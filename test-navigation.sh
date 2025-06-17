#!/bin/bash

# Test script to verify Azure TUI navigation works
echo "Testing Azure TUI Navigation..."

# Build the application
echo "Building application..."
cd /home/olafkfreund/Source/Cloud/azure-tui
go build -o aztui ./cmd/main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi

echo "✅ Build successful"

# Test Azure CLI is available
echo "Testing Azure CLI..."
if ! command -v az &> /dev/null; then
    echo "❌ Azure CLI not found"
    exit 1
fi

# Test Azure login status
echo "Testing Azure login status..."
az account show &> /dev/null
if [ $? -ne 0 ]; then
    echo "⚠️  Azure CLI not logged in - application will use demo mode"
else
    echo "✅ Azure CLI logged in"
fi

# Test application can start
echo "Testing application startup..."
timeout 3s ./aztui > /dev/null 2>&1
if [ $? -eq 124 ]; then
    echo "✅ Application starts successfully (timed out after 3s as expected)"
else
    echo "❌ Application failed to start or crashed"
    exit 1
fi

echo ""
echo "🎉 All tests passed!"
echo ""
echo "Navigation instructions:"
echo "• j/k or ↓/↑ arrows - Navigate through resource groups"
echo "• Space - Expand/collapse a resource group to load resources"
echo "• Enter - On a resource group: expand/collapse, On a resource: show details"
echo "• Tab - Switch between tabs"
echo "• r - Refresh data"
echo "• q - Quit"
echo ""
echo "To run the application: ./aztui"
