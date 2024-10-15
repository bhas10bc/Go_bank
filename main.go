package main

import (
	"fmt"
	"log"

)

func main() {

	// Initialize the store with environment variables
	store, err := NewPostgresStore()
	if err != nil {
		log.Fatal(err)
	}

	// Initialize the store (create tables, etc.)
	err = store.Init()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%+v\n", store)
	
	// Set up and run the API server
	server := NewApiServer(":8080", store, "jwtSecret")
	server.Run()
	fmt.Println("Server started")
}
