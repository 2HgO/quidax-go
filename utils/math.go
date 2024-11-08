package utils

import (
	"math"

	tdb_types "github.com/tigerbeetle/tigerbeetle-go/pkg/types"
)

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

func ToAmount(val float64) tdb_types.Uint128 {
	return tdb_types.ToUint128(uint64(math.Floor(val * 1e9)))
}

func FromAmount(amount tdb_types.Uint128) float64 {
	val := amount.BigInt()
	return float64(val.Uint64()) * 1e-9
}
