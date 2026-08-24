package scheduler_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/005-bot/address-parser-go"
	"github.com/005-bot/monitor-go/internal/parser"
	"github.com/005-bot/monitor-go/internal/parser/organization"
	"github.com/005-bot/monitor-go/internal/parser/outage"
	"github.com/005-bot/monitor-go/internal/publisher"
	"github.com/005-bot/monitor-go/internal/scheduler"
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
	schedMetrics              *scheduler.Metrics
	schedMetricsOnce          sync.Once
	schedScraperMetrics       *scraper.Metrics
	schedScraperMetricsOnce   sync.Once
	schedStoreMetrics         *storage.Metrics
	schedStoreMetricsOnce     sync.Once
	schedParserMetrics        *parser.Metrics
	schedParserMetricsOnce    sync.Once
	schedPublisherMetrics     *publisher.Metrics
	schedPublisherMetricsOnce sync.Once
)

func getSchedMetrics() *scheduler.Metrics {
	schedMetricsOnce.Do(func() { schedMetrics = scheduler.NewMetrics() })
	return schedMetrics
}
func getSchedScraperMetrics() *scraper.Metrics {
	schedScraperMetricsOnce.Do(func() { schedScraperMetrics = scraper.NewMetrics() })
	return schedScraperMetrics
}
func getSchedStoreMetrics() *storage.Metrics {
	schedStoreMetricsOnce.Do(func() { schedStoreMetrics = storage.NewMetrics() })
	return schedStoreMetrics
}
func getSchedParserMetrics() *parser.Metrics {
	schedParserMetricsOnce.Do(func() { schedParserMetrics = parser.NewMetrics() })
	return schedParserMetrics
}
func getSchedPublisherMetrics() *publisher.Metrics {
	schedPublisherMetricsOnce.Do(func() { schedPublisherMetrics = publisher.NewMetrics() })
	return schedPublisherMetrics
}

type testEnv struct {
	scheduler  *scheduler.Service
	mr         *miniredis.Miniredis
	httpServer *httptest.Server
}

func newTestEnv(t *testing.T, html string, intervalSec int) *testEnv {
	t.Helper()

	httpServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("ETag", "test-etag")
		w.Header().Set("Content-Type", "text/html; charset=windows-1251")
		w.WriteHeader(http.StatusOK)

		encoded, err := io.ReadAll(transform.NewReader(
			bytes.NewReader([]byte(html)),
			charmap.Windows1251.NewEncoder(),
		))
		if err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(encoded)
	}))
	t.Cleanup(httpServer.Close)

	mr := miniredis.RunT(t)

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	storageCfg := storage.Config{Prefix: "test-prefix", TTLDays: 5}
	storageSvc := storage.NewService(rdb, storageCfg, getSchedStoreMetrics(), logger(t))

	scraperCfg := scraper.Config{URL: httpServer.URL, Interval: 300}
	scraperSvc := scraper.NewService(scraperCfg, storageSvc, getSchedScraperMetrics(), logger(t))

	addrParser, err := address.NewParser(address.Config{})
	if err != nil {
		t.Fatalf("address.NewParser: %v", err)
	}
	t.Cleanup(addrParser.Stop)

	outageParser := outage.NewParser(addrParser)
	orgParser := organization.NewParser()
	parserSvc := parser.NewService(orgParser, outageParser, getSchedParserMetrics(), logger(t))

	publisherCfg := publisher.Config{Prefix: "test-prefix"}
	publisherSvc := publisher.NewService(rdb, publisherCfg, getSchedPublisherMetrics(), logger(t))

	schedulerCfg := scheduler.Config{Interval: intervalSec, CycleTimeout: 5 * time.Second}
	schedulerSvc, err := scheduler.NewService(
		schedulerCfg, scraperSvc, storageSvc, publisherSvc, parserSvc,
		getSchedMetrics(), logger(t),
	)
	if err != nil {
		t.Fatalf("scheduler.NewService: %v", err)
	}

	subscribeToChannel(t, mr, "test-prefix:outages")

	return &testEnv{
		scheduler:  schedulerSvc,
		mr:         mr,
		httpServer: httpServer,
	}
}

func logger(t *testing.T) *zap.Logger {
	t.Helper()
	l, _ := zap.NewDevelopment()
	t.Cleanup(func() { _ = l.Sync() })
	return l
}

func subscribeToChannel(t *testing.T, mr *miniredis.Miniredis, channel string) {
	t.Helper()
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	pubSub := rdb.Subscribe(context.Background(), channel)
	t.Cleanup(func() { _ = pubSub.Close() })
}

func TestRun_FullCycle(t *testing.T) {
	html := `<html><body><table>
<tr><td></td><td>Железнодорожный район</td><td></td></tr>
<tr><td>Горячее водоснабжение<br/>АО КТТК<br/>т. 264-18-62</td><td>ул. Ленина, 1<br/>плановая проверка</td><td>` + futureDateStr() + ` 10-00</td></tr>
</table></body></html>`

	env := newTestEnv(t, html, 300)

	err := env.scheduler.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
}

func TestRun_EtagNotChanged(t *testing.T) {
	html := `<html><body><table>
<tr><td></td><td>Железнодорожный район</td><td></td></tr>
<tr><td>Горячее водоснабжение<br/>АО КТТК<br/>т. 264-18-62</td><td>ул. Ленина, 1<br/>плановая проверка</td><td>` + futureDateStr() + ` 10-00</td></tr>
</table></body></html>`

	env := newTestEnv(t, html, 300)
	env.mr.Set("test-prefix:etag", "test-etag")

	err := env.scheduler.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
}

func TestRun_EmptyHTML(t *testing.T) {
	html := `<html><body>no table</body></html>`

	env := newTestEnv(t, html, 300)

	err := env.scheduler.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
}

func TestRun_AllRecordsInPast(t *testing.T) {
	html := `<html><body><table>
<tr><td></td><td>Кировский район</td><td></td></tr>
<tr><td>Горячее водоснабжение<br/>АО КТТК<br/>т. 264-18-62</td><td>ул. Ленина, 1<br/>плановая проверка</td><td>01 января 00-00</td></tr>
</table></body></html>`

	env := newTestEnv(t, html, 300)

	err := env.scheduler.Run(context.Background())
	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}
}

func TestStartStop(t *testing.T) {
	html := `<html><body><table>
<tr><td></td><td>Октябрьский район</td><td></td></tr>
<tr><td>Электроснабжение<br/>ПАО Россети Сибирь<br/>т. 8-800-220-0-220</td><td>ул. Мира, 5<br/>аварийный ремонт</td><td>` + futureDateStr() + ` 10-00</td></tr>
</table></body></html>`

	env := newTestEnv(t, html, 1)

	if err := env.scheduler.Start(context.Background()); err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if err := env.scheduler.Stop(context.Background()); err != nil {
		t.Fatalf("Stop() error: %v", err)
	}
}

func TestStatus(t *testing.T) {
	env := newTestEnv(t, `<html><body></body></html>`, 300)

	status := env.scheduler.Status()
	if status.Running {
		t.Error("Status().Running = true, want false")
	}
	if status.LastRunAt.IsZero() {
		t.Error("Status().LastRunAt is zero")
	}
}

func TestNewService_InvalidInterval(t *testing.T) {
	_, err := scheduler.NewService(
		scheduler.Config{Interval: 0},
		nil, nil, nil, nil,
		getSchedMetrics(),
		logger(t),
	)
	if err == nil {
		t.Fatal("NewService with 0 interval expected error")
	}
}

func russianMonthName(m time.Month) string {
	names := map[time.Month]string{
		time.January: "января", time.February: "февраля", time.March: "марта",
		time.April: "апреля", time.May: "мая", time.June: "июня",
		time.July: "июля", time.August: "августа", time.September: "сентября",
		time.October: "октября", time.November: "ноября", time.December: "декабря",
	}
	return names[m]
}

func futureDateStr() string {
	future := time.Now().AddDate(0, 0, 1)
	return fmt.Sprintf("%d %s 10-00", future.Day(), russianMonthName(future.Month()))
}
