package requests

type CreateInstantSwapRequest struct {
	UserID       string  `uri:"user_id"`
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	FromAmount   float64 `json:"from_amount,string"`
}
