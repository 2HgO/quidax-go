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
	mux.HandleFunc("POST /api/v1/accounts", utils.Middleware(a.CreateAccount))
	mux.HandleFunc("PUT /api/v1/accounts", utils.Middleware(a.UpdateWebHookURL, a.middlewares.ValidateAccessToken))

	mux.HandleFunc("POST /api/v1/users", utils.Middleware(a.CreateSubAccount, a.middlewares.ValidateAccessToken))
	mux.HandleFunc("GET /api/v1/users", utils.Middleware(a.FetchAllSubAccounts, a.middlewares.ValidateAccessToken))
	mux.HandleFunc("PUT /api/v1/users/{user_id}", utils.Middleware(a.EditSubAccountDetails, a.middlewares.ValidateAccessToken))
	mux.HandleFunc("GET /api/v1/users/{user_id}", utils.Middleware(a.FetchAccountDetails, a.middlewares.ValidateAccessToken))
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
