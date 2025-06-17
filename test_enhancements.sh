#!/bin/bash

# Test script for Azure TUI enhancements
# This script tests the new table formatting, SSH, and AKS features

echo "🧪 Testing Azure TUI Enhancements"
echo "================================"

# Check if required tools are available
echo "📋 Checking prerequisites..."

# Check Azure CLI
if ! command -v az &> /dev/null; then
    echo "❌ Azure CLI not found. Please install: https://docs.microsoft.com/en-us/cli/azure/install-azure-cli"
    exit 1
else
    echo "✅ Azure CLI found"
fi

# Check kubectl
if ! command -v kubectl &> /dev/null; then
    echo "⚠️  kubectl not found. AKS features may not work properly"
    echo "   Install kubectl: https://kubernetes.io/docs/tasks/tools/"
else
    echo "✅ kubectl found"
fi

# Check if logged into Azure
echo "🔐 Checking Azure authentication..."
if ! az account show &> /dev/null; then
    echo "❌ Not logged into Azure. Please run: az login"
    exit 1
else
    echo "✅ Azure authentication verified"
    SUBSCRIPTION=$(az account show --query name -o tsv)
    echo "   Current subscription: $SUBSCRIPTION"
fi

# Build the application
echo "🔨 Building Azure TUI..."
cd /home/olafkfreund/Source/Cloud/azure-tui
if go build -o azure-tui cmd/main.go; then
    echo "✅ Build successful"
else
    echo "❌ Build failed"
    exit 1
fi

# Check if there are resources to test with
echo "🔍 Checking for test resources..."
RG_COUNT=$(az group list --query "length(@)")
if [ "$RG_COUNT" -eq 0 ]; then
    echo "⚠️  No resource groups found. Consider creating test resources."
else
    echo "✅ Found $RG_COUNT resource group(s)"
fi

VM_COUNT=$(az vm list --query "length(@)")
if [ "$VM_COUNT" -eq 0 ]; then
    echo "⚠️  No VMs found. SSH testing will be limited."
else
    echo "✅ Found $VM_COUNT VM(s) - SSH features can be tested"
fi

AKS_COUNT=$(az aks list --query "length(@)")
if [ "$AKS_COUNT" -eq 0 ]; then
    echo "⚠️  No AKS clusters found. AKS testing will be limited."
else
    echo "✅ Found $AKS_COUNT AKS cluster(s) - AKS features can be tested"
fi

echo ""
echo "🚀 Ready to test! Run the application with:"
echo "   ./azure-tui"
echo ""
echo "📖 Test the following new features:"
echo "   1. Navigate to any resource and check property table formatting"
echo "   2. Select a VM and test SSH actions: [c] SSH, [b] Bastion"
echo "   3. Select an AKS cluster and test: [p] Pods, [D] Deployments, [n] Nodes, [v] Services"
echo "   4. Try [s] Start and [S] Stop actions on resources"
echo ""
echo "Press any key to launch Azure TUI or Ctrl+C to exit..."
read -n 1 -s

# Launch the application
echo "🚀 Launching Azure TUI..."
./azure-tui
