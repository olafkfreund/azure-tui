package openai

import (
	"os"
	"testing"
)

// TestNewAIProviderAuto tests the AI provider auto-detection functionality
func TestNewAIProviderAuto(t *testing.T) {
	// Save original environment
	originalOpenAI := os.Getenv("OPENAI_API_KEY")
	originalGitHub := os.Getenv("GITHUB_TOKEN")
	originalUseCopilot := os.Getenv("USE_GITHUB_COPILOT")

	// Clean up after test
	defer func() {
		_ = os.Setenv("OPENAI_API_KEY", originalOpenAI)
		_ = os.Setenv("GITHUB_TOKEN", originalGitHub)
		_ = os.Setenv("USE_GITHUB_COPILOT", originalUseCopilot)
	}()

	t.Run("no provider configured", func(t *testing.T) {
		_ = os.Unsetenv("OPENAI_API_KEY")
		_ = os.Unsetenv("GITHUB_TOKEN")

		provider := NewAIProviderAuto()
		if provider != nil {
			t.Error("Expected nil provider when no API keys are configured")
		}
	})

	t.Run("openai provider configured", func(t *testing.T) {
		_ = os.Unsetenv("GITHUB_TOKEN")
		_ = os.Setenv("OPENAI_API_KEY", "test-key")
		_ = os.Setenv("USE_GITHUB_COPILOT", "false")

		provider := NewAIProviderAuto()
		if provider == nil {
			t.Error("Expected provider to be initialized with OpenAI key")
		} else if provider.ProviderType != "openai" {
			t.Errorf("Expected provider type 'openai', got '%s'", provider.ProviderType)
		}
	})

	t.Run("github copilot disabled", func(t *testing.T) {
		_ = os.Setenv("GITHUB_TOKEN", "test-github-token")
		_ = os.Setenv("OPENAI_API_KEY", "test-openai-key")
		_ = os.Setenv("USE_GITHUB_COPILOT", "false")

		provider := NewAIProviderAuto()
		if provider == nil {
			t.Error("Expected provider to be initialized")
		} else if provider.ProviderType != "openai" {
			t.Errorf("Expected provider type 'openai' when GitHub Copilot disabled, got '%s'", provider.ProviderType)
		}
	})
}

// TestGetGitHubCopilotToken tests the GitHub token retrieval
func TestGetGitHubCopilotToken(t *testing.T) {
	// Save original environment
	originalToken := os.Getenv("GITHUB_TOKEN")
	defer func() {
		_ = os.Setenv("GITHUB_TOKEN", originalToken)
	}()

	t.Run("token from environment", func(t *testing.T) {
		expectedToken := "test-github-token"
		_ = os.Setenv("GITHUB_TOKEN", expectedToken)

		token := getGitHubCopilotToken()
		if token != expectedToken {
			t.Errorf("Expected token '%s', got '%s'", expectedToken, token)
		}
	})

	t.Run("no token available", func(t *testing.T) {
		_ = os.Unsetenv("GITHUB_TOKEN")

		token := getGitHubCopilotToken()
		// Should return empty string when no token available and no config file
		if token != "" {
			t.Errorf("Expected empty token, got '%s'", token)
		}
	})
}
