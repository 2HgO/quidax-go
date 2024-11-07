package requests

type FetchUserWalletRequest struct {
	UserID   string `uri:"user_id"`
	Currency string `uri:"currency" validate:"oneof=ngn usdt usdc eth bnb sol btc"`
}
