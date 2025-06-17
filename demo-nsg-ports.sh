#!/bin/bash

# Demo script for the enhanced NSG open ports table functionality
# =============================================================

echo "🔒 Azure TUI - Enhanced NSG Open Ports Analysis Demo"
echo "====================================================="
echo ""

echo "📋 New NSG Features:"
echo "--------------------"
echo "✅ Comprehensive Open Ports Table"
echo "   • Port numbers with service identification"
echo "   • Source address analysis (public vs private)"
echo "   • Protocol information (TCP/UDP)"
echo "   • Associated rule names and priorities"
echo "   • Security risk assessment with color coding"
echo ""

echo "✅ Enhanced Security Rules Table"
echo "   • Sorted by priority for logical flow"
echo "   • Color-coded access (Allow/Deny) and direction"
echo "   • Complete source and destination information"
echo "   • Port range handling and display"
echo ""

echo "✅ Security Analysis & Recommendations"
echo "   • Risk assessment for public-facing ports"
echo "   • Detection of sensitive services (SSH, RDP, etc.)"
echo "   • Rule count statistics and summaries"
echo "   • Best practice recommendations"
echo ""

echo "🎯 Port Service Recognition:"
echo "---------------------------"
echo "The system now recognizes common services:"
echo "• Web Services: 80 (HTTP), 443 (HTTPS), 8080 (HTTP Alt)"
echo "• Remote Access: 22 (SSH), 3389 (RDP), 23 (Telnet)"
echo "• Databases: 3306 (MySQL), 5432 (PostgreSQL), 1433 (SQL Server)"
echo "• Container Services: 2376/2377 (Docker), 6443 (Kubernetes API)"
echo "• Monitoring: 3000 (Grafana), 9090 (Prometheus), 5601 (Kibana)"
echo "• And many more..."
echo ""

echo "🛡️  Security Risk Assessment:"
echo "-----------------------------"
echo "• 🔴 HIGH RISK: Sensitive ports open to internet (0.0.0.0/0)"
echo "• 🟡 MEDIUM RISK: Many public ports or privileged ports (<1024)"
echo "• 🟢 LOW RISK: Private network access or standard web ports"
echo ""

echo "🚀 How to Use:"
echo "---------------"
echo "1. Run: go run cmd/main.go"
echo "2. Navigate to any Network Security Group resource"
echo "3. Press 'G' to view NSG details with the new open ports table"
echo "4. Scroll through the comprehensive analysis"
echo ""

echo "💡 Example Output Sections:"
echo "---------------------------"
echo "📋 Basic Information - NSG overview and associations"
echo "🌐 Open Ports Analysis - Table of all open inbound ports"
echo "📜 Security Rules Details - Complete rules with color coding"
echo "🛡️  Security Analysis - Risk assessment and recommendations"
echo ""

echo "✨ Ready to test the enhanced NSG analysis!"
echo "Launch Azure TUI and navigate to any NSG resource to see the new functionality."
