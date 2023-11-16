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
	ErrInvalidSettings      = errors.New("settings must be valid")
	ErrGomongoCanNotConnect = errors.New("gomongo can not connect to mongo")
)

var mongoDatabase *mongo.Database

type ConnectionSettings struct {
	URI          string
	DatabaseName string
	Timeout      time.Duration
}

func (cs *ConnectionSettings) validate() error {
	if cs.URI == "" {
		return fmt.Errorf("%w: URI is invalid", ErrInvalidSettings)
	}

	if cs.DatabaseName == "" {
		return fmt.Errorf("%w: DatabaseName is invalid", ErrInvalidSettings)
	}

	if cs.Timeout <= 0 {
		return fmt.Errorf("%w: Timeout is invalid", ErrInvalidSettings)
	}

	return nil
}

func Init(cs ConnectionSettings) error {
	if err := cs.validate(); err != nil {
		return err
	}

	mongoClientOptions := options.Client().
		ApplyURI(cs.URI).
		SetConnectTimeout(cs.Timeout)

	mongoClient, err := mongo.Connect(context.Background(), mongoClientOptions)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGomongoCanNotConnect, err)
	}

	if err := mongoClient.Ping(context.Background(), nil); err != nil {
		return fmt.Errorf("%w: %w", ErrGomongoCanNotConnect, err)
	}

	mongoDatabase = mongoClient.Database(cs.DatabaseName)

	return nil
}
