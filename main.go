package main

import "github.com/victorguarana/gomongo/gomongo"

func main() {
	// Create a connection settings
	gomongoConnectionSettings := gomongo.ConnectionSettings().
		SetURI("mongodb://localhost:27017").
		SetDatabaseName("test").
		SetTimeout(10)

	// Initialize the connection
	if err := gomongo.Init(gomongoConnectionSettings); err != nil {
		panic(err)
	}
}
