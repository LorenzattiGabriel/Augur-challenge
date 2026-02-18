CREATE TABLE IF NOT EXISTS indicator_actors (
    indicator_id UUID NOT NULL REFERENCES indicators(id) ON DELETE CASCADE,
    actor_id UUID NOT NULL REFERENCES threat_actors(id) ON DELETE CASCADE,
    attribution_confidence INTEGER DEFAULT 50 CHECK (attribution_confidence >= 0 AND attribution_confidence <= 100),
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (indicator_id, actor_id)
);

CREATE INDEX IF NOT EXISTS idx_indicator_actors_actor ON indicator_actors(actor_id);
CREATE INDEX IF NOT EXISTS idx_indicator_actors_indicator ON indicator_actors(indicator_id);
