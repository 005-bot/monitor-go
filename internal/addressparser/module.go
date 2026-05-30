package addressparser

import (
	"context"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module(
		"addressparser",
		fx.Provide(New),
		fx.Invoke(func(ap *AddressParser, lc fx.Lifecycle) {
			lc.Append(fx.StopHook(
				func(_ context.Context) error {
					return ap.Close()
				},
			))
		}),
	)
}
