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
var ErrEmptyCollection = errors.New("collection empty")

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

func First(collectionName string, i interface{}) error {
	collection, err := getCollection(collectionName)
	if err != nil {
		return err
	}

	cursor, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return fmt.Errorf("db first: %w", err)
	}

	var instanceMap map[string]interface{}
	if cursor.Next(context.TODO()) {
		err = bson.Unmarshal(cursor.Current, &instanceMap)
		if err != nil {
			return fmt.Errorf("db first: %w", err)
		}
	} else {
		return fmt.Errorf("db first: %w", ErrEmptyCollection)
	}

	convertMapToStruct(instanceMap, i)

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
