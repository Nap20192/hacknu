-- ALERTS

-- name: InsertAlert :one
INSERT INTO
    alerts (
        locomotive_id,
        severity,
        code,
        metric_name,
        metric_value,
        threshold,
        message,
        recommendation
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8
    )
RETURNING
    id,
    locomotive_id,
    triggered_at,
    resolved_at,
    severity,
    code,
    metric_name,
    metric_value,
    threshold,
    message,
    recommendation,
    acknowledged;

-- name: ListActiveAlertsByLocomotive :many
SELECT
    id,
    locomotive_id,
    triggered_at,
    resolved_at,
    severity,
    code,
    metric_name,
    metric_value,
    threshold,
    message,
    recommendation,
    acknowledged
FROM alerts
WHERE
    locomotive_id = $1
    AND resolved_at IS NULL
ORDER BY triggered_at DESC;

-- name: ListAlertsByLocomotiveRange :many
SELECT
    id,
    locomotive_id,
    triggered_at,
    resolved_at,
    severity,
    code,
    metric_name,
    metric_value,
    threshold,
    message,
    recommendation,
    acknowledged
FROM alerts
WHERE
    locomotive_id = $1
    AND triggered_at >= $2
    AND triggered_at <= $3
ORDER BY triggered_at DESC;

-- name: AcknowledgeAlert :one
UPDATE alerts
SET
    acknowledged = TRUE
WHERE
    id = $1
RETURNING
    id,
    locomotive_id,
    triggered_at,
    resolved_at,
    severity,
    code,
    metric_name,
    metric_value,
    threshold,
    message,
    recommendation,
    acknowledged;

-- name: ResolveAlertNow :one
UPDATE alerts
SET
    resolved_at = NOW()
WHERE
    id = $1
RETURNING
    id,
    locomotive_id,
    triggered_at,
    resolved_at,
    severity,
    code,
    metric_name,
    metric_value,
    threshold,
    message,
    recommendation,
    acknowledged;
