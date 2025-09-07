package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/qtopie/homa/gen/assistant" // Import the generated code
	"github.com/qtopie/homa/internal/assistant/plugins/copilot/shared"
)

// CopilotServiceServerImpl is the implementation of the ChatService
type CopilotServiceServerImpl struct {
	assistant.UnimplementedCopilotServiceServer
	pluginManager *PluginManager
	currentPlugin CopilotPlugin
	currentName   string
	mu            sync.Mutex
}

// NewCopilotServiceServerImpl creates a new instance of CopilotServiceServerImpl
func NewCopilotServiceServerImpl(pluginManager *PluginManager) *CopilotServiceServerImpl {
	return &CopilotServiceServerImpl{
		pluginManager: pluginManager,
	}
}

// Chat implements the server streaming method for ChatService
func (s *CopilotServiceServerImpl) Chat(req *assistant.UserRequest, stream assistant.CopilotService_ChatServer) error {
	// Get the plugin name from the configuration
	copilotPluginName := viperCfg.GetString("plugins.copilot")
	if copilotPluginName == "" {
		return fmt.Errorf("no copilot plugin specified in configuration")
	}

	// Load the plugin only if it has changed
	s.mu.Lock()
	if s.currentName != copilotPluginName {
		log.Printf("Loading copilot plugin: %s", copilotPluginName)

		// Load the plugin dynamically
		err := s.pluginManager.LoadPlugin("copilot", copilotPluginName)
		if err != nil {
			s.mu.Unlock()
			log.Printf("Error loading copilot plugin %s: %v", copilotPluginName, err)
			return err
		}

		// Retrieve the loaded plugin
		plugin, exists := s.pluginManager.GetPlugin("copilot", copilotPluginName)
		if !exists {
			s.mu.Unlock()
			return fmt.Errorf("copilot plugin %s not found", copilotPluginName)
		}

		// Assert the plugin to the CopilotPlugin interface
		copilotPlugin, ok := plugin.(CopilotPlugin)
		if !ok {
			s.mu.Unlock()
			return fmt.Errorf("plugin %s does not implement CopilotPlugin interface", copilotPluginName)
		}

		// Update the current plugin and name
		s.currentPlugin = copilotPlugin
		s.currentName = copilotPluginName
	}
	s.mu.Unlock()

	// Forward the request to the plugin's Chat method
	pluginStream, err := s.currentPlugin.Chat(shared.UserRequest{Message: req.Message})
	if err != nil {
		log.Printf("Error calling Chat on plugin %s: %v", s.currentName, err)
		return err
	}

	// Consume the plugin's stream and forward to gRPC stream
	for chunk := range pluginStream {
		// Send each chunk to the gRPC stream
		resp := &assistant.StreamResponse{
			Content: chunk.Content,
		}
		if err := stream.Send(resp); err != nil {
			log.Printf("Error sending response to gRPC stream: %v", err)
			return err
		}

		// Check if this is the last chunk
		if chunk.IsLast {
			log.Printf("Received end signal from plugin %s", s.currentName)
			break
		}
	}

	log.Printf("Chat request completed for message: %s", req.Message)
	return nil
}

// AutoComplete implements the unary method for AutoComplete
func (s *CopilotServiceServerImpl) AutoComplete(ctx context.Context, req *assistant.UserRequest) (*assistant.AgentResponse, error) {
	log.Printf("Received AutoComplete request: %s", req.Message)

	// Simulate generating a response
	resp := &assistant.AgentResponse{
		Content: "AutoComplete response for: " + req.Message,
	}

	return resp, nil
}