package handlers

import (
	"github.com/005-bot/monitor-go/internal/server/handlers/monitor"
	"github.com/005-bot/monitor-go/internal/server/handlers/parser"
	"github.com/005-bot/monitor-go/internal/server/handlers/scraper"
	"github.com/005-bot/monitor-go/internal/server/handlers/storage"
	"github.com/go-core-fx/fiberfx/handler"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"handlers",
		logger.WithNamedLogger("handlers"),
		fx.Provide(
			fx.Annotate(
				monitor.NewHandler,
				fx.As(new(handler.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			fx.Annotate(
				storage.NewHandler,
				fx.As(new(handler.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			fx.Annotate(
				scraper.NewHandler,
				fx.As(new(handler.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
			fx.Annotate(
				parser.NewHandler,
				fx.As(new(handler.Handler)),
				fx.ResultTags(`group:"handlers"`),
			),
		),
	)
}
