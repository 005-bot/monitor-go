package domain

import "time"

type ResourceType string

const (
	ResourceTypeColdWater   ResourceType = "Холодное водоснабжение"
	ResourceTypeHotWater    ResourceType = "Горячее водоснабжение"
	ResourceTypeElectricity ResourceType = "Электроснабжение"
)

type OrganizationInfo struct {
	ResourceType *ResourceType `json:"resource_type"`
	Resource     string        `json:"resource"`
	Organization string        `json:"organization"`
	Phones       []string      `json:"phones"`
}

type Street struct {
	Name      string   `json:"name"`
	Buildings []string `json:"buildings,omitempty"`
}

type Reason struct {
	Type        string `json:"type"`
	Description string `json:"description"`
}

type WaterDelivery struct {
	Street    string `json:"street"`
	Buildings string `json:"buildings"`
	TimeStart string `json:"time_start"`
	TimeEnd   string `json:"time_end"`
}

type OutageDetails struct {
	Streets         []Street        `json:"streets"`
	Reason          *Reason         `json:"reason,omitempty"`
	WaterDeliveries []WaterDelivery `json:"water_deliveries,omitempty"`
	Comments        *string         `json:"comments,omitempty"`
}

type Record struct {
	Area         string      `json:"area"`
	Organization string      `json:"organization"`
	Address      string      `json:"address"`
	Dates        []time.Time `json:"dates"`
}

type ParsedRecord struct {
	Area         string           `json:"area"`
	Organization OrganizationInfo `json:"organization"`
	Details      OutageDetails    `json:"details"`
	Dates        []time.Time      `json:"dates"`
}

type Outage struct {
	Area             string           `json:"area"`
	OrganizationInfo OrganizationInfo `json:"organization_info"`
	Details          OutageDetails    `json:"details"`
	Period           []time.Time      `json:"period"`
}
