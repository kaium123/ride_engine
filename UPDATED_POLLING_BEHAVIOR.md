# Updated Driver Polling Behavior

## Changes Made

### 1. Removed Online Status Check
**Previous behavior:**
- Drivers had to be "online" to view nearby rides
- API would return error if driver was offline

**New behavior:**
- ✅ Any authenticated driver can poll for nearby rides (no online status check)
- Allows drivers to see available rides even before going online
- More flexible for driver app implementations

### 2. Added Time-Based Filtering (5 Minutes)
**MongoDB Query:**
```javascript
{
  "updated_at": { "$gte": <5_minutes_ago> },
  "status": { "$in": ["requested", "pending"] },
  "pickup_location": {
    "$nearSphere": {
      "$geometry": { "type": "Point", "coordinates": [lng, lat] },
      "$maxDistance": maxDistanceMeters
    }
  }
}
```

**Benefits:**
- Only shows **fresh rides** (updated within last 5 minutes)
- Prevents showing stale/abandoned ride requests
- Reduces irrelevant results for drivers

### 3. Added "Pending" Status Support
**Domain Changes:**
```go
const (
    RideStatusRequested RideStatus = "requested"
    RideStatusPending   RideStatus = "pending"   // NEW
    RideStatusAccepted  RideStatus = "accepted"
    RideStatusStarted   RideStatus = "started"
    RideStatusCompleted RideStatus = "completed"
    RideStatusCancelled RideStatus = "cancelled"
)
```

**Now supports:**
- Rides with status "requested"
- Rides with status "pending"
- Both can be accepted by drivers

## Updated Query Logic

### GetNearbyRequestedRides

**File:** `internal/ride_engine/repository/mongodb/ride_mongodb.go:287`

**Filters applied:**
1. ✅ Status: `requested` OR `pending`
2. ✅ Updated within: Last 5 minutes
3. ✅ Distance: Within specified radius (meters)
4. ✅ Sorted by: Distance (closest first)
5. ✅ Limit: 50 rides

### Example Query

```bash
# Driver at coordinates (23.8103, 90.4125) searching within 10km
GET /api/v1/rides/nearby?lat=23.8103&lng=90.4125&max_distance=10000

# Returns rides matching ALL conditions:
# - pickup_location within 10km of driver
# - status is "requested" OR "pending"
# - updated_at >= (now - 5 minutes)
# - sorted by distance (closest first)
# - maximum 50 results
```

## Updated Service Methods

### 1. GetNearbyRides
**File:** `internal/ride_engine/service/ride_service.go:75`

**Changes:**
- ❌ Removed: Online status check
- ✅ Returns: Rides matching MongoDB filters

```go
func (s *RideService) GetNearbyRides(ctx context.Context, driverID int64, driverLat, driverLng, maxDistance float64) ([]*domain.Ride, error) {
    // Use MongoDB geospatial query to find nearby rides efficiently
    rides, err := s.rideRepoMongo.GetNearbyRequestedRides(ctx, driverLat, driverLng, maxDistance)
    if err != nil {
        logger.Error(ctx, "Failed to get nearby requested rides: %v", err)
        return nil, err
    }

    logger.Info(ctx, fmt.Sprintf("Found %d nearby rides for driver %d within %.2fm", len(rides), driverID, maxDistance))

    return rides, nil
}
```

### 2. GetRideRequestsForDriver
**File:** `internal/ride_engine/service/ride_service.go:241`

**Changes:**
- ❌ Removed: Online status check
- ✅ Returns: Rides with customer details

## Testing Examples

### Test Scenario 1: Fresh Ride (Should Appear)
```bash
# 1. Customer creates ride
curl -X POST http://localhost:8080/api/v1/rides \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pickup_lat": 23.8100,
    "pickup_lng": 90.4120,
    "dropoff_lat": 23.7509,
    "dropoff_lng": 90.3761
  }'

# 2. Driver polls immediately (within 5 minutes)
curl -X GET "http://localhost:8080/api/v1/rides/nearby?lat=23.8103&lng=90.4125&max_distance=10000" \
  -H "Authorization: Bearer $DRIVER_TOKEN"

# Expected: Ride appears in results
```

### Test Scenario 2: Stale Ride (Should NOT Appear)
```bash
# 1. Create ride
# 2. Wait 6+ minutes
# 3. Driver polls

# Expected: Ride does NOT appear (updated_at > 5 minutes ago)
```

### Test Scenario 3: Pending Status Ride
```bash
# 1. Create ride with status "pending" directly in MongoDB
db.rides.updateOne(
  { ride_id: 1 },
  { $set: { status: "pending", updated_at: new Date() } }
)

# 2. Driver polls
curl -X GET "http://localhost:8080/api/v1/rides/nearby?lat=23.8103&lng=90.4125&max_distance=10000" \
  -H "Authorization: Bearer $DRIVER_TOKEN"

# Expected: Ride appears in results (status "pending" is included)
```

### Test Scenario 4: Driver Can Accept Pending Ride
```bash
# Accept a pending ride
curl -X POST "http://localhost:8080/api/v1/rides/accept?ride_id=1" \
  -H "Authorization: Bearer $DRIVER_TOKEN"

# Expected: Success (both "requested" and "pending" can be accepted)
```

## MongoDB Index Performance

### Indexes Used
```javascript
// Compound index for efficient querying
db.rides.createIndex({ "status": 1, "updated_at": -1 })

// Geospatial index for location queries
db.rides.createIndex({ "pickup_location": "2dsphere" })
```

### Query Execution Plan
```bash
# Check query performance
db.rides.find({
  status: { $in: ["requested", "pending"] },
  updated_at: { $gte: ISODate("2025-01-09T10:00:00Z") },
  pickup_location: {
    $nearSphere: {
      $geometry: { type: "Point", coordinates: [90.4125, 23.8103] },
      $maxDistance: 10000
    }
  }
}).explain("executionStats")

# Expected: Uses 2dsphere index (IXSCAN stage)
```

## Migration Notes

### If You Have Existing Rides

**Option 1: Keep existing rides active**
```javascript
// Update all old rides to have recent updated_at
db.rides.updateMany(
  { status: "requested" },
  { $set: { updated_at: new Date() } }
)
```

**Option 2: Only keep recent rides**
```javascript
// No action needed - old rides will naturally expire
// after 5 minutes of no updates
```

### Converting "requested" to "pending"
```javascript
// If you want to use "pending" instead of "requested"
db.rides.updateMany(
  { status: "requested" },
  { $set: { status: "pending" } }
)
```

## Benefits of These Changes

### 1. Reduced Stale Data
- Old abandoned rides don't clutter results
- Drivers see only active, relevant rides
- Better user experience

### 2. Flexible Status Management
- Support for both "requested" and "pending" statuses
- Allows for more complex ride state workflows
- Can differentiate between newly requested and pending verification

### 3. No Online Restriction
- Drivers can browse available rides anytime
- App can show ride availability before driver goes online
- More flexibility in driver app UX

### 4. Better Performance
- Time filter reduces dataset size
- Combined with geospatial index = very fast queries
- Scales better as ride volume increases

## Recommendations

### Client-Side Polling Strategy

```javascript
// Only poll for fresh rides
const poller = new DriverRidePoller(token, lat, lng, maxDistance);

// Poll every 5-10 seconds
poller.startPolling(5000);

// Since rides expire after 5 minutes,
// no need to manually filter results
poller.onRidesUpdate = (rides) => {
  // All rides are guaranteed to be < 5 minutes old
  updateUIWithRides(rides);
};
```

### Keep Rides Fresh
If you want a ride to stay available for > 5 minutes:

```javascript
// Option 1: Update ride periodically (e.g., every 2 minutes)
setInterval(async () => {
  await db.rides.updateOne(
    { ride_id: activeRideId },
    { $set: { updated_at: new Date() } }
  );
}, 120000); // Every 2 minutes

// Option 2: Increase the time window in code
// Change: time.Now().Add(-5 * time.Minute)
// To:     time.Now().Add(-15 * time.Minute)  // 15 minutes
```

## Breaking Changes

⚠️ **Important:** These changes may affect existing behavior:

1. **Online Check Removed**
   - Before: Only online drivers could poll
   - After: Any authenticated driver can poll
   - Impact: May need to update driver app logic

2. **Time Filter Added**
   - Before: All "requested" rides shown
   - After: Only rides updated in last 5 minutes
   - Impact: Old rides won't appear unless updated

3. **Pending Status Added**
   - Before: Only "requested" status
   - After: Both "requested" and "pending"
   - Impact: Existing ride acceptance logic updated

## Summary

✅ Removed unnecessary online status check
✅ Added 5-minute freshness filter for rides
✅ Support for "pending" status in addition to "requested"
✅ Better query performance with time-based filtering
✅ More flexible driver polling behavior

The polling endpoint now returns only **recent, relevant rides** sorted by distance!
