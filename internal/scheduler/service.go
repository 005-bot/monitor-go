package scheduler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/005-bot/monitor-go/internal/parser"
	"github.com/005-bot/monitor-go/internal/publisher"
	"github.com/005-bot/monitor-go/internal/scraper"
	"github.com/005-bot/monitor-go/internal/storage"
	"go.uber.org/zap"
)

type Status struct {
	Running   bool      `json:"running"`
	LastRunAt time.Time `json:"last_run_at"`
}

type Service struct {
	cfg Config

	scraper   *scraper.Service
	storage   *storage.Service
	publisher *publisher.Service
	parser    *parser.Service

	wg        sync.WaitGroup
	mu        sync.RWMutex
	runMu     sync.Mutex
	running   bool
	lastRunAt time.Time
	cancel    context.CancelFunc

	logger  *zap.Logger
	metrics *Metrics
}

func NewService(
	cfg Config,

	scraperSvc *scraper.Service,
	storageSvc *storage.Service,
	publisherSvc *publisher.Service,
	parserSvc *parser.Service,
	metrics *Metrics,
	logger *zap.Logger,
) (*Service, error) {
	if cfg.Interval <= 0 {
		return nil, fmt.Errorf("%w: %d", ErrInvalidInterval, cfg.Interval)
	}

	return &Service{
		cfg: cfg,

		scraper:   scraperSvc,
		storage:   storageSvc,
		publisher: publisherSvc,
		parser:    parserSvc,

		wg:        sync.WaitGroup{},
		mu:        sync.RWMutex{},
		runMu:     sync.Mutex{},
		running:   false,
		lastRunAt: time.Now(),
		cancel:    nil,

		logger:  logger,
		metrics: metrics,
	}, nil
}

func (s *Service) Start(_ context.Context) error {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return nil
	}
	s.running = true
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.mu.Unlock()

	s.wg.Go(func() { s.loop(ctx) })

	return nil
}

func (s *Service) Stop(_ context.Context) error {
	s.mu.Lock()

	if !s.running {
		s.mu.Unlock()
		return nil
	}

	s.running = false
	if s.cancel != nil {
		s.cancel()
	}
	s.mu.Unlock()
	s.wg.Wait()

	return nil
}

func (s *Service) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return Status{
		Running:   s.running,
		LastRunAt: s.lastRunAt,
	}
}

func (s *Service) Run(ctx context.Context) error {
	return s.run(ctx)
}

func (s *Service) loop(ctx context.Context) {
	interval := s.cfg.ResolveInterval()

	s.logger.Info("scheduler started", zap.Duration("interval", interval))

	_ = s.run(ctx)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.run(ctx); err != nil {
				s.logger.Error("scheduler run failed", zap.Error(err))
			}
		case <-ctx.Done():
			s.logger.Info("scheduler stopped")
			return
		}
	}
}

func (s *Service) run(ctx context.Context) error {
	s.runMu.Lock()
	defer s.runMu.Unlock()

	var cancel context.CancelFunc
	ctx, cancel = context.WithTimeout(ctx, s.cfg.ResolveCycleTimeout())
	defer cancel()

	defer func(start time.Time) {
		s.metrics.ObserveDuration(time.Since(start).Seconds())
	}(time.Now())

	s.metrics.IncCycles()
	s.logger.Info("running scheduler cycle")

	records, err := s.scraper.Run(ctx)
	if err != nil {
		s.metrics.IncError("scrape")
		return fmt.Errorf("scrape: %w", err)
	}

	if records == nil {
		s.logger.Info("scheduler cycle: no changes, skipping")
		s.setLastRun()
		return nil
	}

	filtered := s.filterFutureRecords(records)
	if len(filtered) == 0 {
		s.logger.Info("scheduler cycle: all records in the past, skipping")
		s.setLastRun()
		return nil
	}

	parsed, err := s.parser.ParseBatch(ctx, filtered)
	if err != nil {
		s.metrics.IncError("parse")
		return fmt.Errorf("parse: %w", err)
	}

	if len(parsed) == 0 {
		s.logger.Info("scheduler cycle: no records parsed successfully, skipping")
		s.setLastRun()
		return nil
	}

	changes, err := s.storage.Diff(ctx, parsed)
	if err != nil {
		s.metrics.IncError("diff")
		return fmt.Errorf("diff: %w", err)
	}

	if len(changes) == 0 {
		s.logger.Info("scheduler cycle: no changes, skipping")
		s.setLastRun()
		return nil
	}

	s.logger.Info("scheduler cycle: publishing changes", zap.Int("count", len(changes)))

	parsed = s.publishAndFilter(ctx, changes, parsed)
	if len(parsed) == 0 {
		s.logger.Warn("scheduler cycle: no records to commit after publish failures")
		s.setLastRun()
		return nil
	}

	if commitErr := s.storage.Commit(ctx, parsed); commitErr != nil {
		s.metrics.IncError("commit")
		return fmt.Errorf("commit: %w", commitErr)
	}

	s.setLastRun()
	s.logger.Info("scheduler cycle completed", zap.Int("committed", len(parsed)))

	return nil
}

func (s *Service) setLastRun() {
	s.mu.Lock()
	s.lastRunAt = time.Now()
	s.mu.Unlock()
}

func (s *Service) filterFutureRecords(records []domain.Record) []domain.Record {
	now := time.Now()
	filtered := make([]domain.Record, 0, len(records))

	for _, r := range records {
		hasFuture := false
		for _, d := range r.Dates {
			if d.After(now) {
				hasFuture = true
				break
			}
		}
		if hasFuture {
			filtered = append(filtered, r)
		}
	}

	return filtered
}

func (s *Service) publishAndFilter(
	ctx context.Context,
	changes []domain.ParsedRecord,
	parsed []domain.ParsedRecord,
) []domain.ParsedRecord {
	hashesToExclude := make(map[string]struct{}, len(changes))

	for _, rec := range changes {
		pubErr := s.publisher.Publish(ctx, rec)
		if pubErr != nil {
			s.metrics.IncError("publish")
			s.logger.Error(
				"scheduler cycle: publish failed, removing record",
				zap.Error(pubErr),
				zap.String("area", rec.Area),
			)
			hashesToExclude[s.storage.HashV2(rec)] = struct{}{}
		}
	}

	if len(hashesToExclude) == 0 {
		return parsed
	}

	filtered := make([]domain.ParsedRecord, 0, len(parsed))
	for _, rec := range parsed {
		if _, exclude := hashesToExclude[s.storage.HashV2(rec)]; !exclude {
			filtered = append(filtered, rec)
		}
	}

	return filtered
}
