package handlers

import (
	"net/http"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/services"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/utils"
	"go.uber.org/zap"
)

type WalletHandler interface {
	FetchUserWallet(w http.ResponseWriter, r *http.Request)
	FetchUserWallets(w http.ResponseWriter, r *http.Request)

	ServeHttp(*http.ServeMux)
}

func NewWalletHandler(accountService services.AccountService, walletService services.WalletService, middlewares MiddleWareHandler, log *zap.Logger) WalletHandler {
	return &walletHandler{
		handler: handler{accountService: accountService, walletService: walletService, middlewares: middlewares, log: log},
	}
}

type walletHandler struct {
	handler
}

func (ws *walletHandler) ServeHttp(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/wallets/{user_id}", utils.Middleware(ws.FetchUserWallets, ws.middlewares.ValidateAccessToken))
	mux.HandleFunc("GET /api/v1/wallets/{user_id}/{currency}", utils.Middleware(ws.FetchUserWallet, ws.middlewares.ValidateAccessToken))
}

func (ws *walletHandler) FetchUserWallet(w http.ResponseWriter, r *http.Request) {
	req := &requests.FetchUserWalletRequest{UserID: r.PathValue("user_id"), Currency: r.PathValue("currency")}
	if err := utils.Bind(r, req); err != nil {
		errors.HandleBindError(err).Serialize(w)
		return
	}

	res, err := ws.walletService.FetchUserWallet(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (ws *walletHandler) FetchUserWallets(w http.ResponseWriter, r *http.Request) {
	req := &requests.FetchUserWalletsRequest{UserID: r.PathValue("user_id")}
	if err := utils.Bind(r, req); err != nil {
		errors.HandleBindError(err).Serialize(w)
		return
	}

	res, err := ws.walletService.FetchUserWallets(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}
