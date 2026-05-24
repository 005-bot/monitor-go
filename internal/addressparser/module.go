package addressparser

import (
	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"addressparser",
		fx.Provide(New),
	)
}
