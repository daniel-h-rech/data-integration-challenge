package main

import (
	"github.com/daniel-h-rech/data-integration-challenge/api"
	"github.com/daniel-h-rech/data-integration-challenge/data"
	. "log"
	. "os"
)

//go:generate swagger generate spec
func main() {

	if len(Args) < 3 {
		Println("Usage:")
		Println("\t" + "integration <CSV file path> <MongoDB network address>")
		Println()
		Println("Example:")
		Println("\t" + "integration q1_catalog.csv localhost:27017")
		return
	}

	csvFilepath := Args[1]
	mongoDBAddress := Args[2]

	data.LoadCompanyData(csvFilepath, mongoDBAddress)

	api.Start(":8080", mongoDBAddress)
	Println("exit")
}
