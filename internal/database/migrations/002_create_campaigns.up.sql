CREATE TABLE IF NOT EXISTS campaigns (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'historical')),
    start_date DATE,
    end_date DATE,
    target_sectors JSONB DEFAULT '[]',
    target_regions JSONB DEFAULT '[]',
    threat_actor_id UUID REFERENCES threat_actors(id) ON DELETE SET NULL,
    severity VARCHAR(50) DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
CREATE INDEX IF NOT EXISTS idx_campaigns_threat_actor ON campaigns(threat_actor_id);
CREATE INDEX IF NOT EXISTS idx_campaigns_start_date ON campaigns(start_date);
