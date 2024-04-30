package gomongo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/exp/slices"
)

type namespace struct {
	Database   string `bson:"db"`
	Collection string `bson:"coll"`
}

type updateDescription struct {
	RemovedFields bson.A `bson:"removedFields"`
	UpdatedFields bson.M `bson:"updatedFields"`
}

type event struct {
	NS                namespace         `bson:"ns"`
	ClusterTime       time.Time         `bson:"clusterTime"`
	FullDocument      bson.M            `bson:"fullDocument"`
	DocumentKey       bson.M            `bson:"documentKey"`
	UpdateDescription updateDescription `bson:"updateDescription"`
	OperationType     string            `bson:"operationType"`
}

func watch(ctx context.Context, mongoDatabase *mongo.Database, handleEvent func(ctx context.Context, e event) error, collectionNamesToWatch []string) error {
	cs, err := mongoDatabase.Watch(ctx, mongo.Pipeline{}, options.ChangeStream().SetFullDocument(options.UpdateLookup))
	if err != nil {
		return err
	}

	defer cs.Close(ctx)
	for cs.Next(ctx) {
		var e event
		if err := cs.Decode(&e); err != nil {
			return err
		}

		if collectionBellongToWatch(e.NS.Collection, collectionNamesToWatch) {
			err := handleEvent(ctx, e)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func collectionBellongToWatch(collectionName string, collections []string) bool {
	return slices.Contains(collections, collectionName)
}
