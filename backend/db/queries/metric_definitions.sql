-- METRIC DEFINITIONS

-- name: ListMetricDefinitions :many

SELECT
    name,
    display,
    description,
    unit,
    physical_min,
    physical_max,
    normal_min,
    normal_max,
    warn_above,
    warn_below,
    crit_above,
    crit_below,
    health_weight,
    ema_alpha,
    display_opts,
    updated_at
FROM metric_definitions
ORDER BY name;

-- name: GetMetricDefinition :one
SELECT
    name,
    display,
    description,
    unit,
    physical_min,
    physical_max,
    normal_min,
    normal_max,
    warn_above,
    warn_below,
    crit_above,
    crit_below,
    health_weight,
    ema_alpha,
    display_opts,
    updated_at
FROM metric_definitions
WHERE
    name = $1;

-- name: UpsertMetricDefinition :one
INSERT INTO
    metric_definitions (
        name,
        display,
        description,
        unit,
        physical_min,
        physical_max,
        normal_min,
        normal_max,
        warn_above,
        warn_below,
        crit_above,
        crit_below,
        health_weight,
        ema_alpha,
        display_opts,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        NOW()
    )
ON CONFLICT (name) DO
UPDATE
SET
    display = EXCLUDED.display,
    description = EXCLUDED.description,
    unit = EXCLUDED.unit,
    physical_min = EXCLUDED.physical_min,
    physical_max = EXCLUDED.physical_max,
    normal_min = EXCLUDED.normal_min,
    normal_max = EXCLUDED.normal_max,
    warn_above = EXCLUDED.warn_above,
    warn_below = EXCLUDED.warn_below,
    crit_above = EXCLUDED.crit_above,
    crit_below = EXCLUDED.crit_below,
    health_weight = EXCLUDED.health_weight,
    ema_alpha = EXCLUDED.ema_alpha,
    display_opts = EXCLUDED.display_opts,
    updated_at = NOW()
RETURNING
    name,
    display,
    description,
    unit,
    physical_min,
    physical_max,
    normal_min,
    normal_max,
    warn_above,
    warn_below,
    crit_above,
    crit_below,
    health_weight,
    ema_alpha,
    display_opts,
    updated_at;

-- name: DeleteMetricDefinition :execrows
DELETE FROM metric_definitions WHERE name = $1;
