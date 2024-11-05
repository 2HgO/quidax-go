package handlers

import (
	"net/http"

	"github.com/2HgO/quidax-go/services"
)

type handler struct {
	accountService services.AccountService
	middlewares MiddleWareHandler
}

type Handler interface {
	ServeHttp(*http.ServeMux)
}
