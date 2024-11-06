package responses

import "github.com/2HgO/quidax-go/models"

type UserWalletResponseData struct {
	ID            string          `json:"id"`
	Currency      string          `json:"currency"`
	Balance       float64         `json:"balance,string"`
	LockedBalance float64         `json:"locked_balance,string"`
	User          *models.Account `json:"user"`
}
