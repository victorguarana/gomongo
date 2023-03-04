package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func FirstV2(collectionName string, i interface{}) error {
	collection, err := getCollection(collectionName)
	if err != nil {
		return err
	}

	cursor, err := collection.Find(context.TODO(), bson.D{{}})
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
		return fmt.Errorf("mongo first: %w", ErrEmptyCollection)
	}

	convertMapToStruct(instanceMap, i)

	return nil
}
