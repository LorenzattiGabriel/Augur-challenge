package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
)

type CampaignRepository struct {
	db *sql.DB
}

func NewCampaignRepository(db *sql.DB) *CampaignRepository {
	return &CampaignRepository{db: db}
}

func (r *CampaignRepository) GetByID(ctx context.Context, id string) (*model.Campaign, error) {
	query := `
		SELECT id, name, description, status, start_date, end_date,
			   target_sectors, target_regions, threat_actor_id, severity,
			   created_at, updated_at
		FROM campaigns
		WHERE id = $1
	`

	var campaign model.Campaign
	var startDate, endDate sql.NullTime
	var targetSectors, targetRegions, threatActorID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&campaign.ID, &campaign.Name, &campaign.Description, &campaign.Status,
		&startDate, &endDate, &targetSectors, &targetRegions,
		&threatActorID, &campaign.Severity, &campaign.CreatedAt, &campaign.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	if startDate.Valid {
		campaign.StartDate = &startDate.Time
	}
	if endDate.Valid {
		campaign.EndDate = &endDate.Time
	}
	if threatActorID.Valid {
		campaign.ThreatActorID = threatActorID.String
	}

	return &campaign, nil
}

func (r *CampaignRepository) GetIndicatorsTimeline(ctx context.Context, campaignID string, params model.TimelineParams) (*model.CampaignWithTimeline, error) {
	campaign, err := r.GetByID(ctx, campaignID)
	if err != nil {
		return nil, err
	}

	dateTrunc := "day"
	if params.GroupBy == "week" {
		dateTrunc = "week"
	}

	query := fmt.Sprintf(`
		SELECT
			DATE_TRUNC('%s', COALESCE(i.first_seen, ic.added_at)) as period,
			i.id, i.type, i.value
		FROM indicators i
		JOIN indicator_campaigns ic ON ic.indicator_id = i.id
		WHERE ic.campaign_id = $1
	`, dateTrunc)

	args := []interface{}{campaignID}
	argIdx := 2

	if params.StartDate != "" {
		query += fmt.Sprintf(" AND COALESCE(i.first_seen, ic.added_at) >= $%d", argIdx)
		args = append(args, params.StartDate)
		argIdx++
	}
	if params.EndDate != "" {
		query += fmt.Sprintf(" AND COALESCE(i.first_seen, ic.added_at) <= $%d", argIdx)
		args = append(args, params.EndDate)
		argIdx++
	}

	query += " ORDER BY period DESC, i.id"

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get timeline: %w", err)
	}
	defer rows.Close()

	periodMap := make(map[string]*model.TimelinePeriod)
	var periods []string

	for rows.Next() {
		var period time.Time
		var ind model.TimelineIndicator

		if err := rows.Scan(&period, &ind.ID, &ind.Type, &ind.Value); err != nil {
			return nil, fmt.Errorf("failed to scan timeline row: %w", err)
		}

		periodStr := period.Format("2006-01-02")
		if _, exists := periodMap[periodStr]; !exists {
			periodMap[periodStr] = &model.TimelinePeriod{
				Period:     periodStr,
				Indicators: []model.TimelineIndicator{},
				Counts:     make(map[string]int),
			}
			periods = append(periods, periodStr)
		}

		periodMap[periodStr].Indicators = append(periodMap[periodStr].Indicators, ind)
		periodMap[periodStr].Counts[string(ind.Type)]++
	}

	var timeline []model.TimelinePeriod
	for _, p := range periods {
		timeline = append(timeline, *periodMap[p])
	}

	summaryQuery := `
		SELECT
			COUNT(DISTINCT i.id) as total_indicators,
			COUNT(DISTINCT CASE WHEN i.type = 'ip' THEN i.id END) as unique_ips,
			COUNT(DISTINCT CASE WHEN i.type = 'domain' THEN i.id END) as unique_domains,
			COALESCE(
				EXTRACT(DAY FROM MAX(COALESCE(i.first_seen, ic.added_at)) - MIN(COALESCE(i.first_seen, ic.added_at))),
				0
			)::INTEGER as duration_days
		FROM indicators i
		JOIN indicator_campaigns ic ON ic.indicator_id = i.id
		WHERE ic.campaign_id = $1
	`

	var summary model.TimelineSummary
	err = r.db.QueryRowContext(ctx, summaryQuery, campaignID).Scan(
		&summary.TotalIndicators, &summary.UniqueIPs,
		&summary.UniqueDomains, &summary.DurationDays,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	var firstSeen, lastSeen sql.NullTime
	seenQuery := `
		SELECT MIN(COALESCE(i.first_seen, ic.added_at)), MAX(COALESCE(i.last_seen, ic.added_at))
		FROM indicators i
		JOIN indicator_campaigns ic ON ic.indicator_id = i.id
		WHERE ic.campaign_id = $1
	`
	r.db.QueryRowContext(ctx, seenQuery, campaignID).Scan(&firstSeen, &lastSeen)

	campaignDetail := model.CampaignDetail{
		ID:          campaign.ID,
		Name:        campaign.Name,
		Description: campaign.Description,
		Status:      campaign.Status,
	}
	if firstSeen.Valid {
		campaignDetail.FirstSeen = &firstSeen.Time
	}
	if lastSeen.Valid {
		campaignDetail.LastSeen = &lastSeen.Time
	}

	return &model.CampaignWithTimeline{
		Campaign: campaignDetail,
		Timeline: timeline,
		Summary:  summary,
	}, nil
}
