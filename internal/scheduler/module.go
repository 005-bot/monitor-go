package scheduler

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"scheduler",
		logger.WithNamedLogger("scheduler"),
		fx.Provide(New),
		fx.Invoke(func(lc fx.Lifecycle, s *Scheduler) {
			lc.Append(fx.Hook{
				OnStart: s.Start,
				OnStop:  s.Stop,
			})
		}),
	)
}
