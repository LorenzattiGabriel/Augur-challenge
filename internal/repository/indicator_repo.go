package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/LorenzattiGabriel/threat-intel-api/internal/model"
	"github.com/Masterminds/squirrel"
)

type IndicatorRepository struct {
	db *sql.DB
	sq squirrel.StatementBuilderType
}

func NewIndicatorRepository(db *sql.DB) *IndicatorRepository {
	return &IndicatorRepository{
		db: db,
		sq: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *IndicatorRepository) GetByID(ctx context.Context, id string) (*model.IndicatorWithRelations, error) {
	query := `
		SELECT
			i.id, i.type, i.value, i.description, i.severity,
			i.confidence, i.first_seen, i.last_seen, i.is_active,
			i.tags, i.metadata, i.source, i.created_at, i.updated_at
		FROM indicators i
		WHERE i.id = $1
	`

	var indicator model.IndicatorWithRelations
	var description, severity, tags, metadata, source sql.NullString
	var firstSeen, lastSeen sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&indicator.ID, &indicator.Type, &indicator.Value, &description,
		&severity, &indicator.Confidence, &firstSeen, &lastSeen,
		&indicator.IsActive, &tags, &metadata, &source,
		&indicator.CreatedAt, &indicator.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get indicator: %w", err)
	}

	if description.Valid {
		indicator.Description = description.String
	}
	if severity.Valid {
		indicator.Severity = severity.String
	}
	if source.Valid {
		indicator.Source = source.String
	}
	if firstSeen.Valid {
		indicator.FirstSeen = &firstSeen.Time
	}
	if lastSeen.Valid {
		indicator.LastSeen = &lastSeen.Time
	}
	if tags.Valid {
		json.Unmarshal([]byte(tags.String), &indicator.Tags)
	}

	actorQuery := `
		SELECT ta.id, ta.name, ia.attribution_confidence
		FROM threat_actors ta
		JOIN indicator_actors ia ON ia.actor_id = ta.id
		WHERE ia.indicator_id = $1
	`
	actorRows, err := r.db.QueryContext(ctx, actorQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get threat actors: %w", err)
	}
	defer actorRows.Close()

	for actorRows.Next() {
		var actor model.ThreatActorSummary
		if err := actorRows.Scan(&actor.ID, &actor.Name, &actor.Confidence); err != nil {
			return nil, fmt.Errorf("failed to scan threat actor: %w", err)
		}
		indicator.ThreatActors = append(indicator.ThreatActors, actor)
	}

	campaignQuery := `
		SELECT c.id, c.name, c.status = 'active' as active
		FROM campaigns c
		JOIN indicator_campaigns ic ON ic.campaign_id = c.id
		WHERE ic.indicator_id = $1
	`
	campRows, err := r.db.QueryContext(ctx, campaignQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaigns: %w", err)
	}
	defer campRows.Close()

	for campRows.Next() {
		var campaign model.CampaignSummary
		if err := campRows.Scan(&campaign.ID, &campaign.Name, &campaign.Active); err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		indicator.Campaigns = append(indicator.Campaigns, campaign)
	}

	relatedQuery := `
		SELECT DISTINCT i.id, i.type, i.value, 'same_campaign' as relationship
		FROM indicators i
		JOIN indicator_campaigns ic ON ic.indicator_id = i.id
		WHERE ic.campaign_id IN (
			SELECT campaign_id FROM indicator_campaigns WHERE indicator_id = $1
		)
		AND i.id != $1
		ORDER BY i.last_seen DESC NULLS LAST
		LIMIT 5
	`
	relRows, err := r.db.QueryContext(ctx, relatedQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get related indicators: %w", err)
	}
	defer relRows.Close()

	for relRows.Next() {
		var related model.RelatedIndicator
		if err := relRows.Scan(&related.ID, &related.Type, &related.Value, &related.Relationship); err != nil {
			return nil, fmt.Errorf("failed to scan related indicator: %w", err)
		}
		indicator.RelatedIndicators = append(indicator.RelatedIndicators, related)
	}

	return &indicator, nil
}

func (r *IndicatorRepository) Search(ctx context.Context, params model.SearchParams) (*model.SearchResult, error) {
	baseQuery := r.sq.Select(
		"i.id", "i.type", "i.value", "i.confidence",
		"i.first_seen",
		"COUNT(DISTINCT ic.campaign_id) as campaign_count",
		"COUNT(DISTINCT ia.actor_id) as threat_actor_count",
	).
		From("indicators i").
		LeftJoin("indicator_campaigns ic ON ic.indicator_id = i.id").
		LeftJoin("indicator_actors ia ON ia.indicator_id = i.id").
		GroupBy("i.id", "i.type", "i.value", "i.confidence", "i.first_seen")

	if params.Type != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"i.type": params.Type})
	}
	if params.Value != "" {
		baseQuery = baseQuery.Where(squirrel.ILike{"i.value": "%" + params.Value + "%"})
	}
	if params.ThreatActorID != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"ia.actor_id": params.ThreatActorID})
	}
	if params.CampaignID != "" {
		baseQuery = baseQuery.Where(squirrel.Eq{"ic.campaign_id": params.CampaignID})
	}
	if params.FirstSeenAfter != "" {
		baseQuery = baseQuery.Where(squirrel.GtOrEq{"i.first_seen": params.FirstSeenAfter})
	}
	if params.LastSeenBefore != "" {
		baseQuery = baseQuery.Where(squirrel.LtOrEq{"i.last_seen": params.LastSeenBefore})
	}

	countQuery := r.sq.Select("COUNT(DISTINCT i.id)").
		From("indicators i").
		LeftJoin("indicator_campaigns ic ON ic.indicator_id = i.id").
		LeftJoin("indicator_actors ia ON ia.indicator_id = i.id")

	if params.Type != "" {
		countQuery = countQuery.Where(squirrel.Eq{"i.type": params.Type})
	}
	if params.Value != "" {
		countQuery = countQuery.Where(squirrel.ILike{"i.value": "%" + params.Value + "%"})
	}
	if params.ThreatActorID != "" {
		countQuery = countQuery.Where(squirrel.Eq{"ia.actor_id": params.ThreatActorID})
	}
	if params.CampaignID != "" {
		countQuery = countQuery.Where(squirrel.Eq{"ic.campaign_id": params.CampaignID})
	}
	if params.FirstSeenAfter != "" {
		countQuery = countQuery.Where(squirrel.GtOrEq{"i.first_seen": params.FirstSeenAfter})
	}
	if params.LastSeenBefore != "" {
		countQuery = countQuery.Where(squirrel.LtOrEq{"i.last_seen": params.LastSeenBefore})
	}

	countSQL, countArgs, err := countQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build count query: %w", err)
	}

	var total int
	if err := r.db.QueryRowContext(ctx, countSQL, countArgs...).Scan(&total); err != nil {
		return nil, fmt.Errorf("failed to get total count: %w", err)
	}

	offset := (params.Page - 1) * params.Limit
	baseQuery = baseQuery.
		OrderBy("i.created_at DESC").
		Limit(uint64(params.Limit)).
		Offset(uint64(offset))

	querySQL, queryArgs, err := baseQuery.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build search query: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, querySQL, queryArgs...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}
	defer rows.Close()

	var results []model.IndicatorSearchResult
	for rows.Next() {
		var r model.IndicatorSearchResult
		var firstSeen sql.NullTime

		if err := rows.Scan(
			&r.ID, &r.Type, &r.Value, &r.Confidence,
			&firstSeen, &r.CampaignCount, &r.ThreatActorCount,
		); err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}

		if firstSeen.Valid {
			r.FirstSeen = firstSeen.Time.Format(time.RFC3339)
		}
		results = append(results, r)
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &model.SearchResult{
		Data:       results,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

func (r *IndicatorRepository) GetIndicatorsByIDs(ctx context.Context, ids []string) ([]model.Indicator, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, type, value, description, severity, confidence,
			   first_seen, last_seen, is_active, tags, metadata, source,
			   created_at, updated_at
		FROM indicators
		WHERE id IN (%s)
	`, strings.Join(placeholders, ", "))

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get indicators by IDs: %w", err)
	}
	defer rows.Close()

	var indicators []model.Indicator
	for rows.Next() {
		var ind model.Indicator
		var tags, metadata sql.NullString
		var firstSeen, lastSeen sql.NullTime

		if err := rows.Scan(
			&ind.ID, &ind.Type, &ind.Value, &ind.Description,
			&ind.Severity, &ind.Confidence, &firstSeen, &lastSeen,
			&ind.IsActive, &tags, &metadata, &ind.Source,
			&ind.CreatedAt, &ind.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan indicator: %w", err)
		}

		if firstSeen.Valid {
			ind.FirstSeen = &firstSeen.Time
		}
		if lastSeen.Valid {
			ind.LastSeen = &lastSeen.Time
		}
		if tags.Valid {
			json.Unmarshal([]byte(tags.String), &ind.Tags)
		}

		indicators = append(indicators, ind)
	}

	return indicators, nil
}
