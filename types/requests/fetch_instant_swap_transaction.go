package requests

type FetchInstantSwapTransactionRequest struct {
	UserID            string `uri:"user_id" validate:"required"`
	SwapTransactionID string `uri:"swap_transaction_id" validate:"required"`
}
