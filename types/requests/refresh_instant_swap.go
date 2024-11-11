package requests

import "github.com/2HgO/quidax-go/models"

type RefreshInstantSwapRequest struct {
	UserID       string        `uri:"user_id" validate:"required"`
	QuotationID  string        `uri:"quotation_id" validate:"required"`
	FromCurrency string        `json:"from_currency"`
	ToCurrency   string        `json:"to_currency"`
	FromAmount   models.Double `json:"from_amount"`
}
