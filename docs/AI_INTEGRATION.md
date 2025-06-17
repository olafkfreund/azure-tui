# AI Integration

## Overview

Azure TUI integrates with OpenAI to provide intelligent Azure resource analysis, code generation, and optimization recommendations.

## Setup

Set your OpenAI API key as an environment variable:

```bash
export OPENAI_API_KEY="your-openai-api-key-here"
```

Optionally, you can specify which OpenAI model to use:

```bash
export OPENAI_MODEL="gpt-4"  # Default: gpt-4
```

## Features

All AI-powered features work when an OpenAI API key is configured:

- **Resource Analysis** (`a` key): Get intelligent insights about Azure resources
- **Metrics Dashboard** (`M` key): AI-powered performance analysis  
- **Cost Optimization** (`O` key): Smart cost-saving recommendations
- **IaC Generation** (`T`/`B` keys): Generate Terraform/Bicep code
- **Troubleshooting**: AI-powered error analysis and solutions

## Usage

1. **Navigate** to any Azure resource using the tree view
2. **Press the appropriate key** for the AI feature you want to use:
   - `a` - Analyze selected resource
   - `M` - View metrics dashboard with AI insights
   - `O` - Get cost optimization recommendations  
   - `T` - Generate Terraform code for the resource
   - `B` - Generate Bicep code for the resource
3. **View the AI response** in the popup dialog
4. **Press Esc** to close the AI dialog

## Technical Details

The AI provider uses the official OpenAI Go SDK with the following configuration:

- **Default Model**: GPT-4
- **API Endpoint**: https://api.openai.com/v1
- **Authentication**: Bearer token using your API key
- **Timeout**: Standard HTTP client timeout

## Error Handling

If no API key is configured:
- AI features will show "AI provider not configured" message
- All other application features continue to work normally
- You can set the API key and restart the application to enable AI features

## Privacy

- All API calls are made directly to OpenAI
- No data is stored locally or sent to third parties
- Azure resource data is only sent to OpenAI for analysis when you explicitly trigger AI features
