package address_test

import (
	"context"
	"testing"

	"github.com/005-bot/monitor-go/internal/parser/address"
)

func TestNormalize(t *testing.T) {
	p, err := address.NewParser(address.Config{})
	if err != nil {
		t.Fatalf("NewParser: %v", err)
	}

	ctx := context.Background()

	tests := []struct {
		input string
		want  string
	}{
		{"ул. Ленина", "улица Ленина"},
		{"пр. Мира", "проспект Мира"},
		{"Советская", "Советская улица"},
		{"Кольцевая", "Кольцевая улица"},
		{"Красноярский рабочий", "проспект имени газеты Красноярский Рабочий"},
		{"улица Рокоссовского", "улица Рокоссовского"},
		{"Металлургов", "проспект Металлургов"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			m, nerr := p.Normalize(ctx, tt.input)
			if nerr != nil {
				t.Fatalf("Normalize(%q) error: %v", tt.input, nerr)
			}
			if m.Name != tt.want {
				t.Errorf("Normalize(%q) = %q (conf=%.4f), want %q", tt.input, m.Name, m.Confidence, tt.want)
			}
			if m.Confidence < 0.6 {
				t.Errorf("Normalize(%q) confidence=%.4f < 0.6", tt.input, m.Confidence)
			}
		})
	}
}
