#!/bin/bash

echo "🧪 Testing Azure TUI Navigation"
echo "================================"

cd /home/olafkfreund/Source/Cloud/azure-tui

# Build the application
echo "Building application..."
go build -o aztui ./cmd/main.go
if [ $? -ne 0 ]; then
    echo "❌ Build failed"
    exit 1
fi
echo "✅ Build successful"

# Check if we're in a proper terminal
if [ -t 0 ]; then
    echo "✅ Running in proper terminal"
else
    echo "⚠️ Not in proper terminal - keyboard input might not work"
fi

echo ""
echo "🎮 Manual Test Instructions:"
echo "----------------------------"
echo "The application will start in 3 seconds."
echo "Once it starts, try these actions:"
echo ""
echo "1. Wait for resource groups to load (should see 4 groups)"
echo "2. Press 'j' or 'k' keys to navigate up/down"
echo "3. Press 'Space' or 'Enter' on a resource group to expand it"  
echo "4. Navigate to a resource and press 'Enter' to see details"
echo "5. Press 'q' to quit"
echo ""
echo "If nothing happens when you press keys, there's a keyboard input issue."
echo ""

# Countdown
for i in 3 2 1; do
    echo "Starting in $i..."
    sleep 1
done

echo ""
echo "🚀 Starting Azure TUI - Press 'q' to quit"
echo "=========================================="

# Start the application
exec ./aztui
