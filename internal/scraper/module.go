package scraper

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"scraper",
		logger.WithNamedLogger("scraper"),
		fx.Provide(NewMetrics, fx.Private),
		fx.Provide(NewService),
	)
}
