package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type OpenAIConfig struct {
	Model   string
	APIKey  string
	BaseURL string // e.g. https://api.openai.com/v1
}

type Message struct {
	Role    string `json:"role"` // "system", "user", "assistant"
	Content string `json:"content"`
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream,omitempty"`
}

type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Usage interface{} `json:"usage"`
}

// CopilotAgent defines an agent/role for prompt context
type CopilotAgent struct {
	Name   string
	Role   string // e.g. "system", "user"
	Prompt string
}

// BuildPrompt builds a prompt with agent/role context
func BuildPrompt(agent CopilotAgent, userPrompt string) []Message {
	return []Message{
		{Role: agent.Role, Content: agent.Prompt},
		{Role: "user", Content: userPrompt},
	}
}

// SendPrompt sends a prompt to OpenAI and returns the response
func SendPrompt(cfg OpenAIConfig, messages []Message) (string, error) {
	if cfg.APIKey == "" {
		cfg.APIKey = os.Getenv("OPENAI_API_KEY")
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.openai.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-4"
	}
	body, _ := json.Marshal(ChatRequest{
		Model:    cfg.Model,
		Messages: messages,
	})
	req, err := http.NewRequest("POST", cfg.BaseURL+"/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("OpenAI error: %s", string(b))
	}
	var out ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("No response from OpenAI")
	}
	return out.Choices[0].Message.Content, nil
}

// Azure Copilot Agents for different scenarios

var AzureIaCAgent = CopilotAgent{
	Name:   "AzureIaC",
	Role:   "system",
	Prompt: "You are an expert Azure infrastructure-as-code assistant. Generate, review, and explain Terraform or Bicep code for Azure resources. Always ask clarifying questions before making changes. Output only valid code or clear, actionable steps.",
}

var AzureTroubleshootAgent = CopilotAgent{
	Name:   "AzureTroubleshooter",
	Role:   "system",
	Prompt: "You are an expert Azure cloud troubleshooter. Analyze error messages, deployment logs, and IaC output. Suggest clear, step-by-step fixes and always ask the user for missing context before giving advice.",
}

var AzureSecurityAgent = CopilotAgent{
	Name:   "AzureSecurity",
	Role:   "system",
	Prompt: "You are an Azure security best practices advisor. Review Terraform/Bicep code, Azure resource settings, and policies. Identify security risks, suggest remediations, and explain your reasoning in simple terms.",
}

var AzureCostAgent = CopilotAgent{
	Name:   "AzureCost",
	Role:   "system",
	Prompt: "You are an Azure cost optimization expert. Review IaC code, resource lists, and usage data. Suggest ways to reduce costs, right-size resources, and use reserved instances or savings plans. Always ask for business context before making recommendations.",
}

var AzureDocAgent = CopilotAgent{
	Name:   "AzureDoc",
	Role:   "system",
	Prompt: "You are an Azure documentation generator. Given Terraform/Bicep code or resource details, generate clear, concise documentation and usage examples for end users.",
}

var AzureCLIHelpAgent = CopilotAgent{
	Name:   "AzureCLIHelp",
	Role:   "system",
	Prompt: "You are an Azure CLI and PowerShell expert. Given a user goal, generate the minimal, correct Azure CLI or PowerShell commands to accomplish it. Always explain any required parameters or authentication steps.",
}

// Helper to select agent by scenario
func GetAzureAgent(scenario string) CopilotAgent {
	switch scenario {
	case "iac":
		return AzureIaCAgent
	case "troubleshoot":
		return AzureTroubleshootAgent
	case "security":
		return AzureSecurityAgent
	case "cost":
		return AzureCostAgent
	case "doc":
		return AzureDocAgent
	case "cli":
		return AzureCLIHelpAgent
	default:
		return AzureIaCAgent
	}
}

// CopilotConfig supports GitHub Copilot API and agent config
type CopilotConfig struct {
	APIBase string // e.g. https://api.githubcopilot.com
	APIKey  string // Copilot oauth_token
	Model   string // e.g. openai/gpt-4o
	Agent   CopilotAgent
}

// GetCopilotConfig loads config from env or ~/.config/github-copilot/apps.json
func GetCopilotConfig() CopilotConfig {
	cfg := CopilotConfig{
		APIBase: os.Getenv("OPENAI_API_BASE"),
		APIKey:  os.Getenv("OPENAI_API_KEY"),
		Model:   os.Getenv("OPENAI_MODEL"),
		Agent:   AzureIaCAgent,
	}
	if cfg.APIBase == "" {
		cfg.APIBase = "https://api.githubcopilot.com"
	}
	if cfg.Model == "" {
		cfg.Model = "openai/gpt-4o"
	}
	return cfg
}

// ListCopilotModels queries the Copilot API for available models
func ListCopilotModels(cfg CopilotConfig) ([]string, error) {
	req, err := http.NewRequest("GET", cfg.APIBase+"/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Copilot-Integration-Id", "vscode-chat")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("Copilot error: %s", string(b))
	}
	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	var models []string
	for _, m := range out.Data {
		models = append(models, m.ID)
	}
	return models, nil
}

// SendCopilotPrompt sends a prompt to Copilot API with agent/role context
func SendCopilotPrompt(cfg CopilotConfig, userPrompt string) (string, error) {
	messages := BuildPrompt(cfg.Agent, userPrompt)
	body, _ := json.Marshal(ChatRequest{
		Model:    cfg.Model,
		Messages: messages,
	})
	req, err := http.NewRequest("POST", cfg.APIBase+"/chat/completions", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Copilot-Integration-Id", "vscode-chat")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("Copilot error: %s", string(b))
	}
	var out ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("No response from Copilot")
	}
	return out.Choices[0].Message.Content, nil
}

// Usage:
//   cfg := GetCopilotConfig()
//   models, _ := ListCopilotModels(cfg)
//   resp, _ := SendCopilotPrompt(cfg, "Generate a Terraform VM resource for Azure.")
//
//   You can define multiple CopilotAgent structs for different roles/scenarios.
//   The API key is your Copilot oauth_token (see aider docs for how to obtain).
//   Model selection and agent/role are fully configurable.
