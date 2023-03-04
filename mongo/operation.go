package mongo

import (
	"context"
	"errors"
	"fmt"
	"gomongo/database/connection"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrConnectionNotInitialized = errors.New("connection was not initialized")

func Create(collectionName string, object interface{}) error {
	collection, err := getCollection(collectionName)
	if err != nil {
		return err
	}

	objectBSON, err := dataToBSON(object)
	if err != nil {
		return err
	}

	_, err = collection.InsertOne(context.TODO(), objectBSON)
	if err != nil {
		return fmt.Errorf("mongo #create: %w", err)
	}

	return nil
}

func All(collectionName string) (bson.A, error) {
	collection, err := getCollection(collectionName)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, fmt.Errorf("db all: %w", err)
	}

	var allBSON bson.A
	for cursor.Next(context.TODO()) {
		var instanceBSON bson.M
		err = bson.Unmarshal(cursor.Current, &instanceBSON)
		if err != nil {
			return nil, fmt.Errorf("db all: %w", err)
		}
		allBSON = append(allBSON, instanceBSON)
	}

	return allBSON, nil
}

func getCollection(collectionName string) (*mongo.Collection, error) {
	if connection.MongoInstace.Database == nil {
		return nil, ErrConnectionNotInitialized
	}
	return connection.MongoInstace.Database.Collection(collectionName), nil
}

func dataToBSON(data interface{}) (bson.M, error) {
	dataMarshal, err := bson.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("convert data: %w", err)
	}

	var dataBSON bson.M
	if err := bson.Unmarshal(dataMarshal, &dataBSON); err != nil {
		return nil, fmt.Errorf("convert data: %w", err)
	}

	return dataBSON, nil
}
