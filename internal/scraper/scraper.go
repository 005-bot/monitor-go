package scraper

import (
	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/005-bot/monitor-go/internal/storage"
	"go.uber.org/zap"
)

type Scraper struct {
	cfg     Config
	storage *storage.Storage
	logger  *zap.Logger
}

func New(cfg Config, storage *storage.Storage, logger *zap.Logger) *Scraper {
	return &Scraper{
		cfg:     cfg,
		storage: storage,
		logger:  logger,
	}
}

func (s *Scraper) Run() ([]domain.Record, error) {
	return nil, nil
}
