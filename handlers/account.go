package handlers

import (
	"net/http"

	"github.com/2HgO/quidax-go/services"
	"github.com/2HgO/quidax-go/types/requests"
	"github.com/2HgO/quidax-go/types/responses"
	"github.com/2HgO/quidax-go/utils"
)

type AccountHandler interface {
	CreateAccount(w http.ResponseWriter, r *http.Request)
	FetchAccountDetails(w http.ResponseWriter, r *http.Request)

	ServeHttp(*http.ServeMux)
}

func NewAccountHandler(accountService services.AccountService, middlewares MiddleWareHandler) AccountHandler {
	return &accountHandler{
		handler: handler{accountService: accountService, middlewares: middlewares},
	}
}

type accountHandler struct {
	handler
}

func (a *accountHandler) ServeHttp(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/accounts", utils.Middleware(a.CreateAccount))
	mux.HandleFunc("GET /api/v1/users/{user_id}", utils.Middleware(a.FetchAccountDetails, a.middlewares.ValidateAccessToken))
}

func (a *accountHandler) CreateAccount(w http.ResponseWriter, r *http.Request) {
	req := new(requests.CreateAccountRequest)
	err := utils.Bind(r, req)
	if err != nil {
		utils.JSON(w, 500, responses.Response[string]{Data: err.Error()})
		return
	}

	res, err := a.accountService.CreateAccount(r.Context(), req)
	if err != nil {
		utils.JSON(w, 500, responses.Response[string]{Data: err.Error()})
		return
	}

	utils.JSON(w, 201, res)
	return
}

func (a *accountHandler) FetchAccountDetails(w http.ResponseWriter, r *http.Request) {
	req := &requests.FetchAccountDetailsRequest{UserID: r.PathValue("user_id")}

	res, err := a.accountService.FetchAccountDetails(r.Context(), req)
	if err != nil {
		utils.JSON(w, 500, responses.Response[string]{Data: err.Error()})
		return
	}

	utils.JSON(w, 200, res)
	return
}
