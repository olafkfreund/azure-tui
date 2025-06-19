#!/bin/bash

# Demo script for Azure Key Vault Integration in Azure TUI
# This script demonstrates the Key Vault secret management functionality

echo "🔐 Azure TUI - Key Vault Integration Demo"
echo "=========================================="
echo

echo "📋 Prerequisites:"
echo "1. Azure CLI installed and logged in"
echo "2. Access to an Azure Key Vault"
echo "3. Appropriate permissions for secret management"
echo

echo "🎯 Key Vault Features in Azure TUI:"
echo
echo "✅ Secret Management:"
echo "   • List all secrets in a Key Vault"
echo "   • Create new secrets with metadata"
echo "   • Delete existing secrets"
echo "   • View secret details (metadata only, not values)"
echo
echo "🔐 Security Features:"
echo "   • Never displays actual secret values"
echo "   • Uses Azure CLI authentication"
echo "   • Proper error handling for permissions"
echo "   • Audit trail for all operations"
echo

echo "⌨️  Keyboard Shortcuts:"
echo "   K        - List secrets (when Key Vault selected)"
echo "   Shift+K  - Create secret"
echo "   Ctrl+D   - Delete secret"
echo "   ?        - Show help"
echo "   Esc      - Navigate back"
echo

echo "🚀 How to Test:"
echo "1. Run: ./azure-tui"
echo "2. Navigate to a Resource Group containing a Key Vault"
echo "3. Select the Key Vault resource"
echo "4. Press 'K' to list secrets"
echo "5. Press 'Shift+K' to create a demo secret"
echo "6. Press 'Ctrl+D' to delete the demo secret"
echo "7. Use 'Esc' to navigate between views"
echo

echo "🔍 Demo Key Vault Operations:"
echo

# Check if Azure CLI is available
if command -v az &> /dev/null; then
    echo "✅ Azure CLI is available"
    
    # Check if user is logged in
    if az account show &> /dev/null; then
        echo "✅ Azure CLI authenticated"
        
        # List available Key Vaults
        echo "📋 Available Key Vaults in current subscription:"
        az keyvault list --query "[].{Name:name, Location:location, ResourceGroup:resourceGroup}" --output table 2>/dev/null || echo "   No Key Vaults found or insufficient permissions"
    else
        echo "❌ Azure CLI not authenticated. Run: az login"
    fi
else
    echo "❌ Azure CLI not installed"
fi

echo
echo "🎮 Start the Azure TUI application:"
echo "   cd /path/to/azure-tui && ./azure-tui"
echo
echo "💡 Tips:"
echo "   • Use Tab to switch between panels"
echo "   • Press ? for complete keyboard shortcuts"
echo "   • Key Vault shortcuts only appear when a Key Vault is selected"
echo "   • All operations respect Azure RBAC permissions"
echo
echo "🏆 Integration Status: COMPLETE ✅"
