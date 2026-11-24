package organization_test

import (
	"testing"

	domain "github.com/005-bot/apis-go"
	"github.com/005-bot/monitor-go/internal/parser/organization"
)

func assertOrganization(t *testing.T, got, want *domain.OrganizationInfo, wantNil bool) {
	t.Helper()

	if wantNil {
		if got != nil {
			t.Fatalf("Parse() = %+v, want nil", got)
		}
		return
	}
	if got == nil {
		t.Fatal("Parse() = nil, want non-nil")
	}
	if got.Resource != want.Resource {
		t.Errorf("Resource = %q, want %q", got.Resource, want.Resource)
	}
	if got.Organization != want.Organization {
		t.Errorf("Organization = %q, want %q", got.Organization, want.Organization)
	}
	if len(got.Phones) != len(want.Phones) {
		t.Errorf("Phones = %v, want %v", got.Phones, want.Phones)
	} else {
		for i, p := range got.Phones {
			if p != want.Phones[i] {
				t.Errorf("Phone[%d] = %q, want %q", i, p, want.Phones[i])
			}
		}
	}
	if (got.ResourceType == nil) != (want.ResourceType == nil) {
		t.Errorf("ResourceType nil-ness mismatch: got %v, want %v", got.ResourceType, want.ResourceType)
	} else if got.ResourceType != nil && *got.ResourceType != *want.ResourceType {
		t.Errorf("ResourceType = %q, want %q", *got.ResourceType, *want.ResourceType)
	}
}

func TestParse(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    *domain.OrganizationInfo
		wantNil bool
	}{
		{
			name:  "hot water with multiple phones",
			input: "Горячее водоснабжение\nАО КТТК\nт. 264-18-62\nт. 214-93-51",
			want: &domain.OrganizationInfo{
				ResourceType: new(domain.ResourceTypeHotWater),
				Resource:     "Горячее водоснабжение",
				Organization: "АО КТТК",
				Phones:       []string{"264-18-62", "214-93-51"},
			},
		},
		{
			name:  "electricity with one phone",
			input: "Электроснабжение\nПАО Россети Сибирь\nт. 8-800-220-0-220",
			want: &domain.OrganizationInfo{
				ResourceType: new(domain.ResourceTypeElectricity),
				Resource:     "Электроснабжение",
				Organization: "ПАО Россети Сибирь",
				Phones:       []string{"8-800-220-0-220"},
			},
		},
		{
			name:  "hot water with parenthetical org name",
			input: "Горячее водоснабжение с подающего трубопровода\nООО СибЭР (№ 980)\nт. 214-93-51",
			want: &domain.OrganizationInfo{
				ResourceType: new(domain.ResourceTypeHotWater),
				Resource:     "Горячее водоснабжение с подающего трубопровода",
				Organization: "ООО СибЭР (№ 980)",
				Phones:       []string{"214-93-51"},
			},
		},
		{
			name:  "cold water with no phones",
			input: "Холодное водоснабжение\nАО Красмаш",
			want: &domain.OrganizationInfo{
				ResourceType: new(domain.ResourceTypeColdWater),
				Resource:     "Холодное водоснабжение",
				Organization: "АО Красмаш",
				Phones:       []string{},
			},
		},
		{
			name:    "empty input",
			input:   "",
			wantNil: true,
		},
		{
			name:    "single line",
			input:   "Горячее водоснабжение",
			wantNil: true,
		},
		{
			name:  "single line cold water with phone",
			input: "Холодное водоснабжение ООО КрасРСК т. 2113963",
			want: &domain.OrganizationInfo{
				ResourceType: new(domain.ResourceTypeColdWater),
				Resource:     "Холодное водоснабжение",
				Organization: "ООО КрасРСК",
				Phones:       []string{"2113963"},
			},
		},
		{
			name:  "single line electricity with phone",
			input: "Электроснабжение ПАО Россети т. 8-800-220-0-220",
			want: &domain.OrganizationInfo{
				ResourceType: new(domain.ResourceTypeElectricity),
				Resource:     "Электроснабжение",
				Organization: "ПАО Россети",
				Phones:       []string{"8-800-220-0-220"},
			},
		},
		{
			name:  "single line hot water with extra resource text",
			input: "Горячее водоснабжение с подающего трубопровода ООО СибЭР (№ 980) т. 214-93-51",
			want: &domain.OrganizationInfo{
				ResourceType: new(domain.ResourceTypeHotWater),
				Resource:     "Горячее водоснабжение",
				Organization: "с подающего трубопровода ООО СибЭР (№ 980)",
				Phones:       []string{"214-93-51"},
			},
		},
	}

	p := organization.NewParser()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assertOrganization(t, p.Parse(tt.input), tt.want, tt.wantNil)
		})
	}
}
