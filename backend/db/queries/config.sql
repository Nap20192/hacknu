-- name: GetThresholds :many
SELECT id, loco_type, parameter, warn_min, warn_max, crit_min, crit_max, health_weight
FROM threshold_config
WHERE loco_type = $1
ORDER BY parameter;

-- name: UpsertThreshold :exec
INSERT INTO threshold_config (loco_type, parameter, warn_min, warn_max, crit_min, crit_max, health_weight)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (loco_type, parameter)
DO UPDATE SET
    warn_min      = EXCLUDED.warn_min,
    warn_max      = EXCLUDED.warn_max,
    crit_min      = EXCLUDED.crit_min,
    crit_max      = EXCLUDED.crit_max,
    health_weight = EXCLUDED.health_weight;
