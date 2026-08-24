package publisher

import (
	"context"
	"encoding/json"
	"fmt"

	apidev "github.com/005-bot/apis-go"
	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type Service struct {
	rdb     *redis.Client
	channel string
	logger  *zap.Logger
	metrics *Metrics
}

func NewService(rdb *redis.Client, cfg Config, metrics *Metrics, logger *zap.Logger) *Service {
	return &Service{
		rdb:     rdb,
		channel: cfg.Prefix + ":outages",
		logger:  logger,
		metrics: metrics,
	}
}

func (s *Service) Publish(ctx context.Context, record domain.ParsedRecord) error {
	s.metrics.IncPublishes()
	defer s.metrics.ObserveDuration()()

	msg := apidev.Outage{
		Area:             record.Area,
		OrganizationInfo: record.Organization,
		Details:          record.Details,
		Period:           record.Dates,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		s.metrics.IncError("marshal")
		return fmt.Errorf("marshal outage: %w", err)
	}

	res := s.rdb.Publish(ctx, s.channel, string(data))
	if pubErr := res.Err(); pubErr != nil {
		s.metrics.IncError("publish")
		return fmt.Errorf("redis publish: %w", pubErr)
	}

	if res.Val() == 0 {
		s.metrics.IncError("no_subscribers")
		return fmt.Errorf("redis publish: channel %s: %w", s.channel, ErrNoSubscribers)
	}

	s.logger.Info("published outage", zap.String("area", record.Area))

	return nil
}
