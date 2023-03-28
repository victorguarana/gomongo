package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrInvaidURI = errors.New("URI must be valid")
var ErrInvaidDatabaseName = errors.New("DatabaseName must be valid")
var ErrCouldNotConnect = errors.New("Go mongo could not connect to URI")

var mongoDatabase *mongo.Database

func Init(uri, databaseName string) error {
	if uri == "" {
		return ErrInvaidURI
	}

	if databaseName == "" {
		return ErrInvaidDatabaseName
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return ErrCouldNotConnect
	}
	mongoDatabase = client.Database(databaseName)

	return nil
}
