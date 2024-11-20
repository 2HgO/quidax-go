package requests

type FetchDepositRequest struct {
	UserID        string `uri:"user_id"`
	TransactionID string `uri:"transaction_id"`
}
