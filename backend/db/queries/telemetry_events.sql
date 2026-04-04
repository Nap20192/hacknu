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

-- name: PurgeTelemetryOlderThan :execrows
-- Удаляет записи телеметрии старше переданного интервала от текущего момента.
-- Пример: передай '24 hours' чтобы удалить всё старше суток.
DELETE FROM telemetry_events
WHERE ts < NOW() - $1::interval;

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
