package parser_test

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/005-bot/address-parser-go"
	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/005-bot/monitor-go/internal/parser"
	"github.com/005-bot/monitor-go/internal/parser/organization"
	"github.com/005-bot/monitor-go/internal/parser/outage"
	"go.uber.org/zap"
)

//nolint:gochecknoglobals // test metrics cache
var (
	parserMetrics     *parser.Metrics
	parserMetricsOnce sync.Once
)

func getParserMetrics() *parser.Metrics {
	parserMetricsOnce.Do(func() {
		parserMetrics = parser.NewMetrics()
	})
	return parserMetrics
}

func newTestService(t *testing.T) *parser.Service {
	t.Helper()

	orgParser := organization.NewParser()

	addrParser, err := address.NewParser(address.Config{})
	if err != nil {
		t.Fatalf("address.NewParser: %v", err)
	}
	t.Cleanup(addrParser.Stop)

	outageParser := outage.NewParser(addrParser)

	logger, _ := zap.NewDevelopment()
	t.Cleanup(func() { _ = logger.Sync() })

	return parser.NewService(orgParser, outageParser, getParserMetrics(), logger)
}

func nowPlus(days, hours int) time.Time {
	return time.Now().
		AddDate(0, 0, days).
		Truncate(time.Hour).
		Add(time.Duration(hours) * time.Hour)
}

func TestParse_ValidRecord(t *testing.T) {
	svc := newTestService(t)

	record := domain.Record{
		Area:         "Железнодорожный район",
		Organization: "Горячее водоснабжение\nАО КТТК\nт. 264-18-62",
		Address:      "ул. Ленина, 1 – ул. Мира, 5, плановая проверка",
		Dates:        []time.Time{nowPlus(1, 9), nowPlus(1, 17)},
	}

	result, err := svc.Parse(context.Background(), record)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	if result.Area != "Железнодорожный район" {
		t.Errorf("Area = %q, want %q", result.Area, "Железнодорожный район")
	}
	if result.Organization.Organization != "АО КТТК" {
		t.Errorf("Organization = %q, want АО КТТК", result.Organization.Organization)
	}
	if len(result.Details.Streets) == 0 {
		t.Error("expected at least 1 street in details")
	}
}

func TestParse_InvalidOrganization(t *testing.T) {
	svc := newTestService(t)

	record := domain.Record{
		Area:         "Центральный район",
		Organization: "",
		Address:      "ул. Ленина, 1",
		Dates:        []time.Time{nowPlus(1, 9)},
	}

	_, err := svc.Parse(context.Background(), record)
	if err == nil {
		t.Fatal("Parse() expected error for empty organization")
	}
}

func TestParseBatch_SkipsFailures(t *testing.T) {
	svc := newTestService(t)

	records := []domain.Record{
		{
			Area:         "Октябрьский район",
			Organization: "Электроснабжение\nПАО Россети Сибирь\nт. 8-800-220-0-220",
			Address:      "ул. Ленина, 1, плановая проверка",
			Dates:        []time.Time{nowPlus(1, 10)},
		},
		{
			Area:         "Свердловский район",
			Organization: "",
			Address:      "пр. Мира, 5",
			Dates:        []time.Time{nowPlus(1, 11)},
		},
		{
			Area:         "Кировский район",
			Organization: "Холодное водоснабжение\nООО КрасРСК\nт. 211-39-63",
			Address:      "ул. Вавилова, 10, аварийный ремонт",
			Dates:        []time.Time{nowPlus(2, 8)},
		},
	}

	parsed, _ := svc.ParseBatch(context.Background(), records)
	if len(parsed) != 2 {
		t.Fatalf("expected 2 parsed records (skipping 1 failure), got %d", len(parsed))
	}
}

func TestParseBatch_AllEmpty(t *testing.T) {
	svc := newTestService(t)

	parsed, _ := svc.ParseBatch(context.Background(), nil)
	if len(parsed) != 0 {
		t.Fatalf("expected 0 parsed records, got %d", len(parsed))
	}
	parsed, _ = svc.ParseBatch(context.Background(), []domain.Record{})
	if len(parsed) != 0 {
		t.Fatalf("expected 0 parsed records, got %d", len(parsed))
	}
}

func TestParse_OrganizationDetectedResourceType(t *testing.T) {
	svc := newTestService(t)

	tests := []struct {
		orgText      string
		wantResource string
	}{
		{"Горячее водоснабжение\nАО КТТК\nт. 264-18-62", "Горячее водоснабжение"},
		{"Электроснабжение\nПАО Россети Сибирь\nт. 8-800-220-0-220", "Электроснабжение"},
		{"Холодное водоснабжение\nООО КрасРСК\nт. 211-39-63", "Холодное водоснабжение"},
	}

	for _, tt := range tests {
		t.Run(strings.Split(tt.orgText, "\n")[0], func(t *testing.T) {
			record := domain.Record{
				Area:         "Тестовый район",
				Organization: tt.orgText,
				Address:      "ул. Тестовая, 1, плановая проверка",
				Dates:        []time.Time{nowPlus(1, 9)},
			}
			result, err := svc.Parse(context.Background(), record)
			if err != nil {
				t.Fatalf("Parse() error: %v", err)
			}
			if result.Organization.Resource != tt.wantResource {
				t.Errorf("Resource = %q, want %q", result.Organization.Resource, tt.wantResource)
			}
		})
	}
}
