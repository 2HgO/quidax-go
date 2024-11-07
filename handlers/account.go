package handlers

import (
	"net/http"

	"github.com/2HgO/quidax-go/errors"
	"github.com/2HgO/quidax-go/services"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/utils"
	"go.uber.org/zap"
)

type AccountHandler interface {
	CreateAccount(w http.ResponseWriter, r *http.Request)
	UpdateWebHookURL(w http.ResponseWriter, r *http.Request)

	FetchAccountDetails(w http.ResponseWriter, r *http.Request)
	CreateSubAccount(w http.ResponseWriter, r *http.Request)
	EditSubAccountDetails(w http.ResponseWriter, r *http.Request)
	FetchAllSubAccounts(w http.ResponseWriter, r *http.Request)

	ServeHttp(*http.ServeMux)
}

func NewAccountHandler(accountService services.AccountService, middlewares MiddleWareHandler, log *zap.Logger) AccountHandler {
	return &accountHandler{
		handler: handler{accountService: accountService, middlewares: middlewares, log: log},
	}
}

type accountHandler struct {
	handler
}

func (a *accountHandler) ServeHttp(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/accounts", a.CreateAccount)

	mux.HandleFunc("PUT /api/v1/accounts", a.middlewares.AttchValidateAccessToken(a.UpdateWebHookURL))

	mux.HandleFunc("POST /api/v1/users", a.middlewares.AttchValidateAccessToken(a.CreateSubAccount))
	mux.HandleFunc("GET /api/v1/users", a.middlewares.AttchValidateAccessToken(a.FetchAllSubAccounts))
	mux.HandleFunc("PUT /api/v1/users/{user_id}", a.middlewares.AttchValidateAccessToken(a.EditSubAccountDetails))
	mux.HandleFunc("GET /api/v1/users/{user_id}", a.middlewares.AttchValidateAccessToken(a.FetchAccountDetails))
}

func (a *accountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	req := new(requests.CreateAccountRequest)
	err := utils.Bind(r, req)
	if err != nil {
		errors.HandleBindError(err).Serialize(w)
		return
	}

	res, err := a.accountService.CreateAccount(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 201, res)
}

func (a *accountHandler) UpdateWebHookURL(w http.ResponseWriter, r *http.Request) {
	req := new(requests.UpdateWebhookURLRequest)
	err := utils.Bind(r, req)
	if err != nil {
		errors.HandleBindError(err).Serialize(w)
		return
	}

	err = a.accountService.UpdateWebHookURL(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	w.WriteHeader(204)
	w.Write(nil)
}

func (a *accountHandler) FetchAccountDetails(w http.ResponseWriter, r *http.Request) {
	req := &requests.FetchAccountDetailsRequest{UserID: r.PathValue("user_id")}

	res, err := a.accountService.FetchAccountDetails(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (a *accountHandler) CreateSubAccount(w http.ResponseWriter, r *http.Request) {
	req := new(requests.CreateSubAccountRequest)
	err := utils.Bind(r, req)
	if err != nil {
		errors.HandleBindError(err).Serialize(w)
		return
	}

	res, err := a.accountService.CreateSubAccount(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 201, res)
}

func (a *accountHandler) EditSubAccountDetails(w http.ResponseWriter, r *http.Request) {
	req := &requests.EditSubAccountDetailsRequest{UserID: r.PathValue("user_id")}
	err := utils.Bind(r, req)
	if err != nil {
		errors.HandleBindError(err).Serialize(w)
		return
	}

	res, err := a.accountService.EditSubAccountDetails(r.Context(), req)
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}

func (a *accountHandler) FetchAllSubAccounts(w http.ResponseWriter, r *http.Request) {
	res, err := a.accountService.FetchAllSubAccounts(r.Context(), &requests.FetchAllSubAccountsRequest{})
	if err != nil {
		errors.AsAppError(err).Serialize(w)
		return
	}

	utils.JSON(w, 200, res)
}
