package connection

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var errInvaidURI = errors.New("URI must be valid")

type mongoInstace struct {
	Database *mongo.Database
}

// TODO: Private this instance
var MongoInstace mongoInstace

func Init(uri, databaseName string) error {
	if uri == "" {
		return errInvaidURI
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	MongoInstace = mongoInstace{
		Database: client.Database(databaseName),
	}

	return nil
}
