package gomongo

import (
	"context"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrEmptyID                  = errors.New("id can not be nil")
	ErrConnectionNotInitialized = errors.New("connection was not initialized")
	ErrDocumentNotFound         = errors.New("document not found")
	ErrDuplicateKey             = errors.New("duplicate key")
	ErrInvalidIndex             = errors.New("invalid index")
	ErrInvalidCommandOptions    = errors.New("invalid command options")
	ErrIndexNotFound            = errors.New("index not found")
	ErrInvalidOrder             = errors.New("invalid order parameter")
)

type ID *primitive.ObjectID

type OrderBy int

const (
	OrderAsc  OrderBy = 1
	OrderDesc OrderBy = -1
)

type Index struct {
	Keys map[string]OrderBy `bson:"key"`
	Name string
}

type Collection[T any] interface {
	All(ctx context.Context) ([]T, error)
	Create(ctx context.Context, doc T) (ID, error)
	Count(ctx context.Context) (int, error)
	DeleteID(ctx context.Context, id ID) error
	FindID(ctx context.Context, id ID) (T, error)
	FindOne(ctx context.Context, filter any) (T, error)
	First(ctx context.Context) (T, error)
	FirstInserted(ctx context.Context, filter any) (T, error)
	Last(ctx context.Context) (T, error)
	LastInserted(ctx context.Context, filter any) (T, error)
	UpdateID(ctx context.Context, id ID, doc T) error
	Where(ctx context.Context, filter any) ([]T, error)
	WhereWithOrder(ctx context.Context, filter any, orderBy map[string]OrderBy) ([]T, error)

	CreateUniqueIndex(ctx context.Context, index Index) error
	DeleteIndex(ctx context.Context, indexName string) error
	ListIndexes(ctx context.Context) ([]Index, error)

	Drop(ctx context.Context) error

	Name() string
}

type collection[T any] struct {
	mongoCollection *mongo.Collection
}

func NewCollection[T any](database *Database, collectionName string) (Collection[T], error) {
	if err := validateDatabase(database); err != nil {
		return nil, ErrConnectionNotInitialized
	}

	return &collection[T]{
		mongoCollection: database.mongoDatabase.Collection(collectionName),
	}, nil
}

// All returns all objects of a collection
func (c *collection[T]) All(ctx context.Context) ([]T, error) {
	emptyFilter := bson.M{}
	emptyOrder := map[string]OrderBy{}
	return where[T](ctx, c.mongoCollection, emptyFilter, emptyOrder)
}

// Count returns the number of objects of a collection
func (c *collection[T]) Count(ctx context.Context) (int, error) {
	emptyFilter := bson.M{}
	return count(ctx, c.mongoCollection, emptyFilter)
}

// Create inserts a new object into a collection and returns the id of the inserted document
func (c *collection[T]) Create(ctx context.Context, instance T) (ID, error) {
	return create(ctx, c.mongoCollection, instance)
}

// DeleteID deletes an object of a collection by id
func (c *collection[T]) DeleteID(ctx context.Context, id ID) error {
	if err := validateReceivedID(id); err != nil {
		return err
	}

	filter := bson.M{"_id": id}
	return deleteID(ctx, c.mongoCollection, filter)
}

// FindID returns an object of a collection by id
func (c *collection[T]) FindID(ctx context.Context, id ID) (T, error) {
	if err := validateReceivedID(id); err != nil {
		var t T
		return t, err
	}

	filter := bson.M{"_id": id}
	emptyOrder := map[string]OrderBy{}
	return findOne[T](ctx, c.mongoCollection, filter, emptyOrder)
}

// FindOne returns an object of a collection by filter
func (c *collection[T]) FindOne(ctx context.Context, filter any) (T, error) {
	filter = validateReceivedFilter(filter)
	emptyOrder := map[string]OrderBy{}
	return findOne[T](ctx, c.mongoCollection, filter, emptyOrder)
}

// First returns the first object of a collection in natural order
func (c *collection[T]) First(ctx context.Context) (T, error) {
	emptyFilter := bson.M{}
	emptyOrder := map[string]OrderBy{}
	return findOne[T](ctx, c.mongoCollection, emptyFilter, emptyOrder)
}

// FirstInserted returns the first object of a collection ordered by id
func (c *collection[T]) FirstInserted(ctx context.Context, filter any) (T, error) {
	filter = validateReceivedFilter(filter)
	order := map[string]OrderBy{"_id": OrderAsc}
	return findOne[T](ctx, c.mongoCollection, filter, order)
}

// Last returns the last object of a collection in natural order
func (c *collection[T]) Last(ctx context.Context) (T, error) {
	emptyFilter := bson.M{}
	order := map[string]OrderBy{"$natural": OrderDesc}
	return findOne[T](ctx, c.mongoCollection, emptyFilter, order)
}

// LastInserted returns the last object of a collection ordered by id
func (c *collection[T]) LastInserted(ctx context.Context, filter any) (T, error) {
	filter = validateReceivedFilter(filter)
	order := map[string]OrderBy{"_id": OrderDesc}
	return findOne[T](ctx, c.mongoCollection, filter, order)
}

// Update updates an object of a collection by id
func (c *collection[T]) UpdateID(ctx context.Context, id ID, instance T) error {
	if err := validateReceivedID(id); err != nil {
		return err
	}

	filter := bson.M{"_id": id}
	return updateID(ctx, c.mongoCollection, filter, instance)
}

// Where returns all objects of a collection by filter
func (c *collection[T]) Where(ctx context.Context, filter any) ([]T, error) {
	filter = validateReceivedFilter(filter)
	emptyOrder := map[string]OrderBy{}
	return where[T](ctx, c.mongoCollection, filter, emptyOrder)
}

// WhereWithOrder returns all objects of a collection by filter and order
func (c *collection[T]) WhereWithOrder(ctx context.Context, filter any, order map[string]OrderBy) ([]T, error) {
	filter = validateReceivedFilter(filter)
	order, err := validateReceivedOrder(order)
	if err != nil {
		return nil, err
	}
	return where[T](ctx, c.mongoCollection, filter, order)
}

func (c *collection[T]) CreateUniqueIndex(ctx context.Context, index Index) error {
	if err := validateReceivedIndex(index); err != nil {
		return err
	}

	return createUniqueIndex(ctx, c.mongoCollection, index.Name, index.Keys)
}

// ListIndexes returns all indexes of a collection
func (c *collection[T]) ListIndexes(ctx context.Context) ([]Index, error) {
	return listIndexes(ctx, c.mongoCollection)
}

// DeleteIndex deletes an index of a collection
func (c *collection[T]) DeleteIndex(ctx context.Context, indexName string) error {
	return deleteIndex(ctx, c.mongoCollection, indexName)
}

// Drop deletes a collection
func (c *collection[T]) Drop(ctx context.Context) error {
	return drop(ctx, c.mongoCollection)
}

// Name returns the name of a collection
func (c *collection[T]) Name() string {
	return c.mongoCollection.Name()
}

func validateReceivedID(id ID) error {
	if id == nil {
		return ErrEmptyID
	}

	return nil
}

func validateReceivedFilter(filter any) any {
	if filter == nil {
		return bson.M{}
	}

	return filter
}

func validateReceivedOrder(order map[string]OrderBy) (map[string]OrderBy, error) {
	if order == nil {
		return map[string]OrderBy{}, nil
	}

	for key, orderBy := range order {
		if key == "" {
			return order, fmt.Errorf("%w, %s", ErrInvalidOrder, "order key can not be empty")
		}

		if orderBy != OrderAsc && orderBy != OrderDesc {
			return order, fmt.Errorf("%w, %s", ErrInvalidOrder, "order value must be OrderAsc or OrderDesc")
		}
	}

	return order, nil
}

func validateReceivedIndex(index Index) error {
	if len(index.Keys) == 0 {
		return fmt.Errorf("%w: %s", ErrInvalidIndex, "keys can not be empty")
	}

	for key, orderBy := range index.Keys {
		if key == "" {
			return fmt.Errorf("%w: %s", ErrInvalidIndex, "key can not be empty")
		}

		if orderBy != OrderAsc && orderBy != OrderDesc {
			return fmt.Errorf("%w: %s", ErrInvalidIndex, "order must be OrderAsc or OrderDesc")
		}
	}

	return nil
}
