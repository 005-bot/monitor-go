package storage

import (
	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Storage struct {
	rdb    *redis.Client
	prefix string
	ttl    int
	logger *zap.Logger
}

func New(cfg Config, rdb *redis.Client, logger *zap.Logger) *Storage {
	const secondsPerDay = 24 * 60 * 60
	return &Storage{
		rdb:    rdb,
		prefix: cfg.Prefix,
		ttl:    cfg.TTLDays * secondsPerDay,
		logger: logger,
	}
}

func (s *Storage) IsETagChanged(etag string) (bool, error) {
	return false, nil
}

func (s *Storage) Diff(records []domain.ParsedRecord) ([]domain.ParsedRecord, error) {
	return nil, nil
}

func (s *Storage) Commit(records []domain.ParsedRecord) error {
	return nil
}
