package outage //nolint:testpackage // tests internal functions

import (
	"context"
	"testing"

	"github.com/005-bot/address-parser-go"
)

func TestParseReason(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType string
		wantDesc string
	}{
		{
			name:     "standard reason",
			input:    "плановое - устранение дефекта на обратном трубопроводе",
			wantType: "плановое",
			wantDesc: "устранение дефекта на обратном трубопроводе",
		},
		{
			name:     "no separator",
			input:    "плановое",
			wantType: "неизвестное",
			wantDesc: "плановое",
		},
		{
			name:     "hyphen in description",
			input:    "аварийное - отключение - повторное",
			wantType: "аварийное",
			wantDesc: "отключение - повторное",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseReason(tt.input)
			if got.Type != tt.wantType {
				t.Errorf("parseReason(%q).Type = %q, want %q", tt.input, got.Type, tt.wantType)
			}
			if got.Description != tt.wantDesc {
				t.Errorf("parseReason(%q).Description = %q, want %q", tt.input, got.Description, tt.wantDesc)
			}
		})
	}
}

func TestParseCombinedInput(t *testing.T) {
	ctx := context.Background()
	addrParser, parseErr := address.NewParser(address.Config{})
	if parseErr != nil {
		t.Fatalf("NewParser: %v", parseErr)
	}

	p := NewParser(addrParser)

	tests := []struct {
		name     string
		input    string
		wantStr  string
		wantType string
		wantDesc string
	}{
		{
			name:     "single line with all streets and reason",
			input:    "город Красноярск: Ленина 137, 143; Мира 120; плановое - устранение дефекта",
			wantStr:  "улица Ленина 137, 143\nпроспект Мира 120",
			wantType: "плановое",
			wantDesc: "устранение дефекта",
		},
		{
			name:     "single line with one street",
			input:    "город Красноярск: Ленина 137; аварийное - ремонт",
			wantStr:  "улица Ленина 137",
			wantType: "аварийное",
			wantDesc: "ремонт",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, callErr := p.Parse(ctx, tt.input)
			if callErr != nil {
				t.Fatalf("Parse(%q) error: %v", tt.input, callErr)
			}
			if got == nil {
				t.Fatalf("Parse(%q) = nil", tt.input)
			}
			if got.Address() != tt.wantStr {
				t.Errorf("Parse(%q).Address() = %q, want %q", tt.input, got.Address(), tt.wantStr)
			}
			if got.Reason == nil {
				t.Fatalf("Parse(%q).Reason = nil", tt.input)
			}
			if got.Reason.Type != tt.wantType {
				t.Errorf("Parse(%q).Reason.Type = %q, want %q", tt.input, got.Reason.Type, tt.wantType)
			}
			if got.Reason.Description != tt.wantDesc {
				t.Errorf("Parse(%q).Reason.Description = %q, want %q", tt.input, got.Reason.Description, tt.wantDesc)
			}
		})
	}
}

func TestParseMultiLine(t *testing.T) {
	ctx := context.Background()
	addrParser, parseErr := address.NewParser(address.Config{})
	if parseErr != nil {
		t.Fatalf("NewParser: %v", parseErr)
	}

	p := NewParser(addrParser)

	tests := []struct {
		name     string
		input    string
		wantStr  string
		wantType string
		wantDesc string
	}{
		{
			name:     "multi-line with streets and reason separated",
			input:    "город Красноярск: Ленина 137, 143; Мира 120;\nплановое - устранение дефекта",
			wantStr:  "улица Ленина 137, 143\nпроспект Мира 120",
			wantType: "плановое",
			wantDesc: "устранение дефекта",
		},
		{
			name:     "streets only, no reason",
			input:    "город Красноярск: Ленина 137",
			wantStr:  "улица Ленина 137",
			wantType: "",
			wantDesc: "",
		},
		{
			name:     "settlement not in street DB is preserved",
			input:    "город Красноярск: д. Бугачево\nаварийное - устранение аварийной ситуаций на водопроводных сетях",
			wantStr:  "д. Бугачево",
			wantType: "аварийное",
			wantDesc: "устранение аварийной ситуаций на водопроводных сетях",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, callErr := p.Parse(ctx, tt.input)
			if callErr != nil {
				t.Fatalf("Parse(%q) error: %v", tt.input, callErr)
			}
			if got == nil {
				t.Fatalf("Parse(%q) = nil", tt.input)
			}
			if got.Address() != tt.wantStr {
				t.Errorf("Parse(%q).Address() = %q, want %q", tt.input, got.Address(), tt.wantStr)
			}
			if tt.wantType == "" {
				if got.Reason != nil {
					t.Errorf("Parse(%q).Reason = %+v, want nil", tt.input, got.Reason)
				}
				return
			}
			if got.Reason == nil {
				t.Fatalf("Parse(%q).Reason = nil", tt.input)
			}
			if got.Reason.Type != tt.wantType {
				t.Errorf("Parse(%q).Reason.Type = %q, want %q", tt.input, got.Reason.Type, tt.wantType)
			}
			if got.Reason.Description != tt.wantDesc {
				t.Errorf("Parse(%q).Reason.Description = %q, want %q", tt.input, got.Reason.Description, tt.wantDesc)
			}
		})
	}
}

func TestParseEmptyInput(t *testing.T) {
	ctx := context.Background()
	addrParser, parseErr := address.NewParser(address.Config{})
	if parseErr != nil {
		t.Fatalf("NewParser: %v", parseErr)
	}

	p := NewParser(addrParser)

	got, callErr := p.Parse(ctx, "")
	if callErr != nil {
		t.Fatalf("Parse('') error: %v", callErr)
	}
	if got != nil {
		t.Fatal("Parse('') = non-nil")
	}
}

func TestSplitStreetAndNumbers(t *testing.T) {
	tests := []struct {
		input      string
		wantStreet string
		wantNums   string
	}{
		{"Ленина 137", "Ленина", "137"},
		{"Ленина 137, 143", "Ленина", "137, 143"},
		{"Мира 120", "Мира", "120"},
		{"Профсоюзов 16, 18", "Профсоюзов", "16, 18"},
		{"Ленина", "Ленина", ""},
		{"д. Бугачево", "д. Бугачево", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			street, nums := splitStreetAndNumbers(tt.input)
			if street != tt.wantStreet {
				t.Errorf("splitStreetAndNumbers(%q) street = %q, want %q", tt.input, street, tt.wantStreet)
			}
			if nums != tt.wantNums {
				t.Errorf("splitStreetAndNumbers(%q) nums = %q, want %q", tt.input, nums, tt.wantNums)
			}
		})
	}
}

func TestProcessBuildingNumbers(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"137, 143", []string{"137", "143"}},
		{"16, 18", []string{"16", "18"}},
		{"", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := processBuildingNumbers(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("processBuildingNumbers(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("processBuildingNumbers(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseWaterDeliveries(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		want          int
		wantStreet    string
		wantBuildings string
		wantTimeStart string
		wantTimeEnd   string
	}{
		{"no water deliveries", "плановое - ремонт", 0, "", "", "", ""},
		{"single delivery", "Подвоз воды: Ленина 10 с 09-00 до 18-00", 1, "Ленина", "10", "09-00", "18-00"},
		{
			"multiple deliveries",
			"Подвоз воды: Ленина 10 с 09-00 до 18-00; Мира 5 с 10-00 до 17-00",
			2,
			"Ленина",
			"10",
			"09-00",
			"18-00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseWaterDeliveries(tt.input)
			if len(got) != tt.want {
				t.Errorf("parseWaterDeliveries(%q) = %d deliveries, want %d", tt.input, len(got), tt.want)
			}
			if len(got) > 0 && tt.wantStreet != "" {
				if got[0].Street != tt.wantStreet {
					t.Errorf("parseWaterDeliveries(%q)[0].Street = %q, want %q", tt.input, got[0].Street, tt.wantStreet)
				}
				if got[0].Buildings != tt.wantBuildings {
					t.Errorf(
						"parseWaterDeliveries(%q)[0].Buildings = %q, want %q",
						tt.input,
						got[0].Buildings,
						tt.wantBuildings,
					)
				}
				if got[0].TimeStart != tt.wantTimeStart {
					t.Errorf(
						"parseWaterDeliveries(%q)[0].TimeStart = %q, want %q",
						tt.input,
						got[0].TimeStart,
						tt.wantTimeStart,
					)
				}
				if got[0].TimeEnd != tt.wantTimeEnd {
					t.Errorf(
						"parseWaterDeliveries(%q)[0].TimeEnd = %q, want %q",
						tt.input,
						got[0].TimeEnd,
						tt.wantTimeEnd,
					)
				}
			}
		})
	}
}

func TestCityPrefixRegex(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"город Красноярск: Ленина 137", "Ленина 137"},
		{"город Красноярск:Ленина 137", "Ленина 137"},
		{"пгт Солнечный: Мира 1", "Мира 1"},
		{"Ленина 137", "Ленина 137"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := cityPrefixRe.ReplaceAllString(tt.input, "")
			if got != tt.want {
				t.Errorf("cityPrefixRe.ReplaceAll(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
