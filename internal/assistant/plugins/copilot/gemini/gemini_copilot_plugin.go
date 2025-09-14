package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"

	cfg "github.com/qtopie/homa/internal/app/config"
	"github.com/qtopie/homa/internal/assistant/plugins/copilot/shared"
	"golang.org/x/net/proxy"
	"google.golang.org/genai"
)

var (
	httpClient   *http.Client
	geminiApiKey string

	codeCompletionPrompt string
)

func init() {
	geminiApiKey = cfg.GetAppConfig().GetString("services.gemini.api-key")
	if geminiApiKey == "" {
		geminiApiKey = os.Getenv("GOOGLE_API_KEY")
		log.Fatal("GOOGLE_API_KEY environment variable not set")
	}

	proxyUrl := cfg.GetAppConfig().GetString("app.proxy-url")
	if proxyUrl == "" {
		proxyUrl = os.Getenv("https_proxy")
	}

	if len(proxyUrl) > 0 {
		// SOCKS proxy address
		proxyURL, err := url.Parse(proxyUrl)
		if err != nil {
			panic(err)
		}

		// Create a SOCKS5 dialer
		dialer, err := proxy.FromURL(proxyURL, proxy.Direct)
		if err != nil {
			panic(err)
		}

		// Create HTTP client with the SOCKS5 dialer
		httpTransport := &http.Transport{
			Dial: dialer.Dial,
		}
		httpClient = &http.Client{Transport: httpTransport}
	} else {
		httpClient = &http.Client{}
	}

	// 1. Define the template string
	codeCompletionPrompt = `
	You are a highly skilled and efficient code completion assistant. Your task is to generate the most logical and correct completion for the code snippet provided below.

The user is working in a file named: {{.filename}} in workspace {{.workspace}}

The full code content of the file is:
---
{{.code_before_cursor}}
{{.cursor_here}}{{.code_after_cursor}}
---

Your response should be a clean, single block of code that logically follows the cursor position. Do not include any extra text, explanations, or conversational filler. Just the code.
	`

	

}

// GeminiCopilotPlugin is a mock implementation of the CopilotPlugin interface
type GeminiCopilotPlugin struct{}

// Chat simulates streaming data chunks to the client
func (p GeminiCopilotPlugin) Chat(req shared.UserRequest) (<-chan shared.ChunkData, error) {
	ch := make(chan shared.ChunkData)
	ctx := context.Background()

	go func() {
		defer close(ch) // Ensure the channel is closed when done

		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:     geminiApiKey,
			Backend:    genai.BackendGeminiAPI,
			HTTPClient: httpClient,
		})
		if err != nil {
			log.Fatal(err)
		}

		stream := client.Models.GenerateContentStream(
			ctx,
			"gemini-2.5-flash",
			genai.Text(req.Message),
			nil,
		)

		for chunk := range stream {
			part := chunk.Candidates[0].Content.Parts[0]
			ch <- shared.ChunkData{
				Content: part.Text,
			}
		}
	}()

	return ch, nil
}

// AutoComplete simulates generating a single response
func (p GeminiCopilotPlugin) AutoComplete(req shared.UserRequest) (string, error) {
	ctx := context.Background()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:     geminiApiKey,
		Backend:    genai.BackendGeminiAPI,
		HTTPClient: httpClient,
	})
	if err != nil {
		log.Fatal(err)
	}

	data, err := json.Marshal(req)
	if err != nil {
		log.Fatal(err)
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		genai.Text(string(data)),
		&genai.GenerateContentConfig{
			SystemInstruction: &genai.Content{
					Role:  genai.RoleModel,
					Parts: []*genai.Part{{Text: codeCompletionPrompt}},
			},
		},
	)
	if err != nil {
		log.Fatal(err)
	}
	return result.Text(), nil
}

// Export the mock plugin instance
var Plugin GeminiCopilotPlugin
