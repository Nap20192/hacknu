-- ── Default locomotive ───────────────────────────────────────────────────────
INSERT INTO locomotives (id, display_name, loco_type, active)
VALUES ('00000000-0000-0000-0000-000000000001', 'Локомотив ТЭ-001', 'ТЭ10М', TRUE)
ON CONFLICT (id) DO NOTHING;

-- ── Metric definitions from metrics.yaml ────────────────────────────────────
INSERT INTO metric_definitions
    (name, display, description, unit, physical_min, physical_max, normal_min, normal_max,
     warn_above, warn_below, crit_above, crit_below, health_weight, ema_alpha)
VALUES
    ('speed_kmh',        'Скорость',              '', 'км/ч', 0,   300, 0,   120, 120,  NULL, 160,  NULL, 0.10, 0.20),
    ('engine_temp_c',    'Температура двигателя', '', '°C',  -40, 200, 60,   95, 95,   NULL, 110,  NULL, 0.20, 0.05),
    ('brake_pressure_bar','Давление тормозов',    '', 'бар',  0,   10, 5.0,  9.0, NULL, 4.5,  NULL, 3.0,  0.25, 0.10),
    ('fuel_level_pct',   'Уровень топлива',       '', '%',    0,  100, 20,  100,  NULL, 20,   NULL, 10,   0.15, 0.30),
    ('oil_pressure_bar', 'Давление масла',        '', 'бар',  0,   10, 3.5,  7.0, NULL, 3.0,  NULL, 2.0,  0.15, 0.10),
    ('traction_amps',    'Ток тяги',              '', 'А',    0, 1000, 0,   600,  600,  NULL, 750,  NULL, 0.10, 0.15),
    ('voltage_v',        'Напряжение',            '', 'В',   50,  200, 95,  130,  135,  90,   150,  80,   0.05, 0.20),
    ('axle_temp_c',      'Температура букс',      '', '°C',  -40, 200, -20,  60,  60,   NULL, 80,   NULL, 0.00, 0.05)
ON CONFLICT (name) DO NOTHING;
