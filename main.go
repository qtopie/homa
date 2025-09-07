package main

import (
	"fmt"
	"log"
	"net"

	"github.com/go-viper/encoding/ini"
	"github.com/qtopie/homa/gen/assistant"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	viperCfg *viper.Viper
)

func init() {
	// Configure Viper to read the config file
	codecRegistry := viper.NewCodecRegistry()
	codecRegistry.RegisterCodec("ini", ini.Codec{})

	viperCfg = viper.NewWithOptions(
		viper.WithCodecRegistry(codecRegistry),
	)

	viperCfg.SetConfigName("config")
	viperCfg.SetConfigType("ini")
	viperCfg.AddConfigPath(".")
	viperCfg.AddConfigPath("$HOME/.homa")
	if err := viperCfg.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}
}

func main() {
	// Initialize the PluginManager
	pluginManager := NewPluginManager("/opt/homa/plugins")

	// Create the CopilotServiceServerImpl
	copilotService := NewCopilotServiceServerImpl(pluginManager)

	// Start the gRPC server
	address := viperCfg.GetString("app.address")
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
