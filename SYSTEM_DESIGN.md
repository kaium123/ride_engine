# Ride Engine - System Design

## Overview
Ride-sharing platform with geospatial matching, short polling, dual authentication (customers: email/password, drivers: OTP).

**Stack**: Go + Echo + PostgreSQL + MongoDB + Redis

---

## Architecture

```
┌──────────────┐     ┌──────────────┐
│   Customer   │     │    Driver    │
│  Mobile App  │     │  Mobile App  │
└──────┬───────┘     └──────┬───────┘
       │                    │
       └────────┬───────────┘
                │
         ┌──────▼──────┐
         │  API Server │ (Echo v4, JWT Auth)
         │   :8080     │
         └──────┬──────┘
                │
    ┌───────────┼───────────┐
    │           │           │
┌───▼────┐  ┌──▼───┐  ┌────▼────┐
│Postgres│  │MongoDB│ │  Redis  │
│ :5432  │  │:27016 │ │  :6379  │
└────────┘  └───────┘ └─────────┘
```

---

## Components

### 1. Client Layer
- **Customer**: Email/password auth, request rides, track status
- **Driver**: OTP auth, poll nearby rides, manage ride lifecycle

### 2. API Layer (Echo Framework)
**Routes**:
- `/customers` - register, login (email/password)
- `/drivers` - register, login (OTP), update location
- `/rides` - request, nearby, accept, start, complete, cancel, status

### 3. Service Layer
- **CustomerService**: Email/password auth (bcrypt), JWT tokens
- **DriverService**: OTP auth, location tracking
- **RideService**: Geospatial matching, ride lifecycle, status tracking
- **LocationService**: MongoDB geospatial queries (2dsphere)
- **OTPService**: 6-digit OTP generation, Redis cache + PostgreSQL

### 4. Data Layer
**PostgreSQL**: Users (customers, drivers), OTPs
**MongoDB**: Rides (GeoJSON), driver locations (2dsphere indexes)
**Redis**: JWT tokens, OTP cache, sessions


---

## Key Features

### Geospatial Matching
```javascript
// Driver polls every 5s for rides within 5km
cutoffTime := time.Now().Add(-5 * time.Minute) // Calculate cutoff time (5 minutes ago)

filter := bson.M{
  "status": bson.M{
    "$in": []string{"requested", "pending"}, // Support both requested and pending status
  },
  "updated_at": bson.M{
    "$gte": cutoffTime,
  },
  "pickup_location": bson.M{
    "$nearSphere": bson.M{
      "$geometry": bson.M{
        "type":        "Point",
                "coordinates": []float64{lng, lat},
      },
      "$maxDistance": maxDistanceMeters, // in meters
    },
  },
}

opts := options.Find().SetLimit(int64(limit))

cursor, err := r.collection.Find(ctx, filter, opts)
if err != nil {
  logger.Error(ctx, "Failed to get nearby requested rides", err)
  return nil, err
}
defer cursor.Close(ctx)

var rides []*domain.Ride
for cursor.Next(ctx) {
  var doc RideDocument
  if err := cursor.Decode(&doc); err != nil {
    logger.Error(ctx, "Failed to decode ride", err)
    continue
  }
  rides = append(rides, toRideDomain(&doc))
}
```

### Dual Authentication
- **Customer**: bcrypt password → JWT
- **Driver**: 6-digit OTP → JWT
- All JWTs stored in Redis

### Ride Lifecycle
```
requested → accepted → started → completed
   ↓           ↓          ↓
   └───────────┴──────────┴──────→ cancelled
```

### Short Polling
- Driver: poll nearby rides every 5s
- Customer: poll ride status every 3s
- Why? Stateless, horizontally scalable

---

## Database Schema

### PostgreSQL
- **customers**: id, name, email, phone, password (bcrypt)
- **drivers**: id, name, phone, vehicle_no, is_online, current_lat/lng
- **otps**: phone, otp, expires_at

### MongoDB
- **rides**: ride_id, customer_id, driver_id, pickup_location (GeoJSON), dropoff_location (GeoJSON),pickup_location (lat/lng), dropoff_location (lat/lng), status
  - Indexes: 2dsphere on pickup_location, (status, updated_at)
- **driver_locations**: driver_id, location (GeoJSON), updated_at
  - Indexes: 2dsphere on location

### Redis
- `otp:{phone}` (TTL: 2min)
- `jwt:user:{id}` (TTL: configurable)

---

## API Examples

**Customer Request Ride**:
```
POST /rides/ { pickup_lat, pickup_lng, dropoff_lat, dropoff_lng }
→ {
    "id": 22,
    "customer_id": 1,
    "pickup_lat": 23.23,
    "pickup_lng": 90.24,
    "dropoff_lat": 25.23,
    "dropoff_lng": 95.24,
    "status": "requested",
    "requested_at": "2025-11-09T11:26:00.506090169Z"
}
```

**Driver Poll Nearby**:
```
POST /rides/nearby { lat, lng, max_distance: 5000, limit: 10 }
→ [
    {
        "id": 21,
        "customer_id": 10,
        "pickup_lat": 23.7919,
        "pickup_lng": 90.4158,
        "dropoff_lat": 23.8119,
        "dropoff_lng": 90.4358,
        "status": "requested",
        "requested_at": "2025-11-09T11:22:50.875Z"
    },
    {
        "id": 20,
        "customer_id": 9,
        "pickup_lat": 23.7906,
        "pickup_lng": 90.4146,
        "dropoff_lat": 23.8106,
        "dropoff_lng": 90.4346,
        "status": "requested",
        "requested_at": "2025-11-09T11:22:50.786Z"
    },
    {
        "id": 19,
        "customer_id": 8,
        "pickup_lat": 23.7893,
        "pickup_lng": 90.4134,
        "dropoff_lat": 23.8093,
        "dropoff_lng": 90.4334,
        "status": "requested",
        "requested_at": "2025-11-09T11:22:50.696Z"
    }
]
```

**Driver Accept**:
```
POST /rides/accept { ride_id: 1 }
→ { message: "ride accepted" }
```

**Customer Track**:
```
GET /rides/status?ride_id=1
→ {
    "ride_id": 11,
    "customer_id": 1,
    "pickup_lat": 23.23,
    "pickup_lng": 90.24,
    "dropoff_lat": 25.23,
    "dropoff_lng": 95.24,
    "status": "started",
    "requested_at": "2025-11-09 11:13:52",
    "accepted_at": "2025-11-09 11:15:49",
    "started_at": "2025-11-09 11:18:32",
    "driver": {
        "driver_id": 1,
        "name": "",
        "phone": "01875113841",
        "vehicle_no": "",
        "current_lat": 23.7925,
        "current_lng": 90.4078,
        "last_ping_at": "2025-11-09 11:25:20"
    }
}
```

---

## Deployment

```yaml
# docker-compose.yml
services:
  api: { ports: [8080] }
  postgres: { image: postgres:15 }
  mongodb: { image: mongo:7 }
  redis: { image: redis:7-alpine }
```

**Production**:
- Load balancer + 3+ API instances
- PostgreSQL master-replica
- MongoDB replica set (3 nodes)
- Redis cluster

---

## Performance

- MongoDB 2dsphere indexes: O(log n) queries
- 5-minute freshness filter
- Redis caching (OTP, sessions)
- Target: API <100ms, geospatial <50ms

---

## Security

- Passwords: bcrypt
- JWT: stored in Redis, role-based access
- OTP: 2-minute expiry
- Ownership validation: users only see their data
