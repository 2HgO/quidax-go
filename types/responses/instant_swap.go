package responses

import (
	"time"

	"github.com/2HgO/quidax-go/models"
)

type InstantSwapResponseData struct {
	ID             string                            `json:"id"`
	FromCurrency   string                            `json:"from_currency"`
	ToCurrency     string                            `json:"to_currency"`
	FromAmount     float64                           `json:"from_amount,string"`
	ReceivedAmount float64                           `json:"received_amount,string"`
	ExecutionPrice string                            `json:"execution_price"`
	Status         string                            `json:"status"`
	CreatedAt      time.Time                         `json:"created_at"`
	UpdatedAt      time.Time                         `json:"updated_at"`
	SwapQuotation  *InstantSwapQuotationResponseData `json:"swap_quotation"`
	User           *models.Account                   `json:"user"`
}
