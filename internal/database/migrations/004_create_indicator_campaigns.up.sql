CREATE TABLE IF NOT EXISTS indicator_campaigns (
    indicator_id UUID NOT NULL REFERENCES indicators(id) ON DELETE CASCADE,
    campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    PRIMARY KEY (indicator_id, campaign_id)
);

CREATE INDEX IF NOT EXISTS idx_indicator_campaigns_campaign ON indicator_campaigns(campaign_id);
CREATE INDEX IF NOT EXISTS idx_indicator_campaigns_indicator ON indicator_campaigns(indicator_id);
