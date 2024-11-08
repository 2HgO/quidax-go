package services

import (
	"context"
	"time"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/models"
	"github.com/2HgO/quidax-go/types/responses"
	"github.com/2HgO/quidax-go/utils"
	tdb "github.com/tigerbeetle/tigerbeetle-go"
	tdb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
	"go.uber.org/zap"

	"github.com/madflojo/tasks"
)

type SchedulerService interface {
	DropTask(taskID string)
	ScheduleInstantSwapReversal(parent, user *models.Account, confirmationID tdb_types.Uint128, swap *responses.InstantSwapQuotationResponseData)
	GetConfirmationID(taskID string) (tdb_types.Uint128, bool)
	// ScheduleEventRetry(parent *models.Account, event *models.Webhook)
}

func NewSchedulerService(txDatabase tdb.Client, scheduler *tasks.Scheduler, webhookService WebhookService, log *zap.Logger) SchedulerService {
	return &schedulerService{
		service: service{
			transactionDB:  txDatabase,
			webhookService: webhookService,
			log:            log,
		},
		scheduler: scheduler,
	}
}

type schedulerService struct {
	service
	scheduler *tasks.Scheduler
}

func (s *schedulerService) DropTask(taskID string) {
	s.scheduler.Del(taskID)
}

func (s *schedulerService) ScheduleInstantSwapReversal(parent, user *models.Account, confirmationID tdb_types.Uint128, swap *responses.InstantSwapQuotationResponseData) {
	ctx := context.WithValue(context.Background(), "confirmation_id", confirmationID)
	s.scheduler.AddWithID(swap.ID, &tasks.Task{
		TaskContext: tasks.TaskContext{Context: ctx},
		RunOnce:     true,
		Interval:    1 * time.Second,
		StartAfter:  swap.ExpiresAt,
		FuncWithTaskContext: func(t tasks.TaskContext) error {
			s.log.Info("attempting to reverse instant swap transfer...")
			quoteRef, err := tdb_types.HexStringToUint128(swap.ID)
			if err != nil {
				s.log.Error("parsing swap quotation id", zap.Error(err))
				return nil
			}
			quoteRefInt := quoteRef.BigInt()
			transactions, err := s.transactionDB.QueryTransfers(tdb_types.QueryFilter{
				UserData64: quoteRefInt.Uint64(),
				Limit:      2,
			})
			if err != nil {
				s.log.Error("fetching pending transactions", zap.Error(err))
				return nil
			}
			if len(transactions) != 2 {
				s.log.Error("fetching pending transactions", zap.Error(errors.NewFailedDependencyError("transaction not found")))
				return nil
			}

			ref := t.Context.Value("confirmation_id").(tdb_types.Uint128).BigInt()
			res, err := s.transactionDB.CreateTransfers([]tdb_types.Transfer{
				{
					ID:              tdb_types.ID(),
					CreditAccountID: transactions[0].CreditAccountID,
					DebitAccountID:  transactions[0].DebitAccountID,
					Ledger:          transactions[0].Ledger,
					UserData128:     transactions[0].UserData128,
					UserData64:      ref.Uint64(),
					PendingID:       transactions[0].ID,
					Code:            1,
					Flags: tdb_types.TransferFlags{
						Linked:              true,
						VoidPendingTransfer: true,
					}.ToUint16(),
				},
				{
					ID:              tdb_types.ID(),
					CreditAccountID: transactions[1].CreditAccountID,
					DebitAccountID:  transactions[1].DebitAccountID,
					Ledger:          transactions[1].Ledger,
					UserData128:     transactions[1].UserData128,
					UserData64:      ref.Uint64(),
					PendingID:       transactions[1].ID,
					Code:            1,
					Flags: tdb_types.TransferFlags{
						VoidPendingTransfer: true,
					}.ToUint16(),
				},
			})
			if err != nil {
				s.log.Error("reversing pending transaction", zap.Error(err))
				return nil
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
				ID:             tdb_types.ToUint128(ref.Uint64()).String(),
				FromCurrency:   Ledgers[transactions[0].Ledger],
				ToCurrency:     Ledgers[transactions[1].Ledger],
				ExecutionPrice: utils.ApproximateAmount(Ledgers[transactions[0].Ledger], toAmount/fromAmount),
				FromAmount:     utils.ApproximateAmount(Ledgers[transactions[0].Ledger], fromAmount),
				ReceivedAmount: utils.ApproximateAmount(Ledgers[transactions[1].Ledger], toAmount),
				CreatedAt:      now,
				UpdatedAt:      now,
				User:           user,
				Status:         "reversed",
				SwapQuotation: &responses.InstantSwapQuotationResponseData{
					ID:             tdb_types.ToUint128(transactions[0].UserData64).String(),
					FromCurrency:   Ledgers[transactions[0].Ledger],
					ToCurrency:     Ledgers[transactions[1].Ledger],
					QuotedPrice:    utils.ApproximateAmount(Ledgers[transactions[0].Ledger], toAmount/fromAmount),
					QuotedCurrency: Ledgers[transactions[0].Ledger],
					FromAmount:     utils.ApproximateAmount(Ledgers[transactions[0].Ledger], fromAmount),
					ToAmount:       utils.ApproximateAmount(Ledgers[transactions[1].Ledger], toAmount),
					Confirmed:      false,
					ExpiresAt:      time.UnixMicro(int64(transactions[0].Timestamp / 1000)).Add(12 * time.Second),
					CreatedAt:      time.UnixMicro(int64(transactions[0].Timestamp / 1000)),
					User:           user,
				},
			}

			s.webhookService.SendInstantSwapReversedEvent(parent, data)
			return nil
		},
	})
}

func (s *schedulerService) GetConfirmationID(taskID string) (tdb_types.Uint128, bool) {
	task, ok := s.scheduler.Tasks()[taskID]
	if !ok {
		return tdb_types.Uint128{}, ok
	}

	id, ok := task.TaskContext.Context.Value("confirmation_id").(tdb_types.Uint128)
	return id, ok
}
