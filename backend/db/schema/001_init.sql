-- Enable TimescaleDB
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- =============================================================
-- METRIC DEFINITIONS
-- Мастер-список метрик, синхронизируется из metrics.yaml
-- при старте сервиса. Не требует миграций при добавлении метрик.
-- =============================================================
CREATE TABLE IF NOT EXISTS metric_definitions (
    name          TEXT    PRIMARY KEY,
    display       TEXT    NOT NULL,
    description   TEXT    NOT NULL DEFAULT '',
    unit          TEXT    NOT NULL,
    group_id      TEXT    NOT NULL,

    -- Физические границы (для валидации входящих данных)
    physical_min  FLOAT,
    physical_max  FLOAT,

    -- Нормальный рабочий диапазон (для цветовой индикации)
    normal_min    FLOAT,
    normal_max    FLOAT,

    -- Пороги алертов (NULL = не применяется)
    warn_above    FLOAT,
    warn_below    FLOAT,
    crit_above    FLOAT,
    crit_below    FLOAT,

    -- Health Index
    health_weight FLOAT   NOT NULL DEFAULT 0.0,
    ema_alpha     FLOAT   NOT NULL DEFAULT 0.1 CHECK (ema_alpha BETWEEN 0 AND 1),

    -- Настройки отображения (JSON для фронтенда)
    display_opts  JSONB   NOT NULL DEFAULT '{}',

    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =============================================================
-- TELEMETRY EVENTS  (гипертаблица TimescaleDB)
-- Одна строка = один фрейм телеметрии от локомотива.
-- metrics JSONB: {"speed_kmh": 45.2, "engine_temp_c": 87.1, ...}
-- Новая метрика появляется в JSONB без ALTER TABLE.
-- =============================================================
CREATE TABLE IF NOT EXISTS telemetry_events (
    id            BIGSERIAL,
    locomotive_id TEXT        NOT NULL,
    ts            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Обработанные значения (после EMA-сглаживания и валидации)
    metrics       JSONB       NOT NULL DEFAULT '{}',

    -- Сырые значения как пришли от устройства (для отладки/replay)
    raw           JSONB       NOT NULL DEFAULT '{}'
);

SELECT create_hypertable('telemetry_events', 'ts', if_not_exists => TRUE);

-- Поиск по locomotive_id + время
CREATE INDEX IF NOT EXISTS idx_tel_loco_ts
    ON telemetry_events (locomotive_id, ts DESC);

-- JSONB GIN-индекс для запросов по конкретным метрикам
-- Позволяет: WHERE metrics @> '{"speed_kmh": 0}' или jsonb_path_exists
CREATE INDEX IF NOT EXISTS idx_tel_metrics_gin
    ON telemetry_events USING GIN (metrics);

-- =============================================================
-- HEALTH SNAPSHOTS  (гипертаблица TimescaleDB)
-- Рассчитывается из telemetry_events каждые ~1с.
-- =============================================================
CREATE TABLE IF NOT EXISTS health_snapshots (
    id            BIGSERIAL,
    locomotive_id TEXT        NOT NULL,
    ts            TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    score         SMALLINT    NOT NULL CHECK (score BETWEEN 0 AND 100),
    category      TEXT        NOT NULL CHECK (category IN ('normal', 'warning', 'critical')),

    -- Top-N факторов снижения индекса
    -- [{name, display, value, penalty, weight, description}, ...]
    factors       JSONB       NOT NULL DEFAULT '[]',

    -- Снапшот значений метрик на момент расчёта (для истории)
    metrics_snap  JSONB       NOT NULL DEFAULT '{}'
);

SELECT create_hypertable('health_snapshots', 'ts', if_not_exists => TRUE);

CREATE INDEX IF NOT EXISTS idx_health_loco_ts
    ON health_snapshots (locomotive_id, ts DESC);

-- =============================================================
-- ALERTS
-- Активные и исторические предупреждения.
-- =============================================================
CREATE TABLE IF NOT EXISTS alerts (
    id            BIGSERIAL   PRIMARY KEY,
    locomotive_id TEXT        NOT NULL,
    triggered_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    resolved_at   TIMESTAMPTZ,

    severity      TEXT        NOT NULL CHECK (severity IN ('warning', 'critical')),
    code          TEXT        NOT NULL,   -- HIGH_TEMP | LOW_BRAKE | ...

    -- Какая метрика и какое значение вызвало алерт
    metric_name   TEXT        REFERENCES metric_definitions(name) ON DELETE SET NULL,
    metric_value  FLOAT,
    threshold     FLOAT,

    message       TEXT        NOT NULL,
    recommendation TEXT       NOT NULL DEFAULT '',
    acknowledged  BOOLEAN     NOT NULL DEFAULT FALSE
);

CREATE INDEX IF NOT EXISTS idx_alerts_loco_active
    ON alerts (locomotive_id, triggered_at DESC)
    WHERE resolved_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_alerts_loco_ts
    ON alerts (locomotive_id, triggered_at DESC);

-- =============================================================
-- LOCOMOTIVE REGISTRY
-- Список локомотивов с метаданными.
-- =============================================================
CREATE TABLE IF NOT EXISTS locomotives (
    id            TEXT        PRIMARY KEY,  -- "loco-001"
    display_name  TEXT        NOT NULL,
    loco_type     TEXT        NOT NULL DEFAULT 'default',
    series        TEXT        NOT NULL DEFAULT '',
    depot         TEXT        NOT NULL DEFAULT '',
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_seen_at  TIMESTAMPTZ,
    active        BOOLEAN     NOT NULL DEFAULT TRUE
);

-- =============================================================
-- METRIC GROUPS
-- Группировка метрик для UI панелей (из metrics.yaml).
-- =============================================================
CREATE TABLE IF NOT EXISTS metric_groups (
    id      TEXT PRIMARY KEY,
    title   TEXT NOT NULL,
    icon    TEXT NOT NULL DEFAULT '',
    sort    INT  NOT NULL DEFAULT 0
);

-- Какие метрики входят в группу
CREATE TABLE IF NOT EXISTS metric_group_members (
    group_id    TEXT    NOT NULL REFERENCES metric_groups(id)         ON DELETE CASCADE,
    metric_name TEXT    NOT NULL REFERENCES metric_definitions(name)  ON DELETE CASCADE,
    sort        INT     NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, metric_name)
);
