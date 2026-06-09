package domain

import (
	"strings"
	"time"
)

type ResourceType string

const (
	ResourceTypeColdWater   ResourceType = "Холодное водоснабжение"
	ResourceTypeHotWater    ResourceType = "Горячее водоснабжение"
	ResourceTypeElectricity ResourceType = "Электроснабжение"
	ResourceTypeGas         ResourceType = "Газоснабжение"
	ResourceTypeHeating     ResourceType = "Теплоснабжение"
)

func DetectResourceType(resource string) *ResourceType {
	for _, rt := range []ResourceType{
		ResourceTypeColdWater,
		ResourceTypeHotWater,
		ResourceTypeElectricity,
		ResourceTypeGas,
		ResourceTypeHeating,
	} {
		if strings.Contains(strings.ToLower(resource), strings.ToLower(string(rt))) {
			return &rt
		}
	}
	return nil
}

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

func (s Street) String() string {
	if len(s.Buildings) == 0 {
		return s.Name
	}
	return s.Name + " " + strings.Join(s.Buildings, ", ")
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
	Comments        string          `json:"comments,omitempty"`
}

func (d OutageDetails) Address() string {
	parts := make([]string, len(d.Streets))
	for i, s := range d.Streets {
		parts[i] = s.String()
	}
	return strings.Join(parts, "\n")
}

type Outage struct {
	Area             string           `json:"area"`
	OrganizationInfo OrganizationInfo `json:"organization_info"`
	Details          OutageDetails    `json:"details"`
	Period           []time.Time      `json:"period"`
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
