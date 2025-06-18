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

	"github.com/olafkfreund/azure-tui/internal/azure/aci"
	"github.com/olafkfreund/azure-tui/internal/azure/network"
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

// Network dashboard message types
type networkDashboardMsg struct{ content string }
type vnetDetailsMsg struct{ content string }
type nsgDetailsMsg struct{ content string }
type networkTopologyMsg struct{ content string }
type networkAIAnalysisMsg struct{ content string }
type networkResourceCreatedMsg struct {
	resourceType string
	result       resourceactions.ActionResult
}

// Container Instance message types
type containerInstanceDetailsMsg struct{ content string }
type containerInstanceLogsMsg struct{ content string }
type containerInstanceActionMsg struct {
	action string
	result resourceactions.ActionResult
}
type containerInstanceScaleMsg struct {
	cpu    float64
	memory float64
	result resourceactions.ActionResult
}

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
	leftPanelScrollOffset  int // Add independent scrolling for left panel
	rightPanelMaxLines     int
	actionInProgress       bool
	lastActionResult       *resourceactions.ActionResult
	showDashboard          bool
	logEntries             []string
	// New navigation fields
	activeView            string          // "details", "dashboard", "welcome", "network-dashboard", "vnet-details", "nsg-details", "network-topology", "network-ai"
	propertyExpandedIndex int             // For navigating expanded properties
	expandedProperties    map[string]bool // Track which properties are expanded

	// Network-specific fields
	networkDashboardContent string
	vnetDetailsContent      string
	nsgDetailsContent       string
	networkTopologyContent  string
	networkAIContent        string

	// Container Instance-specific content
	containerInstanceDetailsContent string
	containerInstanceLogsContent    string

	// Help popup state
	showHelpPopup bool

	// Navigation stack for back navigation
	navigationStack []string
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
			} else if resource.Type == "Microsoft.ContainerService/managedClusters" {
				result = resourceactions.StartAKSCluster(resource.Name, resource.ResourceGroup)
			} else if resource.Type == "Microsoft.ContainerInstance/containerGroups" {
				err := aci.StartContainerInstance(resource.Name, resource.ResourceGroup)
				if err != nil {
					result = resourceactions.ActionResult{Success: false, Message: fmt.Sprintf("Failed to start container instance: %v", err)}
				} else {
					result = resourceactions.ActionResult{Success: true, Message: fmt.Sprintf("Successfully started container instance %s", resource.Name)}
				}
			}
		case "stop":
			if resource.Type == "Microsoft.Compute/virtualMachines" {
				result = resourceactions.StopVM(resource.Name, resource.ResourceGroup)
			} else if resource.Type == "Microsoft.ContainerService/managedClusters" {
				result = resourceactions.StopAKSCluster(resource.Name, resource.ResourceGroup)
			} else if resource.Type == "Microsoft.ContainerInstance/containerGroups" {
				err := aci.StopContainerInstance(resource.Name, resource.ResourceGroup)
				if err != nil {
					result = resourceactions.ActionResult{Success: false, Message: fmt.Sprintf("Failed to stop container instance: %v", err)}
				} else {
					result = resourceactions.ActionResult{Success: true, Message: fmt.Sprintf("Successfully stopped container instance %s", resource.Name)}
				}
			}
		case "restart":
			if resource.Type == "Microsoft.Compute/virtualMachines" {
				result = resourceactions.RestartVM(resource.Name, resource.ResourceGroup)
			} else if resource.Type == "Microsoft.ContainerInstance/containerGroups" {
				err := aci.RestartContainerInstance(resource.Name, resource.ResourceGroup)
				if err != nil {
					result = resourceactions.ActionResult{Success: false, Message: fmt.Sprintf("Failed to restart container instance: %v", err)}
				} else {
					result = resourceactions.ActionResult{Success: true, Message: fmt.Sprintf("Successfully restarted container instance %s", resource.Name)}
				}
			}
		case "ssh":
			if resource.Type == "Microsoft.Compute/virtualMachines" {
				result = resourceactions.ExecuteVMSSH(resource.Name, resource.ResourceGroup, "azureuser")
			}
		case "bastion":
			if resource.Type == "Microsoft.Compute/virtualMachines" {
				result = resourceactions.ConnectVMBastion(resource.Name, resource.ResourceGroup)
			}
		case "pods":
			if resource.Type == "Microsoft.ContainerService/managedClusters" {
				result = resourceactions.ListAKSPods(resource.Name, resource.ResourceGroup)
			}
		case "deployments":
			if resource.Type == "Microsoft.ContainerService/managedClusters" {
				result = resourceactions.ListAKSDeployments(resource.Name, resource.ResourceGroup)
			}
		case "nodes":
			if resource.Type == "Microsoft.ContainerService/managedClusters" {
				result = resourceactions.GetAKSNodes(resource.Name, resource.ResourceGroup)
			}
		case "services":
			if resource.Type == "Microsoft.ContainerService/managedClusters" {
				result = resourceactions.ListAKSServices(resource.Name, resource.ResourceGroup)
			}
		default:
			result = resourceactions.ActionResult{Success: false, Message: "Unsupported action"}
		}
		return resourceActionMsg{action: action, resource: resource, result: result}
	}
}

// =============================================================================
// NETWORK DASHBOARD AND MANAGEMENT COMMANDS
// =============================================================================

// showNetworkDashboardCmd displays comprehensive network dashboard
func showNetworkDashboardCmd() tea.Cmd {
	return func() tea.Msg {
		// Use the network package's RenderNetworkDashboard function
		dashboardContent := network.RenderNetworkDashboard()
		return networkDashboardMsg{content: dashboardContent}
	}
}

// showVNetDetailsCmd displays detailed VNet information
func showVNetDetailsCmd(vnetName, resourceGroup string) tea.Cmd {
	return func() tea.Msg {
		// Use the network package's RenderVNetDetails function
		vnetContent := network.RenderVNetDetails(vnetName, resourceGroup)
		return vnetDetailsMsg{content: vnetContent}
	}
}

// showNSGDetailsCmd displays detailed NSG information
func showNSGDetailsCmd(nsgName, resourceGroup string) tea.Cmd {
	return func() tea.Msg {
		// Use the network package's RenderNSGDetails function
		nsgContent := network.RenderNSGDetails(nsgName, resourceGroup)
		return nsgDetailsMsg{content: nsgContent}
	}
}

// showNetworkTopologyCmd displays network topology view
func showNetworkTopologyCmd() tea.Cmd {
	return func() tea.Msg {
		// Use the network package's RenderNetworkTopology function
		topologyContent := network.RenderNetworkTopology()
		return networkTopologyMsg{content: topologyContent}
	}
}

// showNetworkAIAnalysisCmd provides AI-powered network analysis
func showNetworkAIAnalysisCmd() tea.Cmd {
	return func() tea.Msg {
		// Use the network package's RenderNetworkAIAnalysis function
		aiContent := network.RenderNetworkAIAnalysis()
		return networkAIAnalysisMsg{content: aiContent}
	}
}

// createNetworkResourceCmd creates network resources
func createNetworkResourceCmd(resourceType string) tea.Cmd {
	return func() tea.Msg {
		var result resourceactions.ActionResult

		switch resourceType {
		case "vnet":
			result = resourceactions.ActionResult{
				Success: true,
				Message: "VNet creation wizard would open here. Use Azure CLI: az network vnet create",
				Output:  "Ready to create Virtual Network",
			}
		case "nsg":
			result = resourceactions.ActionResult{
				Success: true,
				Message: "NSG creation wizard would open here. Use Azure CLI: az network nsg create",
				Output:  "Ready to create Network Security Group",
			}
		case "subnet":
			result = resourceactions.ActionResult{
				Success: true,
				Message: "Subnet creation wizard would open here. Use Azure CLI: az network vnet subnet create",
				Output:  "Ready to create Subnet",
			}
		case "publicip":
			result = resourceactions.ActionResult{
				Success: true,
				Message: "Public IP creation wizard would open here. Use Azure CLI: az network public-ip create",
				Output:  "Ready to create Public IP",
			}
		case "loadbalancer":
			result = resourceactions.ActionResult{
				Success: true,
				Message: "Load Balancer creation wizard would open here. Use Azure CLI: az network lb create",
				Output:  "Ready to create Load Balancer",
			}
		default:
			result = resourceactions.ActionResult{
				Success: false,
				Message: fmt.Sprintf("Unknown network resource type: %s", resourceType),
				Output:  "",
			}
		}

		return networkResourceCreatedMsg{resourceType: resourceType, result: result}
	}
}

// =============================================================================
// CONTAINER INSTANCE MANAGEMENT COMMANDS
// =============================================================================

// showContainerInstanceDetailsCmd displays detailed container instance information
func showContainerInstanceDetailsCmd(name, resourceGroup string) tea.Cmd {
	return func() tea.Msg {
		content := aci.RenderContainerInstanceDetails(name, resourceGroup)
		return containerInstanceDetailsMsg{content: content}
	}
}

// getContainerLogsCmd retrieves container logs
func getContainerLogsCmd(name, resourceGroup, containerName string, tail int) tea.Cmd {
	return func() tea.Msg {
		logs, err := aci.GetContainerLogs(name, resourceGroup, containerName, tail)
		if err != nil {
			return containerInstanceLogsMsg{content: fmt.Sprintf("Error getting logs: %v", err)}
		}

		// Format logs with header
		content := fmt.Sprintf("🐳 Container Logs: %s\n", name)
		content += "═══════════════════════════════════════════════════════════════\n\n"
		content += logs

		return containerInstanceLogsMsg{content: content}
	}
}

// execIntoContainerCmd executes a command in the container
func execIntoContainerCmd(name, resourceGroup, containerName, command string) tea.Cmd {
	return func() tea.Msg {
		err := aci.ExecIntoContainer(name, resourceGroup, containerName, command)
		var result resourceactions.ActionResult

		if err != nil {
			result = resourceactions.ActionResult{
				Success: false,
				Message: fmt.Sprintf("Failed to exec into container: %v", err),
				Output:  "",
			}
		} else {
			result = resourceactions.ActionResult{
				Success: true,
				Message: fmt.Sprintf("Successfully executed command in container %s", name),
				Output:  "Command executed successfully",
			}
		}

		return containerInstanceActionMsg{action: "exec", result: result}
	}
}

// attachToContainerCmd attaches to a running container
func attachToContainerCmd(name, resourceGroup, containerName string) tea.Cmd {
	return func() tea.Msg {
		err := aci.AttachToContainer(name, resourceGroup, containerName)
		var result resourceactions.ActionResult

		if err != nil {
			result = resourceactions.ActionResult{
				Success: false,
				Message: fmt.Sprintf("Failed to attach to container: %v", err),
				Output:  "",
			}
		} else {
			result = resourceactions.ActionResult{
				Success: true,
				Message: fmt.Sprintf("Successfully attached to container %s", name),
				Output:  "Attached to container",
			}
		}

		return containerInstanceActionMsg{action: "attach", result: result}
	}
}

// scaleContainerInstanceCmd scales container instance resources
func scaleContainerInstanceCmd(name, resourceGroup string, cpu, memory float64) tea.Cmd {
	return func() tea.Msg {
		err := aci.UpdateContainerInstance(name, resourceGroup, cpu, memory)
		var result resourceactions.ActionResult

		if err != nil {
			result = resourceactions.ActionResult{
				Success: false,
				Message: fmt.Sprintf("Failed to scale container instance: %v", err),
				Output:  "",
			}
		} else {
			result = resourceactions.ActionResult{
				Success: true,
				Message: fmt.Sprintf("Successfully scaled container instance %s to %.1f CPU, %.1f GB RAM", name, cpu, memory),
				Output:  fmt.Sprintf("New resources: %.1f CPU cores, %.1f GB memory", cpu, memory),
			}
		}

		return containerInstanceScaleMsg{cpu: cpu, memory: memory, result: result}
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
		leftPanelScrollOffset:  0, // Initialize left panel scroll offset
		rightPanelMaxLines:     50,
		showDashboard:          false,
		logEntries:             []string{},
		activeView:             "welcome",
		propertyExpandedIndex:  -1,
		expandedProperties:     make(map[string]bool),
		showHelpPopup:          false,
		navigationStack:        []string{}, // Initialize navigation stack
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

	case networkDashboardMsg:
		m.actionInProgress = false
		m.networkDashboardContent = msg.content
		m.pushView("network-dashboard")
		// Add debug logging
		m.logEntries = append(m.logEntries, "DEBUG: Network Dashboard message received, content length: "+fmt.Sprintf("%d", len(msg.content)))

	case vnetDetailsMsg:
		m.actionInProgress = false
		m.vnetDetailsContent = msg.content
		m.pushView("vnet-details")

	case nsgDetailsMsg:
		m.actionInProgress = false
		m.nsgDetailsContent = msg.content
		m.pushView("nsg-details")

	case networkTopologyMsg:
		m.actionInProgress = false
		m.networkTopologyContent = msg.content
		m.pushView("network-topology")

	case networkAIAnalysisMsg:
		m.actionInProgress = false
		m.networkAIContent = msg.content
		m.pushView("network-ai")

	case networkResourceCreatedMsg:
		m.actionInProgress = false
		m.logEntries = append(m.logEntries, fmt.Sprintf("Created %s: %s", msg.resourceType, msg.result.Message))

	// Container Instance message handlers
	case containerInstanceDetailsMsg:
		m.actionInProgress = false
		m.containerInstanceDetailsContent = msg.content
		m.pushView("container-details")

	case containerInstanceLogsMsg:
		m.actionInProgress = false
		m.containerInstanceLogsContent = msg.content
		m.pushView("container-logs")

	case containerInstanceActionMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		if msg.result.Success && m.selectedResource != nil {
			return m, loadResourceDetailsCmd(*m.selectedResource)
		}

	case containerInstanceScaleMsg:
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
				// Don't reset scroll when switching to maintain position
			}
		case "right", "l":
			// Right navigation - switch to details panel
			if m.selectedPanel == 0 {
				m.selectedPanel = 1
				// Don't reset scroll when switching to maintain position
			}
		case "d":
			// Toggle dashboard view
			m.showDashboard = !m.showDashboard
			if m.showDashboard {
				m.pushView("dashboard")
			} else {
				if m.selectedResource != nil {
					m.pushView("details")
				} else {
					m.pushView("welcome")
				}
			}
		case "j", "down":
			if m.selectedPanel == 0 && m.treeView != nil {
				// Try to navigate first
				m.treeView.SelectNext()
				m.treeView.EnsureSelection()
				if selectedNode := m.treeView.GetSelectedNode(); selectedNode != nil && selectedNode.Type == "resource" {
					if resource, ok := selectedNode.ResourceData.(AzureResource); ok {
						return m, loadResourceDetailsCmd(resource)
					}
				}
			} else if m.selectedPanel == 1 {
				// Right panel scrolling down
				rightContent := m.renderResourcePanel(m.width/3, m.height-2)
				totalLines := strings.Count(rightContent, "\n")
				maxLines := max(0, totalLines-(m.height-6))
				if m.rightPanelScrollOffset < maxLines {
					m.rightPanelScrollOffset++
				}
			}
		case "k", "up":
			if m.selectedPanel == 0 && m.treeView != nil {
				// Navigate normally
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
		case "ctrl+j", "ctrl+down":
			// Dedicated scrolling down for current panel
			if m.selectedPanel == 0 && m.treeView != nil {
				// Left panel scrolling
				treeContent := m.treeView.RenderTreeView(m.width/3-4, m.height-2)
				totalLines := strings.Count(treeContent, "\n")
				maxLines := max(0, totalLines-(m.height-6))
				if m.leftPanelScrollOffset < maxLines {
					m.leftPanelScrollOffset++
				}
			} else if m.selectedPanel == 1 {
				// Right panel scrolling
				rightContent := m.renderResourcePanel(m.width/3, m.height-2)
				totalLines := strings.Count(rightContent, "\n")
				maxLines := max(0, totalLines-(m.height-6))
				if m.rightPanelScrollOffset < maxLines {
					m.rightPanelScrollOffset++
				}
			}
		case "ctrl+k", "ctrl+up":
			// Dedicated scrolling up for current panel
			if m.selectedPanel == 0 {
				// Left panel scrolling up
				if m.leftPanelScrollOffset > 0 {
					m.leftPanelScrollOffset--
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
		case "c":
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.Compute/virtualMachines" {
				m.actionInProgress = true
				return m, executeResourceActionCmd("ssh", *m.selectedResource)
			}
		case "b":
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.Compute/virtualMachines" {
				m.actionInProgress = true
				return m, executeResourceActionCmd("bastion", *m.selectedResource)
			}
		case "p":
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.ContainerService/managedClusters" {
				m.actionInProgress = true
				return m, executeResourceActionCmd("pods", *m.selectedResource)
			}
		case "D":
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.ContainerService/managedClusters" {
				m.actionInProgress = true
				return m, executeResourceActionCmd("deployments", *m.selectedResource)
			}
		case "n":
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.ContainerService/managedClusters" {
				m.actionInProgress = true
				return m, executeResourceActionCmd("nodes", *m.selectedResource)
			}
		case "v":
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.ContainerService/managedClusters" {
				m.actionInProgress = true
				return m, executeResourceActionCmd("services", *m.selectedResource)
			}
		case "N":
			// Show comprehensive network dashboard
			if !m.actionInProgress {
				m.actionInProgress = true
				// Add debug logging
				m.logEntries = append(m.logEntries, "DEBUG: Network Dashboard command triggered")
				return m, showNetworkDashboardCmd()
			}
		case "V":
			// Show VNet details for selected network resource
			if m.selectedResource != nil && !m.actionInProgress && strings.Contains(m.selectedResource.Type, "Network") {
				if strings.Contains(m.selectedResource.Type, "virtualNetworks") {
					m.actionInProgress = true
					return m, showVNetDetailsCmd(m.selectedResource.Name, m.selectedResource.ResourceGroup)
				}
			}
		case "G":
			// Show NSG details for selected network security group
			if m.selectedResource != nil && !m.actionInProgress && strings.Contains(m.selectedResource.Type, "networkSecurityGroups") {
				m.actionInProgress = true
				return m, showNSGDetailsCmd(m.selectedResource.Name, m.selectedResource.ResourceGroup)
			}
		case "Z":
			// Show network topology view
			if !m.actionInProgress {
				m.actionInProgress = true
				return m, showNetworkTopologyCmd()
			}
		case "A":
			// Show AI-powered network analysis
			if !m.actionInProgress {
				m.actionInProgress = true
				return m, showNetworkAIAnalysisCmd()
			}
		case "C":
			// Create VNet action for network resources
			if !m.actionInProgress {
				m.actionInProgress = true
				return m, createNetworkResourceCmd("vnet")
			}
		case "ctrl+n":
			// Create NSG action
			if !m.actionInProgress {
				m.actionInProgress = true
				return m, createNetworkResourceCmd("nsg")
			}
		case "ctrl+s":
			// Create subnet action
			if !m.actionInProgress {
				m.actionInProgress = true
				return m, createNetworkResourceCmd("subnet")
			}
		case "ctrl+p":
			// Create public IP action
			if !m.actionInProgress {
				m.actionInProgress = true
				return m, createNetworkResourceCmd("publicip")
			}
		case "ctrl+l":
			// Create load balancer action
			if !m.actionInProgress {
				m.actionInProgress = true
				return m, createNetworkResourceCmd("loadbalancer")
			}

		// Container Instance Management Actions
		case "L":
			// Get container logs
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.ContainerInstance/containerGroups" {
				m.actionInProgress = true
				return m, getContainerLogsCmd(m.selectedResource.Name, m.selectedResource.ResourceGroup, "", 100)
			}
		case "E":
			// Exec into container
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.ContainerInstance/containerGroups" {
				m.actionInProgress = true
				return m, execIntoContainerCmd(m.selectedResource.Name, m.selectedResource.ResourceGroup, "", "/bin/bash")
			}
		case "a":
			// Attach to container (only for container instances)
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.ContainerInstance/containerGroups" {
				m.actionInProgress = true
				return m, attachToContainerCmd(m.selectedResource.Name, m.selectedResource.ResourceGroup, "")
			}
		case "u":
			// Update/scale container instance
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.ContainerInstance/containerGroups" {
				m.actionInProgress = true
				// Scale up CPU and memory (this could be made interactive in future)
				return m, scaleContainerInstanceCmd(m.selectedResource.Name, m.selectedResource.ResourceGroup, 2.0, 4.0)
			}
		case "I":
			// Show detailed container instance information
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.ContainerInstance/containerGroups" {
				m.actionInProgress = true
				return m, showContainerInstanceDetailsCmd(m.selectedResource.Name, m.selectedResource.ResourceGroup)
			}
		case "R":
			return m, loadDataCmd()
		case "?":
			// Toggle help popup
			m.showHelpPopup = !m.showHelpPopup
		case "escape":
			// Close help popup if open, otherwise navigate back
			if m.showHelpPopup {
				m.showHelpPopup = false
			} else {
				// Try to go back to previous view
				if !m.popView() {
					// If no previous view, stay on current view
					// Could optionally set to welcome view here
				}
			}
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
			if m.leftPanelScrollOffset > 0 {
				panelHelp = " (j/k:scroll)"
			} else {
				panelHelp = " (j/k:navigate/scroll)"
			}
			navigationHelp = "l/→:Details"
		}
		m.statusBar.AddSegment(fmt.Sprintf("▶ %s%s", panelName, panelHelp), colorAqua, bgMedium)
		m.statusBar.AddSegment(navigationHelp, colorPurple, bgMedium)

		// Add expansion hint for AKS resources
		if m.selectedResource != nil && m.selectedResource.Type == "Microsoft.ContainerService/managedClusters" && m.selectedPanel == 1 {
			m.statusBar.AddSegment("e:Expand AKS Properties", colorYellow, bgMedium)
		}

		// Add navigation indicator if there's history
		if len(m.navigationStack) > 0 {
			m.statusBar.AddSegment(fmt.Sprintf("Esc:Back(%d)", len(m.navigationStack)), colorAqua, bgMedium)
		}

		m.statusBar.AddSegment("Tab:Switch d:Dashboard s:Start S:Stop r:Restart R:Refresh ?:Help q:Quit", colorGray, bgLight)
	}

	// Two-panel layout - Fixed width constraints to prevent layout breaking
	leftWidth := m.width / 3
	rightWidth := m.width - leftWidth

	// Ensure minimum widths to prevent layout collapse
	if leftWidth < 20 {
		leftWidth = 20
	}
	if rightWidth < 30 {
		rightWidth = 30
	}

	// Tree panel with strict width enforcement
	treeContent := ""
	if m.treeView != nil {
		treeContentRaw := m.treeView.RenderTreeView(leftWidth-4, m.height-2)
		// Apply scrolling to left panel if it has long content
		if m.selectedPanel == 0 {
			treeContent = m.renderScrollableContentWithOffset(treeContentRaw, m.height-6, m.leftPanelScrollOffset)
		} else {
			treeContent = treeContentRaw
		}
	}

	// Style left panel with STRICT width constraints
	leftPanelStyle := lipgloss.NewStyle().
		Width(leftWidth).
		MaxWidth(leftWidth). // Enforce maximum width
		Foreground(fgMedium).
		Padding(1, 2)

	// Add visual indicator for active panel
	if m.selectedPanel == 0 {
		leftPanelStyle = leftPanelStyle.
			Foreground(fgLight).
			Bold(true)
		// Add enhanced active panel indicator
		treeContent = "🔍 " + strings.Replace(treeContent, "\n", "\n   ", -1)
	}

	leftPanel := leftPanelStyle.Render(treeContent)

	// Details panel with scrolling support and STRICT width constraints
	rightContentRaw := m.renderResourcePanel(rightWidth-4, m.height-2)

	// Ensure content is properly wrapped to prevent layout breaking
	rightContentWrapped := ensureContentWidth(rightContentRaw, rightWidth-8)

	var rightContent string

	// Apply scrolling if right panel is active
	if m.selectedPanel == 1 {
		rightContent = m.renderScrollableContent(rightContentWrapped, m.height-6)
	} else {
		rightContent = rightContentWrapped
	}

	// Style right panel with STRICT width constraints
	rightPanelStyle := lipgloss.NewStyle().
		Width(rightWidth).
		MaxWidth(rightWidth). // Enforce maximum width
		Foreground(fgMedium).
		Padding(1, 2)

	if m.selectedPanel == 1 {
		rightPanelStyle = rightPanelStyle.
			Foreground(fgLight).
			Bold(true)
		// Add enhanced active panel marker
		rightContent = "📊 " + strings.Replace(rightContent, "\n", "\n   ", -1)
	}

	rightPanel := rightPanelStyle.Render(rightContent)

	// Join everything
	statusBarContent := ""
	if m.statusBar != nil {
		statusBarContent = m.statusBar.RenderStatusBar()
	}

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
	fullView := lipgloss.JoinVertical(lipgloss.Left, statusBarContent, mainContent)

	// Render help popup if active
	if m.showHelpPopup {
		// Create a comprehensive help content
		var helpContent strings.Builder
		helpContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("⌨️  Azure TUI - Keyboard Shortcuts"))
		helpContent.WriteString("\n\n")

		helpContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("🧭 Navigation:"))
		helpContent.WriteString("\n")
		helpContent.WriteString("j/k ↑/↓    Navigate up/down in tree\n")
		helpContent.WriteString("h/l ←/→    Switch between panels\n")
		helpContent.WriteString("Space      Expand/collapse resource groups\n")
		helpContent.WriteString("Enter      Open resource in details panel\n")
		helpContent.WriteString("Tab        Switch between panels\n")
		helpContent.WriteString("e          Expand/collapse complex properties\n")
		helpContent.WriteString("Ctrl+j/k   Scroll up/down in current panel\n\n")

		helpContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorAqua).Render("⚡ Resource Actions:"))
		helpContent.WriteString("\n")
		helpContent.WriteString("s          Start resource (VMs, Containers)\n")
		helpContent.WriteString("S          Stop resource (VMs, Containers)\n")
		helpContent.WriteString("r          Restart resource (VMs, Containers)\n")
		helpContent.WriteString("d          Toggle dashboard view\n")
		helpContent.WriteString("R          Refresh all data\n\n")

		helpContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("🌐 Network Management:"))
		helpContent.WriteString("\n")
		helpContent.WriteString("N          Network Dashboard\n")
		helpContent.WriteString("V          VNet Details (for VNets)\n")
		helpContent.WriteString("G          NSG Details (for NSGs)\n")
		helpContent.WriteString("Z          Network Topology\n")
		helpContent.WriteString("A          AI Network Analysis\n")
		helpContent.WriteString("C          Create VNet\n")
		helpContent.WriteString("Ctrl+N     Create NSG\n\n")

		helpContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorPurple).Render("🐳 Container Management:"))
		helpContent.WriteString("\n")
		helpContent.WriteString("L          Get Container Logs\n")
		helpContent.WriteString("E          Exec into Container\n")
		helpContent.WriteString("a          Attach to Container\n")
		helpContent.WriteString("u          Scale Container Resources\n")
		helpContent.WriteString("I          Container Instance Details\n\n")

		helpContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorYellow).Render("🔐 SSH & AKS:"))
		helpContent.WriteString("\n")
		helpContent.WriteString("c          SSH Connect (VMs)\n")
		helpContent.WriteString("b          Bastion Connect (VMs)\n")
		helpContent.WriteString("p          List Pods (AKS)\n")
		helpContent.WriteString("D          List Deployments (AKS)\n")
		helpContent.WriteString("n          List Nodes (AKS)\n")
		helpContent.WriteString("v          List Services (AKS)\n\n")

		helpContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGray).Render("🎮 Interface:"))
		helpContent.WriteString("\n")
		helpContent.WriteString("?          Show/hide this help\n")
		helpContent.WriteString("Esc        Navigate back / Close dialogs\n")
		helpContent.WriteString("q          Quit application\n\n")

		helpContent.WriteString(lipgloss.NewStyle().Italic(true).Foreground(colorGray).Render("Press '?' or 'Esc' to close this help"))

		// Create popup style
		popupStyle := lipgloss.NewStyle().
			Background(bgMedium).
			Foreground(fgLight).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBlue).
			Padding(1, 2).
			Width(70).
			Align(lipgloss.Center, lipgloss.Top)

		styledPopup := popupStyle.Render(helpContent.String())

		// Create a simple centered layout
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, styledPopup)
	}

	return lipgloss.NewStyle().Background(bgDark).Render(fullView)
}

func (m model) renderResourcePanel(width, height int) string {
	// Handle network-specific views first
	switch m.activeView {
	case "network-dashboard":
		return m.networkDashboardContent
	case "vnet-details":
		return m.vnetDetailsContent
	case "nsg-details":
		return m.nsgDetailsContent
	case "network-topology":
		return m.networkTopologyContent
	case "network-ai":
		return m.networkAIContent
	case "container-details":
		return m.containerInstanceDetailsContent
	case "container-logs":
		return m.containerInstanceLogsContent
	}

	// Handle regular resource views
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
	content.WriteString("4. Use Tab to switch between panels\n")
	content.WriteString("5. Press '?' for complete keyboard shortcuts\n\n")

	content.WriteString(sectionStyle.Render("✨ Key Features:"))
	content.WriteString("\n")
	featureStyle := lipgloss.NewStyle().Foreground(colorPurple)
	content.WriteString(fmt.Sprintf("%s Enhanced resource management with comprehensive actions\n", featureStyle.Render("•")))
	content.WriteString(fmt.Sprintf("%s Network topology visualization and analysis\n", featureStyle.Render("•")))
	content.WriteString(fmt.Sprintf("%s Container instance lifecycle management\n", featureStyle.Render("•")))
	content.WriteString(fmt.Sprintf("%s SSH and Bastion connectivity for VMs\n", featureStyle.Render("•")))
	content.WriteString(fmt.Sprintf("%s AI-powered resource insights and analysis\n", featureStyle.Render("•")))
	content.WriteString(fmt.Sprintf("%s Terraform/Bicep code generation\n\n", featureStyle.Render("•")))

	aiStatus := "❌ Disabled (set OPENAI_API_KEY)"
	if m.aiProvider != nil {
		aiStatus = "✅ Enabled"
	}
	statusStyle := lipgloss.NewStyle().Foreground(colorGray)
	content.WriteString(fmt.Sprintf("🤖 AI Features: %s\n\n", statusStyle.Render(aiStatus)))

	helpStyle := lipgloss.NewStyle().Foreground(colorYellow).Bold(true)
	content.WriteString(fmt.Sprintf("💡 Press %s for complete keyboard shortcuts and help\n\n", helpStyle.Render("?")))

	content.WriteString("Select a resource from the left panel to see detailed information and available actions.")

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
		// Wrap AI description to fit panel width
		wrappedAIText := wrapText(m.aiDescription, width-10)
		content.WriteString(aiStyle.Render(wrappedAIText))
		content.WriteString("\n")
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
		content.WriteString(fmt.Sprintf("%s SSH Connect\n", actionStyle.Render("[c]")))
		content.WriteString(fmt.Sprintf("%s Bastion Connect\n", actionStyle.Render("[b]")))

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

	// Actions Section for AKS Clusters
	if resource.Type == "Microsoft.ContainerService/managedClusters" {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🚢 AKS Management Actions"))
		content.WriteString("\n")

		actionStyle := lipgloss.NewStyle().Foreground(colorBlue)
		content.WriteString(fmt.Sprintf("%s Start Cluster\n", actionStyle.Render("[s]")))
		content.WriteString(fmt.Sprintf("%s Stop Cluster\n", actionStyle.Render("[S]")))
		content.WriteString(fmt.Sprintf("%s List Pods\n", actionStyle.Render("[p]")))
		content.WriteString(fmt.Sprintf("%s List Deployments\n", actionStyle.Render("[D]")))
		content.WriteString(fmt.Sprintf("%s List Nodes\n", actionStyle.Render("[n]")))
		content.WriteString(fmt.Sprintf("%s List Services\n", actionStyle.Render("[v]")))

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

	// Actions Section for Container Instances
	if resource.Type == "Microsoft.ContainerInstance/containerGroups" {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🐳 Container Instance Management"))
		content.WriteString("\n")

		actionStyle := lipgloss.NewStyle().Foreground(colorBlue)
		content.WriteString(fmt.Sprintf("%s Start Container Instance\n", actionStyle.Render("[s]")))
		content.WriteString(fmt.Sprintf("%s Stop Container Instance\n", actionStyle.Render("[S]")))
		content.WriteString(fmt.Sprintf("%s Restart Container Instance\n", actionStyle.Render("[r]")))
		content.WriteString(fmt.Sprintf("%s Get Container Logs\n", actionStyle.Render("[L]")))
		content.WriteString(fmt.Sprintf("%s Exec into Container\n", actionStyle.Render("[E]")))
		content.WriteString(fmt.Sprintf("%s Attach to Container\n", actionStyle.Render("[a]")))
		content.WriteString(fmt.Sprintf("%s Scale Container Resources\n", actionStyle.Render("[u]")))
		content.WriteString(fmt.Sprintf("%s Show Detailed Info\n", actionStyle.Render("[I]")))

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

		// Use clean list formatting for better property display
		listData := tui.FormatPropertiesAsSimpleList(m.resourceDetails.Properties)
		content.WriteString(listData)

		// Add expansion hints for complex properties
		content.WriteString("\n")
		helpStyle := lipgloss.NewStyle().Faint(true).Foreground(colorGray)
		content.WriteString(helpStyle.Render("💡 Tip: Press 'e' to expand complex properties like Agent Pools"))
		content.WriteString("\n")
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

// wrapText wraps text to fit within a specified width
func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	lines := strings.Split(text, "\n")

	for _, line := range lines {
		if len(line) <= width {
			result.WriteString(line)
			result.WriteString("\n")
			continue
		}

		words := strings.Fields(line)
		currentLine := ""

		for _, word := range words {
			if len(currentLine+" "+word) > width {
				if currentLine != "" {
					result.WriteString(strings.TrimSpace(currentLine))
					result.WriteString("\n")
					currentLine = word
				} else {
					// Word is longer than width, force break
					result.WriteString(word)
					result.WriteString("\n")
					currentLine = ""
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
			result.WriteString(strings.TrimSpace(currentLine))
			result.WriteString("\n")
		}
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// ensureContentWidth ensures all content fits within the specified width by wrapping text
func ensureContentWidth(content string, maxWidth int) string {
	if maxWidth <= 0 {
		return content
	}

	lines := strings.Split(content, "\n")
	var result strings.Builder

	for _, line := range lines {
		// Check if line is too long
		if len(line) > maxWidth {
			// If it contains ANSI escape codes, preserve them
			wrappedLine := wrapText(line, maxWidth)
			result.WriteString(wrappedLine)
		} else {
			result.WriteString(line)
		}
		result.WriteString("\n")
	}

	return strings.TrimSuffix(result.String(), "\n")
}

// Navigation stack helper functions

// pushView adds the current view to the navigation stack before switching to a new view
func (m *model) pushView(newView string) {
	// Only push if we're actually changing views
	if m.activeView != newView {
		m.navigationStack = append(m.navigationStack, m.activeView)
		m.activeView = newView
	}
}

// popView goes back to the previous view from the navigation stack
func (m *model) popView() bool {
	if len(m.navigationStack) == 0 {
		return false // No previous view to go back to
	}

	// Get the last view from the stack
	lastIndex := len(m.navigationStack) - 1
	previousView := m.navigationStack[lastIndex]

	// Remove it from the stack
	m.navigationStack = m.navigationStack[:lastIndex]

	// Switch to the previous view
	m.activeView = previousView

	// Reset scroll offsets when going back
	m.rightPanelScrollOffset = 0
	m.leftPanelScrollOffset = 0

	return true
}

// clearNavigationStack clears the navigation history (useful for returning to main menu)
func (m *model) clearNavigationStack() {
	m.navigationStack = []string{}
}

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
	return m.renderScrollableContentWithOffset(content, maxHeight, m.rightPanelScrollOffset)
}

// renderScrollableContentWithOffset applies scrolling to content with a custom offset
func (m model) renderScrollableContentWithOffset(content string, maxHeight int, scrollOffset int) string {
	lines := strings.Split(content, "\n")
	totalLines := len(lines)

	// Calculate visible range
	startLine := scrollOffset
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

// createShortcutsMap creates a comprehensive keyboard shortcuts map for the help popup
func createShortcutsMap() map[string]string {
	return map[string]string{
		// Navigation
		"j/k ↑/↓": "Navigate up/down in tree",
		"h/l ←/→": "Switch between panels",
		"Space":   "Expand/collapse resource groups",
		"Enter":   "Open resource in details panel",
		"Tab":     "Switch between panels",
		"e":       "Expand/collapse complex properties",

		// Resource Actions
		"s": "Start resource (VMs, Containers)",
		"S": "Stop resource (VMs, Containers)",
		"r": "Restart resource (VMs, Containers)",
		"d": "Toggle dashboard view",
		"R": "Refresh all data",

		// Network Management
		"N":      "Network Dashboard",
		"V":      "VNet Details (for VNets)",
		"G":      "NSG Details (for NSGs)",
		"Z":      "Network Topology",
		"A":      "AI Network Analysis",
		"C":      "Create VNet",
		"Ctrl+N": "Create NSG",
		"Ctrl+S": "Create Subnet",
		"Ctrl+P": "Create Public IP",
		"Ctrl+L": "Create Load Balancer",

		// Container Instance Management
		"L": "Get Container Logs",
		"E": "Exec into Container",
		"a": "Attach to Container",
		"u": "Scale Container Resources",
		"I": "Container Instance Details",

		// SSH & AKS Management
		"c": "SSH Connect (VMs)",
		"b": "Bastion Connect (VMs)",
		"p": "List Pods (AKS)",
		"D": "List Deployments (AKS)",
		"n": "List Nodes (AKS)",
		"v": "List Services (AKS)",

		// Interface
		"?":   "Show/hide this help",
		"Esc": "Navigate back / Close dialogs",
		"q":   "Quit application",
	}
}

func main() {
	m := initModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting Azure Dashboard: %v\n", err)
	}
}
