#!/bin/bash

# Azure TUI Help Popup Demonstration
# Shows off the new scrolling, formatting, and ESC key improvements

echo "🎯 Azure TUI Help Popup Improvements Demo"
echo "========================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}📋 What's New in the Help Popup:${NC}"
echo ""

echo -e "${GREEN}1. 📜 Scrolling Functionality${NC}"
echo "   • Navigate through long content with j/k or ↑/↓"
echo "   • Scroll indicators show when more content is available"
echo "   • Smooth, bounded scrolling prevents going past edges"
echo "   • Scroll position resets when popup is reopened"
echo ""

echo -e "${GREEN}2. 📊 Improved Table Formatting${NC}"
echo "   • Professional column alignment for easy scanning"
echo "   • Color-coded sections (Navigation, Search, Actions, etc.)"
echo "   • Consistent 12-character padding for shortcuts"
echo "   • Wider popup (78 chars) for better readability"
echo ""

echo -e "${GREEN}3. 🚪 Fixed ESC Key Behavior${NC}"
echo "   • ESC now immediately closes help popup"
echo "   • '?' key also closes the popup for convenience"
echo "   • Clean popup handling with proper priority"
echo "   • No conflicts with other popup systems"
echo ""

echo -e "${CYAN}🎮 Interactive Demo Instructions:${NC}"
echo ""
echo "1. Launch Azure TUI with the command below"
echo "2. Press '?' to open the improved help popup"
echo "3. Use j/k or ↑/↓ to scroll through the content"
echo "4. Notice the beautiful table formatting and colors"
echo "5. Try ESC or '?' to close the popup"
echo "6. Reopen and notice scroll position is reset"
echo ""

echo -e "${YELLOW}📁 All categories available in help:${NC}"
echo "   🧭 Navigation  🔍 Search  ⚡ Resource Actions"
echo "   🌐 Network     🏗️ Terraform  🐳 Container Mgmt"
echo "   🔐 SSH & AKS   🔑 Key Vault  🎮 Interface"
echo ""

# Build if needed
if [ ! -f "./azure-tui" ]; then
    echo -e "${BLUE}📦 Building Azure TUI...${NC}"
    just build
    echo ""
fi

echo -e "${GREEN}🚀 Launching Azure TUI...${NC}"
echo -e "${CYAN}💡 Press '?' to see the improvements in action!${NC}"
echo ""

# Launch the application
./azure-tui
