package requests

type FetchAccountDetailsRequest struct {
	UserID string `uri:"user_id" validate:"required"`
}
