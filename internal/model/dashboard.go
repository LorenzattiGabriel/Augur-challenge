package model

type DashboardSummary struct {
	TimeRange             string                 `json:"time_range"`
	NewIndicators         map[string]int         `json:"new_indicators"`
	ActiveCampaigns       int                    `json:"active_campaigns"`
	TopThreatActors       []ThreatActorWithCount `json:"top_threat_actors"`
	IndicatorDistribution map[string]int         `json:"indicator_distribution"`
}
