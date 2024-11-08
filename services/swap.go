package services

import (
	"context"
	"database/sql"
	"math/big"
	"time"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/models"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/types/responses"
	"github.com/2HgO/quidax-go/utils"
	"github.com/google/uuid"
	tdb "github.com/tigerbeetle/tigerbeetle-go"
	tdb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"go.uber.org/zap"
)

type InstantSwapService interface {
	CreateInstantSwap(ctx context.Context, req *requests.CreateInstantSwapRequest) (*responses.Response[*responses.InstantSwapQuotationResponseData], error)
	ConfirmInstantSwap(ctx context.Context, req *requests.ConfirmInstanSwapRequest) (*responses.Response[*responses.InstantSwapResponseData], error)
	RefreshInstantSwap(ctx context.Context, req *requests.RefreshInstantSwapRequest) (*responses.Response[*responses.InstantSwapQuotationResponseData], error)
	FetchInstantSwapTransaction(ctx context.Context, req *requests.FetchInstantSwapTransactionRequest) (*responses.Response[*responses.InstantSwapResponseData], error)
	GetInstantSwapTransactions(ctx context.Context, req *requests.GetInstantSwapTransactionsRequest) (*responses.Response[[]*responses.InstantSwapResponseData], error)
	QuoteInstantSwap(ctx context.Context, req *requests.CreateInstantSwapRequest) (*responses.Response[*responses.QuoteInstantSwapResponseData], error)
}

func NewInstantSwapService(
	txDatabase tdb.Client,
	dataDatabase *sql.DB,
	accountService AccountService,
	walletService WalletService,
	scheduler SchedulerService,
	webhookService WebhookService,
	log *zap.Logger,
) InstantSwapService {
	return &instantSwapService{
		service: service{
			transactionDB:  txDatabase,
			dataDB:         dataDatabase,
			accountService: accountService,
			walletService:  walletService,
			webhookService: webhookService,
			scheduler:      scheduler,
			log:            log,
		},
	}
}

type instantSwapService struct {
	InstantSwapService

	service
}

type normalizedSwapTransaction struct {
	fromToken  string
	toToken    string
	fromAmount float64
	toAmount   float64
}

func (i *instantSwapService) normalizeTransaction(from string, to string, amount float64) normalizedSwapTransaction {
	fromAmount := utils.ApproximateAmount(from, amount)
	toAmount := utils.ApproximateAmount(to, Rates[from][to]*fromAmount)

	return normalizedSwapTransaction{
		fromToken:  from,
		toToken:    to,
		fromAmount: fromAmount,
		toAmount:   toAmount,
	}
}

func (i *instantSwapService) CreateInstantSwap(ctx context.Context, req *requests.CreateInstantSwapRequest) (*responses.Response[*responses.InstantSwapQuotationResponseData], error) {
	transactionDetails := i.normalizeTransaction(req.FromCurrency, req.ToCurrency, req.FromAmount)
	fromWallet, err := i.walletService.FetchUserWallet(ctx, &requests.FetchUserWalletRequest{UserID: req.UserID, Currency: req.FromCurrency})
	if err != nil {
		return nil, err
	}
	fromWalletID, err := tdb_types.HexStringToUint128(fromWallet.Data.ID)
	if err != nil {
		return nil, err
	}
	toWallet, err := i.walletService.FetchUserWallet(ctx, &requests.FetchUserWalletRequest{UserID: req.UserID, Currency: req.ToCurrency})
	if err != nil {
		return nil, err
	}
	toWalletID, err := tdb_types.HexStringToUint128(toWallet.Data.ID)
	if err != nil {
		return nil, err
	}

	ref := tdb_types.ID().BigInt()
	now := time.Now()
	timeout := now.Add(time.Second * 12)
	transactions := []tdb_types.Transfer{
		{
			ID:              tdb_types.ID(),
			CreditAccountID: tdb_types.ToUint128(uint64(LedgerIDs[req.FromCurrency])),
			DebitAccountID:  fromWalletID,
			Amount:          utils.ToAmount(transactionDetails.fromAmount),
			UserData128:     tdb_types.BytesToUint128(uuid.MustParse(fromWallet.Data.User.ID)),
			UserData64:      ref.Uint64(),
			// ?todo: store fee information in userdata32
			Ledger: LedgerIDs[req.FromCurrency],
			Code:   1,
			Flags: tdb_types.TransferFlags{
				Linked:  true,
				Pending: true,
			}.ToUint16(),
		},
		{
			ID:              tdb_types.ID(),
			DebitAccountID:  tdb_types.ToUint128(uint64(LedgerIDs[req.ToCurrency])),
			CreditAccountID: toWalletID,
			Amount:          utils.ToAmount(transactionDetails.toAmount),
			Ledger:          LedgerIDs[req.ToCurrency],
			UserData128:     tdb_types.BytesToUint128(uuid.MustParse(fromWallet.Data.User.ID)),
			UserData64:      ref.Uint64(),
			Code:            1,
			Flags: tdb_types.TransferFlags{
				Pending: true,
			}.ToUint16(),
		},
	}

	res, err := i.transactionDB.CreateTransfers(transactions)
	if err != nil {
		return nil, errors.HandleTxDBError(err)
	}
	if len(res) > 0 {
		for _, r := range res {
			if r.Result == tdb_types.TransferExceedsCredits {
				return nil, errors.NewFailedDependencyError("Insufficient Balance")
			}
		}
		return nil, errors.NewFailedDependencyError(res[0].Result.String())
	}

	data := &responses.InstantSwapQuotationResponseData{
		ID:             tdb_types.ToUint128(ref.Uint64()).String(),
		FromCurrency:   req.FromCurrency,
		ToCurrency:     req.ToCurrency,
		QuotedPrice:    utils.ApproximateAmount(req.FromCurrency, transactionDetails.toAmount/transactionDetails.fromAmount),
		QuotedCurrency: req.FromCurrency,
		FromAmount:     transactionDetails.fromAmount,
		ToAmount:       transactionDetails.toAmount,
		Confirmed:      false,
		ExpiresAt:      timeout,
		CreatedAt:      now,
		User:           fromWallet.Data.User,
	}

	parent := ctx.Value("user").(*models.Account)
	confirmationRef := tdb_types.ID()
	i.scheduler.ScheduleInstantSwapReversal(parent, fromWallet.Data.User, confirmationRef, data)
	go i.webhookService.SendWalletUpdatedEvent(parent, fromWallet.Data)

	return &responses.Response[*responses.InstantSwapQuotationResponseData]{
		Status: "successful",
		Data:   data,
	}, nil
}

func (i *instantSwapService) processSwap(ref big.Int, ts time.Time, transactions []tdb_types.Transfer, parent, user *models.Account) {
	failed := false

	fromAmount := utils.FromAmount(transactions[0].Amount)
	toAmount := utils.FromAmount(transactions[1].Amount)
	confirmedTransactions := []tdb_types.Transfer{
		{
			ID:              tdb_types.ID(),
			CreditAccountID: transactions[0].CreditAccountID,
			DebitAccountID:  transactions[0].DebitAccountID,
			Amount:          transactions[0].Amount,
			Ledger:          transactions[0].Ledger,
			UserData128:     transactions[0].UserData128,
			UserData64:      ref.Uint64(),
			PendingID:       transactions[0].ID,
			Code:            1,
			Flags: tdb_types.TransferFlags{
				Linked:              true,
				PostPendingTransfer: true,
			}.ToUint16(),
		},
		{
			ID:              tdb_types.ID(),
			CreditAccountID: transactions[1].CreditAccountID,
			DebitAccountID:  transactions[1].DebitAccountID,
			Amount:          transactions[1].Amount,
			Ledger:          transactions[1].Ledger,
			UserData128:     transactions[1].UserData128,
			UserData64:      ref.Uint64(),
			PendingID:       transactions[1].ID,
			Code:            1,
			Flags: tdb_types.TransferFlags{
				PostPendingTransfer: true,
			}.ToUint16(),
		},
	}

	res, err := i.transactionDB.CreateTransfers(confirmedTransactions)
	if err != nil {
		i.log.Error("processing swap transaction", zap.Error(err))
		return
	}

	if len(res) > 0 {
		for _, r := range res {
			if r.Result == tdb_types.TransferExceedsCredits {
				for i := range confirmedTransactions {
					confirmedTransactions[i].Flags = tdb_types.TransferFlags{
						Linked:              i == 0,
						VoidPendingTransfer: true,
					}.ToUint16()
				}
				res, err = i.transactionDB.CreateTransfers(confirmedTransactions)
				if err != nil {
					i.log.Error("failing swap transaction", zap.Error(err))
				}
				failed = true
				goto failedTransfer
			}
			i.log.Error("processing swap transaction", zap.String("info", r.Result.String()))
		}
		// ?todo: retry transfer confirmation?
		return
	}

failedTransfer:
	data := &responses.InstantSwapResponseData{
		ID:             tdb_types.ToUint128(ref.Uint64()).String(),
		FromCurrency:   Ledgers[transactions[0].Ledger],
		ToCurrency:     Ledgers[transactions[1].Ledger],
		ExecutionPrice: utils.ApproximateAmount(Ledgers[transactions[0].Ledger], toAmount/fromAmount),
		FromAmount:     utils.ApproximateAmount(Ledgers[transactions[0].Ledger], fromAmount),
		ReceivedAmount: utils.ApproximateAmount(Ledgers[transactions[1].Ledger], toAmount),
		CreatedAt:      ts,
		UpdatedAt:      ts,
		User:           user,
		Status:         "confirmed",
		SwapQuotation: &responses.InstantSwapQuotationResponseData{
			ID:             tdb_types.ToUint128(transactions[0].UserData64).String(),
			FromCurrency:   Ledgers[transactions[0].Ledger],
			ToCurrency:     Ledgers[transactions[1].Ledger],
			QuotedPrice:    utils.ApproximateAmount(Ledgers[transactions[0].Ledger], toAmount/fromAmount),
			QuotedCurrency: Ledgers[transactions[0].Ledger],
			FromAmount:     utils.ApproximateAmount(Ledgers[transactions[0].Ledger], fromAmount),
			ToAmount:       utils.ApproximateAmount(Ledgers[transactions[1].Ledger], toAmount),
			Confirmed:      true,
			ExpiresAt:      time.Unix(int64(transactions[0].Timeout), 0),
			CreatedAt:      time.UnixMicro(int64(transactions[0].Timestamp / 1000)),
			User:           user,
		},
	}

	switch failed {
	case true:
		i.webhookService.SendInstantSwapFailedEvent(parent, data)

		// todo: send wallet updated event for debit wallet
	default:
		i.webhookService.SendInstantSwapCompletedEvent(parent, data)

		// todo: send wallet updated event for credit and debit wallets
	}
}

func (i *instantSwapService) ConfirmInstantSwap(ctx context.Context, req *requests.ConfirmInstanSwapRequest) (*responses.Response[*responses.InstantSwapResponseData], error) {
	user, err := i.accountService.FetchAccountDetails(ctx, &requests.FetchAccountDetailsRequest{UserID: req.UserID})
	if err != nil {
		return nil, err
	}

	quoteRef, err := tdb_types.HexStringToUint128(req.QuotationID)
	if err != nil {
		return nil, err
	}
	quoteRefInt := quoteRef.BigInt()
	transactions, err := i.transactionDB.QueryTransfers(tdb_types.QueryFilter{
		UserData64: quoteRefInt.Uint64(),
		Limit:      2,
	})
	if err != nil {
		return nil, errors.HandleTxDBError(err)
	}
	if len(transactions) != 2 {
		return nil, errors.NewFailedDependencyError("transaction not found")
	}

	userId := uuid.UUID(transactions[0].UserData128.Bytes())
	if userId.String() != user.Data.ID {
		return nil, errors.NewValidationError("invalid user id provided")
	}

	fromAmount := utils.FromAmount(transactions[0].Amount)
	toAmount := utils.FromAmount(transactions[1].Amount)

	refId, ok := i.scheduler.GetConfirmationID(req.QuotationID)
	if !ok {
		return nil, errors.NewFailedDependencyError("swap has already been processed")
	}
	ref := refId.BigInt()
	now := time.Now()

	// * stop transfer reversal if not already started
	i.scheduler.DropTask(req.QuotationID)
	go i.processSwap(ref, now, transactions, ctx.Value("user").(*models.Account), user.Data)

	return &responses.Response[*responses.InstantSwapResponseData]{
		Status: "successful",
		Data: &responses.InstantSwapResponseData{
			ID:             tdb_types.ToUint128(ref.Uint64()).String(),
			FromCurrency:   Ledgers[transactions[0].Ledger],
			ToCurrency:     Ledgers[transactions[1].Ledger],
			ExecutionPrice: utils.ApproximateAmount(Ledgers[transactions[0].Ledger], toAmount/fromAmount),
			FromAmount:     utils.ApproximateAmount(Ledgers[transactions[0].Ledger], fromAmount),
			ReceivedAmount: utils.ApproximateAmount(Ledgers[transactions[1].Ledger], toAmount),
			CreatedAt:      now,
			UpdatedAt:      now,
			User:           user.Data,
			Status:         "pending",
			SwapQuotation: &responses.InstantSwapQuotationResponseData{
				ID:             tdb_types.ToUint128(transactions[0].UserData64).String(),
				FromCurrency:   Ledgers[transactions[0].Ledger],
				ToCurrency:     Ledgers[transactions[1].Ledger],
				QuotedPrice:    utils.ApproximateAmount(Ledgers[transactions[0].Ledger], toAmount/fromAmount),
				QuotedCurrency: Ledgers[transactions[0].Ledger],
				FromAmount:     utils.ApproximateAmount(Ledgers[transactions[0].Ledger], fromAmount),
				ToAmount:       utils.ApproximateAmount(Ledgers[transactions[1].Ledger], toAmount),
				Confirmed:      true,
				ExpiresAt:      time.Unix(int64(transactions[0].Timeout), 0),
				CreatedAt:      time.UnixMicro(int64(transactions[0].Timestamp / 1000)),
				User:           user.Data,
			},
		},
	}, nil
}

func (i *instantSwapService) FetchInstantSwapTransaction(ctx context.Context, req *requests.FetchInstantSwapTransactionRequest) (*responses.Response[*responses.InstantSwapResponseData], error) {
	user, err := i.accountService.FetchAccountDetails(ctx, &requests.FetchAccountDetailsRequest{UserID: req.UserID})
	if err != nil {
		return nil, err
	}
	id, err := tdb_types.HexStringToUint128(req.SwapTransactionID)
	if err != nil {
		return nil, err
	}
	idInt := id.BigInt()
	res, err := i.transactionDB.QueryTransfers(tdb_types.QueryFilter{
		UserData128: tdb_types.BytesToUint128(uuid.MustParse(user.Data.ID)),
		UserData64:  idInt.Uint64(),
		Limit:       9000,
		Code:        1,
		Flags: tdb_types.QueryFilterFlags{
			Reversed: true,
		}.ToUint32(),
	})
	if err != nil {
		return nil, errors.HandleTxDBError(err)
	}
	quoteIds := make([]tdb_types.Uint128, 0)
	for _, tx := range res {
		quoteIds = append(quoteIds, tx.PendingID)
	}
	res2, err := i.transactionDB.LookupTransfers(quoteIds)
	if err != nil {
		return nil, errors.HandleTxDBError(err)
	}
	if len(res)+len(res2) < 4 {
		return nil, errors.NewNotFoundError("swap not found")
	}

	data, err := i.groupTransactions(append(res, res2...), user.Data)
	if err != nil {
		return nil, err
	}

	return &responses.Response[*responses.InstantSwapResponseData]{
		Status: "successful",
		Data:   data[0],
	}, nil
}

func (i *instantSwapService) GetInstantSwapTransactions(ctx context.Context, req *requests.GetInstantSwapTransactionsRequest) (*responses.Response[[]*responses.InstantSwapResponseData], error) {
	user, err := i.accountService.FetchAccountDetails(ctx, &requests.FetchAccountDetailsRequest{UserID: req.UserID})
	if err != nil {
		return nil, err
	}

	res, err := i.transactionDB.QueryTransfers(tdb_types.QueryFilter{
		UserData128: tdb_types.BytesToUint128(uuid.MustParse(user.Data.ID)),
		Limit:       9000,
		Code:        1,
		Flags: tdb_types.QueryFilterFlags{
			Reversed: true,
		}.ToUint32(),
	})
	if err != nil {
		return nil, errors.HandleTxDBError(err)
	}
	quoteIds := make([]tdb_types.Uint128, 0)
	for _, tx := range res {
		quoteIds = append(quoteIds, tx.PendingID)
	}
	res2, err := i.transactionDB.LookupTransfers(quoteIds)
	if err != nil {
		return nil, errors.HandleTxDBError(err)
	}

	data, err := i.groupTransactions(append(res, res2...), user.Data)
	if err != nil {
		return nil, err
	}

	return &responses.Response[[]*responses.InstantSwapResponseData]{
		Status: "successful",
		Data:   data,
	}, nil
}

func (i *instantSwapService) groupTransactions(txs []tdb_types.Transfer, user *models.Account) ([]*responses.InstantSwapResponseData, error) {
	txMap := map[string][2]tdb_types.Transfer{}
	pendingTxMap := map[string]tdb_types.Transfer{}
	txIds := []string{}
	for _, tx := range txs {
		crAccount := tx.CreditAccountID.BigInt()
		switch {
		case tx.TransferFlags().Pending:
			id := tx.ID.String()
			pendingTxMap[id] = tx
		default:
			id := tdb_types.ToUint128(tx.UserData64).String()
			if _, ok := txMap[id]; !ok {
				txIds = append(txIds, id)
				txMap[id] = [2]tdb_types.Transfer{}
			}
			switch tx.Ledger == uint32(crAccount.Uint64()) {
			case true:
				txMap[id] = [2]tdb_types.Transfer{tx, txMap[id][1]}
			default:
				txMap[id] = [2]tdb_types.Transfer{txMap[id][0], tx}
			}
		}
	}

	result := make([]*responses.InstantSwapResponseData, 0)
	for _, id := range txIds {
		transactions := txMap[id]
		quotations := [2]tdb_types.Transfer{pendingTxMap[transactions[0].PendingID.String()], pendingTxMap[transactions[1].PendingID.String()]}

		status := ""
		switch {
		case transactions[0].TransferFlags().PostPendingTransfer:
			status = "confirmed"
		case time.UnixMicro(int64(quotations[0].Timestamp / 1000)).Add(time.Second * 12).Before(time.UnixMicro(int64(transactions[0].Timestamp / 1000))):
			status = "reversed"
		default:
			status = "failed"
		}

		result = append(result, &responses.InstantSwapResponseData{
			ID:             tdb_types.ToUint128(transactions[0].UserData64).String(),
			FromCurrency:   Ledgers[transactions[0].Ledger],
			ToCurrency:     Ledgers[transactions[1].Ledger],
			ExecutionPrice: utils.ApproximateAmount(Ledgers[transactions[0].Ledger], utils.FromAmount(transactions[1].Amount)/utils.FromAmount(transactions[0].Amount)),
			FromAmount:     utils.ApproximateAmount(Ledgers[transactions[0].Ledger], utils.FromAmount(transactions[0].Amount)),
			ReceivedAmount: utils.ApproximateAmount(Ledgers[transactions[1].Ledger], utils.FromAmount(transactions[1].Amount)),
			CreatedAt:      time.UnixMicro(int64(transactions[0].Timestamp / 1000)),
			UpdatedAt:      time.UnixMicro(int64(transactions[0].Timestamp / 1000)),
			User:           user,
			Status:         status,
			SwapQuotation: &responses.InstantSwapQuotationResponseData{
				ID:             tdb_types.ToUint128(quotations[0].UserData64).String(),
				FromCurrency:   Ledgers[quotations[0].Ledger],
				ToCurrency:     Ledgers[quotations[1].Ledger],
				QuotedPrice:    utils.ApproximateAmount(Ledgers[quotations[0].Ledger], utils.FromAmount(quotations[1].Amount)/utils.FromAmount(quotations[0].Amount)),
				QuotedCurrency: Ledgers[quotations[0].Ledger],
				FromAmount:     utils.ApproximateAmount(Ledgers[quotations[0].Ledger], utils.FromAmount(quotations[0].Amount)),
				ToAmount:       utils.ApproximateAmount(Ledgers[quotations[1].Ledger], utils.FromAmount(quotations[1].Amount)),
				Confirmed:      status != "reversed",
				ExpiresAt:      time.UnixMicro(int64(quotations[0].Timestamp / 1000)).Add(12 * time.Second),
				CreatedAt:      time.UnixMicro(int64(quotations[0].Timestamp / 1000)),
				User:           user,
			},
		})
	}

	return result, nil
}
