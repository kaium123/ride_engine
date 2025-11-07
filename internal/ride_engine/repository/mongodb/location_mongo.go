package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"vcs.technonext.com/carrybee/ride_engine/pkg/database"
)

var (
	ErrLocationNotFound = errors.New("location not found")
)

// LocationPoint represents a GeoJSON point
type LocationPoint struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"` // [longitude, latitude]
}

// DriverLocation represents a driver's location in MongoDB
type DriverLocation struct {
	DriverID  string        `bson:"driver_id"`
	Location  LocationPoint `bson:"location"`
	Timestamp time.Time     `bson:"timestamp"`
	IsOnline  bool          `bson:"is_online"`
	Metadata  interface{}   `bson:"metadata,omitempty"`
}

// RideLocation represents ride tracking location in MongoDB
type RideLocation struct {
	RideID    string        `bson:"ride_id"`
	DriverID  string        `bson:"driver_id"`
	Location  LocationPoint `bson:"location"`
	Timestamp time.Time     `bson:"timestamp"`
	Status    string        `bson:"status"`
}

type LocationRepository struct {
	db                    *database.MongoDB
	driverLocationsColl   *mongo.Collection
	rideLocationsColl     *mongo.Collection
}

func NewLocationRepository(db *database.MongoDB) *LocationRepository {
	return &LocationRepository{
		db:                    db,
		driverLocationsColl:   db.Collection("driver_locations"),
		rideLocationsColl:     db.Collection("ride_locations"),
	}
}

// SaveDriverLocation saves a driver's location
func (r *LocationRepository) SaveDriverLocation(ctx context.Context, driverID string, lat, lng float64, isOnline bool) error {
	location := DriverLocation{
		DriverID: driverID,
		Location: LocationPoint{
			Type:        "Point",
			Coordinates: []float64{lng, lat}, // GeoJSON uses [longitude, latitude]
		},
		Timestamp: time.Now(),
		IsOnline:  isOnline,
	}

	_, err := r.driverLocationsColl.InsertOne(ctx, location)
	return err
}

// GetDriverLocationHistory gets a driver's location history
func (r *LocationRepository) GetDriverLocationHistory(ctx context.Context, driverID string, limit int64) ([]DriverLocation, error) {
	filter := bson.M{"driver_id": driverID}
	opts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetLimit(limit)

	cursor, err := r.driverLocationsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var locations []DriverLocation
	if err := cursor.All(ctx, &locations); err != nil {
		return nil, err
	}

	return locations, nil
}

// GetLatestDriverLocation gets the most recent location for a driver
func (r *LocationRepository) GetLatestDriverLocation(ctx context.Context, driverID string) (*DriverLocation, error) {
	filter := bson.M{"driver_id": driverID}
	opts := options.FindOne().SetSort(bson.D{{Key: "timestamp", Value: -1}})

	var location DriverLocation
	err := r.driverLocationsColl.FindOne(ctx, filter, opts).Decode(&location)
	if err == mongo.ErrNoDocuments {
		return nil, ErrLocationNotFound
	}
	if err != nil {
		return nil, err
	}

	return &location, nil
}

// FindNearbyDrivers finds drivers within a certain radius (in meters)
func (r *LocationRepository) FindNearbyDrivers(ctx context.Context, lat, lng float64, maxDistanceMeters int) ([]DriverLocation, error) {
	// Using $geoNear aggregation for distance-based search
	pipeline := mongo.Pipeline{
		{{Key: "$geoNear", Value: bson.D{
			{Key: "near", Value: bson.D{
				{Key: "type", Value: "Point"},
				{Key: "coordinates", Value: []float64{lng, lat}},
			}},
			{Key: "distanceField", Value: "distance"},
			{Key: "maxDistance", Value: maxDistanceMeters},
			{Key: "spherical", Value: true},
			{Key: "query", Value: bson.D{
				{Key: "is_online", Value: true},
			}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$driver_id"},
			{Key: "driver_id", Value: bson.D{{Key: "$first", Value: "$driver_id"}}},
			{Key: "location", Value: bson.D{{Key: "$first", Value: "$location"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$first", Value: "$timestamp"}}},
			{Key: "is_online", Value: bson.D{{Key: "$first", Value: "$is_online"}}},
			{Key: "distance", Value: bson.D{{Key: "$first", Value: "$distance"}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "distance", Value: 1}}}},
	}

	cursor, err := r.driverLocationsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var locations []DriverLocation
	if err := cursor.All(ctx, &locations); err != nil {
		return nil, err
	}

	return locations, nil
}

// SaveRideLocation saves a location point during a ride
func (r *LocationRepository) SaveRideLocation(ctx context.Context, rideID, driverID string, lat, lng float64, status string) error {
	location := RideLocation{
		RideID:   rideID,
		DriverID: driverID,
		Location: LocationPoint{
			Type:        "Point",
			Coordinates: []float64{lng, lat},
		},
		Timestamp: time.Now(),
		Status:    status,
	}

	_, err := r.rideLocationsColl.InsertOne(ctx, location)
	return err
}

// GetRideLocationHistory gets the location history for a specific ride
func (r *LocationRepository) GetRideLocationHistory(ctx context.Context, rideID string) ([]RideLocation, error) {
	filter := bson.M{"ride_id": rideID}
	opts := options.Find().SetSort(bson.D{{Key: "timestamp", Value: 1}})

	cursor, err := r.rideLocationsColl.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var locations []RideLocation
	if err := cursor.All(ctx, &locations); err != nil {
		return nil, err
	}

	return locations, nil
}

// DeleteOldDriverLocations deletes driver location records older than the specified duration
func (r *LocationRepository) DeleteOldDriverLocations(ctx context.Context, olderThan time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-olderThan)
	filter := bson.M{"timestamp": bson.M{"$lt": cutoffTime}}

	result, err := r.driverLocationsColl.DeleteMany(ctx, filter)
	if err != nil {
		return 0, err
	}

	return result.DeletedCount, nil
}

// GetOnlineDriversCount gets the count of currently online drivers
func (r *LocationRepository) GetOnlineDriversCount(ctx context.Context) (int64, error) {
	// Get the latest location for each driver and count online ones
	pipeline := mongo.Pipeline{
		{{Key: "$sort", Value: bson.D{{Key: "timestamp", Value: -1}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$driver_id"},
			{Key: "is_online", Value: bson.D{{Key: "$first", Value: "$is_online"}}},
			{Key: "timestamp", Value: bson.D{{Key: "$first", Value: "$timestamp"}}},
		}}},
		{{Key: "$match", Value: bson.D{
			{Key: "is_online", Value: true},
			{Key: "timestamp", Value: bson.D{
				{Key: "$gte", Value: time.Now().Add(-5 * time.Minute)}, // Only count if updated in last 5 minutes
			}},
		}}},
		{{Key: "$count", Value: "online_count"}},
	}

	cursor, err := r.driverLocationsColl.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result []bson.M
	if err := cursor.All(ctx, &result); err != nil {
		return 0, err
	}

	if len(result) == 0 {
		return 0, nil
	}

	count, ok := result[0]["online_count"].(int32)
	if !ok {
		return 0, nil
	}

	return int64(count), nil
}
