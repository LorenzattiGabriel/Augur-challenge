CREATE TABLE IF NOT EXISTS indicators (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    type VARCHAR(50) NOT NULL CHECK (type IN ('ip', 'domain', 'url', 'hash')),
    value VARCHAR(2048) NOT NULL,
    description TEXT,
    severity VARCHAR(50) DEFAULT 'medium' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    confidence INTEGER DEFAULT 50 CHECK (confidence >= 0 AND confidence <= 100),
    first_seen TIMESTAMP WITH TIME ZONE,
    last_seen TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    tags JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',
    source VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_indicators_type ON indicators(type);
CREATE INDEX IF NOT EXISTS idx_indicators_value ON indicators(value);
CREATE INDEX IF NOT EXISTS idx_indicators_first_seen ON indicators(first_seen);
CREATE INDEX IF NOT EXISTS idx_indicators_last_seen ON indicators(last_seen);
CREATE INDEX IF NOT EXISTS idx_indicators_is_active ON indicators(is_active);
CREATE INDEX IF NOT EXISTS idx_indicators_confidence ON indicators(confidence);
CREATE INDEX IF NOT EXISTS idx_indicators_created_at ON indicators(created_at);
CREATE INDEX IF NOT EXISTS idx_indicators_type_active ON indicators(type, is_active);
CREATE INDEX IF NOT EXISTS idx_indicators_type_created ON indicators(type, created_at);
