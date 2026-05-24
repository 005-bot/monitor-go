package parser

import (
	"github.com/go-core-fx/logger"
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"parser",
		logger.WithNamedLogger("parser"),
		fx.Provide(NewOrganizationParser),
		fx.Provide(NewOutageDetailsParser),
	)
}
