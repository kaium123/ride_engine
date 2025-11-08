package mongodb

import (
	"context"
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

	// Create 2dsphere index on location field for geospatial queries
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
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
