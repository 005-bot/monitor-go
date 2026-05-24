package internal

import (
	"context"

	"github.com/005-bot/monitor-go/internal/addressparser"
	"github.com/005-bot/monitor-go/internal/config"
	"github.com/005-bot/monitor-go/internal/parser"
	"github.com/005-bot/monitor-go/internal/publisher"
	"github.com/005-bot/monitor-go/internal/scheduler"
	"github.com/005-bot/monitor-go/internal/scraper"
	"github.com/005-bot/monitor-go/internal/server"
	"github.com/005-bot/monitor-go/internal/storage"
	"github.com/go-core-fx/fiberfx"
	"github.com/go-core-fx/healthfx"
	"github.com/go-core-fx/logger"
	"github.com/go-core-fx/redisfx"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

func Run(version healthfx.Version) {
	fx.New(
		// CORE MODULES
		logger.Module(),
		logger.WithFxDefaultLogger(),
		// badgerfx.Module(),
		// bunfx.Module(),
		// cachefx.Module(),
		fiberfx.Module(),
		// gocqlfx.Module(),
		// gocqlxfx.Module(),
		// sqlfx.Module(),
		// goosefx.Module(),
		// gormfx.Module(),
		healthfx.Module(),
		// openrouterfx.Module(),
		redisfx.Module(),
		// sqlxfx.Module(),
		// telegofx.Module(true),
		// validatorfx.Module(),
		// watermillfx.Module(),
		//
		// APP MODULES
		config.Module(),
		server.Module(),
		//
		// BUSINESS MODULES
		fx.Supply(version),
		addressparser.Module(),
		storage.Module(),
		publisher.Module(),
		scraper.Module(),
		parser.Module(),
		scheduler.Module(),

		fx.Invoke(func(lc fx.Lifecycle, logger *zap.Logger) {
			lc.Append(fx.Hook{
				OnStart: func(_ context.Context) error {
					logger.Info("app started")
					return nil
				},
				OnStop: func(_ context.Context) error {
					logger.Info("app stopped")
					return nil
				},
			})
		}),
	).Run()
}
