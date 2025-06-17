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

	"github.com/olafkfreund/azure-tui/internal/tui"
)

// Data structures
type Subscription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TenantID  string `json:"tenantId"`
	IsDefault bool   `json:"isDefault"`
}

type ResourceGroup struct {
	Name     string `json:"name"`
	Location string `json:"location"`
}

type AzureResource struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Type     string `json:"type"`
	Location string `json:"location"`
}

// Message types
type subscriptionsLoadedMsg struct {
	subscriptions []Subscription
}

type subscriptionsErrorMsg struct {
	error string
}

type resourceGroupsLoadedMsg struct {
	groups []ResourceGroup
}

type resourceGroupsErrorMsg struct {
	error string
}

type resourcesInGroupMsg struct {
	groupName string
	resources []AzureResource
}

type resourcesInGroupErrMsg struct {
	groupName string
	error     string
}

type resourceDetailsMsg struct {
	resourceName string
	details      string
}

// Model
type model struct {
	treeView         *tui.TreeView
	tabManager       *tui.TabManager
	statusBar        *tui.StatusBar
	width            int
	height           int
	ready            bool
	subscriptions    []Subscription
	resourceGroups   []ResourceGroup
	resourcesInGroup []AzureResource
	loadingState     string // "subscriptions", "groups", "ready", "error"
}

// Azure data fetching functions
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
			ID:        s.ID,
			Name:      s.Name,
			TenantID:  s.TenantID,
			IsDefault: s.IsDefault,
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
		groups = append(groups, ResourceGroup{
			Name:     g.Name,
			Location: g.Location,
		})
	}

	return groups, nil
}

func fetchResourcesInGroup(groupName string) ([]AzureResource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "resource", "list",
		"--resource-group", groupName,
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch resources: %v", err)
	}

	var azResources []struct {
		ID       string `json:"id"`
		Name     string `json:"name"`
		Type     string `json:"type"`
		Location string `json:"location"`
	}

	if err := json.Unmarshal(output, &azResources); err != nil {
		return nil, fmt.Errorf("failed to parse resource data: %v", err)
	}

	var resources []AzureResource
	for _, r := range azResources {
		resources = append(resources, AzureResource{
			ID:       r.ID,
			Name:     r.Name,
			Type:     r.Type,
			Location: r.Location,
		})
	}

	return resources, nil
}

// Helper functions
func extractResourceGroupFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

// Commands
func loadSubscriptionsCmd() tea.Cmd {
	return func() tea.Msg {
		subs, err := fetchSubscriptions()
		if err != nil {
			return subscriptionsErrorMsg{error: err.Error()}
		}
		return subscriptionsLoadedMsg{subscriptions: subs}
	}
}

func loadResourceGroupsCmd() tea.Cmd {
	return func() tea.Msg {
		groups, err := fetchResourceGroups()
		if err != nil {
			return resourceGroupsErrorMsg{error: err.Error()}
		}
		return resourceGroupsLoadedMsg{groups: groups}
	}
}

func loadResourcesInGroupCmd(groupName string) tea.Cmd {
	return func() tea.Msg {
		resources, err := fetchResourcesInGroup(groupName)
		if err != nil {
			return resourcesInGroupErrMsg{groupName, err.Error()}
		}
		return resourcesInGroupMsg{groupName, resources}
	}
}

func (m *model) showResourceDetails(node *tui.TreeNode) tea.Cmd {
	return func() tea.Msg {
		if node.ResourceData == nil {
			return resourceDetailsMsg{
				resourceName: node.Name,
				details:      "No resource data available",
			}
		}

		// Extract resource information
		if resource, ok := node.ResourceData.(AzureResource); ok {
			resourceGroup := extractResourceGroupFromID(resource.ID)

			details := fmt.Sprintf(`Resource Details:

Name: %s
Type: %s  
Location: %s
Resource Group: %s
Resource ID: %s

Press 'Tab' to switch tabs, 'q' to quit`,
				resource.Name,
				resource.Type,
				resource.Location,
				resourceGroup,
				resource.ID)

			return resourceDetailsMsg{
				resourceName: resource.Name,
				details:      details,
			}
		}

		return resourceDetailsMsg{
			resourceName: node.Name,
			details:      "Unable to parse resource data",
		}
	}
}

// Initialize model
func initModel() model {
	treeView := tui.NewTreeView()
	tabManager := tui.NewTabManager()
	statusBar := tui.CreatePowerlineStatusBar(80)

	// Add a default tab
	tabManager.AddTab(tui.Tab{
		Title:    "Azure Resources",
		Content:  "Welcome to Azure TUI\n\nLoading Azure data...",
		Type:     "main",
		Closable: false,
	})

	return model{
		treeView:     treeView,
		tabManager:   tabManager,
		statusBar:    statusBar,
		loadingState: "subscriptions",
	}
}

// BubbleTea methods
func (m model) Init() tea.Cmd {
	return tea.Batch(
		loadSubscriptionsCmd(),
		loadResourceGroupsCmd(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		if m.statusBar != nil {
			m.statusBar.Width = msg.Width
		}
		return m, nil

	case subscriptionsLoadedMsg:
		m.subscriptions = msg.subscriptions
		m.loadingState = "groups"
		if m.tabManager != nil && len(m.tabManager.Tabs) > 0 {
			m.tabManager.Tabs[0].Content = fmt.Sprintf("Azure TUI\n\n✅ Loaded %d subscriptions\n🔄 Loading resource groups...", len(msg.subscriptions))
		}
		return m, nil

	case subscriptionsErrorMsg:
		m.loadingState = "error"
		if m.tabManager != nil && len(m.tabManager.Tabs) > 0 {
			m.tabManager.Tabs[0].Content = fmt.Sprintf("Azure TUI\n\n❌ Failed to load subscriptions: %s", msg.error)
		}
		return m, nil

	case resourceGroupsLoadedMsg:
		m.resourceGroups = msg.groups
		m.loadingState = "ready"

		// Populate tree view with resource groups
		if m.treeView != nil {
			for _, group := range msg.groups {
				groupNode := m.treeView.AddResourceGroup(group.Name, group.Location)
				m.treeView.AddResource(groupNode, "Loading...", "placeholder", nil)
			}
			// Ensure first item is selected
			m.treeView.EnsureSelection()
		}

		if m.tabManager != nil && len(m.tabManager.Tabs) > 0 {
			m.tabManager.Tabs[0].Content = fmt.Sprintf(`Azure TUI

✅ Loaded %d subscriptions
✅ Loaded %d resource groups

Navigation:
• j/k or ↓/↑ - Navigate
• Space - Expand/collapse resource group
• Enter - View resource details
• Tab - Switch tabs
• r - Refresh
• q - Quit`, len(m.subscriptions), len(msg.groups))
		}
		return m, nil

	case resourceGroupsErrorMsg:
		m.loadingState = "error"
		if m.tabManager != nil && len(m.tabManager.Tabs) > 0 {
			m.tabManager.Tabs[0].Content = fmt.Sprintf("Azure TUI\n\n❌ Failed to load resource groups: %s", msg.error)
		}
		return m, nil

	case resourcesInGroupMsg:
		m.resourcesInGroup = msg.resources
		// Update tree view with actual resources
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
		return m, nil

	case resourcesInGroupErrMsg:
		// Handle resource loading error - could show an error in the tab
		return m, nil

	case resourceDetailsMsg:
		// Create a new tab with resource details
		if m.tabManager != nil {
			resourceTab := tui.Tab{
				Title:    fmt.Sprintf("📦 %s", msg.resourceName),
				Content:  msg.details,
				Type:     "resource",
				Closable: true,
			}
			m.tabManager.AddTab(resourceTab)
		}
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.treeView != nil {
				m.treeView.SelectNext()
				m.treeView.EnsureSelection()
			}
			return m, nil
		case "k", "up":
			if m.treeView != nil {
				m.treeView.SelectPrevious()
				m.treeView.EnsureSelection()
			}
			return m, nil
		case " ":
			if m.treeView != nil {
				selectedNode, expanded := m.treeView.ToggleExpansion()
				if expanded && selectedNode != nil && selectedNode.Type == "group" {
					return m, loadResourcesInGroupCmd(selectedNode.Name)
				}
			}
			return m, nil
		case "enter":
			if m.treeView != nil {
				selectedNode := m.treeView.GetSelectedNode()
				if selectedNode != nil {
					switch selectedNode.Type {
					case "group":
						// Expand/collapse resource group
						selectedNode.Expanded = !selectedNode.Expanded
						if selectedNode.Expanded {
							return m, loadResourcesInGroupCmd(selectedNode.Name)
						}
					case "resource":
						// Show resource details in a new tab
						return m, m.showResourceDetails(selectedNode)
					}
				}
			}
			return m, nil
		case "r":
			return m, tea.Batch(loadSubscriptionsCmd(), loadResourceGroupsCmd())
		case "tab":
			if m.tabManager != nil {
				m.tabManager.SwitchTab(1)
			}
			return m, nil
		case "shift+tab":
			if m.tabManager != nil {
				m.tabManager.SwitchTab(-1)
			}
			return m, nil
		}
	}

	return m, nil
}

func (m model) View() string {
	if !m.ready {
		return "Loading Azure TUI..."
	}

	// Update status bar
	if m.statusBar != nil {
		m.statusBar.Segments = []tui.PowerlineSegment{}
		m.statusBar.AddSegment("☁️ Azure TUI", lipgloss.Color("39"), lipgloss.Color("15"))

		switch m.loadingState {
		case "subscriptions":
			m.statusBar.AddSegment("Loading Subscriptions", lipgloss.Color("11"), lipgloss.Color("0"))
		case "groups":
			m.statusBar.AddSegment("Loading Resource Groups", lipgloss.Color("11"), lipgloss.Color("0"))
		case "ready":
			m.statusBar.AddSegment(fmt.Sprintf("%d Groups", len(m.resourceGroups)), lipgloss.Color("10"), lipgloss.Color("0"))
		case "error":
			m.statusBar.AddSegment("Error", lipgloss.Color("9"), lipgloss.Color("15"))
		}
	}

	// Render tree view
	treeContent := ""
	if m.treeView != nil {
		treeContent = m.treeView.RenderTreeView(m.width/3, m.height-3)
	}

	// Render tabs content
	tabsContent := ""
	if m.tabManager != nil {
		tabsContent = tui.RenderTabs(m.tabManager, "Azure TUI")
	}

	// Join panels horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, treeContent, tabsContent)

	// Add status bar
	statusBarContent := ""
	if m.statusBar != nil {
		statusBarContent = m.statusBar.RenderStatusBar()
	}

	// Join with status bar vertically
	fullView := lipgloss.JoinVertical(lipgloss.Left, statusBarContent, mainContent)

	return fullView
}

func main() {
	m := initModel()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting Azure TUI: %v\n", err)
	}
}
