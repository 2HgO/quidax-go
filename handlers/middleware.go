package handlers

import (
	"context"
	"net/http"
	"strings"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/services"
	"github.com/2HgO/quidax-go/utils"
	"go.uber.org/zap"
)

type MiddleWareHandler interface {
	AttachValidateAccessToken(http.HandlerFunc) http.HandlerFunc
}

type middlewareHandler struct {
	accountService services.AccountService
	log            *zap.Logger
}

func NewMiddlewareHandler(account services.AccountService, log *zap.Logger) MiddleWareHandler {
	return &middlewareHandler{accountService: account, log: log}
}

func (m *middlewareHandler) AttachValidateAccessToken(h http.HandlerFunc) http.HandlerFunc {
	return utils.Middleware(h, m.validateAccessToken)
}

func (m *middlewareHandler) validateAccessToken(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")
		if token == "" {
			errors.NewInvalidTokenError().Serialize(w)
			return
		}

		res, err := m.accountService.GetAccountByAccessToken(r.Context(), token)
		if err != nil {
			errors.AsAppError(err).Serialize(w)
			return
		}

		h.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "user", res)))
	}
}

func RecoveryMW(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				switch rErr := err.(type) {
				case error:
					errors.AsAppError(rErr).Serialize(w)
				default:
					errors.NewUnknownError(err).Serialize(w)
				}
			}
		}()

		h.ServeHTTP(w, r)
	})
}
