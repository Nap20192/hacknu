-- LOCOMOTIVES

-- name: ListLocomotives :many
SELECT
    id,
    display_name,
    loco_type,
    registered_at,
    last_seen_at,
    active
FROM locomotives
ORDER BY id;

-- name: ListActiveLocomotives :many
SELECT
    id,
    display_name,
    loco_type,
    registered_at,
    last_seen_at,
    active
FROM locomotives
WHERE
    active = TRUE
ORDER BY id;

-- name: GetLocomotive :one
SELECT
    id,
    display_name,
    loco_type,
    registered_at,
    last_seen_at,
    active
FROM locomotives
WHERE
    id = $1;

-- name: UpsertLocomotive :one
INSERT INTO
    locomotives (
        id,
        display_name,
        loco_type,
        last_seen_at,
        active
    )
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO
UPDATE
SET
    display_name = EXCLUDED.display_name,
    loco_type = EXCLUDED.loco_type,
    last_seen_at = EXCLUDED.last_seen_at,
    active = EXCLUDED.active
RETURNING
    id,
    display_name,
    loco_type,
    registered_at,
    last_seen_at,
    active;

-- name: UpdateLocomotiveLastSeen :one
UPDATE locomotives
SET
    last_seen_at = $2
WHERE
    id = $1
RETURNING
    id,
    display_name,
    loco_type,
    registered_at,
    last_seen_at,
    active;

-- name: SetLocomotiveActive :one
UPDATE locomotives
SET
    active = $2
WHERE
    id = $1
RETURNING
    id,
    display_name,
    loco_type,
    registered_at,
    last_seen_at,
    active;
