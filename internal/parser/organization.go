package parser

import (
	"github.com/005-bot/monitor-go/internal/domain"
)

type OrganizationParser struct{}

func NewOrganizationParser() *OrganizationParser {
	return &OrganizationParser{}
}

func (p *OrganizationParser) Parse(input string) (*domain.OrganizationInfo, error) {
	return nil, nil
}
