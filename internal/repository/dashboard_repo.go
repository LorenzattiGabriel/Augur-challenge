package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
)

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (r *DashboardRepository) GetSummary(ctx context.Context, timeRange string) (*model.DashboardSummary, error) {
	interval := "7 days"
	switch timeRange {
	case "24h":
		interval = "24 hours"
	case "7d":
		interval = "7 days"
	case "30d":
		interval = "30 days"
	}

	summary := &model.DashboardSummary{
		TimeRange:             timeRange,
		NewIndicators:         make(map[string]int),
		IndicatorDistribution: make(map[string]int),
		TopThreatActors:       []model.ThreatActorWithCount{},
	}

	newIndicatorsQuery := fmt.Sprintf(`
		SELECT type, COUNT(*) as count
		FROM indicators
		WHERE created_at >= NOW() - INTERVAL '%s'
		GROUP BY type
	`, interval)

	rows, err := r.db.QueryContext(ctx, newIndicatorsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get new indicators: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var indicatorType string
		var count int
		if err := rows.Scan(&indicatorType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan new indicators: %w", err)
		}
		summary.NewIndicators[indicatorType] = count
	}

	activeCampaignsQuery := `SELECT COUNT(*) FROM campaigns WHERE status = 'active'`
	if err := r.db.QueryRowContext(ctx, activeCampaignsQuery).Scan(&summary.ActiveCampaigns); err != nil {
		return nil, fmt.Errorf("failed to get active campaigns: %w", err)
	}

	topActorsQuery := `
		SELECT ta.id, ta.name, ta.description, ta.country, ta.motivation,
			   ta.first_seen, ta.last_seen, ta.confidence_level,
			   ta.created_at, ta.updated_at,
			   COUNT(DISTINCT ia.indicator_id) as indicator_count
		FROM threat_actors ta
		LEFT JOIN indicator_actors ia ON ia.actor_id = ta.id
		GROUP BY ta.id, ta.name, ta.description, ta.country, ta.motivation,
				 ta.first_seen, ta.last_seen, ta.confidence_level,
				 ta.created_at, ta.updated_at
		ORDER BY indicator_count DESC
		LIMIT 5
	`

	actorRows, err := r.db.QueryContext(ctx, topActorsQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get top actors: %w", err)
	}
	defer actorRows.Close()

	for actorRows.Next() {
		var actor model.ThreatActorWithCount
		var description, country, motivation sql.NullString
		var firstSeen, lastSeen sql.NullTime

		if err := actorRows.Scan(
			&actor.ID, &actor.Name, &description, &country, &motivation,
			&firstSeen, &lastSeen, &actor.ConfidenceLevel,
			&actor.CreatedAt, &actor.UpdatedAt, &actor.IndicatorCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan actor: %w", err)
		}

		if description.Valid {
			actor.Description = description.String
		}
		if country.Valid {
			actor.Country = country.String
		}
		if motivation.Valid {
			actor.Motivation = motivation.String
		}
		if firstSeen.Valid {
			actor.FirstSeen = &firstSeen.Time
		}
		if lastSeen.Valid {
			actor.LastSeen = &lastSeen.Time
		}

		summary.TopThreatActors = append(summary.TopThreatActors, actor)
	}

	distributionQuery := `SELECT type, COUNT(*) as count FROM indicators GROUP BY type`

	distRows, err := r.db.QueryContext(ctx, distributionQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get distribution: %w", err)
	}
	defer distRows.Close()

	for distRows.Next() {
		var indicatorType string
		var count int
		if err := distRows.Scan(&indicatorType, &count); err != nil {
			return nil, fmt.Errorf("failed to scan distribution: %w", err)
		}
		summary.IndicatorDistribution[indicatorType] = count
	}

	return summary, nil
}
