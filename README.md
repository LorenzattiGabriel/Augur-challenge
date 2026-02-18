# Threat Intelligence Dashboard API

REST API in Go for a threat intelligence dashboard. Query indicators of compromise (IoCs), threat campaigns, and malicious actors.

## Tech Stack

- **Language**: Go 1.22
- **Router**: Chi v5
- **Database**: PostgreSQL 16
- **Cache**: Ristretto (in-memory)
- **Rate Limiting**: httprate
- **Documentation**: OpenAPI 3.0

## Prerequisites

- Go 1.22+
- Docker and Docker Compose
- PostgreSQL 16 (or use Docker)

## Quick Start

### With Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/LorenzattiGabriel/threat-intel-api.git
cd threat-intel-api

# Start all services
make docker-up

# Seed the database with test data (10K indicators)
make seed

# View logs
make docker-logs
```

The API will be available at `http://localhost:8080`
Swagger UI at `http://localhost:8081`

### Local Development

```bash
# Install dependencies
go mod download

# Start PostgreSQL
docker-compose up -d postgres

# Run migrations and seed
make seed

# Start the API
make run
```

## API Endpoints

### 1. GET /api/indicators/{id}

Get detailed information about a specific indicator including relationships.

```bash
curl http://localhost:8080/api/indicators/550e8400-e29b-41d4-a716-446655440000
```

**Response:**
```json
{
  "success": true,
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "type": "ip",
    "value": "192.168.1.100",
    "confidence": 85,
    "first_seen": "2024-11-15T10:30:00Z",
    "last_seen": "2024-12-20T14:22:00Z",
    "threat_actors": [
      {"id": "actor-123", "name": "APT-Dragon", "confidence": 90}
    ],
    "campaigns": [
      {"id": "camp-456", "name": "Operation ShadowNet", "active": true}
    ],
    "related_indicators": [
      {"id": "uuid", "type": "domain", "value": "malicious.example.com", "relationship": "same_campaign"}
    ]
  }
}
```

### 2. GET /api/indicators/search

Search indicators with filters and pagination.

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| type | string | ip, domain, url, hash |
| value | string | Partial match on value |
| threat_actor | uuid | Filter by actor |
| campaign | uuid | Filter by campaign |
| first_seen_after | date | ISO date |
| last_seen_before | date | ISO date |
| page | int | Page number (default: 1) |
| limit | int | Results per page (default: 20, max: 100) |

```bash
curl "http://localhost:8080/api/indicators/search?type=ip&page=1&limit=20"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "data": [
      {
        "id": "uuid",
        "type": "ip",
        "value": "10.0.0.1",
        "confidence": 75,
        "first_seen": "2024-10-01T08:00:00Z",
        "campaign_count": 2,
        "threat_actor_count": 1
      }
    ],
    "total": 156,
    "page": 1,
    "limit": 20,
    "total_pages": 8
  }
}
```

### 3. GET /api/campaigns/{id}/indicators

Get campaign indicators organized in a timeline.

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| group_by | string | day or week (default: day) |
| start_date | date | Start date |
| end_date | date | End date |

```bash
curl "http://localhost:8080/api/campaigns/camp-456/indicators?group_by=day"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "campaign": {
      "id": "camp-456",
      "name": "Operation ShadowNet",
      "status": "active"
    },
    "timeline": [
      {
        "period": "2024-10-01",
        "indicators": [{"id": "uuid", "type": "ip", "value": "10.0.0.1"}],
        "counts": {"ip": 5, "domain": 3, "url": 12}
      }
    ],
    "summary": {
      "total_indicators": 234,
      "unique_ips": 45,
      "unique_domains": 67,
      "duration_days": 75
    }
  }
}
```

### 4. GET /api/dashboard/summary

High-level statistics for the dashboard.

**Parameters:**
| Parameter | Type | Description |
|-----------|------|-------------|
| time_range | string | 24h, 7d, 30d (default: 7d) |

```bash
curl "http://localhost:8080/api/dashboard/summary?time_range=7d"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "time_range": "7d",
    "new_indicators": {"ip": 145, "domain": 89, "url": 234, "hash": 67},
    "active_campaigns": 12,
    "top_threat_actors": [
      {"id": "actor-123", "name": "APT-Dragon", "indicator_count": 456}
    ],
    "indicator_distribution": {"ip": 3421, "domain": 2876, "url": 2134, "hash": 1569}
  }
}
```

## Optimized SQL Query Examples

### Query 1: Indicator with Relations (Avoiding N+1)

```sql
-- Instead of multiple queries, we use correlated subqueries
SELECT
    i.id, i.type, i.value, i.confidence, i.first_seen, i.last_seen,
    -- Campaigns as JSON array in a subquery
    COALESCE(
        (SELECT json_agg(json_build_object('id', c.id, 'name', c.name, 'active', c.status = 'active'))
         FROM campaigns c
         JOIN indicator_campaigns ic ON ic.campaign_id = c.id
         WHERE ic.indicator_id = i.id),
        '[]'
    ) as campaigns,
    -- Actors as JSON array
    COALESCE(
        (SELECT json_agg(json_build_object('id', a.id, 'name', a.name, 'confidence', ia.attribution_confidence))
         FROM threat_actors a
         JOIN indicator_actors ia ON ia.actor_id = a.id
         WHERE ia.indicator_id = i.id),
        '[]'
    ) as actors
FROM indicators i
WHERE i.id = $1;

-- Optimization: A single query returns the indicator with all its relations
-- instead of 3 separate queries (indicator, campaigns, actors)
```

### Query 2: Dashboard Summary with Aggregations

```sql
-- All dashboard statistics in a single query
SELECT
    -- New indicators by type within time range
    json_build_object(
        'ip', COUNT(*) FILTER (WHERE type = 'ip' AND created_at >= NOW() - INTERVAL '7 days'),
        'domain', COUNT(*) FILTER (WHERE type = 'domain' AND created_at >= NOW() - INTERVAL '7 days'),
        'url', COUNT(*) FILTER (WHERE type = 'url' AND created_at >= NOW() - INTERVAL '7 days'),
        'hash', COUNT(*) FILTER (WHERE type = 'hash' AND created_at >= NOW() - INTERVAL '7 days')
    ) as new_indicators,

    -- Total distribution
    json_build_object(
        'ip', COUNT(*) FILTER (WHERE type = 'ip'),
        'domain', COUNT(*) FILTER (WHERE type = 'domain'),
        'url', COUNT(*) FILTER (WHERE type = 'url'),
        'hash', COUNT(*) FILTER (WHERE type = 'hash')
    ) as distribution
FROM indicators;

-- Optimization: Uses PostgreSQL FILTER clause to calculate multiple
-- conditional aggregations in a single table scan
```

## Architecture

```
threat-intel-api/
├── cmd/api/main.go           # Entry point
├── internal/
│   ├── config/               # Configuration via env vars
│   ├── database/             # PostgreSQL connection + migrations
│   ├── model/                # Domain structs
│   ├── repository/           # Data access layer (SQL)
│   ├── service/              # Business logic + cache
│   ├── handler/              # HTTP handlers
│   ├── middleware/           # Rate limit, logging, recovery
│   └── cache/                # In-memory cache with Ristretto
├── api/openapi.yaml          # OpenAPI specification
├── scripts/seed.go           # Script to populate test data
├── Dockerfile
├── docker-compose.yaml
└── Makefile
```

### Design Decisions

1. **Layered Architecture**: Clear separation between handlers, services, and repositories for easier testing and maintenance.

2. **Cache Strategy**:
   - Indicator detail: 2 min TTL
   - Search: 30 sec TTL
   - Dashboard: 5 min TTL (less volatile data)

3. **PostgreSQL vs SQLite**: PostgreSQL offers better concurrency, native JSONB, and advanced aggregation functions like `FILTER`.

4. **Rate Limiting**: 100 req/min per IP+endpoint using sliding window.

## Assumptions

- All indicator IDs are UUIDs
- Timestamps are stored and returned in UTC (ISO 8601 format)
- The API is stateless; authentication would be added via JWT middleware
- Search is case-insensitive for indicator values
- Cache invalidation is time-based (TTL), not event-based

## Future Improvements

With more time I would implement:

- [ ] JWT authentication with roles
- [ ] Elasticsearch for full-text search
- [ ] Redis for distributed cache
- [ ] Datadog metrics
- [ ] OpenTelemetry tracing
- [ ] GraphQL as an alternative to REST
- [ ] Bulk import/export of indicators
- [ ] MongoDB for storing raw threat intel feeds (NoSQL)
- [ ] Apache Kafka for event streaming
- [ ] WebSockets for real-time indicator updates
- [ ] S3 for  sample storage

## Tests

```bash
# Run all tests
make test

# Tests with coverage
make test-cover
```

### Test Coverage

| Layer | Tests |
|-------|-------|
| Cache | 6 tests |
| Middleware | 4 tests |
| Handlers | 19 tests |
| Services | 20 tests |
| **Total** | **49 tests** |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| SERVER_PORT | 8080 | Server port |
| DB_HOST | localhost | PostgreSQL host |
| DB_PORT | 5432 | PostgreSQL port |
| DB_USER | postgres | Database user |
| DB_PASSWORD | postgres | Database password |
| DB_NAME | threat_intel | Database name |
| RATE_LIMIT_RPM | 100 | Requests per minute |
| CACHE_MAX_SIZE_MB | 100 | Maximum cache size |

## License

MIT
