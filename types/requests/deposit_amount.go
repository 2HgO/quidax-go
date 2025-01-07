package requests

import "github.com/2HgO/quidax-go/models"

type DepositAmountRequest struct {
	UserID   string        `uri:"user_id"`
	Currency string        `uri:"currency" validate:"required,oneof=ngn usdt usdc eth bnb sol btc"`
	Amount   models.Double `json:"amount"`
}
