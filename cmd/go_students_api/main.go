package main

import (
	"fmt"
	"log"

	"github.com/prashantkumbhar2002/go_students_api/internal/config"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()

	fmt.Println("Welcome to the Students API")
	fmt.Printf("Environment: %s\n", cfg.Env)
	fmt.Printf("Storage Path: %s\n", cfg.StoragePath)
	fmt.Printf("Server will run on: %s:%d\n", cfg.HTTPServer.Host, cfg.HTTPServer.Port)

	// TODO: Initialize logger
	// TODO: Initialize storage (database)
	// TODO: Initialize router & handlers
	// TODO: Start HTTP server

	log.Println("Starting Students API server...")
}
