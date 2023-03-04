package test

import (
	"gomongo/database/connection"
)

// Example of how to use
func main() {
	err := connection.Init("URI", "DatabaseName")
	if err != nil {
		panic(err)
	}
}
