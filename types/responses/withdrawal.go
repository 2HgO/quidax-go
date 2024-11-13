package responses

import (
	"time"

	"github.com/2HgO/quidax-go/models"
)

type WithdrawalResponseData struct {
	ID              string                  `json:"id"`
	Reference       string                  `json:"reference"`
	Type            models.RecipientType    `json:"type"`
	Currency        string                  `json:"currency"`
	Amount          float64                 `json:"amount,string"`
	Fee             float64                 `json:"fee,string"`
	Total           float64                 `json:"total,string"`
	TransactionID   string                  `json:"txid"`
	TransactionNote string                  `json:"transaction_note"`
	Narration       string                  `json:"narration"`
	Status          models.WithdrawalStatus `json:"status"`
	Reason          *string                 `json:"reason"`
	CreatedAt       time.Time               `json:"created_at"`
	DoneAt          time.Time               `json:"done_at"`
	Recipient       *models.Recipient       `json:"recipient"`
	Wallet          *UserWalletResponseData `json:"wallet"`
	User            *models.Account         `json:"user"`
}

type WithdrawalResponseDataRecipient struct {
	Type    string                                  `json:"type"`
	Details *WithdrawalResponseDataRecipientDetails `json:"details"`
}

type WithdrawalResponseDataRecipientDetails struct {
	Name           *string `json:"name"`
	DestinationTag *string `json:"destination_tag"`
	Address        *string `json:"address"`
}
