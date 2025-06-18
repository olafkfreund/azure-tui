#!/bin/bash

# Azure TUI Container Instance Demo Script
# Demonstrates the new Container Instance management capabilities

echo "🐳 Azure TUI - Container Instance Management Demo"
echo "=================================================="
echo

# Check if container instance exists
echo "📋 Step 1: Checking available container instances..."
az container list --output table
echo

# Show detailed container instance information
echo "🔍 Step 2: Container Instance Details..."
echo "Container: cadmin (Resource Group: con_demo_01)"
az container show --name cadmin --resource-group con_demo_01 --query '{name:name,state:provisioningState,ip:ipAddress.ip,image:containers[0].image,cpu:containers[0].resources.requests.cpu,memory:containers[0].resources.requests.memoryInGb}' --output table
echo

# Show container logs (last 10 lines)
echo "📜 Step 3: Container Logs (last 10 lines)..."
az container logs --name cadmin --resource-group con_demo_01 --tail 10
echo

# Show Azure TUI with Container Instance support
echo "🚀 Step 4: Launching Azure TUI with Container Instance Support..."
echo 
echo "New Container Instance Features Available:"
echo "• [s] Start Container Instance"
echo "• [S] Stop Container Instance"
echo "• [r] Restart Container Instance"
echo "• [L] Get Container Logs"
echo "• [E] Exec into Container"
echo "• [a] Attach to Container"
echo "• [u] Scale Container Resources"
echo "• [I] Show Detailed Information"
echo
echo "Navigate to the con_demo_01 resource group and select the 'cadmin' container"
echo "to see the new container management options!"
echo
echo "Press Enter to launch Azure TUI..."
read -r

# Launch Azure TUI
./azure-tui
