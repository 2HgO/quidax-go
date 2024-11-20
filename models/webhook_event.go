package models

import "encoding/json"

type Webhook struct {
	Event WebhookEvent `json:"event"`
	Data  any          `json:"data"`
}

type WebhookEvent uint8

const (
	WalletUpdated_WebhookEvent WebhookEvent = iota + 1

	SwapTransactionCompleted_WebhookEvent
	SwapTransactionReversed_WebhookEvent
	SwapTransactionFailed_WebhookEvent

	WithdrawalSuccessful_WebhookEvent
	WithdrawalRejected_WebhookEvent

	DepositSuccessful_WebhookEvent
)

func (w WebhookEvent) String() string {
	switch w {
	case WalletUpdated_WebhookEvent:
		return "wallet.updated"
	case SwapTransactionCompleted_WebhookEvent:
		return "swap_transaction.completed"
	case SwapTransactionReversed_WebhookEvent:
		return "swap_transaction.reversed"
	case SwapTransactionFailed_WebhookEvent:
		return "swap_transaction.failed"
	case WithdrawalSuccessful_WebhookEvent:
		return "withdraw.successful"
	case WithdrawalRejected_WebhookEvent:
		return "withdraw.rejected"
	case DepositSuccessful_WebhookEvent:
		return "deposit.successful"
	default:
		panic("unreachable")
	}
}

func (w WebhookEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.String())
}
