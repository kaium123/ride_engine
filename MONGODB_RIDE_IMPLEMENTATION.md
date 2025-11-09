# MongoDB Ride Implementation with Driver Short Polling

## Overview

This document describes the implementation of MongoDB-based ride storage with geospatial queries and driver short polling for the ride-hailing application.

## What Was Implemented

### 1. MongoDB Ride Repository (`internal/ride_engine/repository/mongodb/ride_mongodb.go`)

**Features:**
- **Geospatial indexing** for efficient location-based queries
- **Auto-incrementing ride_id** using MongoDB counters collection
- **GeoJSON Point format** for pickup and dropoff locations
- **2dsphere indexes** for geospatial queries

**Key Methods:**
- `Create()` - Creates new ride with auto-generated ride_id
- `GetNearbyRequestedRides()` - **Main geospatial query** for driver polling
- `GetByID()`, `Update()` - Standard CRUD operations
- `GetRequestedRides()` - Get all rides with "requested" status
- `GetByCustomerID()`, `GetByDriverID()` - Get rides by user

**Indexes Created:**
```javascript
// Geospatial indexes
{ "pickup_location": "2dsphere" }
{ "dropoff_location": "2dsphere" }

// Performance indexes
{ "status": 1 }
{ "customer_id": 1 }
{ "driver_id": 1 }
{ "status": 1, "requested_at": -1 }  // Compound index for efficient polling
{ "ride_id": 1 }  // Unique index
```

### 2. Updated Ride Service (`internal/ride_engine/service/ride_service.go`)

**Changes:**
- Replaced PostgreSQL repository with MongoDB repository
- Updated `GetNearbyRides()` to use MongoDB geospatial query (`$nearSphere`)
- Updated `GetRideRequestsForDriver()` for efficient driver polling
- All ride operations now use MongoDB

**Geospatial Query:**
```go
// MongoDB automatically sorts by distance from driver's location
rides := s.rideRepoMongo.GetNearbyRequestedRides(ctx, driverLat, driverLng, maxDistance)
```

### 3. Driver Polling Endpoint

**Route:** `GET /api/v1/rides/nearby`

**Authentication:** Bearer token (driver must be logged in)

**Query Parameters:**
- `lat` (required) - Driver's current latitude
- `lng` (required) - Driver's current longitude
- `max_distance` (optional) - Maximum search radius in meters (default: 10,000m = 10km)

**Response:**
```json
[
  {
    "id": 1,
    "customer_id": 123,
    "pickup_lat": 23.8103,
    "pickup_lng": 90.4125,
    "dropoff_lat": 23.7509,
    "dropoff_lng": 90.3761,
    "status": "requested",
    "requested_at": "2025-01-09T10:30:00Z"
  }
]
```

**Business Rules:**
- Only **online drivers** can poll for rides
- Only returns rides with status **"requested"**
- Results are automatically sorted by distance (closest first)
- Limited to 50 rides per query

### 4. Ride Document Schema

```javascript
{
  _id: ObjectId,
  ride_id: 1,  // Auto-incrementing sequence
  customer_id: 123,
  driver_id: null,  // Optional, set when accepted

  // GeoJSON for geospatial queries
  pickup_location: {
    type: "Point",
    coordinates: [90.4125, 23.8103]  // [lng, lat]
  },
  dropoff_location: {
    type: "Point",
    coordinates: [90.3761, 23.7509]  // [lng, lat]
  },

  // Regular fields for compatibility
  pickup_lat: 23.8103,
  pickup_lng: 90.4125,
  dropoff_lat: 23.7509,
  dropoff_lng: 90.3761,

  status: "requested",
  fare: null,
  requested_at: ISODate("2025-01-09T10:30:00Z"),
  accepted_at: null,
  started_at: null,
  completed_at: null,
  cancelled_at: null,
  created_at: ISODate("2025-01-09T10:30:00Z"),
  updated_at: ISODate("2025-01-09T10:30:00Z")
}
```

## How Driver Short Polling Works

### Flow:

1. **Driver goes online** (sets status to online in database)
2. **Driver's app polls** `GET /api/v1/rides/nearby?lat=23.8103&lng=90.4125&max_distance=10000`
3. **Server checks** if driver is online
4. **MongoDB query** finds rides within radius using `$nearSphere` operator
5. **Returns** sorted list of nearby available rides
6. **Driver selects** a ride and accepts it via `POST /api/v1/rides/accept`

### Short Polling Recommendations:

**Client-side implementation:**
```javascript
// Poll every 5-10 seconds when driver is idle
setInterval(async () => {
  const response = await fetch('/api/v1/rides/nearby?lat=23.8103&lng=90.4125&max_distance=10000', {
    headers: { 'Authorization': 'Bearer ' + token }
  });
  const rides = await response.json();

  // Update UI with available rides
  updateRidesList(rides);
}, 5000);  // Poll every 5 seconds
```

**Benefits of MongoDB geospatial queries:**
- **O(log n) complexity** for nearby queries (vs O(n) for PostgreSQL)
- **Automatic distance sorting** from driver's location
- **Efficient at scale** - handles millions of rides
- **Native GeoJSON support** - no custom distance calculations needed

## API Usage Examples

### Customer Requests a Ride
```bash
curl -X POST http://localhost:8080/api/v1/rides \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <customer_token>" \
  -d '{
    "pickup_lat": 23.8103,
    "pickup_lng": 90.4125,
    "dropoff_lat": 23.7509,
    "dropoff_lng": 90.3761
  }'
```

**Response:**
```json
{
  "id": 1,
  "customer_id": 123,
  "pickup_lat": 23.8103,
  "pickup_lng": 90.4125,
  "dropoff_lat": 23.7509,
  "dropoff_lng": 90.3761,
  "status": "requested",
  "requested_at": "2025-01-09T10:30:00Z"
}
```

### Driver Polls for Nearby Rides
```bash
curl -X GET "http://localhost:8080/api/v1/rides/nearby?lat=23.8100&lng=90.4120&max_distance=5000" \
  -H "Authorization: Bearer <driver_token>"
```

**Response:**
```json
[
  {
    "id": 1,
    "customer_id": 123,
    "pickup_lat": 23.8103,
    "pickup_lng": 90.4125,
    "dropoff_lat": 23.7509,
    "dropoff_lng": 90.3761,
    "status": "requested",
    "requested_at": "2025-01-09T10:30:00Z"
  },
  {
    "id": 2,
    "customer_id": 456,
    "pickup_lat": 23.8090,
    "pickup_lng": 90.4115,
    "dropoff_lat": 23.7600,
    "dropoff_lng": 90.3800,
    "status": "requested",
    "requested_at": "2025-01-09T10:28:00Z"
  }
]
```

### Driver Accepts a Ride
```bash
curl -X POST "http://localhost:8080/api/v1/rides/accept?ride_id=1" \
  -H "Authorization: Bearer <driver_token>"
```

**Response:**
```json
{
  "message": "Ride accepted successfully"
}
```

## Performance Considerations

### MongoDB Geospatial Query Performance

**Traditional approach (PostgreSQL):**
```sql
-- O(n) - checks ALL rides
SELECT * FROM rides
WHERE status = 'requested'
AND ST_Distance_Sphere(
  POINT(pickup_lng, pickup_lat),
  POINT(90.4125, 23.8103)
) <= 10000;
```

**MongoDB approach:**
```javascript
// O(log n) - uses 2dsphere index
db.rides.find({
  status: "requested",
  pickup_location: {
    $nearSphere: {
      $geometry: { type: "Point", coordinates: [90.4125, 23.8103] },
      $maxDistance: 10000
    }
  }
}).limit(50)
```

### Polling Frequency Recommendations

| Driver State | Polling Interval | Reason |
|--------------|------------------|--------|
| Idle/Waiting | 5-10 seconds | Catch new rides quickly |
| On a ride | Stop polling | Driver unavailable |
| Offline | Stop polling | Driver not available |

### Scaling Considerations

- **Connection pooling**: Configure MongoDB connection pool size
- **Index monitoring**: Monitor index usage with `db.rides.getIndexes()`
- **TTL indexes**: Consider adding TTL for old completed rides
- **Sharding**: Shard rides collection by location for global scale

## Migration from PostgreSQL

If you have existing rides in PostgreSQL, you can migrate them:

```javascript
// Migration script (pseudo-code)
db.rides.insertMany(
  postgresRides.map(ride => ({
    ride_id: ride.id,
    customer_id: ride.customer_id,
    driver_id: ride.driver_id,
    pickup_location: {
      type: "Point",
      coordinates: [ride.pickup_lng, ride.pickup_lat]
    },
    dropoff_location: {
      type: "Point",
      coordinates: [ride.dropoff_lng, ride.dropoff_lat]
    },
    pickup_lat: ride.pickup_lat,
    pickup_lng: ride.pickup_lng,
    dropoff_lat: ride.dropoff_lat,
    dropoff_lng: ride.dropoff_lng,
    status: ride.status,
    fare: ride.fare,
    requested_at: ride.requested_at,
    accepted_at: ride.accepted_at,
    started_at: ride.started_at,
    completed_at: ride.completed_at,
    cancelled_at: ride.cancelled_at,
    created_at: new Date(),
    updated_at: new Date()
  }))
)
```

## Testing

### Test Nearby Rides Query

```bash
# 1. Start MongoDB and app
make docker-up
go run cmd/api/main.go

# 2. Create test customer and driver
# 3. Customer requests ride
# 4. Driver goes online
# 5. Driver polls for rides

# Expected: Driver sees the requested ride
```

### MongoDB Shell Verification

```javascript
// Check indexes
db.rides.getIndexes()

// Check rides count
db.rides.countDocuments({ status: "requested" })

// Test geospatial query manually
db.rides.find({
  status: "requested",
  pickup_location: {
    $nearSphere: {
      $geometry: { type: "Point", coordinates: [90.4125, 23.8103] },
      $maxDistance: 10000
    }
  }
}).limit(10)
```

## Future Enhancements

1. **Real-time updates** - Replace short polling with WebSockets or Server-Sent Events
2. **Driver matching algorithm** - Automatically assign rides based on proximity and availability
3. **Ride analytics** - Use MongoDB aggregation pipeline for insights
4. **Ride history** - Archive completed rides to separate collection
5. **Fare estimation** - Calculate fare based on distance (use $geoNear for distance calculation)

## Conclusion

The implementation successfully:
- ✅ Stores ride data in MongoDB with geospatial indexing
- ✅ Implements efficient driver short polling endpoint
- ✅ Uses MongoDB's native geospatial queries for performance
- ✅ Maintains backward compatibility with existing API
- ✅ Only online drivers can view available rides
- ✅ Results sorted by distance automatically
