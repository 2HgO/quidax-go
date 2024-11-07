package requests

type CreateWithdrawalRequest struct {
	UserID          string  `uri:"user_id"`
	FundUid         string  `json:"fund_uid" validate:"required"`
	Currency        string  `json:"currency" validate:"required,oneof=ngn usdt usdc eth bnb sol btc"`
	Amount          float64 `json:"amount,string" validate:"required,gt=0"`
	TransactionNote string  `json:"transaction_note"`
	Narration       string  `json:"narration"`
}
