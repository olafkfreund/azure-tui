# AI Integration Workflow Guide 🤖

## Overview
This guide provides comprehensive instructions for integrating AI capabilities into Azure TUI, including setup, configuration, and advanced usage patterns.

---

## Table of Contents
1. [AI Setup & Configuration](#ai-setup--configuration)
2. [Supported AI Providers](#supported-ai-providers)
3. [Feature Workflows](#feature-workflows)
4. [Advanced AI Usage](#advanced-ai-usage)
5. [Custom AI Agents](#custom-ai-agents)
6. [Troubleshooting AI Issues](#troubleshooting-ai-issues)

---

## AI Setup & Configuration

### OpenAI Setup (Recommended)
```bash
# Set your OpenAI API key
export AZURE_TUI_AI_API_KEY="sk-your-openai-api-key"

# Optional: Specify model
export AZURE_TUI_AI_MODEL="gpt-4"

# Optional: Custom endpoint
export AZURE_TUI_AI_ENDPOINT="https://api.openai.com/v1"
```

### Configuration File
Create `~/.config/aztui/config.yaml`:
```yaml
ai:
  provider: "openai"
  api_key: "sk-your-key-here"
  model: "gpt-4"
  endpoint: "https://api.openai.com/v1"
  timeout: 30s
  max_tokens: 4096
  temperature: 0.1
  
  # Feature-specific settings
  features:
    resource_analysis:
      enabled: true
      system_prompt: "You are an Azure expert analyzing resources for optimization."
    
    terraform_generation:
      enabled: true
      include_variables: true
      include_outputs: true
    
    cost_optimization:
      enabled: true
      currency: "USD"
      region: "eastus"
```

### Azure OpenAI Setup
```yaml
ai:
  provider: "azure_openai"
  azure_endpoint: "https://your-resource.openai.azure.com/"
  api_key: "your-azure-openai-key"
  deployment_name: "gpt-4"
  api_version: "2024-02-15-preview"
```

---

## Supported AI Providers

### 1. OpenAI
- **Models**: GPT-4, GPT-4 Turbo, GPT-3.5 Turbo
- **Features**: All AI features supported
- **Setup**: API key required

### 2. Azure OpenAI
- **Models**: GPT-4, GPT-3.5 Turbo (deployed models)
- **Features**: All AI features supported
- **Setup**: Azure OpenAI resource required

### 3. Local Models (Planned)
- **Ollama Integration**: Local model support
- **Models**: Llama 2, Code Llama, Mistral
- **Features**: Limited by model capabilities

---

## Feature Workflows

### 1. Resource Analysis Workflow

#### Step-by-Step Process:
1. **Select Resource**: Navigate to any Azure resource in the tree view
2. **Trigger Analysis**: Press `a` key
3. **AI Processing**: The system:
   - Gathers resource metadata
   - Analyzes configuration
   - Reviews health status
   - Compares against best practices
4. **Review Results**: AI provides comprehensive analysis

#### What Gets Analyzed:
```yaml
Resource Context:
  - Resource type and configuration
  - Current health and performance metrics
  - Cost data and trends
  - Security configuration
  - Compliance status
  - Dependencies and relationships

Analysis Output:
  - Health assessment
  - Performance optimization recommendations
  - Cost reduction opportunities
  - Security improvements
  - Best practices compliance
  - Action items with priorities
```

#### Example Analysis Prompt:
```
Analyze this Azure resource for optimization opportunities:

Resource Type: Virtual Machine
Name: webapp-vm-01
Location: East US
Size: Standard_D4s_v3
OS: Ubuntu 20.04
Current Metrics:
  CPU: 12% average utilization
  Memory: 45% average utilization
  Network: <1 Mbps throughput
Cost: $142/month

Provide recommendations for:
1. Cost optimization
2. Performance improvements
3. Security enhancements
4. Best practices compliance
```

### 2. Infrastructure as Code Generation

#### Terraform Generation Workflow:
1. **Select Resource**: Choose target Azure resource
2. **Press `T`**: Trigger Terraform generation
3. **AI Processing**: 
   - Analyzes resource configuration
   - Generates complete Terraform code
   - Includes provider requirements
   - Adds variables and outputs
4. **Review Generated Code**: Complete `.tf` file ready for use

#### Generated Terraform Structure:
```hcl
# Provider configuration
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

# Variables
variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
}

# Main resource
resource "azurerm_virtual_machine" "webapp_vm" {
  name                = var.vm_name
  location            = var.location
  resource_group_name = var.resource_group_name
  
  # ... complete configuration
}

# Outputs
output "vm_id" {
  description = "ID of the virtual machine"
  value       = azurerm_virtual_machine.webapp_vm.id
}
```

#### Bicep Generation Workflow:
1. **Select Resource**: Choose target Azure resource
2. **Press `B`**: Trigger Bicep generation
3. **AI Processing**: Generates ARM Bicep template
4. **Review Template**: Complete `.bicep` file

### 3. Cost Optimization Workflow

#### Automated Cost Analysis:
1. **Press `O`**: Trigger cost optimization analysis
2. **AI Processing**:
   - Reviews resource utilization
   - Analyzes cost patterns
   - Identifies optimization opportunities
   - Calculates potential savings
3. **Recommendations**: Specific cost reduction steps

#### Cost Analysis Output:
```markdown
💰 Cost Optimization Analysis

Current Monthly Cost: $847.32

🎯 Top Opportunities:
1. VM Rightsizing - Save $156/month (18%)
   • webapp-vm-01: D4s → D2s (-$89/month)
   • db-vm-02: D8s → D4s (-$67/month)

2. Storage Optimization - Save $67/month (8%)
   • Convert Premium SSD to Standard SSD for dev storage
   • Archive old backup data to Cool storage

3. Network Optimization - Save $34/month (4%)
   • Remove unused Load Balancer
   • Optimize bandwidth allocation

Total Potential Savings: $257/month (30% reduction)
```

---

## Advanced AI Usage

### Multi-Turn Conversations
The AI system supports context-aware conversations:

```yaml
Conversation Flow:
1. Initial Analysis: "Analyze this VM for cost optimization"
2. Follow-up: "What about security improvements?"
3. Deep Dive: "Show me the Terraform for the optimized version"
4. Implementation: "What's the migration strategy?"
```

### Custom System Prompts
Configure specialized AI behavior:

```yaml
ai:
  custom_prompts:
    security_expert:
      system: "You are a cybersecurity expert specializing in Azure security."
      temperature: 0.1
      
    cost_optimizer:
      system: "You are a cloud cost optimization specialist."
      focus: ["cost_reduction", "resource_utilization", "pricing_models"]
      
    devops_engineer:
      system: "You are a DevOps engineer focused on automation and IaC."
      tools: ["terraform", "bicep", "azure_cli", "powershell"]
```

### Batch Processing
Process multiple resources with AI:

```bash
# Analyze all VMs in resource group
aztui --ai-batch-analyze --resource-type="Microsoft.Compute/virtualMachines" \
      --resource-group="production-rg" \
      --output="analysis-report.json"

# Generate Terraform for entire resource group
aztui --ai-terraform-batch --resource-group="infrastructure-rg" \
      --output-dir="./terraform/"
```

---

## Custom AI Agents

### Creating Specialized Agents

#### Security Audit Agent
```yaml
agents:
  security_auditor:
    name: "Azure Security Auditor"
    description: "Specialized in Azure security compliance and best practices"
    system_prompt: |
      You are an Azure security expert conducting security audits.
      Focus on:
      - Azure Security Center recommendations
      - Network security configurations
      - Identity and access management
      - Data encryption and protection
      - Compliance frameworks (SOC, ISO, GDPR)
    
    skills:
      - security_assessment
      - compliance_checking
      - vulnerability_analysis
      - remediation_planning
    
    parameters:
      temperature: 0.1
      max_tokens: 2048
```

#### Cost Optimization Agent
```yaml
agents:
  cost_optimizer:
    name: "Azure Cost Optimizer"
    description: "Focuses on cost reduction and resource optimization"
    system_prompt: |
      You are a cloud cost optimization specialist.
      Analyze resources for:
      - Underutilized resources
      - Right-sizing opportunities
      - Reserved instance benefits
      - Storage optimization
      - Network cost reduction
    
    skills:
      - utilization_analysis
      - pricing_optimization
      - reservation_planning
      - storage_tiering
```

### Agent Usage
```bash
# Use specific agent for analysis
aztui --agent="security_auditor" --analyze-resource="vm-prod-01"

# Switch agents during session
:agent cost_optimizer
:analyze current_resource
```

---

## AI Integration Patterns

### 1. Pipeline Integration
Integrate AI analysis into CI/CD pipelines:

```yaml
# GitHub Actions example
- name: Azure Resource Analysis
  run: |
    aztui --ai-analyze --subscription="$AZURE_SUBSCRIPTION" \
          --output-format="json" > analysis.json
    
    # Process results
    cat analysis.json | jq '.recommendations[] | select(.priority == "high")'
```

### 2. Automated Reporting
Generate regular AI-powered reports:

```bash
#!/bin/bash
# Weekly optimization report
aztui --ai-batch-optimize \
      --subscription="production" \
      --format="markdown" \
      --output="weekly-optimization-$(date +%Y%m%d).md"
```

### 3. Infrastructure Validation
Use AI to validate infrastructure changes:

```bash
# Before deployment
aztui --ai-validate-terraform --file="main.tf" \
      --check-security --check-cost --check-compliance
```

---

## Troubleshooting AI Issues

### Common Issues

#### 1. API Key Problems
```bash
# Test API key
curl -H "Authorization: Bearer $AZURE_TUI_AI_API_KEY" \
     https://api.openai.com/v1/models

# Check configuration
aztui --check-ai-config
```

#### 2. Rate Limiting
```yaml
ai:
  rate_limiting:
    requests_per_minute: 60
    retry_attempts: 3
    backoff_multiplier: 2
```

#### 3. Large Resource Analysis
For complex resources, use streaming responses:
```yaml
ai:
  streaming: true
  chunk_size: 1024
  timeout: 120s
```

### Debug Mode
Enable AI debug logging:
```bash
export AZURE_TUI_AI_DEBUG="true"
aztui --verbose
```

### Fallback Strategies
Configure fallback behavior:
```yaml
ai:
  fallback:
    on_error: "show_basic_info"
    offline_mode: "demo_analysis"
    timeout_action: "partial_results"
```

---

## Performance Optimization

### Caching AI Responses
```yaml
ai:
  cache:
    enabled: true
    ttl: 3600s
    storage: "~/.cache/aztui/ai/"
    max_size: "100MB"
```

### Parallel Processing
```yaml
ai:
  parallel:
    max_concurrent: 5
    batch_size: 10
    timeout_per_request: 30s
```

---

This comprehensive guide covers all aspects of AI integration in Azure TUI. For additional support, consult the troubleshooting section or check the GitHub issues.
