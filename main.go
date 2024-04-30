package main

import (
	"context"
	"fmt"
	"time"

	"github.com/victorguarana/gomongo/gomongo"
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

}
