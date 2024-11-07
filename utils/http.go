package utils

import (
	"encoding/json"
	"net/http"
)

type MW func(http.HandlerFunc) http.HandlerFunc

func JSON(w http.ResponseWriter, code int, data any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}

func Middleware(final http.HandlerFunc, h ...MW) http.HandlerFunc {
	for i := len(h) - 1; i >= 0; i-- {
		final = h[i](final)
	}
	return final
}
