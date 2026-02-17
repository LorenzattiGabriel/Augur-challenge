package model

type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

type SearchParams struct {
	Type           string `json:"type,omitempty"`
	Value          string `json:"value,omitempty"`
	ThreatActorID  string `json:"threat_actor,omitempty"`
	CampaignID     string `json:"campaign,omitempty"`
	FirstSeenAfter string `json:"first_seen_after,omitempty"`
	LastSeenBefore string `json:"last_seen_before,omitempty"`
	Page           int    `json:"page"`
	Limit          int    `json:"limit"`
}

type SearchResult struct {
	Data       []IndicatorSearchResult `json:"data"`
	Total      int                     `json:"total"`
	Page       int                     `json:"page"`
	Limit      int                     `json:"limit"`
	TotalPages int                     `json:"total_pages"`
}

type IndicatorSearchResult struct {
	ID               string `json:"id"`
	Type             string `json:"type"`
	Value            string `json:"value"`
	Confidence       int    `json:"confidence"`
	FirstSeen        string `json:"first_seen,omitempty"`
	CampaignCount    int    `json:"campaign_count"`
	ThreatActorCount int    `json:"threat_actor_count"`
}

type TimelineParams struct {
	GroupBy   string `json:"group_by"`
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}
