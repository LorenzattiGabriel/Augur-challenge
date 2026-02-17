package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func NewPostgresDB(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	schema := `
	CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

	CREATE TABLE IF NOT EXISTS threat_actors (
		id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
		name VARCHAR(255) NOT NULL UNIQUE,
		description TEXT,
		country VARCHAR(100),
		motivation VARCHAR(100),
		first_seen TIMESTAMP WITH TIME ZONE,
		last_seen TIMESTAMP WITH TIME ZONE,
		confidence_level INTEGER DEFAULT 50 CHECK (confidence_level >= 0 AND confidence_level <= 100),
		created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
	);

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

	CREATE TABLE IF NOT EXISTS indicator_campaigns (
		indicator_id UUID NOT NULL REFERENCES indicators(id) ON DELETE CASCADE,
		campaign_id UUID NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
		added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		notes TEXT,
		PRIMARY KEY (indicator_id, campaign_id)
	);

	CREATE TABLE IF NOT EXISTS indicator_actors (
		indicator_id UUID NOT NULL REFERENCES indicators(id) ON DELETE CASCADE,
		actor_id UUID NOT NULL REFERENCES threat_actors(id) ON DELETE CASCADE,
		attribution_confidence INTEGER DEFAULT 50 CHECK (attribution_confidence >= 0 AND attribution_confidence <= 100),
		added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
		PRIMARY KEY (indicator_id, actor_id)
	);

	CREATE INDEX IF NOT EXISTS idx_indicators_type ON indicators(type);
	CREATE INDEX IF NOT EXISTS idx_indicators_value ON indicators(value);
	CREATE INDEX IF NOT EXISTS idx_indicators_first_seen ON indicators(first_seen);
	CREATE INDEX IF NOT EXISTS idx_indicators_last_seen ON indicators(last_seen);
	CREATE INDEX IF NOT EXISTS idx_indicators_is_active ON indicators(is_active);
	CREATE INDEX IF NOT EXISTS idx_indicators_confidence ON indicators(confidence);
	CREATE INDEX IF NOT EXISTS idx_indicators_created_at ON indicators(created_at);
	CREATE INDEX IF NOT EXISTS idx_campaigns_status ON campaigns(status);
	CREATE INDEX IF NOT EXISTS idx_campaigns_threat_actor ON campaigns(threat_actor_id);
	CREATE INDEX IF NOT EXISTS idx_campaigns_start_date ON campaigns(start_date);
	CREATE INDEX IF NOT EXISTS idx_indicator_campaigns_campaign ON indicator_campaigns(campaign_id);
	CREATE INDEX IF NOT EXISTS idx_indicator_campaigns_indicator ON indicator_campaigns(indicator_id);
	CREATE INDEX IF NOT EXISTS idx_indicator_actors_actor ON indicator_actors(actor_id);
	CREATE INDEX IF NOT EXISTS idx_indicator_actors_indicator ON indicator_actors(indicator_id);
	CREATE INDEX IF NOT EXISTS idx_actors_name ON threat_actors(name);
	CREATE INDEX IF NOT EXISTS idx_indicators_type_active ON indicators(type, is_active);
	CREATE INDEX IF NOT EXISTS idx_indicators_type_created ON indicators(type, created_at);
	`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}
