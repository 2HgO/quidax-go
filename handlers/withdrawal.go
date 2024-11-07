package handlers

import (
	"net/http"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/services"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/utils"
	"go.uber.org/zap"
)

type WithdrawalHandler interface {
	CreateWithdrawal(w http.ResponseWriter, r *http.Request)
	FetchWithdrawal(w http.ResponseWriter, r *http.Request)
	FetchWithdrawalByRef(w http.ResponseWriter, r *http.Request)
	FetchWithdrawals(w http.ResponseWriter, r *http.Request)

	ServeHttp(*http.ServeMux)
}

func NewWithdrawalHandler(accountService services.AccountService, withdrawalService services.WithdrawalService, middlewares MiddleWareHandler, log *zap.Logger) WithdrawalHandler {
	return &withdrawalHandler{
		handler: handler{accountService: accountService, withdrawalService: withdrawalService, middlewares: middlewares, log: log},
	}
}

type withdrawalHandler struct {
	handler
}

func (wd *withdrawalHandler) ServeHttp(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/users/{user_id}/withdraws", wd.middlewares.AttchValidateAccessToken(wd.CreateWithdrawal))
	mux.HandleFunc("GET /api/v1/users/{user_id}/withdraws", wd.middlewares.AttchValidateAccessToken(wd.FetchWithdrawals))
	mux.HandleFunc("GET /api/v1/users/{user_id}/withdraws/reference/{reference}", wd.middlewares.AttchValidateAccessToken(wd.FetchWithdrawalByRef))
	mux.HandleFunc("GET /api/v1/users/{user_id}/withdraws/{withdrawal_id}", wd.middlewares.AttchValidateAccessToken(wd.FetchWithdrawal))
}

func (wd *withdrawalHandler) CreateWithdrawal(w http.ResponseWriter, r *http.Request) {
	req := &requests.CreateWithdrawalRequest{UserID: r.PathValue("user_id")}
	err := utils.Bind(r, req)
	if err != nil {
		errors.HandleBindError(err).Serialize(w)
		return
	}

	res, err := wd.withdrawalService.CreateUserWithdrawal(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 201, res)
}
func (wd *withdrawalHandler) FetchWithdrawal(w http.ResponseWriter, r *http.Request) {
	req := &requests.FetchWithdrawalRequest{UserID: r.PathValue("user_id"), WithdrawalID: r.PathValue("withdrawal_id")}

	res, err := wd.withdrawalService.FetchWithdrawal(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}
func (wd *withdrawalHandler) FetchWithdrawalByRef(w http.ResponseWriter, r *http.Request) {
	req := &requests.FetchWithdrawalRequest{UserID: r.PathValue("user_id"), Reference: r.PathValue("reference")}

	res, err := wd.withdrawalService.FetchWithdrawal(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (wd *withdrawalHandler) FetchWithdrawals(w http.ResponseWriter, r *http.Request) {
	req := &requests.FetchWithdrawalsRequest{UserID: r.PathValue("user_id")}
	if err := utils.Bind(r, req); err != nil {
		errors.HandleBindError(err).Serialize(w)
		return
	}

	res, err := wd.withdrawalService.FetchWithdrawals(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}
