# GitHub Copilot Integration Guide

## Overview

The Azure TUI application now supports **GitHub Copilot** as the preferred AI provider, in addition to OpenAI. This provides enhanced Azure-specific analysis and recommendations using GitHub's AI capabilities.

## Setup Options

### Option 1: GitHub Copilot (Recommended)

```bash
# Set your GitHub token
export GITHUB_TOKEN="your-github-token"

# Optional: Explicitly enable GitHub Copilot mode
export USE_GITHUB_COPILOT="true"
```

**Benefits:**
- 🚀 Enhanced Azure-specific knowledge
- 🔒 Better integration with development workflows
- 📊 Optimized for cloud infrastructure analysis

### Option 2: OpenAI API (Alternative)

```bash
# Set your OpenAI API key
export OPENAI_API_KEY="your-openai-api-key"
```

## How It Works

The application automatically detects which AI provider to use:

1. **GitHub Copilot Priority**: If `GITHUB_TOKEN` is set, it uses GitHub Copilot API
2. **OpenAI Fallback**: If only `OPENAI_API_KEY` is set, it uses OpenAI
3. **Auto-Detection**: The app intelligently switches between providers based on available credentials

## Features Enabled

All AI-powered features work with both providers:

- **Resource Analysis** (`a` key): Get intelligent insights about Azure resources
- **Metrics Dashboard** (`M` key): AI-powered performance analysis
- **Cost Optimization** (`O` key): Smart cost-saving recommendations
- **IaC Generation** (`T`/`B` keys): Generate Terraform/Bicep code
- **Troubleshooting**: AI-powered error analysis and solutions

## Technical Implementation

### Code Changes Made

1. **Enhanced AI Provider** (`internal/openai/ai.go`):
   - Added GitHub Copilot API endpoint support
   - Automatic provider detection
   - Model selection based on provider

2. **Updated Initialization** (`cmd/main.go`):
   - Support for both GitHub token and OpenAI key
   - Graceful fallback between providers

3. **Documentation Updates**:
   - README.md: Added GitHub Copilot setup instructions
   - FEATURES.md: Updated AI integration section

### Configuration Details

The AI provider configuration follows this logic:

```go
// Priority order:
1. If USE_GITHUB_COPILOT="true" AND GITHUB_TOKEN is set → Use GitHub Copilot
2. If GITHUB_TOKEN is set (regardless of OpenAI key) → Use GitHub Copilot  
3. If only OPENAI_API_KEY is set → Use OpenAI
4. If neither is set → AI features disabled
```

## Testing

To verify the integration is working:

1. **Set Environment Variables**:
   ```bash
   export GITHUB_TOKEN="your-token"
   export USE_GITHUB_COPILOT="true"
   ```

2. **Run the Application**:
   ```bash
   go run cmd/main.go
   ```

3. **Test AI Features**:
   - Navigate to a resource group
   - Press `a` for AI analysis
   - Check that responses are generated

## Troubleshooting

### Common Issues

1. **"AI provider not configured"**:
   - Ensure either `GITHUB_TOKEN` or `OPENAI_API_KEY` is set
   - Check token permissions and validity

2. **"AI analysis failed"**:
   - Verify GitHub token has appropriate permissions
   - Check network connectivity
   - Try with OpenAI as fallback

3. **No AI responses**:
   - Confirm the token is valid and not expired
   - Check if GitHub Copilot access is enabled for your account

### Debug Mode

You can verify which provider is being used by checking the logs or adding debug output to the AI provider initialization.

## Migration from OpenAI

If you're currently using OpenAI and want to switch to GitHub Copilot:

1. **Keep existing setup** (works as-is)
2. **Add GitHub token**:
   ```bash
   export GITHUB_TOKEN="your-github-token"
   ```
3. **Remove OpenAI key** (optional):
   ```bash
   unset OPENAI_API_KEY
   ```

The application will automatically prefer GitHub Copilot when both are available.

## Benefits of GitHub Copilot Integration

- **Azure-Specific Training**: Better understanding of Azure services and patterns
- **Developer-Friendly**: Integrates with existing GitHub workflows
- **Enhanced Code Generation**: Superior Terraform/Bicep code generation
- **Cost Efficiency**: Often included with GitHub subscriptions
- **Better Context**: Understands cloud infrastructure patterns and best practices

---

**Status**: ✅ GitHub Copilot integration is complete and ready for use!
