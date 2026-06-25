package date

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

var ErrUnknownMonth = errors.New("unknown month")

//nolint:gochecknoglobals // lookup table, not mutable state
var russianMonths = map[string]time.Month{
	"января":   time.January,
	"февраля":  time.February,
	"марта":    time.March,
	"апреля":   time.April,
	"мая":      time.May,
	"июня":     time.June,
	"июля":     time.July,
	"августа":  time.August,
	"сентября": time.September,
	"октября":  time.October,
	"ноября":   time.November,
	"декабря":  time.December,
}

var datePattern = regexp.MustCompile(`(\d{1,2})\s+([а-яё]+)\s+(\d{2})[-:](\d{2})`)

func ParseDates(input string) ([]time.Time, error) {
	matches := datePattern.FindAllStringSubmatch(input, -1)
	if len(matches) == 0 {
		return nil, nil
	}

	now := time.Now()
	year := now.Year()

	result := make([]time.Time, 0, len(matches))
	for _, m := range matches {
		dayStr, monthName, hourStr, minStr := m[1], m[2], m[3], m[4]

		month, ok := russianMonths[strings.ToLower(monthName)]
		if !ok {
			return nil, fmt.Errorf("%w: %s", ErrUnknownMonth, monthName)
		}

		day := parseInt(dayStr)
		hour := parseInt(hourStr)
		minute := parseInt(minStr)

		if hour == 24 && minute == 0 {
			hour = 23
			minute = 59
		}

		result = append(result, time.Date(year, month, day, hour, minute, 0, 0, time.Local))
	}

	return result, nil
}

func FormatDates(dates []time.Time) string {
	parts := make([]string, len(dates))
	for i, d := range dates {
		monthName := russianMonthName(d.Month())
		parts[i] = fmt.Sprintf("%d %s %02d-%02d", d.Day(), monthName, d.Hour(), d.Minute())
	}
	return strings.Join(parts, " ")
}

//nolint:gochecknoglobals // lookup table, not mutable state
var russianMonthNames = map[time.Month]string{
	time.January:   "января",
	time.February:  "февраля",
	time.March:     "марта",
	time.April:     "апреля",
	time.May:       "мая",
	time.June:      "июня",
	time.July:      "июля",
	time.August:    "августа",
	time.September: "сентября",
	time.October:   "октября",
	time.November:  "ноября",
	time.December:  "декабря",
}

func russianMonthName(m time.Month) string {
	return russianMonthNames[m]
}

func parseInt(s string) int {
	n := 0
	for _, c := range []byte(s) {
		if c >= '0' && c <= '9' {
			n = n*10 + int(c-'0') //nolint:mnd // decimal digit conversion
		}
	}
	return n
}
