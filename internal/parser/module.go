package parser

import (
	"github.com/005-bot/address-parser-go"
	"github.com/005-bot/monitor-go/internal/parser/organization"
	"github.com/005-bot/monitor-go/internal/parser/outage"
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"parser",
		logger.WithNamedLogger("parser"),
		address.Module(),
		fx.Provide(
			NewMetrics,
			fx.Private,
		),
		fx.Provide(
			organization.NewParser,
			outage.NewParser,
			NewService,
		),
	)
}
