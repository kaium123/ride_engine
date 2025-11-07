# Quick Start Guide

Get the Ride Engine up and running in 5 minutes!

## Prerequisites

- Docker and Docker Compose installed
- Go 1.21+ installed
- Git (optional)

## Step 1: Start Databases

```bash
# Start PostgreSQL and MongoDB
docker-compose up -d

# Verify containers are running
docker-compose ps
```

Expected output:
```
NAME                        STATUS
ride_engine_mongodb         Up
ride_engine_postgres        Up
ride_engine_pgadmin         Up (optional)
ride_engine_mongo_express   Up (optional)
```

## Step 2: Configure Environment (Optional)

The application works with default settings. To customize:

```bash
cp .env.example .env
# Edit .env if needed
```

## Step 3: Run the Application

```bash
# Option 1: Run directly
go run cmd/api/main.go

# Option 2: Build and run
make build
./bin/ride_engine

# Option 3: Use Make
make run
```

You should see:
```
✓ PostgreSQL connected successfully
✓ MongoDB connected successfully
Starting Ride Engine API server on port 8080...
Server is running on http://localhost:8080
```

## Step 4: Test the API

### Register a Rider
```bash
curl -X POST http://localhost:8080/api/v1/riders \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "01711111111",
    "email": "rider@example.com"
  }'
```

### Register a Driver
```bash
curl -X POST http://localhost:8080/api/v1/drivers \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "01722222222"
  }'
```

### Driver Login - Send OTP
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "01722222222"
  }'
```

**Check the console for the OTP code!**

### Driver Login - Verify OTP
```bash
curl -X POST http://localhost:8080/api/v1/login/verify \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "01722222222",
    "otp": "123456"
  }'
```

Save the `driver_id` from the response!

### Set Driver Online
```bash
curl -X POST http://localhost:8080/api/v1/drivers/status \
  -H "Content-Type: application/json" \
  -d '{
    "driver_id": "<your_driver_id>",
    "is_online": true
  }'
```

### Update Driver Location
```bash
curl -X POST http://localhost:8080/api/v1/drivers/location \
  -H "Content-Type: application/json" \
  -d '{
    "driver_id": "<your_driver_id>",
    "latitude": 23.8103,
    "longitude": 90.4125
  }'
```

### Request a Ride
```bash
curl -X POST http://localhost:8080/api/v1/rides \
  -H "Content-Type: application/json" \
  -d '{
    "rider_id": "<your_rider_id>",
    "pickup_location": {
      "latitude": 23.8103,
      "longitude": 90.4125
    },
    "dropoff_location": {
      "latitude": 23.7509,
      "longitude": 90.3761
    }
  }'
```

### Find Nearby Rides
```bash
curl "http://localhost:8080/api/v1/drivers/rides/nearby?driver_id=<your_driver_id>&max_distance=10"
```

### Accept a Ride
```bash
curl -X POST "http://localhost:8080/api/v1/drivers/rides/accept?ride_id=<ride_id>" \
  -H "Content-Type: application/json" \
  -d '{
    "driver_id": "<your_driver_id>"
  }'
```

## Management Tools

### pgAdmin (PostgreSQL)
- URL: http://localhost:5050
- Email: admin@rideengine.com
- Password: admin

### Mongo Express (MongoDB)
- URL: http://localhost:8081
- Username: admin
- Password: admin

## Common Commands

```bash
# Start databases
make docker-up

# Stop databases
make docker-down

# View logs
make docker-logs

# Reset databases
make db-reset

# Build application
make build

# Run tests (when available)
make test

# Clean build artifacts
make clean
```

## Troubleshooting

### Port Already in Use
If you get "port already in use" errors:

1. Edit `docker-compose.yml` and change the port mappings
2. Update `.env` file with new ports

### Can't Connect to Database
```bash
# Check if containers are running
docker-compose ps

# Restart containers
docker-compose restart

# View logs for errors
docker-compose logs postgres
docker-compose logs mongodb
```

### Application Won't Start
```bash
# Verify Go dependencies
go mod tidy

# Rebuild
make clean
make build

# Check environment variables
cat .env
```

## Next Steps

- Read [README.md](README.md) for complete API documentation
- Read [README_DATABASE.md](README_DATABASE.md) for database details
- Explore the codebase structure
- Add more features!

## Architecture Overview

```
PostgreSQL (Port 5432)     MongoDB (Port 27017)
     |                            |
     |-- Users                    |-- Driver Locations (GeoJSON)
     |-- Drivers                  |-- Ride Tracking
     |-- Rides                    |
            \                    /
             \                  /
              \                /
            Ride Engine API (Port 8080)
                    |
              HTTP Endpoints
```

## Data Flow

1. **Driver comes online** → Updates PostgreSQL + Saves location to MongoDB
2. **Rider requests ride** → Stored in PostgreSQL
3. **Driver location updates** → PostgreSQL (current) + MongoDB (history)
4. **Find nearby drivers** → Query MongoDB with geospatial search
5. **Accept ride** → Update PostgreSQL
6. **Track ride** → Save locations to MongoDB

Happy coding! 🚀
