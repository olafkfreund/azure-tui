package tui

import (
	"fmt"
	"sort"
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

// GetAllVisibleNodes returns all currently visible nodes in order for navigation
func (tv *TreeView) GetAllVisibleNodes() []*TreeNode {
	var nodes []*TreeNode
	tv.collectVisibleNodes(tv.Root, &nodes)
	return nodes
}

// collectVisibleNodes recursively collects all visible nodes
func (tv *TreeView) collectVisibleNodes(node *TreeNode, nodes *[]*TreeNode) {
	if node.Type == "root" {
		for _, child := range node.Children {
			tv.collectVisibleNodes(child, nodes)
		}
		return
	}

	*nodes = append(*nodes, node)

	// Include children only if expanded
	if node.Expanded {
		for _, child := range node.Children {
			tv.collectVisibleNodes(child, nodes)
		}
	}
}

// GetSelectedNode returns the currently selected node
func (tv *TreeView) GetSelectedNode() *TreeNode {
	return tv.findSelectedNode(tv.Root)
}

// findSelectedNode recursively finds the selected node
func (tv *TreeView) findSelectedNode(node *TreeNode) *TreeNode {
	if node.Selected {
		return node
	}
	for _, child := range node.Children {
		if result := tv.findSelectedNode(child); result != nil {
			return result
		}
	}
	return nil
}

// SelectNext moves selection to the next visible node
func (tv *TreeView) SelectNext() *TreeNode {
	visibleNodes := tv.GetAllVisibleNodes()
	if len(visibleNodes) == 0 {
		return nil
	}

	// Find current selection
	currentIndex := -1
	for i, node := range visibleNodes {
		if node.Selected {
			currentIndex = i
			break
		}
	}

	// Clear current selection
	tv.clearAllSelections(tv.Root)

	// Select next node (wrap around)
	nextIndex := (currentIndex + 1) % len(visibleNodes)
	visibleNodes[nextIndex].Selected = true

	// Update scroll if needed
	if nextIndex >= tv.ScrollOffset+tv.MaxVisible {
		tv.ScrollOffset = nextIndex - tv.MaxVisible + 1
	}

	return visibleNodes[nextIndex]
}

// SelectPrevious moves selection to the previous visible node
func (tv *TreeView) SelectPrevious() *TreeNode {
	visibleNodes := tv.GetAllVisibleNodes()
	if len(visibleNodes) == 0 {
		return nil
	}

	// Find current selection
	currentIndex := -1
	for i, node := range visibleNodes {
		if node.Selected {
			currentIndex = i
			break
		}
	}

	// Clear current selection
	tv.clearAllSelections(tv.Root)

	// Select previous node (wrap around)
	prevIndex := (currentIndex - 1 + len(visibleNodes)) % len(visibleNodes)
	visibleNodes[prevIndex].Selected = true

	// Update scroll if needed
	if prevIndex < tv.ScrollOffset {
		tv.ScrollOffset = prevIndex
	}

	return visibleNodes[prevIndex]
}

// clearAllSelections recursively clears all selections
func (tv *TreeView) clearAllSelections(node *TreeNode) {
	node.Selected = false
	for _, child := range node.Children {
		tv.clearAllSelections(child)
	}
}

// ToggleExpansion toggles the expansion of the currently selected node
func (tv *TreeView) ToggleExpansion() (*TreeNode, bool) {
	selectedNode := tv.GetSelectedNode()
	if selectedNode == nil || selectedNode.Type != "group" {
		return nil, false
	}

	selectedNode.Expanded = !selectedNode.Expanded
	return selectedNode, selectedNode.Expanded
}

// EnsureSelection ensures at least one node is selected
func (tv *TreeView) EnsureSelection() {
	if tv.GetSelectedNode() != nil {
		return
	}

	// Select first visible node
	visibleNodes := tv.GetAllVisibleNodes()
	if len(visibleNodes) > 0 {
		visibleNodes[0].Selected = true
	}
}

// RenderTreeView renders the tree view as a string
func (tv *TreeView) RenderTreeView(width, height int) string {
	style := lipgloss.NewStyle().
		Width(width).
		Height(height).
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

	// Ensure we always have some content
	if content == "" {
		content = "☁️ Azure Resources\n\n🔄 Loading resource groups...\n\nPress ? for help"
	}

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
	if len(sb.Segments) == 0 {
		return "🚀 Azure TUI | Loading..."
	}

	var leftSide strings.Builder
	var rightSide strings.Builder

	// Render left segments
	for i, segment := range sb.Segments {
		style := lipgloss.NewStyle().
			Background(segment.Background).
			Foreground(segment.Foreground).
			Padding(0, 1)

		leftSide.WriteString(style.Render(segment.Text))

		// Add powerline separator (simplified)
		if i < len(sb.Segments)-1 {
			leftSide.WriteString(" ")
		}
	}

	// Render right segments
	for i := len(sb.RightAlign) - 1; i >= 0; i-- {
		segment := sb.RightAlign[i]
		style := lipgloss.NewStyle().
			Background(segment.Background).
			Foreground(segment.Foreground).
			Padding(0, 1)

		rightSide.WriteString(style.Render(segment.Text))

		// Add separator between segments
		if i > 0 {
			rightSide.WriteString(" ")
		}
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
	style := lipgloss.NewStyle().Padding(1, 2)
	title := msg.Title
	if msg.Level == "error" {
		title = "❌ " + title
	} else if msg.Level == "alarm" {
		title = "⚠️  " + title
	}
	return style.Render(fmt.Sprintf("%s\n\n%s", title, msg.Content))
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
	return lipgloss.NewStyle().Width(50).Height(20).Align(lipgloss.Center, lipgloss.Center).Padding(1).Render(b.String())
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
	trendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Padding(1)
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
func RenderResourceActions(resourceType, resourceName string, actions []string) string {
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Background(lipgloss.Color("236")).Padding(0, 2)
	content.WriteString(headerStyle.Render(fmt.Sprintf("⚡ Actions: %s", resourceName)))
	content.WriteString("\n\n")

	content.WriteString("Available actions:\n\n")

	actionIcons := map[string]string{
		"start":       "▶️",
		"stop":        "⏹️",
		"restart":     "🔄",
		"ssh":         "🔐",
		"bastion":     "🏰",
		"scale":       "📈",
		"connect":     "🔗",
		"pods":        "🐳",
		"deployments": "🚀",
		"browse":      "🌐",
		"logs":        "📋",
		"backup":      "💾",
		"security":    "🔒",
		"metrics":     "📊",
		"edit":        "✏️",
		"delete":      "🗑️",
		"view":        "📄",
		"terraform":   "🔧",
		"bicep":       "🔨",
		"ai":          "🤖",
	}

	for i, action := range actions {
		icon := actionIcons[action]
		if icon == "" {
			icon = "⚙️"
		}

		actionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
		content.WriteString(fmt.Sprintf("%d. %s %s\n", i+1, icon, actionStyle.Render(strings.Title(action))))
	}

	content.WriteString("\n")
	controlsStyle := lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("8"))
	content.WriteString(controlsStyle.Render("Select action by number • Press 'q' to cancel"))

	return content.String()
}

// RenderEditDialog renders a dialog for editing resource configuration
func RenderEditDialog(resourceName, resourceType string, currentConfig map[string]string) string {
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Background(lipgloss.Color("236")).Padding(0, 2)
	content.WriteString(headerStyle.Render(fmt.Sprintf("✏️ Edit: %s", resourceName)))
	content.WriteString("\n\n")

	content.WriteString(fmt.Sprintf("Resource Type: %s\n\n", resourceType))

	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	content.WriteString(sectionStyle.Render("Current Configuration:"))
	content.WriteString("\n")

	for key, value := range currentConfig {
		keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
		content.WriteString(fmt.Sprintf("%s: %s\n", keyStyle.Render(key), value))
	}

	content.WriteString("\n")
	controlsStyle := lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("8"))
	content.WriteString(controlsStyle.Render("Use arrow keys to navigate • Enter to edit • ESC to cancel"))

	return content.String()
}

// RenderDeleteConfirmation renders a confirmation dialog for resource deletion
func RenderDeleteConfirmation(resourceName, resourceType string) string {
	style := lipgloss.NewStyle().Padding(1).Foreground(lipgloss.Color("9"))
	content := fmt.Sprintf("⚠️  Delete Resource\n\nAre you sure you want to delete:\n\nName: %s\nType: %s\n\nThis action cannot be undone!\n\nPress 'y' to confirm, 'n' to cancel", resourceName, resourceType)
	return style.Render(content)
}

// RenderStructuredResourceDetails renders comprehensive resource information
func RenderStructuredResourceDetails(details map[string]interface{}) string {
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Padding(0, 2)
	content.WriteString(headerStyle.Render("📋 Resource Details"))
	content.WriteString("\n\n")

	// Basic Information Section
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	content.WriteString(sectionStyle.Render("📍 Basic Information"))
	content.WriteString("\n")

	if name, ok := details["name"].(string); ok {
		content.WriteString(fmt.Sprintf("Name:           %s\n", name))
	}
	if resourceType, ok := details["type"].(string); ok {
		content.WriteString(fmt.Sprintf("Type:           %s\n", resourceType))
	}
	if location, ok := details["location"].(string); ok {
		content.WriteString(fmt.Sprintf("Location:       %s\n", location))
	}
	if resourceGroup, ok := details["resourceGroup"].(string); ok {
		content.WriteString(fmt.Sprintf("Resource Group: %s\n", resourceGroup))
	}
	if status, ok := details["status"].(string); ok {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		if status != "Succeeded" && status != "Running" {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		}
		content.WriteString(fmt.Sprintf("Status:         %s\n", statusStyle.Render(status)))
	}

	// Timestamps Section
	content.WriteString("\n")
	content.WriteString(sectionStyle.Render("📅 Timestamps"))
	content.WriteString("\n")

	if createdTime, ok := details["createdTime"].(string); ok && createdTime != "" {
		content.WriteString(fmt.Sprintf("Created:        %s\n", createdTime))
	}
	if modifiedTime, ok := details["modifiedTime"].(string); ok && modifiedTime != "" {
		content.WriteString(fmt.Sprintf("Last Modified:  %s\n", modifiedTime))
	}

	// Tags Section
	if tags, ok := details["tags"].(map[string]string); ok && len(tags) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🏷️  Tags"))
		content.WriteString("\n")

		for key, value := range tags {
			tagStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
			content.WriteString(fmt.Sprintf("%s: %s\n", tagStyle.Render(key), value))
		}
	}

	// SKU/Pricing Section
	if sku, ok := details["sku"].(map[string]interface{}); ok && len(sku) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("💰 SKU/Pricing"))
		content.WriteString("\n")

		for key, value := range sku {
			content.WriteString(fmt.Sprintf("%s: %v\n", strings.Title(key), value))
		}
	}

	// Properties Section (condensed)
	if properties, ok := details["properties"].(map[string]interface{}); ok && len(properties) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("⚙️  Configuration"))
		content.WriteString("\n")

		// Show only important properties to avoid clutter
		importantProps := []string{"vmSize", "osType", "provisioningState", "adminUsername", "computerName", "dnsSettings", "ipConfigurations"}
		for _, prop := range importantProps {
			if value, exists := properties[prop]; exists {
				propStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
				content.WriteString(fmt.Sprintf("%s: %s\n", propStyle.Render(strings.Title(prop)), fmt.Sprintf("%v", value)))
			}
		}
	}

	return content.String()
}

// RenderEnhancedMetricsDashboard renders real-time metrics with graphs
func RenderEnhancedMetricsDashboard(resourceName string, metrics map[string]interface{}, trends map[string][]float64) string {
	var dashboard strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Padding(0, 2)
	dashboard.WriteString(headerStyle.Render(fmt.Sprintf("📊 Live Metrics: %s", resourceName)))
	dashboard.WriteString("\n\n")

	// Current Metrics Row
	metricsStyle := lipgloss.NewStyle().Padding(1).Margin(0, 1)

	// CPU Section
	cpuContent := "🖥️  CPU Usage\n"
	if cpu, exists := metrics["cpu_usage"]; exists {
		cpuValue := fmt.Sprintf("%.1f%%", cpu)
		cpuStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		if val, ok := cpu.(float64); ok && val > 80 {
			cpuStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		} else if val, ok := cpu.(float64); ok && val > 60 {
			cpuStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		}
		cpuContent += cpuStyle.Render(cpuValue)

		// Add CPU trend graph
		if cpuTrend, exists := trends["cpu"]; exists && len(cpuTrend) > 0 {
			cpuContent += "\n" + generateTrendGraph(cpuTrend, 20, 100)
		}
	}
	dashboard.WriteString(metricsStyle.Render(cpuContent))

	// Memory Section
	memContent := "💾 Memory Usage\n"
	if mem, exists := metrics["memory_usage"]; exists {
		memValue := fmt.Sprintf("%.1f%%", mem)
		memStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		if val, ok := mem.(float64); ok && val > 85 {
			memStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		} else if val, ok := mem.(float64); ok && val > 70 {
			memStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
		}
		memContent += memStyle.Render(memValue)

		// Add memory trend graph
		if memTrend, exists := trends["memory"]; exists && len(memTrend) > 0 {
			memContent += "\n" + generateTrendGraph(memTrend, 20, 100)
		}
	}
	dashboard.WriteString(metricsStyle.Render(memContent))

	dashboard.WriteString("\n")

	// Network Section
	netContent := "🌐 Network I/O\n"
	if netIn, exists := metrics["network_in"]; exists {
		netContent += fmt.Sprintf("In:  %.2f MB/s\n", netIn)
	}
	if netOut, exists := metrics["network_out"]; exists {
		netContent += fmt.Sprintf("Out: %.2f MB/s\n", netOut)
	}
	// Add network trend graph
	if netTrend, exists := trends["network"]; exists && len(netTrend) > 0 {
		netContent += generateTrendGraph(netTrend, 20, 0) // Auto-scale for network
	}
	dashboard.WriteString(metricsStyle.Render(netContent))

	// Disk Section
	diskContent := "💿 Disk I/O\n"
	if diskRead, exists := metrics["disk_read"]; exists {
		diskContent += fmt.Sprintf("Read:  %.2f MB/s\n", diskRead)
	}
	if diskWrite, exists := metrics["disk_write"]; exists {
		diskContent += fmt.Sprintf("Write: %.2f MB/s\n", diskWrite)
	}
	// Add disk trend graph
	if diskTrend, exists := trends["disk"]; exists && len(diskTrend) > 0 {
		diskContent += generateTrendGraph(diskTrend, 20, 0) // Auto-scale for disk
	}
	dashboard.WriteString(metricsStyle.Render(diskContent))

	dashboard.WriteString("\n\n")

	// Controls and refresh info
	controlsStyle := lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("8"))
	dashboard.WriteString(controlsStyle.Render("⚡ Auto-refresh: 30s | [r]efresh now | [a]lerts | [h]istory | [q]uit"))

	return dashboard.String()
}

// RenderAKSDetails renders comprehensive AKS cluster information
func RenderAKSDetails(clusterName string, aksDetails map[string]interface{}) string {
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39")).Background(lipgloss.Color("236")).Padding(0, 2)
	content.WriteString(headerStyle.Render(fmt.Sprintf("🚢 AKS Cluster: %s", clusterName)))
	content.WriteString("\n\n")

	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))

	// Cluster Overview
	content.WriteString(sectionStyle.Render("📋 Cluster Overview"))
	content.WriteString("\n")

	if status, ok := aksDetails["status"].(string); ok {
		statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
		if status != "Running" && status != "Succeeded" {
			statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		}
		content.WriteString(fmt.Sprintf("Status:         %s\n", statusStyle.Render(status)))
	}

	if kubeVersion, ok := aksDetails["kubernetesVersion"].(string); ok {
		content.WriteString(fmt.Sprintf("Kubernetes:     %s\n", kubeVersion))
	}

	if nodeCount, ok := aksDetails["nodeCount"].(int); ok {
		content.WriteString(fmt.Sprintf("Total Nodes:    %d\n", nodeCount))
	}

	// Node Pools
	if nodePools, ok := aksDetails["nodePools"].([]interface{}); ok && len(nodePools) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🖥️  Node Pools"))
		content.WriteString("\n")

		for _, pool := range nodePools {
			if poolMap, ok := pool.(map[string]interface{}); ok {
				name := poolMap["name"].(string)
				count := poolMap["count"].(int)
				vmSize := poolMap["vmSize"].(string)
				osType := poolMap["osType"].(string)

				poolStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14"))
				content.WriteString(fmt.Sprintf("%s: %d × %s (%s)\n", poolStyle.Render(name), count, vmSize, osType))
			}
		}
	}

	// Pods Summary
	if pods, ok := aksDetails["pods"].([]interface{}); ok {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🐳 Pods Summary"))
		content.WriteString("\n")

		podCounts := make(map[string]int)
		nsCounts := make(map[string]int)

		for _, pod := range pods {
			if podMap, ok := pod.(map[string]interface{}); ok {
				status := podMap["status"].(string)
				namespace := podMap["namespace"].(string)
				podCounts[status]++
				nsCounts[namespace]++
			}
		}

		content.WriteString(fmt.Sprintf("Total Pods:     %d\n", len(pods)))

		for status, count := range podCounts {
			statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
			if status != "Running" {
				statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
			}
			content.WriteString(fmt.Sprintf("%s: %s\n", statusStyle.Render(strings.Title(status)), fmt.Sprintf("%d", count)))
		}

		content.WriteString("\nTop Namespaces:\n")
		for ns, count := range nsCounts {
			if count > 1 { // Only show namespaces with multiple pods
				nsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
				content.WriteString(fmt.Sprintf("%s: %d pods\n", nsStyle.Render(ns), count))
			}
		}
	}

	// Deployments Summary
	if deployments, ok := aksDetails["deployments"].([]interface{}); ok && len(deployments) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🚀 Deployments"))
		content.WriteString("\n")

		content.WriteString(fmt.Sprintf("Total Deployments: %d\n", len(deployments)))

		// Show first few deployments
		for i, deploy := range deployments {
			if i >= 5 { // Limit to first 5
				content.WriteString(fmt.Sprintf("... and %d more\n", len(deployments)-5))
				break
			}

			if deployMap, ok := deploy.(map[string]interface{}); ok {
				name := deployMap["name"].(string)
				namespace := deployMap["namespace"].(string)
				ready := deployMap["ready"].(string)

				deployStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("13"))
				content.WriteString(fmt.Sprintf("%s (%s): %s\n", deployStyle.Render(name), namespace, ready))
			}
		}
	}

	// Services Summary
	if services, ok := aksDetails["services"].([]interface{}); ok && len(services) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🔗 Services"))
		content.WriteString("\n")

		typeCounts := make(map[string]int)
		for _, svc := range services {
			if svcMap, ok := svc.(map[string]interface{}); ok {
				svcType := svcMap["type"].(string)
				typeCounts[svcType]++
			}
		}

		content.WriteString(fmt.Sprintf("Total Services: %d\n", len(services)))
		for svcType, count := range typeCounts {
			typeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
			content.WriteString(fmt.Sprintf("%s: %d\n", typeStyle.Render(svcType), count))
		}
	}

	return content.String()
}

// Helper function to generate ASCII trend graphs
func generateTrendGraph(data []float64, width int, maxValue float64) string {
	if len(data) == 0 {
		return ""
	}

	// Auto-scale if maxValue is 0
	if maxValue == 0 {
		for _, val := range data {
			if val > maxValue {
				maxValue = val
			}
		}
	}

	if maxValue == 0 {
		maxValue = 1 // Avoid division by zero
	}

	// Create trend line
	blocks := []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}
	var trend strings.Builder

	// Sample data to fit width
	step := len(data) / width
	if step < 1 {
		step = 1
	}

	for i := 0; i < width && i*step < len(data); i++ {
		value := data[i*step]
		level := int((value / maxValue) * float64(len(blocks)-1))
		if level >= len(blocks) {
			level = len(blocks) - 1
		}
		if level < 0 {
			level = 0
		}
		trend.WriteString(blocks[level])
	}

	return trend.String()
}

// TableData represents data for table formatting
type TableData struct {
	Headers []string
	Rows    [][]string
	Title   string
}

// formatPropertyName formats property names for display
func formatPropertyName(prop string) string {
	// Convert camelCase to Title Case
	result := ""
	for i, r := range prop {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result += " "
		}
		if i == 0 {
			result += strings.ToUpper(string(r))
		} else {
			result += string(r)
		}
	}
	return result
}

// formatValue formats interface{} values for display
func formatValue(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case string:
		return v
	case bool:
		if v {
			return "✓ Yes"
		}
		return "✗ No"
	case float64:
		return fmt.Sprintf("%.2f", v)
	case int:
		return fmt.Sprintf("%d", v)
	case []interface{}:
		return fmt.Sprintf("Array (%d items)", len(v))
	case map[string]interface{}:
		return fmt.Sprintf("Object (%d properties)", len(v))
	default:
		str := fmt.Sprintf("%v", v)
		if len(str) > 100 {
			return str[:97] + "..."
		}
		return str
	}
}

// RenderTable renders data in a clean, borderless table format
func RenderTable(data TableData) string {
	if len(data.Rows) == 0 {
		return ""
	}

	var content strings.Builder

	// Title
	if data.Title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		content.WriteString(titleStyle.Render(data.Title))
		content.WriteString("\n\n")
	}

	// Calculate column widths
	colWidths := make([]int, len(data.Headers))
	for i, header := range data.Headers {
		colWidths[i] = len(header)
	}

	for _, row := range data.Rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Add padding between columns
	for i := range colWidths {
		colWidths[i] += 4 // More spacing for better readability
	}

	// Render headers
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("33"))
	for i, header := range data.Headers {
		content.WriteString(headerStyle.Render(fmt.Sprintf("%-*s", colWidths[i], header)))
	}
	content.WriteString("\n")

	// Simple underline for headers (just spaces for clean look)
	content.WriteString("\n")

	// Render rows
	for _, row := range data.Rows {
		for i, cell := range row {
			if i < len(colWidths) {
				content.WriteString(fmt.Sprintf("%-*s", colWidths[i], cell))
			}
		}
		content.WriteString("\n")
	}

	return content.String()
}

// FormatPropertiesAsTable formats resource properties as a table
func FormatPropertiesAsTable(properties map[string]interface{}) string {
	if len(properties) == 0 {
		return ""
	}

	tableData := TableData{
		Headers: []string{"Property", "Value"},
		Title:   "⚙️  Configuration Properties",
	}

	// Convert properties to table rows
	for key, value := range properties {
		formattedKey := formatPropertyName(key)
		formattedValue := formatValue(value)

		// Truncate long values for table display
		if len(formattedValue) > 50 {
			formattedValue = formattedValue[:47] + "..."
		}

		tableData.Rows = append(tableData.Rows, []string{formattedKey, formattedValue})
	}

	// Sort rows by property name for consistency
	sort.Slice(tableData.Rows, func(i, j int) bool {
		return tableData.Rows[i][0] < tableData.Rows[j][0]
	})

	return RenderTable(tableData)
}

// RenderSimpleList renders data as a clean property list (alternative to table)
func RenderSimpleList(data TableData) string {
	if len(data.Rows) == 0 {
		return ""
	}

	var content strings.Builder

	// Title
	if data.Title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
		content.WriteString(titleStyle.Render(data.Title))
		content.WriteString("\n\n")
	}

	// Find the longest property name for alignment
	maxKeyLength := 0
	for _, row := range data.Rows {
		if len(row) > 0 && len(row[0]) > maxKeyLength {
			maxKeyLength = len(row[0])
		}
	}

	// Render rows as property: value pairs
	keyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")).Bold(true)
	valueStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7"))

	for _, row := range data.Rows {
		if len(row) >= 2 {
			key := fmt.Sprintf("%-*s", maxKeyLength, row[0])
			content.WriteString(fmt.Sprintf("%s: %s\n",
				keyStyle.Render(key),
				valueStyle.Render(row[1])))
		}
	}

	return content.String()
}

// FormatPropertiesAsSimpleList formats resource properties as a simple list
func FormatPropertiesAsSimpleList(properties map[string]interface{}) string {
	if len(properties) == 0 {
		return ""
	}

	tableData := TableData{
		Title: "⚙️  Configuration Properties",
	}

	// Convert properties to rows
	for key, value := range properties {
		formattedKey := formatPropertyName(key)
		formattedValue := formatValue(value)

		// Truncate long values for display
		if len(formattedValue) > 60 {
			formattedValue = formattedValue[:57] + "..."
		}

		tableData.Rows = append(tableData.Rows, []string{formattedKey, formattedValue})
	}

	// Sort rows by property name for consistency
	sort.Slice(tableData.Rows, func(i, j int) bool {
		return tableData.Rows[i][0] < tableData.Rows[j][0]
	})

	return RenderSimpleList(tableData)
}
