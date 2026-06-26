package parser

import "errors"

var (
	ErrParseOrganization = errors.New("failed to parse organization info")
	ErrParseOutage       = errors.New("failed to parse outage details")
)
