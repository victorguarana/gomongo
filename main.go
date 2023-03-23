package main

import (
	"fmt"
	"gomongo/database/mongo"
)

///////////////////////////////////
// Example of how to use gomongo //
///////////////////////////////////

var carroCollection = mongo.NewCollection("carros")

type Carro struct {
	ID     string
	Marca  string
	Modelo string
	Ano    int
}

func main() {
	err := mongo.Init("mongodb://localhost:27017", "Loja")
	if err != nil {
		panic(err)
	}

	//////////////////////////////
	// Example: Create document //
	//////////////////////////////
	carro := Carro{
		Marca:  "Honda",
		Modelo: "City",
		Ano:    2022,
	}

	carro.ID, err = carroCollection.Create(carro)
	if err != nil {
		panic(err)
	}
	fmt.Println("Created:", carro)

	/////////////////////////////////
	// Example: Get first document //
	/////////////////////////////////
	firstCarro := Carro{}
	carroInterface, err := carroCollection.First()
	if err != nil {
		panic(err)
	}

	err = mongo.InterfaceToStruct(carroInterface, &firstCarro)
	if err != nil {
		panic(err)
	}

	fmt.Println("First:", firstCarro)

	//////////////////////////////////
	// Example: Search with FindOne //
	//////////////////////////////////
	findFilter := map[string]string{"modelo": "City"}
	findCarroInterface, err := carroCollection.FindOne(findFilter)
	if err != nil {
		panic(err)
	}

	var findCarro Carro
	err = mongo.InterfaceToStruct(findCarroInterface, &findCarro)
	if err != nil {
		panic(err)
	}
	fmt.Println("Find:", findCarro)

	////////////////////////////////////////////////
	// Example: Update one document on Collection //
	////////////////////////////////////////////////
	findCarro.Ano = 2023
	findCarro.Modelo = "Civic"
	err = carroCollection.UpdateID(findCarro.ID, findCarro)
	if err != nil {
		panic(err)
	}
	fmt.Println("Update:", findCarro)

	//////////////////////////////////
	// Example: Count all documents //
	//////////////////////////////////
	count, err := carroCollection.Count()
	if err != nil {
		panic(err)
	}
	fmt.Println("Count:", count)

	////////////////////////////////////////////////
	// Example: Delete one document on Collection //
	////////////////////////////////////////////////
	deleteCarro := Carro{
		Marca:  "Fiat",
		Modelo: "Argo",
		Ano:    2022,
	}

	deleteID, _ := carroCollection.Create(deleteCarro)
	deleteCarro.ID = deleteID

	err = carroCollection.DeleteID(deleteCarro.ID)
	if err != nil {
		panic(err)
	}

	fmt.Println("Deleted:", deleteCarro)

	///////////////////////////////////////////
	// Example: List documents on Collection //
	///////////////////////////////////////////
	listaCarrosInterface, err := carroCollection.All()
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
	fmt.Println("All:", listaCarros)

	////////////////////////////////
	// Example: Search with Where //
	////////////////////////////////
	whereCarrosInterface, err := carroCollection.Where(map[string]string{"marca": "Honda"})
	if err != nil {
		panic(err)
	}

	var whereCarros []Carro
	for _, value := range whereCarrosInterface {
		var c Carro
		err = mongo.InterfaceToStruct(value, &c)
		if err != nil {
			panic(err)
		}

		whereCarros = append(whereCarros, c)
	}
	fmt.Println("Where:", whereCarros)
}
