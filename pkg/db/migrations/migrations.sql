CREATE TABLE cities (
    id serial primary key,
    city VARCHAR(100),
    latitude DOUBLE PRECISION NOT NULL,
    longitude DOUBLE PRECISION NOT NULL,
    country VARCHAR(100)
);

CREATE TABLE forecasts (
    id serial primary key,
    city_id INTEGER NOT NULL REFERENCES cities(id),
    date BIGINT NOT NULL,
    temperature DOUBLE PRECISION NOT NULL,
    weather JSONB NOT NULL,
    UNIQUE (city_id, date)
);

