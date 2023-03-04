package main

import (
	"fmt"
	"gomongo/database/connection"
	"gomongo/database/mongo"
)

var carroCollectionName = "carros"

type Carro struct {
	Marca  string
	Modelo string
	Ano    int
}

// Example of how to use
func main() {
	err := connection.Init("mongodb://localhost:27017", "Loja")
	if err != nil {
		panic(err)
	}

	// Example: Create document
	carro := Carro{
		Marca:  "Honda",
		Modelo: "City",
		Ano:    2022,
	}

	err = mongo.Create(carroCollectionName, carro)
	if err != nil {
		panic(err)
	}

	// Example: List documents on Collection
	listaCarros, err := mongo.All(carroCollectionName)
	if err != nil {
		panic(err)
	}
	fmt.Println(listaCarros)
}
