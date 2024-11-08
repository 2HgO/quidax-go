package requests

type RefreshInstantSwapRequest struct {
	UserID       string  `uri:"user_id" validate:"required"`
	QuotationID  string  `uri:"quotation_id" validate:"required"`
	FromCurrency string  `json:"from_currency"`
	ToCurrency   string  `json:"to_currency"`
	FromAmount   float64 `json:"from_amount,string"`
}
