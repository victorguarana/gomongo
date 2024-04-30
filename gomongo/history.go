package gomongo

import (
	"context"
	"time"
)

type IHistoryCollection interface {
	All(ctx context.Context) ([]History, error)
	Count(ctx context.Context) (int, error)
	First(ctx context.Context) (History, error)
	Last(ctx context.Context) (History, error)
	Where(ctx context.Context, filter any) ([]History, error)

	Drop(ctx context.Context) error

	Name() string
}

type History struct {
	CreatedAt      time.Time
	CollectionName string
	ObjectID       ID
	Modified       map[string]any
	UpdatedFields  map[string]UpdatedField
	Action         string
}

type UpdatedField struct {
	Old any
	New any
}
