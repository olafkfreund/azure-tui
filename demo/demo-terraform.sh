#!/bin/bash

# Azure TUI Terraform Integration Demo Script
# =============================================

echo "🏗️ Azure TUI Terraform Integration Demo"
echo "========================================"
echo ""

echo "📋 Terraform Features Available:"
echo "--------------------------------"
echo "• F  - Terraform Management Menu (Main entry point)"
echo "• Browse Files - View and edit Terraform files"
echo "• Create Templates - Generate new Terraform templates"  
echo "• Plan & Apply - Run terraform plan and apply"
echo "• View State - Show current Terraform state"
echo "• AI Assistant - Get AI help with Terraform"
echo "• Settings - Configure Terraform options"
echo ""

echo "🎯 Usage Instructions:"
echo "----------------------"
echo "1. Run: go run cmd/main.go"
echo "2. Navigate with j/k or arrow keys"
echo "3. Use Tab to switch between panels"
echo "4. Press 'F' to open Terraform Management"
echo "5. Use arrow keys to navigate the Terraform menu"
echo "6. Press Esc to go back to previous views"
echo ""

echo "✨ Enhanced Features:"
echo "--------------------"
echo "• Complete Terraform file management"
echo "• AI-powered Terraform code generation"
echo "• Integration with existing Azure resources"
echo "• Template generation for common Azure resources"
echo "• Plan, apply, and destroy operations"
echo "• State management and visualization"
echo ""

echo "🔧 Configuration:"
echo "-----------------"
echo "The Terraform integration uses:"
echo "• Source folder: ./terraform (configurable)"
echo "• Default location: uksouth (configurable)"
echo "• AI provider: OpenAI (configurable)"
echo "• Auto-format: enabled"
echo "• Validate on save: enabled"
echo ""

echo "🚀 How to Test:"
echo "---------------"
echo "1. Launch Azure TUI:"
echo "   ./azure-tui"
echo "   # or"
echo "   go run cmd/main.go"
echo ""
echo "2. Press 'F' to access Terraform Management:"
echo "   • The Terraform menu will display available options"
echo "   • Navigate with arrow keys"
echo "   • Select options with Enter"
echo ""
echo "3. Explore Terraform Features:"
echo "   • Browse existing Terraform files"
echo "   • Generate new templates for Azure resources"
echo "   • Use AI assistance for Terraform code"
echo "   • Run plan and apply operations"
echo ""
echo "4. Navigation:"
echo "   • Use 'Esc' to navigate back"
echo "   • Use '?' to see all keyboard shortcuts"
echo "   • Terraform shortcut 'F' is now available globally"
echo ""

echo "🔍 Demo Terraform Operations:"
echo "-----------------------------"

# Check if Terraform is installed
if command -v terraform &> /dev/null; then
    echo "✅ Terraform CLI is available"
    terraform version
else
    echo "❌ Terraform CLI not found. Install from: https://terraform.io"
fi

echo ""

# Check if Azure CLI is available
if command -v az &> /dev/null; then
    echo "✅ Azure CLI is available"
    
    # Check if user is logged in
    if az account show &> /dev/null; then
        echo "✅ Azure CLI authenticated"
        
        # Show current subscription
        echo "📋 Current Azure subscription:"
        az account show --query "{Name:name, SubscriptionId:id, TenantId:tenantId}" --output table 2>/dev/null || echo "   Failed to get subscription info"
    else
        echo "❌ Azure CLI not authenticated. Run: az login"
    fi
else
    echo "❌ Azure CLI not installed"
fi

echo ""
echo "💡 Tips:"
echo "--------"
echo "• Press 'F' from anywhere in Azure TUI to access Terraform"
echo "• Use Tab to switch between tree and details panels"
echo "• All Terraform operations respect Azure RBAC permissions"
echo "• AI assistance helps generate Terraform code for selected resources"
echo "• The Terraform manager integrates with existing Azure credentials"
echo ""
echo "🏆 Integration Status: COMPLETE ✅"
echo ""
echo "Ready to launch Azure TUI with Terraform integration!"
echo "Use: go run cmd/main.go"
