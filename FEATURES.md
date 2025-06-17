# Azure TUI - Advanced Features

## AI-Powered Resource Management

The Azure TUI now includes comprehensive AI-powered features for intelligent resource management, analysis, and code generation.

### Key Features

#### 🤖 AI Resource Analysis

- Press `a` to get AI-powered analysis of any selected resource
- Provides configuration summary, optimization recommendations, and security considerations
- Works with OpenAI API (set `OPENAI_API_KEY` environment variable)

#### 📊 Interactive Metrics Dashboard

- Press `M` to view real-time metrics for any resource
- Shows CPU, memory, disk, and network usage with color-coded alerts
- Includes ASCII trend graphs and interactive controls

#### ✏️ Resource Configuration Editor

- Press `E` to edit resource configurations in a guided dialog
- Shows current settings with field-by-field editing
- Validates changes before applying

#### 🗑️ Safe Resource Deletion

- Press `Ctrl+D` to initiate resource deletion with confirmation dialog
- Shows resource details and requires explicit confirmation
- Prevents accidental deletions with clear warnings

#### 🔧 Infrastructure as Code Generation

- Press `T` to generate Terraform code for any resource
- Press `B` to generate Bicep code with AI assistance
- Creates complete, deployable infrastructure templates

#### 💰 Cost Optimization Analysis

- Press `O` to get AI-powered cost optimization suggestions
- Analyzes entire resource groups for savings opportunities
- Provides right-sizing and reserved instance recommendations

### Navigation & Shortcuts

| Key | Action |
|-----|--------|
| `↑/↓` or `j/k` | Navigate tree items |
| `Space` | Expand/collapse tree nodes |
| `Enter` | Open resource in content tab |
| `Tab/Shift+Tab` | Switch between content tabs |
| `Ctrl+W` | Close current content tab |
| `F2` | Toggle tree view/traditional mode |
| `a` | AI analysis of selected resource |
| `M` | Show metrics dashboard |
| `E` | Edit resource configuration |
| `Ctrl+D` | Delete resource (with confirmation) |
| `T` | Generate Terraform code |
| `B` | Generate Bicep code |
| `O` | Cost optimization analysis |
| `?` | Show all keyboard shortcuts |
| `Esc` | Close any open dialog/popup |
| `q` | Quit application |

### Tab System

The enhanced TUI features a sophisticated interface system with two modes:

#### **Tree View Mode (Default)**
- **NeoVim-style Interface**: Hierarchical tree navigation on the left
- **Content Tabs**: Right-side tabbed area for opened resources  
- **Powerline Statusbar**: Modern status bar with Azure context
- **Vim Navigation**: Familiar `j/k` keys and space for expand/collapse

#### **Traditional Mode (F2 to toggle)**
- **Two-Panel Layout**: Left panel for resource groups, right for resources
- **Resource Tabs**: Traditional tab system for opened resources
- **Legacy Navigation**: Arrow key navigation between panels

#### **Tab Features**
- **Azure Service Icons**: Resource-specific icons for easy identification
- **Multi-Resource Management**: Open multiple resources simultaneously
- **Tab Management**: Full support for opening, closing, and switching tabs
- **Persistent State**: Tabs maintain their content when switching between them

### Visual Enhancements

- **Modern Styling**: Uses lipgloss for consistent, beautiful UI
- **Azure Icons**: Service-specific icons throughout the interface
- **Color Coding**: Resource status, warnings, and selections
- **Responsive Layout**: Panels automatically adjust content
- **Unicode Alignment**: Proper text alignment with Unicode support

### AI Integration

The application integrates with AI providers:

- **GitHub Copilot**: Recommended for enhanced Azure-specific analysis
- **OpenAI GPT-4**: Alternative for general analysis and code generation
- **Custom Agents**: Specialized agents for different Azure scenarios

Set up AI integration (choose one):

```bash
# Option 1: GitHub Copilot (Recommended)
export GITHUB_TOKEN="your-github-token"
export USE_GITHUB_COPILOT="true"  # optional, auto-detected

# Option 2: OpenAI API
export OPENAI_API_KEY="your-openai-api-key"
```

### Demo Mode

If Azure CLI is not configured or unavailable, the application runs in demo mode with sample data, allowing you to explore all features without an active Azure subscription.

### Architecture

The application follows a modular architecture:

- **TUI Layer**: Bubble Tea framework with lipgloss styling
- **Azure Integration**: Azure SDK and CLI integration
- **AI Services**: OpenAI client support
- **IaC Support**: Terraform and Bicep code analysis/generation
- **Configuration**: YAML-based user preferences

This enhanced version transforms the basic Azure resource browser into a comprehensive, AI-powered cloud management platform.
