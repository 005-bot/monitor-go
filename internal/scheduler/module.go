package scheduler

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"scheduler",
		logger.WithNamedLogger("scheduler"),
		fx.Provide(NewMetrics, fx.Private),
		fx.Provide(NewService),
		fx.Invoke(func(lc fx.Lifecycle, svc *Service) {
			lc.Append(fx.Hook{
				OnStart: svc.Start,
				OnStop:  svc.Stop,
			})
		}),
	)
}
