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
	TemporaryInstantSwapQuotation(w http.ResponseWriter, r *http.Request)

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
	mux.HandleFunc("POST /api/v1/users/{user_id}/temporary_swap_quotation", i.middlewares.AttchValidateAccessToken(i.TemporaryInstantSwapQuotation))
	mux.HandleFunc("POST /api/v1/users/{user_id}/swap_quotation", i.middlewares.AttchValidateAccessToken(i.CreateInstantSwap))
	markets := map[string]any{}
	for k := range services.Rates {
		for j := range services.Rates {
			markets[k+j] = map[string]any{
				"ticker": map[string]float64{
					"open": services.Rates[k][j],
					"buy": services.Rates[k][j],
					"sell": services.Rates[j][k],
				},
				"market": k+j,
			}
		}
	}
	mux.HandleFunc("GET /api/v1/markets/tickers/{market}", i.middlewares.AttchValidateAccessToken(func(w http.ResponseWriter, r *http.Request) {
		utils.JSON(w, 200, map[string]any{
			"data": markets[r.PathValue("market")],
		})
	}))
	mux.HandleFunc("POST /api/v1/users/{user_id}/swap_quotation/{quotation_id}/confirm", i.middlewares.AttchValidateAccessToken(i.ConfirmInstantSwap))
	mux.HandleFunc("GET /api/v1/users/{user_id}/swap_transactions/{swap_transaction_id}", i.middlewares.AttchValidateAccessToken(i.FetchInstantSwapTransaction))
	mux.HandleFunc("GET /api/v1/users/{user_id}/swap_transactions", i.middlewares.AttchValidateAccessToken(i.GetInstantSwapTransactions))
}

func (i *instantSwapHandler) CreateInstantSwap(w http.ResponseWriter, r *http.Request) {
	req := utils.Bind[requests.CreateInstantSwapRequest](r)

	res, err := i.swapService.CreateInstantSwap(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 201, res)
}

func (i *instantSwapHandler) ConfirmInstantSwap(w http.ResponseWriter, r *http.Request) {
	req := utils.Bind[requests.ConfirmInstanSwapRequest](r)

	res, err := i.swapService.ConfirmInstantSwap(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (i *instantSwapHandler) FetchInstantSwapTransaction(w http.ResponseWriter, r *http.Request) {
	req := utils.Bind[requests.FetchInstantSwapTransactionRequest](r)

	res, err := i.swapService.FetchInstantSwapTransaction(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (i *instantSwapHandler) GetInstantSwapTransactions(w http.ResponseWriter, r *http.Request) {
	req := utils.Bind[requests.GetInstantSwapTransactionsRequest](r)

	res, err := i.swapService.GetInstantSwapTransactions(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (i *instantSwapHandler) TemporaryInstantSwapQuotation(w http.ResponseWriter, r *http.Request) {
	req := utils.Bind[requests.CreateInstantSwapRequest](r)

	res, err := i.swapService.QuoteInstantSwap(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}
