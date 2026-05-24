package scheduler

import (
	"context"

	"github.com/005-bot/monitor-go/internal/parser"
	"github.com/005-bot/monitor-go/internal/publisher"
	"github.com/005-bot/monitor-go/internal/scraper"
	"github.com/005-bot/monitor-go/internal/storage"
	"go.uber.org/zap"
)

type Scheduler struct {
	scraper      *scraper.Scraper
	storage      *storage.Storage
	publisher    *publisher.Publisher
	orgParser    *parser.OrganizationParser
	outageParser *parser.OutageDetailsParser
	cfg          Config
	logger       *zap.Logger
}

func New(
	scraper *scraper.Scraper,
	storage *storage.Storage,
	publisher *publisher.Publisher,
	orgParser *parser.OrganizationParser,
	outageParser *parser.OutageDetailsParser,
	cfg Config,
	logger *zap.Logger,
) *Scheduler {
	return &Scheduler{
		scraper:      scraper,
		storage:      storage,
		publisher:    publisher,
		orgParser:    orgParser,
		outageParser: outageParser,
		cfg:          cfg,
		logger:       logger,
	}
}

func (s *Scheduler) Start(ctx context.Context) error {
	return nil
}

func (s *Scheduler) Stop(ctx context.Context) error {
	return nil
}
