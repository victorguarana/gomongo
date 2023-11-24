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

type ConnectionSettings struct {
	URI          string
	DatabaseName string
	Timeout      time.Duration
	PingTimeout  time.Duration
}

func NewConnectionSettings() *ConnectionSettings {
	return &ConnectionSettings{
		PingTimeout: 10 * time.Second,
	}
}

func (cs *ConnectionSettings) SetURI(uri string) *ConnectionSettings {
	cs.URI = uri
	return cs
}

func (cs *ConnectionSettings) SetDatabaseName(databaseName string) *ConnectionSettings {
	cs.DatabaseName = databaseName
	return cs
}

func (cs *ConnectionSettings) SetTimeout(timeout time.Duration) *ConnectionSettings {
	cs.Timeout = timeout
	return cs
}

func (cs *ConnectionSettings) SetPingTimeout(pingTimeout time.Duration) *ConnectionSettings {
	cs.PingTimeout = pingTimeout
	return cs
}

func (cs *ConnectionSettings) validate() error {
	if cs.URI == "" {
		return fmt.Errorf("%w: URI is invalid", ErrInvalidSettings)
	}

	if cs.DatabaseName == "" {
		return fmt.Errorf("%w: Database Name is invalid", ErrInvalidSettings)
	}

	if cs.PingTimeout <= 0 {
		return fmt.Errorf("%w: Ping Timeout is invalid", ErrInvalidSettings)
	}

	return nil
}

func Init(cs *ConnectionSettings) error {
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

	ctx, cancelFunc := context.WithTimeout(context.Background(), cs.PingTimeout)
	defer cancelFunc()

	if err := mongoClient.Ping(ctx, nil); err != nil {
		return fmt.Errorf("%w: %w", ErrGomongoCanNotConnect, err)
	}

	mongoDatabase = mongoClient.Database(cs.DatabaseName)

	return nil
}
