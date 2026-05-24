package publisher

import (
	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Publisher struct {
	rdb     *redis.Client
	channel string
	logger  *zap.Logger
}

func New(cfg Config, rdb *redis.Client, logger *zap.Logger) *Publisher {
	return &Publisher{
		rdb:     rdb,
		channel: cfg.Prefix + ":outages",
		logger:  logger,
	}
}

func (p *Publisher) Publish(outage domain.ParsedRecord) error {
	return nil
}
