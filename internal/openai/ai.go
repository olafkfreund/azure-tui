package openai

import (
	"context"
	"fmt"
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

// DescribeResource provides AI-powered description and recommendations for a specific Azure resource
func (ai *AIProvider) DescribeResource(resourceType, resourceName, resourceDetails string) (string, error) {
	prompt := fmt.Sprintf("Analyze this Azure %s resource named '%s' and provide:\n1. Brief description of what it does\n2. Current configuration summary\n3. Optimization recommendations\n4. Security considerations\n\nResource details:\n%s",
		resourceType, resourceName, resourceDetails)
	return ai.Ask(prompt, "Azure Resource Analysis")
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
