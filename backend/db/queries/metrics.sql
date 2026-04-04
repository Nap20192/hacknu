-- name: UpsertMetricDefinition :exec
-- Вызывается при старте сервера, синхронизирует metrics.yaml → БД
INSERT INTO metric_definitions (
    name, display, description, unit, group_id,
    physical_min, physical_max,
    normal_min, normal_max,
    warn_above, warn_below, crit_above, crit_below,
    health_weight, ema_alpha, display_opts, updated_at
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8, $9,
    $10, $11, $12, $13,
    $14, $15, $16, NOW()
)
ON CONFLICT (name) DO UPDATE SET
    display       = EXCLUDED.display,
    description   = EXCLUDED.description,
    unit          = EXCLUDED.unit,
    group_id      = EXCLUDED.group_id,
    physical_min  = EXCLUDED.physical_min,
    physical_max  = EXCLUDED.physical_max,
    normal_min    = EXCLUDED.normal_min,
    normal_max    = EXCLUDED.normal_max,
    warn_above    = EXCLUDED.warn_above,
    warn_below    = EXCLUDED.warn_below,
    crit_above    = EXCLUDED.crit_above,
    crit_below    = EXCLUDED.crit_below,
    health_weight = EXCLUDED.health_weight,
    ema_alpha     = EXCLUDED.ema_alpha,
    display_opts  = EXCLUDED.display_opts,
    updated_at    = NOW();

-- name: GetAllMetricDefinitions :many
SELECT *
FROM metric_definitions
ORDER BY group_id, name;

-- name: GetMetricDefinition :one
SELECT *
FROM metric_definitions
WHERE name = $1;

-- name: UpsertMetricGroup :exec
INSERT INTO metric_groups (id, title, icon, sort)
VALUES ($1, $2, $3, $4)
ON CONFLICT (id) DO UPDATE SET
    title = EXCLUDED.title,
    icon  = EXCLUDED.icon,
    sort  = EXCLUDED.sort;

-- name: UpsertGroupMember :exec
INSERT INTO metric_group_members (group_id, metric_name, sort)
VALUES ($1, $2, $3)
ON CONFLICT (group_id, metric_name) DO UPDATE SET sort = EXCLUDED.sort;

-- name: GetGroupsWithMetrics :many
SELECT
    g.id,
    g.title,
    g.icon,
    g.sort,
    jsonb_agg(
        jsonb_build_object(
            'name',         m.name,
            'display',      m.display,
            'unit',         m.unit,
            'normal_min',   m.normal_min,
            'normal_max',   m.normal_max,
            'physical_min', m.physical_min,
            'physical_max', m.physical_max,
            'display_opts', m.display_opts
        ) ORDER BY mgm.sort
    ) AS metrics
FROM metric_groups g
JOIN metric_group_members mgm ON mgm.group_id = g.id
JOIN metric_definitions   m   ON m.name = mgm.metric_name
GROUP BY g.id, g.title, g.icon, g.sort
ORDER BY g.sort;
