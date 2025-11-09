CREATE TABLE customers (
   id serial primary key,
   name VARCHAR(255) NOT NULL,
   email VARCHAR(255) NOT NULL UNIQUE,
   phone VARCHAR(20) NOT NULL UNIQUE,
   password VARCHAR(255) NOT NULL,
   created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_phone ON customers(phone);

CREATE TABLE drivers (
     id serial primary key,
     name VARCHAR(255) NOT NULL,
     phone VARCHAR(20) NOT NULL UNIQUE,
     vehicle_no VARCHAR(50),
     is_online BOOLEAN NOT NULL DEFAULT FALSE,
     current_lat DOUBLE PRECISION,
     current_lng DOUBLE PRECISION,
     last_ping_at TIMESTAMP,
     last_updated_at TIMESTAMP,
     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_drivers_phone ON drivers(phone);
CREATE INDEX idx_drivers_online ON drivers(is_online);
CREATE INDEX idx_drivers_last_ping ON drivers(last_ping_at);
CREATE INDEX idx_drivers_location ON drivers(current_lat, current_lng);

CREATE TABLE otp_records (
     id serial primary key,
     phone VARCHAR(20) NOT NULL,
     otp VARCHAR(10) NOT NULL,
     purpose VARCHAR(50) NOT NULL,
     is_verified BOOLEAN NOT NULL DEFAULT FALSE,
     is_expired BOOLEAN NOT NULL DEFAULT FALSE,
     expires_at TIMESTAMP NOT NULL,
     verified_at TIMESTAMP,
     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_otp_phone ON otp_records(phone);
CREATE INDEX idx_otp_expires_at ON otp_records(expires_at);
CREATE INDEX idx_otp_created_at ON otp_records(created_at);