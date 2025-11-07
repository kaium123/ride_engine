package service

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LocationService struct {
	collection *mongo.Collection
}

type DriverLocation struct {
	DriverID  int64     `bson:"driver_id"`
	Location  GeoJSON   `bson:"location"`
	UpdatedAt time.Time `bson:"updated_at"`
}

type GeoJSON struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"` // [longitude, latitude]
}

func NewLocationService(db *mongo.Database) *LocationService {
	collection := db.Collection("driver_locations")

	// Create 2dsphere index on location field
	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	}
	collection.Indexes().CreateOne(context.Background(), indexModel)

	return &LocationService{collection: collection}
}

// UpdateDriverLocation updates driver's current location
func (s *LocationService) UpdateDriverLocation(ctx context.Context, driverID int64, lat, lng float64) error {
	location := DriverLocation{
		DriverID: driverID,
		Location: GeoJSON{
			Type:        "Point",
			Coordinates: []float64{lng, lat}, // MongoDB uses [longitude, latitude]
		},
		UpdatedAt: time.Now(),
	}

	filter := bson.M{"driver_id": driverID}
	update := bson.M{"$set": location}
	opts := options.Update().SetUpsert(true)

	_, err := s.collection.UpdateOne(ctx, filter, update, opts)
	return err
}

// FindNearestDrivers finds drivers within maxDistance (in meters)
func (s *LocationService) FindNearestDrivers(ctx context.Context, lat, lng float64, maxDistance float64, limit int) ([]int64, error) {
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

	cursor, err := s.collection.Find(ctx, filter, options.Find().SetLimit(int64(limit)))
	if err != nil {
		return nil, fmt.Errorf("failed to find nearest drivers: %w", err)
	}
	defer cursor.Close(ctx)

	var driverIDs []int64
	for cursor.Next(ctx) {
		var location DriverLocation
		if err := cursor.Decode(&location); err != nil {
			continue
		}
		driverIDs = append(driverIDs, location.DriverID)
	}

	return driverIDs, nil
}
