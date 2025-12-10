package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prashantkumbhar2002/go_students_api/internal/config"
	"github.com/prashantkumbhar2002/go_students_api/internal/http/handlers/students"
	"github.com/prashantkumbhar2002/go_students_api/internal/storage/sqlite"
)

func main() {
	// Load configuration
	cfg := config.MustLoad()

	fmt.Println("Welcome to the Students API")
	fmt.Printf("Environment: %s\n", cfg.Env)
	fmt.Printf("Storage Path: %s\n", cfg.StoragePath)
	fmt.Printf("Server will run on: %s:%d\n", cfg.HTTPServer.Host, cfg.HTTPServer.Port)

	// TODO: Initialize logger


	// Initialize storage (database)
	storage, err := sqlite.NewSqlite(cfg)
	if err != nil {
		log.Fatalf("Error initializing SQLite storage: %v", err)
	}

	log.Println("SQLite storage initialized successfully")

	// Initialize router & handlers
	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("This is Home page,.... It works!"))
	})

	router.HandleFunc("POST /students", students.New(storage))

	router.HandleFunc("GET /slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.Write([]byte("This is Slow page,.... It works!"))
	})


	// Start HTTP server
	server := &http.Server{
		Addr:        fmt.Sprintf("%s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port),
		Handler:     router,
		ReadTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout: cfg.HTTPServer.IdleTimeout,
	}

	// Create context that listens for shutdown signals (Ctrl+C, SIGINT, SIGTERM)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Buffered channel to receive server errors
	// Buffer size 1 prevents goroutine from blocking if error occurs before select
	serverErrors := make(chan error, 1)

	// Start server in goroutine so main thread can listen for shutdown signals
	go func() {
		log.Printf("Starting server on %s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port)

		err := server.ListenAndServe()

		// Only send error if it's NOT the expected shutdown error
		// http.ErrServerClosed is returned when Shutdown() is called - this is normal
		if err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	// Wait for either:
	// 1. Server error (startup failure or runtime error)
	// 2. Shutdown signal (Ctrl+C, SIGINT, SIGTERM)
	select {
	case err := <-serverErrors:
		// Server encountered an error
		log.Fatalf("Server error: %v", err)

	case <-ctx.Done():
		// Shutdown signal received
		log.Println("Shutdown signal received, initiating graceful shutdown...")

		// Create a context with timeout for the shutdown process
		// Server has shutdown timeout to finish active requests
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTPServer.ShutdownTimeout)
		defer cancel()

		// Attempt graceful shutdown
		// This stops accepting new requests and waits for active ones to complete
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Error during shutdown: %v", err)
			// Force close if graceful shutdown fails
			server.Close()
		}

		log.Println("Server stopped gracefully")
	}
}
