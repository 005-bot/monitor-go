package address

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/hbollon/go-edlib"
	_ "modernc.org/sqlite" // register sqlite3 driver (pure Go, no CGO)
)

//go:embed streets.db
var streetsDB []byte

var nonWordRe = regexp.MustCompile(`[^\p{L}\p{N}\s\-]+`)

var spaceRe = regexp.MustCompile(`\s+`)

type Parser struct {
	mu       sync.RWMutex
	names    []string
	rows     []streetRow
	exactMap map[string]streetRow
	tempDir  string
}

func NewParser(cfg Config) (*Parser, error) {
	dbPath := cfg.DBPath
	var tempDir string
	if dbPath == "" {
		var err error
		tempDir, err = createTempDir()
		if err != nil {
			return nil, fmt.Errorf("create temp dir: %w", err)
		}

		dbPath = tempDir + "/streets.db"
		if writeErr := writeEmbeddedDB(dbPath); writeErr != nil {
			_ = os.RemoveAll(tempDir)
			return nil, fmt.Errorf("write embedded db: %w", writeErr)
		}
	}

	p := &Parser{
		mu:       sync.RWMutex{},
		names:    make([]string, 0),
		rows:     make([]streetRow, 0),
		exactMap: make(map[string]streetRow),
		tempDir:  tempDir,
	}

	if err := p.loadStreets(dbPath); err != nil {
		return nil, fmt.Errorf("load streets: %w", err)
	}

	return p, nil
}

func (p *Parser) loadStreets(dbPath string) error {
	db, openErr := sql.Open("sqlite", dbPath)
	if openErr != nil {
		return fmt.Errorf("open sqlite: %w", openErr)
	}
	defer db.Close()

	ctx := context.Background()
	rows, queryErr := db.QueryContext(ctx, `SELECT name_normalized, name_original FROM streets`)
	if queryErr != nil {
		return fmt.Errorf("query streets: %w", queryErr)
	}
	defer rows.Close()

	for rows.Next() {
		var normName, origName string
		if scanErr := rows.Scan(&normName, &origName); scanErr != nil {
			return fmt.Errorf("scan row: %w", scanErr)
		}

		row := streetRow{ //nolint:exhaustruct // only needed fields are scanned from DB
			NameOriginal:   origName,
			NameNormalized: normName,
		}

		p.names = append(p.names, normName)
		p.rows = append(p.rows, row)
		p.exactMap[normName] = row
	}

	if err := rows.Err(); err != nil {
		return fmt.Errorf("rows iteration: %w", err)
	}

	return nil
}

func (p *Parser) Stop() {
	if p.tempDir != "" {
		_ = os.RemoveAll(p.tempDir)
	}
}

func (p *Parser) Normalize(_ context.Context, rawInput string) (*Match, error) {
	candidate := cleanName(rawInput)

	p.mu.RLock()
	defer p.mu.RUnlock()

	if row, ok := p.exactMap[candidate]; ok {
		return &Match{
			Name:           row.NameOriginal,
			NormalizedName: row.NameNormalized,
			Confidence:     1.0,
		}, nil
	}

	match := p.fuzzyMatch(candidate)
	if match == nil {
		return nil, fmt.Errorf("%w: %q", ErrNoMatch, rawInput)
	}

	return match, nil
}

func (p *Parser) fuzzyMatch(name string) *Match {
	const (
		ratioWeight  = 0.3
		lcsWeight    = 0.7
		minThreshold = 0.6
	)

	var best *Match

	for i, normName := range p.names {
		similarity, err := edlib.StringsSimilarity(name, normName, edlib.Levenshtein)
		if err != nil {
			continue
		}

		lcsLen := edlib.LCS(name, normName)

		lcsScore := float64(lcsLen) / float64(max(utf8.RuneCountInString(name), 1))

		confidence := float64(similarity)*ratioWeight + lcsScore*lcsWeight

		if confidence < minThreshold {
			continue
		}

		if best == nil || confidence > best.Confidence {
			row := p.rows[i]
			best = &Match{
				Name:           row.NameOriginal,
				NormalizedName: row.NameNormalized,
				Confidence:     confidence,
			}
		}
	}

	return best
}

func cleanName(name string) string {
	name = strings.ToLower(name)
	name = nonWordRe.ReplaceAllString(name, "")
	name = spaceRe.ReplaceAllString(name, " ")
	return strings.TrimSpace(name)
}
