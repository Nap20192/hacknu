-- Enable TimescaleDB
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- =============================================================
-- METRIC DEFINITIONS
-- =============================================================
CREATE TABLE metric_definitions (
    name          TEXT    PRIMARY KEY,
    display       TEXT    NOT NULL,
    description   TEXT    NOT NULL DEFAULT '',
    unit          TEXT    NOT NULL,
    group_id      TEXT    NOT NULL,
    physical_min  FLOAT,
    physical_max  FLOAT,
    normal_min    FLOAT,
    normal_max    FLOAT,
    warn_above    FLOAT,
    warn_below    FLOAT,
    crit_above    FLOAT,
    crit_below    FLOAT,
    health_weight FLOAT   NOT NULL DEFAULT 0.0,
    ema_alpha     FLOAT   NOT NULL DEFAULT 0.1 CHECK (ema_alpha BETWEEN 0 AND 1),
    display_opts  JSONB   NOT NULL DEFAULT '{}',
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================
-- METRIC GROUPS
-- =============================================================
CREATE TABLE metric_groups (
    id    TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    icon  TEXT NOT NULL DEFAULT '',
    sort  INT  NOT NULL DEFAULT 0
);

CREATE TABLE metric_group_members (
    group_id    TEXT NOT NULL REFERENCES metric_groups(id)        ON DELETE CASCADE,
    metric_name TEXT NOT NULL REFERENCES metric_definitions(name) ON DELETE CASCADE,
    sort        INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, metric_name)
);

-- =============================================================
-- LOCOMOTIVE REGISTRY
-- =============================================================
CREATE TABLE locomotives (
    id            TEXT        PRIMARY KEY,
    display_name  TEXT        NOT NULL,
    loco_type     TEXT        NOT NULL DEFAULT 'default',
    series        TEXT        NOT NULL DEFAULT '',
    depot         TEXT        NOT NULL DEFAULT '',
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at  TIMESTAMPTZ,
    active        BOOLEAN     NOT NULL DEFAULT TRUE
);

-- =============================================================
-- TELEMETRY EVENTS  (TimescaleDB hypertable)
-- metrics JSONB: {"speed_kmh": 45.2, "engine_temp_c": 87.1, ...}
-- =============================================================
CREATE TABLE telemetry_events (
    id            BIGSERIAL,
    locomotive_id TEXT        NOT NULL,
    ts            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metrics       JSONB       NOT NULL DEFAULT '{}',
    raw           JSONB       NOT NULL DEFAULT '{}'
);

SELECT create_hypertable('telemetry_events', 'ts');

CREATE INDEX idx_tel_loco_ts    ON telemetry_events (locomotive_id, ts DESC);
CREATE INDEX idx_tel_metrics_gin ON telemetry_events USING GIN (metrics);

-- =============================================================
-- HEALTH SNAPSHOTS  (TimescaleDB hypertable)
-- =============================================================
CREATE TABLE health_snapshots (
    id            BIGSERIAL,
    locomotive_id TEXT        NOT NULL,
    ts            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    score         SMALLINT    NOT NULL CHECK (score BETWEEN 0 AND 100),
    category      TEXT        NOT NULL CHECK (category IN ('normal', 'warning', 'critical')),
    factors       JSONB       NOT NULL DEFAULT '[]',
    metrics_snap  JSONB       NOT NULL DEFAULT '{}'
);

SELECT create_hypertable('health_snapshots', 'ts');

CREATE INDEX idx_health_loco_ts ON health_snapshots (locomotive_id, ts DESC);

-- =============================================================
-- ALERTS
-- =============================================================
CREATE TABLE alerts (
    id             BIGSERIAL   PRIMARY KEY,
    locomotive_id  TEXT        NOT NULL,
    triggered_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at    TIMESTAMPTZ,
    severity       TEXT        NOT NULL CHECK (severity IN ('warning', 'critical')),
    code           TEXT        NOT NULL,
    metric_name    TEXT        REFERENCES metric_definitions(name) ON DELETE SET NULL,
    metric_value   FLOAT,
    threshold      FLOAT,
    message        TEXT        NOT NULL,
    recommendation TEXT        NOT NULL DEFAULT '',
    acknowledged   BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX idx_alerts_loco_active ON alerts (locomotive_id, triggered_at DESC)
    WHERE resolved_at IS NULL;
CREATE INDEX idx_alerts_loco_ts ON alerts (locomotive_id, triggered_at DESC);
