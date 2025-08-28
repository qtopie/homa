package main

import (
	"log"
	"time"

	"github.com/qtopie/homa/gen/assistant" // Import the generated code
)

// ChatServiceServerImpl is the implementation of the ChatService
type ChatServiceServerImpl struct {
    assistant.UnimplementedChatServiceServer
}

// Chat implements the server streaming method for ChatService
func (s *ChatServiceServerImpl) Chat(req *assistant.UserRequest, stream assistant.ChatService_ChatServer) error {
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