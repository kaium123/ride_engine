// Initialize MongoDB database for Ride Engine Location Tracking
// This script sets up collections and indexes for geospatial queries

// Use the ride_engine database
db = db.getSiblingDB('ride_engine');

print("Initializing MongoDB for ride_engine...");

// Drop existing collections if they exist (use with caution in production)
db.driver_locations.drop();

// Create collection for driver locations with schema validation
db.createCollection('driver_locations', {
    validator: {
        $jsonSchema: {
            bsonType: "object",
            required: ["driver_id", "location", "updated_at"],
            properties: {
                driver_id: {
                    bsonType: "long",
                    description: "Driver ID from PostgreSQL (BIGINT) - required"
                },
                location: {
                    bsonType: "object",
                    required: ["type", "coordinates"],
                    properties: {
                        type: {
                            enum: ["Point"],
                            description: "GeoJSON type - must be 'Point'"
                        },
                        coordinates: {
                            bsonType: "array",
                            minItems: 2,
                            maxItems: 2,
                            items: {
                                bsonType: "double"
                            },
                            description: "[longitude, latitude] - required"
                        }
                    },
                    description: "GeoJSON Point object for driver location"
                },
                updated_at: {
                    bsonType: "date",
                    description: "Timestamp of last location update - required"
                }
            }
        }
    }
});

print("✓ Collection 'driver_locations' created");

// Create 2dsphere geospatial index for location queries
// This enables efficient $nearSphere queries
db.driver_locations.createIndex(
    { location: "2dsphere" },
    { name: "location_2dsphere" }
);

print("✓ 2dsphere geospatial index created on 'location'");

// Create index on driver_id for efficient lookups and updates
db.driver_locations.createIndex(
    { driver_id: 1 },
    { unique: true, name: "driver_id_unique" }
);

print("✓ Unique index created on 'driver_id'");

// Create compound index for driver_id and updated_at
db.driver_locations.createIndex(
    { driver_id: 1, updated_at: -1 },
    { name: "driver_updated_compound" }
);

print("✓ Compound index created on 'driver_id' and 'updated_at'");

// Create TTL index to auto-delete stale location data after 7 days
db.driver_locations.createIndex(
    { updated_at: 1 },
    {
        expireAfterSeconds: 604800,  // 7 days in seconds
        name: "updated_at_ttl"
    }
);

print("✓ TTL index created (auto-delete after 7 days)");

// Insert sample driver location data
db.driver_locations.insertOne({
    driver_id: NumberLong(1),
    location: {
        type: "Point",
        coordinates: [90.4125, 23.8103]  // Dhaka coordinates [lng, lat]
    },
    updated_at: new Date()
});

print("✓ Sample driver location inserted");

// Display collection stats
print("\n=== Collection Statistics ===");
const stats = db.driver_locations.stats();
print("Collection: driver_locations");
print("Document count: " + stats.count);
print("Storage size: " + stats.storageSize + " bytes");
print("Index count: " + stats.nindexes);

// Display indexes
print("\n=== Indexes ===");
db.driver_locations.getIndexes().forEach(function(index) {
    print("- " + index.name + ": " + JSON.stringify(index.key));
});

// Test geospatial query
print("\n=== Testing Geospatial Query ===");
const nearbyDrivers = db.driver_locations.find({
    location: {
        $nearSphere: {
            $geometry: {
                type: "Point",
                coordinates: [90.4100, 23.8100]
            },
            $maxDistance: 5000  // 5km radius
        }
    }
}).toArray();

print("Found " + nearbyDrivers.length + " driver(s) within 5km");

print("\n✓ MongoDB initialization completed successfully!");
print("\nUsage Examples:");
print("================");
print("\n1. Update driver location:");
print('   db.driver_locations.updateOne(');
print('       { driver_id: NumberLong(1) },');
print('       { $set: {');
print('           location: { type: "Point", coordinates: [90.4125, 23.8103] },');
print('           updated_at: new Date()');
print('       }},');
print('       { upsert: true }');
print('   );');
print("\n2. Find drivers within 10km radius:");
print('   db.driver_locations.find({');
print('       location: {');
print('           $nearSphere: {');
print('               $geometry: { type: "Point", coordinates: [90.4100, 23.8100] },');
print('               $maxDistance: 10000');
print('           }');
print('       }');
print('   });');
print("\n3. Get specific driver location:");
print('   db.driver_locations.findOne({ driver_id: NumberLong(1) });');
