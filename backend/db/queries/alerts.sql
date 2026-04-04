-- name: InsertAlert :one
INSERT INTO alerts (locomotive_id, severity, code, metric_name, metric_value, threshold, message, recommendation)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetActiveAlerts :many
SELECT *
FROM alerts
WHERE locomotive_id = $1
  AND resolved_at IS NULL
ORDER BY triggered_at DESC;

-- name: ResolveAlert :exec
UPDATE alerts
SET resolved_at = NOW()
WHERE id = $1 AND resolved_at IS NULL;

-- name: AcknowledgeAlert :exec
UPDATE alerts SET acknowledged = TRUE WHERE id = $1;

-- name: GetAlertHistory :many
SELECT *
FROM alerts
WHERE locomotive_id = $1
  AND triggered_at BETWEEN $2 AND $3
ORDER BY triggered_at DESC;

-- name: ResolveAlertByMetric :exec
-- Автоматически закрывает алерт когда метрика вернулась в норму
UPDATE alerts
SET resolved_at = NOW()
WHERE locomotive_id = $1
  AND metric_name   = $2
  AND code          = $3
  AND resolved_at IS NULL;
