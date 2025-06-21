#!/bin/bash

echo "🔄 Azure TUI Log Streaming Service Demo"
echo "========================================"
echo ""

# Make the logstream script executable
chmod +x logstream.go

echo "📋 Available Commands:"
echo ""
echo "1. Stream logs for a specific resource:"
echo "   go run logstream.go /subscriptions/46b2dfbe-fe9e-4433-b327-b2dc32c8af5e/resourceGroups/dem01_group/providers/Microsoft.Network/networkInterfaces/dem01211_z1"
echo ""
echo "2. Stream logs for entire subscription:"
echo "   go run logstream.go --subscription 46b2dfbe-fe9e-4433-b327-b2dc32c8af5e"
echo ""
echo "3. Stream logs for resource group:"
echo "   go run logstream.go --resource-group dem01_group"
echo ""
echo "4. Stream from Log Analytics workspace:"
echo "   go run logstream.go --workspace your-workspace-id"
echo ""

echo "🎯 Testing with demo data (resource-specific):"
echo "Press Ctrl+C to stop the stream"
echo ""

# Run the log streaming service with demo data
go run logstream.go /subscriptions/46b2dfbe-fe9e-4433-b327-b2dc32c8af5e/resourceGroups/dem01_group/providers/Microsoft.Network/networkInterfaces/dem01211_z1
