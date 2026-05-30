package addressparser

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"os"
	"regexp"

	_ "modernc.org/sqlite" // sqlite driver for embedded DB
)

const minConfidence = 0.5

var (
	nonWordRE    = regexp.MustCompile(`[^\p{L}\p{N}_\s\-]`)
	multiSpaceRE = regexp.MustCompile(`\s+`)
)

type Match struct {
	Name           string
	NormalizedName string
	Confidence     float64
}

type AddressParser struct {
	db     *sql.DB
	dbPath string
}

func New() (*AddressParser, error) {
	f, err := os.CreateTemp("", "streets-*.db")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}

	if _, writeErr := f.Write(streetsDB); writeErr != nil {
		writeErr = errors.Join(writeErr, f.Close(), os.Remove(f.Name()))
		return nil, fmt.Errorf("write temp file: %w", writeErr)
	}

	dbPath := f.Name()
	if closeErr := f.Close(); closeErr != nil {
		_ = os.Remove(dbPath)
		return nil, fmt.Errorf("close temp file: %w", closeErr)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		err = errors.Join(err, os.Remove(dbPath))
		return nil, fmt.Errorf("open database: %w", err)
	}

	return &AddressParser{db: db, dbPath: dbPath}, nil
}

func (p *AddressParser) Close() error {
	var err error
	if closeErr := p.db.Close(); closeErr != nil {
		err = errors.Join(err, fmt.Errorf("close database: %w", closeErr))
	}
	if removeErr := os.Remove(p.dbPath); removeErr != nil && !errors.Is(removeErr, os.ErrNotExist) {
		err = errors.Join(err, fmt.Errorf("remove database: %w", removeErr))
	}
	return err
}

func (p *AddressParser) Normalize(ctx context.Context, rawInput string) (*Match, error) {
	name := CleanName(rawInput)
	if name == "" {
		return nil, ErrEmptyInput
	}
	return p.fuzzyDBMatch(ctx, name)
}

func (p *AddressParser) fuzzyDBMatch(ctx context.Context, name string) (*Match, error) {
	rows, err := p.db.QueryContext(ctx, "SELECT name_normalized, name_original FROM streets")
	if err != nil {
		return nil, fmt.Errorf("query database: %w", err)
	}
	defer rows.Close()

	var candidate *Match

	for rows.Next() {
		var normName, origName string
		if rowErr := rows.Scan(&normName, &origName); rowErr != nil {
			return nil, fmt.Errorf("scan street row: %w", rowErr)
		}

		if normName == name {
			return &Match{
				Name:           origName,
				NormalizedName: normName,
				Confidence:     1.0,
			}, nil
		}

		sm := newSequenceMatcher(name, normName)
		r := sm.ratio()
		lcs := sm.longestMatchSize()

		lcsScore := float64(lcs) / math.Max(float64(len([]rune(name))), 1)
		score := r*0.3 + lcsScore*0.7 //nolint:mnd // magic numbers

		if candidate == nil || score > candidate.Confidence {
			candidate = &Match{
				Name:           origName,
				NormalizedName: normName,
				Confidence:     score,
			}
		}
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		return nil, fmt.Errorf("iterate streets: %w", rowsErr)
	}

	if candidate != nil && candidate.Confidence < minConfidence {
		return nil, ErrNoMatch
	}
	return candidate, nil
}
