package publisher_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/005-bot/monitor-go/internal/publisher"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals // test metrics cache
var (
	publisherMetrics     *publisher.Metrics
	publisherMetricsOnce sync.Once
)

func getPublisherMetrics() *publisher.Metrics {
	publisherMetricsOnce.Do(func() {
		publisherMetrics = publisher.NewMetrics()
	})
	return publisherMetrics
}

func newTestService(t *testing.T) (*publisher.Service, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	t.Cleanup(func() { _ = rdb.Close() })

	cfg := publisher.Config{
		Prefix: "test-prefix",
	}

	logger, _ := zap.NewDevelopment()
	t.Cleanup(func() { _ = logger.Sync() })

	svc := publisher.NewService(rdb, cfg, getPublisherMetrics(), logger)

	return svc, mr
}

func makeRecord(area string) domain.ParsedRecord {
	return domain.ParsedRecord{
		Area: area,
		Organization: domain.OrganizationInfo{
			Resource:     "Электроснабжение",
			Organization: "ПАО Россети Сибирь",
			Phones:       []string{"8-800-220-0-220"},
		},
		Details: domain.OutageDetails{
			Streets: []domain.Street{{Name: "ул. Ленина"}},
		},
		Dates: []time.Time{time.Now().Add(24 * time.Hour)},
	}
}

func TestPublish_Success(t *testing.T) {
	svc, mr := newTestService(t)

	sub := newSubscriber(t, mr)
	defer sub.Close()

	record := makeRecord("Железнодорожный район")
	err := svc.Publish(context.Background(), record)
	if err != nil {
		t.Fatalf("Publish() error: %v", err)
	}
}

func TestPublish_NoSubscribers(t *testing.T) {
	svc, _ := newTestService(t)

	record := makeRecord("Центральный район")
	err := svc.Publish(context.Background(), record)
	if err == nil {
		t.Fatal("Publish() expected error (no subscribers)")
	}
}

type testSubscriber struct {
	pubSub *redis.PubSub
	rdb    *redis.Client
}

func newSubscriber(t *testing.T, mr *miniredis.Miniredis) *testSubscriber {
	t.Helper()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	t.Cleanup(func() { _ = rdb.Close() })

	pubSub := rdb.Subscribe(context.Background(), "test-prefix:outages")

	return &testSubscriber{
		pubSub: pubSub,
		rdb:    rdb,
	}
}

func (s *testSubscriber) Close() {
	_ = s.pubSub.Close()
	s.rdb.Close()
}
