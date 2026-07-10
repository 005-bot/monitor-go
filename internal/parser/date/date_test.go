package date_test

import (
	"testing"
	"time"

	"github.com/005-bot/monitor-go/internal/parser/date"
)

func mustParse(t *testing.T, input string) []time.Time {
	t.Helper()
	dates, err := date.ParseDates(input)
	if err != nil {
		t.Fatalf("ParseDates(%q) unexpected error: %v", input, err)
	}
	return dates
}

func TestParseDates_SingleDate(t *testing.T) {
	now := time.Now()
	dates := mustParse(t, "12 марта 10-00")
	if len(dates) != 1 {
		t.Fatalf("expected 1 date, got %d", len(dates))
	}
	d := dates[0]
	if d.Month() != time.March {
		t.Errorf("month = %v, want March", d.Month())
	}
	if d.Day() != 12 {
		t.Errorf("day = %d, want 12", d.Day())
	}
	if d.Hour() != 10 {
		t.Errorf("hour = %d, want 10", d.Hour())
	}
	if d.Minute() != 0 {
		t.Errorf("minute = %d, want 0", d.Minute())
	}
	if d.Year() != now.Year() {
		t.Errorf("year = %d, want %d", d.Year(), now.Year())
	}
}

func TestParseDates_MultipleDates(t *testing.T) {
	dates := mustParse(t, "12 марта 10-00 13 марта 14-30")
	if len(dates) != 2 {
		t.Fatalf("expected 2 dates, got %d", len(dates))
	}
}

func TestParseDates_Midnight24(t *testing.T) {
	dates := mustParse(t, "15 апреля 24-00")
	if len(dates) != 1 {
		t.Fatalf("expected 1 date, got %d", len(dates))
	}
	d := dates[0]
	if d.Hour() != 23 || d.Minute() != 59 {
		t.Errorf("expected 23:59, got %d:%d", d.Hour(), d.Minute())
	}
}

func TestParseDates_ColonSeparator(t *testing.T) {
	dates := mustParse(t, "12 марта 10:00")
	if len(dates) != 1 {
		t.Fatalf("expected 1 date, got %d", len(dates))
	}
	d := dates[0]
	if d.Hour() != 10 || d.Minute() != 0 {
		t.Errorf("expected 10:00, got %d:%d", d.Hour(), d.Minute())
	}
}

func TestParseDates_EmptyInput(t *testing.T) {
	dates, err := date.ParseDates("")
	if err != nil {
		t.Fatalf("ParseDates('') unexpected error: %v", err)
	}
	if len(dates) != 0 {
		t.Fatalf("expected 0 dates, got %d", len(dates))
	}
}

func TestParseDates_NoDate(t *testing.T) {
	dates, err := date.ParseDates("some random text without dates")
	if err != nil {
		t.Fatalf("ParseDates() unexpected error: %v", err)
	}
	if len(dates) != 0 {
		t.Fatalf("expected 0 dates, got %d", len(dates))
	}
}

func TestParseDates_AllMonths(t *testing.T) {
	months := []string{
		"января", "февраля", "марта", "апреля", "мая", "июня",
		"июля", "августа", "сентября", "октября", "ноября", "декабря",
	}
	for i, m := range months {
		dates := mustParse(t, "1 "+m+" 10-00")
		if len(dates) != 1 {
			t.Fatalf("month %q: expected 1 date, got %d", m, len(dates))
		}
		if dates[0].Month() != time.Month(i+1) {
			t.Errorf("month %q: got %v, want %v", m, dates[0].Month(), time.Month(i+1))
		}
	}
}

func TestParseDates_UnknownMonth(t *testing.T) {
	_, err := date.ParseDates("12 неизвестный 10-00")
	if err == nil {
		t.Fatal("expected error for unknown month")
	}
}

func TestParseDates_SingleDigitDay(t *testing.T) {
	dates := mustParse(t, "1 мая 08-00")
	if len(dates) != 1 {
		t.Fatalf("expected 1 date, got %d", len(dates))
	}
	if dates[0].Day() != 1 {
		t.Errorf("day = %d, want 1", dates[0].Day())
	}
}

func TestFormatDates(t *testing.T) {
	now := time.Now()
	d1 := time.Date(now.Year(), time.March, 12, 10, 0, 0, 0, time.Local)
	d2 := time.Date(now.Year(), time.March, 13, 14, 30, 0, 0, time.Local)
	result := date.FormatDates([]time.Time{d1, d2})
	expected := "12 марта 10-00 13 марта 14-30"
	if result != expected {
		t.Errorf("FormatDates = %q, want %q", result, expected)
	}
}

func TestFormatDates_Empty(t *testing.T) {
	result := date.FormatDates(nil)
	if result != "" {
		t.Errorf("FormatDates(nil) = %q, want empty", result)
	}
	result = date.FormatDates([]time.Time{})
	if result != "" {
		t.Errorf("FormatDates([]) = %q, want empty", result)
	}
}
