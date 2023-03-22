package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrConnectionNotInitialized = errors.New("connection was not initialized")
var ErrEmptyCollection = errors.New("collection empty")
var ErrNothingDeleted = errors.New("nothing was deleted")
var ErrIDNotExist = errors.New("id must exist")
var ErrDocumentNotFound = errors.New("document not found")

func All(collectionName string) ([]interface{}, error) {
	return Where(collectionName, bson.M{})
}

func Create(collectionName string, object interface{}) (string, error) {
	var id string
	collection, err := getCollection(collectionName)
	if err != nil {
		return id, err
	}

	bson, err := dataToBSON(object)
	if err != nil {
		return id, err
	}

	delete(bson, "id")

	result, err := collection.InsertOne(context.TODO(), bson)
	if err != nil {
		return id, err
	}

	id = result.InsertedID.(primitive.ObjectID).Hex()

	return id, nil
}

func Count(collectionName string) (int, error) {
	collection, err := getCollection(collectionName)
	if err != nil {
		return 0, err
	}

	filter := bson.M{}
	count, err := collection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func DeleteID(collectionName string, id string) error {
	if id == "" {
		return ErrIDNotExist
	}

	collection, err := getCollection(collectionName)
	if err != nil {
		return err
	}

	idPrimitive, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": idPrimitive}
	result, err := collection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrNothingDeleted
	}

	return nil
}

func FindOne(collectionName string, filter interface{}) (interface{}, error) {
	collection, err := getCollection(collectionName)
	if err != nil {
		return nil, err
	}

	result := collection.FindOne(context.TODO(), filter)
	if result.Err() != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) {
			return nil, ErrDocumentNotFound
		}
		return nil, result.Err()
	}

	var instance interface{}
	err = result.Decode(&instance)
	if err != nil {
		return nil, err
	}

	return instance, nil
}

func First(collectionName string) (interface{}, error) {
	return FindOne(collectionName, map[string]string{})
}

func UpdateID(collectionName string, id string, object interface{}) error {
	if id == "" {
		return ErrIDNotExist
	}

	collection, err := getCollection(collectionName)
	if err != nil {
		return err
	}

	idPrimitive, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	objectBSON, err := dataToBSON(object)
	if err != nil {
		return err
	}
	delete(objectBSON, "id")

	filter := bson.M{"_id": idPrimitive}
	update := bson.M{
		"$set": objectBSON,
	}

	result, err := collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}

func Where(collectionName string, filter interface{}) ([]interface{}, error) {
	collection, err := getCollection(collectionName)
	if err != nil {
		return nil, err
	}

	cursor, err := collection.Find(context.TODO(), filter)
	if err != nil {
		return nil, err
	}

	var all []interface{}
	for cursor.Next(context.TODO()) {
		var instance interface{}
		err = bson.Unmarshal(cursor.Current, &instance)
		if err != nil {
			return nil, err
		}
		all = append(all, instance)
	}

	return all, nil
}

func getCollection(collectionName string) (*mongo.Collection, error) {
	if mongoDB == nil {
		return nil, ErrConnectionNotInitialized
	}
	return mongoDB.Collection(collectionName), nil
}
