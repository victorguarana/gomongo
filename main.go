package main

import (
	"fmt"
	"gomongo/database/connection"
	"gomongo/database/mongo"
)

///////////////////////////////////
// Example of how to use gomongo //
///////////////////////////////////

var carroCollectionName = "carros"

type Carro struct {
	ID     string
	Marca  string
	Modelo string
	Ano    int
}

func main() {
	err := connection.Init("mongodb://localhost:27017", "Loja")
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

	carro.ID, err = mongo.Create(carroCollectionName, &carro)
	if err != nil {
		panic(err)
	}
	fmt.Println(carro)

	/////////////////////////////////
	// Example: Get first document //
	/////////////////////////////////
	firstCarro := Carro{}
	carroInterface, err := mongo.First(carroCollectionName)
	if err != nil {
		panic(err)
	}

	err = mongo.InterfaceToStruct(carroInterface, &firstCarro)
	if err != nil {
		panic(err)
	}

	fmt.Println(firstCarro)

	////////////////////////////////////////////////
	// Example: Update one document on Collection //
	////////////////////////////////////////////////
	firstCarro.Ano = 2023
	firstCarro.Modelo = "Civic"
	err = mongo.UpdateByID(carroCollectionName, firstCarro)
	if err != nil {
		panic(err)
	}

	//////////////////////////////////
	// Example: Count all documents //
	//////////////////////////////////
	count, err := mongo.Count(carroCollectionName)
	if err != nil {
		panic(err)
	}
	fmt.Println("Total de documentos no mongo:", count)

	////////////////////////////////////////////////
	// Example: Delete one document on Collection //
	////////////////////////////////////////////////
	deleteCarro := Carro{
		Marca:  "Fiat",
		Modelo: "Argo",
		Ano:    2022,
	}

	deleteID, _ := mongo.Create(carroCollectionName, &deleteCarro)
	deleteCarro.ID = deleteID

	err = mongo.DeleteByID(carroCollectionName, deleteCarro)
	if err != nil {
		panic(err)
	}

	//////////////////////////////////
	// Example: Search with FindOne //
	//////////////////////////////////
	whereFilter := map[string]string{"marca": "Honda"}
	findCarroInterface, err := mongo.FindOne(carroCollectionName, whereFilter)
	if err != nil {
		panic(err)
	}

	var findCarro Carro
	err = mongo.InterfaceToStruct(findCarroInterface, &findCarro)
	if err != nil {
		panic(err)
	}
	fmt.Println(findCarro)

	///////////////////////////////////////////
	// Example: List documents on Collection //
	///////////////////////////////////////////
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
