-- HEALTH SNAPSHOTS

-- name: InsertHealthSnapshot :one

INSERT INTO
    health_snapshots (
        locomotive_id,
        ts,
        score,
        category,
        factors,
        metrics_snap
    )
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING
    id,
    locomotive_id,
    ts,
    score,
    category,
    factors,
    metrics_snap;

-- name: GetLatestHealthSnapshot :one
SELECT
    id,
    locomotive_id,
    ts,
    score,
    category,
    factors,
    metrics_snap
FROM health_snapshots
WHERE
    locomotive_id = $1
ORDER BY ts DESC
LIMIT 1;

-- name: ListHealthSnapshotsLatest :many
SELECT
    id,
    locomotive_id,
    ts,
    score,
    category,
    factors,
    metrics_snap
FROM health_snapshots
WHERE
    locomotive_id = $1
ORDER BY ts DESC
LIMIT $2;

-- name: ListHealthSnapshotsRange :many
SELECT
    id,
    locomotive_id,
    ts,
    score,
    category,
    factors,
    metrics_snap
FROM health_snapshots
WHERE
    locomotive_id = $1
    AND ts >= $2
    AND ts <= $3
ORDER BY ts ASC;
