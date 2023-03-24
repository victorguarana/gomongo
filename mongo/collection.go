package mongo

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrConnectionNotInitialized = errors.New("connection was not initialized")
var ErrIDNotExist = errors.New("id must exist")
var ErrDocumentNotFound = errors.New("document not found")

type Collection interface {
	All() ([]interface{}, error)
	Create(interface{}) (string, error)
	Count() (int, error)
	DeleteID(string) error
	FindOne(interface{}) (interface{}, error)
	First() (interface{}, error)
	UpdateID(string, interface{}) error
	Where(interface{}) ([]interface{}, error)
}

type collection struct {
	name            string
	mongoCollection *mongo.Collection
}

func NewCollection(collectionName string) Collection {
	return collection{name: collectionName}
}

func (c *collection) validate() error {
	if c.mongoCollection != nil {
		return nil
	}

	if mongoDatabase == nil {
		return ErrConnectionNotInitialized
	}

	c.mongoCollection = mongoDatabase.Collection(c.name)

	return nil
}

func (c collection) All() ([]interface{}, error) {
	return c.Where(bson.M{})
}

func (c collection) Create(object interface{}) (string, error) {
	var id string

	err := c.validate()
	if err != nil {
		return id, err
	}

	bson, err := dataToBSON(object)
	if err != nil {
		return id, err
	}

	delete(bson, "id")

	result, err := c.mongoCollection.InsertOne(context.TODO(), bson)
	if err != nil {
		return id, err
	}

	id = result.InsertedID.(primitive.ObjectID).Hex()

	return id, nil
}

func (c collection) Count() (int, error) {
	err := c.validate()
	if err != nil {
		return 0, err
	}

	filter := bson.M{}
	count, err := c.mongoCollection.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (c collection) DeleteID(id string) error {
	if id == "" {
		return ErrIDNotExist
	}

	err := c.validate()
	if err != nil {
		return err
	}

	idPrimitive, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"_id": idPrimitive}
	result, err := c.mongoCollection.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}

func (c collection) FindOne(filter interface{}) (interface{}, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}

	result := c.mongoCollection.FindOne(context.TODO(), filter)
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

func (c collection) First() (interface{}, error) {
	return c.FindOne(map[string]string{})
}

func (c collection) UpdateID(id string, object interface{}) error {
	if id == "" {
		return ErrIDNotExist
	}

	err := c.validate()
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

	result, err := c.mongoCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}

func (c collection) Where(filter interface{}) ([]interface{}, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}

	cursor, err := c.mongoCollection.Find(context.TODO(), filter)
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
