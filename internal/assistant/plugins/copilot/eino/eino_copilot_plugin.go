package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/cloudwego/eino-ext/components/model/gemini"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
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

type QueryWeatherParams struct {
	City *string `json:"city,omitempty" jsonschema:"description=City"`
}

// 处理函数
func QueryWeatherFunc(_ context.Context, params *QueryWeatherParams) (string, error) {
	log.Println("querying weather")
	reqUrl := "https://wttr.in/?T"
	if len(*params.City) > 0 {
		reqUrl = "https://wttr.in/" + url.PathEscape(*params.City) + "?T"
		log.Println("request url", reqUrl)
	}
	resp, err := http.DefaultClient.Get(reqUrl)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	reply := string(data)
	log.Println("weather of", params.City, reply)
	return reply, nil
}

type LoggerCallback struct {
	callbacks.HandlerBuilder // 可以用 callbacks.HandlerBuilder 来辅助实现 callback
}

func (cb *LoggerCallback) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	fmt.Println("==================")
	inputStr, _ := json.MarshalIndent(input, "", "  ")
	fmt.Printf("[OnStart] %s\n", string(inputStr))
	return ctx
}

func (cb *LoggerCallback) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	fmt.Println("=========[OnEnd]=========")
	outputStr, _ := json.MarshalIndent(output, "", "  ")
	fmt.Println(string(outputStr))
	return ctx
}

func (cb *LoggerCallback) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	fmt.Println("=========[OnError]=========")
	fmt.Println(err)
	return ctx
}

func (cb *LoggerCallback) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo,
	output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {

	var graphInfoName = react.GraphName

	go func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("[OnEndStream] panic err:", err)
			}
		}()

		defer output.Close() // remember to close the stream in defer

		fmt.Println("=========[OnEndStream]=========")
		for {
			frame, err := output.Recv()
			if errors.Is(err, io.EOF) {
				// finish
				break
			}
			if err != nil {
				fmt.Printf("internal error: %s\n", err)
				return
			}

			s, err := json.Marshal(frame)
			if err != nil {
				fmt.Printf("internal error: %s\n", err)
				return
			}

			if info.Name == graphInfoName { // 仅打印 graph 的输出, 否则每个 stream 节点的输出都会打印一遍
				fmt.Printf("%s: %s\n", info.Name, string(s))
			}
		}

	}()
	return ctx
}

func (cb *LoggerCallback) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo,
	input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	defer input.Close()
	return ctx
}

// EinoCopilotPlugin is a mock implementation of the CopilotPlugin interface
type EinoCopilotPlugin struct{}

// Chat simulates streaming data chunks to the client
func (p EinoCopilotPlugin) Chat(req shared.UserRequest) (<-chan shared.ChunkData, error) {
	ch := make(chan shared.ChunkData)

	go func() {
		defer close(ch) // Ensure the channel is closed when done

		ctx := context.Background()

		// SOCKS proxy address
		proxyURL, err := url.Parse("socks5://127.0.0.1:1080")
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
		httpClient := &http.Client{Transport: httpTransport}

		client, err := genai.NewClient(ctx, &genai.ClientConfig{
			APIKey:     geminiApiKey,
			Backend:    genai.BackendGeminiAPI,
			HTTPClient: httpClient,
		})
		if err != nil {
			log.Fatal(err)
		}

		chatModel, err := gemini.NewChatModel(context.Background(), &gemini.Config{
			Client: client,
			Model:  "gemini-2.5-flash",
		})
		if err != nil {
			panic(err)
		}

		// prepare persona (system prompt) (optional)
		persona := `# Character: 你是聪明的个人助理，擅长使用工具并分析数据帮伙伴解决问题`

		// 使用 InferTool 创建工具
		updateTool, err := utils.InferTool(
			"query_weather", // tool name
			"A tool to query weather, will use default location if no city provide", // tool description
			QueryWeatherFunc)
		if err != nil {
			panic(err)
		}

		ragent, err := react.NewAgent(ctx, &react.AgentConfig{
			ToolCallingModel: chatModel,
			ToolsConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{updateTool},
			},
			// StreamToolCallChecker: toolCallChecker, // uncomment it to replace the default tool call checker with custom one
		})
		if err != nil {
			panic(err)
		}

		opt := []agent.AgentOption{
			agent.WithComposeOptions(compose.WithCallbacks(&LoggerCallback{})),
			//react.WithChatModelOptions(ark.WithCache(cacheOption)),
		}

		sr, err := ragent.Stream(ctx, []*schema.Message{
			{
				Role:    schema.System,
				Content: persona,
			},
			{
				Role:    schema.User,
				Content: req.Message,
			},
		}, opt...)
		if err != nil {
			log.Printf("failed to stream: %v", err)
			return
		}

		defer sr.Close() // remember to close the stream

		log.Printf("\n\n===== start streaming =====\n\n")

		for {
			msg, err := sr.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					// finish
					break
				}
				// error
				log.Printf("failed to recv: %v", err)
				return
			}

			// 打字机打印
			ch <- shared.ChunkData{
				Content: msg.Content,
			}
		}

		log.Printf("\n\n===== finished =====\n")
	}()

	return ch, nil
}

// AutoComplete simulates generating a single response
func (p EinoCopilotPlugin) AutoComplete(req shared.UserRequest) (string, error) {
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
var Plugin EinoCopilotPlugin
