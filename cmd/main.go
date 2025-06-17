package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/olafkfreund/azure-tui/internal/azure/resourceactions"
	"github.com/olafkfreund/azure-tui/internal/azure/resourcedetails"
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
type resourceActionMsg struct {
	action   string
	resource AzureResource
	result   resourceactions.ActionResult
}
type errorMsg struct{ error string }

type model struct {
	treeView         *tui.TreeView
	statusBar        *tui.StatusBar
	width, height    int
	ready            bool
	subscriptions    []Subscription
	resourceGroups   []ResourceGroup
	allResources     []AzureResource
	selectedResource *AzureResource
	resourceDetails  *resourcedetails.ResourceDetails
	loadingState     string
	selectedPanel    int
	actionInProgress bool
	lastActionResult *resourceactions.ActionResult
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
	return model{
		treeView:      tui.NewTreeView(),
		statusBar:     tui.CreatePowerlineStatusBar(80),
		loadingState:  "loading",
		selectedPanel: 0,
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
		case "j", "down":
			if m.selectedPanel == 0 && m.treeView != nil {
				m.treeView.SelectNext()
				m.treeView.EnsureSelection()
				if selectedNode := m.treeView.GetSelectedNode(); selectedNode != nil && selectedNode.Type == "resource" {
					if resource, ok := selectedNode.ResourceData.(AzureResource); ok {
						return m, loadResourceDetailsCmd(resource)
					}
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
		if m.selectedPanel == 1 {
			panelName = "Details"
		}
		m.statusBar.AddSegment(fmt.Sprintf("Panel: %s", panelName), colorAqua, bgMedium)
		m.statusBar.AddSegment("Tab:Switch s:Start S:Stop r:Restart R:Refresh q:Quit", colorGray, bgLight)
	}

	// Two-panel layout - NO BORDERS AT ALL
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth

	// Tree panel
	treeContent := ""
	if m.treeView != nil {
		treeContent = m.treeView.RenderTreeView(leftWidth-4, m.height-2)
	}

	leftBg := bgMedium
	if m.selectedPanel == 0 {
		leftBg = bgLight
	}

	leftPanel := lipgloss.NewStyle().
		Width(leftWidth).
		Background(leftBg).
		Foreground(fgMedium).
		Padding(1, 2).
		Render(treeContent)

	// Details panel
	rightContent := m.renderResourcePanel(rightWidth-4, m.height-2)

	rightBg := bgMedium
	if m.selectedPanel == 1 {
		rightBg = bgLight
	}

	rightPanel := lipgloss.NewStyle().
		Width(rightWidth).
		Background(rightBg).
		Foreground(fgMedium).
		Padding(1, 2).
		Render(rightContent)

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
	return m.renderResourceDetailsAndActions(width, height)
}

func (m model) renderWelcomePanel(width, height int) string {
	content := `📊 Azure Resource Dashboard

Welcome to Azure TUI Dashboard!

🎯 Getting Started:
1. Navigate through resource groups in the left panel
2. Press Space/Enter to expand a resource group  
3. Select a resource to view details and actions
4. Use Tab to switch between panels

🎮 Quick Actions:
• s - Start resource (VMs)
• S - Stop resource (VMs)
• r - Restart resource (VMs)
• R - Refresh all data
• Tab - Switch panels
• q - Quit

💡 Select a resource from the left panel to see detailed information and available actions here.`

	return content
}

func (m model) renderResourceDetailsAndActions(width, height int) string {
	resource := m.selectedResource

	content := fmt.Sprintf(`📦 Resource Details

🏷️  Name: %s
🏗️  Type: %s
📍 Location: %s
📁 Resource Group: %s
🆔 Resource ID: %s`,
		resource.Name, resource.Type, resource.Location, resource.ResourceGroup, resource.ID)

	if resource.Status != "" {
		statusColor := "🔴"
		if strings.Contains(strings.ToLower(resource.Status), "running") {
			statusColor = "🟢"
		} else if strings.Contains(strings.ToLower(resource.Status), "deallocated") {
			statusColor = "🟡"
		}
		content += fmt.Sprintf("\n⚡ Status: %s %s", statusColor, resource.Status)
	}

	if len(resource.Tags) > 0 {
		content += "\n\n🏷️  Tags:"
		for key, value := range resource.Tags {
			content += fmt.Sprintf("\n   • %s: %s", key, value)
		}
	}

	if resource.Type == "Microsoft.Compute/virtualMachines" {
		content += "\n\n🎮 Available Actions:"
		content += "\n   [s] Start VM"
		content += "\n   [S] Stop VM"
		content += "\n   [r] Restart VM"

		if m.actionInProgress {
			content += "\n\n⏳ Action in progress..."
		}

		if m.lastActionResult != nil {
			status := "❌"
			if m.lastActionResult.Success {
				status = "✅"
			}
			content += fmt.Sprintf("\n\n%s Last Action: %s", status, m.lastActionResult.Message)
		}
	}

	if m.resourceDetails != nil {
		content += "\n\n📋 Additional Properties:"
		content += fmt.Sprintf("\n   • ID: %s", m.resourceDetails.ID)
		content += fmt.Sprintf("\n   • Type: %s", m.resourceDetails.Type)
		content += fmt.Sprintf("\n   • Location: %s", m.resourceDetails.Location)
		content += fmt.Sprintf("\n   • Resource Group: %s", m.resourceDetails.ResourceGroup)

		if m.resourceDetails.Status != "" {
			content += fmt.Sprintf("\n   • Status: %s", m.resourceDetails.Status)
		}

		if len(m.resourceDetails.Tags) > 0 {
			content += "\n\n🏷️  Additional Tags:"
			for key, value := range m.resourceDetails.Tags {
				content += fmt.Sprintf("\n   • %s: %s", key, value)
			}
		}

		if len(m.resourceDetails.Properties) > 0 {
			content += "\n\n⚙️  Additional Properties:"
			for key, value := range m.resourceDetails.Properties {
				content += fmt.Sprintf("\n   • %s: %v", key, value)
			}
		}
	}

	return content
}

func main() {
	m := initModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting Azure Dashboard: %v\n", err)
	}
}
