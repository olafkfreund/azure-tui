package ai

import (
	"context"
	"net/http"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type AIProvider struct {
	Client *openai.Client
}

func NewAIProvider(apiKey string) *AIProvider {
	// Check for MCP server endpoint in environment or use default
	mcpEndpoint := os.Getenv("AZURE_MCP_ENDPOINT")
	if mcpEndpoint == "" {
		mcpEndpoint = "http://localhost:5030/v1" // Default MCP server endpoint
	}
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = mcpEndpoint
	config.HTTPClient = &http.Client{}
	return &AIProvider{Client: openai.NewClientWithConfig(config)}
}

func (ai *AIProvider) Ask(question string, contextStr string) (string, error) {
	resp, err := ai.Client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    openai.GPT4,
		Messages: []openai.ChatCompletionMessage{{Role: openai.ChatMessageRoleUser, Content: contextStr + "\n" + question}},
	})
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

func (ai *AIProvider) SummarizeResourceGroups(groups []string) (string, error) {
	prompt := "Summarize the following Azure resource groups and suggest improvements or optimizations:\n" + strings.Join(groups, ", ")
	return ai.Ask(prompt, "Azure Resource Management")
}
