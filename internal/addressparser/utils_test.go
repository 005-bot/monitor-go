package addressparser_test

import (
	"testing"

	"github.com/005-bot/monitor-go/internal/addressparser"
)

func TestCleanName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"ул. Ленина", "ул ленина"},
		{"пр. Мира", "пр мира"},
		{"ул. Советская", "ул советская"},
		{"  Ул.   Ленина  ", "ул ленина"},
		{"Красноярский рабочий", "красноярский рабочий"},
		{"1-19", "1-19"},
		{"", ""},
		{"  ", ""},
		{"ул. (Ленина)", "ул ленина"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := addressparser.CleanName(tt.input)
			if got != tt.expected {
				t.Errorf("cleanName(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
