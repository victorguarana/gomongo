package gomongo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrGomongoCanNotConnect = errors.New("gomongo can not connect to mongodb")
)

type Database struct {
	mongoDatabse *mongo.Database
}

func NewDatabase(cs ConnectionSettings) (*Database, error) {
	if err := cs.validate(); err != nil {
		return nil, err
	}

	mongoClient, err := mongoClient(&cs, context.Background())
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGomongoCanNotConnect, err)
	}

	if err := pingMongoServer(&cs, mongoClient, context.Background()); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrGomongoCanNotConnect, err)
	}

	return &Database{
		mongoClient.Database(cs.DatabaseName),
	}, nil
}

func mongoClient(cs *ConnectionSettings, ctx context.Context) (*mongo.Client, error) {
	return mongo.Connect(ctx, clientOptions(cs))
}

func clientOptions(cs *ConnectionSettings) *options.ClientOptions {
	clientOptions := options.Client().ApplyURI(cs.URI)
	if cs.ConnectionTimeout > 0 {
		clientOptions.SetConnectTimeout(cs.ConnectionTimeout)
	}

	return clientOptions
}

func pingMongoServer(cs *ConnectionSettings, mongoClient *mongo.Client, ctx context.Context) error {
	pingTimeout := cs.ConnectionTimeout
	if pingTimeout <= 0 {
		pingTimeout = 30 * time.Second
	}

	ctx, cancelFunc := context.WithTimeout(ctx, pingTimeout)
	defer cancelFunc()

	return mongoClient.Ping(ctx, nil)
}
