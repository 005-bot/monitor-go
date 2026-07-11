package scraper

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/005-bot/monitor-go/internal/domain"
	"github.com/005-bot/monitor-go/internal/parser/date"
	"github.com/005-bot/monitor-go/internal/storage"
	"github.com/PuerkitoBio/goquery"
	"go.uber.org/zap"
	"golang.org/x/net/html"
	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var ErrHTTPRequest = errors.New("http request failed")

const requestTimeout = 30 * time.Second

type Service struct {
	url     string
	client  *http.Client
	storage *storage.Service
	logger  *zap.Logger
	metrics *Metrics
}

func NewService(cfg Config, storageSvc *storage.Service, metrics *Metrics, logger *zap.Logger) *Service {
	return &Service{
		url: cfg.URL,
		client: &http.Client{
			Timeout: requestTimeout,
		},
		storage: storageSvc,
		logger:  logger,
		metrics: metrics,
	}
}

func (s *Service) Run(ctx context.Context) ([]domain.Record, error) {
	s.metrics.IncTotal()
	defer s.metrics.ObserveDuration()()
	s.logger.Info("running scraper")

	etag, err := s.fetchEtag(ctx)
	if err != nil {
		s.metrics.IncError("etag_fetch")
		return nil, fmt.Errorf("fetch etag: %w", err)
	}

	if etag == "" {
		s.logger.Warn("etag not found in response, scraping anyway")
	} else {
		var changed bool
		changed, err = s.storage.IsEtagChanged(ctx, etag)
		if err != nil && !errors.Is(err, storage.ErrNotFound) {
			return nil, fmt.Errorf("etag check: %w", err)
		}
		if !changed {
			s.logger.Info("etag not changed, skipping scrape")
			return nil, nil
		}
		s.logger.Info("etag changed, scraping")
	}

	records, err := s.fetchAndParse(ctx)
	if err != nil {
		s.metrics.IncError("fetch_parse")
		return nil, fmt.Errorf("fetch and parse: %w", err)
	}

	s.logger.Info("scrape completed", zap.Int("records", len(records)))
	return records, nil
}

func (s *Service) fetchEtag(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, s.url, nil)
	if err != nil {
		return "", fmt.Errorf("create head request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("head request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%w: HEAD %s returned %s", ErrHTTPRequest, s.url, resp.Status)
	}

	return resp.Header.Get("ETag"), nil
}

func (s *Service) fetchAndParse(ctx context.Context) ([]domain.Record, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return nil, fmt.Errorf("create get request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get request: %w", err)
	}
	defer func() {
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: GET %s returned %s", ErrHTTPRequest, s.url, resp.Status)
	}

	utf8Reader := transform.NewReader(resp.Body, charmap.Windows1251.NewDecoder())

	doc, err := goquery.NewDocumentFromReader(utf8Reader)
	if err != nil {
		return nil, fmt.Errorf("parse html: %w", err)
	}

	return s.parseRecords(doc), nil
}

func (s *Service) parseRecords(doc *goquery.Document) []domain.Record {
	table := doc.Find("table")
	if table.Length() == 0 {
		s.logger.Warn("no table found in document")
		return nil
	}

	var records []domain.Record
	var currentArea string

	table.Find("tr").Each(func(_ int, row *goquery.Selection) {
		text := strings.TrimSpace(row.Text())
		if text == "" {
			return
		}

		cells := row.Find("td")
		if cells.Length() != 3 { //nolint:mnd // expected 3 cells per row
			s.logger.Debug("skipping row with non-3 cells", zap.Int("cells", cells.Length()))
			return
		}

		cell0Text := strings.TrimSpace(cells.Eq(0).Text())
		cell1Text := cells.Eq(1).Text()
		cell2Text := strings.TrimSpace(cells.Eq(2).Text()) //nolint:mnd // third cell

		if cell0Text == "" && strings.Contains(cell1Text, "район") && cell2Text == "" {
			currentArea = collapseWhitespaces(cell1Text)
			return
		}

		if currentArea == "" {
			return
		}

		if cell0Text == "" || cell1Text == "" || cell2Text == "" {
			s.logger.Debug(
				"skipping row with empty cell",
				zap.String("area", currentArea),
			)
			return
		}

		org := getText(cells.Eq(0))
		addr := getText(cells.Eq(1))
		datesRaw := collapseWhitespaces(cells.Eq(2).Text()) //nolint:mnd // third cell

		parsedDates, err := date.ParseDates(datesRaw)
		if err != nil {
			s.logger.Warn("failed to parse dates", zap.String("raw", datesRaw), zap.Error(err))
			return
		}

		records = append(records, domain.Record{
			Area:         currentArea,
			Organization: normalizeMultiline(org),
			Address:      normalizeMultiline(addr),
			Dates:        parsedDates,
		})
	})

	return records
}

func getText(s *goquery.Selection) string {
	var buf strings.Builder
	for _, node := range s.Nodes {
		writeNodeText(node, &buf)
	}
	return normalizeLines(buf.String())
}

func normalizeLines(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.Join(strings.Fields(line), " ")
	}
	return strings.Join(lines, "\n")
}

func writeNodeText(n *html.Node, buf *strings.Builder) {
	if n.Type == html.TextNode {
		text := strings.Join(strings.Fields(n.Data), " ")
		if text == "" {
			return
		}
		if buf.Len() > 0 && buf.String()[buf.Len()-1] != '\n' {
			buf.WriteByte(' ')
		}
		buf.WriteString(text)
		return
	}
	if n.Type == html.ElementNode && n.Data == "br" {
		buf.WriteByte('\n')
		return
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		writeNodeText(c, buf)
	}
}

func collapseWhitespaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func normalizeMultiline(s string) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(line)
	}
	return strings.Join(lines, "\n")
}
