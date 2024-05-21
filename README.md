[![Go Reference](https://pkg.go.dev/badge/github.com/victorguarana/gomongo.svg)](https://pkg.go.dev/github.com/victorguarana/gomongo)
# GoMongo

GoMongo is an Object-Relational Mapping (ORM) library for MongoDB in Go. It simplifies the process of interacting with MongoDB, allowing developers to perform Mongo operations in an intuitive and efficient manner.

## Installation

To install GoMongo, you can use the following go get command:

```bash
go get github.com/victorguarana/gomongo
```

## Basic Usage
The code bellow sets up a connection to a MongoDB database, creates a collection for movies, inserts a movie document, updates the movie document, retrieves all movies, and finally deletes the movie document. This code is used to demonstrate basic CRUD (Create, Read, Update, Delete) operations with MongoDB using the gomongo package.

```go
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/victorguarana/gomongo"
)

type Movie struct {
	ID   gomongo.ID `bson:"_id"`
	Name string
	Year int
}

func main() {
	// Setting up the connection to the database
	connectionSettings := gomongo.ConnectionSettings{
		URI:               "mongodb://localhost:27017",
		DatabaseName:      "mydatabase",
		ConnectionTimeout: 60 * time.Second,
	}
	database, err := gomongo.NewDatabase(context.Background(), connectionSettings)
	if err != nil {
		panic(err)
	}

	// Creating a collection
	moviesCollection, err := gomongo.NewCollection[Movie](database, "mymovies")
	if err != nil {
		panic(err)
	}

	// Inserting a movie
	starWarsIV := Movie{
		Name: "Star Wars",
		Year: 1977,
	}
	starWarsIV.ID, err = moviesCollection.Create(context.Background(), starWarsIV)
	if err != nil {
		panic(err)
	}

	// Updating a movie
	starWarsIV.Name = "Star Wars: Episode IV - A New Hope"
	err = moviesCollection.UpdateID(context.Background(), starWarsIV.ID, starWarsIV)
	if err != nil {
		panic(err)
	}

	// Listing all movies
	allMovies, err := moviesCollection.All(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("All movies from Mongo: ", allMovies)

	// Deleting a movie
	err = moviesCollection.DeleteID(context.Background(), starWarsIV.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println("CRUD operations performed successfully!")
}
```

### Available Collection Interface
```go
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
```

## Contributing

Contributions are welcome! Before submitting a pull request, make sure the code is properly tested and follows the code style guidelines.

## License

This project is licensed under the [MIT License](https://github.com/victorguarana/gomongo/blob/main/LICENSE).

