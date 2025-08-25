package main

import (
	"log"

	dapr "github.com/dapr/go-sdk/service/grpc"
)

func main() {
	// Create a new Dapr gRPC service
	port := ":1234" // The port your app will listen on
	s, err := dapr.NewService(port)
	if err != nil {
		log.Fatalf("Failed to start Dapr service: %v", err)
	}

	// Start the Dapr runtime
	log.Printf("Starting Dapr gRPC service on port %s...", port)
	if err := s.Start(); err != nil {
		log.Fatalf("Dapr service failed to start: %v", err)
	}
}
