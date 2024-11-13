package handlers

import (
	"net/http"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/services"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/types/responses"
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
	mux.HandleFunc("GET /api/v1/users/{user_id}/wallets", ws.middlewares.AttchValidateAccessToken(ws.FetchUserWallets))
	mux.HandleFunc("GET /api/v1/users/{user_id}/wallets/{currency}", ws.middlewares.AttchValidateAccessToken(ws.FetchUserWallet))
	mux.HandleFunc("GET /api/v1/users/{user_id}/wallets/{currency}/address", ws.middlewares.AttchValidateAccessToken(func(w http.ResponseWriter, r *http.Request) {
		walletRes, err := ws.walletService.FetchUserWallet(r.Context(), &requests.FetchUserWalletRequest{UserID: r.PathValue("user_id"), Currency: r.PathValue("currency")})
		if err != nil {
			errors.AsAppError(err).Serialize(w)
			return
		}
		wallet := walletRes.Data
		utils.JSON(w, 200, responses.Response[any]{
			Status: "successful",
			Data: map[string]any{
				"id":              wallet.ID,
				"reference":       wallet.ID,
				"currency":        wallet.Currency,
				"address":         "",
				"destination_tag": "deposit_not_supported",
				"total_payments":  "0",
				"network":         "",
			},
		})
	}))
}

func (ws *walletHandler) FetchUserWallet(w http.ResponseWriter, r *http.Request) {
	req := utils.Bind[requests.FetchUserWalletRequest](r)

	res, err := ws.walletService.FetchUserWallet(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (ws *walletHandler) FetchUserWallets(w http.ResponseWriter, r *http.Request) {
	req := utils.Bind[requests.FetchUserWalletsRequest](r)

	res, err := ws.walletService.FetchUserWallets(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}
