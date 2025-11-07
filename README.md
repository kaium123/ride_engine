# Ride Engine - Pathao Ride Simulator (v0.0.1)

A clean architecture implementation of a ride-hailing simulator in Go, similar to Pathao.

## Project Structure

```
ride_engine/
├── cmd/
│   └── api/
│       └── main.go                 # Application entry point
├── internal/
│   └── ride_engine/
│       ├── domain/                 # Domain entities and business logic
│       │   ├── user.go
│       │   ├── location.go
│       │   └── ride.go
│       ├── repository/             # Repository interfaces
│       │   ├── user_repository.go
│       │   ├── driver_repository.go
│       │   ├── ride_repository.go
│       │   └── memory/             # In-memory implementations
│       │       ├── user_memory.go
│       │       ├── driver_memory.go
│       │       └── ride_memory.go
│       ├── usecase/                # Business use cases
│       │   ├── rider_usecase.go
│       │   └── driver_usecase.go
│       └── delivery/               # HTTP delivery layer
│           └── http/
│               ├── response.go
│               ├── rider_handler.go
│               ├── driver_handler.go
│               └── router.go
├── pkg/
│   └── utils/
│       └── id_generator.go
└── README.md
```

## Clean Architecture Layers

1. **Domain Layer**: Contains business entities and rules
2. **Repository Layer**: Defines data persistence interfaces and implementations
3. **Use Case Layer**: Implements business logic and orchestrates data flow
4. **Delivery Layer**: HTTP handlers and routing

## Features

### Rider Features
- Register with email and phone number
- Request a ride with pickup and dropoff locations
- Get ride status
- Cancel a ride

### Driver Features
- Register with phone number
- Login with OTP
- Go online/offline
- Update location continuously
- View nearby ride requests
- Accept, start, complete, or cancel rides

## API Documentation

Base URL: `http://localhost:8080`

### Rider Endpoints

#### 1. Register Rider
```
POST /api/v1/riders
Content-Type: application/json

{
  "phone": "0171xxx",
  "email": "abc@xyz.com"
}

Response (201):
{
  "success": true,
  "data": {
    "id": "xxx",
    "phone": "0171xxx",
    "email": "abc@xyz.com",
    "type": "rider",
    "created_at": "2025-01-04T12:00:00Z"
  }
}
```

#### 2. Request Ride
```
POST /api/v1/rides
Content-Type: application/json

{
  "rider_id": "xxx",
  "pickup_location": {
    "latitude": 23.8103,
    "longitude": 90.4125
  },
  "dropoff_location": {
    "latitude": 23.7509,
    "longitude": 90.3761
  }
}

Response (201):
{
  "success": true,
  "data": {
    "id": "ride_xxx",
    "rider_id": "xxx",
    "pickup_location": {...},
    "dropoff_location": {...},
    "status": "requested",
    "requested_at": "2025-01-04T12:00:00Z"
  }
}
```

#### 3. Get Ride Status
```
GET /api/v1/rides/status?ride_id=ride_xxx

Response (200):
{
  "success": true,
  "data": {
    "id": "ride_xxx",
    "rider_id": "xxx",
    "driver_id": "driver_xxx",
    "status": "accepted",
    ...
  }
}
```

#### 4. Cancel Ride
```
POST /api/v1/rides/cancel?ride_id=ride_xxx
Content-Type: application/json

{
  "rider_id": "xxx"
}

Response (200):
{
  "success": true,
  "message": "Ride cancelled successfully"
}
```

### Driver Endpoints

#### 1. Register Driver
```
POST /api/v1/drivers
Content-Type: application/json

{
  "phone": "0171xxx"
}

Response (201):
{
  "success": true,
  "data": {
    "id": "driver_xxx",
    "phone": "0171xxx",
    "type": "driver",
    "is_online": false,
    "created_at": "2025-01-04T12:00:00Z"
  }
}
```

#### 2. Send OTP
```
POST /api/v1/login
Content-Type: application/json

{
  "phone": "0171xxx"
}

Response (200):
{
  "success": true,
  "message": "OTP sent successfully"
}

Note: OTP will be printed in console for simulation
```

#### 3. Verify OTP (Login)
```
POST /api/v1/login/verify
Content-Type: application/json

{
  "phone": "0171xxx",
  "otp": "123456"
}

Response (200):
{
  "success": true,
  "data": {
    "id": "driver_xxx",
    "phone": "0171xxx",
    "type": "driver",
    "is_online": false
  }
}
```

#### 4. Set Online Status
```
POST /api/v1/drivers/status
Content-Type: application/json

{
  "driver_id": "driver_xxx",
  "is_online": true
}

Response (200):
{
  "success": true,
  "message": "Driver is now online"
}
```

#### 5. Update Location
```
POST /api/v1/drivers/location
Content-Type: application/json

{
  "driver_id": "driver_xxx",
  "latitude": 23.8103,
  "longitude": 90.4125
}

Response (200):
{
  "success": true,
  "message": "Location updated successfully"
}
```

#### 6. Get Nearby Rides
```
GET /api/v1/drivers/rides/nearby?driver_id=driver_xxx&max_distance=10

Response (200):
{
  "success": true,
  "data": [
    {
      "id": "ride_xxx",
      "rider_id": "xxx",
      "pickup_location": {...},
      "dropoff_location": {...},
      "status": "requested"
    }
  ]
}
```

#### 7. Accept Ride
```
POST /api/v1/drivers/rides/accept?ride_id=ride_xxx
Content-Type: application/json

{
  "driver_id": "driver_xxx"
}

Response (200):
{
  "success": true,
  "message": "Ride accepted successfully"
}
```

#### 8. Start Ride
```
POST /api/v1/drivers/rides/start?ride_id=ride_xxx
Content-Type: application/json

{
  "driver_id": "driver_xxx"
}

Response (200):
{
  "success": true,
  "message": "Ride started successfully"
}
```

#### 9. Complete Ride
```
POST /api/v1/drivers/rides/complete?ride_id=ride_xxx
Content-Type: application/json

{
  "driver_id": "driver_xxx"
}

Response (200):
{
  "success": true,
  "message": "Ride completed successfully"
}
```

#### 10. Cancel Ride
```
POST /api/v1/drivers/rides/cancel?ride_id=ride_xxx
Content-Type: application/json

{
  "driver_id": "driver_xxx"
}

Response (200):
{
  "success": true,
  "message": "Ride cancelled successfully"
}
```

## Running the Application

### Prerequisites
- Go 1.21 or higher

### Build and Run

```bash
# Build the application
go build -o bin/ride_engine ./cmd/api

# Run the application
./bin/ride_engine

# Or run directly
go run cmd/api/main.go
```

The server will start on `http://localhost:8080`

## Testing the API

### Example Flow

1. **Register a Rider**
```bash
curl -X POST http://localhost:8080/api/v1/riders \
  -H "Content-Type: application/json" \
  -d '{"phone":"01711111111","email":"rider@example.com"}'
```

2. **Register a Driver**
```bash
curl -X POST http://localhost:8080/api/v1/drivers \
  -H "Content-Type: application/json" \
  -d '{"phone":"01722222222"}'
```

3. **Driver Login - Send OTP**
```bash
curl -X POST http://localhost:8080/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"01722222222"}'
```

4. **Driver Login - Verify OTP** (use OTP from console)
```bash
curl -X POST http://localhost:8080/api/v1/login/verify \
  -H "Content-Type: application/json" \
  -d '{"phone":"01722222222","otp":"123456"}'
```

5. **Driver Goes Online**
```bash
curl -X POST http://localhost:8080/api/v1/drivers/status \
  -H "Content-Type: application/json" \
  -d '{"driver_id":"<driver_id>","is_online":true}'
```

6. **Driver Updates Location**
```bash
curl -X POST http://localhost:8080/api/v1/drivers/location \
  -H "Content-Type: application/json" \
  -d '{"driver_id":"<driver_id>","latitude":23.8103,"longitude":90.4125}'
```

7. **Rider Requests a Ride**
```bash
curl -X POST http://localhost:8080/api/v1/rides \
  -H "Content-Type: application/json" \
  -d '{"rider_id":"<rider_id>","pickup_location":{"latitude":23.8103,"longitude":90.4125},"dropoff_location":{"latitude":23.7509,"longitude":90.3761}}'
```

8. **Driver Views Nearby Rides**
```bash
curl "http://localhost:8080/api/v1/drivers/rides/nearby?driver_id=<driver_id>&max_distance=10"
```

9. **Driver Accepts Ride**
```bash
curl -X POST "http://localhost:8080/api/v1/drivers/rides/accept?ride_id=<ride_id>" \
  -H "Content-Type: application/json" \
  -d '{"driver_id":"<driver_id>"}'
```

10. **Driver Starts Ride**
```bash
curl -X POST "http://localhost:8080/api/v1/drivers/rides/start?ride_id=<ride_id>" \
  -H "Content-Type: application/json" \
  -d '{"driver_id":"<driver_id>"}'
```

11. **Driver Completes Ride**
```bash
curl -X POST "http://localhost:8080/api/v1/drivers/rides/complete?ride_id=<ride_id>" \
  -H "Content-Type: application/json" \
  -d '{"driver_id":"<driver_id>"}'
```

## Ride Status Flow

```
requested → accepted → started → completed
    ↓           ↓          ↓
  cancelled  cancelled  cancelled
```

## Business Rules

1. Only riders can request rides
2. Only drivers with phone numbers can be onboarded
3. Riders must provide both email and phone
4. Driver must be online to accept rides
5. OTP expires in 5 minutes
6. Distance calculation uses Haversine formula
7. Ride status transitions follow a strict flow

## Future Enhancements

- Add database persistence (PostgreSQL/MongoDB)
- Implement real-time location tracking with WebSockets
- Add payment processing
- Implement ride pricing calculation
- Add driver ratings and reviews
- Implement ride history
- Add authentication and authorization (JWT)
- Implement notification service for OTPs
- Add logging and monitoring
- Add unit and integration tests

## License

MIT
