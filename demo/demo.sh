#!/usr/bin/env bash

# Azure TUI - Enhanced Features Demo
# This script demonstrates the new AI-powered features

set -e

echo "🚀 Azure TUI - Enhanced Features Demo"
echo "====================================="
echo ""

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    exit 1
fi

# Check if the application builds
echo "🔨 Building Azure TUI..."
if ! go build -o main cmd/main.go; then
    echo "❌ Build failed. Please check the code for errors."
    exit 1
fi
echo "✅ Build successful!"
echo ""

# Create demo configuration
echo "📝 Setting up demo environment..."
mkdir -p ~/.config/azure-tui

cat > ~/.config/azure-tui/config.yaml << 'EOF'
# Azure TUI Configuration
naming:
  standard: "demo-{{type}}-{{name}}"
  environment: "dev"

ai:
  provider: "openai"
  model: "gpt-4"
  enabled: true

features:
  metrics_dashboard: true
  ai_analysis: true
  iac_generation: true
  cost_optimization: true
EOF

echo "✅ Demo configuration created!"
echo ""

# Show feature overview
echo "🎯 Enhanced Features Overview:"
echo "────────────────────────────────────"
echo "• 🤖 AI-powered resource analysis"
echo "• 📊 Interactive metrics dashboard"
echo "• ✏️  Resource configuration editor"
echo "• 🗑️  Safe resource deletion with confirmation"
echo "• 🔧 Terraform/Bicep code generation"
echo "• 💰 Cost optimization recommendations"
echo "• 🏠 Modern tabbed interface with Azure icons"
echo "• ⌨️  Comprehensive keyboard shortcuts"
echo ""

echo "🎮 How to Use:"
echo "──────────────"
echo "1. Navigate resource groups with ↑/↓ arrows"
echo "2. Navigate resources with ←/→ arrows"
echo "3. Press Enter to open a resource in a new tab"
echo "4. Press 'a' for AI analysis of selected resource"
echo "5. Press 'M' for metrics dashboard"
echo "6. Press 'E' to edit resource configuration"
echo "7. Press 'T' for Terraform code generation"
echo "8. Press 'B' for Bicep code generation"
echo "9. Press 'O' for cost optimization analysis"
echo "10. Press F1 to see all shortcuts"
echo "11. Press Esc to close any dialog"
echo "12. Press 'q' to quit"
echo ""

echo "🔑 AI Features (Optional):"
echo "─────────────────────────"
echo "To enable full AI features, set your OpenAI API key:"
echo "export OPENAI_API_KEY='your-api-key-here'"
echo ""
echo "Without an API key, the app runs in demo mode with sample data."
echo ""

echo "🎪 Starting Azure TUI Demo..."
echo "Press Ctrl+C to exit when done exploring."
echo ""

# Add a small delay to let user read the instructions
sleep 3

# Run the application
echo "🚀 Launching Azure TUI..."
./main

echo ""
echo "👋 Thanks for trying Azure TUI Enhanced Features!"
echo "For more information, see FEATURES.md"
