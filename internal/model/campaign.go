package model

import "time"

type Campaign struct {
	ID            string     `json:"id"`
	Name          string     `json:"name"`
	Description   string     `json:"description,omitempty"`
	Status        string     `json:"status"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	TargetSectors []string   `json:"target_sectors,omitempty"`
	TargetRegions []string   `json:"target_regions,omitempty"`
	ThreatActorID string     `json:"threat_actor_id,omitempty"`
	Severity      string     `json:"severity,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

type CampaignWithTimeline struct {
	Campaign CampaignDetail   `json:"campaign"`
	Timeline []TimelinePeriod `json:"timeline"`
	Summary  TimelineSummary  `json:"summary"`
}

type CampaignDetail struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description,omitempty"`
	FirstSeen   *time.Time `json:"first_seen,omitempty"`
	LastSeen    *time.Time `json:"last_seen,omitempty"`
	Status      string     `json:"status"`
}

type TimelinePeriod struct {
	Period     string              `json:"period"`
	Indicators []TimelineIndicator `json:"indicators"`
	Counts     map[string]int      `json:"counts"`
}

type TimelineIndicator struct {
	ID    string        `json:"id"`
	Type  IndicatorType `json:"type"`
	Value string        `json:"value"`
}

type TimelineSummary struct {
	TotalIndicators int `json:"total_indicators"`
	UniqueIPs       int `json:"unique_ips"`
	UniqueDomains   int `json:"unique_domains"`
	DurationDays    int `json:"duration_days"`
}
