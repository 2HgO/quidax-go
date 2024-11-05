package requests

type ConfirmInstanSwapRequest struct {
	UserID      string `uri:"user_id"`
	QuotationID string `uri:"quotation_id"`
}
