# Configuration Guide 🔧

## Overview
Complete guide for configuring Azure TUI with all available options, environment variables, and customization settings.

---

## Table of Contents
1. [Configuration File Structure](#configuration-file-structure)
2. [Environment Variables](#environment-variables)
3. [Theme Customization](#theme-customization)
4. [Azure Integration](#azure-integration)
5. [AI Configuration](#ai-configuration)
6. [Performance Settings](#performance-settings)
7. [Security Configuration](#security-configuration)

---

## Configuration File Structure

### Default Configuration Location
- **Linux/macOS**: `~/.config/aztui/config.yaml`
- **Windows**: `%APPDATA%/aztui/config.yaml`
- **Custom**: Use `--config` flag or `AZURE_TUI_CONFIG` environment variable

### Complete Configuration Template
```yaml
# Azure TUI Configuration File
# Version: 2.0

# ===== UI Configuration =====
ui:
  theme: "azure"                    # azure, dark, light, custom
  interface_mode: "tree"            # tree, traditional
  auto_refresh: true                # Enable automatic refresh
  refresh_interval: 30s             # Refresh interval
  mouse_support: true               # Enable mouse interactions
  unicode_support: true             # Enable Unicode characters
  
  # Panel configuration
  panels:
    tree_width: 40                  # Tree panel width (percentage)
    content_width: 60               # Content panel width (percentage)
    status_height: 3                # Status bar height (lines)
  
  # Display options
  display:
    show_icons: true                # Show resource type icons
    show_health_status: true        # Show health indicators
    show_resource_count: true       # Show resource counts
    truncate_long_names: true       # Truncate long resource names
    max_name_length: 50             # Maximum name length before truncation

# ===== Azure Configuration =====
azure:
  # Authentication
  cli_path: "/usr/bin/az"           # Azure CLI path
  default_subscription: ""          # Default subscription ID
  default_tenant: ""                # Default tenant ID
  
  # Timeouts and limits
  timeout: 30s                      # API timeout
  max_retries: 3                    # Maximum retry attempts
  retry_delay: 1s                   # Delay between retries
  
  # Resource loading
  load_resources_on_expand: true    # Load resources when expanding groups
  cache_duration: 300s              # Resource cache duration
  max_resources_per_group: 1000     # Limit resources per group
  
  # SDK Configuration
  sdk:
    enabled: true                   # Enable Azure SDK (fallback to CLI)
    client_id: ""                   # Service principal client ID
    client_secret: ""               # Service principal secret
    certificate_path: ""            # Certificate path for auth

# ===== AI Configuration =====
ai:
  # Provider settings
  provider: "openai"                # openai, azure_openai, local
  api_key: ""                       # API key (or use environment variable)
  model: "gpt-4"                    # Model name
  endpoint: "https://api.openai.com/v1"  # API endpoint
  
  # Request settings
  timeout: 60s                      # AI request timeout
  max_tokens: 4096                  # Maximum tokens per request
  temperature: 0.1                  # Response creativity (0.0-1.0)
  top_p: 1.0                        # Nucleus sampling
  
  # Feature configuration
  features:
    resource_analysis: true         # Enable AI resource analysis
    terraform_generation: true     # Enable Terraform generation
    bicep_generation: true          # Enable Bicep generation
    cost_optimization: true         # Enable cost analysis
    security_analysis: true         # Enable security analysis
  
  # Azure OpenAI specific
  azure_openai:
    endpoint: ""                    # Azure OpenAI endpoint
    deployment_name: ""             # Deployment name
    api_version: "2024-02-15-preview"  # API version
  
  # Caching and performance
  cache:
    enabled: true                   # Enable response caching
    ttl: 3600s                      # Cache time-to-live
    max_size: "100MB"               # Maximum cache size
  
  # Rate limiting
  rate_limiting:
    requests_per_minute: 60         # Request rate limit
    burst_limit: 10                 # Burst request limit

# ===== Performance Configuration =====
performance:
  # Memory settings
  max_memory_usage: "512MB"         # Maximum memory usage
  gc_threshold: "256MB"             # Garbage collection threshold
  
  # Concurrency
  max_concurrent_requests: 5        # Maximum concurrent API requests
  worker_pool_size: 10              # Worker pool size
  
  # Caching
  resource_cache_size: 1000         # Number of resources to cache
  subscription_cache_ttl: 3600s     # Subscription cache TTL
  
  # Network
  connection_timeout: 10s           # Network connection timeout
  read_timeout: 30s                 # Network read timeout
  keep_alive: true                  # Enable HTTP keep-alive

# ===== Logging Configuration =====
logging:
  level: "info"                     # debug, info, warn, error
  output: "file"                    # console, file, both
  file_path: "~/.local/share/aztui/logs/aztui.log"
  max_file_size: "10MB"             # Maximum log file size
  max_backups: 5                    # Number of backup log files
  compress_backups: true            # Compress old log files
  
  # Component logging
  components:
    azure_api: "info"               # Azure API calls
    ai_requests: "debug"            # AI API requests
    ui_events: "warn"               # UI event logging
    performance: "info"             # Performance metrics

# ===== Security Configuration =====
security:
  # API key handling
  mask_api_keys: true               # Mask API keys in logs
  store_keys_encrypted: false       # Encrypt stored API keys
  
  # TLS settings
  tls_verify: true                  # Verify TLS certificates
  tls_min_version: "1.2"            # Minimum TLS version
  
  # Audit logging
  audit_log: true                   # Enable audit logging
  audit_file: "~/.local/share/aztui/audit.log"

# ===== Keyboard Shortcuts =====
keybindings:
  # Navigation
  move_up: "k"
  move_down: "j"
  move_left: "h"
  move_right: "l"
  expand_collapse: "space"
  select: "enter"
  
  # Actions
  ai_analysis: "a"
  terraform_gen: "T"
  bicep_gen: "B"
  metrics_dashboard: "M"
  cost_optimization: "O"
  edit_resource: "E"
  delete_resource: "ctrl+d"
  
  # View controls
  refresh_groups: "r"
  refresh_health: "ctrl+r"
  toggle_auto_refresh: "h"
  search: "/"
  help: "?"
  quit: "q"
  
  # Tab management
  new_tab: "ctrl+t"
  close_tab: "ctrl+w"
  next_tab: "ctrl+tab"
  prev_tab: "ctrl+shift+tab"

# ===== Theme Configuration =====
themes:
  azure:
    primary: "#0078d4"              # Azure blue
    secondary: "#106ebe"            # Darker blue
    accent: "#00bcf2"               # Light blue
    success: "#107c10"              # Green
    warning: "#ff8c00"              # Orange
    error: "#d13438"                # Red
    text: "#323130"                 # Dark gray
    background: "#ffffff"           # White
    surface: "#f3f2f1"              # Light gray
  
  dark:
    primary: "#0078d4"
    secondary: "#106ebe"
    accent: "#00bcf2"
    success: "#107c10"
    warning: "#ff8c00"
    error: "#d13438"
    text: "#ffffff"
    background: "#1e1e1e"
    surface: "#2d2d30"
  
  # Custom theme support
  custom:
    primary: "#custom_color"
    # ... define all colors

# ===== Advanced Configuration =====
advanced:
  # Experimental features
  experimental:
    streaming_ai: false             # Enable streaming AI responses
    real_time_metrics: false       # Enable real-time metrics
    bulk_operations: false          # Enable bulk operations
  
  # Debug options
  debug:
    enabled: false                  # Enable debug mode
    verbose_logging: false          # Extra verbose logging
    api_request_logging: false      # Log all API requests
    performance_metrics: false     # Enable performance metrics
  
  # Feature flags
  features:
    resource_actions: true          # Enable resource actions
    terraform_validation: true     # Validate generated Terraform
    cost_forecasting: false         # Enable cost forecasting
    compliance_checking: false     # Enable compliance checks

# ===== Plugin Configuration =====
plugins:
  enabled: false                    # Enable plugin system
  directory: "~/.config/aztui/plugins"
  auto_load: true                   # Auto-load plugins on startup
  
  # Available plugins
  available:
    - name: "resource-validator"
      enabled: false
    - name: "cost-reporter"
      enabled: false
    - name: "security-scanner"
      enabled: false
```

---

## Environment Variables

### Core Variables
```bash
# Application settings
export AZURE_TUI_CONFIG="/path/to/config.yaml"
export AZURE_TUI_THEME="azure"
export AZURE_TUI_LOG_LEVEL="info"
export AZURE_TUI_DEBUG="false"

# Azure configuration
export AZURE_TUI_SUBSCRIPTION="your-subscription-id"
export AZURE_TUI_TENANT="your-tenant-id"
export AZURE_TUI_TIMEOUT="30s"

# AI configuration
export AZURE_TUI_AI_API_KEY="your-api-key"
export AZURE_TUI_AI_MODEL="gpt-4"
export AZURE_TUI_AI_ENDPOINT="https://api.openai.com/v1"

# UI configuration
export AZURE_TUI_INTERFACE_MODE="tree"
export AZURE_TUI_AUTO_REFRESH="true"
export AZURE_TUI_MOUSE_SUPPORT="true"
```

### Azure Authentication
```bash
# Service Principal authentication
export AZURE_CLIENT_ID="your-client-id"
export AZURE_CLIENT_SECRET="your-client-secret"
export AZURE_TENANT_ID="your-tenant-id"

# Certificate authentication
export AZURE_CLIENT_CERTIFICATE_PATH="/path/to/cert.pem"

# Managed Identity (when running on Azure)
export AZURE_USE_MSI="true"
```

### AI Provider Configuration
```bash
# OpenAI
export OPENAI_API_KEY="sk-your-key"
export OPENAI_MODEL="gpt-4"

# Azure OpenAI
export AZURE_OPENAI_ENDPOINT="https://your-resource.openai.azure.com/"
export AZURE_OPENAI_API_KEY="your-key"
export AZURE_OPENAI_DEPLOYMENT="gpt-4"

# Local models (Ollama)
export OLLAMA_ENDPOINT="http://localhost:11434"
export OLLAMA_MODEL="llama2"
```

---

## Theme Customization

### Built-in Themes
1. **azure**: Microsoft Azure theme (default)
2. **dark**: Dark mode theme
3. **light**: Light mode theme
4. **terminal**: Terminal-native colors

### Custom Theme Creation
Create `~/.config/aztui/themes/mytheme.yaml`:
```yaml
name: "My Custom Theme"
description: "A custom theme for Azure TUI"

colors:
  # Primary colors
  primary: "#0078d4"
  secondary: "#106ebe"
  accent: "#00bcf2"
  
  # Status colors
  success: "#107c10"
  warning: "#ff8c00"
  error: "#d13438"
  info: "#0078d4"
  
  # Text colors
  text_primary: "#323130"
  text_secondary: "#605e5c"
  text_disabled: "#a19f9d"
  
  # Background colors
  background: "#ffffff"
  surface: "#f3f2f1"
  surface_variant: "#e1dfdd"
  
  # Interactive elements
  hover: "#f3f2f1"
  selected: "#deecf9"
  focused: "#cfe4fa"
  pressed: "#b8d4f1"

# Component-specific styling
components:
  tree_view:
    background: "${colors.background}"
    text: "${colors.text_primary}"
    selected_background: "${colors.selected}"
    selected_text: "${colors.text_primary}"
  
  status_bar:
    background: "${colors.primary}"
    text: "#ffffff"
  
  tabs:
    background: "${colors.surface}"
    text: "${colors.text_primary}"
    active_background: "${colors.background}"
    active_text: "${colors.primary}"
  
  health_indicators:
    healthy: "${colors.success}"
    warning: "${colors.warning}"
    critical: "${colors.error}"
    unknown: "${colors.text_disabled}"
```

### Theme Application
```bash
# Use custom theme
aztui --theme="mytheme"

# Or set in config
echo "ui.theme: mytheme" >> ~/.config/aztui/config.yaml
```

---

## Azure Integration

### Subscription Management
```yaml
azure:
  subscriptions:
    - id: "sub-1-id"
      name: "Production"
      is_default: true
      
    - id: "sub-2-id"
      name: "Development"
      is_default: false
  
  # Auto-discovery
  auto_discover_subscriptions: true
  exclude_subscriptions:
    - "test-subscription-id"
```

### Resource Group Filtering
```yaml
azure:
  resource_groups:
    # Include patterns
    include_patterns:
      - "prod-*"
      - "staging-*"
    
    # Exclude patterns
    exclude_patterns:
      - "temp-*"
      - "*-delete"
    
    # Maximum groups to load
    max_groups: 100
```

### Resource Type Configuration
```yaml
azure:
  resource_types:
    # Supported resource types
    supported:
      - "Microsoft.Compute/virtualMachines"
      - "Microsoft.Storage/storageAccounts"
      - "Microsoft.Network/virtualNetworks"
      - "Microsoft.Web/sites"
    
    # Custom icons
    icons:
      "Microsoft.Compute/virtualMachines": "🖥️"
      "Microsoft.Storage/storageAccounts": "💾"
      "Microsoft.Network/virtualNetworks": "🌐"
```

---

## Performance Tuning

### For Large Environments
```yaml
performance:
  # Reduce resource loading
  azure:
    lazy_loading: true
    pagination_size: 50
    max_resources_per_group: 500
  
  # Optimize UI rendering
  ui:
    virtual_scrolling: true
    render_throttle: 16ms
    max_visible_items: 100
  
  # Aggressive caching
  cache:
    aggressive_caching: true
    preload_subscriptions: true
    background_refresh: true
```

### For Slow Connections
```yaml
performance:
  # Longer timeouts
  azure:
    timeout: 60s
    connection_timeout: 30s
  
  # Reduced refresh rates
  ui:
    auto_refresh: false
    refresh_interval: 300s
  
  # Minimal data loading
  data:
    minimal_metadata: true
    skip_health_checks: true
```

---

## Security Best Practices

### API Key Management
```yaml
security:
  # Never store API keys in config files
  api_keys:
    use_environment_variables: true
    use_key_vault: false  # Future feature
    encryption: false     # Future feature
  
  # Audit logging
  audit:
    log_api_calls: true
    log_resource_access: true
    log_ai_requests: false  # Contains sensitive data
```

### Network Security
```yaml
security:
  network:
    # TLS configuration
    tls_verify: true
    tls_min_version: "1.2"
    certificate_pinning: false
    
    # Proxy support
    proxy:
      http_proxy: ""
      https_proxy: ""
      no_proxy: "localhost,127.0.0.1"
```

---

## Validation and Testing

### Configuration Validation
```bash
# Validate configuration file
aztui --validate-config

# Test Azure connectivity
aztui --test-azure-connection

# Test AI connectivity
aztui --test-ai-connection

# Check all settings
aztui --check-all
```

### Configuration Migration
```bash
# Migrate from v1.x config
aztui --migrate-config --from-version="1.0"

# Backup current config
aztui --backup-config

# Restore config
aztui --restore-config --backup-file="config-backup.yaml"
```

---

## Configuration Examples

### Development Environment
```yaml
ui:
  theme: "dark"
  auto_refresh: false
  
azure:
  timeout: 10s
  max_resources_per_group: 100
  
ai:
  provider: "openai"
  model: "gpt-3.5-turbo"
  
logging:
  level: "debug"
  output: "console"
```

### Production Environment
```yaml
ui:
  theme: "azure"
  auto_refresh: true
  refresh_interval: 60s
  
azure:
  timeout: 30s
  max_retries: 5
  
ai:
  provider: "azure_openai"
  model: "gpt-4"
  cache:
    enabled: true
    ttl: 7200s
  
logging:
  level: "info"
  output: "file"
  
security:
  audit_log: true
  mask_api_keys: true
```

### Minimal Configuration
```yaml
# Only essential settings
azure:
  default_subscription: "your-subscription-id"
  
ai:
  api_key: "your-api-key"
```

---

This configuration guide provides comprehensive coverage of all Azure TUI settings. Start with the minimal configuration and gradually add features as needed.
