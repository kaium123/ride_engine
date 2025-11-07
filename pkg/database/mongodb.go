package database

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/event"
	"time"
	"vcs.technonext.com/carrybee/ride_engine/pkg/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"vcs.technonext.com/carrybee/ride_engine/pkg/config"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoDB(cfg config.MongoDBConfig) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	commandMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			fmt.Printf("MongoDB Query: %s\n", evt.Command.String())
		},
	}

	clientOptions := options.Client().ApplyURI(cfg.URI)
	clientOptions.SetMaxPoolSize(50)
	clientOptions.SetMinPoolSize(10)
	clientOptions.SetMonitor(commandMonitor)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		logger.Error(ctx, err)
		return nil, err
	}

	database := client.Database(cfg.Database)

	logger.Info(ctx, "MongoDB connected successfully")
	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

func (m *MongoDB) Close() error {
	logger.Info(context.Background(), "Closing MongoDB connection...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.Client.Disconnect(ctx)
}

func (m *MongoDB) HealthCheck(ctx context.Context) error {
	return m.Client.Ping(ctx, nil)
}

func (m *MongoDB) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}
