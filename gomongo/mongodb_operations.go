package gomongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func where[T any](ctx context.Context, mongoCollection *mongo.Collection, filter any, order map[string]OrderBy) ([]T, error) {
	cursor, err := mongoCollection.Find(ctx, filter, options.Find().SetSort(order))
	if err != nil {
		return nil, err
	}

	return mongoCursorToSlice[T](ctx, cursor)
}

func mongoCursorToSlice[T any](ctx context.Context, cursor *mongo.Cursor) ([]T, error) {
	var instanceSlice = []T{}

	for cursor.Next(ctx) {
		var instance T
		err := cursor.Decode(&instance)
		if err != nil {
			return nil, err
		}

		instanceSlice = append(instanceSlice, instance)
	}

	return instanceSlice, nil
}

func findOne[T any](ctx context.Context, mongoCollection *mongo.Collection, filter any, order map[string]OrderBy) (T, error) {
	var instance T
	result := mongoCollection.FindOne(ctx, filter, options.FindOne().SetSort(order))
	if err := singleResultError(result); err != nil {
		return instance, err
	}

	return singleResultToInstance[T](result)
}

func singleResultError(result *mongo.SingleResult) error {
	if err := result.Err(); err != nil {
		if errors.Is(result.Err(), mongo.ErrNoDocuments) || errors.Is(result.Err(), mongo.ErrNilDocument) {
			return ErrDocumentNotFound
		}
		return err
	}
	return nil
}

func singleResultToInstance[T any](result *mongo.SingleResult) (T, error) {
	var instance T
	err := result.Decode(&instance)

	return instance, err
}

func create[T any](ctx context.Context, mongoCollection *mongo.Collection, doc T) (ID, error) {
	docBSON, err := dataToBSON(doc)
	if err != nil {
		return nil, err
	}

	delete(docBSON, "_id")
	result, err := mongoCollection.InsertOne(ctx, docBSON)
	if err != nil {
		return nil, insertOneError(err)
	}

	return insertOneResultToID(result)
}

func dataToBSON[T any](doc T) (bson.M, error) {
	dataMarshal, err := bson.Marshal(doc)
	if err != nil {
		return nil, fmt.Errorf("convert data: %w", err)
	}

	var dataBSON bson.M
	if err := bson.Unmarshal(dataMarshal, &dataBSON); err != nil {
		return nil, fmt.Errorf("convert data: %w", err)
	}

	return dataBSON, nil
}

func insertOneError(err error) error {
	if mongo.IsDuplicateKeyError(err) {
		return ErrDuplicateKey
	}

	return err
}

func insertOneResultToID(result *mongo.InsertOneResult) (ID, error) {
	id, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, fmt.Errorf("cannot convert id to ObjectID")
	}

	return &id, nil
}

func deleteID(ctx context.Context, mongoCollection *mongo.Collection, filter any) error {
	result, err := mongoCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if err := deleteResultError(result); err != nil {
		return err
	}

	return nil
}

func deleteResultError(result *mongo.DeleteResult) error {
	if result.DeletedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}

func updateID[T any](ctx context.Context, mongoCollection *mongo.Collection, filter any, doc T) error {
	docBSON, err := dataToBSON(doc)
	if err != nil {
		return err
	}

	delete(docBSON, "_id")

	update := bson.M{"$set": docBSON}
	result, err := mongoCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if err := updateResultErrors(result); err != nil {
		return err
	}

	return nil
}

func updateResultErrors(result *mongo.UpdateResult) error {
	if result.MatchedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}

func count(ctx context.Context, mongoCollection *mongo.Collection, filter any) (int, error) {
	count, err := mongoCollection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func createUniqueIndex(ctx context.Context, mongoCollection *mongo.Collection, name string, keys map[string]OrderBy) error {
	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(true),
	}

	if name != "" {
		indexModel.Options.SetName(name)
	}

	_, err := mongoCollection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return err
	}

	return nil
}

func listIndexes(ctx context.Context, mongoCollection *mongo.Collection) ([]Index, error) {
	cursor, err := mongoCollection.Indexes().List(ctx)
	if err != nil {
		// mongo.CommandError
		return nil, err
	}

	return mongoCursorToSliceIndex(ctx, cursor)
}

func mongoCursorToSliceIndex(ctx context.Context, cursor *mongo.Cursor) ([]Index, error) {
	var indexes []Index

	for cursor.Next(ctx) {
		var index Index
		err := cursor.Decode(&index)
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, index)
	}

	return indexes, nil
}

func deleteIndex(ctx context.Context, mongoCollection *mongo.Collection, indexName string) error {
	_, err := mongoCollection.Indexes().DropOne(ctx, indexName)
	if err != nil {
		var mongoCommandError mongo.CommandError
		if ok := errors.As(err, &mongoCommandError); ok {
			return mongoCommandErrorToCustomError(mongoCommandError)
		}
		return err
	}

	return nil
}

func mongoCommandErrorToCustomError(mongoCommandError mongo.CommandError) error {
	switch mongoCommandError.Code {
	case 72:
		return fmt.Errorf("%w: %s", ErrInvalidCommandOptions, fmt.Errorf(mongoCommandError.Message))
	case 27:
		return fmt.Errorf("%w: %s", ErrIndexNotFound, fmt.Errorf(mongoCommandError.Message))
	}

	return fmt.Errorf("mongo command error: %s: %s", mongoCommandError.Name, mongoCommandError.Message)
}

func drop(ctx context.Context, mongoCollection *mongo.Collection) error {
	return mongoCollection.Drop(ctx)
}
