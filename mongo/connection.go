package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ErrInvaidURI = errors.New("URI must be valid")

var mongoDB *mongo.Database

func Init(uri, databaseName string) error {
	if uri == "" {
		return ErrInvaidURI
	}

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		return err
	}

	mongoDB = client.Database(databaseName)

	return nil
}
