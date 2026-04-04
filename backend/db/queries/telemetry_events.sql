-- TELEMETRY EVENTS

-- name: InsertTelemetryEvent :one

INSERT INTO
    telemetry_events (
        locomotive_id,
        ts,
        metrics,
        raw
    )
VALUES ($1, $2, $3, $4)
RETURNING
    id,
    locomotive_id,
    ts,
    metrics,
    raw;

-- name: ListTelemetryByLocomotiveLatest :many
SELECT
    id,
    locomotive_id,
    ts,
    metrics,
    raw
FROM telemetry_events
WHERE
    locomotive_id = $1
ORDER BY ts DESC
LIMIT $2;

-- name: ListTelemetryByLocomotiveRange :many
SELECT
    id,
    locomotive_id,
    ts,
    metrics,
    raw
FROM telemetry_events
WHERE
    locomotive_id = $1
    AND ts >= $2
    AND ts <= $3
ORDER BY ts ASC;
