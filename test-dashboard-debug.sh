#!/bin/bash

echo "🧪 Testing Dashboard Debug Fixes"
echo "================================="

# Clear previous debug logs
rm -f debug.txt

echo "Starting Azure TUI..."
echo "1. Navigate to a resource"
echo "2. Press Shift+D to trigger the enhanced dashboard"
echo "3. Press q to quit"
echo "4. Check debug.txt for error messages"
echo ""
echo "Expected: JSON parsing errors should be fixed"
echo ""

# Run the TUI (this will be interactive)
./azure-tui

echo ""
echo "Debug output:"
echo "============="
if [ -f debug.txt ]; then
    cat debug.txt
else
    echo "No debug.txt file found"
fi
