package mainv2

import (
	"fmt"
	"gomongo/database/connection"
	"gomongo/database/mongov2"
)

/////////////////////////////////////
// Example of how to use gomongoV2 //
/////////////////////////////////////

var collection = mongov2.NewCollection("carros")

type Carro struct {
	ID     string
	Marca  string
	Modelo string
	Ano    int
}

func Example() {
	err := connection.Init("mongodb://localhost:27017", "Loja")
	if err != nil {
		panic(err)
	}

	/////////////////////////////
	// Example: First document //
	/////////////////////////////
	carro := Carro{}
	err = collection.FirstV2(&carro)
	if err != nil {
		panic(err)
	}
	fmt.Println("First:", carro)
}
