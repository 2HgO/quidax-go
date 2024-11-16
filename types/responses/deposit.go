package responses

import (
	"time"

	"github.com/2HgO/quidax-go/models"
)

type DepositResponseData struct {
	ID        string                  `json:"id"`
	Type      string                  `json:"type"`
	User      *models.Account         `json:"user"`
	Wallet    *UserWalletResponseData `json:"wallet"`
	Currency  string                  `json:"currency"`
	Amount    float64                 `json:"amount,string"`
	CreatedAt time.Time               `json:"created_at"`
	DoneAt    time.Time               `json:"done_at"`
	Fee       float64                 `json:"fee,string"`
	Status    string                  `json:"status"`
	TxID      string                  `json:"txid"`
}
