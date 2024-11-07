package requests

type CreateInstantSwapRequest struct {
	UserID       string  `uri:"user_id"`
	FromCurrency string  `json:"from_currency" validate:"required,oneof=ngn usdt usdc eth bnb sol btc"`
	ToCurrency   string  `json:"to_currency" validate:"required,oneof=ngn usdt usdc eth bnb sol btc"`
	FromAmount   float64 `json:"from_amount,string" validate:"required,gt=0"`
}
