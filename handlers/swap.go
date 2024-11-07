package handlers

import (
	"net/http"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/services"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/utils"
	"go.uber.org/zap"
)

type InstantSwapHandler interface {
	CreateInstantSwap(w http.ResponseWriter, r *http.Request)
	ConfirmInstantSwap(w http.ResponseWriter, r *http.Request)
	FetchInstantSwapTransaction(w http.ResponseWriter, r *http.Request)
	GetInstantSwapTransactions(w http.ResponseWriter, r *http.Request)

	ServeHttp(*http.ServeMux)
}

func NewInstantSwapHandler(accountService services.AccountService, swapService services.InstantSwapService, middlewares MiddleWareHandler, log *zap.Logger) InstantSwapHandler {
	return &instantSwapHandler{
		handler: handler{accountService: accountService, swapService: swapService, middlewares: middlewares, log: log},
	}
}

type instantSwapHandler struct {
	handler
}

func (i *instantSwapHandler) ServeHttp(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/users/{user_id}/swap_quotation", i.middlewares.AttchValidateAccessToken(i.CreateInstantSwap))
	mux.HandleFunc("POST /api/v1/users/{user_id}/swap_quotation/{quotation_id}/confirm", i.middlewares.AttchValidateAccessToken(i.ConfirmInstantSwap))
	mux.HandleFunc("GET /api/v1/users/{user_id}/swap_transactions/{swap_transaction_id}", i.middlewares.AttchValidateAccessToken(i.FetchInstantSwapTransaction))
	mux.HandleFunc("GET /api/v1/users/{user_id}/swap_transactions", i.middlewares.AttchValidateAccessToken(i.GetInstantSwapTransactions))
}

func (i *instantSwapHandler) CreateInstantSwap(w http.ResponseWriter, r *http.Request) {
	var req = &requests.CreateInstantSwapRequest{UserID: r.PathValue("user_id")}
	if err := utils.Bind(r, req); err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	res, err := i.swapService.CreateInstantSwap(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 201, res)
}

func (i *instantSwapHandler) ConfirmInstantSwap(w http.ResponseWriter, r *http.Request) {
	var req = &requests.ConfirmInstanSwapRequest{UserID: r.PathValue("user_id"), QuotationID: r.PathValue("quotation_id")}

	res, err := i.swapService.ConfirmInstantSwap(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (i *instantSwapHandler) FetchInstantSwapTransaction(w http.ResponseWriter, r *http.Request) {
	var req = &requests.FetchInstantSwapTransactionRequest{UserID: r.PathValue("user_id"), SwapTransactionID: r.PathValue("swap_transaction_id")}

	res, err := i.swapService.FetchInstantSwapTransaction(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (i *instantSwapHandler) GetInstantSwapTransactions(w http.ResponseWriter, r *http.Request) {
	var req = &requests.GetInstantSwapTransactionsRequest{UserID: r.PathValue("user_id")}

	res, err := i.swapService.GetInstantSwapTransactions(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}
