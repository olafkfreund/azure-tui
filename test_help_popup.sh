#!/bin/bash

# Test Help Popup Improvements
# This script tests the improved help popup with scrolling and table formatting

echo "🔍 Testing Azure TUI Help Popup Improvements..."
echo "=============================================="
echo ""

# Build the application
echo "📦 Building Azure TUI..."
cd /home/olafkfreund/Source/Cloud/azure-tui
just build

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    
    echo "🎮 Testing Help Popup Features:"
    echo ""
    echo "1. 📋 **Scrolling Functionality**:"
    echo "   - Press '?' to open help popup"
    echo "   - Use j/k or ↑/↓ to scroll through content"
    echo "   - Content should show scroll indicators when needed"
    echo ""
    echo "2. 📊 **Table Formatting**:"
    echo "   - Shortcuts should be properly aligned"
    echo "   - Colors should make sections easily distinguishable"
    echo "   - Width should be optimized for readability"
    echo ""
    echo "3. 🚪 **ESC Key Behavior**:"
    echo "   - ESC should close help popup immediately"
    echo "   - '?' should also close the help popup"
    echo "   - Scroll position should reset when reopening"
    echo ""
    
    echo "🚀 Launching Azure TUI for testing..."
    echo "💡 Press '?' to open the help popup and test the improvements!"
    echo ""
    
    # Launch the application for manual testing
    ./azure-tui
    
else
    echo "❌ Build failed!"
    exit 1
fi
