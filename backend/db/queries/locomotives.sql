-- name: UpsertLocomotive :exec
INSERT INTO locomotives (id, display_name, loco_type, series, depot)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    loco_type    = EXCLUDED.loco_type,
    series       = EXCLUDED.series,
    depot        = EXCLUDED.depot;

-- name: TouchLocomotive :exec
UPDATE locomotives
SET last_seen_at = NOW()
WHERE id = $1;

-- name: GetLocomotive :one
SELECT * FROM locomotives WHERE id = $1;

-- name: ListLocomotives :many
SELECT * FROM locomotives WHERE active = TRUE ORDER BY id;
