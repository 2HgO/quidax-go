package requests

type FetchInstantSwapTransactionRequest struct {
	UserID            string `uri:"user_id"`
	SwapTransactionID string `uri:"swap_transaction_id"`
}
