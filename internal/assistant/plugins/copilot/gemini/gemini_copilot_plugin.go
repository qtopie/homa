package main

import (
	"context"
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

		for chunk, _ := range stream {
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

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.0-flash",
		genai.Text(req.Message),
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	return result.Text(), nil
}

// Export the mock plugin instance
var Plugin GeminiCopilotPlugin
