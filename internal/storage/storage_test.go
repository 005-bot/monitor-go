package storage_test

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/005-bot/monitor-go/internal/storage"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals // test metrics cache
var (
	storageMetrics     *storage.Metrics
	storageMetricsOnce sync.Once
)

func getStorageMetrics() *storage.Metrics {
	storageMetricsOnce.Do(func() {
		storageMetrics = storage.NewMetrics()
	})
	return storageMetrics
}

func newTestService(t *testing.T) (*storage.Service, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	t.Cleanup(func() { _ = rdb.Close() })

	cfg := storage.Config{
		Prefix:  "test-prefix",
		TTLDays: 5,
	}

	logger, _ := zap.NewDevelopment()
	t.Cleanup(func() { _ = logger.Sync() })

	svc := storage.NewService(rdb, cfg, getStorageMetrics(), logger)

	return svc, mr
}

func makeRecord(street, reason string) domain.ParsedRecord {
	return domain.ParsedRecord{
		Area: "р-н 1",
		Organization: domain.OrganizationInfo{
			Resource:     "Электроснабжение",
			Organization: "Орг 1",
			Phones:       []string{"111-11-11"},
		},
		Details: domain.OutageDetails{
			Streets: []domain.Street{{Name: street}},
			Reason:  &domain.Reason{Type: reason},
		},
		Dates: []time.Time{time.Now().Add(24 * time.Hour)},
	}
}

func TestGetEtag_NotFound(t *testing.T) {
	svc, _ := newTestService(t)
	etag, err := svc.GetEtag(context.Background())
	if err != nil {
		t.Fatalf("GetEtag() error: %v", err)
	}
	if etag != "" {
		t.Errorf("GetEtag() = %q, want empty", etag)
	}
}

func TestIsEtagChanged_FirstTime(t *testing.T) {
	svc, _ := newTestService(t)
	changed, err := svc.IsEtagChanged(context.Background(), "etag-123")
	if !errors.Is(err, storage.ErrNotFound) {
		t.Fatalf("IsEtagChanged() expected ErrNotFound, got %v", err)
	}
	if !changed {
		t.Error("IsEtagChanged() = false, want true (first time)")
	}
}

func TestIsEtagChanged_SameEtag(t *testing.T) {
	svc, mr := newTestService(t)
	mr.Set("test-prefix:etag", "etag-123")

	changed, err := svc.IsEtagChanged(context.Background(), "etag-123")
	if err != nil {
		t.Fatalf("IsEtagChanged() error: %v", err)
	}
	if changed {
		t.Error("IsEtagChanged() = true, want false (same etag)")
	}
}

func TestIsEtagChanged_DifferentEtag(t *testing.T) {
	svc, mr := newTestService(t)
	mr.Set("test-prefix:etag", "etag-old")

	changed, err := svc.IsEtagChanged(context.Background(), "etag-new")
	if err != nil {
		t.Fatalf("IsEtagChanged() error: %v", err)
	}
	if !changed {
		t.Error("IsEtagChanged() = false, want true (different etag)")
	}
}

func TestHashV2_Deterministic(t *testing.T) {
	svc, _ := newTestService(t)
	r1 := makeRecord("ул. Ленина", "плановая проверка")
	r2 := makeRecord("ул. Ленина", "плановая проверка")

	h1 := svc.HashV2(r1)
	h2 := svc.HashV2(r2)
	if h1 != h2 {
		t.Error("HashV2 should be deterministic for identical records")
	}
}

func TestHashV2_DifferentStreetDifferentHash(t *testing.T) {
	svc, _ := newTestService(t)
	r1 := makeRecord("ул. Ленина", "плановая")
	r2 := makeRecord("ул. Мира", "плановая")

	h1 := svc.HashV2(r1)
	h2 := svc.HashV2(r2)
	if h1 == h2 {
		t.Error("HashV2 should differ for different streets")
	}
}

func TestDiff_NoExistingRecords(t *testing.T) {
	svc, _ := newTestService(t)

	records := []domain.ParsedRecord{
		makeRecord("ул. Ленина", "плановая"),
	}

	changed, err := svc.Diff(context.Background(), records)
	if err != nil {
		t.Fatalf("Diff() error: %v", err)
	}
	if len(changed) != 1 {
		t.Fatalf("expected 1 changed record, got %d", len(changed))
	}
}

func TestDiff_AllExisting(t *testing.T) {
	svc, mr := newTestService(t)

	record := makeRecord("ул. Ленина", "плановая")
	hash := svc.HashV2(record)
	mr.HSet("test-prefix:records_v2", hash, `{"area":"р-н 1"}`)

	changed, err := svc.Diff(context.Background(), []domain.ParsedRecord{record})
	if err != nil {
		t.Fatalf("Diff() error: %v", err)
	}
	if len(changed) != 0 {
		t.Fatalf("expected 0 changed records, got %d", len(changed))
	}
}

func TestDiff_EmptyInput(t *testing.T) {
	svc, _ := newTestService(t)

	changed, err := svc.Diff(context.Background(), nil)
	if err != nil {
		t.Fatalf("Diff(nil) error: %v", err)
	}
	if len(changed) != 0 {
		t.Fatalf("expected 0 changed records, got %d", len(changed))
	}

	changed, err = svc.Diff(context.Background(), []domain.ParsedRecord{})
	if err != nil {
		t.Fatalf("Diff(empty) error: %v", err)
	}
	if len(changed) != 0 {
		t.Fatalf("expected 0 changed records, got %d", len(changed))
	}
}

func TestCommit_NewRecords(t *testing.T) {
	svc, mr := newTestService(t)

	records := []domain.ParsedRecord{
		makeRecord("ул. Ленина", "плановая"),
	}

	err := svc.Commit(context.Background(), records)
	if err != nil {
		t.Fatalf("Commit() error: %v", err)
	}

	keys := mr.Keys()
	if !slices.Contains(keys, "test-prefix:records_v2") {
		t.Error("Commit() did not create records_v2 hash")
	}
}

func TestCommit_EmptyInput(t *testing.T) {
	svc, _ := newTestService(t)

	err := svc.Commit(context.Background(), nil)
	if err != nil {
		t.Fatalf("Commit(nil) error: %v", err)
	}

	err = svc.Commit(context.Background(), []domain.ParsedRecord{})
	if err != nil {
		t.Fatalf("Commit(empty) error: %v", err)
	}
}

func TestGetEtag_AfterSet(t *testing.T) {
	svc, mr := newTestService(t)
	mr.Set("test-prefix:etag", "my-etag-value")

	etag, err := svc.GetEtag(context.Background())
	if err != nil {
		t.Fatalf("GetEtag() error: %v", err)
	}
	if etag != "my-etag-value" {
		t.Errorf("GetEtag() = %q, want %q", etag, "my-etag-value")
	}
}
