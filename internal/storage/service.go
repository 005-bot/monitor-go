//nolint:gosec // MD5 used for dedup hashing, not security
package storage

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

const ttlSecondsPerDay = 86400

type Service struct {
	rdb          *redis.Client
	logger       *zap.Logger
	metrics      *Metrics
	keyETag      string
	keyRecordsV2 string
	ttl          time.Duration
}

func NewService(rdb *redis.Client, cfg Config, metrics *Metrics, logger *zap.Logger) *Service {
	return &Service{
		rdb:          rdb,
		logger:       logger,
		metrics:      metrics,
		keyETag:      cfg.Prefix + ":etag",
		keyRecordsV2: cfg.Prefix + ":records_v2",
		ttl:          time.Duration(cfg.TTLDays) * ttlSecondsPerDay * time.Second,
	}
}

func (s *Service) GetEtag(ctx context.Context) (string, error) {
	defer s.metrics.ObserveDuration("etag_get")()
	s.metrics.opsTotal.WithLabelValues("etag_get").Inc()

	val, err := s.rdb.Get(ctx, s.keyETag).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("etag get: %w", err)
	}

	return val, nil
}

func (s *Service) IsEtagChanged(ctx context.Context, etag string) (bool, error) {
	defer s.metrics.ObserveDuration("etag_check")()
	s.metrics.opsTotal.WithLabelValues("etag_check").Inc()

	redisArgs := redis.SetArgs{Get: true} //nolint:exhaustruct // only setting Get
	old, err := s.rdb.SetArgs(ctx, s.keyETag, etag, redisArgs).Result()
	if errors.Is(err, redis.Nil) {
		return true, ErrNotFound
	}
	if err != nil {
		return false, fmt.Errorf("etag set: %w", err)
	}

	return old != etag, nil
}

func (s *Service) Diff(ctx context.Context, records []domain.ParsedRecord) ([]domain.ParsedRecord, error) {
	defer s.metrics.ObserveDuration("diff")()
	s.metrics.opsTotal.WithLabelValues("diff").Inc()

	if len(records) == 0 {
		return nil, nil
	}

	hashes := make(map[string]domain.ParsedRecord, len(records))
	for _, r := range records {
		hashes[s.HashV2(r)] = r
	}

	existing, err := s.rdb.HKeys(ctx, s.keyRecordsV2).Result()
	if err != nil {
		return nil, fmt.Errorf("diff hkeys: %w", err)
	}

	existingSet := make(map[string]struct{}, len(existing))
	for _, h := range existing {
		existingSet[h] = struct{}{}
	}

	changed := make([]domain.ParsedRecord, 0, len(records))
	for h, r := range hashes {
		if _, ok := existingSet[h]; !ok {
			changed = append(changed, r)
		}
	}

	s.logger.Info("diff completed", zap.Int("total", len(records)), zap.Int("changed", len(changed)))

	return changed, nil
}

func (s *Service) Commit(ctx context.Context, records []domain.ParsedRecord) error {
	defer s.metrics.ObserveDuration("commit")()
	s.metrics.opsTotal.WithLabelValues("commit").Inc()

	if len(records) == 0 {
		return nil
	}

	pipe := s.rdb.Pipeline()

	recordsV2 := make(map[string]string, len(records))
	fields := make([]string, 0, len(records))
	for _, r := range records {
		h := s.HashV2(r)
		data, err := json.Marshal(r)
		if err != nil {
			return fmt.Errorf("commit marshal record %s: %w", h, err)
		}
		recordsV2[h] = string(data)
		fields = append(fields, h)
	}

	pipe.HMSet(ctx, s.keyRecordsV2, recordsV2)
	pipe.HExpire(ctx, s.keyRecordsV2, s.ttl, fields...)

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("commit pipeline: %w", err)
	}

	s.logger.Info("committed records", zap.Int("count", len(records)))

	return nil
}

func (s *Service) HashV2(record domain.ParsedRecord) string {
	streetNames := make([]string, len(record.Details.Streets))
	for i, st := range record.Details.Streets {
		streetNames[i] = st.Name
	}

	var reasonType string
	if record.Details.Reason != nil {
		reasonType = record.Details.Reason.Type
	}

	firstDate := ""
	if len(record.Dates) > 0 {
		firstDate = record.Dates[0].Format("2006-01-02 15:04")
	}

	hash := md5.Sum([]byte(strings.Join(streetNames, ",") + reasonType + firstDate)) //nolint:gosec // MD5 for dedup
	return hex.EncodeToString(hash[:])
}
