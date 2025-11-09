-- Initialize PostgreSQL database for Ride Engine
-- This script creates the tables for customers, drivers, and rides

-- Drop existing tables (use with caution in production)
DROP TABLE IF EXISTS rides CASCADE;
DROP TABLE IF EXISTS otp_records CASCADE;
DROP TABLE IF EXISTS drivers CASCADE;
DROP TABLE IF EXISTS customers CASCADE;

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

-- Online drivers table (tracks active/online drivers separately)
CREATE TABLE online_drivers (
    driver_id BIGINT PRIMARY KEY REFERENCES drivers(id) ON DELETE CASCADE,
    is_online BOOLEAN NOT NULL DEFAULT TRUE,
    last_ping_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    went_online_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    current_lat DOUBLE PRECISION,
    current_lng DOUBLE PRECISION,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for online drivers
CREATE INDEX idx_online_drivers_is_online ON online_drivers(is_online);
CREATE INDEX idx_online_drivers_last_ping ON online_drivers(last_ping_at);
CREATE INDEX idx_online_drivers_location ON online_drivers(current_lat, current_lng);

-- Comments for documentation
COMMENT ON TABLE customers IS 'Stores customer/rider information with authentication';
COMMENT ON TABLE drivers IS 'Stores driver information with real-time location and online status';
COMMENT ON TABLE rides IS 'Stores ride requests and tracking information';
COMMENT ON TABLE otp_records IS 'Stores OTP records for audit trail and security monitoring';
COMMENT ON TABLE online_drivers IS 'Tracks currently online/active drivers with real-time location and ping status';
COMMENT ON COLUMN drivers.last_ping_at IS 'Last heartbeat from driver app - used for auto offline detection';
COMMENT ON COLUMN rides.fare IS 'Calculated fare for the ride';
COMMENT ON COLUMN otp_records.purpose IS 'Purpose of OTP: driver_login, customer_verification, password_reset, etc';
COMMENT ON COLUMN online_drivers.last_ping_at IS 'Last location ping from driver - used to detect inactive drivers';
COMMENT ON COLUMN online_drivers.went_online_at IS 'Timestamp when driver went online in current session';

-- Insert sample data (optional - remove in production)
-- Sample customer
INSERT INTO customers (name, email, phone, password) VALUES
('John Doe', 'john@example.com', '01234567890', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi');
-- password is 'password'

-- Sample driver
INSERT INTO drivers (name, phone, vehicle_no, is_online) VALUES
('Jane Driver', '01875113841', 'DHA-1234', false);

COMMIT;

-- Display table information
SELECT 'Database initialization completed!' as status;
SELECT 'Tables created: customers, drivers, rides, otp_records' as info;
SELECT 'Sample data inserted (1 customer, 1 driver)' as info;
