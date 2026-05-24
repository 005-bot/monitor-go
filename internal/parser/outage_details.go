package parser

import (
	"github.com/005-bot/monitor-go/internal/addressparser"
	"github.com/005-bot/monitor-go/internal/domain"
)

type OutageDetailsParser struct {
	addressParser *addressparser.AddressParser
}

func NewOutageDetailsParser(ap *addressparser.AddressParser) *OutageDetailsParser {
	return &OutageDetailsParser{addressParser: ap}
}

func (p *OutageDetailsParser) Parse(input string) (*domain.OutageDetails, error) {
	return nil, nil
}
