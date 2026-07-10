package scraper_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/005-bot/monitor-go/internal/scraper"
	"github.com/005-bot/monitor-go/internal/storage"
	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

//nolint:gochecknoglobals // test metrics cache
var (
	scraperMetrics     *scraper.Metrics
	scraperMetricsOnce sync.Once
	storeMetrics       *storage.Metrics
	storeMetricsOnce   sync.Once
)

func getScraperMetrics() *scraper.Metrics {
	scraperMetricsOnce.Do(func() {
		scraperMetrics = scraper.NewMetrics()
	})
	return scraperMetrics
}

func getStoreMetrics() *storage.Metrics {
	storeMetricsOnce.Do(func() {
		storeMetrics = storage.NewMetrics()
	})
	return storeMetrics
}

func newTestServer(htmlContent string, etag string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if etag != "" {
			w.Header().Set("ETag", etag)
		}
		w.Header().Set("Content-Type", "text/html; charset=windows-1251")
		w.WriteHeader(http.StatusOK)

		encoded, err := io.ReadAll(transform.NewReader(
			bytes.NewReader([]byte(htmlContent)),
			charmap.Windows1251.NewEncoder(),
		))
		if err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(encoded)
	}))
}

func newStorageService(t *testing.T, mr *miniredis.Miniredis) *storage.Service {
	t.Helper()

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

	return storage.NewService(rdb, cfg, getStoreMetrics(), logger)
}

func TestRun_Success(t *testing.T) {
	html := `<html><body><table>
<tr><td></td><td>Железнодорожный район</td><td></td></tr>
<tr><td>Горячее водоснабжение<br/>АО КТТК<br/>т. 264-18-62</td><td>ул. Ленина, 1<br/>плановая проверка</td><td>12 марта 10-00</td></tr>
</table></body></html>`

	srv := newTestServer(html, "etag-123")
	defer srv.Close()

	mr := miniredis.RunT(t)

	storageSvc := newStorageService(t, mr)

	cfg := scraper.Config{
		URL:      srv.URL,
		Interval: 300,
	}

	logger, _ := zap.NewDevelopment()
	t.Cleanup(func() { _ = logger.Sync() })

	svc := scraper.NewService(cfg, storageSvc, getScraperMetrics(), logger)

	records, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].Area != "Железнодорожный район" {
		t.Errorf("Area = %q, want Железнодорожный район", records[0].Area)
	}
}

func TestRun_EtagNotChanged(t *testing.T) {
	html := `<html><body><table>
<tr><td></td><td>Железнодорожный район</td><td></td></tr>
<tr><td>Горячее водоснабжение<br/>АО КТТК<br/>т. 264-18-62</td><td>ул. Ленина, 1<br/>плановая проверка</td><td>12 марта 10-00</td></tr>
</table></body></html>`

	srv := newTestServer(html, "etag-123")
	defer srv.Close()

	mr := miniredis.RunT(t)
	mr.Set("test-prefix:etag", "etag-123")

	storageSvc := newStorageService(t, mr)

	cfg := scraper.Config{
		URL:      srv.URL,
		Interval: 300,
	}

	logger, _ := zap.NewDevelopment()
	t.Cleanup(func() { _ = logger.Sync() })

	svc := scraper.NewService(cfg, storageSvc, getScraperMetrics(), logger)

	records, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if records != nil {
		t.Fatalf("expected nil records (not changed), got %d", len(records))
	}
}

func TestRun_EtagChanged(t *testing.T) {
	html := `<html><body><table>
<tr><td></td><td>Октябрьский район</td><td></td></tr>
<tr><td>Электроснабжение<br/>ПАО Россети Сибирь<br/>т. 8-800-220-0-220</td><td>ул. Мира, 5<br/>аварийный ремонт</td><td>13 марта 14-00</td></tr>
</table></body></html>`

	srv := newTestServer(html, "etag-456")
	defer srv.Close()

	mr := miniredis.RunT(t)
	mr.Set("test-prefix:etag", "etag-123")

	storageSvc := newStorageService(t, mr)

	cfg := scraper.Config{
		URL:      srv.URL,
		Interval: 300,
	}

	logger, _ := zap.NewDevelopment()
	t.Cleanup(func() { _ = logger.Sync() })

	svc := scraper.NewService(cfg, storageSvc, getScraperMetrics(), logger)

	records, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}

func TestRun_NoEtagInResponse(t *testing.T) {
	html := `<html><body><table>
<tr><td></td><td>Советский район</td><td></td></tr>
<tr><td>Горячее водоснабжение<br/>ООО СибЭР<br/>т. 214-93-51</td><td>ул. Вавилова, 10<br/>плановая проверка</td><td>15 марта 09-00</td></tr>
</table></body></html>`

	srv := newTestServer(html, "")
	defer srv.Close()

	mr := miniredis.RunT(t)

	storageSvc := newStorageService(t, mr)

	cfg := scraper.Config{
		URL:      srv.URL,
		Interval: 300,
	}

	logger, _ := zap.NewDevelopment()
	t.Cleanup(func() { _ = logger.Sync() })

	svc := scraper.NewService(cfg, storageSvc, getScraperMetrics(), logger)

	records, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
}

func TestRun_NoTable(t *testing.T) {
	html := `<html><body><p>no table here</p></body></html>`

	srv := newTestServer(html, "etag-789")
	defer srv.Close()

	mr := miniredis.RunT(t)

	storageSvc := newStorageService(t, mr)

	cfg := scraper.Config{
		URL:      srv.URL,
		Interval: 300,
	}

	logger, _ := zap.NewDevelopment()
	t.Cleanup(func() { _ = logger.Sync() })

	svc := scraper.NewService(cfg, storageSvc, getScraperMetrics(), logger)

	records, err := svc.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
	if len(records) != 0 {
		t.Fatalf("expected 0 records, got %d", len(records))
	}
}
