# Database Initialization Scripts

This directory contains initialization scripts for setting up the Ride Engine databases.

## Files

- **init-postgres.sql**: PostgreSQL database initialization script
- **init-mongo.js**: MongoDB initialization script

## Usage

### Option 1: Using Docker Compose (Recommended)

The easiest way to initialize the databases is using docker-compose, which will automatically run these scripts when the containers are first created:

```bash
# From the project root
docker-compose up -d
```

This will:
- ✅ Start PostgreSQL on port 5436
- ✅ Start MongoDB on port 27016
- ✅ Start Redis on port 6379
- ✅ Automatically run init scripts on first startup

### Option 2: Manual Execution

If you're running databases locally or separately, you can execute these scripts manually:

#### PostgreSQL

```bash
# Using psql
psql -h localhost -p 5436 -U root -d ride_engine -f scripts/init-postgres.sql

# Or if PostgreSQL is running in Docker
docker exec -i ride_engine-postgres psql -U root -d ride_engine < scripts/init-postgres.sql
```

#### MongoDB

```bash
# Using mongosh
mongosh "mongodb://root:secret@localhost:27016/ride_engine?authSource=admin" < scripts/init-mongo.js

# Or if MongoDB is running in Docker
docker exec -i ride_engine-mongo mongosh "mongodb://root:secret@localhost:27017/ride_engine?authSource=admin" < scripts/init-mongo.js
```

## What Gets Created

### PostgreSQL

**Tables:**
- `customers` - Customer accounts with authentication
- `drivers` - Driver accounts with location tracking
- `rides` - Ride requests and tracking

**Indexes:**
- Email and phone indexes for fast lookups
- Status indexes for ride queries
- Location indexes for spatial queries
- Last ping timestamp index for online/offline detection

**Views:**
- `active_rides` - Shows all active rides with customer and driver details
- `online_drivers` - Shows currently online drivers with activity status

**Sample Data:**
- 1 sample customer (email: john@example.com, password: password)
- 1 sample driver (phone: 01875113841)

### MongoDB

**Collections:**
- `driver_locations` - Real-time driver locations with geospatial data

**Indexes:**
- 2dsphere geospatial index for location queries
- Unique index on driver_id
- TTL index to auto-delete old data after 7 days

**Sample Data:**
- 1 sample driver location (driver_id: 1, Dhaka coordinates)

## Database Schemas

### PostgreSQL Tables

```sql
customers (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255),
  email VARCHAR(255) UNIQUE,
  phone VARCHAR(20) UNIQUE,
  password VARCHAR(255),
  created_at TIMESTAMP
)

drivers (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255),
  phone VARCHAR(20) UNIQUE,
  vehicle_no VARCHAR(50),
  is_online BOOLEAN,
  current_lat DOUBLE PRECISION,
  current_lng DOUBLE PRECISION,
  last_ping_at TIMESTAMP,
  last_updated_at TIMESTAMP,
  created_at TIMESTAMP
)

rides (
  id BIGSERIAL PRIMARY KEY,
  customer_id BIGINT REFERENCES customers(id),
  driver_id BIGINT REFERENCES drivers(id),
  pickup_lat DOUBLE PRECISION,
  pickup_lng DOUBLE PRECISION,
  dropoff_lat DOUBLE PRECISION,
  dropoff_lng DOUBLE PRECISION,
  status VARCHAR(20),
  fare DECIMAL(10,2),
  requested_at TIMESTAMP,
  accepted_at TIMESTAMP,
  started_at TIMESTAMP,
  completed_at TIMESTAMP,
  cancelled_at TIMESTAMP
)
```

### MongoDB Collections

```javascript
driver_locations {
  driver_id: NumberLong(1),
  location: {
    type: "Point",
    coordinates: [90.4125, 23.8103]  // [longitude, latitude]
  },
  updated_at: ISODate("2025-11-05T07:30:00Z")
}
```

## Testing the Setup

After initialization, you can verify the setup:

### PostgreSQL

```bash
# Connect to PostgreSQL
psql -h localhost -p 5436 -U root -d ride_engine

# List tables
\dt

# Check customers
SELECT * FROM customers;

# Check drivers
SELECT * FROM drivers;

# Check views
SELECT * FROM active_rides;
SELECT * FROM online_drivers;
```

### MongoDB

```bash
# Connect to MongoDB
mongosh "mongodb://root:secret@localhost:27016/ride_engine?authSource=admin"

# Show collections
show collections

# Check driver locations
db.driver_locations.find().pretty()

# Test geospatial query (find drivers within 5km)
db.driver_locations.find({
  location: {
    $nearSphere: {
      $geometry: { type: "Point", coordinates: [90.4100, 23.8100] },
      $maxDistance: 5000
    }
  }
})

# Show indexes
db.driver_locations.getIndexes()
```

### Redis

```bash
# Connect to Redis
redis-cli

# Test connection
PING

# Check keys (should be empty initially)
KEYS *
```

## Resetting the Databases

If you need to reset the databases:

### Using Docker Compose

```bash
# Stop and remove containers with volumes
docker-compose down -v

# Restart (will reinitialize)
docker-compose up -d
```

### Manual Reset

**PostgreSQL:**
```sql
DROP TABLE IF EXISTS rides CASCADE;
DROP TABLE IF EXISTS drivers CASCADE;
DROP TABLE IF EXISTS customers CASCADE;
```

Then re-run the init script.

**MongoDB:**
```javascript
db.driver_locations.drop()
```

Then re-run the init script.

## Notes

- The init scripts are **idempotent** - they can be run multiple times safely
- Sample data includes a test customer with password "password" (bcrypt hashed)
- MongoDB uses TTL index to auto-delete location data older than 7 days
- PostgreSQL uses BIGSERIAL (int64) for auto-incrementing primary keys
- All coordinates use GeoJSON format: [longitude, latitude]
- Redis requires no initialization script (used for OTP storage with TTL)

## Troubleshooting

**Scripts not running in Docker:**
- Ensure scripts directory is mounted correctly in docker-compose.yml
- Scripts only run on **first container creation**
- To force re-run, delete the volume and recreate: `docker-compose down -v && docker-compose up -d`

**MongoDB geospatial queries not working:**
- Verify 2dsphere index exists: `db.driver_locations.getIndexes()`
- Ensure coordinates are in [longitude, latitude] format
- MongoDB coordinates range: longitude [-180, 180], latitude [-90, 90]

**PostgreSQL connection issues:**
- Check port 5436 is not in use
- Verify credentials: root/secret
- Wait for container to be fully ready (check `docker-compose logs postgres`)

**Redis not accessible:**
- Check port 6379 is available
- Redis requires no password by default in this setup
- Verify with `redis-cli ping`
