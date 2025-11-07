-- Customers table (riders/passengers)
CREATE TABLE customers (
   id BIGSERIAL PRIMARY KEY,
   name VARCHAR(255) NOT NULL,
   email VARCHAR(255) NOT NULL UNIQUE,
   phone VARCHAR(20) NOT NULL UNIQUE,
   password VARCHAR(255) NOT NULL,
   created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for customers
CREATE INDEX idx_customers_email ON customers(email);
CREATE INDEX idx_customers_phone ON customers(phone);

-- Drivers table
CREATE TABLE drivers (
     id BIGSERIAL PRIMARY KEY,
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

-- Indexes for drivers
CREATE INDEX idx_drivers_phone ON drivers(phone);
CREATE INDEX idx_drivers_online ON drivers(is_online);
CREATE INDEX idx_drivers_last_ping ON drivers(last_ping_at);
CREATE INDEX idx_drivers_location ON drivers(current_lat, current_lng);

-- Rides table
CREATE TABLE rides (
   id BIGSERIAL PRIMARY KEY,
   customer_id BIGINT NOT NULL REFERENCES customers(id) ON DELETE CASCADE,
   driver_id BIGINT REFERENCES drivers(id) ON DELETE SET NULL,
   pickup_lat DOUBLE PRECISION NOT NULL,
   pickup_lng DOUBLE PRECISION NOT NULL,
   dropoff_lat DOUBLE PRECISION NOT NULL,
   dropoff_lng DOUBLE PRECISION NOT NULL,
   status VARCHAR(20) NOT NULL CHECK (status IN ('requested', 'accepted', 'started', 'completed', 'cancelled')),
   fare DECIMAL(10,2),
   requested_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
   accepted_at TIMESTAMP,
   started_at TIMESTAMP,
   completed_at TIMESTAMP,
   cancelled_at TIMESTAMP
);

-- Indexes for rides
CREATE INDEX idx_rides_customer_id ON rides(customer_id);
CREATE INDEX idx_rides_driver_id ON rides(driver_id);
CREATE INDEX idx_rides_status ON rides(status);
CREATE INDEX idx_rides_requested_at ON rides(requested_at);

-- OTP records table (for audit trail)
CREATE TABLE otp_records (
     id BIGSERIAL PRIMARY KEY,
     phone VARCHAR(20) NOT NULL,
     otp VARCHAR(10) NOT NULL,
     purpose VARCHAR(50) NOT NULL,
     is_verified BOOLEAN NOT NULL DEFAULT FALSE,
     is_expired BOOLEAN NOT NULL DEFAULT FALSE,
     expires_at TIMESTAMP NOT NULL,
     verified_at TIMESTAMP,
     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for OTP records
CREATE INDEX idx_otp_phone ON otp_records(phone);
CREATE INDEX idx_otp_expires_at ON otp_records(expires_at);
CREATE INDEX idx_otp_created_at ON otp_records(created_at);