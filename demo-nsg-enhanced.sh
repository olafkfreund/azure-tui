#!/bin/zsh

# Azure TUI - NSG Open Ports Analysis Demo
# =======================================

print_header() {
    echo ""
    echo "🔒 Azure TUI - NSG Open Ports Analysis Demo"
    echo "==========================================="
    echo ""
}

print_section() {
    echo ""
    echo "📋 $1"
    echo "$(echo $1 | sed 's/./─/g')"
    echo ""
}

print_feature() {
    echo "✅ $1"
}

print_example() {
    echo "💡 $1"
}

print_header

print_section "What's New in NSG Analysis"

print_feature "Comprehensive Open Ports Table"
echo "   • All inbound ports with service identification"
echo "   • Color-coded security risk assessment"
echo "   • Source address analysis (public vs private)"
echo "   • Rule association and priority information"
echo ""

print_feature "Enhanced Security Rules Display"
echo "   • Priority-based sorting for logical flow"
echo "   • Color-coded Allow/Deny and Inbound/Outbound"
echo "   • Complete source and destination details"
echo "   • Port range handling and display"
echo ""

print_feature "Automated Security Analysis"
echo "   • Risk detection for public-facing sensitive ports"
echo "   • Best practice compliance checking"
echo "   • Statistical summaries and recommendations"
echo "   • Actionable security insights"

print_section "Service Recognition Database"

echo "The system now recognizes 80+ common services:"
echo ""
echo "🌐 Web Services:"
echo "   • 80 (HTTP), 443 (HTTPS), 8080 (HTTP Alt), 8443 (HTTPS Alt)"
echo ""
echo "🔐 Remote Access:"
echo "   • 22 (SSH), 3389 (RDP), 23 (Telnet)"
echo ""
echo "🗄️  Database Services:"
echo "   • 3306 (MySQL), 5432 (PostgreSQL), 1433 (SQL Server)"
echo "   • 27017 (MongoDB), 6379 (Redis)"
echo ""
echo "🚢 Container & Orchestration:"
echo "   • 2376/2377 (Docker), 6443 (Kubernetes API), 10250 (Kubelet)"
echo ""
echo "📊 Monitoring & DevOps:"
echo "   • 3000 (Grafana), 9090 (Prometheus), 5601 (Kibana)"
echo "   • 9200 (Elasticsearch)"

print_section "Security Risk Assessment"

echo "🔴 HIGH RISK: Sensitive ports open to internet (0.0.0.0/0)"
echo "   • SSH (22), RDP (3389), Telnet (23)"
echo "   • Database ports: 3306, 5432, 1433, 27017"
echo "   • Management interfaces"
echo ""
echo "🟡 MEDIUM RISK: Many public ports or privileged ports (<1024)"
echo "   • Multiple services exposed to internet"
echo "   • System ports without justification"
echo ""
echo "🟢 LOW RISK: Standard web ports or private network access"
echo "   • HTTP/HTTPS to internet"
echo "   • Internal services on private networks"

print_section "How to Use"

echo "1. Launch Azure TUI:"
echo "   ./azure-tui"
echo "   # or"
echo "   go run cmd/main.go"
echo ""
echo "2. Navigate to NSG Resources:"
echo "   • Use j/k or arrow keys to navigate"
echo "   • Look for 'Microsoft.Network/networkSecurityGroups' resources"
echo "   • Expand resource groups with Space/Enter"
echo ""
echo "3. View NSG Details:"
echo "   • Select an NSG resource"
echo "   • Press 'G' to open enhanced NSG details"
echo ""
echo "4. Analyze the Output:"
echo "   • Scroll through the comprehensive analysis"
echo "   • Review open ports table"
echo "   • Check security recommendations"

print_section "Example Output Structure"

echo "📋 Basic Information"
echo "   Resource group, location, rule counts, associations"
echo ""
echo "🌐 Open Ports Analysis Table"
echo "   Port | Protocol | Source | Rule Name | Priority | Service"
echo "   -----|----------|--------|-----------|----------|--------"
echo "   22   | TCP      | 0.0.0.0/0 | AllowSSH | 1000    | SSH (TCP)"
echo "   80   | TCP      | *      | AllowHTTP | 1010    | HTTP (TCP)"
echo "   443  | TCP      | *      | AllowHTTPS| 1020    | HTTPS (TCP)"
echo ""
echo "📜 Security Rules Details"
echo "   Complete rules table sorted by priority"
echo ""
echo "🛡️  Security Analysis"
echo "   Risk assessment and recommendations"

print_section "Advanced Features"

print_feature "Port Range Handling"
echo "   • Single ports: 80, 443, 22"
echo "   • Port ranges: 8000-8010"
echo "   • Multiple ports: 80,443,8080"
echo "   • Wildcard handling: * (all ports)"
echo ""

print_feature "Intelligent Grouping"
echo "   • Ports grouped by source address"
echo "   • Public vs private access separation"
echo "   • Risk-based sorting and highlighting"
echo ""

print_feature "Professional Formatting"
echo "   • Terminal-width responsive tables"
echo "   • Color-coded risk indicators"
echo "   • Consistent spacing and alignment"

print_section "Ready to Test!"

echo "🚀 Launch Azure TUI and navigate to any NSG resource"
echo "📊 Press 'G' to see the enhanced open ports analysis"
echo "🔍 Explore the comprehensive security assessment"
echo ""
echo "The NSG open ports table enhancement is now complete and ready for use!"
echo ""
echo "📖 For detailed documentation, see:"
echo "   docs/NSG_OPEN_PORTS_ENHANCEMENT_COMPLETE.md"
