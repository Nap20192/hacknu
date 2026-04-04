-- Enable TimescaleDB if available (gracefully skipped on plain PostgreSQL)
DO $$
BEGIN
    CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'TimescaleDB not available, skipping: %', SQLERRM;
END
$$;

-- =============================================================
-- METRIC DEFINITIONS
-- =============================================================
CREATE TABLE metric_definitions (
    name TEXT PRIMARY KEY,
    display TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    unit TEXT NOT NULL,
    physical_min real,
    physical_max real,
    normal_min real,
    normal_max real,
    warn_above real,
    warn_below real,
    crit_above real,
    crit_below real,
    health_weight real NOT NULL DEFAULT 0.0,
    ema_alpha real NOT NULL DEFAULT 0.1 CHECK (ema_alpha BETWEEN 0 AND 1),
    display_opts JSONB NOT NULL DEFAULT '{}',
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================
-- LOCOMOTIVE REGISTRY
-- =============================================================
CREATE TABLE locomotives (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    display_name TEXT NOT NULL,
    loco_type TEXT NOT NULL DEFAULT 'default',
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at TIMESTAMPTZ,
    active BOOLEAN NOT NULL DEFAULT TRUE
);

-- =============================================================
-- TELEMETRY EVENTS  (TimescaleDB hypertable)
-- metrics JSONB: {"speed_kmh": 45.2, "engine_temp_c": 87.1, ...}
-- =============================================================
CREATE TABLE telemetry_events (
    id BIGSERIAL NOT NULL,
    locomotive_id UUID NOT NULL,
    ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metrics JSONB NOT NULL DEFAULT '{}',
    raw JSONB NOT NULL DEFAULT '{}'
);

-- sqlc:ignore
DO $$
BEGIN
    PERFORM create_hypertable('telemetry_events', 'ts', if_not_exists => TRUE);
EXCEPTION
    WHEN undefined_function THEN
        RAISE NOTICE 'create_hypertable not available, skipping';
END
$$;

-- TimescaleDB: уникальные индексы обязаны включать колонку партиционирования (ts)
ALTER TABLE telemetry_events ADD PRIMARY KEY (id, ts);

CREATE INDEX idx_tel_loco_ts ON telemetry_events (locomotive_id, ts DESC);

CREATE INDEX idx_tel_metrics_gin ON telemetry_events USING GIN (metrics);

-- =============================================================
-- HEALTH SNAPSHOTS  (TimescaleDB hypertable)
-- =============================================================
CREATE TABLE health_snapshots (
    id BIGSERIAL NOT NULL,
    locomotive_id UUID NOT NULL,
    ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    score SMALLINT NOT NULL CHECK (score BETWEEN 0 AND 100),
    category TEXT NOT NULL CHECK (
        category IN (
            'normal',
            'warning',
            'critical'
        )
    ),
    factors JSONB NOT NULL DEFAULT '[]',
    metrics_snap JSONB NOT NULL DEFAULT '{}'
);

-- sqlc:ignore
DO $$
BEGIN
    PERFORM create_hypertable('health_snapshots', 'ts', if_not_exists => TRUE);
EXCEPTION
    WHEN undefined_function THEN
        RAISE NOTICE 'create_hypertable not available, skipping';
END
$$;

ALTER TABLE health_snapshots ADD PRIMARY KEY (id, ts);

CREATE INDEX idx_health_loco_ts ON health_snapshots (locomotive_id, ts DESC);

-- =============================================================
-- ALERTS
-- =============================================================
CREATE TABLE alerts (
    id BIGSERIAL PRIMARY KEY,
    locomotive_id UUID NOT NULL,
    triggered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at TIMESTAMPTZ,
    severity TEXT NOT NULL CHECK (
        severity IN ('warning', 'critical')
    ),
    code TEXT NOT NULL,
    metric_name TEXT REFERENCES metric_definitions (name) ON DELETE SET NULL,
    metric_value real,
    threshold real,
    message TEXT NOT NULL,
    recommendation TEXT NOT NULL DEFAULT '',
    acknowledged BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_alerts_loco_active ON alerts (
    locomotive_id,
    triggered_at DESC
)
WHERE
    resolved_at IS NULL;

CREATE INDEX idx_alerts_loco_ts ON alerts (
    locomotive_id,
    triggered_at DESC
);
