package openai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

// GitHubCopilotConfig represents the structure of the GitHub Copilot config file
type GitHubCopilotConfig struct {
	Apps map[string]struct {
		OAuthToken string `json:"oauth_token"`
	} `json:"apps"`
}

// getGitHubCopilotToken attempts to read the OAuth token from GitHub Copilot config
func getGitHubCopilotToken() string {
	// Check environment variable first
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		return token
	}

	// Try to read from GitHub Copilot config file
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	configPath := filepath.Join(homeDir, ".config", "github-copilot", "apps.json")

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return ""
	}

	// Read and parse config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	var config GitHubCopilotConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return ""
	}

	// Extract token from any app entry (usually there's only one)
	for _, app := range config.Apps {
		if app.OAuthToken != "" {
			return app.OAuthToken
		}
	}

	return ""
}

type AIProvider struct {
	Client       *openai.Client
	ProviderType string // "openai" or "github_copilot"
}

func NewAIProvider(apiKey string) *AIProvider {
	config := openai.DefaultConfig(apiKey)
	config.HTTPClient = &http.Client{}
	return &AIProvider{
		Client:       openai.NewClientWithConfig(config),
		ProviderType: "openai",
	}
}

// GitHubCopilotTransport adds required headers for GitHub Copilot API
type GitHubCopilotTransport struct {
	Transport http.RoundTripper
}

func (t *GitHubCopilotTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Add required headers for GitHub Copilot
	req.Header.Set("Copilot-Integration-Id", "vscode-chat")
	req.Header.Set("Editor-Version", "vscode/1.85.0")
	req.Header.Set("Editor-Plugin-Version", "copilot-chat/0.11.1")
	req.Header.Set("User-Agent", "GitHubCopilotChat/0.11.1")

	// Use default transport if none specified
	transport := t.Transport
	if transport == nil {
		transport = http.DefaultTransport
	}

	return transport.RoundTrip(req)
}

// NewGitHubCopilotProvider creates an AI provider using GitHub Copilot API
func NewGitHubCopilotProvider(githubToken string) *AIProvider {
	// GitHub Copilot uses OpenAI-compatible chat completions API
	config := openai.DefaultConfig(githubToken)
	config.BaseURL = "https://api.githubcopilot.com"

	// Create custom HTTP client with required headers for GitHub Copilot
	client := &http.Client{
		Transport: &GitHubCopilotTransport{},
	}
	config.HTTPClient = client

	return &AIProvider{
		Client:       openai.NewClientWithConfig(config),
		ProviderType: "github_copilot",
	}
}

// testGitHubCopilotAccess tests if the GitHub Copilot API is accessible
func testGitHubCopilotAccess(token string) bool {
	client := &http.Client{
		Transport: &GitHubCopilotTransport{},
		Timeout:   5 * time.Second,
	}

	req, err := http.NewRequest("GET", "https://api.githubcopilot.com/models", nil)
	if err != nil {
		return false
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error but don't interrupt the flow
			fmt.Printf("Warning: failed to close response body: %v\n", closeErr)
		}
	}()

	return resp.StatusCode == 200
}

// NewAIProviderAuto automatically detects and creates the appropriate AI provider
func NewAIProviderAuto() *AIProvider {
	// Check for GitHub Copilot first (recommended)
	if githubToken := getGitHubCopilotToken(); githubToken != "" {
		// Use GitHub Copilot by default when token is available
		// Set USE_GITHUB_COPILOT=false to disable explicitly
		if useGitHubCopilot := os.Getenv("USE_GITHUB_COPILOT"); useGitHubCopilot != "false" {
			provider := NewGitHubCopilotProvider(githubToken)

			// Test if GitHub Copilot access works by trying a simple API call
			if testGitHubCopilotAccess(githubToken) {
				return provider
			}
			// If GitHub Copilot test fails, fall back to OpenAI
		}
	}

	// Fall back to OpenAI
	if openaiKey := os.Getenv("OPENAI_API_KEY"); openaiKey != "" {
		return NewAIProvider(openaiKey)
	}

	// No AI provider available
	return nil
}

func (ai *AIProvider) Ask(question string, contextStr string) (string, error) {
	resp, err := ai.Client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model:    ai.getModel(),
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

// DescribeResource provides AI-powered description and recommendations for a specific Azure resource
func (ai *AIProvider) DescribeResource(resourceType, resourceName, resourceDetails string) (string, error) {
	prompt := fmt.Sprintf("Analyze this Azure %s resource named '%s' and provide:\n1. Brief description of what it does\n2. Current configuration summary\n3. Optimization recommendations\n4. Security considerations\n\nResource details:\n%s",
		resourceType, resourceName, resourceDetails)
	return ai.Ask(prompt, "Azure Resource Analysis")
}

// getModel returns the appropriate model for the AI provider
func (ai *AIProvider) getModel() string {
	if ai.ProviderType == "github_copilot" {
		// GitHub Copilot: use gpt-4o-mini as it's available and efficient
		return "gpt-4o-mini"
	}

	// For OpenAI, check environment variable or use default
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4" // Default to GPT-4
	}
	return model
}

// AnalyzeMetrics provides AI-powered analysis of resource metrics and performance data
func (ai *AIProvider) AnalyzeMetrics(resourceName, metricsData string) (string, error) {
	prompt := fmt.Sprintf("Analyze these metrics for Azure resource '%s' and provide:\n1. Performance insights\n2. Trending analysis\n3. Alerting recommendations\n4. Scaling suggestions\n\nMetrics data:\n%s",
		resourceName, metricsData)
	return ai.Ask(prompt, "Azure Metrics Analysis")
}

// SuggestCostOptimizations analyzes resource configuration for cost savings
func (ai *AIProvider) SuggestCostOptimizations(resources []string, resourceDetails map[string]string) (string, error) {
	var resourceInfo strings.Builder
	for _, resource := range resources {
		if details, exists := resourceDetails[resource]; exists {
			resourceInfo.WriteString(fmt.Sprintf("Resource: %s\nDetails: %s\n\n", resource, details))
		}
	}

	prompt := fmt.Sprintf("Analyze these Azure resources for cost optimization opportunities:\n1. Right-sizing recommendations\n2. Reserved instance opportunities\n3. Unused resources to delete\n4. Alternative service options\n\n%s", resourceInfo.String())
	return ai.Ask(prompt, "Azure Cost Optimization")
}

// GenerateTerraformCode generates Terraform code for a resource based on requirements
func (ai *AIProvider) GenerateTerraformCode(resourceType, requirements string) (string, error) {
	prompt := fmt.Sprintf("Generate Terraform code for an Azure %s with these requirements:\n%s\n\nProvide:\n1. Complete .tf file content\n2. Required variables\n3. Output values\n4. Brief usage explanation",
		resourceType, requirements)
	return ai.Ask(prompt, "Terraform Code Generation")
}

// GenerateBicepCode generates Bicep code for a resource based on requirements
func (ai *AIProvider) GenerateBicepCode(resourceType, requirements string) (string, error) {
	prompt := fmt.Sprintf("Generate Bicep code for an Azure %s with these requirements:\n%s\n\nProvide:\n1. Complete .bicep file content\n2. Required parameters\n3. Output values\n4. Brief usage explanation",
		resourceType, requirements)
	return ai.Ask(prompt, "Bicep Code Generation")
}

// TroubleshootError analyzes error messages and provides troubleshooting guidance
func (ai *AIProvider) TroubleshootError(errorMessage, context string) (string, error) {
	prompt := fmt.Sprintf("Analyze this Azure error and provide troubleshooting steps:\nError: %s\nContext: %s\n\nProvide:\n1. Root cause analysis\n2. Step-by-step fix\n3. Prevention recommendations",
		errorMessage, context)
	return ai.Ask(prompt, "Azure Troubleshooting")
}
