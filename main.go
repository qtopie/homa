package main

import (
	"fmt"
	"log"
	"net"

	dapr "github.com/dapr/go-sdk/service/grpc"
	"github.com/qtopie/homa/gen/assistant" // Import the generated code
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	viper.SetConfigName("config") 
	viper.SetConfigType("ini")
	viper.AddConfigPath("$HOME/.homa") // call multiple times to add many search paths
	viper.AddConfigPath(".")           // optionally look for config in the working directory
	err := viper.ReadInConfig()        // Find and read the config file
	if err != nil {                    // Handle errors reading the config file
		panic(fmt.Errorf("fatal error config file: %w", err))
	}

	// Define the port for the gRPC server
	address := viper.GetString("address")
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

	fmt.Println("Starting process on", address)
	if err := daprService.Start(); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
