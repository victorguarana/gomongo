package mongo

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"gomongo/database/connection"

// 	"go.mongodb.org/mongo-driver/bson"
// 	"go.mongodb.org/mongo-driver/mongo"
// )

// var ErrDatabaseNotInitialized = errors.New("database was not initialized")

// // var ErrEmptyCollection = errors.New("collection empty")
// // var ErrNothingDeleted = errors.New("on delete")

// func Create(collectionName string, object bson.M) error {
// 	collection, err := getCollection(collectionName)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = collection.InsertOne(context.TODO(), object)
// 	if err != nil {
// 		return fmt.Errorf("mongo #create: %w", err)
// 	}

// 	return nil
// }

// func getCollection(collectionName string) (*mongo.Collection, error) {
// 	if connection.MongoInstace.Database == nil {
// 		return nil, ErrDatabaseNotInitialized
// 	}
// 	return connection.MongoInstace.Database.Collection(collectionName), nil
// }

// // type DB struct {
// // 	database *mongo.Database
// // }

// // func (db DB) First(collectionName string) (bson.M, error) {
// // 	collection := db.database.Collection(collectionName)

// // 	cursor, err := collection.Find(context.TODO(), bson.D{{}})
// // 	if err != nil {
// // 		return nil, fmt.Errorf("db first: %w", err)
// // 	}

// // 	var instanceBSON bson.M
// // 	if cursor.Next(context.TODO()) {
// // 		err = bson.Unmarshal(cursor.Current, &instanceBSON)
// // 		if err != nil {
// // 			return nil, fmt.Errorf("db first: %w", err)
// // 		}
// // 	} else {
// // 		return nil, fmt.Errorf("db first: %w", ErrEmptyCollection)
// // 	}

// // 	return instanceBSON, nil
// // }

// // func (db DB) Delete(collectionName string, object bson.M) error {
// // 	collection := db.database.Collection(collectionName)

// // 	result, err := collection.DeleteOne(context.TODO(), object)
// // 	if err != nil {
// // 		return fmt.Errorf("db delete: %w", err)
// // 	}

// // 	if result.DeletedCount == 0 {
// // 		return fmt.Errorf("db delete: %w", ErrNothingDeleted)
// // 	}

// // 	return nil
// // }

// // func (db DB) All(collectionName string) (bson.A, error) {
// // 	collection := db.database.Collection(collectionName)

// // 	cursor, err := collection.Find(context.TODO(), bson.D{{}})
// // 	if err != nil {
// // 		return nil, fmt.Errorf("db all: %w", err)
// // 	}

// // 	var allBSON bson.A
// // 	for cursor.Next(context.TODO()) {
// // 		var instanceBSON bson.M
// // 		err = bson.Unmarshal(cursor.Current, &instanceBSON)
// // 		if err != nil {
// // 			return nil, fmt.Errorf("db all: %w", err)
// // 		}
// // 		allBSON = append(allBSON, instanceBSON)
// // 	}

// // 	return allBSON, nil
// // }
