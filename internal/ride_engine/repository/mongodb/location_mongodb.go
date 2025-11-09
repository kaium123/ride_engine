package mongodb

import (
	"context"
	"errors"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/repository"
)

// LocationMongoRepository implements LocationRepository using MongoDB
type LocationMongoRepository struct {
	collection *mongo.Collection
}

// NewLocationMongoRepository creates a new MongoDB location repository
func NewLocationMongoRepository(db *mongo.Database) repository.LocationRepository {
	collection := db.Collection("driver_locations")

	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}}, // Create 2dsphere index on location field for geospatial queries
	}
	collection.Indexes().CreateOne(context.Background(), indexModel)

	return &LocationMongoRepository{collection: collection}
}

func (r *LocationMongoRepository) UpdateDriverLocation(ctx context.Context, driverID int64, lat, lng float64) error {
	location := repository.DriverLocation{
		DriverID: driverID,
		Location: repository.GeoJSON{
			Type:        "Point",
			Coordinates: []float64{lng, lat}, // MongoDB uses [longitude, latitude]
		},
		UpdatedAt: time.Now(),
	}

	filter := bson.M{"driver_id": driverID}
	update := bson.M{"$set": location}
	opts := options.Update().SetUpsert(true)

	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Error(ctx, err)
		return err
	}

	return nil
}

func (r *LocationMongoRepository) FindNearestDrivers(ctx context.Context, lat, lng float64, maxDistance float64, limit int) ([]int64, error) {
	cutoffTime := time.Now().Add(-2 * time.Minute) // Only consider drivers whose location was updated within the last 2 minutes

	filter := bson.M{
		"location": bson.M{
			"$nearSphere": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{lng, lat},
				},
				"$maxDistance": maxDistance, // in meters
			},
		},

		"updated_at": bson.M{
			"$gte": cutoffTime, // Filter: only include drivers who updated their location within last 2 minutes
		},
	}

	cursor, err := r.collection.Find(ctx, filter, options.Find().SetLimit(int64(limit)))
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}
	defer cursor.Close(ctx)

	var driverIDs []int64
	for cursor.Next(ctx) {
		var location repository.DriverLocation
		if err := cursor.Decode(&location); err != nil {
			logger.Error(ctx, err)
			continue
		}
		driverIDs = append(driverIDs, location.DriverID)
	}

	return driverIDs, nil
}

func (r *LocationMongoRepository) GetDriverLocation(ctx context.Context, driverID int64) (lat, lng float64, updatedAt *time.Time, err error) {
	filter := bson.M{"driver_id": driverID}

	var location repository.DriverLocation
	err = r.collection.FindOne(ctx, filter).Decode(&location)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return 0, 0, nil, errors.New("driver location not found")
		}
		logger.Error(ctx, err)
		return 0, 0, nil, err
	}

	// Extract coordinates [lng, lat]
	if len(location.Location.Coordinates) >= 2 {
		lng = location.Location.Coordinates[0]
		lat = location.Location.Coordinates[1]
	}

	return lat, lng, &location.UpdatedAt, nil
}
