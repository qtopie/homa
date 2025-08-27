package main

import (
	"context"
	"log"
	"net"
	"time"

	"github.com/dapr/go-sdk/service/common"
	dapr "github.com/dapr/go-sdk/service/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Define the first streaming service
type streamingServiceOne struct {
	UnimplementedStreamingServiceOneServer
}

// StreamDataOne implements server streaming for the first service
func (s *streamingServiceOne) StreamDataOne(req *StreamRequest, stream StreamingServiceOne_StreamDataOneServer) error {
	log.Printf("Received request for StreamDataOne: %s", req.Message)

	// Simulate streaming data
	for i := 0; i < 5; i++ {
		resp := &StreamResponse{
			Data: req.Message + " - Response from Service One " + time.Now().Format(time.RFC3339),
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
		time.Sleep(1 * time.Second) // Simulate delay
	}

	return nil
}

// Define the second streaming service
type streamingServiceTwo struct {
	UnimplementedStreamingServiceTwoServer
}

// StreamDataTwo implements server streaming for the second service
func (s *streamingServiceTwo) StreamDataTwo(req *StreamRequest, stream StreamingServiceTwo_StreamDataTwoServer) error {
	log.Printf("Received request for StreamDataTwo: %s", req.Message)

	// Simulate streaming data
	for i := 0; i < 5; i++ {
		resp := &StreamResponse{
			Data: req.Message + " - Response from Service Two " + time.Now().Format(time.RFC3339),
		}
		if err := stream.Send(resp); err != nil {
			return err
		}
		time.Sleep(1 * time.Second) // Simulate delay
	}

	return nil
}

func main() {
	// Define the port for the gRPC server
	port := ":1234"

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register the first streaming service
	RegisterStreamingServiceOneServer(grpcServer, &streamingServiceOne{})

	// Register the second streaming service
	RegisterStreamingServiceTwoServer(grpcServer, &streamingServiceTwo{})

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Create a new Dapr service and attach it to the same gRPC server
	daprService, err := dapr.NewServiceWithGrpcServer(port, grpcServer)
	if err != nil {
		log.Fatalf("Failed to create Dapr service: %v", err)
	}

	// Add a unary service invocation handler
	if err := daprService.AddServiceInvocationHandler("echo", echoHandler); err != nil {
		log.Fatalf("Error adding invocation handler: %v", err)
	}

	// Start the combined gRPC server
	log.Printf("Starting combined gRPC server on port %s...", port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port)
	}
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}

// Dapr service invocation handler
func echoHandler(ctx context.Context, in *common.InvocationEvent) (out *common.Content, err error) {
	log.Printf("echo - ContentType:%s, Verb:%s, QueryString:%s, %+v", in.ContentType, in.Verb, in.QueryString, string(in.Data))
	out = &common.Content{
		Data:        in.Data,
		ContentType: in.ContentType,
		DataTypeURL: in.DataTypeURL,
	}
	return
}
