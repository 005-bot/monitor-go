package address

import "time"

type Match struct {
	Name           string  `json:"name"`
	NormalizedName string  `json:"normalized_name"`
	Confidence     float64 `json:"confidence"`
}

type streetRow struct {
	OsmID          int64
	NameNormalized string
	NameOriginal   string
	Region         string
	LastUpdated    time.Time
}
