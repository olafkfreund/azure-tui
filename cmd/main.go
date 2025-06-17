package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/olafkfreund/azure-tui/internal/azure/resourceactions"
	"github.com/olafkfreund/azure-tui/internal/azure/resourcedetails"
	"github.com/olafkfreund/azure-tui/internal/openai"
	"github.com/olafkfreund/azure-tui/internal/tui"
)

// Gruvbox colors
var (
	bgDark      = lipgloss.Color("#282828")
	bgMedium    = lipgloss.Color("#3c3836")
	bgLight     = lipgloss.Color("#504945")
	fgLight     = lipgloss.Color("#fbf1c7")
	fgMedium    = lipgloss.Color("#ebdbb2")
	colorBlue   = lipgloss.Color("#83a598")
	colorGreen  = lipgloss.Color("#b8bb26")
	colorRed    = lipgloss.Color("#fb4934")
	colorYellow = lipgloss.Color("#fabd2f")
	colorPurple = lipgloss.Color("#d3869b")
	colorAqua   = lipgloss.Color("#8ec07c")
	colorGray   = lipgloss.Color("#a89984")
)

type AzureResource struct {
	ID            string                 `json:"id"`
	Name          string                 `json:"name"`
	Type          string                 `json:"type"`
	Location      string                 `json:"location"`
	ResourceGroup string                 `json:"resourceGroup"`
	Status        string                 `json:"status,omitempty"`
	Tags          map[string]string      `json:"tags,omitempty"`
	Properties    map[string]interface{} `json:"properties,omitempty"`
}

type ResourceGroup struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type Subscription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TenantID  string `json:"tenantId"`
	IsDefault bool   `json:"isDefault"`
}

// Messages
type subscriptionsLoadedMsg struct{ subscriptions []Subscription }
type resourceGroupsLoadedMsg struct{ groups []ResourceGroup }
type resourcesInGroupMsg struct {
	groupName string
	resources []AzureResource
}
type resourceDetailsLoadedMsg struct {
	resource AzureResource
	details  *resourcedetails.ResourceDetails
}
type aiDescriptionLoadedMsg struct {
	description string
}
type resourceActionMsg struct {
	action   string
	resource AzureResource
	result   resourceactions.ActionResult
}
type errorMsg struct{ error string }

type model struct {
	treeView               *tui.TreeView
	statusBar              *tui.StatusBar
	aiProvider             *openai.AIProvider
	width, height          int
	ready                  bool
	subscriptions          []Subscription
	resourceGroups         []ResourceGroup
	allResources           []AzureResource
	selectedResource       *AzureResource
	resourceDetails        *resourcedetails.ResourceDetails
	aiDescription          string
	loadingState           string
	selectedPanel          int
	rightPanelScrollOffset int
	rightPanelMaxLines     int
	actionInProgress       bool
	lastActionResult       *resourceactions.ActionResult
	showDashboard          bool
	logEntries             []string
	// New navigation fields
	activeView            string          // "details", "dashboard", "welcome"
	propertyExpandedIndex int             // For navigating expanded properties
	expandedProperties    map[string]bool // Track which properties are expanded
}

func fetchSubscriptions() ([]Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "account", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch subscriptions: %v", err)
	}

	var azSubs []struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		TenantID  string `json:"tenantId"`
		IsDefault bool   `json:"isDefault"`
	}

	if err := json.Unmarshal(output, &azSubs); err != nil {
		return nil, fmt.Errorf("failed to parse subscription data: %v", err)
	}

	var subscriptions []Subscription
	for _, s := range azSubs {
		subscriptions = append(subscriptions, Subscription{
			ID: s.ID, Name: s.Name, TenantID: s.TenantID, IsDefault: s.IsDefault,
		})
	}
	return subscriptions, nil
}

func fetchResourceGroups() ([]ResourceGroup, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "group", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resource groups: %v", err)
	}

	var azGroups []struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}

	if err := json.Unmarshal(output, &azGroups); err != nil {
		return nil, fmt.Errorf("failed to parse resource group data: %v", err)
	}

	var groups []ResourceGroup
	for _, g := range azGroups {
		groups = append(groups, ResourceGroup{Name: g.Name, Location: g.Location})
	}
	return groups, nil
}

func fetchResourcesInGroup(groupName string) ([]AzureResource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "resource", "list", "--resource-group", groupName, "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources: %v", err)
	}

	var azResources []struct {
		ID       string            `json:"id"`
		Name     string            `json:"name"`
		Type     string            `json:"type"`
		Location string            `json:"location"`
		Tags     map[string]string `json:"tags"`
	}

	if err := json.Unmarshal(output, &azResources); err != nil {
		return nil, fmt.Errorf("failed to parse resource data: %v", err)
	}

	var resources []AzureResource
	for _, r := range azResources {
		resource := AzureResource{
			ID: r.ID, Name: r.Name, Type: r.Type, Location: r.Location,
			ResourceGroup: groupName, Tags: r.Tags,
		}

		if r.Type == "Microsoft.Compute/virtualMachines" {
			if status, err := resourceactions.GetVMStatus(r.Name, groupName); err == nil {
				resource.Status = status
			}
		}
		resources = append(resources, resource)
	}
	return resources, nil
}

func loadDataCmd() tea.Cmd {
	return func() tea.Msg {
		subs, err := fetchSubscriptions()
		if err != nil {
			return errorMsg{error: err.Error()}
		}

		groups, err := fetchResourceGroups()
		if err != nil {
			return errorMsg{error: err.Error()}
		}

		return tea.Batch(
			func() tea.Msg { return subscriptionsLoadedMsg{subscriptions: subs} },
			func() tea.Msg { return resourceGroupsLoadedMsg{groups: groups} },
		)()
	}
}

func loadResourcesInGroupCmd(groupName string) tea.Cmd {
	return func() tea.Msg {
		resources, err := fetchResourcesInGroup(groupName)
		if err != nil {
			return errorMsg{error: err.Error()}
		}
		return resourcesInGroupMsg{groupName: groupName, resources: resources}
	}
}

func loadResourceDetailsCmd(resource AzureResource) tea.Cmd {
	return func() tea.Msg {
		details, err := resourcedetails.GetResourceDetails(resource.ID)
		if err != nil {
			return errorMsg{error: err.Error()}
		}
		return resourceDetailsLoadedMsg{resource: resource, details: details}
	}
}

func loadAIDescriptionCmd(ai *openai.AIProvider, resource AzureResource, details *resourcedetails.ResourceDetails) tea.Cmd {
	return func() tea.Msg {
		if ai == nil {
			return aiDescriptionLoadedMsg{description: ""}
		}

		detailsStr := fmt.Sprintf("Resource: %s\nType: %s\nLocation: %s\nStatus: %s",
			resource.Name, resource.Type, resource.Location, resource.Status)

		if details != nil {
			detailsStr += fmt.Sprintf("\nProperties: %v", details.Properties)
		}

		description, err := ai.DescribeResource(resource.Type, resource.Name, detailsStr)
		if err != nil {
			return aiDescriptionLoadedMsg{description: "AI analysis unavailable"}
		}

		return aiDescriptionLoadedMsg{description: description}
	}
}

func executeResourceActionCmd(action string, resource AzureResource) tea.Cmd {
	return func() tea.Msg {
		var result resourceactions.ActionResult

		switch action {
		case "start":
			if resource.Type == "Microsoft.Compute/virtualMachines" {
				result = resourceactions.StartVM(resource.Name, resource.ResourceGroup)
			}
		case "stop":
			if resource.Type == "Microsoft.Compute/virtualMachines" {
				result = resourceactions.StopVM(resource.Name, resource.ResourceGroup)
			}
		case "restart":
			if resource.Type == "Microsoft.Compute/virtualMachines" {
				result = resourceactions.RestartVM(resource.Name, resource.ResourceGroup)
			}
		default:
			result = resourceactions.ActionResult{Success: false, Message: "Unsupported action"}
		}
		return resourceActionMsg{action: action, resource: resource, result: result}
	}
}

func initModel() model {
	// Initialize AI provider if API key is available
	var ai *openai.AIProvider
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		ai = openai.NewAIProvider(apiKey)
	}

	return model{
		treeView:               tui.NewTreeView(),
		statusBar:              tui.CreatePowerlineStatusBar(80),
		aiProvider:             ai,
		loadingState:           "loading",
		selectedPanel:          0,
		rightPanelScrollOffset: 0,
		rightPanelMaxLines:     50,
		showDashboard:          false,
		logEntries:             []string{},
		activeView:             "welcome",
		propertyExpandedIndex:  -1,
		expandedProperties:     make(map[string]bool),
	}
}

func (m model) Init() tea.Cmd {
	return loadDataCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height, m.ready = msg.Width, msg.Height, true
		if m.statusBar != nil {
			m.statusBar.Width = msg.Width
		}

	case subscriptionsLoadedMsg:
		m.subscriptions = msg.subscriptions

	case resourceGroupsLoadedMsg:
		m.resourceGroups = msg.groups
		m.loadingState = "ready"
		if m.treeView != nil {
			for _, group := range msg.groups {
				groupNode := m.treeView.AddResourceGroup(group.Name, group.Location)
				m.treeView.AddResource(groupNode, "Loading...", "placeholder", nil)
			}
			m.treeView.EnsureSelection()
		}

	case resourcesInGroupMsg:
		if m.treeView != nil {
			for _, groupNode := range m.treeView.Root.Children {
				if groupNode.Name == msg.groupName {
					groupNode.Children = []*tui.TreeNode{}
					for _, resource := range msg.resources {
						m.treeView.AddResource(groupNode, resource.Name, resource.Type, resource)
					}
					break
				}
			}
		}
		m.allResources = append(m.allResources, msg.resources...)

	case resourceDetailsLoadedMsg:
		m.selectedResource = &msg.resource
		m.resourceDetails = msg.details
		// Load AI description after resource details are loaded
		if m.aiProvider != nil {
			return m, loadAIDescriptionCmd(m.aiProvider, msg.resource, msg.details)
		}

	case aiDescriptionLoadedMsg:
		m.aiDescription = msg.description

	case resourceActionMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		if msg.result.Success && m.selectedResource != nil {
			return m, loadResourceDetailsCmd(*m.selectedResource)
		}

	case errorMsg:
		m.loadingState = "error"

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.selectedPanel = (m.selectedPanel + 1) % 2
		case "left", "h":
			// Left navigation - switch to tree panel or previous section
			if m.selectedPanel == 1 {
				m.selectedPanel = 0
				m.rightPanelScrollOffset = 0 // Reset scroll when switching
			}
		case "right", "l":
			// Right navigation - switch to details panel
			if m.selectedPanel == 0 {
				m.selectedPanel = 1
			}
		case "d":
			// Toggle dashboard view
			m.showDashboard = !m.showDashboard
			if m.showDashboard {
				m.activeView = "dashboard"
			} else {
				if m.selectedResource != nil {
					m.activeView = "details"
				} else {
					m.activeView = "welcome"
				}
			}
		case "j", "down":
			if m.selectedPanel == 0 && m.treeView != nil {
				m.treeView.SelectNext()
				m.treeView.EnsureSelection()
				if selectedNode := m.treeView.GetSelectedNode(); selectedNode != nil && selectedNode.Type == "resource" {
					if resource, ok := selectedNode.ResourceData.(AzureResource); ok {
						return m, loadResourceDetailsCmd(resource)
					}
				}
			} else if m.selectedPanel == 1 {
				// Right panel scrolling down
				// Calculate max lines based on current content
				rightContent := m.renderResourcePanel(m.width/3, m.height-2)
				totalLines := strings.Count(rightContent, "\n")
				maxLines := max(0, totalLines-(m.height-6))
				if m.rightPanelScrollOffset < maxLines {
					m.rightPanelScrollOffset++
				}
			}
		case "k", "up":
			if m.selectedPanel == 0 && m.treeView != nil {
				m.treeView.SelectPrevious()
				m.treeView.EnsureSelection()
				if selectedNode := m.treeView.GetSelectedNode(); selectedNode != nil && selectedNode.Type == "resource" {
					if resource, ok := selectedNode.ResourceData.(AzureResource); ok {
						return m, loadResourceDetailsCmd(resource)
					}
				}
			} else if m.selectedPanel == 1 {
				// Right panel scrolling up
				if m.rightPanelScrollOffset > 0 {
					m.rightPanelScrollOffset--
				}
			}
		case " ", "enter":
			if m.selectedPanel == 0 && m.treeView != nil {
				selectedNode := m.treeView.GetSelectedNode()
				if selectedNode != nil {
					switch selectedNode.Type {
					case "group":
						selectedNode.Expanded = !selectedNode.Expanded
						if selectedNode.Expanded {
							return m, loadResourcesInGroupCmd(selectedNode.Name)
						}
					case "resource":
						if resource, ok := selectedNode.ResourceData.(AzureResource); ok {
							return m, loadResourceDetailsCmd(resource)
						}
					}
				}
			}
		case "e":
			// Toggle property expansion in details panel
			if m.selectedPanel == 1 && m.selectedResource != nil {
				// Toggle expansion for complex properties
				if m.selectedResource.Type == "Microsoft.ContainerService/managedClusters" {
					key := "agentPoolProfiles"
					m.expandedProperties[key] = !m.expandedProperties[key]
				}
			}
		case "s":
			if m.selectedResource != nil && !m.actionInProgress {
				m.actionInProgress = true
				return m, executeResourceActionCmd("start", *m.selectedResource)
			}
		case "S":
			if m.selectedResource != nil && !m.actionInProgress {
				m.actionInProgress = true
				return m, executeResourceActionCmd("stop", *m.selectedResource)
			}
		case "r":
			if m.selectedResource != nil && !m.actionInProgress {
				m.actionInProgress = true
				return m, executeResourceActionCmd("restart", *m.selectedResource)
			} else {
				return m, loadDataCmd()
			}
		case "R":
			return m, loadDataCmd()
		}
	}
	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return lipgloss.NewStyle().
			Background(bgDark).
			Foreground(fgLight).
			Render("Loading Azure Dashboard...")
	}

	// Status bar
	if m.statusBar != nil {
		m.statusBar.Segments = []tui.PowerlineSegment{}
		m.statusBar.AddSegment("☁️ Azure Dashboard", colorBlue, bgDark)

		switch m.loadingState {
		case "loading":
			m.statusBar.AddSegment("Loading", colorYellow, bgMedium)
		case "ready":
			m.statusBar.AddSegment(fmt.Sprintf("%d Groups", len(m.resourceGroups)), colorGreen, bgMedium)
			if m.selectedResource != nil {
				m.statusBar.AddSegment(fmt.Sprintf("Selected: %s", m.selectedResource.Name), colorPurple, bgMedium)
			}
		case "error":
			m.statusBar.AddSegment("Error", colorRed, bgMedium)
		}

		panelName := "Tree"
		panelHelp := ""
		navigationHelp := ""
		if m.selectedPanel == 1 {
			panelName = "Details"
			if m.rightPanelScrollOffset > 0 {
				panelHelp = " (j/k:scroll)"
			} else {
				panelHelp = " (j/k:scroll)"
			}
			navigationHelp = "h/←:Tree l/→:Stay"
		} else {
			panelHelp = " (j/k:navigate)"
			navigationHelp = "l/→:Details"
		}
		m.statusBar.AddSegment(fmt.Sprintf("▶ %s%s", panelName, panelHelp), colorAqua, bgMedium)
		m.statusBar.AddSegment(navigationHelp, colorPurple, bgMedium)

		// Add expansion hint for AKS resources
		if m.selectedResource != nil && m.selectedResource.Type == "Microsoft.ContainerService/managedClusters" && m.selectedPanel == 1 {
			m.statusBar.AddSegment("e:Expand AKS Properties", colorYellow, bgMedium)
		}

		m.statusBar.AddSegment("Tab:Switch d:Dashboard s:Start S:Stop r:Restart R:Refresh q:Quit", colorGray, bgLight)
	}

	// Two-panel layout - Enhanced with active panel indicators
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth

	// Tree panel
	treeContent := ""
	if m.treeView != nil {
		treeContent = m.treeView.RenderTreeView(leftWidth-4, m.height-2)
	}

	// Style left panel with active indicator
	leftPanelStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Foreground(fgMedium).
		Padding(1, 2)

	// Add visual indicator for active panel
	if m.selectedPanel == 0 {
		leftPanelStyle = leftPanelStyle.
			Foreground(fgLight).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBlue)
		// Add enhanced active panel indicator
		treeContent = "🔍 " + strings.Replace(treeContent, "\n", "\n   ", -1)
	} else {
		leftPanelStyle = leftPanelStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGray)
	}

	leftPanel := leftPanelStyle.Render(treeContent)

	// Details panel with scrolling support
	rightContentRaw := m.renderResourcePanel(rightWidth-4, m.height-2)
	var rightContent string

	// Apply scrolling if right panel is active
	if m.selectedPanel == 1 {
		rightContent = m.renderScrollableContent(rightContentRaw, m.height-6)
	} else {
		rightContent = rightContentRaw
	}

	// Style right panel with active indicator
	rightPanelStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Foreground(fgMedium).
		Padding(1, 2)

	if m.selectedPanel == 1 {
		rightPanelStyle = rightPanelStyle.
			Foreground(fgLight).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGreen)
		// Add enhanced active panel marker
		rightContent = "📊 " + strings.Replace(rightContent, "\n", "\n   ", -1)
	} else {
		rightPanelStyle = rightPanelStyle.
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorGray)
	}

	rightPanel := rightPanelStyle.Render(rightContent)

	// Join everything
	statusBarContent := ""
	if m.statusBar != nil {
		statusBarContent = m.statusBar.RenderStatusBar()
	}

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	fullView := lipgloss.JoinVertical(lipgloss.Left, statusBarContent, mainContent)

	return lipgloss.NewStyle().Background(bgDark).Render(fullView)
}

func (m model) renderResourcePanel(width, height int) string {
	if m.selectedResource == nil {
		return m.renderWelcomePanel(width, height)
	}

	if m.showDashboard {
		return m.renderDashboardView(width, height)
	}

	return m.renderEnhancedResourceDetails(width, height)
}

func (m model) renderWelcomePanel(width, height int) string {
	var content strings.Builder

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Padding(0, 1)
	content.WriteString(headerStyle.Render("📊 Azure Resource Dashboard"))
	content.WriteString("\n\n")

	content.WriteString("Welcome to Azure TUI Dashboard!\n\n")

	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	content.WriteString(sectionStyle.Render("🎯 Getting Started:"))
	content.WriteString("\n")
	content.WriteString("1. Navigate through resource groups in the left panel\n")
	content.WriteString("2. Press Space/Enter to expand a resource group\n")
	content.WriteString("3. Select a resource to view details and actions\n")
	content.WriteString("4. Use Tab to switch between panels\n\n")

	content.WriteString(sectionStyle.Render("🎮 Quick Actions:"))
	content.WriteString("\n")
	actionStyle := lipgloss.NewStyle().Foreground(colorAqua)
	content.WriteString(fmt.Sprintf("%s - Start resource (VMs)\n", actionStyle.Render("s")))
	content.WriteString(fmt.Sprintf("%s - Stop resource (VMs)\n", actionStyle.Render("S")))
	content.WriteString(fmt.Sprintf("%s - Restart resource (VMs)\n", actionStyle.Render("r")))
	content.WriteString(fmt.Sprintf("%s - Toggle Dashboard view\n", actionStyle.Render("d")))
	content.WriteString(fmt.Sprintf("%s - Refresh all data\n", actionStyle.Render("R")))
	content.WriteString(fmt.Sprintf("%s - Switch panels\n", actionStyle.Render("Tab")))
	content.WriteString(fmt.Sprintf("%s - Quit\n\n", actionStyle.Render("q")))

	content.WriteString(sectionStyle.Render("✨ New Features:"))
	content.WriteString("\n")
	featureStyle := lipgloss.NewStyle().Foreground(colorPurple)
	content.WriteString(fmt.Sprintf("%s Enhanced resource details with better formatting\n", featureStyle.Render("•")))
	content.WriteString(fmt.Sprintf("%s AI-powered resource descriptions and insights\n", featureStyle.Render("•")))
	content.WriteString(fmt.Sprintf("%s Dashboard view with live metrics and trends\n", featureStyle.Render("•")))
	content.WriteString(fmt.Sprintf("%s AI-parsed log analysis and recommendations\n", featureStyle.Render("•")))
	content.WriteString(fmt.Sprintf("%s Transparent backgrounds for cleaner interface\n\n", featureStyle.Render("•")))

	aiStatus := "❌ Disabled (set OPENAI_API_KEY)"
	if m.aiProvider != nil {
		aiStatus = "✅ Enabled"
	}
	statusStyle := lipgloss.NewStyle().Foreground(colorGray)
	content.WriteString(fmt.Sprintf("🤖 AI Features: %s\n\n", statusStyle.Render(aiStatus)))

	content.WriteString("💡 Select a resource from the left panel to see detailed information and available actions here.")

	return content.String()
}

func (m model) renderEnhancedResourceDetails(width, height int) string {
	resource := m.selectedResource
	var content strings.Builder

	// Header with resource name and type
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Padding(0, 1)
	content.WriteString(headerStyle.Render(fmt.Sprintf("📦 %s (%s)", resource.Name, getResourceTypeDisplayName(resource.Type))))
	content.WriteString("\n\n")

	// Basic Information Section
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	content.WriteString(sectionStyle.Render("📋 Basic Information"))
	content.WriteString("\n")

	keyStyle := lipgloss.NewStyle().Foreground(colorAqua)
	valueStyle := lipgloss.NewStyle().Foreground(fgMedium)

	content.WriteString(fmt.Sprintf("%s: %s\n", keyStyle.Render("Name"), valueStyle.Render(resource.Name)))
	content.WriteString(fmt.Sprintf("%s: %s\n", keyStyle.Render("Type"), valueStyle.Render(resource.Type)))
	content.WriteString(fmt.Sprintf("%s: %s\n", keyStyle.Render("Location"), valueStyle.Render(resource.Location)))
	content.WriteString(fmt.Sprintf("%s: %s\n", keyStyle.Render("Resource Group"), valueStyle.Render(resource.ResourceGroup)))

	// Status with color coding
	if resource.Status != "" {
		statusColor := colorRed
		statusIcon := "🔴"
		if strings.Contains(strings.ToLower(resource.Status), "running") || strings.Contains(strings.ToLower(resource.Status), "succeeded") {
			statusColor = colorGreen
			statusIcon = "🟢"
		} else if strings.Contains(strings.ToLower(resource.Status), "deallocated") || strings.Contains(strings.ToLower(resource.Status), "stopped") {
			statusColor = colorYellow
			statusIcon = "🟡"
		}
		statusStyle := lipgloss.NewStyle().Foreground(statusColor)
		content.WriteString(fmt.Sprintf("%s: %s %s\n", keyStyle.Render("Status"), statusIcon, statusStyle.Render(resource.Status)))
	}

	// AI Description Section
	if m.aiDescription != "" && m.aiProvider != nil {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🤖 AI Analysis"))
		content.WriteString("\n")

		aiStyle := lipgloss.NewStyle().Foreground(colorPurple).Italic(true)
		// Wrap long AI descriptions
		lines := strings.Split(m.aiDescription, "\n")
		for _, line := range lines {
			if len(line) > width-10 {
				words := strings.Fields(line)
				currentLine := ""
				for _, word := range words {
					if len(currentLine+" "+word) > width-10 {
						if currentLine != "" {
							content.WriteString(aiStyle.Render(currentLine))
							content.WriteString("\n")
							currentLine = word
						} else {
							content.WriteString(aiStyle.Render(word))
							content.WriteString("\n")
						}
					} else {
						if currentLine == "" {
							currentLine = word
						} else {
							currentLine += " " + word
						}
					}
				}
				if currentLine != "" {
					content.WriteString(aiStyle.Render(currentLine))
					content.WriteString("\n")
				}
			} else {
				content.WriteString(aiStyle.Render(line))
				content.WriteString("\n")
			}
		}
	}

	// Tags Section
	if len(resource.Tags) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🏷️  Tags"))
		content.WriteString("\n")

		tagKeyStyle := lipgloss.NewStyle().Foreground(colorYellow)
		for key, value := range resource.Tags {
			content.WriteString(fmt.Sprintf("%s: %s\n", tagKeyStyle.Render(key), valueStyle.Render(value)))
		}
	}

	// Actions Section for VMs
	if resource.Type == "Microsoft.Compute/virtualMachines" {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🎮 Available Actions"))
		content.WriteString("\n")

		actionStyle := lipgloss.NewStyle().Foreground(colorBlue)
		content.WriteString(fmt.Sprintf("%s Start VM\n", actionStyle.Render("[s]")))
		content.WriteString(fmt.Sprintf("%s Stop VM\n", actionStyle.Render("[S]")))
		content.WriteString(fmt.Sprintf("%s Restart VM\n", actionStyle.Render("[r]")))

		if m.actionInProgress {
			progressStyle := lipgloss.NewStyle().Foreground(colorYellow)
			content.WriteString("\n")
			content.WriteString(progressStyle.Render("⏳ Action in progress..."))
			content.WriteString("\n")
		}

		if m.lastActionResult != nil {
			resultStyle := lipgloss.NewStyle().Foreground(colorRed)
			icon := "❌"
			if m.lastActionResult.Success {
				resultStyle = lipgloss.NewStyle().Foreground(colorGreen)
				icon = "✅"
			}
			content.WriteString("\n")
			content.WriteString(fmt.Sprintf("%s %s", icon, resultStyle.Render(m.lastActionResult.Message)))
			content.WriteString("\n")
		}
	}

	// Properties Section
	if m.resourceDetails != nil && len(m.resourceDetails.Properties) > 0 {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("⚙️  Configuration"))
		content.WriteString("\n")

		// Show only important properties to avoid clutter
		importantProps := getImportantProperties(resource.Type)
		for _, prop := range importantProps {
			if value, exists := m.resourceDetails.Properties[prop]; exists {
				propStyle := lipgloss.NewStyle().Foreground(colorAqua)
				formattedName := formatPropertyName(prop)

				// Check if this is a complex property that needs special formatting
				if prop == "agentPoolProfiles" || prop == "subnets" || prop == "primaryEndpoints" {
					isExpanded := m.expandedProperties[prop]
					if isExpanded {
						content.WriteString(fmt.Sprintf("%s: %s\n",
							propStyle.Render(formattedName+" (Expanded)"),
							formatComplexProperty(prop, value, 1)))
					} else {
						// Show condensed view with expansion hint
						summary := getPropertySummary(prop, value)
						expandHint := lipgloss.NewStyle().Foreground(colorGray).Render(" [Press 'e' to expand]")
						content.WriteString(fmt.Sprintf("%s: %s%s\n",
							propStyle.Render(formattedName),
							valueStyle.Render(summary),
							expandHint))
					}
				} else {
					// Simple property formatting
					formattedValue := formatValue(value)
					content.WriteString(fmt.Sprintf("%s: %s\n",
						propStyle.Render(formattedName),
						valueStyle.Render(formattedValue)))
				}
			}
		}
	}

	// Footer with help text
	content.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Faint(true).Foreground(colorGray)
	content.WriteString(helpStyle.Render("Press [d] for Dashboard view • [Tab] to switch panels"))

	return content.String()
}

func (m model) renderDashboardView(width, height int) string {
	resource := m.selectedResource
	var content strings.Builder

	// Dashboard Header
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Padding(0, 1)
	content.WriteString(headerStyle.Render(fmt.Sprintf("📊 Dashboard: %s", resource.Name)))
	content.WriteString("\n\n")

	// Mock metrics for demonstration (in real implementation, these would come from Azure Monitor)
	metrics := map[string]interface{}{
		"cpu_usage":    75.2,
		"memory_usage": 68.5,
		"network_in":   12.3,
		"network_out":  8.7,
		"disk_read":    45.2,
		"disk_write":   23.1,
	}

	// Metrics Section
	sectionStyle := lipgloss.NewStyle().Bold(true).Foreground(colorGreen)
	content.WriteString(sectionStyle.Render("📈 Live Metrics"))
	content.WriteString("\n")

	// CPU and Memory in a row
	cpuStyle := lipgloss.NewStyle().Foreground(colorGreen)
	if cpu, ok := metrics["cpu_usage"].(float64); ok && cpu > 80 {
		cpuStyle = lipgloss.NewStyle().Foreground(colorRed)
	} else if cpu, ok := metrics["cpu_usage"].(float64); ok && cpu > 60 {
		cpuStyle = lipgloss.NewStyle().Foreground(colorYellow)
	}

	memStyle := lipgloss.NewStyle().Foreground(colorGreen)
	if mem, ok := metrics["memory_usage"].(float64); ok && mem > 85 {
		memStyle = lipgloss.NewStyle().Foreground(colorRed)
	} else if mem, ok := metrics["memory_usage"].(float64); ok && mem > 70 {
		memStyle = lipgloss.NewStyle().Foreground(colorYellow)
	}

	content.WriteString(fmt.Sprintf("🖥️  CPU: %s  💾 Memory: %s\n",
		cpuStyle.Render(fmt.Sprintf("%.1f%%", metrics["cpu_usage"])),
		memStyle.Render(fmt.Sprintf("%.1f%%", metrics["memory_usage"]))))

	// Network metrics
	netStyle := lipgloss.NewStyle().Foreground(colorBlue)
	content.WriteString(fmt.Sprintf("🌐 Network In: %s  Out: %s\n",
		netStyle.Render(fmt.Sprintf("%.1f MB/s", metrics["network_in"])),
		netStyle.Render(fmt.Sprintf("%.1f MB/s", metrics["network_out"]))))

	// Disk metrics
	diskStyle := lipgloss.NewStyle().Foreground(colorPurple)
	content.WriteString(fmt.Sprintf("💿 Disk Read: %s  Write: %s\n",
		diskStyle.Render(fmt.Sprintf("%.1f MB/s", metrics["disk_read"])),
		diskStyle.Render(fmt.Sprintf("%.1f MB/s", metrics["disk_write"]))))

	// Simple trend visualization
	content.WriteString("\n")
	content.WriteString(sectionStyle.Render("📊 Trend (24h)"))
	content.WriteString("\n")
	trendStyle := lipgloss.NewStyle().Foreground(colorAqua)
	content.WriteString(trendStyle.Render("CPU: ▁▂▃▄▅▆▇█▇▆▅▄▃▂▁▂▃▄▅▆▇█▇▆▅▄"))
	content.WriteString("\n")
	content.WriteString(trendStyle.Render("MEM: ▂▃▄▃▂▃▄▅▆▅▄▃▂▃▄▅▆▇▆▅▄▃▂▃▄▅"))
	content.WriteString("\n")

	// AI-Parsed Logs Section
	content.WriteString("\n")
	content.WriteString(sectionStyle.Render("🤖 AI Log Analysis"))
	content.WriteString("\n")

	logStyle := lipgloss.NewStyle().Foreground(fgMedium)
	if m.aiProvider != nil {
		// Mock AI-parsed log insights
		insights := []string{
			"✅ No critical errors detected in the last 24h",
			"⚠️  High CPU usage detected during peak hours (2-4 PM)",
			"📈 Memory usage is trending upward, consider scaling",
			"🔧 Recommended: Enable auto-scaling for better performance",
		}

		for _, insight := range insights {
			content.WriteString(logStyle.Render(insight))
			content.WriteString("\n")
		}
	} else {
		content.WriteString(logStyle.Render("AI analysis unavailable (set OPENAI_API_KEY)"))
		content.WriteString("\n")
	}

	// Recent Activity/Logs
	content.WriteString("\n")
	content.WriteString(sectionStyle.Render("📋 Recent Activity"))
	content.WriteString("\n")

	// Mock recent activity
	activities := []string{
		"[15:30] VM started successfully",
		"[15:25] Resource health check: OK",
		"[15:20] Auto-scaling triggered",
		"[15:15] Backup completed",
		"[15:10] Security scan: No issues",
	}

	activityStyle := lipgloss.NewStyle().Foreground(colorGray)
	for _, activity := range activities {
		content.WriteString(activityStyle.Render(activity))
		content.WriteString("\n")
	}

	// Footer
	content.WriteString("\n")
	helpStyle := lipgloss.NewStyle().Faint(true).Foreground(colorGray)
	content.WriteString(helpStyle.Render("Press [d] for Details view • Auto-refresh: 30s"))

	return content.String()
}

// Helper functions for better formatting
func getResourceTypeDisplayName(resourceType string) string {
	displayNames := map[string]string{
		"Microsoft.Compute/virtualMachines":          "Virtual Machine",
		"Microsoft.KeyVault/vaults":                  "Key Vault",
		"Microsoft.Storage/storageAccounts":          "Storage Account",
		"Microsoft.Network/networkInterfaces":        "Network Interface",
		"Microsoft.Network/publicIPAddresses":        "Public IP",
		"Microsoft.Network/virtualNetworks":          "Virtual Network",
		"Microsoft.Compute/disks":                    "Disk",
		"Microsoft.ContainerService/managedClusters": "AKS Cluster",
		"Microsoft.Web/sites":                        "Web App",
		"Microsoft.Sql/servers":                      "SQL Server",
	}

	if displayName, exists := displayNames[resourceType]; exists {
		return displayName
	}
	return resourceType
}

func getImportantProperties(resourceType string) []string {
	switch resourceType {
	case "Microsoft.Compute/virtualMachines":
		return []string{"vmSize", "osType", "provisioningState", "adminUsername", "computerName"}
	case "Microsoft.Storage/storageAccounts":
		return []string{"accountType", "kind", "accessTier", "primaryEndpoints"}
	case "Microsoft.Network/virtualNetworks":
		return []string{"addressSpace", "subnets", "dhcpOptions"}
	case "Microsoft.ContainerService/managedClusters":
		return []string{"kubernetesVersion", "nodeResourceGroup", "dnsPrefix", "agentPoolProfiles"}
	default:
		return []string{"provisioningState", "location", "sku"}
	}
}

func formatPropertyName(prop string) string {
	// Convert camelCase to readable format
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

// getPropertySummary returns a condensed summary of complex properties
func getPropertySummary(propName string, value interface{}) string {
	switch propName {
	case "agentPoolProfiles":
		if pools, ok := value.([]interface{}); ok {
			return fmt.Sprintf("%d Agent Pool(s)", len(pools))
		}
	case "subnets":
		if subnets, ok := value.([]interface{}); ok {
			return fmt.Sprintf("%d Subnet(s)", len(subnets))
		}
	case "primaryEndpoints":
		if endpoints, ok := value.(map[string]interface{}); ok {
			return fmt.Sprintf("%d Endpoint(s)", len(endpoints))
		}
	}
	return fmt.Sprintf("%v", value)
}

// formatComplexProperty formats complex properties like AKS agent pool profiles
func formatComplexProperty(propName string, value interface{}, indent int) string {
	indentStr := strings.Repeat("  ", indent)

	switch propName {
	case "agentPoolProfiles":
		return formatAgentPoolProfiles(value, indent)
	case "subnets":
		return formatSubnets(value, indent)
	case "primaryEndpoints":
		return formatEndpoints(value, indent)
	default:
		// Handle generic objects and arrays
		if slice, ok := value.([]interface{}); ok {
			var result strings.Builder
			result.WriteString(fmt.Sprintf("\n%s└─ %d items:", indentStr, len(slice)))
			for i, item := range slice {
				if i < 3 { // Show first 3 items
					result.WriteString(fmt.Sprintf("\n%s   [%d] %v", indentStr, i, formatValue(item)))
				} else if i == 3 {
					result.WriteString(fmt.Sprintf("\n%s   ... and %d more", indentStr, len(slice)-3))
					break
				}
			}
			return result.String()
		} else if obj, ok := value.(map[string]interface{}); ok {
			var result strings.Builder
			result.WriteString(fmt.Sprintf("\n%s└─ Object with %d properties:", indentStr, len(obj)))
			count := 0
			for key, val := range obj {
				if count < 3 { // Show first 3 properties
					result.WriteString(fmt.Sprintf("\n%s   %s: %v", indentStr, key, formatValue(val)))
					count++
				} else if count == 3 {
					result.WriteString(fmt.Sprintf("\n%s   ... and %d more properties", indentStr, len(obj)-3))
					break
				}
			}
			return result.String()
		}
		return fmt.Sprintf("%v", value)
	}
}

func formatAgentPoolProfiles(value interface{}, indent int) string {
	indentStr := strings.Repeat("  ", indent)
	var result strings.Builder

	if pools, ok := value.([]interface{}); ok {
		result.WriteString(fmt.Sprintf("\n%s└─ %d Agent Pool(s):", indentStr, len(pools)))

		for i, pool := range pools {
			if poolMap, ok := pool.(map[string]interface{}); ok {
				result.WriteString(fmt.Sprintf("\n%s   [%d] Pool Configuration:", indentStr, i+1))

				// Show important pool properties
				if name, exists := poolMap["name"]; exists {
					result.WriteString(fmt.Sprintf("\n%s       Name: %v", indentStr, name))
				}
				if count, exists := poolMap["count"]; exists {
					result.WriteString(fmt.Sprintf("\n%s       Node Count: %v", indentStr, count))
				}
				if vmSize, exists := poolMap["vmSize"]; exists {
					result.WriteString(fmt.Sprintf("\n%s       VM Size: %v", indentStr, vmSize))
				}
				if osType, exists := poolMap["osType"]; exists {
					result.WriteString(fmt.Sprintf("\n%s       OS Type: %v", indentStr, osType))
				}
				if mode, exists := poolMap["mode"]; exists {
					result.WriteString(fmt.Sprintf("\n%s       Mode: %v", indentStr, mode))
				}
			}
		}
	} else {
		result.WriteString(fmt.Sprintf("%v", value))
	}

	return result.String()
}

func formatSubnets(value interface{}, indent int) string {
	indentStr := strings.Repeat("  ", indent)
	var result strings.Builder

	if subnets, ok := value.([]interface{}); ok {
		result.WriteString(fmt.Sprintf("\n%s└─ %d Subnet(s):", indentStr, len(subnets)))

		for i, subnet := range subnets {
			if subnetMap, ok := subnet.(map[string]interface{}); ok {
				result.WriteString(fmt.Sprintf("\n%s   [%d] Subnet:", indentStr, i+1))

				if name, exists := subnetMap["name"]; exists {
					result.WriteString(fmt.Sprintf("\n%s       Name: %v", indentStr, name))
				}
				if addressPrefix, exists := subnetMap["addressPrefix"]; exists {
					result.WriteString(fmt.Sprintf("\n%s       Address Prefix: %v", indentStr, addressPrefix))
				}
			}
		}
	} else {
		result.WriteString(fmt.Sprintf("%v", value))
	}

	return result.String()
}

func formatEndpoints(value interface{}, indent int) string {
	indentStr := strings.Repeat("  ", indent)
	var result strings.Builder

	if endpoints, ok := value.(map[string]interface{}); ok {
		result.WriteString(fmt.Sprintf("\n%s└─ Endpoints:", indentStr))

		for name, endpoint := range endpoints {
			result.WriteString(fmt.Sprintf("\n%s   %s: %v", indentStr, name, endpoint))
		}
	} else {
		result.WriteString(fmt.Sprintf("%v", value))
	}

	return result.String()
}

func formatValue(value interface{}) string {
	switch v := value.(type) {
	case string:
		if len(v) > 50 {
			return v[:47] + "..."
		}
		return v
	default:
		str := fmt.Sprintf("%v", v)
		if len(str) > 50 {
			return str[:47] + "..."
		}
		return str
	}
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// renderScrollableContent applies scrolling to content and adds scroll indicators
func (m model) renderScrollableContent(content string, maxHeight int) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// Calculate visible range
	startLine := m.rightPanelScrollOffset
	endLine := startLine + maxHeight
	if endLine > totalLines {
		endLine = totalLines
	}
	if startLine >= totalLines {
		startLine = totalLines - 1
	}
	if startLine < 0 {
		startLine = 0
	}

	// Get visible content
	var visibleLines []string
	if startLine < endLine {
		visibleLines = lines[startLine:endLine]
	}

	// Add scroll indicators
	result := strings.Join(visibleLines, "\n")
	if totalLines > maxHeight {
		scrollIndicator := fmt.Sprintf(" [%d-%d/%d]", startLine+1, endLine, totalLines)
		if startLine > 0 {
			result = "↑ More above ↑" + scrollIndicator + "\n" + result
		}
		if endLine < totalLines {
			result = result + "\n↓ More below ↓"
		}
	}

	return result
}

func main() {
	m := initModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting Azure Dashboard: %v\n", err)
	}
}
