# Driver Polling Quick Start Guide

## Quick Setup & Testing

### 1. Start the Application

```bash
# Start databases (MongoDB, PostgreSQL, Redis)
docker-compose up -d

# Run the application
go run cmd/api/main.go
```

### 2. Test Flow

#### Step 1: Register and Login Customer
```bash
# Register customer
curl -X POST http://localhost:8080/api/v1/customers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "01711111111",
    "password": "password123"
  }'

# Login customer
curl -X POST http://localhost:8080/api/v1/customers/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "01711111111",
    "password": "password123"
  }'

# Save the token from response
CUSTOMER_TOKEN="<token_from_response>"
```

#### Step 2: Register and Login Driver
```bash
# Register driver
curl -X POST http://localhost:8080/api/v1/drivers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Driver One",
    "phone": "01722222222",
    "vehicle_no": "DHA-1234"
  }'

# Send OTP
curl -X POST http://localhost:8080/api/v1/drivers/send-otp \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "01722222222"
  }'

# Check console for OTP, then verify
curl -X POST http://localhost:8080/api/v1/drivers/verify-otp \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "01722222222",
    "otp": "<otp_from_console>"
  }'

# Save the token from response
DRIVER_TOKEN="<token_from_response>"
```

#### Step 3: Driver Goes Online and Updates Location
```bash
# Set driver online
curl -X POST http://localhost:8080/api/v1/drivers/online \
  -H "Authorization: Bearer $DRIVER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "is_online": true
  }'

# Update driver location (Dhaka coordinates)
curl -X POST http://localhost:8080/api/v1/drivers/location \
  -H "Authorization: Bearer $DRIVER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "latitude": 23.8103,
    "longitude": 90.4125
  }'
```

#### Step 4: Customer Requests a Ride
```bash
# Request ride near driver's location
curl -X POST http://localhost:8080/api/v1/rides \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pickup_lat": 23.8100,
    "pickup_lng": 90.4120,
    "dropoff_lat": 23.7509,
    "dropoff_lng": 90.3761
  }'

# Note the ride_id from response
RIDE_ID=<ride_id_from_response>
```

#### Step 5: Driver Polls for Nearby Rides (SHORT POLLING)
```bash
# Driver polls every 5-10 seconds
curl -X GET "http://localhost:8080/api/v1/rides/nearby?lat=23.8103&lng=90.4125&max_distance=10000" \
  -H "Authorization: Bearer $DRIVER_TOKEN"

# Response should show the nearby ride:
# [
#   {
#     "id": 1,
#     "customer_id": 123,
#     "pickup_lat": 23.8100,
#     "pickup_lng": 90.4120,
#     "dropoff_lat": 23.7509,
#     "dropoff_lng": 90.3761,
#     "status": "requested",
#     "requested_at": "2025-01-09T10:30:00Z"
#   }
# ]
```

#### Step 6: Driver Accepts the Ride
```bash
curl -X POST "http://localhost:8080/api/v1/rides/accept?ride_id=$RIDE_ID" \
  -H "Authorization: Bearer $DRIVER_TOKEN"
```

#### Step 7: Complete Ride Flow
```bash
# Start ride
curl -X POST "http://localhost:8080/api/v1/rides/start?ride_id=$RIDE_ID" \
  -H "Authorization: Bearer $DRIVER_TOKEN"

# Complete ride
curl -X POST "http://localhost:8080/api/v1/rides/complete?ride_id=$RIDE_ID" \
  -H "Authorization: Bearer $DRIVER_TOKEN"
```

## Short Polling Client Example (JavaScript)

```javascript
// Driver app - polls for nearby rides every 5 seconds
class DriverRidePoller {
  constructor(authToken, lat, lng, maxDistance = 10000) {
    this.authToken = authToken;
    this.lat = lat;
    this.lng = lng;
    this.maxDistance = maxDistance;
    this.pollingInterval = null;
    this.isPolling = false;
  }

  async fetchNearbyRides() {
    try {
      const response = await fetch(
        `http://localhost:8080/api/v1/rides/nearby?lat=${this.lat}&lng=${this.lng}&max_distance=${this.maxDistance}`,
        {
          headers: {
            'Authorization': `Bearer ${this.authToken}`
          }
        }
      );

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const rides = await response.json();
      console.log(`Found ${rides.length} nearby rides`);
      this.onRidesUpdate(rides);

      return rides;
    } catch (error) {
      console.error('Error fetching nearby rides:', error);
      this.onError(error);
    }
  }

  startPolling(intervalMs = 5000) {
    if (this.isPolling) {
      console.log('Already polling');
      return;
    }

    console.log(`Starting polling every ${intervalMs}ms`);
    this.isPolling = true;

    // Fetch immediately
    this.fetchNearbyRides();

    // Then poll at interval
    this.pollingInterval = setInterval(() => {
      this.fetchNearbyRides();
    }, intervalMs);
  }

  stopPolling() {
    if (this.pollingInterval) {
      clearInterval(this.pollingInterval);
      this.pollingInterval = null;
      this.isPolling = false;
      console.log('Stopped polling');
    }
  }

  updateLocation(lat, lng) {
    this.lat = lat;
    this.lng = lng;
  }

  // Override these methods in your UI
  onRidesUpdate(rides) {
    // Update UI with new rides
    console.log('Rides updated:', rides);
  }

  onError(error) {
    // Handle error in UI
    console.error('Polling error:', error);
  }
}

// Usage:
const poller = new DriverRidePoller(
  'driver_auth_token_here',
  23.8103,  // driver's latitude
  90.4125,  // driver's longitude
  10000     // search radius in meters
);

// Customize callbacks
poller.onRidesUpdate = (rides) => {
  // Update your UI with rides
  document.getElementById('rides-list').innerHTML = rides.map(ride => `
    <div class="ride-card">
      <p>Ride #${ride.id}</p>
      <p>Pickup: ${ride.pickup_lat}, ${ride.pickup_lng}</p>
      <p>Dropoff: ${ride.dropoff_lat}, ${ride.dropoff_lng}</p>
      <button onclick="acceptRide(${ride.id})">Accept</button>
    </div>
  `).join('');
};

// Start polling when driver goes online
poller.startPolling(5000);  // Poll every 5 seconds

// Stop polling when driver goes offline or accepts a ride
// poller.stopPolling();
```

## Testing with Different Locations

### Dhaka, Bangladesh (Test Locations)

```bash
# Driver at Gulshan
lat=23.7808, lng=90.4156

# Customer at Banani (2km away)
pickup_lat=23.7956, pickup_lng=90.4063

# Customer at Dhanmondi (8km away)
pickup_lat=23.7461, pickup_lng=90.3742

# Customer at Mirpur (15km away - should NOT appear with 10km radius)
pickup_lat=23.8223, pickup_lng=90.3654
```

### Test Different Radii

```bash
# 1km radius
curl -X GET "http://localhost:8080/api/v1/rides/nearby?lat=23.8103&lng=90.4125&max_distance=1000" \
  -H "Authorization: Bearer $DRIVER_TOKEN"

# 5km radius
curl -X GET "http://localhost:8080/api/v1/rides/nearby?lat=23.8103&lng=90.4125&max_distance=5000" \
  -H "Authorization: Bearer $DRIVER_TOKEN"

# 10km radius (default)
curl -X GET "http://localhost:8080/api/v1/rides/nearby?lat=23.8103&lng=90.4125&max_distance=10000" \
  -H "Authorization: Bearer $DRIVER_TOKEN"
```

## Monitoring MongoDB

### Check Ride Documents
```bash
# Connect to MongoDB container
docker exec -it <mongodb_container_name> mongosh

# Use database
use ride_engine

# Check all rides
db.rides.find().pretty()

# Check requested rides
db.rides.find({ status: "requested" }).pretty()

# Check geospatial query
db.rides.find({
  status: "requested",
  pickup_location: {
    $nearSphere: {
      $geometry: { type: "Point", coordinates: [90.4125, 23.8103] },
      $maxDistance: 10000
    }
  }
}).limit(10).pretty()

# Check indexes
db.rides.getIndexes()

# Check counters collection
db.counters.find().pretty()
```

## Troubleshooting

### Driver Not Seeing Rides

**Check:**
1. Is driver online? `is_online: true`
2. Is driver's location updated recently?
3. Is ride status "requested"?
4. Is ride within max_distance radius?
5. Are MongoDB indexes created?

```bash
# Verify driver is online
curl -X GET http://localhost:8080/api/v1/drivers/me \
  -H "Authorization: Bearer $DRIVER_TOKEN"

# Check MongoDB indexes
db.rides.getIndexes()
```

### No Rides in Response

```bash
# Check if there are any requested rides in MongoDB
db.rides.countDocuments({ status: "requested" })

# If 0, create a test ride
curl -X POST http://localhost:8080/api/v1/rides \
  -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "pickup_lat": 23.8100,
    "pickup_lng": 90.4120,
    "dropoff_lat": 23.7509,
    "dropoff_lng": 90.3761
  }'
```

### Geospatial Query Not Working

```bash
# Verify 2dsphere index exists
db.rides.getIndexes()

# Should see:
# { "pickup_location": "2dsphere" }

# If not, create it manually:
db.rides.createIndex({ "pickup_location": "2dsphere" })
```

## Performance Tips

1. **Polling Frequency**: 5-10 seconds is optimal (balance between responsiveness and server load)
2. **Max Distance**: Use 10km for urban areas, 20km for suburban
3. **Stop Polling**: Always stop polling when driver accepts a ride or goes offline
4. **Connection Pooling**: Reuse HTTP connections in production
5. **Error Handling**: Implement exponential backoff on errors

## Production Considerations

1. **Rate Limiting**: Add rate limiting to prevent abuse (e.g., max 12 requests/minute per driver)
2. **Caching**: Cache results for 5 seconds to reduce MongoDB load
3. **Monitoring**: Track polling frequency and response times
4. **Upgrade to WebSockets**: For real-time updates, replace short polling with WebSockets
5. **Load Balancing**: Use MongoDB replica set for read scaling

## Next Steps

- Implement WebSockets for real-time ride updates
- Add push notifications for ride requests
- Implement automatic driver-ride matching
- Add ride history and analytics
- Implement fare calculation based on distance
