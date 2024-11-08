package main

import (
	"net/http"

	"github.com/2HgO/quidax-go/db"
	"github.com/2HgO/quidax-go/handlers"
	"github.com/2HgO/quidax-go/services"
	"github.com/madflojo/tasks"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(
			NewHttpServer,
			fx.Annotate(
				NewServeMux,
				fx.ParamTags(`group:"handlers"`),
			),
			fx.Annotate(
				handlers.NewAccountHandler,
				fx.As(new(handlers.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			fx.Annotate(
				handlers.NewWalletHandler,
				fx.As(new(handlers.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			fx.Annotate(
				handlers.NewInstantSwapHandler,
				fx.As(new(handlers.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			fx.Annotate(
				handlers.NewWithdrawalHandler,
				fx.As(new(handlers.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			handlers.NewMiddlewareHandler,
			services.NewInstantSwapService,
			services.NewWithdrawalService,
			services.NewWalletService,
			services.NewWebhookService,
			services.NewSchedulerService,
			services.NewAccountService,
			db.GetDataDBConnection,
			db.GetTxDBConnection,
			tasks.New,
			zap.NewProduction,
		),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}
