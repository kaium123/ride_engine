# Ride Engine - Monolithic Architecture

## Overview
This is a ride-hailing backend system built as a clean monolithic application in Go. It uses PostgreSQL for relational data, MongoDB for geo-location queries, and Redis for OTP storage.

## Architecture

```
┌─────────────────────────────┐
│        API Layer            │
│  (REST Handlers / Routers)  │
└──────────────┬──────────────┘
               │
┌──────────────┴──────────────┐
│       Service Layer          │
│ (Business Logic Modules)     │
│  ├── CustomerService         │
│  ├── DriverService           │
│  ├── RideService             │
│  ├── OTPService              │
│  └── LocationService         │
└──────────────┬──────────────┘
               │
┌──────────────┴──────────────┐
│        Data Layer            │
│ (Repositories / ORM / DB)    │
└──────────────┬──────────────┘
               │
┌──────────────┴──────────────┐
│  PostgreSQL / MongoDB / Redis│
└─────────────────────────────┘
```

## Project Structure

```
ride_engine/
├── cmd/
│   └── api/
│       └── main.go                    # Application entry point
├── internal/
│   └── ride_engine/
│       ├── domain/                    # Domain models
│       │   ├── models.go             # Customer, Driver, Ride
│       │   └── location.go           # Location with distance calculation
│       ├── handler/                   # HTTP handlers
│       │   ├── customer_handler.go
│       │   ├── driver_handler.go
│       │   └── ride_handler.go
│       ├── service/                   # Business logic
│       │   ├── customer_service.go
│       │   ├── driver_service.go
│       │   ├── ride_service.go
│       │   ├── otp_service.go
│       │   └── location_service.go
│       └── repository/                # Data access
│           ├── customer_repository.go
│           └── postgres/
│               ├── models.go         # GORM models
│               ├── customer_postgres.go
│               ├── driver_postgres.go
│               └── ride_postgres.go
├── pkg/
│   ├── config/
│   │   └── config.go                 # Configuration
│   ├── database/
│   │   ├── postgres.go               # PostgreSQL connection
│   │   ├── mongodb.go                # MongoDB connection
│   │   └── redis.go                  # Redis connection
│   └── utils/
│       ├── jwt.go                    # JWT utilities
│       └── password.go               # Password hashing
└── go.mod
```

## Database Schema

### PostgreSQL Tables

**customers**
```sql
CREATE TABLE customers (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  email VARCHAR(255) UNIQUE NOT NULL,
  phone VARCHAR(20) UNIQUE NOT NULL,
  password VARCHAR(255) NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**drivers**
```sql
CREATE TABLE drivers (
  id BIGSERIAL PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  phone VARCHAR(20) UNIQUE NOT NULL,
  vehicle_no VARCHAR(50),
  is_online BOOLEAN DEFAULT FALSE,
  current_lat DOUBLE PRECISION,
  current_lng DOUBLE PRECISION,
  last_ping_at TIMESTAMP,
  last_updated_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**rides**
```sql
CREATE TABLE rides (
  id BIGSERIAL PRIMARY KEY,
  customer_id BIGINT NOT NULL REFERENCES customers(id),
  driver_id BIGINT REFERENCES drivers(id),
  pickup_lat DOUBLE PRECISION NOT NULL,
  pickup_lng DOUBLE PRECISION NOT NULL,
  dropoff_lat DOUBLE PRECISION NOT NULL,
  dropoff_lng DOUBLE PRECISION NOT NULL,
  status VARCHAR(20) NOT NULL,
  fare DECIMAL(10,2),
  requested_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  accepted_at TIMESTAMP,
  started_at TIMESTAMP,
  completed_at TIMESTAMP,
  cancelled_at TIMESTAMP
);
```

### MongoDB Collections

**driver_locations**
```json
{
  "driver_id": 123,
  "location": {
    "type": "Point",
    "coordinates": [90.4125, 23.8103]
  },
  "updated_at": ISODate("2025-11-05T07:30:00Z")
}
```

Index:
```javascript
db.driver_locations.createIndex({ location: "2dsphere" })
```

### Redis Keys

```
otp:<phone_number> = "123456" (TTL: 2 minutes)
```

## API Endpoints

### Customer Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/customers/register` | Register new customer |
| POST | `/api/v1/customers/login` | Login customer |

### Driver Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/drivers/register` | Register new driver |
| POST | `/api/v1/drivers/login/request-otp` | Request OTP |
| POST | `/api/v1/drivers/login/verify-otp` | Verify OTP & login |
| POST | `/api/v1/drivers/location` | Update driver location |
| POST | `/api/v1/drivers/status` | Set online/offline |

### Ride Endpoints
| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/rides` | Request a ride |
| GET | `/api/v1/rides/nearby` | Get nearby rides (driver) |
| POST | `/api/v1/rides/accept` | Accept a ride |
| POST | `/api/v1/rides/start` | Start a ride |
| POST | `/api/v1/rides/complete` | Complete a ride |
| POST | `/api/v1/rides/cancel` | Cancel a ride |

## Key Features

### 1. Customer Authentication
- Email + password-based registration and login
- JWT token generation for authenticated sessions
- Password hashing using bcrypt

### 2. Driver Authentication
- Phone number-based registration
- OTP-based login (6-digit OTP with 2-minute expiry)
- JWT token generation after OTP verification

### 3. Driver Location Tracking
- **Dual storage**:
  - PostgreSQL: Stores `last_ping_at`, `current_lat`, `current_lng`
  - MongoDB: Stores geo-location with 2dsphere index
- **Automatic online/offline detection**:
  - Driver marked online when location ping received
  - Background worker checks every 30 seconds
  - Driver marked offline if no ping for >60 seconds

### 4. Location Service
- MongoDB-based geo queries using `$nearSphere`
- Find nearest drivers within specified radius
- Haversine formula for distance calculations

### 5. Ride Management
- Status transitions: `requested → accepted → started → completed/cancelled`
- Nearby ride queries for drivers
- Real-time ride tracking

## Background Workers

### Driver Activity Monitor
- Runs every 30 seconds
- Marks drivers offline if `last_ping_at > 60 seconds ago`
- Automatically started with the application
- Gracefully shut down with the server

## Environment Variables

```bash
# Server
SERVER_PORT=8080

# PostgreSQL
POSTGRES_HOST=localhost
POSTGRES_PORT=5436
POSTGRES_USER=root
POSTGRES_PASSWORD=secret
POSTGRES_DB=ride_engine
POSTGRES_SSLMODE=disable

# MongoDB
MONGODB_URI=mongodb://root:secret@localhost:27016/?authSource=admin
MONGODB_DATABASE=ride_engine

# Redis
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

# JWT
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRATION_HOURS=24
```

## Running the Application

1. **Start dependencies**:
   ```bash
   docker-compose up -d  # or start PostgreSQL, MongoDB, Redis manually
   ```

2. **Install Go dependencies**:
   ```bash
   go mod tidy
   ```

3. **Run the server**:
   ```bash
   go run cmd/api/main.go
   ```

4. **The server will**:
   - Connect to PostgreSQL, MongoDB, and Redis
   - Run database migrations automatically
   - Start the HTTP server on port 8080
   - Start the driver activity monitor in the background

## Migration Path to Microservices

When you need to scale, this monolithic structure naturally splits into:

1. **customer-service**: Customer registration, login, profile
2. **driver-service**: Driver registration, OTP, location tracking
3. **ride-service**: Ride requests, matching, status management
4. **location-service**: Geo-location queries and tracking

Each service already has clear boundaries and can be extracted independently.

## Testing

Example customer registration:
```bash
curl -X POST http://localhost:8080/api/v1/customers/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "01234567890",
    "password": "securepassword"
  }'
```

Example driver OTP request:
```bash
curl -X POST http://localhost:8080/api/v1/drivers/login/request-otp \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "01875113841"
  }'
```

## Security Considerations

1. **Passwords**: Hashed using bcrypt with default cost
2. **JWT**: HMAC-SHA256 signed tokens
3. **OTP**: Stored in Redis with 2-minute TTL
4. **Environment variables**: Never commit secrets to git
5. **Production**: Use proper secret management (AWS Secrets Manager, Vault, etc.)

## Future Enhancements

- [ ] JWT middleware for route protection
- [ ] Rate limiting for OTP requests
- [ ] SMS integration for OTP delivery
- [ ] Real-time WebSocket updates for ride status
- [ ] Fare calculation based on distance
- [ ] Driver ratings and reviews
- [ ] Payment integration
- [ ] Admin dashboard
