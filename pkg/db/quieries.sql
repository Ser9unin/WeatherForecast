-- name: NewCitiesList :one
INSERT INTO cities(city, latitude, longitude, country) 
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING
RETURNING ID;

-- name: NewForecast :exec
INSERT INTO forecasts(city_id, date, temperature, weather) 
VALUES ($1, $2, $3, $4)
ON CONFLICT (city_id, date) 
DO UPDATE SET temperature = EXCLUDED.temperature, weather = EXCLUDED.weather;

-- name: CitiesList :many
SELECT id, city, latitude, longitude, country
FROM cities
ORDER BY city;

-- name: City :one
SELECT city, latitude, longitude, country
FROM cities
WHERE id = $1;

-- name: ShortFcastForCity :many
SELECT f.city_id, f.date, f.temperature
FROM forecasts f
WHERE f.city_id = $1;

-- name: FullFcastByTime :many
SELECT f.date, f.temperature, f.weather
FROM forecasts f
WHERE f.city_id = $1
ORDER BY ABS(f.date - $2)
LIMIT 2;