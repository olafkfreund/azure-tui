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

// GenerateTerraformCode generates Terraform code using AI with enhanced capabilities
func (ai *AIProvider) GenerateTerraformCode(req TerraformRequest) (*TerraformResponse, error) {
	if ai.Client == nil {
		return nil, fmt.Errorf("AI client not initialized")
	}

	prompt := ai.buildTerraformPrompt(req)

	chatReq := openai.ChatCompletionRequest{
		Model: ai.getModel(),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: ai.getTerraformSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.1, // Low temperature for more consistent code generation
		TopP:        0.95,
	}

	resp, err := ai.Client.CreateChatCompletion(context.Background(), chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Terraform code: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI provider")
	}

	content := resp.Choices[0].Message.Content
	return ai.parseTerraformResponse(content), nil
}

// GenerateTerraformCodeSimple generates simple Terraform code (legacy function for compatibility)
func (ai *AIProvider) GenerateTerraformCodeSimple(resourceType, requirements string) (string, error) {
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

// TerraformRequest represents a request for Terraform code generation
type TerraformRequest struct {
	ResourceType     string            `json:"resource_type"`
	Description      string            `json:"description"`
	Requirements     []string          `json:"requirements"`
	Location         string            `json:"location"`
	Environment      string            `json:"environment"`
	Tags             map[string]string `json:"tags"`
	ExistingCode     string            `json:"existing_code,omitempty"`
	ModificationType string            `json:"modification_type,omitempty"` // "create", "modify", "optimize"
}

// TerraformResponse represents the response from AI for Terraform operations
type TerraformResponse struct {
	Code          string   `json:"code"`
	Explanation   string   `json:"explanation"`
	Variables     []string `json:"variables"`
	Outputs       []string `json:"outputs"`
	BestPractices []string `json:"best_practices"`
}

// GenerateTerraformCodeAdvanced generates Terraform code using AI with enhanced capabilities
func (ai *AIProvider) GenerateTerraformCodeAdvanced(req TerraformRequest) (*TerraformResponse, error) {
	if ai.Client == nil {
		return nil, fmt.Errorf("AI client not initialized")
	}

	prompt := ai.buildTerraformPrompt(req)

	chatReq := openai.ChatCompletionRequest{
		Model: ai.getModel(),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: ai.getTerraformSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.1, // Low temperature for more consistent code generation
		TopP:        0.95,
	}

	resp, err := ai.Client.CreateChatCompletion(context.Background(), chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate Terraform code: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI provider")
	}

	content := resp.Choices[0].Message.Content
	return ai.parseTerraformResponse(content), nil
}

// OptimizeTerraformCode optimizes existing Terraform code
func (ai *AIProvider) OptimizeTerraformCode(code, focus string) (*TerraformResponse, error) {
	if ai.Client == nil {
		return nil, fmt.Errorf("AI client not initialized")
	}

	prompt := fmt.Sprintf("Please optimize the following Terraform code with focus on: %s\n\nCurrent code:\n```hcl\n%s\n```\n\nPlease provide:\n1. Optimized code\n2. Explanation of changes\n3. Best practices applied\n4. Any new variables or outputs needed", focus, code)

	chatReq := openai.ChatCompletionRequest{
		Model: ai.getModel(),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: ai.getTerraformSystemPrompt(),
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   2000,
		Temperature: 0.1,
		TopP:        0.95,
	}

	resp, err := ai.Client.CreateChatCompletion(context.Background(), chatReq)
	if err != nil {
		return nil, fmt.Errorf("failed to optimize Terraform code: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from AI provider")
	}

	content := resp.Choices[0].Message.Content
	return ai.parseTerraformResponse(content), nil
}

// ExplainTerraformCode explains existing Terraform code
func (ai *AIProvider) ExplainTerraformCode(code string) (string, error) {
	if ai.Client == nil {
		return "", fmt.Errorf("AI client not initialized")
	}

	prompt := fmt.Sprintf("Please explain the following Terraform code in simple terms:\n\n```hcl\n%s\n```\n\nProvide a clear explanation of:\n1. What resources are being created\n2. How they are configured\n3. Any dependencies between resources\n4. Best practices being followed or that could be improved", code)

	chatReq := openai.ChatCompletionRequest{
		Model: ai.getModel(),
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a helpful Terraform expert. Explain code clearly and concisely.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		MaxTokens:   1000,
		Temperature: 0.2,
		TopP:        0.95,
	}

	resp, err := ai.Client.CreateChatCompletion(context.Background(), chatReq)
	if err != nil {
		return "", fmt.Errorf("failed to explain Terraform code: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from AI provider")
	}

	return resp.Choices[0].Message.Content, nil
}

// buildTerraformPrompt builds a detailed prompt for Terraform code generation
func (ai *AIProvider) buildTerraformPrompt(req TerraformRequest) string {
	var prompt strings.Builder

	switch req.ModificationType {
	case "modify":
		prompt.WriteString(fmt.Sprintf("Please modify the following Terraform code for %s:\n\n", req.ResourceType))
		prompt.WriteString(fmt.Sprintf("Existing code:\n```hcl\n%s\n```\n\n", req.ExistingCode))
		prompt.WriteString(fmt.Sprintf("Modifications needed: %s\n\n", req.Description))
	case "optimize":
		prompt.WriteString(fmt.Sprintf("Please optimize the following Terraform code for %s:\n\n", req.ResourceType))
		prompt.WriteString(fmt.Sprintf("Current code:\n```hcl\n%s\n```\n\n", req.ExistingCode))
		prompt.WriteString("Focus on best practices, security, and efficiency.\n\n")
	default: // create
		prompt.WriteString(fmt.Sprintf("Please generate Terraform code for: %s\n\n", req.ResourceType))
		prompt.WriteString(fmt.Sprintf("Description: %s\n\n", req.Description))
	}

	prompt.WriteString("Requirements:\n")
	for _, req := range req.Requirements {
		prompt.WriteString(fmt.Sprintf("- %s\n", req))
	}

	prompt.WriteString(fmt.Sprintf("\nConfiguration:\n"))
	prompt.WriteString(fmt.Sprintf("- Location: %s\n", req.Location))
	prompt.WriteString(fmt.Sprintf("- Environment: %s\n", req.Environment))

	if len(req.Tags) > 0 {
		prompt.WriteString("- Tags:\n")
		for key, value := range req.Tags {
			prompt.WriteString(fmt.Sprintf("  - %s: %s\n", key, value))
		}
	}

	prompt.WriteString("\nPlease provide:\n")
	prompt.WriteString("1. Complete Terraform code\n")
	prompt.WriteString("2. Brief explanation of the resources\n")
	prompt.WriteString("3. List of variables needed\n")
	prompt.WriteString("4. List of outputs provided\n")
	prompt.WriteString("5. Best practices applied\n\n")
	prompt.WriteString("Use Azure best practices and follow Terraform conventions.")

	return prompt.String()
}

// getTerraformSystemPrompt returns the system prompt for Terraform operations
func (ai *AIProvider) getTerraformSystemPrompt() string {
	return `You are an expert Terraform developer specializing in Azure infrastructure. 

Your expertise includes:
- Azure resource types and their configurations
- Terraform best practices and conventions
- Security and compliance standards
- Cost optimization strategies
- Infrastructure as Code patterns

When generating Terraform code:
1. Use the latest Azure provider syntax
2. Follow naming conventions
3. Include appropriate tags
4. Use variables for configurable values
5. Provide meaningful outputs
6. Include security best practices
7. Use data sources when appropriate
8. Add comments for complex configurations

Always generate valid, working Terraform code that follows Azure and Terraform best practices.`
}

// parseTerraformResponse parses the AI response and extracts structured information
func (ai *AIProvider) parseTerraformResponse(content string) *TerraformResponse {
	response := &TerraformResponse{
		Variables:     []string{},
		Outputs:       []string{},
		BestPractices: []string{},
	}

	// Extract code blocks (look for ```hcl or ```terraform)
	lines := strings.Split(content, "\n")
	var codeLines []string
	inCodeBlock := false

	for _, line := range lines {
		if strings.HasPrefix(line, "```") {
			if strings.Contains(line, "hcl") || strings.Contains(line, "terraform") {
				inCodeBlock = true
				continue
			} else if inCodeBlock {
				inCodeBlock = false
				continue
			}
		}

		if inCodeBlock {
			codeLines = append(codeLines, line)
		}
	}

	if len(codeLines) > 0 {
		response.Code = strings.Join(codeLines, "\n")
	} else {
		// If no code block found, use the entire response as code
		response.Code = content
	}

	// Extract explanation (everything that's not code)
	response.Explanation = content

	// Parse variables and outputs from the code
	response.Variables = ai.extractVariablesFromCode(response.Code)
	response.Outputs = ai.extractOutputsFromCode(response.Code)

	return response
}

// extractVariablesFromCode extracts variable names from Terraform code
func (ai *AIProvider) extractVariablesFromCode(code string) []string {
	variables := []string{}
	lines := strings.Split(code, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "variable ") {
			// Extract variable name from: variable "name" {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				varName := strings.Trim(parts[1], `"`)
				variables = append(variables, varName)
			}
		}
	}

	return variables
}

// extractOutputsFromCode extracts output names from Terraform code
func (ai *AIProvider) extractOutputsFromCode(code string) []string {
	outputs := []string{}
	lines := strings.Split(code, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "output ") {
			// Extract output name from: output "name" {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				outputName := strings.Trim(parts[1], `"`)
				outputs = append(outputs, outputName)
			}
		}
	}

	return outputs
}
