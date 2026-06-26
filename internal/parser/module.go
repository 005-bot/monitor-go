package parser

import (
	"github.com/005-bot/monitor-go/internal/parser/address"
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
			organization.NewParser,
			outage.NewParser,
			NewService,
		),
	)
}
