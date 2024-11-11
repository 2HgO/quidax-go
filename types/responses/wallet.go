package responses

import "github.com/2HgO/quidax-go/models"

type UserWalletResponseData struct {
	ID             string          `json:"id"`
	Name           string          `json:"name"`
	Currency       string          `json:"currency"`
	Balance        float64         `json:"balance,string"`
	LockedBalance  float64         `json:"locked,string"`
	DepositAddress *string         `json:"deposit_address"`
	DefaultNetwork *string         `json:"default_network"`
	Networks       []any           `json:"networks"`
	User           *models.Account `json:"user"`
}
