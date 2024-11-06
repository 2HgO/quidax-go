package requests

type FetchUserWalletRequest struct {
	UserID   string `uri:"user_id"`
	Currency string `uri:"currency"`
}
