package addressparser_test

import (
	"context"
	"errors"
	"testing"

	"github.com/005-bot/monitor-go/internal/addressparser"
)

func TestNormalize(t *testing.T) {
	ap := newParser(t)

	tests := []struct {
		input    string
		expected string
		desc     string
	}{
		{"ул. Ленина", "улица Ленина", "street with abbreviation prefix"},
		{"пр. Мира", "проспект Мира", "avenue with abbreviation prefix"},
		{"ул. Советская", "Советская улица", "street with reversed order"},
		{"ул. Пушкина", "улица Пушкина", "street with abbreviation"},
		{"Вильского", "улица Вильского", "bare surname street"},
		{"Петра Словцова", "улица Петра Словцова", "compound surname street"},
		{"Гусарова", "улица Гусарова", "bare surname"},
		{"Лесопарковая", "Лесопарковая улица", "adjective as bare name"},
		{"Мирошниченко", "улица Мирошниченко", "bare surname"},
		{"Мира", "проспект Мира", "short name to avenue"},
		{"Кольцевая", "Кольцевая улица", "adjective street"},
		{"Красноярский рабочий", "проспект имени газеты Красноярский Рабочий", "complex named avenue"},
		{"Пировская", "Пировская улица", "adjective street"},
		{"Курейская", "Курейская улица", "adjective street"},
		{"Каратузский", "Каратузский переулок", "lane"},
		{"Бийхемская", "Бийхемская улица", "adjective street"},
		{"Назаровская", "Назаровская улица", "adjective street"},
		{"Мате Залки", "улица Мате Залки", "compound name street"},
		{"Космонавтов", "улица Космонавтов", "bare name street"},
		{"Харламова", "улица Харламова", "bare surname"},
		{"Тельмана", "улица Тельмана", "bare surname"},
		{"Джамбульская", "Джамбульская улица", "adjective street"},
		{"Новгородская", "Новгородская улица", "adjective street"},
		{"Металлургов", "проспект Металлургов", "avenue"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result, nErr := ap.Normalize(context.Background(), tt.input)
			if nErr != nil {
				t.Fatalf("Normalize(%q) error: %v", tt.input, nErr)
			}
			if result == nil {
				t.Fatalf("Normalize(%q) = nil, want match with name=%q", tt.input, tt.expected)
			}
			if result.Name != tt.expected {
				t.Errorf("Normalize(%q).Name = %q, want %q", tt.input, result.Name, tt.expected)
			}
			if result.Confidence < 0.6 {
				t.Errorf("Normalize(%q).Confidence = %f, want >= 0.6", tt.input, result.Confidence)
			}
		})
	}
}

func TestNormalizeNoMatch(t *testing.T) {
	ap := newParser(t)

	result, err := ap.Normalize(context.Background(), "1-19")
	if !errors.Is(err, addressparser.ErrNoMatch) {
		t.Fatalf("Normalize(\"1-19\") error = %v, want ErrNoMatch", err)
	}
	if result != nil {
		t.Errorf("Normalize(\"1-19\") = %+v, want nil", result)
	}
}

func TestNormalizeEmpty(t *testing.T) {
	ap := newParser(t)

	result, err := ap.Normalize(context.Background(), "")
	if !errors.Is(err, addressparser.ErrEmptyInput) {
		t.Fatalf("Normalize(\"\") error = %v, want ErrEmptyInput", err)
	}
	if result != nil {
		t.Errorf("Normalize(\"\") = %+v, want nil", result)
	}
}

func TestNormalizeUnknown(t *testing.T) {
	ap := newParser(t)

	result, err := ap.Normalize(context.Background(), "НеизвестнаяУлица123")
	if err != nil && !errors.Is(err, addressparser.ErrNoMatch) {
		t.Fatalf("Normalize(\"НеизвестнаяУлица123\") error: %v", err)
	}
	if errors.Is(err, addressparser.ErrNoMatch) && result != nil {
		t.Fatalf("Normalize returned result %+v with ErrNoMatch", result)
	}
	if result != nil {
		t.Logf("Normalize returned match %+v (may be low-confidence crossing threshold)", result)
	}
}

func TestNormalizeCaseInsensitive(t *testing.T) {
	ap := newParser(t)

	upper, err := ap.Normalize(context.Background(), "УЛ. ЛЕНИНА")
	if err != nil {
		t.Fatalf("Normalize(\"УЛ. ЛЕНИНА\") error: %v", err)
	}
	lower, err := ap.Normalize(context.Background(), "ул. ленина")
	if err != nil {
		t.Fatalf("Normalize(\"ул. ленина\") error: %v", err)
	}
	mixed, err := ap.Normalize(context.Background(), "Ул. Ленина")
	if err != nil {
		t.Fatalf("Normalize(\"Ул. Ленина\") error: %v", err)
	}

	if upper == nil || lower == nil || mixed == nil {
		t.Fatal("Normalize should find match regardless of case")
	}

	if upper.Name != "улица Ленина" {
		t.Errorf("upper case result.Name = %q, want \"улица Ленина\"", upper.Name)
	}
	if lower.Name != "улица Ленина" {
		t.Errorf("lower case result.Name = %q, want \"улица Ленина\"", lower.Name)
	}
	if mixed.Name != "улица Ленина" {
		t.Errorf("mixed case result.Name = %q, want \"улица Ленина\"", mixed.Name)
	}
}

func TestNormalizeConfidenceOrder(t *testing.T) {
	ap := newParser(t)

	exact, err := ap.Normalize(context.Background(), "улица ленина")
	if err != nil {
		t.Fatalf("Normalize(\"улица ленина\") error: %v", err)
	}
	fuzzy, err := ap.Normalize(context.Background(), "ул. Ленина")
	if err != nil {
		t.Fatalf("Normalize(\"ул. Ленина\") error: %v", err)
	}

	if exact == nil || fuzzy == nil {
		t.Fatal("Normalize should find match")
	}

	if exact.Confidence != 1.0 {
		t.Errorf("exact match confidence = %f, want 1.0", exact.Confidence)
	}
	if fuzzy.Confidence >= 1.0 {
		t.Errorf("fuzzy match confidence = %f, want < 1.0", fuzzy.Confidence)
	}
}

func newParser(t *testing.T) *addressparser.AddressParser {
	t.Helper()
	ap, err := addressparser.New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	t.Cleanup(func() { _ = ap.Close() })
	return ap
}
