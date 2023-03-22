package mongov2

import (
	"context"
	"errors"
	"fmt"
	"gomongo/database/connection"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrConnectionNotInitialized = errors.New("connection was not initialized")
var ErrDocumentNotFound = errors.New("document not found")

type CollectionI interface {
	FirstV2(i interface{}) error
}

type collection struct {
	collection *mongo.Collection
}

func NewCollection(collectionName string) CollectionI {
	c := connection.MongoInstace.Database.Collection(collectionName)
	return collection{collection: c}
}

func (c collection) FirstV2(i interface{}) error {
	cursor, err := c.collection.Find(context.TODO(), bson.D{{}})
	if err != nil {
		return fmt.Errorf("mongo first: %w", err)
	}

	var instanceMap map[string]interface{}
	if cursor.Next(context.TODO()) {
		err = bson.Unmarshal(cursor.Current, &instanceMap)
		if err != nil {
			return fmt.Errorf("mongo first: %w", err)
		}
	} else {
		return fmt.Errorf("mongo first: %w", ErrDocumentNotFound)
	}

	convertMapToStruct(instanceMap, i)

	return nil
}

func convertMapToStruct(m map[string]interface{}, s interface{}) {
	stValue := reflect.ValueOf(s).Elem()
	sType := stValue.Type()
	for i := 0; i < sType.NumField(); i++ {
		field := sType.Field(i)
		if value, ok := m[field.Name]; ok {
			if stValue.Field(i).Type().String() == "int" {
				stValue.Field(i).Set(reflect.ValueOf(int(value.(int32))))
			} else {
				stValue.Field(i).Set(reflect.ValueOf(value))
			}
		}
	}
}
