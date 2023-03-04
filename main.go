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

	err = mongo.Create(carroCollectionName, &carro)
	if err != nil {
		panic(err)
	}

	// Example: Get first document
	carro = Carro{}
	carroInterface, err := mongo.First(carroCollectionName)
	if err != nil {
		panic(err)
	}

	err = mongo.InterfaceToStruct(carroInterface, &carro)
	if err != nil {
		panic(err)
	}

	fmt.Println(carro)

	// Example: List documents on Collection
	listaCarrosInterface, err := mongo.All(carroCollectionName)
	if err != nil {
		panic(err)
	}

	var listaCarros []Carro
	for _, value := range listaCarrosInterface {
		var c Carro
		err = mongo.InterfaceToStruct(value, &c)
		if err != nil {
			panic(err)
		}

		listaCarros = append(listaCarros, c)
	}
	fmt.Println(listaCarros)
}
