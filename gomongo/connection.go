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
	ErrGomongoCanNotConnect = errors.New("gomongo can not connect to mongodb")
)

var mongoDatabase *mongo.Database

type connectionSettings struct {
	uri          string
	databaseName string
	timeout      time.Duration
	pingTimeout  time.Duration
}

func ConnectionSettings() *connectionSettings {
	return &connectionSettings{
		pingTimeout: 10 * time.Second,
	}
}

func (cs *connectionSettings) SetURI(uri string) *connectionSettings {
	cs.uri = uri
	return cs
}

func (cs *connectionSettings) SetDatabaseName(databaseName string) *connectionSettings {
	cs.databaseName = databaseName
	return cs
}

func (cs *connectionSettings) SetTimeout(timeout time.Duration) *connectionSettings {
	cs.timeout = timeout
	return cs
}

func (cs *connectionSettings) SetPingTimeout(pingTimeout time.Duration) *connectionSettings {
	cs.pingTimeout = pingTimeout
	return cs
}

func (cs *connectionSettings) validate() error {
	if cs.uri == "" {
		return fmt.Errorf("%w: URI is invalid", ErrInvalidSettings)
	}

	if cs.databaseName == "" {
		return fmt.Errorf("%w: Database Name is invalid", ErrInvalidSettings)
	}

	if cs.pingTimeout <= 0 {
		return fmt.Errorf("%w: Ping Timeout is invalid", ErrInvalidSettings)
	}

	return nil
}

func Init(cs *connectionSettings) error {
	if err := cs.validate(); err != nil {
		return err
	}

	mongoClientOptions := options.Client().
		ApplyURI(cs.uri).
		SetConnectTimeout(cs.timeout)

	mongoClient, err := mongo.Connect(context.Background(), mongoClientOptions)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGomongoCanNotConnect, err)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), cs.pingTimeout)
	defer cancelFunc()

	if err := mongoClient.Ping(ctx, nil); err != nil {
		return fmt.Errorf("%w: %w", ErrGomongoCanNotConnect, err)
	}

	mongoDatabase = mongoClient.Database(cs.databaseName)

	return nil
}
