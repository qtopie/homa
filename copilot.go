package main

import (
	"context"
	"log"
	"time"

	"github.com/qtopie/homa/gen/assistant" // Import the generated code
)

// CopilotServiceServerImpl is the implementation of the ChatService
type CopilotServiceServerImpl struct {
    assistant.UnimplementedCopilotServiceServer
}

// Chat implements the server streaming method for ChatService
func (s *CopilotServiceServerImpl) Chat(req *assistant.UserRequest, stream assistant.CopilotService_ChatServer) error {
    log.Printf("Received Chat request: %s", req.Message)

    // Simulate streaming responses
    for i := 0; i < 5; i++ {
        resp := &assistant.StreamResponse{
            Content: req.Message + " - Response " + time.Now().Format(time.RFC3339),
        }
        if err := stream.Send(resp); err != nil {
            return err
        }
        time.Sleep(1 * time.Second) // Simulate delay
    }

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