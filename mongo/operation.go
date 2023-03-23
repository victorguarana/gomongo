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

type Collection struct {
	Name            string
	collectionMongo *mongo.Collection
}

func NewCollection(collectionName string) Collection {
	return Collection{Name: collectionName}
}

func (c *Collection) validate() error {
	if c.collectionMongo != nil {
		return nil
	}

	if mongoDatabase == nil {
		return ErrConnectionNotInitialized
	} else {
		c.collectionMongo = mongoDatabase.Collection(c.Name)
	}
	return nil
}

func (c *Collection) All() ([]interface{}, error) {
	return c.Where(bson.M{})
}

func (c *Collection) Create(object interface{}) (string, error) {
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

	result, err := c.collectionMongo.InsertOne(context.TODO(), bson)
	if err != nil {
		return id, err
	}

	id = result.InsertedID.(primitive.ObjectID).Hex()

	return id, nil
}

func (c *Collection) Count() (int, error) {
	err := c.validate()
	if err != nil {
		return 0, err
	}

	filter := bson.M{}
	count, err := c.collectionMongo.CountDocuments(context.TODO(), filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (c *Collection) DeleteID(id string) error {
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
	result, err := c.collectionMongo.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return ErrNothingDeleted
	}

	return nil
}

func (c *Collection) FindOne(filter interface{}) (interface{}, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}

	result := c.collectionMongo.FindOne(context.TODO(), filter)
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

func (c *Collection) First() (interface{}, error) {
	return c.FindOne(map[string]string{})
}

func (c *Collection) UpdateID(id string, object interface{}) error {
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

	result, err := c.collectionMongo.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrDocumentNotFound
	}

	return nil
}

func (c *Collection) Where(filter interface{}) ([]interface{}, error) {
	err := c.validate()
	if err != nil {
		return nil, err
	}

	cursor, err := c.collectionMongo.Find(context.TODO(), filter)
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
