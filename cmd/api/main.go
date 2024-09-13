package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"space-booking/internal/server"
	"syscall"
	"time"
)

func main() {
	// Create a new server instance
	srv := server.NewServer()

	// Create a listener on the desired address
	listener, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		log.Fatalf("Error creating listener: %v", err)
	}

	// Channel to receive errors from the server
	errChan := make(chan error, 1)

	// Start the server in a goroutine
	go func() {
		log.Printf("Server started on %s...", srv.Addr)
		if err := srv.Serve(listener); err != nil && err != http.ErrServerClosed {
			// Log the error
			log.Printf("Server encountered an error: %v", err)
			// Send the error to errChan
			errChan <- err
		}
	}()

	// Set up channel to receive OS signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	// Wait for an interrupt or server error
	select {
	case err := <-errChan:
		// Server encountered an unexpected error
		log.Fatalf("Server error: %v", err)
	case sig := <-stop:
		// Received an interrupt signal, shut down gracefully
		log.Printf("Received signal %s, initiating graceful shutdown", sig)

		// Create a deadline to wait for the server to shut down
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Attempt a graceful shutdown
		if err := srv.Shutdown(ctx); err != nil {
			log.Fatalf("Could not gracefully shut down the server: %v", err)
		}

		log.Println("Server gracefully stopped")
	}
}
