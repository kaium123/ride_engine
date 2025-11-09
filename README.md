# Installation Guide - Ride Engine

This guide provides two methods to run the Ride Engine application.

---

## Prerequisites

- **Docker & Docker Compose** (required for both methods)
- **Go 1.21+** (required for local development only)
- **Make** (command-line build tool)

---

## Method 1: Run Application Locally (Development)

This method runs databases in Docker containers while running the Go application natively on your machine.

### Step 1: Start Database Services

```bash
make docker-up-database
```

This command starts:
- **PostgreSQL** on port `5436`
- **MongoDB** on port `27016`
- **Redis** on port `6389`

### Step 2: Run the Application

```bash
make run
```

The API server will start on `http://localhost:8080`

### Step 3: Verify Installation

```bash
curl http://localhost:8080/health
```

Expected response: `{"status":"ok"}`

### Step 4: Stop Database Services (when done)

```bash
make docker-down
```

---

## Method 2: Run Everything in Docker (Production-like)

This method runs both the application and all databases in Docker containers.

### Step 1: Build and Start All Services

```bash
make docker-up
```

This command:
1. Builds the Go application Docker image
2. Starts PostgreSQL, MongoDB, Redis containers
3. Runs database migrations automatically
4. Starts the API server container

All services will be running in Docker containers.

### Step 2: Verify Installation

```bash
curl http://localhost:8080/health
```

Expected response: `{"status":"ok"}`

### Step 3: View Logs

```bash
# View API logs
docker logs -f ride_engine-app

# View all services logs
docker-compose logs -f
```

### Step 4: Stop All Services

```bash
make docker-down
```

## Swagger Documentation
```bash
http://localhost:8080/swagger/index.html
```

---

## Environment Configuration

The application uses environment variables defined in your configuration. Key settings:

- **Database URLs**: Automatically configured in docker-compose.yml
- **Redis**: localhost:6389 (local) or redis:6379 (docker)
- **MongoDB**: localhost:27016 (local) or mongodb:27017 (docker)
- **PostgreSQL**: localhost:5436 (local) or postgres:5432 (docker)

---

## Database Access

### PostgreSQL
```bash
# Local development
psql -h localhost -p 5432 -U root -d ride_engine

# Docker
docker exec -it ride_engine-postgres psql -U root -d ride_engine

### MongoDB
# Local development
mongosh --port 27016 -u root -p secret --authenticationDatabase admin

# Docker
docker exec -it ride_engine-mongo mongosh -u root -p secret --authenticationDatabase admin
```

### Redis
```bash
# Local development
redis-cli

# Docker
docker exec -it ride_engine-redis redis-cli
```

---

## Testing the API

### Register a Customer
```bash
curl -X POST http://localhost:8080/api/v1/customers/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "1234567890",
    "password": "password123"
  }'
```

### Register a Driver
```bash
curl -X POST http://localhost:8080/api/v1/drivers/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Mike Driver",
    "phone": "9876543210",
    "vehicle_no": "ABC-123"
  }'
```

### Request OTP (Driver Login)
```bash
curl -X POST http://localhost:8080/api/v1/drivers/login/request-otp \
  -H "Content-Type: application/json" \
  -d '{"phone": "9876543210"}'
```

**Note**: In development mode, the OTP is always `123456`

### Verify OTP & Login
```bash
curl -X POST http://localhost:8080/api/v1/drivers/login/verify-otp \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "9876543210",
    "otp": "123456"
  }'
```

### Update drivers location
```bash
curl --location 'http://localhost:8080/api/v1/drivers/location' \
--header "Content-Type: application/json" \
--header "Authorization: Bearer $TOKEN" \
--data '{
    "latitude": 23.8003,
    "longitude": 90.4025
}'
```

### Customer Login
```bash
curl --location 'http://localhost:8080/api/v1/customers/login' \
--header 'Content-Type: application/json' \
--data-raw '{
    "email": "john@example.com",
    "password": "password123"
}'
```

### Find nearest drivers
```bash
curl --location 'http://localhost:8080/api/v1/drivers/nearby' \
--header "Content-Type: application/json" \
--header "Authorization: Bearer $TOKEN" \
--data '{
    "latitude": 23.8103,
    "longitude": 90.4125,
    "radius": 5000,
    "limit": 3
}'
```

### Request a Ride (Customer)
```bash
# First login to get token
TOKEN="<your_jwt_token_from_login>"

curl -X POST http://localhost:8080/api/v1/rides/ \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pickup_lat": 23.8100,
    "pickup_lng": 90.4120,
    "dropoff_lat": 23.7509,
    "dropoff_lng": 90.3761
  }'
```


# Find nearest rides
```bash
curl --location 'http://localhost:8080/api/v1/rides/nearby' \
--header "Content-Type: application/json" \
--header "Authorization: Bearer $TOKEN" \
--data '{
    "lat": 23.8103,
    "lng": 90.4125,
    "radius": 5000,
    "limit": 3
}'
```

### Accept ride
```bash
curl --location --request POST 'http://localhost:8080/api/v1/rides/accept?ride_id=11' \
--header "Authorization: Bearer $TOKEN" \
--data ''
```

### Start ride
```bash
curl --location 'http://localhost:8080/api/v1/rides/start?ride_id=11' \
--header "Content-Type: application/json" \
--header "Authorization: Bearer $TOKEN" \
--data '{
    "pickup_lat": 23.23,
    "pickup_lng": 90.24,
    "dropoff_lat": 25.23,
    "dropoff_lng": 95.24
}'
```

### Ride complete
```bash
curl --location 'http://localhost:8080/api/v1/rides/complete?ride_id=1' \
--header "Content-Type: application/json" \
--header "Authorization: Bearer $TOKEN" \
--data '{
    "pickup_lat": 23.23,
    "pickup_lng": 90.24,
    "dropoff_lat": 25.23,
    "dropoff_lng": 95.24
}'
```

### Ride Cancel
```bash
curl --location 'http://localhost:8080/api/v1/rides/cancel?ride_id=2' \
--header "Content-Type: application/json" \
--header "Authorization: Bearer $TOKEN" \
--data '{
    "pickup_lat": 23.23,
    "pickup_lng": 90.24,
    "dropoff_lat": 25.23,
    "dropoff_lng": 95.24
}'
```

### Ride status

```bash
curl --location --request GET 'http://localhost:8080/api/v1/rides/status?ride_id=11' \
--header "Content-Type: application/json" \
--header "Authorization: Bearer $TOKEN" \
--data '{
    "pickup_lat": 23.23,
    "pickup_lng": 90.24,
    "dropoff_lat": 25.23,
    "dropoff_lng": 95.24
}'
```