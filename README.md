# Threat Intelligence Dashboard API

REST API en Go para un dashboard de threat intelligence. Permite consultar indicadores de compromiso (IoCs), campañas de amenazas y actores maliciosos.

## Stack Tecnológico

- **Lenguaje**: Go 1.22
- **Router**: Chi v5
- **Base de datos**: PostgreSQL 16
- **Cache**: Ristretto (in-memory)
- **Rate Limiting**: httprate
- **Documentación**: OpenAPI 3.0

## Requisitos Previos

- Go 1.22+
- Docker y Docker Compose
- PostgreSQL 16 (o usar Docker)

## Inicio Rápido

### Con Docker (Recomendado)

```bash
# Clonar el repositorio
git clone https://github.com/LorenzattiGabriel/threat-intel-api.git
cd threat-intel-api

# Iniciar todos los servicios
make docker-up

# Poblar la base de datos con datos de prueba (10K indicadores)
make seed

# Ver logs
make docker-logs
```

La API estará disponible en `http://localhost:8080`
Swagger UI en `http://localhost:8081`

### Desarrollo Local

```bash
# Instalar dependencias
go mod download

# Iniciar PostgreSQL
docker-compose up -d postgres

# Ejecutar migraciones y seed
make seed

# Iniciar la API
make run
```

## Endpoints de la API

### 1. GET /api/indicators/{id}

Obtiene información detallada de un indicador específico incluyendo relaciones.

```bash
curl http://localhost:8080/api/indicators/550e8400-e29b-41d4-a716-446655440000
```

**Respuesta:**
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

Búsqueda de indicadores con filtros y paginación.

**Parámetros:**
| Parámetro | Tipo | Descripción |
|-----------|------|-------------|
| type | string | ip, domain, url, hash |
| value | string | Búsqueda parcial en valor |
| threat_actor | uuid | Filtrar por actor |
| campaign | uuid | Filtrar por campaña |
| first_seen_after | date | Fecha ISO |
| last_seen_before | date | Fecha ISO |
| page | int | Página (default: 1) |
| limit | int | Resultados por página (default: 20, max: 100) |

```bash
curl "http://localhost:8080/api/indicators/search?type=ip&page=1&limit=20"
```

**Respuesta:**
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

Indicadores de una campaña organizados en timeline.

**Parámetros:**
| Parámetro | Tipo | Descripción |
|-----------|------|-------------|
| group_by | string | day o week (default: day) |
| start_date | date | Fecha de inicio |
| end_date | date | Fecha de fin |

```bash
curl "http://localhost:8080/api/campaigns/camp-456/indicators?group_by=day"
```

**Respuesta:**
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

Estadísticas de alto nivel para el dashboard.

**Parámetros:**
| Parámetro | Tipo | Descripción |
|-----------|------|-------------|
| time_range | string | 24h, 7d, 30d (default: 7d) |

```bash
curl "http://localhost:8080/api/dashboard/summary?time_range=7d"
```

**Respuesta:**
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

## Ejemplos de Queries SQL Optimizadas

### Query 1: Indicador con Relaciones (Evitar N+1)

```sql
-- En lugar de hacer múltiples queries, usamos subqueries correlacionadas
SELECT
    i.id, i.type, i.value, i.confidence, i.first_seen, i.last_seen,
    -- Campaigns como JSON array en una subquery
    COALESCE(
        (SELECT json_agg(json_build_object('id', c.id, 'name', c.name, 'active', c.status = 'active'))
         FROM campaigns c
         JOIN indicator_campaigns ic ON ic.campaign_id = c.id
         WHERE ic.indicator_id = i.id),
        '[]'
    ) as campaigns,
    -- Actors como JSON array
    COALESCE(
        (SELECT json_agg(json_build_object('id', a.id, 'name', a.name, 'confidence', ia.attribution_confidence))
         FROM threat_actors a
         JOIN indicator_actors ia ON ia.actor_id = a.id
         WHERE ia.indicator_id = i.id),
        '[]'
    ) as actors
FROM indicators i
WHERE i.id = $1;

-- Optimización: Una sola query retorna el indicador con todas sus relaciones
-- en lugar de 3 queries separadas (indicador, campaigns, actors)
```

### Query 2: Dashboard Summary con Agregaciones

```sql
-- Todas las estadísticas del dashboard en una sola query
SELECT
    -- Nuevos indicadores por tipo en el rango de tiempo
    json_build_object(
        'ip', COUNT(*) FILTER (WHERE type = 'ip' AND created_at >= NOW() - INTERVAL '7 days'),
        'domain', COUNT(*) FILTER (WHERE type = 'domain' AND created_at >= NOW() - INTERVAL '7 days'),
        'url', COUNT(*) FILTER (WHERE type = 'url' AND created_at >= NOW() - INTERVAL '7 days'),
        'hash', COUNT(*) FILTER (WHERE type = 'hash' AND created_at >= NOW() - INTERVAL '7 days')
    ) as new_indicators,

    -- Distribución total
    json_build_object(
        'ip', COUNT(*) FILTER (WHERE type = 'ip'),
        'domain', COUNT(*) FILTER (WHERE type = 'domain'),
        'url', COUNT(*) FILTER (WHERE type = 'url'),
        'hash', COUNT(*) FILTER (WHERE type = 'hash')
    ) as distribution
FROM indicators;

-- Optimización: Usa FILTER clause de PostgreSQL para calcular múltiples
-- agregaciones condicionales en un solo scan de la tabla
```

## Arquitectura

```
threat-intel-api/
├── cmd/api/main.go           # Entry point
├── internal/
│   ├── config/               # Configuración via env vars
│   ├── database/             # Conexión PostgreSQL + migraciones
│   ├── model/                # Structs de dominio
│   ├── repository/           # Capa de acceso a datos (SQL)
│   ├── service/              # Lógica de negocio + cache
│   ├── handler/              # HTTP handlers
│   ├── middleware/           # Rate limit, logging, recovery
│   └── cache/                # Cache in-memory con Ristretto
├── api/openapi.yaml          # Especificación OpenAPI
├── scripts/seed.go           # Script para poblar datos
├── Dockerfile
├── docker-compose.yaml
└── Makefile
```

### Decisiones de Diseño

1. **Arquitectura en Capas**: Separación clara entre handlers, services y repositories para facilitar testing y mantenimiento.

2. **Cache Strategy**:
   - Indicador detalle: 2 min TTL
   - Búsqueda: 30 seg TTL
   - Dashboard: 5 min TTL (datos menos volátiles)

3. **PostgreSQL vs SQLite**: PostgreSQL ofrece mejor concurrencia, JSONB nativo, y funciones de agregación avanzadas como `FILTER`.

4. **Rate Limiting**: 100 req/min por IP+endpoint usando sliding window.

## Mejoras Futuras

Con más tiempo implementaría:

- [ ] Autenticación JWT con roles
- [ ] Websockets para actualizaciones en tiempo real
- [ ] Elasticsearch para búsqueda full-text
- [ ] Redis para cache distribuido
- [ ] Métricas con Prometheus
- [ ] Tracing con OpenTelemetry
- [ ] GraphQL como alternativa a REST
- [ ] Bulk import/export de indicadores
- [ ] Integración con feeds de threat intel (MISP, OTX)

## Tests

```bash
# Ejecutar todos los tests
make test

# Tests con coverage
make test-cover
```

## Variables de Entorno

| Variable | Default | Descripción |
|----------|---------|-------------|
| SERVER_PORT | 8080 | Puerto del servidor |
| DB_HOST | localhost | Host de PostgreSQL |
| DB_PORT | 5432 | Puerto de PostgreSQL |
| DB_USER | postgres | Usuario de DB |
| DB_PASSWORD | postgres | Contraseña de DB |
| DB_NAME | threat_intel | Nombre de la DB |
| RATE_LIMIT_RPM | 100 | Requests por minuto |
| CACHE_MAX_SIZE_MB | 100 | Tamaño máximo del cache |

## Licencia

MIT
