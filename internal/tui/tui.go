package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// This package will contain the Bubble Tea TUI logic and models.
// TODO: Implement main menu, environment dashboard, and navigation.

// PopupMsg is a message for showing popups (alarms/errors)
type PopupMsg struct {
	Title   string
	Content string
	Level   string // "error", "alarm", "info"
}

// MatrixGraphMsg is a message for showing a matrix/graph in the TUI
type MatrixGraphMsg struct {
	Title  string
	Rows   [][]string // 2D matrix of values
	Labels []string   // Row/column labels
}

// Tab represents a TUI tab/window (e.g., resource, connection, monitor, etc.)
type Tab struct {
	Title    string
	Content  string            // Could be a rendered panel, connection, or monitoring view
	Type     string            // e.g. "resource", "aks", "vm", "monitor", "health", "shell"
	Meta     map[string]string // Extra info (e.g. resource ID, connection params)
	Closable bool
}

// TabManager manages multiple tabs/windows in the TUI
// Supports tab open/close, switching, and nested tabs
// (Stateful, to be embedded in the main model)
type TabManager struct {
	Tabs        []Tab
	ActiveIndex int
}

func NewTabManager() *TabManager {
	return &TabManager{Tabs: []Tab{}, ActiveIndex: 0}
}

func (tm *TabManager) AddTab(tab Tab) {
	tm.Tabs = append(tm.Tabs, tab)
	tm.ActiveIndex = len(tm.Tabs) - 1
}

func (tm *TabManager) CloseTab(idx int) {
	if idx < 0 || idx >= len(tm.Tabs) {
		return
	}
	tm.Tabs = append(tm.Tabs[:idx], tm.Tabs[idx+1:]...)
	if tm.ActiveIndex >= len(tm.Tabs) {
		tm.ActiveIndex = len(tm.Tabs) - 1
	}
	if tm.ActiveIndex < 0 {
		tm.ActiveIndex = 0
	}
}

func (tm *TabManager) SwitchTab(delta int) {
	if len(tm.Tabs) == 0 {
		tm.ActiveIndex = 0
		return
	}
	tm.ActiveIndex = (tm.ActiveIndex + delta + len(tm.Tabs)) % len(tm.Tabs)
}

func (tm *TabManager) ActiveTab() *Tab {
	if len(tm.Tabs) == 0 {
		return nil
	}
	return &tm.Tabs[tm.ActiveIndex]
}

// TreeNode represents a node in the resource tree
type TreeNode struct {
	Name         string
	Type         string // "group", "resource", "folder"
	Icon         string
	Children     []*TreeNode
	Expanded     bool
	Selected     bool
	ResourceData interface{} // stores actual resource data
	Level        int         // nesting level for indentation
}

// TreeView manages the hierarchical display of resources
type TreeView struct {
	Root         *TreeNode
	SelectedPath []int // path to selected node
	ScrollOffset int
	MaxVisible   int
}

// NewTreeView creates a new tree view
func NewTreeView() *TreeView {
	return &TreeView{
		Root:         &TreeNode{Name: "Azure Resources", Type: "root", Icon: "☁️", Expanded: true},
		SelectedPath: []int{},
		ScrollOffset: 0,
		MaxVisible:   20,
	}
}

// AddResourceGroup adds a resource group to the tree
func (tv *TreeView) AddResourceGroup(name, location string) *TreeNode {
	node := &TreeNode{
		Name:     name,
		Type:     "group",
		Icon:     "🗂️",
		Children: []*TreeNode{},
		Expanded: false,
		Level:    1,
	}
	tv.Root.Children = append(tv.Root.Children, node)
	return node
}

// AddResource adds a resource to a resource group
func (tv *TreeView) AddResource(groupNode *TreeNode, name, resourceType string, data interface{}) {
	icon := GetResourceIcon(resourceType)
	resource := &TreeNode{
		Name:         name,
		Type:         "resource",
		Icon:         icon,
		Children:     []*TreeNode{},
		Expanded:     false,
		ResourceData: data,
		Level:        2,
	}
	groupNode.Children = append(groupNode.Children, resource)
}

// GetResourceIcon returns appropriate icon for resource type
func GetResourceIcon(resourceType string) string {
	icons := map[string]string{
		"Microsoft.Compute/virtualMachines":          "🖥️",
		"Microsoft.KeyVault/vaults":                  "🔑",
		"Microsoft.Storage/storageAccounts":          "💾",
		"Microsoft.Network/networkInterfaces":        "🔌",
		"Microsoft.Network/publicIPAddresses":        "🌐",
		"Microsoft.Network/virtualNetworks":          "🔗",
		"Microsoft.Compute/disks":                    "💽",
		"Microsoft.Insights/actionGroups":            "🚨",
		"Microsoft.Insights/metricAlerts":            "📊",
		"Microsoft.ContainerService/managedClusters": "🚢",
		"Microsoft.Web/sites":                        "🌐",
		"Microsoft.Sql/servers":                      "🗄️",
		"Microsoft.DocumentDB/databaseAccounts":      "📄",
	}
	if icon, exists := icons[resourceType]; exists {
		return icon
	}
	return "📦"
}

// RenderTreeView renders the tree view as a string
func (tv *TreeView) RenderTreeView(width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(1, 1)

	var lines []string
	tv.renderNode(tv.Root, &lines, 0)

	// Show loading message if tree is empty
	if len(lines) == 0 {
		lines = append(lines, "☁️ Azure Resources")
		lines = append(lines, "")
		lines = append(lines, "🔄 Loading resource groups...")
		lines = append(lines, "")
		lines = append(lines, "Press ? for help")
	}

	// Handle scrolling
	visibleLines := lines
	if len(lines) > tv.MaxVisible {
		start := tv.ScrollOffset
		end := start + tv.MaxVisible
		if end > len(lines) {
			end = len(lines)
		}
		if start < len(lines) {
			visibleLines = lines[start:end]
		}
	}

	// Add scroll indicators
	if tv.ScrollOffset > 0 {
		visibleLines = append([]string{"  ↑ More above ↑"}, visibleLines...)
	}
	if tv.ScrollOffset+tv.MaxVisible < len(lines) {
		visibleLines = append(visibleLines, "  ↓ More below ↓")
	}

	content := strings.Join(visibleLines, "\n")
	return style.Render(content)
}

// renderNode recursively renders tree nodes
func (tv *TreeView) renderNode(node *TreeNode, lines *[]string, depth int) {
	if node.Type == "root" {
		// Render root children directly
		for _, child := range node.Children {
			tv.renderNode(child, lines, depth)
		}
		return
	}

	// Create indentation
	indent := strings.Repeat("  ", depth)

	// Create expand/collapse indicator
	indicator := ""
	if len(node.Children) > 0 {
		if node.Expanded {
			indicator = "▼ "
		} else {
			indicator = "▶ "
		}
	} else {
		indicator = "  "
	}

	// Create the line
	line := fmt.Sprintf("%s%s%s %s", indent, indicator, node.Icon, node.Name)

	// Highlight if selected
	if node.Selected {
		line = lipgloss.NewStyle().
			Background(lipgloss.Color("33")).
			Foreground(lipgloss.Color("230")).
			Render(line)
	}

	*lines = append(*lines, line)

	// Render children if expanded
	if node.Expanded {
		for _, child := range node.Children {
			tv.renderNode(child, lines, depth+1)
		}
	}
}

// PowerlineSegment represents a segment in the powerline statusbar
type PowerlineSegment struct {
	Text       string
	Background lipgloss.Color
	Foreground lipgloss.Color
	Separator  string
}

// StatusBar represents a powerline-style status bar
type StatusBar struct {
	Segments   []PowerlineSegment
	RightAlign []PowerlineSegment
	Height     int
	Width      int
}

// CreatePowerlineStatusBar creates a powerline-style status bar
func CreatePowerlineStatusBar(width int) *StatusBar {
	return &StatusBar{
		Segments:   []PowerlineSegment{},
		RightAlign: []PowerlineSegment{},
		Height:     1,
		Width:      width,
	}
}

// AddSegment adds a segment to the status bar
func (sb *StatusBar) AddSegment(text string, bg, fg lipgloss.Color) {
	segment := PowerlineSegment{
		Text:       text,
		Background: bg,
		Foreground: fg,
		Separator:  "",
	}
	sb.Segments = append(sb.Segments, segment)
}

// AddRightSegment adds a right-aligned segment
func (sb *StatusBar) AddRightSegment(text string, bg, fg lipgloss.Color) {
	segment := PowerlineSegment{
		Text:       text,
		Background: bg,
		Foreground: fg,
		Separator:  "",
	}
	sb.RightAlign = append(sb.RightAlign, segment)
}

// RenderStatusBar renders the powerline status bar
func (sb *StatusBar) RenderStatusBar() string {
	var leftSide strings.Builder
	var rightSide strings.Builder

	// Render left segments
	for i, segment := range sb.Segments {
		style := lipgloss.NewStyle().
			Background(segment.Background).
			Foreground(segment.Foreground).
			Padding(0, 1)

		leftSide.WriteString(style.Render(segment.Text))

		// Add powerline separator
		if i < len(sb.Segments)-1 {
			nextBg := sb.Segments[i+1].Background
			separator := lipgloss.NewStyle().
				Background(nextBg).
				Foreground(segment.Background).
				Render("")
			leftSide.WriteString(separator)
		}
	}

	// Render right segments
	for i := len(sb.RightAlign) - 1; i >= 0; i-- {
		segment := sb.RightAlign[i]
		style := lipgloss.NewStyle().
			Background(segment.Background).
			Foreground(segment.Foreground).
			Padding(0, 1)

		// Add powerline separator before segment
		if i < len(sb.RightAlign)-1 {
			prevBg := sb.RightAlign[i+1].Background
			separator := lipgloss.NewStyle().
				Background(segment.Background).
				Foreground(prevBg).
				Render("")
			rightSide.WriteString(separator)
		}

		rightSide.WriteString(style.Render(segment.Text))
	}

	// Combine left and right with spacing
	leftStr := leftSide.String()
	rightStr := rightSide.String()

	// Calculate spaces needed
	leftWidth := lipgloss.Width(leftStr)
	rightWidth := lipgloss.Width(rightStr)
	spacesNeeded := sb.Width - leftWidth - rightWidth
	if spacesNeeded < 0 {
		spacesNeeded = 0
	}

	spaces := strings.Repeat(" ", spacesNeeded)

	return leftStr + spaces + rightStr
}

// RenderPopup renders a popup window for alarms/errors
func RenderPopup(msg PopupMsg) string {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	title := msg.Title
	if msg.Level == "error" {
		title = "❌ " + title
	} else if msg.Level == "alarm" {
		title = "⚠️  " + title
	}
	return border.Render(fmt.Sprintf("%s\n\n%s", title, msg.Content))
}

// RenderMatrixGraph renders a simple ASCII matrix/graph
func RenderMatrixGraph(msg MatrixGraphMsg) string {
	var b strings.Builder
	b.WriteString(msg.Title + "\n\n")
	if len(msg.Labels) > 0 {
		b.WriteString("    ")
		for _, label := range msg.Labels {
			b.WriteString(fmt.Sprintf("%8s", label))
		}
		b.WriteString("\n")
	}
	for i, row := range msg.Rows {
		if len(msg.Labels) > 0 && i < len(msg.Labels) {
			b.WriteString(fmt.Sprintf("%4s", msg.Labels[i]))
		} else {
			b.WriteString("    ")
		}
		for _, val := range row {
			b.WriteString(fmt.Sprintf("%8s", val))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// RenderTabs renders the tab bar and the active tab's content
func RenderTabs(tm *TabManager, status string) string {
	if tm == nil || len(tm.Tabs) == 0 {
		return "No tabs open."
	}
	var tabBar strings.Builder
	for i, tab := range tm.Tabs {
		if i == tm.ActiveIndex {
			tabBar.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Render("[" + tab.Title + "] "))
		} else {
			tabBar.WriteString(lipgloss.NewStyle().Faint(true).Render(tab.Title + " "))
		}
	}
	content := tm.Tabs[tm.ActiveIndex].Content
	statusLine := lipgloss.NewStyle().Faint(true).Render(status)
	return tabBar.String() + "\n" + content + "\n" + statusLine
}

// RenderShortcutsPopup renders a popup with all keyboard shortcuts
func RenderShortcutsPopup(shortcuts map[string]string) string {
	var b strings.Builder
	b.WriteString("Keyboard Shortcuts:\n\n")
	for k, v := range shortcuts {
		b.WriteString(fmt.Sprintf("%-8s : %s\n", k, v))
	}
	return lipgloss.NewStyle().Width(50).Height(20).Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.RoundedBorder()).Render(b.String())
}

// RenderTabsWithActive renders a tab bar with the active tab highlighted, supporting a main (non-closable) tab and resource tabs with Azure icons.
func RenderTabsWithActive(tabs []Tab, activeIdx int) string {
	if len(tabs) == 0 {
		return ""
	}

	// Azure service icons mapping - using Unicode symbols that represent Azure services
	// Inspired by https://code.benco.io/icon-collection/azure-icons/
	azureIcons := map[string]string{
		"main":               "⌂", // Home/Dashboard
		"resource":           "▤", // Resource groups
		"vm":                 "⧉", // Virtual machines
		"aks":                "⬢", // Kubernetes (hexagon)
		"storage":            "⬚", // Storage accounts
		"database":           "⛁", // SQL Database
		"network":            "⬡", // Virtual networks
		"keyvault":           "⚿", // Key vault
		"monitor":            "◉", // Monitor/metrics
		"logs":               "≡", // Log analytics
		"security":           "⛨", // Security center
		"compute":            "⚙", // Compute services
		"container":          "⬡", // Container instances
		"function":           "λ", // Azure functions
		"servicebus":         "⇄", // Service bus
		"eventhub":           "◈", // Event hubs
		"cosmosdb":           "◯", // Cosmos DB
		"redis":              "◆", // Redis cache
		"search":             "⌕", // Cognitive search
		"apimanagement":      "⚏", // API management
		"applicationgateway": "⊞", // Application gateway
		"loadbalancer":       "⚌", // Load balancer
		"publicip":           "◎", // Public IP
		"firewall":           "⚡", // Azure firewall
		"vpn":                "⟐", // VPN gateway
		"dns":                "⌘", // DNS zones
		"cdn":                "⊙", // CDN
		"backup":             "⊡", // Backup
		"recovery":           "↻", // Site recovery
		"automation":         "⚆", // Automation
		"devops":             "⚒", // DevOps
		"ml":                 "◉", // Machine learning
		"cognitive":          "⬢", // Cognitive services
		"iot":                "⬡", // IoT services
		"blockchain":         "⬢", // Blockchain
		"batch":              "▣", // Batch
		"logic":              "⬢", // Logic apps
		"analysis":           "⟐", // Analysis services
		"powerbi":            "◈", // Power BI
		"webapp":             "⊞", // Web apps
		"sqlserver":          "⛁", // SQL Server
		"postgresql":         "🐘", // PostgreSQL
		"mysql":              "🐬", // MySQL
		"mariadb":            "⚭", // MariaDB
		"appservice":         "⊞", // App Service
		"containerregistry":  "⬢", // Container Registry
		"containerinstance":  "⬡", // Container Instances
		"default":            "▫", // Default/unknown
	}

	var tabBar strings.Builder

	for i, tab := range tabs {
		// Get appropriate icon
		icon := azureIcons[tab.Type]
		if icon == "" {
			icon = azureIcons["default"]
		}

		// Style active vs inactive tabs
		var tabStyle lipgloss.Style
		if i == activeIdx {
			tabStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("39")).
				Padding(0, 1)
		} else {
			tabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Background(lipgloss.Color("0")).
				Padding(0, 1)
		}

		// Format tab title with icon
		tabTitle := fmt.Sprintf("%s %s", icon, tab.Title)
		if tab.Closable && i != 0 { // First tab (main) is not closable
			tabTitle += " ✕"
		}

		tabBar.WriteString(tabStyle.Render(tabTitle))

		// Add separator between tabs
		if i < len(tabs)-1 {
			tabBar.WriteString(" ")
		}
	}

	return tabBar.String()
}

// RenderMetricsDashboard renders an interactive dashboard for Azure resource metrics
func RenderMetricsDashboard(resourceName string, metrics map[string]interface{}) string {
	var dashboard strings.Builder

	// Header with resource name and current time
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Background(lipgloss.Color("236")).Padding(0, 2)
	dashboard.WriteString(headerStyle.Render(fmt.Sprintf("📊 Metrics Dashboard: %s", resourceName)))
	dashboard.WriteString("\n\n")

	// CPU Usage (example metric)
	if cpu, exists := metrics["cpu_usage"]; exists {
		cpuValue := fmt.Sprintf("%.1f%%", cpu)
		cpuStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		if val, ok := cpu.(float64); ok && val > 80 {
			cpuStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		}
		dashboard.WriteString(fmt.Sprintf("CPU Usage:    %s\n", cpuStyle.Render(cpuValue)))
	}

	// Memory Usage
	if mem, exists := metrics["memory_usage"]; exists {
		memValue := fmt.Sprintf("%.1f%%", mem)
		memStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		if val, ok := mem.(float64); ok && val > 85 {
			memStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		}
		dashboard.WriteString(fmt.Sprintf("Memory Usage: %s\n", memStyle.Render(memValue)))
	}

	// Network I/O
	if netIn, exists := metrics["network_in"]; exists {
		dashboard.WriteString(fmt.Sprintf("Network In:   %s MB/s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(fmt.Sprintf("%.2f", netIn))))
	}
	if netOut, exists := metrics["network_out"]; exists {
		dashboard.WriteString(fmt.Sprintf("Network Out:  %s MB/s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render(fmt.Sprintf("%.2f", netOut))))
	}

	// Disk I/O
	if diskRead, exists := metrics["disk_read"]; exists {
		dashboard.WriteString(fmt.Sprintf("Disk Read:    %s MB/s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render(fmt.Sprintf("%.2f", diskRead))))
	}
	if diskWrite, exists := metrics["disk_write"]; exists {
		dashboard.WriteString(fmt.Sprintf("Disk Write:   %s MB/s\n", lipgloss.NewStyle().Foreground(lipgloss.Color("13")).Render(fmt.Sprintf("%.2f", diskWrite))))
	}

	// Add a simple ASCII graph for trending
	dashboard.WriteString("\n")
	trendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Border(lipgloss.RoundedBorder()).Padding(1)
	trendContent := "CPU Trend (24h):\n"
	trendContent += "▁▂▃▄▅▆▇█▇▆▅▄▃▂▁▂▃▄▅▆▇█▇▆▅▄"
	dashboard.WriteString(trendStyle.Render(trendContent))

	// Add interactive controls hint
	dashboard.WriteString("\n\n")
	controlsStyle := lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("8"))
	dashboard.WriteString(controlsStyle.Render("Controls: [r]efresh • [a]lerts • [e]xport • [q]uit"))

	return dashboard.String()
}

// RenderResourceActions renders available actions for a selected resource
func RenderResourceActions(resourceType, resourceName string) string {
	var actions strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	actions.WriteString(headerStyle.Render(fmt.Sprintf("⚙️  Actions for %s: %s", resourceType, resourceName)))
	actions.WriteString("\n\n")

	// Common actions for all resources
	actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
	actions.WriteString(fmt.Sprintf("%s [v]iew details\n", actionStyle.Render("📄")))
	actions.WriteString(fmt.Sprintf("%s [m]etrics dashboard\n", actionStyle.Render("📊")))
	actions.WriteString(fmt.Sprintf("%s [l]ogs\n", actionStyle.Render("📋")))
	actions.WriteString(fmt.Sprintf("%s [e]dit configuration\n", actionStyle.Render("✏️")))
	actions.WriteString(fmt.Sprintf("%s [d]elete resource\n", actionStyle.Render("🗑️")))
	actions.WriteString(fmt.Sprintf("%s [t]erraform code\n", actionStyle.Render("🔧")))
	actions.WriteString(fmt.Sprintf("%s [b]icep code\n", actionStyle.Render("🔨")))
	actions.WriteString(fmt.Sprintf("%s [a]i analysis\n", actionStyle.Render("🤖")))

	// Resource-specific actions
	switch strings.ToLower(resourceType) {
	case "vm", "virtualmachine":
		actions.WriteString(fmt.Sprintf("%s [s]sh connect\n", actionStyle.Render("🔌")))
		actions.WriteString(fmt.Sprintf("%s [r]estart\n", actionStyle.Render("🔄")))
		actions.WriteString(fmt.Sprintf("%s [p]ower off\n", actionStyle.Render("⏻")))
	case "aks", "kubernetes":
		actions.WriteString(fmt.Sprintf("%s [k]ubectl connect\n", actionStyle.Render("☸️")))
		actions.WriteString(fmt.Sprintf("%s [n]odes status\n", actionStyle.Render("🖥️")))
		actions.WriteString(fmt.Sprintf("%s [p]ods status\n", actionStyle.Render("📦")))
	case "storage", "storageaccount":
		actions.WriteString(fmt.Sprintf("%s [f]ile browser\n", actionStyle.Render("📁")))
		actions.WriteString(fmt.Sprintf("%s [u]pload file\n", actionStyle.Render("⬆️")))
		actions.WriteString(fmt.Sprintf("%s [c]ontainers\n", actionStyle.Render("🗂️")))
	case "database", "sql":
		actions.WriteString(fmt.Sprintf("%s [q]uery editor\n", actionStyle.Render("💾")))
		actions.WriteString(fmt.Sprintf("%s [b]ackup now\n", actionStyle.Render("💿")))
		actions.WriteString(fmt.Sprintf("%s [u]sers\n", actionStyle.Render("👥")))
	}

	return actions.String()
}

// RenderEditDialog renders a dialog for editing resource configuration
func RenderEditDialog(resourceName, resourceType string, currentConfig map[string]string) string {
	var dialog strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("4")).Padding(0, 2)
	dialog.WriteString(headerStyle.Render(fmt.Sprintf("✏️  Edit %s: %s", resourceType, resourceName)))
	dialog.WriteString("\n\n")

	// Configuration fields
	for key, value := range currentConfig {
		fieldStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
		valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Background(lipgloss.Color("0")).Padding(0, 1)
		dialog.WriteString(fmt.Sprintf("%s: %s\n", fieldStyle.Render(key), valueStyle.Render(value)))
	}

	dialog.WriteString("\n")
	controlsStyle := lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("8"))
	dialog.WriteString(controlsStyle.Render("Controls: [enter]edit field • [tab]next field • [esc]cancel • [ctrl+s]save"))

	borderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1)
	return borderStyle.Render(dialog.String())
}

// RenderDeleteConfirmation renders a confirmation dialog for resource deletion
func RenderDeleteConfirmation(resourceName, resourceType string) string {
	var dialog strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Background(lipgloss.Color("1")).Padding(0, 2)
	dialog.WriteString(headerStyle.Render("⚠️  DELETE CONFIRMATION"))
	dialog.WriteString("\n\n")

	warningStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	dialog.WriteString(warningStyle.Render(fmt.Sprintf("Are you sure you want to delete this %s?", resourceType)))
	dialog.WriteString("\n\n")

	resourceStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("11"))
	dialog.WriteString(fmt.Sprintf("Resource: %s\n", resourceStyle.Render(resourceName)))
	dialog.WriteString(fmt.Sprintf("Type: %s\n\n", resourceStyle.Render(resourceType)))

	dialog.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Render("⚠️  This action cannot be undone!"))
	dialog.WriteString("\n\n")

	controlsStyle := lipgloss.NewStyle().Bold(true)
	dialog.WriteString(controlsStyle.Foreground(lipgloss.Color("1")).Render("[y]es, delete "))
	dialog.WriteString(controlsStyle.Foreground(lipgloss.Color("10")).Render("[n]o, cancel"))

	borderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("1")).Padding(1)
	return borderStyle.Render(dialog.String())
}
