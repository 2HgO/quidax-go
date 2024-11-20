package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/models"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/types/responses"
	"github.com/2HgO/quidax-go/utils"
	sq "github.com/Masterminds/squirrel"
	tdb "github.com/tigerbeetle/tigerbeetle-go"
	tdb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"go.uber.org/zap"

	"github.com/madflojo/tasks"
)

type SchedulerService interface {
	ScheduleInstantSwapReversal(string, time.Time)
	// ScheduleEventRetry(parent *models.Account, event *models.Webhook)
}

func NewSchedulerService(dataDB *sql.DB, txDatabase tdb.Client, scheduler *tasks.Scheduler, accountService AccountService, walletService WalletService, webhookService WebhookService, log *zap.Logger) SchedulerService {
	return &schedulerService{
		service{
			transactionDB:  txDatabase,
			webhookService: webhookService,
			accountService: accountService,
			walletService:  walletService,
			log:            log,
			dataDB:         dataDB,
		},
		scheduler,
	}
}

type schedulerService struct {
	service
	scheduler *tasks.Scheduler
}

func (s *schedulerService) ScheduleInstantSwapReversal(id string, dueAt time.Time) {
	s.scheduler.AddWithID(id, &tasks.Task{
		RunOnce:    true,
		Interval:   1 * time.Second,
		StartAfter: dueAt,
		TaskFunc: func() error {
			s.log.Info("attempting to reverse instant swap transfer...")
			row := sq.
				Select("instant_swaps.id", "quotation_id", "from_wallet_id", "to_wallet_id", "quotation_rate", "execution_rate", "swap_tx_id_0", "swap_tx_id_1", "quote_tx_id_0", "quote_tx_id_1", "wallets.token", "wallets.account_id").
				From("instant_swaps").
				Join("wallets on wallets.id = instant_swaps.from_wallet_id").
				Where(sq.Eq{"quotation_id": id}).
				RunWith(s.dataDB).
				QueryRow()
			if row == nil {
				return errors.NewNotFoundError("swap not found")
			}

			var swap models.InstantSwap
			var wallet models.Wallet
			err := row.Scan(
				&swap.ID,
				&swap.QuotationID,
				&swap.FromWalletID,
				&swap.ToWalletID,
				&swap.QuotationRate,
				&swap.ExecutionRate,
				&swap.SwapTxID0,
				&swap.SwapTxID1,
				&swap.QuoteTxID0,
				&swap.QuoteTxID1,
				&wallet.Token,
				&wallet.AccountID,
			)
			if err != nil {
				s.log.Error("fetching instant swap for reversal", zap.Error(err))
				return err
			}

			user, err := s.accountService.FetchAccountDetails(context.WithValue(context.Background(), "skip_check", true), &requests.FetchAccountDetailsRequest{UserID: wallet.AccountID})
			if err != nil {
				s.log.Error("fetching user details for instant swap reversal", zap.Error(err))
				return err
			}

			qtx0, _ := tdb_types.HexStringToUint128(swap.QuoteTxID0)
			qtx1, _ := tdb_types.HexStringToUint128(swap.QuoteTxID1)
			transactions, err := s.transactionDB.LookupTransfers([]tdb_types.Uint128{qtx0, qtx1})
			if err != nil {
				return errors.HandleTxDBError(err)
			}
			if len(transactions) != 2 {
				s.log.Error("fetching pending transactions", zap.Error(errors.NewFailedDependencyError("transaction not found")))
				return errors.NewFailedDependencyError("transaction not found")
			}

			stx0, _ := tdb_types.HexStringToUint128(swap.SwapTxID0)
			stx1, _ := tdb_types.HexStringToUint128(swap.SwapTxID1)
			res, err := s.transactionDB.CreateTransfers([]tdb_types.Transfer{
				{
					ID:              stx0,
					CreditAccountID: transactions[0].CreditAccountID,
					DebitAccountID:  transactions[0].DebitAccountID,
					Ledger:          transactions[0].Ledger,
					UserData128:     transactions[0].UserData128,
					PendingID:       transactions[0].ID,
					Code:            1,
					Flags: tdb_types.TransferFlags{
						Linked:              true,
						VoidPendingTransfer: true,
					}.ToUint16(),
				},
				{
					ID:              stx1,
					CreditAccountID: transactions[1].CreditAccountID,
					DebitAccountID:  transactions[1].DebitAccountID,
					Ledger:          transactions[1].Ledger,
					UserData128:     transactions[1].UserData128,
					PendingID:       transactions[1].ID,
					Code:            1,
					Flags: tdb_types.TransferFlags{
						VoidPendingTransfer: true,
					}.ToUint16(),
				},
			})
			if err != nil {
				s.log.Error("reversing pending transaction", zap.Error(err))
				return err
			}

			if len(res) > 0 {
				for _, r := range res {
					if r.Result != tdb_types.TransferPendingTransferNotPending {
						s.log.Error("reversing pending transactions", zap.String("error status", r.Result.String()))
					}
				}
				return nil
			}

			fromAmount := utils.FromAmount(transactions[0].Amount)
			toAmount := utils.FromAmount(transactions[1].Amount)
			now := time.Now()
			data := &responses.InstantSwapResponseData{
				ID:             swap.ID,
				FromCurrency:   Ledgers[transactions[0].Ledger],
				ToCurrency:     Ledgers[transactions[1].Ledger],
				ExecutionPrice: swap.ExecutionRate,
				FromAmount:     utils.ApproximateAmount(Ledgers[transactions[0].Ledger], fromAmount),
				ReceivedAmount: utils.ApproximateAmount(Ledgers[transactions[1].Ledger], toAmount),
				CreatedAt:      now,
				UpdatedAt:      now,
				User:           user.Data,
				Status:         "reversed",
				SwapQuotation: &responses.InstantSwapQuotationResponseData{
					ID:             swap.QuotationID,
					FromCurrency:   Ledgers[transactions[0].Ledger],
					ToCurrency:     Ledgers[transactions[1].Ledger],
					QuotedPrice:    swap.QuotationRate,
					QuotedCurrency: Ledgers[transactions[1].Ledger],
					FromAmount:     utils.ApproximateAmount(Ledgers[transactions[0].Ledger], fromAmount),
					ToAmount:       utils.ApproximateAmount(Ledgers[transactions[1].Ledger], toAmount),
					Confirmed:      false,
					ExpiresAt:      time.UnixMicro(int64(transactions[0].Timestamp / 1000)).Add(12 * time.Second),
					CreatedAt:      time.UnixMicro(int64(transactions[0].Timestamp / 1000)),
					User:           user.Data,
				},
			}
			if data.SwapQuotation.FromCurrency == "ngn" {
				data.SwapQuotation.QuotedCurrency = data.SwapQuotation.FromCurrency
			}

			s.webhookService.SendInstantSwapReversedEvent(user.Data.WebhookDetails, data)
			return nil
		},
	})
}
