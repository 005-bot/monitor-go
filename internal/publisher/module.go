package publisher

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"publisher",
		logger.WithNamedLogger("publisher"),
		fx.Provide(NewMetrics, fx.Private),
		fx.Provide(NewService),
	)
}
