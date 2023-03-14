package mongo

import (
	"context"
	"errors"
	"fmt"
	"gomongo/database/connection"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrConnectionNotInitialized = errors.New("connection was not initialized")
var ErrEmptyCollection = errors.New("collection empty")

func All(collectionName string) ([]interface{}, error) {
	collection, err := getCollection(collectionName)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, fmt.Errorf("mongo all: %w", err)
	}

	var all []interface{}
	for cursor.Next(context.TODO()) {
		var instance interface{}
		err = bson.Unmarshal(cursor.Current, &instance)
		if err != nil {
			return nil, fmt.Errorf("mongo all: %w", err)
		}
		all = append(all, instance)
	}

	return all, nil
}

func Create(collectionName string, object interface{}) (primitive.ObjectID, error) {
	var id primitive.ObjectID
	collection, err := getCollection(collectionName)
	if err != nil {
		return id, err
	}

	result, err := collection.InsertOne(context.TODO(), object)
	if err != nil {
		err = fmt.Errorf("mongo #create: %w", err)
		return id, err
	}

	id = result.InsertedID.(primitive.ObjectID)

	return id, nil
}

func First(collectionName string) (interface{}, error) {
	collection, err := getCollection(collectionName)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, fmt.Errorf("mongo first: %w", err)
	}

	var instance interface{}
	if cursor.Next(context.TODO()) {
		err = bson.Unmarshal(cursor.Current, &instance)
		if err != nil {
			return nil, fmt.Errorf("mongo first: %w", err)
		}
	} else {
		return nil, fmt.Errorf("mongo first: %w", ErrEmptyCollection)
	}

	return instance, nil
}

func getCollection(collectionName string) (*mongo.Collection, error) {
	if connection.MongoInstace.Database == nil {
		return nil, ErrConnectionNotInitialized
	}
	return connection.MongoInstace.Database.Collection(collectionName), nil
}
