DELETE FROM metric_definitions WHERE name IN (
    'speed_kmh','engine_temp_c','brake_pressure_bar','fuel_level_pct',
    'oil_pressure_bar','traction_amps','voltage_v','axle_temp_c'
);
DELETE FROM locomotives WHERE id = '00000000-0000-0000-0000-000000000001';
