CREATE TABLE online_drivers (
    driver_id BIGINT PRIMARY KEY REFERENCES drivers(id) ON DELETE CASCADE,
    is_online BOOLEAN NOT NULL DEFAULT TRUE,
    last_ping_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    went_online_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    current_lat DOUBLE PRECISION,
    current_lng DOUBLE PRECISION,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_online_drivers_is_online ON online_drivers(is_online);
CREATE INDEX idx_online_drivers_last_ping ON online_drivers(last_ping_at);
CREATE INDEX idx_online_drivers_location ON online_drivers(current_lat, current_lng);