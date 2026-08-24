package organization

import (
	"regexp"
	"strings"

	domain "github.com/005-bot/apis-go"
)

var phonePrefixRe = regexp.MustCompile(`(?i)^\s*(?:т\.?\s*|тел\.?\s*|т:\s*)`)

var phonePrefixMidRe = regexp.MustCompile(`(?i)\s*(?:т\.\s*|тел\.\s*|т:\s*)`)

const minLines = 2

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Parse(input string) *domain.OrganizationInfo {
	lines := cleanInput(input)

	if len(lines) >= minLines {
		return p.parseMultiline(lines)
	}

	if len(lines) == 1 {
		return p.parseSingleLine(lines[0])
	}

	return nil
}

func (p *Parser) parseMultiline(lines []string) *domain.OrganizationInfo {
	resourceType := domain.DetectResourceType(lines[0])
	if resourceType == nil {
		return nil
	}
	return &domain.OrganizationInfo{
		ResourceType: resourceType,
		Resource:     lines[0],
		Organization: lines[1],
		Phones:       extractPhones(lines[2:]),
	}
}

func (p *Parser) parseSingleLine(line string) *domain.OrganizationInfo {
	loc := phonePrefixMidRe.FindStringIndex(line)

	var combo string
	var phonesStr string
	if loc != nil {
		combo = strings.TrimSpace(line[:loc[0]])
		phonesStr = strings.TrimSpace(line[loc[0]:])
	} else {
		combo = line
	}

	if combo == "" {
		return nil
	}

	resource, resourceType := detectResourceInLine(combo)
	if resource == "" || resourceType == nil {
		return nil
	}

	if !strings.HasPrefix(combo, resource) {
		return nil
	}

	org := strings.TrimSpace(strings.TrimPrefix(combo, resource))
	if org == "" {
		return nil
	}

	var phones []string
	if phonesStr != "" {
		cleaned := phonePrefixRe.ReplaceAllString(phonesStr, "")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned != "" {
			phones = []string{cleaned}
		}
	}

	return &domain.OrganizationInfo{
		ResourceType: resourceType,
		Resource:     resource,
		Organization: org,
		Phones:       phones,
	}
}

func detectResourceInLine(line string) (string, *domain.ResourceType) {
	lower := strings.ToLower(line)
	for _, rt := range []domain.ResourceType{
		domain.ResourceTypeColdWater,
		domain.ResourceTypeHotWater,
		domain.ResourceTypeElectricity,
		domain.ResourceTypeGas,
		domain.ResourceTypeHeating,
	} {
		rtLower := strings.ToLower(string(rt))
		if idx := strings.Index(lower, rtLower); idx >= 0 {
			return line[idx : idx+len(rtLower)], &rt
		}
	}
	return "", nil
}

func cleanInput(input string) []string {
	lines := strings.Split(input, "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	return result
}

func extractPhones(lines []string) []string {
	phones := make([]string, 0, len(lines))

	for _, line := range lines {
		cleaned := phonePrefixRe.ReplaceAllString(line, "")
		cleaned = strings.TrimSpace(cleaned)
		if cleaned != "" {
			phones = append(phones, cleaned)
		}
	}

	return phones
}
