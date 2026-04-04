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

-- name: RecalculateEmaAlphaPreview :many
-- Dry-run: показывает текущий и пересчитанный alpha без изменений в БД.
-- N = phys_range / warn_band — эквивалентный период EMA.
-- alpha = 2 / (N + 1) — стандартная конвертация периода в коэффициент сглаживания.
-- Поправка health_weight: критичные метрики сглаживаются сильнее.
SELECT
    name,
    ema_alpha                                              AS current_alpha,
    ROUND(
        GREATEST(0.05, LEAST(0.90,
            CASE
                WHEN NULLIF(physical_max - physical_min, 0) IS NOT NULL
                 AND NULLIF(
                         COALESCE(warn_above, crit_above, physical_max)
                         - COALESCE(warn_below, crit_below, physical_min),
                         0
                     ) IS NOT NULL
                THEN
                    2.0 / (
                        NULLIF(physical_max - physical_min, 0)
                        / NULLIF(
                            COALESCE(warn_above, crit_above, physical_max)
                            - COALESCE(warn_below, crit_below, physical_min),
                            0
                          )
                        + 1.0
                    ) * (1.0 - health_weight * 0.3)
                ELSE 0.30
            END
        ))::numeric,
        3
    )                                                      AS new_alpha
FROM metric_definitions
ORDER BY name;

-- name: RecalculateEmaAlpha :many
-- Пересчитывает ema_alpha для всех метрик на основе их физических диапазонов,
-- предупредительных порогов и health_weight. Возвращает обновлённые строки.
WITH ranges AS (
    SELECT
        name,
        health_weight,
        NULLIF(physical_max - physical_min, 0)              AS phys_range,
        NULLIF(
            COALESCE(warn_above, crit_above, physical_max)
            - COALESCE(warn_below, crit_below, physical_min),
            0
        )                                                   AS warn_band
    FROM metric_definitions
),
alpha_calc AS (
    SELECT
        name,
        GREATEST(0.05, LEAST(0.90,
            CASE
                WHEN phys_range IS NOT NULL AND warn_band IS NOT NULL
                THEN 2.0 / (phys_range / warn_band + 1.0)
                     * (1.0 - health_weight * 0.3)
                ELSE 0.30
            END
        ))                                                  AS new_alpha
    FROM ranges
)
UPDATE metric_definitions md
SET
    ema_alpha  = ROUND(alpha_calc.new_alpha::numeric, 3),
    updated_at = NOW()
FROM alpha_calc
WHERE md.name = alpha_calc.name
RETURNING
    md.name,
    md.display,
    md.description,
    md.unit,
    md.physical_min,
    md.physical_max,
    md.normal_min,
    md.normal_max,
    md.warn_above,
    md.warn_below,
    md.crit_above,
    md.crit_below,
    md.health_weight,
    md.ema_alpha,
    md.display_opts,
    md.updated_at;
