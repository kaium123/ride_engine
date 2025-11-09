package mongodb

import (
	"context"
	"errors"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"vcs.technonext.com/carrybee/ride_engine/internal/ride_engine/domain"
)

var (
	ErrRideNotFound = errors.New("ride not found")
)

// GeoJSONPoint represents a GeoJSON point for MongoDB geospatial queries
type GeoJSONPoint struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"` // [longitude, latitude]
}

// RideDocument represents a ride in MongoDB
type RideDocument struct {
	ID              primitive.ObjectID `bson:"_id,omitempty"`
	RideID          int64              `bson:"ride_id"`
	CustomerID      int64              `bson:"customer_id"`
	DriverID        *int64             `bson:"driver_id,omitempty"`
	PickupLocation  GeoJSONPoint       `bson:"pickup_location"`
	DropoffLocation GeoJSONPoint       `bson:"dropoff_location"`
	PickupLat       float64            `bson:"pickup_lat"`
	PickupLng       float64            `bson:"pickup_lng"`
	DropoffLat      float64            `bson:"dropoff_lat"`
	DropoffLng      float64            `bson:"dropoff_lng"`
	Status          string             `bson:"status"`
	Fare            *float64           `bson:"fare,omitempty"`
	RequestedAt     time.Time          `bson:"requested_at"`
	AcceptedAt      *time.Time         `bson:"accepted_at,omitempty"`
	StartedAt       *time.Time         `bson:"started_at,omitempty"`
	CompletedAt     *time.Time         `bson:"completed_at,omitempty"`
	CancelledAt     *time.Time         `bson:"cancelled_at,omitempty"`
	CreatedAt       time.Time          `bson:"created_at"`
	UpdatedAt       time.Time          `bson:"updated_at"`
}

type RideMongoRepository struct {
	collection *mongo.Collection
	db         *mongo.Database
}

// NewRideMongoRepository creates a new MongoDB ride repository
func NewRideMongoRepository(db *mongo.Database) *RideMongoRepository {
	collection := db.Collection("rides")

	pickupIndexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "pickup_location", Value: "2dsphere"}}, // Create geospatial index on pickup_location for finding nearby rides
	}

	dropoffIndexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "dropoff_location", Value: "2dsphere"}}, // Create geospatial index on dropoff_location
	}

	statusIndexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "status", Value: 1}}, // Create index on status for efficient filtering
	}

	customerIndexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "customer_id", Value: 1}}, // Create index on customer_id
	}

	driverIndexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "driver_id", Value: 1}}, // Create index on driver_id
	}

	compoundIndexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "status", Value: 1},
			{Key: "requested_at", Value: -1}, // Create compound index on status and requested_at for efficient polling
		},
	}

	rideIDIndexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "ride_id", Value: 1}},
		Options: options.Index().SetUnique(true), // Create unique index on ride_id for auto-increment simulation
	}

	// Create all indexes
	ctx := context.Background()
	collection.Indexes().CreateOne(ctx, pickupIndexModel)
	collection.Indexes().CreateOne(ctx, dropoffIndexModel)
	collection.Indexes().CreateOne(ctx, statusIndexModel)
	collection.Indexes().CreateOne(ctx, customerIndexModel)
	collection.Indexes().CreateOne(ctx, driverIndexModel)
	collection.Indexes().CreateOne(ctx, compoundIndexModel)
	collection.Indexes().CreateOne(ctx, rideIDIndexModel)

	return &RideMongoRepository{
		collection: collection,
		db:         db,
	}
}

// getNextRideID generates next sequence ID for ride_id
func (r *RideMongoRepository) getNextRideID(ctx context.Context) (int64, error) {
	counterCollection := r.db.Collection("counters")

	filter := bson.M{"_id": "ride_id"}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var result struct {
		Seq int64 `bson:"seq"`
	}

	err := counterCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&result)
	if err != nil {
		logger.Error(ctx, err)
		return 0, err
	}

	return result.Seq, nil
}

// toRideDocument converts domain.Ride to RideDocument
func toRideDocument(ride *domain.Ride) *RideDocument {
	now := time.Now()
	doc := &RideDocument{
		RideID:     ride.ID,
		CustomerID: ride.CustomerID,
		DriverID:   ride.DriverID,
		PickupLocation: GeoJSONPoint{
			Type:        "Point",
			Coordinates: []float64{ride.PickupLng, ride.PickupLat},
		},
		DropoffLocation: GeoJSONPoint{
			Type:        "Point",
			Coordinates: []float64{ride.DropoffLng, ride.DropoffLat},
		},
		PickupLat:   ride.PickupLat,
		PickupLng:   ride.PickupLng,
		DropoffLat:  ride.DropoffLat,
		DropoffLng:  ride.DropoffLng,
		Status:      string(ride.Status),
		Fare:        ride.Fare,
		RequestedAt: ride.RequestedAt,
		AcceptedAt:  ride.AcceptedAt,
		StartedAt:   ride.StartedAt,
		CompletedAt: ride.CompletedAt,
		CancelledAt: ride.CancelledAt,
		UpdatedAt:   now,
	}

	if doc.RideID == 0 {
		doc.CreatedAt = now
	}

	return doc
}

// toRideDomain converts RideDocument to domain.Ride
func toRideDomain(doc *RideDocument) *domain.Ride {
	return &domain.Ride{
		ID:          doc.RideID,
		CustomerID:  doc.CustomerID,
		DriverID:    doc.DriverID,
		PickupLat:   doc.PickupLat,
		PickupLng:   doc.PickupLng,
		DropoffLat:  doc.DropoffLat,
		DropoffLng:  doc.DropoffLng,
		Status:      domain.RideStatus(doc.Status),
		Fare:        doc.Fare,
		RequestedAt: doc.RequestedAt,
		AcceptedAt:  doc.AcceptedAt,
		StartedAt:   doc.StartedAt,
		CompletedAt: doc.CompletedAt,
		CancelledAt: doc.CancelledAt,
	}
}

// Create creates a new ride in MongoDB
func (r *RideMongoRepository) Create(ctx context.Context, ride *domain.Ride) error {
	rideID, err := r.getNextRideID(ctx)
	if err != nil {
		logger.Error(ctx, "Failed to generate ride ID", err)
		return err
	}

	ride.ID = rideID
	doc := toRideDocument(ride)

	_, err = r.collection.InsertOne(ctx, doc)
	if err != nil {
		logger.Error(ctx, "Failed to insert ride", err)
		return err
	}

	return nil
}

// GetByID retrieves a ride by its ID
func (r *RideMongoRepository) GetByID(ctx context.Context, id int64) (*domain.Ride, error) {
	var doc RideDocument

	filter := bson.M{"ride_id": id}
	err := r.collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrRideNotFound
		}
		logger.Error(ctx, "Failed to get ride by ID", err)
		return nil, err
	}

	return toRideDomain(&doc), nil
}

// Update updates an existing ride
func (r *RideMongoRepository) Update(ctx context.Context, ride *domain.Ride) error {
	doc := toRideDocument(ride)

	filter := bson.M{"ride_id": ride.ID}
	update := bson.M{
		"$set": bson.M{
			"driver_id":    doc.DriverID,
			"status":       doc.Status,
			"fare":         doc.Fare,
			"accepted_at":  doc.AcceptedAt,
			"started_at":   doc.StartedAt,
			"completed_at": doc.CompletedAt,
			"cancelled_at": doc.CancelledAt,
			"updated_at":   time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		logger.Error(ctx, "Failed to update ride", err)
		return err
	}

	if result.MatchedCount == 0 {
		return ErrRideNotFound
	}

	return nil
}

// GetRequestedRides retrieves all rides with "requested" status
func (r *RideMongoRepository) GetRequestedRides(ctx context.Context) ([]*domain.Ride, error) {
	filter := bson.M{"status": "requested"}
	opts := options.Find().SetSort(bson.D{{Key: "requested_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error(ctx, "Failed to get requested rides", err)
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

	return rides, nil
}

// GetNearbyRequestedRides retrieves rides within a certain radius using geospatial query
// This is the key method for driver polling - finds available rides near driver's location
// Filters: status in ["requested", "pending"], updated within last 5 minutes, within radius
// Params: lat, lng (driver location), maxDistanceMeters (search radius), limit (max results)
func (r *RideMongoRepository) GetNearbyRequestedRides(ctx context.Context, lat, lng, maxDistanceMeters float64, limit int) ([]*domain.Ride, error) {

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

	return rides, nil
}

// GetByCustomerID retrieves all rides for a customer
func (r *RideMongoRepository) GetByCustomerID(ctx context.Context, customerID int64) ([]*domain.Ride, error) {
	filter := bson.M{"customer_id": customerID}
	opts := options.Find().SetSort(bson.D{{Key: "requested_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error(ctx, "Failed to get rides by customer ID", err)
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

	return rides, nil
}

// GetByDriverID retrieves all rides for a driver
func (r *RideMongoRepository) GetByDriverID(ctx context.Context, driverID int64) ([]*domain.Ride, error) {
	filter := bson.M{"driver_id": driverID}
	opts := options.Find().SetSort(bson.D{{Key: "requested_at", Value: -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error(ctx, "Failed to get rides by driver ID", err)
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

	return rides, nil
}
