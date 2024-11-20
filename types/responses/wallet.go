package responses

import (
	"time"

	"github.com/2HgO/quidax-go/models"
)

type UserWalletResponseData struct {
	ID                string          `json:"id"`
	Name              string          `json:"name"`
	Currency          string          `json:"currency"`
	Balance           float64         `json:"balance,string"`
	LockedBalance     float64         `json:"locked,string"`
	DepositAddress    *string         `json:"deposit_address"`
	DefaultNetwork    *string         `json:"default_network"`
	ConvertedBalance  float64         `json:"converted_balance,string"`
	Networks          []any           `json:"networks"`
	User              *models.Account `json:"user"`
	CreatedAt         time.Time       `json:"created_at"`
	UpdatedAt         time.Time       `json:"updated_at"`
	ReferenceCurrency string          `json:"reference_currency"`
	IsCrypto          bool            `json:"is_crypto"`
}
