package main

import (
	"fmt"
	"log"
	"net/http"

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


	// Initialize router & handlers
	router := http.NewServeMux()

	router.HandleFunc("GET /", func (w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is Home page,.... It works!"))
	})

	// Start HTTP server
	server := http.Server{
		Addr: fmt.Sprintf("%s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler: router,
		// ReadTimeout: cfg.HTTPServer.Timeout,
		// IdleTimeout: cfg.HTTPServer.IdleTimeout,
	}
    
	log.Printf("Starting Students API server at address: %s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Failed to start server: %s", err.Error())
	}

	log.Println("Server started on port", cfg.HTTPServer.Port)
}
