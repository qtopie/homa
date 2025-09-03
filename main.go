package main

import (
	"fmt"
	"log"
	"net"

	"github.com/qtopie/homa/gen/assistant"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	// Configure Viper to read the config file
	viper.SetConfigName("config")
	viper.SetConfigType("ini")
	viper.AddConfigPath("$HOME/.homa")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	// Initialize the PluginManager
	pluginManager := NewPluginManager("/opt/homa/plugins")

	// Create the CopilotServiceServerImpl
	copilotService := NewCopilotServiceServerImpl(pluginManager)

	// Start the gRPC server
	address := viper.GetString("app.system.address")
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("Failed to listen on %s: %v", address, err)
	}

	grpcServer := grpc.NewServer()
	assistant.RegisterCopilotServiceServer(grpcServer, copilotService)
	reflection.Register(grpcServer)

	fmt.Println("Starting process on", address)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
