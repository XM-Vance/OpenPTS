DROP TRIGGER IF EXISTS mechanism_energy_set_updated_at ON mechanism_energy_monthly;
DROP TABLE IF EXISTS mechanism_energy_monthly;

DROP TRIGGER IF EXISTS tou_rules_set_updated_at ON tou_rules;
DROP TABLE IF EXISTS tou_rules;

DROP TABLE IF EXISTS weather_forecasts;
DROP TABLE IF EXISTS weather_actuals;

DROP TRIGGER IF EXISTS weather_locations_set_updated_at ON weather_locations;
DROP TABLE IF EXISTS weather_locations;
