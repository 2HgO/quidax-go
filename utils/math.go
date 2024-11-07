package utils

import "math"

func ApproximateAmount(currency string, amount float64) float64 {
	switch currency {
	case "sol":
		return math.Floor(amount*1000000) / 1000000
	case "btc":
		return math.Floor(amount*100000000) / 100000000
	case "bnb":
		return math.Floor(amount*100000) / 100000
	case "eth":
		return math.Floor(amount*1000000) / 1000000
	default:
		return math.Floor(amount*100) / 100
	}
}
