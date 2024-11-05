package responses

import (
	"time"

	"github.com/2HgO/quidax-go/models"
)

type InstantSwapQuotationResponseData struct {
	ID             string          `json:"id"`
	FromCurrency   string          `json:"from_currency"`
	ToCurrency     string          `json:"to_currency"`
	QuotedPrice    float64         `json:"quoted_price,string"`
	QuotedCurrency string          `json:"quoted_currency"`
	FromAmount     float64         `json:"from_amount,string"`
	ToAmount       float64         `json:"to_amount,string"`
	Confirmed      bool            `json:"confirmed"`
	ExpiresAt      time.Time       `json:"expires_at"`
	CreatedAt      time.Time       `json:"created_at"`
	User           *models.Account `json:"user"`
}
