package services

import (
	"context"
	"database/sql"
	"math"
	"time"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/models"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/types/responses"
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

func NewInstantSwapService(txDatabase tdb.Client, dataDatabase *sql.DB, accountService AccountService, walletService WalletService, log *zap.Logger) InstantSwapService {
	return &instantSwapService{
		service: service{
			transactionDB:  txDatabase,
			dataDB:         dataDatabase,
			accountService: accountService,
			walletService:  walletService,
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

func (i *instantSwapService) approximateAmount(currency string, amount float64) float64 {
	switch currency {
	case "sol":
		return math.Floor(amount*1000000) / 1000000
	case "btc":
		return math.Floor(amount*100000000) / 100000000
	case "bnb":
		return math.Floor(amount*100000) / 100000
	case "eth":
		return math.Floor(amount*1000000) / 1000000
	default:
		return math.Floor(amount*100) / 100
	}
}

func (i *instantSwapService) normalizeTransaction(from string, to string, amount float64) normalizedSwapTransaction {
	fromAmount := i.approximateAmount(from, amount)
	toAmount := i.approximateAmount(to, Rates[from][to]*fromAmount)

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

	if fromWallet.Data.Balance < transactionDetails.fromAmount {
		return nil, errors.NewFailedDependencyError("Insufficient Balance")
	}

	ref := tdb_types.ID().BigInt()
	now := time.Now()
	timeout := now.Add(time.Hour)
	transactions := []tdb_types.Transfer{
		{
			ID:              tdb_types.ID(),
			CreditAccountID: tdb_types.ToUint128(uint64(LedgerIDs[req.FromCurrency])),
			DebitAccountID:  fromWalletID,
			Amount:          tdb_types.ToUint128(uint64(math.Floor(transactionDetails.fromAmount * 10e8))),
			Timeout:         uint32(timeout.Unix()),
			UserData128:     tdb_types.BytesToUint128(uuid.MustParse(fromWallet.Data.User.ID)),
			UserData64:      ref.Uint64(),
			UserData32:      uint32(float32(math.Floor(Rates[req.FromCurrency][req.ToCurrency] * 10e8))),
			Ledger:          LedgerIDs[req.FromCurrency],
			Code:            1,
			Flags: tdb_types.TransferFlags{
				Linked:  true,
				Pending: true,
			}.ToUint16(),
		},
		{
			ID:              tdb_types.ID(),
			DebitAccountID:  tdb_types.ToUint128(uint64(LedgerIDs[req.ToCurrency])),
			CreditAccountID: toWalletID,
			Amount:          tdb_types.ToUint128(uint64(math.Floor(transactionDetails.toAmount * 10e8))),
			Ledger:          LedgerIDs[req.ToCurrency],
			Timeout:         uint32(timeout.Unix()),
			UserData128:     tdb_types.BytesToUint128(uuid.MustParse(fromWallet.Data.User.ID)),
			UserData64:      ref.Uint64(),
			UserData32:      uint32(float32(math.Floor(Rates[req.FromCurrency][req.ToCurrency] * 10e8))),
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
		return nil, errors.NewFailedDependencyError(res[0].Result.String())
	}

	return &responses.Response[*responses.InstantSwapQuotationResponseData]{
		Status: "successful",
		Data: &responses.InstantSwapQuotationResponseData{
			ID:             tdb_types.ToUint128(ref.Uint64()).String(),
			FromCurrency:   req.FromCurrency,
			ToCurrency:     req.ToCurrency,
			QuotedPrice:    i.approximateAmount(req.FromCurrency, float64(float32(transactions[0].UserData32)) * 1e-9),
			QuotedCurrency: req.FromCurrency,
			FromAmount:     transactionDetails.fromAmount,
			ToAmount:       transactionDetails.toAmount,
			Confirmed:      false,
			ExpiresAt:      timeout,
			CreatedAt:      now,
			User:           fromWallet.Data.User,
		},
	}, nil
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

	fromAmount := transactions[0].Amount.BigInt()
	toAmount := transactions[1].Amount.BigInt()

	ref := tdb_types.ID().BigInt()
	now := time.Now()
	timeExceeded := transactions[0].Timeout < uint32(now.Unix())
	confirmedTransactions := []tdb_types.Transfer{
		{
			ID:              tdb_types.ID(),
			CreditAccountID: transactions[0].CreditAccountID,
			DebitAccountID:  transactions[0].DebitAccountID,
			Amount:          transactions[0].Amount,
			Ledger:          transactions[0].Ledger,
			UserData128:     transactions[0].UserData128,
			UserData64:      ref.Uint64(),
			UserData32:      transactions[0].UserData32,
			PendingID:       transactions[0].ID,
			Code:            1,
			Flags: tdb_types.TransferFlags{
				Linked:              true,
				PostPendingTransfer: !timeExceeded,
				VoidPendingTransfer: timeExceeded,
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
			UserData32:      transactions[1].UserData32,
			PendingID:       transactions[1].ID,
			Code:            1,
			Flags: tdb_types.TransferFlags{
				PostPendingTransfer: !timeExceeded,
				VoidPendingTransfer: timeExceeded,
			}.ToUint16(),
		},
	}

	res, err := i.transactionDB.CreateTransfers(confirmedTransactions)
	if err != nil {
		return nil, errors.HandleTxDBError(err)
	}

	if len(res) > 0 {
		return nil, errors.NewUnknownError(res[0].Result.String())
	}

	if timeExceeded {
		return nil, errors.NewFailedDependencyError("transaction timed out")
	}

	return &responses.Response[*responses.InstantSwapResponseData]{
		Status: "successful",
		Data: &responses.InstantSwapResponseData{
			ID:           tdb_types.ToUint128(ref.Uint64()).String(),
			FromCurrency: Ledgers[transactions[0].Ledger],
			ToCurrency:   Ledgers[transactions[1].Ledger],
			ExecutionPrice: i.approximateAmount(Ledgers[transactions[0].Ledger], float64(float32(confirmedTransactions[0].UserData32)) * 1e-9),
			FromAmount:     i.approximateAmount(Ledgers[transactions[0].Ledger], float64(fromAmount.Uint64()) * 1e-9),
			ReceivedAmount: i.approximateAmount(Ledgers[transactions[1].Ledger], float64(toAmount.Uint64()) * 1e-9),
			CreatedAt:      now,
			UpdatedAt:      now,
			User:           user.Data,
			Status:         "confirmed",
			SwapQuotation: &responses.InstantSwapQuotationResponseData{
				ID:           tdb_types.ToUint128(transactions[0].UserData64).String(),
				FromCurrency: Ledgers[transactions[0].Ledger],
				ToCurrency:   Ledgers[transactions[1].Ledger],
				QuotedPrice:    i.approximateAmount(Ledgers[transactions[0].Ledger], float64(float32(transactions[0].UserData32)) * 1e-9),
				QuotedCurrency: Ledgers[transactions[0].Ledger],
				FromAmount:     i.approximateAmount(Ledgers[transactions[0].Ledger], float64(fromAmount.Uint64()) * 1e-9),
				ToAmount:       i.approximateAmount(Ledgers[transactions[1].Ledger], float64(toAmount.Uint64()) * 1e-9),
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
	if (len(res) + len(res2) < 4) {
		return nil, errors.NewNotFoundError("swap not found")
	}

	data, err := i.groupTransactions(append(res, res2...), user.Data)
	if err != nil {
		return nil, err
	}

	return &responses.Response[*responses.InstantSwapResponseData]{
		Status: "successful",
		Data: data[0],
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
		Data: data,
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
		transactionFromAmount := transactions[0].Amount.BigInt()
		transactionToAmount := transactions[1].Amount.BigInt()

		quotations := [2]tdb_types.Transfer{pendingTxMap[transactions[0].PendingID.String()], pendingTxMap[transactions[1].PendingID.String()]}
		quotationFromAmount := quotations[0].Amount.BigInt()
		quotationToAmount := quotations[1].Amount.BigInt()

		status := ""
		switch {
		case transactions[0].TransferFlags().PostPendingTransfer:
			status = "confirmed"
		default:
			status = "failed"
		}

		result = append(result, &responses.InstantSwapResponseData{
			ID:           tdb_types.ToUint128(transactions[0].UserData64).String(),
			FromCurrency: Ledgers[transactions[0].Ledger],
			ToCurrency:   Ledgers[transactions[1].Ledger],
			ExecutionPrice: i.approximateAmount(Ledgers[transactions[0].Ledger], float64(float32(transactions[0].UserData32)) * 1e-9),
			FromAmount:     i.approximateAmount(Ledgers[transactions[0].Ledger], float64(transactionFromAmount.Uint64()) * 1e-9),
			ReceivedAmount: i.approximateAmount(Ledgers[transactions[1].Ledger], float64(transactionToAmount.Uint64()) * 1e-9),
			CreatedAt:      time.UnixMicro(int64(transactions[0].Timestamp / 1000)),
			UpdatedAt:      time.UnixMicro(int64(transactions[0].Timestamp / 1000)),
			User:           user,
			Status:         status,
			SwapQuotation: &responses.InstantSwapQuotationResponseData{
				ID:             tdb_types.ToUint128(quotations[0].UserData64).String(),
				FromCurrency:   Ledgers[quotations[0].Ledger],
				ToCurrency:     Ledgers[quotations[1].Ledger],
				QuotedPrice:    i.approximateAmount(Ledgers[quotations[0].Ledger], float64(float32(quotations[0].UserData32)) * 1e-9),
				QuotedCurrency: Ledgers[quotations[0].Ledger],
				FromAmount:     i.approximateAmount(Ledgers[quotations[0].Ledger], float64(quotationFromAmount.Uint64()) * 1e-9),
				ToAmount:       i.approximateAmount(Ledgers[quotations[1].Ledger], float64(quotationToAmount.Uint64()) * 1e-9),
				Confirmed:      transactions[0].TransferFlags().PostPendingTransfer,
				ExpiresAt:      time.Unix(int64(quotations[0].Timeout), 0),
				CreatedAt:      time.UnixMicro(int64(quotations[0].Timestamp / 1000)),
				User:           user,
			},
		})
	}

	return result, nil
}
