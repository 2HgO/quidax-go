package main

import (
	"net/http"

	"github.com/2HgO/quidax-go/db"
	"github.com/2HgO/quidax-go/handlers"
	"github.com/2HgO/quidax-go/services"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

// "net/http"

// "go.uber.org/fx"

func main() {
	fx.New(
		fx.Provide(
			NewHttpServer,
			NewServeMux,
			fx.Annotate(
				handlers.NewAccountHandler,
				fx.As(new(handlers.Handler)),
			),
			handlers.NewMiddlewareHandler,
			services.NewAccountService,
			db.GetDataDBConnection,
			db.GetTxDBConnection,
			zap.NewProduction,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}
