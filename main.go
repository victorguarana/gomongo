package test

import (
	"gomongo/database/connection"
	"gomongo/database/mongo"
)

// Example of how to use
func main() {
	err := connection.Init("URI", "DatabaseName")
	if err != nil {
		panic(err)
	}

	create()
}

func create() {
	object := map[string]string{"Name": "Victor"}

	mongo.Create("CollectionName", object)
}
