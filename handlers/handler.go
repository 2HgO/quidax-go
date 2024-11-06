package handlers

import (
	"net/http"

	"github.com/2HgO/quidax-go/services"
	"go.uber.org/zap"
)

type handler struct {
	accountService services.AccountService
	walletService  services.WalletService
	middlewares    MiddleWareHandler

	log *zap.Logger
}

type Handler interface {
	ServeHttp(*http.ServeMux)
}
