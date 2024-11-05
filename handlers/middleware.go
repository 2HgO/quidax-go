package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/2HgO/quidax-go/services"
	"github.com/2HgO/quidax-go/utils"
)

type MiddleWareHandler interface {
	ValidateAccessToken(http.HandlerFunc) http.HandlerFunc
}

type middlewareHandler struct {
	accountService services.AccountService
}

func NewMiddlewareHandler(account services.AccountService) MiddleWareHandler {
	return &middlewareHandler{accountService: account}
}

func (m *middlewareHandler) ValidateAccessToken(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
		if token == "" {
			utils.JSON(w, 401, map[string]any{"data": "invalid token provided"})
			return
		}

		res, err := m.accountService.GetAccountByAccessToken(r.Context(), token)
		if err != nil {
			utils.JSON(w, 500, map[string]any{"data": err.Error()})
			return
		}

		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "user", res)))
	}
}
