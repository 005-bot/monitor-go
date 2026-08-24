package outage

import (
	"context"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/005-bot/address-parser-go"
	domain "github.com/005-bot/apis-go"
)

var parentheticalRe = regexp.MustCompile(`\(.+?\)`)

var commaSplitRe = regexp.MustCompile(`,\s*`)

var cityPrefixRe = regexp.MustCompile(`^(?:город|пгт)\s+\S+?:\s*`)

const (
	unknownReasonType = "неизвестное"
	splitParts        = 2
)

type Parser struct {
	addressParser *address.Parser
}

func NewParser(addressParser *address.Parser) *Parser {
	return &Parser{
		addressParser: addressParser,
	}
}

func (p *Parser) Parse(ctx context.Context, input string) (*domain.OutageDetails, error) {
	lines := p.cleanInput(input)

	if len(lines) == 0 {
		return nil, nil //nolint:nilnil // empty input returns nil details
	}

	if len(lines) == 1 && strings.Contains(lines[0], " - ") {
		return p.parseCombinedInput(ctx, lines[0])
	}

	var streets []domain.Street
	streetIdx := 0

	for streetIdx < len(lines) {
		if strings.Contains(lines[streetIdx], " - ") {
			break
		}

		chunk := p.parseStreets(ctx, lines[streetIdx])
		streets = append(streets, chunk...)
		streetIdx++
	}

	var reason *domain.Reason
	var waterDeliveries []domain.WaterDelivery
	comments := ""

	remaining := lines[streetIdx:]
	var reasonLines []string
	for _, line := range remaining {
		if deliveries := parseWaterDeliveries(line); deliveries != nil {
			waterDeliveries = append(waterDeliveries, deliveries...)
		} else {
			reasonLines = append(reasonLines, line)
		}
	}

	if len(reasonLines) > 0 {
		reasonLine := strings.Join(reasonLines, " ")
		r := parseReason(reasonLine)
		reason = &r
	}

	if streets == nil {
		streets = []domain.Street{}
	}

	return &domain.OutageDetails{
		Streets:         streets,
		Reason:          reason,
		WaterDeliveries: waterDeliveries,
		Comments:        comments,
	}, nil
}

func (p *Parser) parseCombinedInput(ctx context.Context, line string) (*domain.OutageDetails, error) {
	reasonTypeStr, reasonDesc, found := strings.Cut(line, " - ")
	if !found {
		return nil, nil //nolint:nilnil // no reason separator
	}

	preReason := strings.TrimSpace(reasonTypeStr)
	desc := strings.TrimSpace(reasonDesc)

	streetsStr := preReason
	reasonType := preReason
	if lastSemi := strings.LastIndex(preReason, ";"); lastSemi >= 0 {
		streetsStr = strings.TrimSpace(preReason[:lastSemi])
		reasonType = strings.TrimSpace(preReason[lastSemi+1:])
	}

	streets := p.parseStreets(ctx, streetsStr)

	if streets == nil {
		streets = []domain.Street{}
	}

	return &domain.OutageDetails{
		Streets:         streets,
		Reason:          &domain.Reason{Type: reasonType, Description: desc},
		WaterDeliveries: nil,
		Comments:        "",
	}, nil
}

func (p *Parser) cleanInput(input string) []string {
	lines := strings.Split(strings.TrimSpace(input), "\n")
	result := make([]string, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" {
			result = append(result, line)
		}
	}

	return result
}

func (p *Parser) parseStreets(ctx context.Context, addressLine string) []domain.Street {
	addressLine = cityPrefixRe.ReplaceAllString(addressLine, "")
	parts := strings.Split(addressLine, ";")
	streets := make([]domain.Street, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		streetName, numbersStr := splitStreetAndNumbers(part)
		if streetName == "" {
			continue
		}

		match, err := p.addressParser.Normalize(ctx, streetName)
		if err == nil && match != nil && match.Confidence >= 0.6 {
			streetName = match.Name
		}

		buildings := processBuildingNumbers(numbersStr)

		streets = append(streets, domain.Street{
			Name:      streetName,
			Buildings: buildings,
		})
	}

	return streets
}

func splitStreetAndNumbers(streetPart string) (string, string) {
	tokens := strings.Fields(streetPart)
	for i := 1; i < len(tokens); i++ {
		currRune, _ := utf8.DecodeRuneInString(tokens[i])
		prevRune, _ := utf8.DecodeRuneInString(tokens[i-1])
		if currRune != utf8.RuneError && isDigit(currRune) && prevRune != utf8.RuneError && isAlpha(prevRune) {
			return strings.Join(tokens[:i], " "), strings.Join(tokens[i:], " ")
		}
	}
	return strings.Join(tokens, " "), ""
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAlpha(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= 'а' && r <= 'я') || (r >= 'А' && r <= 'Я')
}

func processBuildingNumbers(numbersStr string) []string {
	if numbersStr == "" {
		return nil
	}

	cleaned := parentheticalRe.ReplaceAllString(numbersStr, "")

	entries := commaSplitRe.Split(cleaned, -1)
	buildings := make([]string, 0, len(entries))

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if idx := strings.Index(entry, "("); idx >= 0 {
			entry = strings.TrimSpace(entry[:idx])
		}
		if entry == "" {
			continue
		}
		buildings = append(buildings, entry)
	}

	return buildings
}

func parseReason(line string) domain.Reason {
	reasonType, reasonDesc, found := strings.Cut(line, " - ")
	if !found {
		return domain.Reason{Type: unknownReasonType, Description: line}
	}

	return domain.Reason{
		Type:        strings.TrimSpace(reasonType),
		Description: strings.TrimSpace(reasonDesc),
	}
}

func parseWaterDeliveries(line string) []domain.WaterDelivery {
	if !strings.HasPrefix(line, "Подвоз воды: ") {
		return nil
	}

	data := strings.TrimPrefix(line, "Подвоз воды: ")
	data = strings.TrimSpace(data)
	entries := strings.Split(data, ";")
	deliveries := make([]domain.WaterDelivery, 0, len(entries))

	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}

		parts := strings.SplitN(entry, " с ", splitParts)
		if len(parts) != splitParts {
			continue
		}

		addressPart := strings.TrimSpace(parts[0])
		timePart := strings.TrimSpace(parts[1])

		lastSpace := strings.LastIndex(addressPart, " ")
		if lastSpace < 0 {
			continue
		}

		street := strings.TrimSpace(addressPart[:lastSpace])
		buildings := strings.TrimSpace(addressPart[lastSpace:])

		timeParts := strings.SplitN(timePart, " до ", splitParts)
		if len(timeParts) != splitParts {
			continue
		}

		deliveries = append(deliveries, domain.WaterDelivery{
			Street:    street,
			Buildings: buildings,
			TimeStart: strings.TrimSpace(timeParts[0]),
			TimeEnd:   strings.TrimSpace(timeParts[1]),
		})
	}

	return deliveries
}
