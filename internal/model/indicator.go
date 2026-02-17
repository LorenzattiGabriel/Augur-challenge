package model

import "time"

type IndicatorType string

const (
	IndicatorTypeIP     IndicatorType = "ip"
	IndicatorTypeDomain IndicatorType = "domain"
	IndicatorTypeURL    IndicatorType = "url"
	IndicatorTypeHash   IndicatorType = "hash"
)

type Indicator struct {
	ID          string        `json:"id"`
	Type        IndicatorType `json:"type"`
	Value       string        `json:"value"`
	Description string        `json:"description,omitempty"`
	Severity    string        `json:"severity,omitempty"`
	Confidence  int           `json:"confidence"`
	FirstSeen   *time.Time    `json:"first_seen,omitempty"`
	LastSeen    *time.Time    `json:"last_seen,omitempty"`
	IsActive    bool          `json:"is_active"`
	Tags        []string      `json:"tags,omitempty"`
	Metadata    string        `json:"metadata,omitempty"`
	Source      string        `json:"source,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

type IndicatorWithRelations struct {
	Indicator
	ThreatActors      []ThreatActorSummary `json:"threat_actors"`
	Campaigns         []CampaignSummary    `json:"campaigns"`
	RelatedIndicators []RelatedIndicator   `json:"related_indicators"`
}

type RelatedIndicator struct {
	ID           string        `json:"id"`
	Type         IndicatorType `json:"type"`
	Value        string        `json:"value"`
	Relationship string        `json:"relationship"`
}

type ThreatActorSummary struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Confidence int    `json:"confidence"`
}

type CampaignSummary struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active"`
}
