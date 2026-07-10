package parser

import (
	"context"
	"fmt"
	"time"

	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/005-bot/monitor-go/internal/parser/organization"
	"github.com/005-bot/monitor-go/internal/parser/outage"
	"go.uber.org/zap"
)

type Service struct {
	orgParser    *organization.Parser
	outageParser *outage.Parser
	logger       *zap.Logger
	metrics      *Metrics
}

func NewService(
	orgParser *organization.Parser,
	outageParser *outage.Parser,
	metrics *Metrics,
	logger *zap.Logger,
) *Service {
	return &Service{
		orgParser:    orgParser,
		outageParser: outageParser,
		logger:       logger,
		metrics:      metrics,
	}
}

func (s *Service) Parse(ctx context.Context, record domain.Record) (domain.ParsedRecord, error) {
	defer func(start time.Time) {
		s.metrics.ObserveDuration("parse", time.Since(start).Seconds())
	}(time.Now())

	orgInfo := s.orgParser.Parse(record.Organization)
	if orgInfo == nil {
		return domain.ParsedRecord{}, fmt.Errorf("%w: %q", ErrParseOrganization, record.Organization)
	}

	details, err := s.outageParser.Parse(ctx, record.Address)
	if err != nil {
		return domain.ParsedRecord{}, fmt.Errorf("%w: %w", ErrParseOutage, err)
	}
	if details == nil {
		details = &domain.OutageDetails{} //nolint:exhaustruct // zero value is acceptable for empty details
	}

	return domain.ParsedRecord{
		Area:         record.Area,
		Organization: *orgInfo,
		Details:      *details,
		Dates:        record.Dates,
	}, nil
}

func (s *Service) ParseBatch(ctx context.Context, records []domain.Record) ([]domain.ParsedRecord, error) {
	defer func(start time.Time) {
		s.metrics.ObserveDuration("parse_batch", time.Since(start).Seconds())
	}(time.Now())
	s.metrics.IncOperations("parse_batch")

	parsed := make([]domain.ParsedRecord, 0, len(records))

	for _, record := range records {
		result, err := s.Parse(ctx, record)
		if err != nil {
			s.logger.Warn("skipping record parse failure", zap.Error(err), zap.String("area", record.Area))
			continue
		}
		parsed = append(parsed, result)
	}

	return parsed, nil
}
