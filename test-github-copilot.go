package main

package main

import (
	"fmt"
	"os"

	"github.com/olafkfreund/azure-tui/internal/openai"
)

func main() {
	fmt.Println("=== Testing GitHub Copilot Integration ===")

	// Test 1: Check environment variables
	fmt.Println("\n1. Environment Variables:")
	apiKey := os.Getenv("OPENAI_API_KEY")
	githubToken := os.Getenv("GITHUB_TOKEN")
	useGitHub := os.Getenv("USE_GITHUB_COPILOT")

	fmt.Printf("   OPENAI_API_KEY: %s\n", maskKey(apiKey))
	fmt.Printf("   GITHUB_TOKEN: %s\n", maskKey(githubToken))
	fmt.Printf("   USE_GITHUB_COPILOT: %s\n", useGitHub)

	// Test 2: Initialize AI Provider
	fmt.Println("\n2. AI Provider Initialization:")
	if apiKey == "" && githubToken == "" {
		fmt.Println("   ❌ No API keys found. Set either OPENAI_API_KEY or GITHUB_TOKEN")
		fmt.Println("\n   To use GitHub Copilot:")
		fmt.Println("   export GITHUB_TOKEN=\"your-github-token\"")
		fmt.Println("   export USE_GITHUB_COPILOT=\"true\"")
		fmt.Println("\n   To use OpenAI:")
		fmt.Println("   export OPENAI_API_KEY=\"your-openai-api-key\"")
		return
	}

	aiProvider := openai.NewAIProvider(apiKey)
	if aiProvider != nil {
		fmt.Println("   ✅ AI Provider initialized successfully")
		
		// Determine which provider is being used
		if githubToken != "" && (useGitHub == "true" || apiKey == "") {
			fmt.Println("   🚀 Using GitHub Copilot API")
		} else {
			fmt.Println("   🤖 Using OpenAI API")
		}
	} else {
		fmt.Println("   ❌ Failed to initialize AI Provider")
		return
	}

	// Test 3: Simple AI request (if token is available)
	fmt.Println("\n3. Testing AI Request:")
	if apiKey != "" || githubToken != "" {
		response, err := aiProvider.Ask("What is Azure?", "Azure Cloud")
		if err != nil {
			fmt.Printf("   ❌ AI request failed: %v\n", err)
			fmt.Println("   This might be normal if the token doesn't have proper permissions")
		} else {
			fmt.Printf("   ✅ AI response received (length: %d characters)\n", len(response))
			// Don't print the full response to keep test output clean
			if len(response) > 100 {
				fmt.Printf("   Preview: %s...\n", response[:100])
			} else {
				fmt.Printf("   Response: %s\n", response)
			}
		}
	}

	fmt.Println("\n=== Test Complete ===")
}

// maskKey partially masks sensitive keys for display
func maskKey(key string) string {
	if key == "" {
		return "(not set)"
	}
	if len(key) <= 8 {
		return "***masked***"
	}
	return key[:4] + "***" + key[len(key)-4:]
}
