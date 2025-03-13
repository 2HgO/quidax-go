package handlers

import (
	"net/http"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/services"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/utils"
	"go.uber.org/zap"
)

type DepositHandler interface {
	DepositAmount(http.ResponseWriter, *http.Request)
	FetchDeposit(http.ResponseWriter, *http.Request)
	FetchDeposits(http.ResponseWriter, *http.Request)

	Handler
}

func NewDepositHandler(depositService services.DepositService, middlewares MiddleWareHandler, log *zap.Logger) DepositHandler {
	return &depositHandler{
		handler: handler{depositService: depositService, middlewares: middlewares, log: log},
	}
}

type depositHandler struct {
	handler
}

func (d *depositHandler) ServeHttp(mux *http.ServeMux) {
	mux.Handle("POST /api/v1/users/{user_id}/deposits/{currency}", d.middlewares.AttachValidateAccessToken(d.DepositAmount))
	mux.Handle("GET /api/v1/users/{user_id}/deposits", d.middlewares.AttachValidateAccessToken(d.FetchDeposits))
	mux.Handle("GET /api/v1/users/{user_id}/deposits/currency/{currency}", d.middlewares.AttachValidateAccessToken(d.FetchDeposits))
	mux.Handle("GET /api/v1/users/{user_id}/deposits/{transaction_id}", d.middlewares.AttachValidateAccessToken(d.FetchDeposit))
}

func (d *depositHandler) DepositAmount(w http.ResponseWriter, r *http.Request) {
	req := utils.Bind[requests.DepositAmountRequest](r)

	res, err := d.depositService.CreateDeposit(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 201, res)
}

func (d *depositHandler) FetchDeposit(w http.ResponseWriter, r *http.Request) {
	req := utils.Bind[requests.FetchDepositRequest](r)

	res, err := d.depositService.FetchDeposit(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (d *depositHandler) FetchDeposits(w http.ResponseWriter, r *http.Request) {
	req := utils.Bind[requests.FetchDepositsRequest](r)

	res, err := d.depositService.FetchDeposits(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}
