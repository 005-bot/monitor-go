package domain

import (
	"time"

	domain "github.com/005-bot/apis-go"
)

// Record is a raw scraping record produced by the scraper before parsing.
type Record struct {
	Area         string      `json:"area"`
	Organization string      `json:"organization"`
	Address      string      `json:"address"`
	Dates        []time.Time `json:"dates"`
}

// ParsedRecord is a parsed scraping record ready for storage and publishing.
type ParsedRecord struct {
	Area         string                  `json:"area"`
	Organization domain.OrganizationInfo `json:"organization"`
	Details      domain.OutageDetails    `json:"details"`
	Dates        []time.Time             `json:"dates"`
}
