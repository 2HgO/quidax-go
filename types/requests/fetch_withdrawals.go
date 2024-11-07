package requests

import "github.com/2HgO/quidax-go/models"

type FetchWithdrawalsRequest struct {
	UserID   string                   `uri:"user_id"`
	Currency *string                  `query:"currency" validate:"omitempty,oneof=ngn usdt usdc eth bnb sol btc"`
	State    *models.WithdrawalStatus `query:"state"`
}
