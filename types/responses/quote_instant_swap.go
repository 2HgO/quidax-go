package responses

type QuoteInstantSwapResponseData struct {
	FromCurrency   string  `json:"from_currency"`
	ToCurrency     string  `json:"to_currency"`
	QuotedPrice    float64 `json:"quoted_price,string"`
	QuotedCurrency string  `json:"quoted_currency"`
	FromAmount     float64 `json:"from_amount,string"`
	ToAmount       float64 `json:"to_amount,string"`
}
