# Ride Engine API Documentation

## Overview

The Ride Engine API is a comprehensive ride-sharing platform that manages customers, drivers, and rides. The API provides endpoints for user registration, authentication, location tracking, and ride management.

## Quick Start with Docker

### Prerequisites
- Docker and Docker Compose installed on your machine
- Git (to clone the repository)

### Running the Application

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd ride_engine
   ```

2. **Start all services using Docker Compose**
   ```bash
   docker-compose up -d
   ```

   This command will start:
   - **Ride Engine API** (port 8080)
   - **Swagger UI** (port 8081)
   - **PostgreSQL** (port 5436)
   - **MongoDB** (port 27016)
   - **Redis** (port 6389)

3. **Check if services are running**
   ```bash
   docker-compose ps
   ```

4. **View logs**
   ```bash
   docker-compose logs -f app
   ```

5. **Stop all services**
   ```bash
   docker-compose down
   ```

## Accessing the Swagger Documentation

Once the application is running, you can access the interactive Swagger UI in two ways:

### Option 1: Standalone Swagger UI Container (Recommended for Interviewers)
**http://localhost:8081**

This is a dedicated Swagger UI container that serves the API documentation independently.

### Option 2: Built-in Swagger Endpoint
**http://localhost:8080/swagger/index.html**

This is served directly by the Ride Engine API application.

Both options provide:
- Complete API documentation
- Interactive testing interface
- Request/response examples
- Model schemas
- Authentication support

## API Endpoints

### Base URL
```
http://localhost:8080/api/v1
```

### Health Check
```
GET /health
```

### Customer Endpoints

#### Register a New Customer
```
POST /api/v1/customers/register
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "+1234567890",
  "password": "securepassword"
}
```

#### Customer Login
```
POST /api/v1/customers/login
Content-Type: application/json

{
  "email": "john@example.com",
  "password": "securepassword"
}
```

### Driver Endpoints

#### Register a New Driver
```
POST /api/v1/drivers/register
Content-Type: application/json

{
  "name": "Jane Driver",
  "phone": "+1234567891",
  "vehicle_no": "ABC-1234"
}
```

#### Request OTP for Driver Login
```
POST /api/v1/drivers/login/request-otp
Content-Type: application/json

{
  "phone": "+1234567891"
}
```

#### Verify OTP and Login
```
POST /api/v1/drivers/login/verify-otp
Content-Type: application/json

{
  "phone": "+1234567891",
  "otp": "123456"
}
```

#### Update Driver Location (Requires Authentication)
```
POST /api/v1/drivers/location
Authorization: Bearer <token>
Content-Type: application/json

{
  "latitude": 23.8103,
  "longitude": 90.4125
}
```

#### Set Driver Online Status (Requires Authentication)
```
POST /api/v1/drivers/status
Authorization: Bearer <token>
Content-Type: application/json

{
  "is_online": true
}
```

### Ride Endpoints

#### Find Nearest Drivers (Requires Authentication)
```
GET /api/v1/rides/nearby?lat=23.8103&lng=90.4125&radius=3000&limit=5
Authorization: Bearer <token>
```

#### Request a Ride (Requires Authentication)
```
POST /api/v1/rides
Authorization: Bearer <token>
Content-Type: application/json

{
  "pickup_lat": 23.8103,
  "pickup_lng": 90.4125,
  "dropoff_lat": 23.7500,
  "dropoff_lng": 90.3800
}
```

#### Accept a Ride (Requires Authentication - Driver)
```
POST /api/v1/rides/accept?ride_id=1
Authorization: Bearer <token>
```

#### Start a Ride (Requires Authentication)
```
POST /api/v1/rides/start?ride_id=1
Authorization: Bearer <token>
```

#### Complete a Ride (Requires Authentication)
```
POST /api/v1/rides/complete?ride_id=1
Authorization: Bearer <token>
```

#### Cancel a Ride (Requires Authentication)
```
POST /api/v1/rides/cancel?ride_id=1
Authorization: Bearer <token>
```

## Authentication

The API uses JWT (JSON Web Token) for authentication. After logging in (either as a customer or driver), you'll receive a token in the response:

```json
{
  "customer": { ... },
  "token": "eyJhbGciOiJIUzI1NiIs..."
}
```

Include this token in the `Authorization` header for protected endpoints:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIs...
```

### Using Authentication in Swagger

1. Click the **Authorize** button in Swagger UI
2. Enter your token in the format: `Bearer <your-token>`
3. Click **Authorize**
4. Now you can test protected endpoints

## Testing the API

### Using Swagger UI (Recommended)

1. Navigate to http://localhost:8080/swagger/index.html
2. Click on any endpoint to expand it
3. Click "Try it out"
4. Fill in the required parameters
5. Click "Execute"
6. View the response

### Using cURL

**Register a Customer:**
```bash
curl -X POST http://localhost:8080/api/v1/customers/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "+1234567890",
    "password": "securepassword"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/customers/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepassword"
  }'
```

**Request a Ride (with token):**
```bash
curl -X POST http://localhost:8080/api/v1/rides \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <your-token>" \
  -d '{
    "pickup_lat": 23.8103,
    "pickup_lng": 90.4125,
    "dropoff_lat": 23.7500,
    "dropoff_lng": 90.3800
  }'
```

## Database Access

### PostgreSQL
- **Host:** localhost
- **Port:** 5436
- **Database:** ride_engine
- **Username:** root
- **Password:** secret

### MongoDB
- **Host:** localhost
- **Port:** 27016
- **Database:** ride_engine
- **Username:** root
- **Password:** secret

### Redis
- **Host:** localhost
- **Port:** 6389

## Architecture

The application follows a clean architecture with:

- **Handlers** - HTTP request handlers
- **Services** - Business logic layer
- **Repositories** - Data access layer
- **Models** - Domain models

### Technology Stack

- **Language:** Go 1.25.1
- **Web Framework:** Standard `net/http`
- **Databases:**
  - PostgreSQL (User data, Rides)
  - MongoDB (Location data)
  - Redis (Sessions, OTP)
- **Documentation:** Swagger/OpenAPI 2.0
- **Containerization:** Docker & Docker Compose

## Development

### Running Locally (without Docker)

1. **Install dependencies**
   ```bash
   go mod download
   ```

2. **Generate Swagger docs**
   ```bash
   swag init
   ```

3. **Run the application**
   ```bash
   go run main.go serve
   ```

### Regenerate Swagger Documentation

If you make changes to the API annotations:

```bash
swag init
```

## Troubleshooting

### Issue: Cannot connect to databases

**Solution:** Ensure all Docker containers are running:
```bash
docker-compose ps
```

If any service is down, restart it:
```bash
docker-compose restart <service-name>
```

### Issue: Port already in use

**Solution:** Stop the conflicting service or change the port mapping in `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"  # Change 8081 to any available port
```

### Issue: Swagger UI not loading

**Solution:**
1. Ensure Swagger docs are generated: `swag init`
2. Rebuild the Docker image: `docker-compose build app`
3. Restart the service: `docker-compose up -d app`

## Support

For issues, questions, or contributions:
- **Maintainer:** Mohammad Kaium
- **Email:** mohammadkaiom79@gmail.com

## License

MIT License
