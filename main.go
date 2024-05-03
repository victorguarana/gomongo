package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/victorguarana/gomongo/gomongo"
)

type HistoryEntry struct {
	Message    string         `json:"message"`
	Old        map[string]any `json:"old"`
	New        map[string]any `json:"new"`
	Author     string         `json:"author"`
	CreatedAt  time.Time      `json:"createdAt"`
	Collection string         `json:"collection"`
}
type Movie struct {
	ID   gomongo.ID `bson:"_id"`
	Name string
	Year int
}

func main() {
	cs := gomongo.ConnectionSettings{
		URI:          "mongodb://localhost:27017",
		DatabaseName: "test",
	}

	db, err := gomongo.NewDatabase(context.Background(), cs)
	if err != nil {
		log.Fatal(err)
	}

	movieCollection, err := gomongo.NewCollection[Movie](db, "movies")
	if err != nil {
		log.Fatal(err)
	}

	// historyRepository realiza a gestão do histórico
	historyWatcher, historyRepository, err := gomongo.NewWatcher(db, "history")
	if err != nil {
		log.Fatal(err)
	}

	ctx, ctxCancel := context.WithCancel(context.Background())
	defer ctxCancel()
	go historyWatcher.Watch(ctx, movieCollection.Name())

	// History usage

	r, err := historyRepository.All(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	list := []HistoryEntry{}
	for _, doc := range r {
		message := ""
		name, ok := doc.Modified["name"].(string)
		if ok && len(name) > 0 {
			message = fmt.Sprintf("%s %s (id: %s) ", doc.CollectionName, doc.Modified["name"], doc.ObjectID)
		} else {
			message = fmt.Sprintf("%s (id: %s) ", doc.CollectionName, doc.ObjectID)
		}
		switch doc.Action {
		case "insert":
			message += "was created."
		case "update":
			message += fmt.Sprintf("has been updated. The following fields have been changed: %s", listUpdatedFields(doc))
		case "delete":
			message += "was deleted."
		}
		author, ok := doc.Modified["author"].(string)
		if !ok {
			author = "Undefined"
		}
		old := make(map[string]any)
		current := make(map[string]any)

		for name, u := range doc.UpdatedFields {
			if name != "author" {
				if u.Old != nil {
					old[name] = u.Old
				}
				if u.New != nil {
					current[name] = u.New
				}
			}
		}

		list = append(list, HistoryEntry{
			Message:    message,
			Old:        old,
			New:        current,
			CreatedAt:  doc.CreatedAt,
			Collection: doc.CollectionName,
			Author:     author,
		})

		fmt.Println(list)
	}
}

func listUpdatedFields(d gomongo.History) string {
	var s []string
	for name := range d.UpdatedFields {
		s = append(s, name)
	}
	return strings.Join(s, ", ")
}
