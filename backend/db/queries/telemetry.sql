-- name: InsertTelemetry :exec
INSERT INTO telemetry_events (locomotive_id, ts, metrics, raw)
VALUES ($1, $2, $3, $4);

-- name: GetLatestTelemetry :one
SELECT id, locomotive_id, ts, metrics, raw
FROM telemetry_events
WHERE locomotive_id = $1
ORDER BY ts DESC
LIMIT 1;

-- name: GetRecentTelemetry :many
-- $2 = секунды назад, например 300 = последние 5 минут
SELECT id, locomotive_id, ts, metrics, raw
FROM telemetry_events
WHERE locomotive_id = $1
  AND ts >= NOW() - ($2 || ' seconds')::interval
ORDER BY ts ASC;

-- name: GetTelemetryWindow :many
-- Агрегация по 10-секундным бакетам для графиков
-- avg_metrics: средние значения всех метрик в бакете
SELECT
    time_bucket('10 seconds', ts)   AS bucket,
    COUNT(*)                        AS sample_count,
    jsonb_object_agg(
        key,
        round(avg_val::numeric, 4)
    ) AS avg_metrics
FROM (
    SELECT
        ts,
        key,
        AVG((value::text)::float) AS avg_val
    FROM telemetry_events,
         jsonb_each(metrics) AS kv(key, value)
    WHERE locomotive_id = $1
      AND ts BETWEEN $2 AND $3
    GROUP BY time_bucket('10 seconds', ts), key
) sub
GROUP BY bucket
ORDER BY bucket;

-- name: GetMetricSeries :many
-- Временной ряд одной метрики для графика
SELECT
    ts,
    (metrics ->> $2)::float AS value
FROM telemetry_events
WHERE locomotive_id = $1
  AND ts BETWEEN $3 AND $4
  AND metrics ? $2
ORDER BY ts ASC;
