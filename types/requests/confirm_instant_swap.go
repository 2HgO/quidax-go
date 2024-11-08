package requests

type ConfirmInstanSwapRequest struct {
	UserID      string `uri:"user_id" validate:"required"`
	QuotationID string `uri:"quotation_id" validate:"required"`
}
