package config

import (
	"context"

	"github.com/005-bot/address-parser-go"
	"github.com/005-bot/monitor-go/internal/publisher"
	"github.com/005-bot/monitor-go/internal/scheduler"
	"github.com/005-bot/monitor-go/internal/scraper"
	"github.com/005-bot/monitor-go/internal/storage"
	"github.com/go-core-fx/fiberfx"
	"github.com/go-core-fx/fiberfx/openapi"
	"github.com/go-core-fx/redisfx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Module() fx.Option {
	return fx.Module(
		"config",
		fx.Provide(New, fx.Private),
		fx.Provide(
			func(cfg Config) fiberfx.Config {
				return fiberfx.Config{
					Address:     cfg.HTTP.Address,
					ProxyHeader: cfg.HTTP.ProxyHeader,
					Proxies:     cfg.HTTP.Proxies,
				}
			},
			func(cfg Config) openapi.Config {
				return openapi.Config{
					Enabled:    cfg.HTTP.OpenAPI.Enabled,
					PublicHost: cfg.HTTP.OpenAPI.PublicHost,
					PublicPath: cfg.HTTP.OpenAPI.PublicPath,
				}
			},
			func(cfg Config) redisfx.Config {
				return redisfx.Config{
					URL: cfg.Redis.URL,
				}
			},
			func(cfg Config) storage.Config {
				return storage.Config{
					Prefix:  cfg.Storage.Prefix,
					TTLDays: cfg.Storage.TTLDays,
				}
			},
			func(cfg Config) scraper.Config {
				return scraper.Config{
					URL:      cfg.Scraper.URL,
					Interval: cfg.Scraper.Interval,
				}
			},
			func(cfg Config) address.Config {
				return address.Config{
					DBPath: cfg.Parser.AddressDBPath,
				}
			},
			func(cfg Config) publisher.Config {
				return publisher.Config{
					Prefix: cfg.Publisher.Prefix,
				}
			},
			func(cfg Config) scheduler.Config {
				return scheduler.Config{
					Interval:     cfg.Scraper.Interval,
					CycleTimeout: cfg.Scraper.Timeout,
				}
			},
		),
		fx.Invoke(func(lc fx.Lifecycle, cfg Config, logger *zap.Logger) {
			lc.Append(fx.StartHook(func(_ context.Context) error {
				if err := cfg.Validate(); err != nil {
					logger.Error("invalid configuration", zap.Error(err))

					return err
				}

				return nil
			}))
		}),
	)
}
