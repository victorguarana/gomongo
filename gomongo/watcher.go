package gomongo

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/r3labs/diff"
)

// IHistoryCollection should always be implemented by Collection
var _ IHistoryCollection = Collection[History]{}

// Watcher should always implement IWatcher
var _ IWatcher = Watcher{}

type IWatcher interface {
	Watch(ctx context.Context, collections ...string) error
}

type Watcher struct {
	database          Database
	historyCollection Collection[History]
}

func NewWatcher(database Database, historyCollectionName string) (Watcher, Collection[History], error) {
	if err := validateDatabase(database); err != nil {
		return Watcher{}, Collection[History]{}, err
	}

	historyCollection, err := NewCollection[History](database, historyCollectionName)
	if err != nil {
		return Watcher{}, Collection[History]{}, err
	}

	w := Watcher{
		database:          database,
		historyCollection: historyCollection,
	}

	return w, w.historyCollection, nil
}

func (w Watcher) Watch(ctx context.Context, collectionNamesToWatch ...string) error {
	err := watch(ctx, w.database.mongoDatabase, w.handleEvents, collectionNamesToWatch)
	if err != nil {
		return err
	}

	return nil
}

func (w Watcher) handleEvents(ctx context.Context, e event) error {
	id, ok := e.DocumentKey["_id"].(ID)
	if !ok {
		return fmt.Errorf("could not get id from event document")
	}

	last, err := lastInsertedByObjectID(ctx, w.historyCollection, id)
	if err != nil && err != ErrDocumentNotFound {
		return fmt.Errorf("failed to get last entry")
	}

	updatedFields, err := updatedFields(last.Modified, e.FullDocument)
	if err != nil {
		return err
	}

	history := History{
		CreatedAt:      e.ClusterTime,
		CollectionName: e.NS.Collection,
		ObjectID:       id,
		Modified:       e.FullDocument,
		Action:         e.OperationType,
		UpdatedFields:  updatedFields,
	}

	_, err = w.historyCollection.Create(ctx, history)
	if err != nil && !errors.Is(err, ErrDuplicateKey) {
		return err
	}

	return nil
}

func lastInsertedByObjectID(ctx context.Context, c Collection[History], objectID ID) (History, error) {
	filter := map[string]any{"objectid": objectID}
	return c.LastInserted(ctx, filter)
}

func updatedFields(last any, history any) (map[string]UpdatedField, error) {
	changes, err := diff.Diff(last, history)
	if err != nil {
		return nil, fmt.Errorf("failed to get diff between entries")
	}

	updatedFields := make(map[string]UpdatedField, len(changes))
	for _, change := range changes {
		field := strings.Join(change.Path, ".")
		if field != "_id" {
			updatedFields[field] = UpdatedField{
				Old: change.From,
				New: change.To,
			}
		}
	}

	return updatedFields, nil
}
