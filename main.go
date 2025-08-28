package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/dapr/go-sdk/service/common"
	dapr "github.com/dapr/go-sdk/service/grpc"
	"github.com/qtopie/homa/gen/assistant" // Import the generated code
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Define the port for the gRPC server
	address := ":1234"
	lis, err := net.Listen("tcp", address)
	if err != nil {
		err = fmt.Errorf("failed to TCP listen on %s: %w", address, err)
		return
	}

	// Create a new gRPC server
	grpcServer := grpc.NewServer()

	// Register the ChatService implementation
	chatService := &CopilotServiceServerImpl{}
	assistant.RegisterCopilotServiceServer(grpcServer, chatService)

	// Enable reflection for debugging
	reflection.Register(grpcServer)

	// Create a new Dapr service and attach it to the same gRPC server
	daprService := dapr.NewServiceWithGrpcServer(lis, grpcServer)

	// Add a unary service invocation handler
	if err := daprService.AddServiceInvocationHandler("echo", echoHandler); err != nil {
		log.Fatalf("Error adding invocation handler: %v", err)
	}

	fmt.Println("Starting process on", address)
	if err := daprService.Start(); err != nil {
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
