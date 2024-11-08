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

	WithdrawalCompleted_WebhookEvent
	WithdrawalFailed_WebhookEvent
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
	case WithdrawalCompleted_WebhookEvent:
		return "withdrawal.completed"
	case WithdrawalFailed_WebhookEvent:
		return "withdrawal.failed"
	default:
		panic("unreachable")
	}
}

func (w WebhookEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(w.String())
}
