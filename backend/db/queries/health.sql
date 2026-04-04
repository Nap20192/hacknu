-- name: InsertHealthSnapshot :exec
INSERT INTO health_snapshots (locomotive_id, ts, score, category, factors, metrics_snap)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetLatestHealth :one
SELECT id, locomotive_id, ts, score, category, factors, metrics_snap
FROM health_snapshots
WHERE locomotive_id = $1
ORDER BY ts DESC
LIMIT 1;

-- name: GetHealthHistory :many
-- Агрегация по 1-минутным бакетам для тренда индекса здоровья
SELECT
    time_bucket('1 minute', ts)  AS bucket,
    ROUND(AVG(score))::smallint  AS avg_score,
    MIN(score)                   AS min_score,
    MAX(score)                   AS max_score
FROM health_snapshots
WHERE locomotive_id = $1
  AND ts BETWEEN $2 AND $3
GROUP BY bucket
ORDER BY bucket;
