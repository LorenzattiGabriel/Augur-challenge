package model

import "time"

type ThreatActor struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Description     string     `json:"description,omitempty"`
	Country         string     `json:"country,omitempty"`
	Motivation      string     `json:"motivation,omitempty"`
	FirstSeen       *time.Time `json:"first_seen,omitempty"`
	LastSeen        *time.Time `json:"last_seen,omitempty"`
	ConfidenceLevel int        `json:"confidence_level"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type ThreatActorWithCount struct {
	ThreatActor
	IndicatorCount int `json:"indicator_count"`
}
