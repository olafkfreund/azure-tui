package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/olafkfreund/azure-tui/internal/azure/azuresdk"
	"github.com/olafkfreund/azure-tui/internal/azure/tfbicep"
	"github.com/olafkfreund/azure-tui/internal/openai"
	"github.com/olafkfreund/azure-tui/internal/tui"
)

// Azure SDK client for resource group listing
var azureClient *azuresdk.AzureClient

func init() {
	var err error
	azureClient, err = azuresdk.NewAzureClient()
	if err != nil {
		panic("Failed to initialize Azure SDK client: " + err.Error())
	}
}

var titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("63"))
var subtitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("33"))
var helpStyle = lipgloss.NewStyle().Faint(true)

// Subscription, Tenant, and Resource Group info

type Subscription struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	TenantID  string `json:"tenantId"`
	IsDefault bool   `json:"isDefault"`
}

type Tenant struct {
	ID   string `json:"id"`
	Name string `json:"displayName"`
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

type AKSCluster struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	ResourceGroup string `json:"resourceGroup"`
}

type resourcesInGroupMsg struct {
	groupName string
	resources []AzureResource
}
type resourcesInGroupErrMsg struct {
	groupName string
	err       string
}

type aksClustersMsg []AKSCluster
type aksClusterErrMsg string

type keyVaultsMsg []struct {
	Name          string
	Location      string
	ResourceGroup string
}
type keyVaultErrMsg string

// Storage Account messages

type storageAccountsMsg []struct {
	Name          string
	Location      string
	ResourceGroup string
}
type storageErrMsg string

// IaC file scan messages

type iacFilesMsg []struct{ Path, Type string }
type iacFilesErrMsg string

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "--create" {
			// Example: aztui --create vm --name demo01 --folder ./iac
			// TODO: Parse args, call AI provider to generate code, write file, print result
			os.Exit(0)
		}
		if os.Args[1] == "--deploy" {
			// Example: aztui --deploy ./iac/main.tf
			// TODO: Run deployment, print output, call AI on error
			os.Exit(0)
		}
		if os.Args[1] == "vnet-summary" {
			subID := "<your-subscription-id>" // TODO: get from config or flag
			resourceGroup := ""               // Optionally set
			_ = listAndSummarizeVNetsCLI(subID, resourceGroup)
			return
		}
	}
	p := tea.NewProgram(initialModel(), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if err := p.Start(); err != nil {
		os.Exit(1)
	}
}

// initialModel returns the starting state for the TUI.
func initialModel() tea.Model {
	// Initialize AI provider with API key from environment
	apiKey := os.Getenv("OPENAI_API_KEY")
	var aiProvider *openai.AIProvider
	if apiKey != "" {
		aiProvider = openai.NewAIProvider(apiKey)
	}

	return &model{
		profiles:              []string{"default"},
		currentProfile:        0,
		environments:          []string{"East US", "West Europe"},
		currentEnv:            0,
		loading:               true,
		tabManager:            tui.NewTabManager(),
		showShortcutsPopup:    false,
		activeTabIdx:          0,
		resourceTabs:          []tui.Tab{},
		aiProvider:            aiProvider,
		currentResourceConfig: make(map[string]string),
		resourceMetrics:       make(map[string]interface{}),
		termWidth:             80, // Default fallback dimensions
		termHeight:            24, // Default fallback dimensions
		maxVisibleGroups:      10, // Default scrolling limits
		maxVisibleResources:   15, // Default scrolling limits
		// Initialize with empty data - will be loaded from Azure
		subscriptions: []Subscription{},
		tenants:       []Tenant{},
		currentSub:    0,
		currentTenant: 0,
		// Initialize tree view and status bar
		treeView:         tui.NewTreeView(),
		statusBar:        tui.CreatePowerlineStatusBar(80),
		leftPanelWidth:   40,
		showTreeView:     true,
		contentTabs:      []tui.Tab{},
		activeContentTab: 0,
	}
}

type model struct {
	profiles       []string
	currentProfile int
	environments   []string
	currentEnv     int

	subscriptions []Subscription
	tenants       []Tenant
	currentSub    int
	currentTenant int
	loading       bool
	loadErr       string

	resourceGroups      []ResourceGroup
	resourceLoadErr     string
	resourceGroupIdx    int
	resourcesInGroup    []AzureResource
	resourcesInGroupErr string
	selectedGroup       string

	resourceIdx      int
	selectedResource string

	aksClusters []AKSCluster
	aksErr      string
	aksLoading  bool

	// Add AKS prompt state
	promptingAKS   bool
	promptStep     int
	promptName     string
	promptRG       string
	promptLocation string
	promptMsg      string

	// Add Key Vaults state
	keyVaults []struct {
		Name          string
		Location      string
		ResourceGroup string
	}
	keyVaultErr        string
	keyVaultsLoading   bool
	promptingKeyVault  bool
	promptKeyVaultStep int
	promptKeyVaultName string
	promptKeyVaultRG   string
	promptKeyVaultLoc  string
	promptKeyVaultMsg  string

	// Add Storage Accounts state
	storageAccounts []struct {
		Name          string
		Location      string
		ResourceGroup string
	}
	storageErr        string
	storageLoading    bool
	promptingStorage  bool
	promptStorageStep int
	promptStorageName string
	promptStorageRG   string
	promptStorageLoc  string
	promptStorageMsg  string

	// Add advanced TUI features
	usageMatrix  [][]string
	usageHeaders []string
	alarms       []struct {
		Name    string
		Status  string
		Details string
	}
	showMatrixPopup bool
	showAlarmsPopup bool
	matrixViewport  viewport.Model
	alarmsViewport  viewport.Model

	// Add IaC file scan state
	iacFiles       []struct{ Path, Type string }
	iacScanErr     string
	iacScanLoading bool
	iacDir         string // last scanned dir
	selectedIacIdx int
	showIacPanel   bool

	// Add IaC file viewing popup state
	showIacFilePopup    bool
	iacFilePopupContent string

	// Resource creation and deployment state
	creatingResource     bool
	createStep           int
	createResourceType   string
	createResourceName   string
	createResourceFolder string
	createResourceCode   string
	createResourceVars   map[string]string
	createAIMessage      string
	showCreatePopup      bool
	deployingResource    bool
	deployOutput         string
	showDeployPopup      bool
	aiConfirmPending     bool
	aiSuggestMessage     string

	// Caching for responsiveness
	resourceGroupsCache map[string][]ResourceGroup // subID -> groups
	resourcesCache      map[string][]AzureResource // groupName -> resources
	cacheTimestamp      time.Time
	isLoading           bool

	tabManager         *tui.TabManager // Multi-tab/window manager
	showShortcutsPopup bool            // Show keyboard shortcuts popup

	// Tab management
	activeTabIdx int       // index of the currently active tab (0 = main browser)
	resourceTabs []tui.Tab // tabs for opened resources (excluding main browser)

	// AI-powered features
	aiProvider  *openai.AIProvider
	showAIPopup bool
	aiMessage   string
	aiLoading   bool

	// Resource analysis and editing
	showResourceActions   bool
	showEditDialog        bool
	showDeleteDialog      bool
	showMetricsDialog     bool
	editingResourceName   string
	editingResourceType   string
	currentResourceConfig map[string]string
	resourceMetrics       map[string]interface{}

	// Terminal dimensions
	termWidth  int
	termHeight int

	// Tree view and interface components
	treeView       *tui.TreeView
	statusBar      *tui.StatusBar
	leftPanelWidth int
	showTreeView   bool

	// Content tabs (right side)
	contentTabs      []tui.Tab
	activeContentTab int

	// Scrolling and navigation
	groupScrollOffset    int
	resourceScrollOffset int
	maxVisibleGroups     int
	maxVisibleResources  int
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			subs, tenants, err := fetchAzureSubsAndTenants()
			if err != nil {
				return loadErrMsg(err.Error())
			}
			return loadedMsg{subs, tenants}
		},
		// Load real Azure resources immediately
		func() tea.Msg {
			// Try to load real Azure resource groups first
			groups, err := fetchResourceGroups("")
			if err != nil {
				// Only fallback to demo if Azure is completely unavailable
				return loadErrMsg("Failed to load Azure resources: " + err.Error())
			}
			return resourcesMsg(groups)
		},
	)
}

type loadedMsg struct {
	subs    []Subscription
	tenants []Tenant
}
type loadErrMsg string
type resourcesMsg []ResourceGroup
type resourceLoadErrMsg string

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		// Recalculate scroll limits based on new terminal size
		m.maxVisibleGroups = (m.termHeight - 12) / 2 // Account for borders and padding
		m.maxVisibleResources = (m.termHeight - 12) / 2
		// Ensure minimum values
		if m.maxVisibleGroups < 5 {
			m.maxVisibleGroups = 5
		}
		if m.maxVisibleResources < 5 {
			m.maxVisibleResources = 5
		}
		// Update status bar width
		if m.statusBar != nil {
			m.statusBar.Width = msg.Width
		}
		// Update tree view size
		if m.treeView != nil {
			m.treeView.MaxVisible = (msg.Height - 8) // Account for borders and status bar
		}
		return m, nil
	case loadedMsg:
		m.subscriptions = msg.subs
		m.tenants = msg.tenants
		m.loading = false
		m.currentSub = 0
		m.currentTenant = 0
		// Load resources for the selected subscription
		if len(m.subscriptions) > 0 {
			return m, loadResourcesCmd(m.subscriptions[m.currentSub].ID)
		}
		// Fallback: load demo resources directly
		return m, loadResourcesCmd("demo-sub")
	case resourcesMsg:
		m.resourceGroups = msg
		m.isLoading = false
		// Cache resource groups for the current subscription
		if m.resourceGroupsCache == nil {
			m.resourceGroupsCache = make(map[string][]ResourceGroup)
		}
		if len(m.subscriptions) > 0 {
			subID := m.subscriptions[m.currentSub].ID
			m.resourceGroupsCache[subID] = msg
		}

		// Populate tree view with resource groups
		m.treeView = tui.NewTreeView()
		for _, rg := range m.resourceGroups {
			m.treeView.AddResourceGroup(rg.Name, rg.Location)
		}

		// Set initial selection in tree view
		if len(m.resourceGroups) > 0 && len(m.treeView.Root.Children) > 0 {
			m.treeView.Root.Children[0].Selected = true
		}

		// Auto-select first resource group and load its resources for better UX
		if len(m.resourceGroups) > 0 && m.selectedGroup == "" {
			m.selectedGroup = m.resourceGroups[0].Name
			m.resourceGroupIdx = 0
			return m, loadResourcesInGroupCmd(m.selectedGroup)
		}
		return m, nil
	case resourceLoadErrMsg:
		m.resourceLoadErr = string(msg)
		m.isLoading = false
		return m, nil
	case resourcesInGroupMsg:
		if msg.groupName == m.selectedGroup {
			m.resourcesInGroup = msg.resources
			m.resourcesInGroupErr = ""

			// Add resources to the corresponding resource group in tree view
			for _, child := range m.treeView.Root.Children {
				if child.Name == msg.groupName {
					// Clear existing resources in this group
					child.Children = []*tui.TreeNode{}
					// Add all resources to this group
					for _, resource := range msg.resources {
						m.treeView.AddResource(child, resource.Name, resource.Type, resource)
					}
					break
				}
			}
		}
		return m, nil
	case resourcesInGroupErrMsg:
		if msg.groupName == m.selectedGroup {
			m.resourcesInGroup = nil
			m.resourcesInGroupErr = msg.err
		}
		return m, nil
	case aksClustersMsg:
		m.aksClusters = msg
		m.aksErr = ""
		m.aksLoading = false
		return m, nil
	case aksClusterErrMsg:
		m.aksErr = string(msg)
		m.aksLoading = false
		return m, nil
	case keyVaultsMsg:
		m.keyVaults = msg
		m.keyVaultErr = ""
		m.keyVaultsLoading = false
		return m, nil
	case keyVaultErrMsg:
		m.keyVaultErr = string(msg)
		m.keyVaultsLoading = false
		return m, nil
	case storageAccountsMsg:
		m.storageAccounts = msg
		m.storageErr = ""
		m.storageLoading = false
		return m, nil
	case storageErrMsg:
		m.storageErr = string(msg)
		m.storageLoading = false
		return m, nil
	case iacFilesMsg:
		m.iacFiles = msg
		m.iacScanErr = ""
		m.iacScanLoading = false
		return m, nil
	case iacFilesErrMsg:
		m.iacFiles = nil
		m.iacScanErr = string(msg)
		m.iacScanLoading = false
		return m, nil
	case loadErrMsg:
		m.loadErr = string(msg)
		m.loading = false
		return m, nil
	case tea.MouseMsg:
		// Handle mouse events for better interactivity
		if msg.Type == tea.MouseWheelUp {
			// Scroll up in resource groups panel
			if m.groupScrollOffset > 0 {
				m.groupScrollOffset--
			}
		} else if msg.Type == tea.MouseWheelDown {
			// Scroll down in resource groups panel
			if len(m.resourceGroups) > m.maxVisibleGroups && m.groupScrollOffset < len(m.resourceGroups)-m.maxVisibleGroups {
				m.groupScrollOffset++
			}
		} else if msg.Type == tea.MouseLeft {
			// Handle mouse clicks for selection (simplified)
			// Could be enhanced to detect which panel was clicked
		}
		return m, nil
	case tea.KeyMsg:
		if m.loading || m.isLoading {
			return m, nil
		}
		// Tab navigation
		switch msg.String() {
		case "tab":
			if len(m.resourceTabs) > 0 {
				m.activeTabIdx = (m.activeTabIdx + 1) % (len(m.resourceTabs) + 1)
			}
			return m, nil
		case "shift+tab":
			if len(m.resourceTabs) > 0 {
				m.activeTabIdx = (m.activeTabIdx - 1 + len(m.resourceTabs) + 1) % (len(m.resourceTabs) + 1)
			}
			return m, nil
		case "ctrl+w":
			// Prevent closing main tab
			if m.activeTabIdx > 0 && m.activeTabIdx <= len(m.resourceTabs) {
				m.resourceTabs = append(m.resourceTabs[:m.activeTabIdx-1], m.resourceTabs[m.activeTabIdx:]...)
				if m.activeTabIdx > len(m.resourceTabs) {
					m.activeTabIdx = len(m.resourceTabs)
				}
			}
			return m, nil
		case "enter":
			if m.showTreeView && len(m.resourcesInGroup) > 0 && m.resourceIdx < len(m.resourcesInGroup) {
				// Tree view mode: open resource in right-side content tab
				res := m.resourcesInGroup[m.resourceIdx]

				// Check if content tab for this resource already exists
				found := -1
				for i, tab := range m.contentTabs {
					if tab.Meta["id"] == res.ID {
						found = i
						break
					}
				}

				if found == -1 {
					// Create new content tab
					content := fmt.Sprintf("Resource Details\n%s\n\nName: %s\nType: %s\nLocation: %s\nResource Group: %s\nID: %s\n\nActions:\n• Press 'a' for AI analysis\n• Press 'M' for metrics\n• Press 'E' to edit\n• Press 'T' for Terraform\n• Press 'B' for Bicep",
						strings.Repeat("=", 50), res.Name, res.Type, res.Location, m.selectedGroup, res.ID)

					tab := tui.Tab{
						Title:    res.Name,
						Content:  content,
						Type:     res.Type,
						Meta:     map[string]string{"id": res.ID, "type": res.Type, "resourceGroup": m.selectedGroup},
						Closable: true,
					}
					m.contentTabs = append(m.contentTabs, tab)
					m.activeContentTab = len(m.contentTabs) - 1 // switch to new tab
				} else {
					m.activeContentTab = found // switch to existing tab
				}
				return m, nil
			} else if m.activeTabIdx == 0 && len(m.resourcesInGroup) > 0 && m.resourceIdx < len(m.resourcesInGroup) {
				// Legacy mode: open resource in traditional tab
				res := m.resourcesInGroup[m.resourceIdx]
				// Check if tab for this resource already exists
				found := -1
				for i, tab := range m.resourceTabs {
					if tab.Meta["id"] == res.ID {
						found = i
						break
					}
				}
				if found == -1 {
					// Open new tab
					tab := tui.Tab{
						Title:    res.Name,
						Content:  fmt.Sprintf("Name: %s\nType: %s\nLocation: %s\nID: %s", res.Name, res.Type, res.Location, res.ID),
						Type:     res.Type,
						Meta:     map[string]string{"id": res.ID, "type": res.Type},
						Closable: true,
					}
					m.resourceTabs = append(m.resourceTabs, tab)
					m.activeTabIdx = len(m.resourceTabs) // switch to new tab
				} else {
					m.activeTabIdx = found + 1 // switch to existing tab
				}
				return m, nil
			}
		}
		// ...existing navigation for resource groups/resources...
		if m.promptingAKS {
			// Interactive AKS prompt flow
			input := msg.String()
			switch m.promptStep {
			case 0:
				if input != "enter" && input != "" {
					m.promptName = input
					m.promptStep = 1
					m.promptMsg = "Enter resource group for AKS cluster:"
				}
			case 1:
				if input != "enter" && input != "" {
					m.promptRG = input
					m.promptStep = 2
					m.promptMsg = "Enter location (e.g. westeurope):"
				}
			case 2:
				if input != "enter" && input != "" {
					m.promptLocation = input
					m.promptingAKS = false
					go createAKSCluster(m.promptName, m.promptRG, m.promptLocation)
					m.aksLoading = true
					return m, loadAKSClustersCmd()
				}
			}
			return m, nil
		}
		if m.promptingKeyVault {
			// Interactive Key Vault prompt flow
			input := msg.String()
			switch m.promptKeyVaultStep {
			case 0:
				if input != "enter" && input != "" {
					m.promptKeyVaultName = input
					m.promptKeyVaultStep = 1
					m.promptKeyVaultMsg = "Enter resource group for Key Vault:"
				}
			case 1:
				if input != "enter" && input != "" {
					m.promptKeyVaultRG = input
					m.promptKeyVaultStep = 2
					m.promptKeyVaultMsg = "Enter location (e.g. westeurope):"
				}
			case 2:
				if input != "enter" && input != "" {
					m.promptKeyVaultLoc = input
					m.promptingKeyVault = false
					go createKeyVault(m.promptKeyVaultName, m.promptKeyVaultRG, m.promptKeyVaultLoc)
					m.keyVaultsLoading = true
					return m, loadKeyVaultsCmd()
				}
			}
			return m, nil
		}
		if m.promptingStorage {
			// Interactive Storage Account prompt flow
			input := msg.String()
			switch m.promptStorageStep {
			case 0:
				if input != "enter" && input != "" {
					m.promptStorageName = input
					m.promptStorageStep = 1
					m.promptStorageMsg = "Enter resource group for Storage Account:"
				}
			case 1:
				if input != "enter" && input != "" {
					m.promptStorageRG = input
					m.promptStorageStep = 2
					m.promptStorageMsg = "Enter location (e.g. westeurope):"
				}
			case 2:
				if input != "enter" && input != "" {
					m.promptStorageLoc = input
					m.promptingStorage = false
					go createStorageAccount(m.promptStorageName, m.promptStorageRG, m.promptStorageLoc)
					m.storageLoading = true
					return m, loadStorageAccountsCmd()
				}
			}
			return m, nil
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "down", "j":
			if len(m.resourceGroups) > 0 {
				m.resourceGroupIdx = (m.resourceGroupIdx + 1) % len(m.resourceGroups)
				// Auto-scroll if needed
				if m.resourceGroupIdx >= m.groupScrollOffset+m.maxVisibleGroups {
					m.groupScrollOffset = m.resourceGroupIdx - m.maxVisibleGroups + 1
				}
				m.selectedGroup = m.resourceGroups[m.resourceGroupIdx].Name

				// Update tree view selection if in tree mode
				if m.showTreeView && m.treeView != nil {
					for i, child := range m.treeView.Root.Children {
						child.Selected = (i == m.resourceGroupIdx)
					}
				}

				return m, loadResourcesInGroupCmd(m.selectedGroup)
			}
		case "up", "k":
			if len(m.resourceGroups) > 0 {
				m.resourceGroupIdx = (m.resourceGroupIdx - 1 + len(m.resourceGroups)) % len(m.resourceGroups)
				// Auto-scroll if needed
				if m.resourceGroupIdx < m.groupScrollOffset {
					m.groupScrollOffset = m.resourceGroupIdx
				}
				m.selectedGroup = m.resourceGroups[m.resourceGroupIdx].Name

				// Update tree view selection if in tree mode
				if m.showTreeView && m.treeView != nil {
					for i, child := range m.treeView.Root.Children {
						child.Selected = (i == m.resourceGroupIdx)
					}
				}

				return m, loadResourcesInGroupCmd(m.selectedGroup)
			}
		case "right":
			if len(m.resourcesInGroup) > 0 {
				m.resourceIdx = (m.resourceIdx + 1) % len(m.resourcesInGroup)
				// Auto-scroll if needed
				if m.resourceIdx >= m.resourceScrollOffset+m.maxVisibleResources {
					m.resourceScrollOffset = m.resourceIdx - m.maxVisibleResources + 1
				}
				m.selectedResource = m.resourcesInGroup[m.resourceIdx].Name
			}
		case "left":
			if len(m.resourcesInGroup) > 0 {
				m.resourceIdx = (m.resourceIdx - 1 + len(m.resourcesInGroup)) % len(m.resourcesInGroup)
				// Auto-scroll if needed
				if m.resourceIdx < m.resourceScrollOffset {
					m.resourceScrollOffset = m.resourceIdx
				}
				m.selectedResource = m.resourcesInGroup[m.resourceIdx].Name
			}
		case " ", "space":
			// Expand/collapse tree nodes (vim-like space functionality)
			if m.showTreeView && len(m.resourceGroups) > 0 && m.resourceGroupIdx < len(m.resourceGroups) {
				// Find the corresponding tree node and toggle expansion
				for _, child := range m.treeView.Root.Children {
					if child.Name == m.selectedGroup {
						child.Expanded = !child.Expanded
						break
					}
				}
			}
			return m, nil
		case "d":
			// Show details for selected resource
			if m.selectedResource != "" && m.resourceIdx < len(m.resourcesInGroup) {
				resource := m.resourcesInGroup[m.resourceIdx]
				details, err := fetchResourceDetails(resource.ID)
				if err != nil {
					m.selectedResource = resource.Name + " (details error: " + err.Error() + ")"
				} else {
					m.selectedResource = resource.Name + "\n" + details
				}
			}
		case "ctrl+k":
			// Load AKS clusters
			m.aksLoading = true
			return m, loadAKSClustersCmd()
		case "K":
			// Open AKS connection tab (simulate connection)
			if len(m.aksClusters) > 0 {
				aks := m.aksClusters[0] // For demo, use first
				title := fmt.Sprintf("AKS: %s", aks.Name)
				content := fmt.Sprintf("Connected to AKS cluster: %s\nResource Group: %s\nLocation: %s", aks.Name, aks.ResourceGroup, aks.Location)
				m.tabManager.AddTab(tui.Tab{Title: title, Content: content, Type: "aks", Meta: map[string]string{"name": aks.Name}, Closable: true})
			}
			return m, nil
		case "ctrl+V":
			// Open VM connection tab (simulate connection)
			if len(m.resourcesInGroup) > 0 {
				for _, r := range m.resourcesInGroup {
					if r.Type == "Microsoft.Compute/virtualMachines" {
						title := fmt.Sprintf("VM: %s", r.Name)
						content := fmt.Sprintf("Connected to VM: %s\nResource ID: %s\nLocation: %s", r.Name, r.ID, r.Location)
						m.tabManager.AddTab(tui.Tab{Title: title, Content: content, Type: "vm", Meta: map[string]string{"id": r.ID}, Closable: true})
					}
				}
			}
			return m, nil
		case "V":
			// Start interactive Key Vault creation prompt
			m.promptingKeyVault = true
			m.promptKeyVaultStep = 0
			m.promptKeyVaultMsg = "Enter Key Vault name:"
			return m, nil
		case "X":
			// Delete selected Key Vault
			if len(m.keyVaults) > 0 {
				go deleteKeyVault(m.keyVaults[0].Name, m.keyVaults[0].ResourceGroup)
				m.keyVaultsLoading = true
				return m, loadKeyVaultsCmd()
			}
		case "s":
			// Load Storage Accounts
			m.storageLoading = true
			return m, loadStorageAccountsCmd()
		case "S":
			// Start interactive Storage Account creation prompt
			m.promptingStorage = true
			m.promptStorageStep = 0
			m.promptStorageMsg = "Enter Storage Account name:"
			return m, nil
		case "Y":
			// Delete selected Storage Account
			if len(m.storageAccounts) > 0 {
				go deleteStorageAccount(m.storageAccounts[0].Name, m.storageAccounts[0].ResourceGroup)
				m.storageLoading = true
				return m, loadStorageAccountsCmd()
			}
		case "m":
			// Show usage matrix popup
			m.showMatrixPopup = true
			m.matrixViewport = viewport.New(80, 20)
			m.matrixViewport.SetContent(renderUsageMatrix(m.usageHeaders, m.usageMatrix))
			return m, nil
		case "A":
			// Show alarms popup
			m.showAlarmsPopup = true
			m.alarmsViewport = viewport.New(80, 20)
			m.alarmsViewport.SetContent(renderAlarms(m.alarms))
			return m, nil
		case "esc":
			// Hide popups
			m.showMatrixPopup = false
			m.showAlarmsPopup = false
			m.showIacPanel = false
			m.showIacFilePopup = false
			m.showCreatePopup = false
			m.showDeployPopup = false
			m.showShortcutsPopup = false
			m.showAIPopup = false
			m.showMetricsDialog = false
			m.showEditDialog = false
			m.showDeleteDialog = false
			m.showResourceActions = false
			return m, nil
		case "F":
			// Prompt for directory to scan (for now, use current dir or last used)
			m.iacScanLoading = true
			m.iacDir = "." // TODO: prompt user for dir
			return m, scanIaCFilesCmd(m.iacDir)
		case "i":
			// Toggle IaC file panel
			m.showIacPanel = !m.showIacPanel
			return m, nil
		case "n":
			// Handle multiple 'n' scenarios
			if m.showDeleteDialog {
				// Cancel delete action
				m.showDeleteDialog = false
			} else if m.showIacPanel && len(m.iacFiles) > 0 {
				// Navigate IaC files
				m.selectedIacIdx = (m.selectedIacIdx + 1) % len(m.iacFiles)
			}
		case "p":
			if m.showIacPanel && len(m.iacFiles) > 0 {
				m.selectedIacIdx = (m.selectedIacIdx - 1 + len(m.iacFiles)) % len(m.iacFiles)
			}
		case "c":
			// Start resource creation workflow
			m.creatingResource = true
			m.createStep = 0
			m.createAIMessage = "AI: What type of resource do you want to create? (e.g. vm, storage, vnet)"
			m.showCreatePopup = true
			return m, nil
		case "y":
			// Handle multiple 'y' scenarios
			if m.showDeleteDialog {
				// Confirm delete action
				m.showDeleteDialog = false
				// Here you would actually delete the resource
				// For now, just show a message
				m.aiMessage = fmt.Sprintf("Resource '%s' would be deleted here. (Demo mode - no actual deletion)", m.editingResourceName)
				m.showAIPopup = true
			} else if m.showIacPanel && len(m.iacFiles) > 0 {
				// Deploy selected IaC file
				m.deployingResource = true
				m.deployOutput = ""
				m.showDeployPopup = true
				go runIaCDeployment(m.iacFiles[m.selectedIacIdx].Path, m.iacFiles[m.selectedIacIdx].Type, m)
			}
			return m, nil
		case "r":
			// Force refresh resource groups from Azure for current subscription
			if len(m.subscriptions) > 0 {
				subID := m.subscriptions[m.currentSub].ID
				m.isLoading = true
				return m, loadResourcesCmd(subID)
			}
			return m, nil
		// --- Tab/Window and Shortcuts Popup ---
		case "ctrl+t":
			// Open a new tab (demo: blank tab, real: could prompt for type)
			title := fmt.Sprintf("Tab %d", len(m.tabManager.Tabs)+1)
			m.tabManager.AddTab(tui.Tab{Title: title, Content: "New tab opened.", Type: "blank", Closable: true})
			return m, nil
		case "?":
			// Show shortcuts popup
			m.showShortcutsPopup = true
			return m, nil
		case "F2":
			// Toggle between tree view and traditional tab interface
			m.showTreeView = !m.showTreeView
			return m, nil
		case "a":
			// Show AI analysis for selected resource
			if m.activeTabIdx == 0 && len(m.resourcesInGroup) > 0 && m.resourceIdx < len(m.resourcesInGroup) {
				res := m.resourcesInGroup[m.resourceIdx]
				if m.aiProvider != nil {
					m.aiLoading = true
					m.showAIPopup = true
					go func() {
						details, _ := fetchResourceDetails(res.ID)
						analysis, err := m.aiProvider.DescribeResource(res.Type, res.Name, details)
						if err != nil {
							m.aiMessage = "AI analysis failed: " + err.Error()
						} else {
							m.aiMessage = analysis
						}
						m.aiLoading = false
					}()
				} else {
					m.aiMessage = "AI provider not configured. Set OPENAI_API_KEY environment variable."
					m.showAIPopup = true
				}
			}
			return m, nil
		case "M":
			// Show metrics dashboard for selected resource
			if m.activeTabIdx == 0 && len(m.resourcesInGroup) > 0 && m.resourceIdx < len(m.resourcesInGroup) {
				res := m.resourcesInGroup[m.resourceIdx]
				m.showMetricsDialog = true
				m.editingResourceName = res.Name
				m.editingResourceType = res.Type
				// Generate demo metrics data
				m.resourceMetrics = map[string]interface{}{
					"cpu_usage":    75.5,
					"memory_usage": 82.3,
					"network_in":   12.5,
					"network_out":  8.7,
					"disk_read":    45.2,
					"disk_write":   23.1,
				}
			}
			return m, nil
		case "E":
			// Show edit dialog for selected resource
			if m.activeTabIdx == 0 && len(m.resourcesInGroup) > 0 && m.resourceIdx < len(m.resourcesInGroup) {
				res := m.resourcesInGroup[m.resourceIdx]
				m.showEditDialog = true
				m.editingResourceName = res.Name
				m.editingResourceType = res.Type
				// Generate demo config data
				m.currentResourceConfig = map[string]string{
					"Name":           res.Name,
					"Type":           res.Type,
					"Location":       res.Location,
					"Resource Group": m.selectedGroup,
					"Status":         "Running",
				}
			}
			return m, nil
		case "D":
			// Show delete confirmation dialog
			if m.activeTabIdx == 0 && len(m.resourcesInGroup) > 0 && m.resourceIdx < len(m.resourcesInGroup) {
				res := m.resourcesInGroup[m.resourceIdx]
				m.showDeleteDialog = true
				m.editingResourceName = res.Name
				m.editingResourceType = res.Type
			}
			return m, nil
		case "T":
			// Generate Terraform code for selected resource
			if m.activeTabIdx == 0 && len(m.resourcesInGroup) > 0 && m.resourceIdx < len(m.resourcesInGroup) {
				res := m.resourcesInGroup[m.resourceIdx]
				if m.aiProvider != nil {
					m.aiLoading = true
					m.showAIPopup = true
					go func() {
						requirements := fmt.Sprintf("Resource: %s\nType: %s\nLocation: %s", res.Name, res.Type, res.Location)
						code, err := m.aiProvider.GenerateTerraformCode(res.Type, requirements)
						if err != nil {
							m.aiMessage = "Terraform generation failed: " + err.Error()
						} else {
							m.aiMessage = "Generated Terraform Code:\n\n" + code
						}
						m.aiLoading = false
					}()
				} else {
					m.aiMessage = "AI provider not configured. Set OPENAI_API_KEY environment variable."
					m.showAIPopup = true
				}
			}
			return m, nil
		case "B":
			// Generate Bicep code for selected resource
			if m.activeTabIdx == 0 && len(m.resourcesInGroup) > 0 && m.resourceIdx < len(m.resourcesInGroup) {
				res := m.resourcesInGroup[m.resourceIdx]
				if m.aiProvider != nil {
					m.aiLoading = true
					m.showAIPopup = true
					go func() {
						requirements := fmt.Sprintf("Resource: %s\nType: %s\nLocation: %s", res.Name, res.Type, res.Location)
						code, err := m.aiProvider.GenerateBicepCode(res.Type, requirements)
						if err != nil {
							m.aiMessage = "Bicep generation failed: " + err.Error()
						} else {
							m.aiMessage = "Generated Bicep Code:\n\n" + code
						}
						m.aiLoading = false
					}()
				} else {
					m.aiMessage = "AI provider not configured. Set OPENAI_API_KEY environment variable."
					m.showAIPopup = true
				}
			}
			return m, nil
		case "O":
			// Cost optimization analysis for current resource group
			if m.aiProvider != nil && len(m.resourcesInGroup) > 0 {
				m.aiLoading = true
				m.showAIPopup = true
				go func() {
					var resources []string
					resourceDetails := make(map[string]string)
					for _, res := range m.resourcesInGroup {
						resources = append(resources, res.Name)
						details, _ := fetchResourceDetails(res.ID)
						resourceDetails[res.Name] = details
					}
					optimization, err := m.aiProvider.SuggestCostOptimizations(resources, resourceDetails)
					if err != nil {
						m.aiMessage = "Cost optimization analysis failed: " + err.Error()
					} else {
						m.aiMessage = "Cost Optimization Suggestions:\n\n" + optimization
					}
					m.aiLoading = false
				}()
			} else if m.aiProvider == nil {
				m.aiMessage = "AI provider not configured. Set OPENAI_API_KEY environment variable."
				m.showAIPopup = true
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *model) View() string {
	// Handle loading state
	if m.loading {
		loadingText := "🔄 Loading Azure resources...\n\nConnecting to Azure CLI..."
		if m.loadErr != "" {
			loadingText = fmt.Sprintf("❌ Error loading Azure resources:\n\n%s\n\nPlease ensure you're logged in with 'az login'", m.loadErr)
		}

		content := lipgloss.NewStyle().
			Width(m.termWidth-4).
			Height(m.termHeight-4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("63")).
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(2, 4).
			Align(lipgloss.Center, lipgloss.Center).
			Render(loadingText)

		return renderMainBoxLipgloss(content, m.termWidth, m.termHeight)
	}

	// Handle error state when not loading
	if m.loadErr != "" {
		errorText := fmt.Sprintf("❌ Azure Connection Error\n\n%s\n\nTroubleshooting:\n• Run 'az login' to authenticate\n• Check your internet connection\n• Verify you have access to Azure subscriptions\n\nPress 'q' to quit", m.loadErr)

		content := lipgloss.NewStyle().
			Width(m.termWidth-4).
			Height(m.termHeight-4).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("196")).
			Background(lipgloss.Color("236")).
			Foreground(lipgloss.Color("252")).
			Padding(2, 4).
			Align(lipgloss.Center, lipgloss.Center).
			Render(errorText)

		return renderMainBoxLipgloss(content, m.termWidth, m.termHeight)
	}

	// Update status bar with current subscription and tenant information
	if m.statusBar != nil {
		// Clear existing segments
		m.statusBar.Segments = []tui.PowerlineSegment{}
		m.statusBar.RightAlign = []tui.PowerlineSegment{}

		// Always show Azure TUI branding
		m.statusBar.AddSegment("🚀 Azure TUI", lipgloss.Color("39"), lipgloss.Color("15"))

		// Add subscription info if available
		if len(m.subscriptions) > 0 && m.currentSub < len(m.subscriptions) {
			sub := m.subscriptions[m.currentSub]
			m.statusBar.AddSegment("☁️ "+sub.Name, lipgloss.Color("33"), lipgloss.Color("15"))
		} else {
			m.statusBar.AddSegment("☁️ Loading...", lipgloss.Color("240"), lipgloss.Color("15"))
		}

		// Add tenant info
		if len(m.tenants) > 0 && m.currentTenant < len(m.tenants) {
			tenant := m.tenants[m.currentTenant]
			m.statusBar.AddSegment("🏢 "+tenant.Name, lipgloss.Color("33"), lipgloss.Color("15"))
		}

		// Add resource group count
		if len(m.resourceGroups) > 0 {
			resourceText := fmt.Sprintf("📁 %d groups", len(m.resourceGroups))
			m.statusBar.AddSegment(resourceText, lipgloss.Color("63"), lipgloss.Color("15"))
		} else {
			m.statusBar.AddSegment("📁 Loading groups...", lipgloss.Color("240"), lipgloss.Color("15"))
		}

		// Add right-aligned segments
		m.statusBar.AddRightSegment("⌨️ Press ? for help", lipgloss.Color("240"), lipgloss.Color("252"))
	}

	// Create main layout with tree view on left and content tabs on right
	var mainContent string

	if m.showTreeView && m.treeView != nil {
		// Tree view + content tabs layout (NeoVim-style)
		leftPanelHeight := m.termHeight - 3 // Account for status bar
		treeViewContent := m.treeView.RenderTreeView(m.leftPanelWidth, leftPanelHeight)

		// Right panel for content tabs
		rightPanelWidth := m.termWidth - m.leftPanelWidth - 2
		if rightPanelWidth < 20 {
			rightPanelWidth = 20
		}

		var rightContent string
		if len(m.contentTabs) > 0 && m.activeContentTab < len(m.contentTabs) {
			// Render active content tab
			activeTab := m.contentTabs[m.activeContentTab]
			rightContent = lipgloss.NewStyle().
				Width(rightPanelWidth).
				Height(leftPanelHeight).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("252")).
				Padding(1, 2).
				Render(activeTab.Content)
		} else {
			// Show welcome message in right panel
			welcomeText := "Welcome to Azure TUI\n\n" +
				"TREE VIEW INTERFACE\n\n" +
				"Navigate with:\n" +
				"• j/k or ↑↓ - Navigate tree\n" +
				"• Space - Expand/collapse\n" +
				"• Enter - Open resource\n" +
				"• ? - Show all shortcuts\n\n" +
				"STATUS: Tree view mode active"

			rightContent = lipgloss.NewStyle().
				Width(rightPanelWidth).
				Height(leftPanelHeight).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("63")).
				Background(lipgloss.Color("236")).
				Foreground(lipgloss.Color("252")).
				Padding(1, 2).
				Render(welcomeText)
		}

		// Combine tree view and content panels horizontally
		mainContent = lipgloss.JoinHorizontal(lipgloss.Top, treeViewContent, rightContent)
	} else {
		// Fallback to traditional tab layout for resource tabs
		tabs := []tui.Tab{{Title: "Resource Browser", Closable: false, Type: "resourcegroup"}}
		tabs = append(tabs, m.resourceTabs...)
		tabBar := tui.RenderTabsWithActive(tabs, m.activeTabIdx)

		var content string
		if m.activeTabIdx == 0 {
			// Main resource browser: left = groups, right = resources
			leftPanel := renderLeftPanel(m)
			rightPanel := renderRightPanel(m)
			panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)
			content = lipgloss.JoinVertical(lipgloss.Left, panels)
		} else {
			// Resource tab: show details for that resource
			resTab := m.resourceTabs[m.activeTabIdx-1]
			content = resTab.Content
		}

		mainContent = tabBar + "\n" + content
	}

	// Render status bar
	var statusBarContent string
	if m.statusBar != nil {
		statusBarContent = m.statusBar.RenderStatusBar()
	}

	// Combine main content with status bar (status bar at top)
	fullView := lipgloss.JoinVertical(lipgloss.Left, statusBarContent, mainContent)

	baseView := renderMainBoxLipgloss(fullView, m.termWidth, m.termHeight)

	// Show dialogs and popups on top of the main view
	if m.showAIPopup {
		var aiContent string
		if m.aiLoading {
			aiContent = "🤖 AI is analyzing... Please wait."
		} else {
			aiContent = m.aiMessage
		}
		popup := tui.RenderPopup(tui.PopupMsg{
			Title:   "AI Analysis",
			Content: aiContent,
			Level:   "info",
		})
		return baseView + "\n\n" + popup
	}

	if m.showMetricsDialog {
		metricsContent := tui.RenderMetricsDashboard(m.editingResourceName, m.resourceMetrics)
		return baseView + "\n\n" + metricsContent
	}

	if m.showEditDialog {
		editContent := tui.RenderEditDialog(m.editingResourceName, m.editingResourceType, m.currentResourceConfig)
		return baseView + "\n\n" + editContent
	}

	if m.showDeleteDialog {
		deleteContent := tui.RenderDeleteConfirmation(m.editingResourceName, m.editingResourceType)
		return baseView + "\n\n" + deleteContent
	}

	if m.showResourceActions {
		actionsContent := tui.RenderResourceActions(m.editingResourceType, m.editingResourceName)
		return baseView + "\n\n" + actionsContent
	}

	if m.showShortcutsPopup {
		shortcuts := map[string]string{
			"j/k,↑/↓": "Navigate tree/resources",
			"Space":   "Expand/collapse tree",
			"Enter":   "Open resource tab",
			"Tab":     "Switch tabs",
			"Ctrl+W":  "Close tab",
			"F2":      "Toggle tree/traditional",
			"a":       "AI analysis",
			"M":       "Metrics dashboard",
			"E":       "Edit resource",
			"Ctrl+D":  "Delete resource",
			"T":       "Generate Terraform",
			"B":       "Generate Bicep",
			"O":       "Cost optimization",
			"Ctrl+K":  "Load AKS clusters",
			"?":       "Show shortcuts",
			"Esc":     "Close popups",
			"q":       "Quit",
		}
		shortcutsContent := tui.RenderShortcutsPopup(shortcuts)
		return baseView + "\n\n" + shortcutsContent
	}

	return baseView
}

// --- Modern Panel Renderers ---
var (
	panelBorder      = lipgloss.RoundedBorder()
	panelBg          = lipgloss.Color("236")
	panelFg          = lipgloss.Color("252")
	panelBorderColor = lipgloss.Color("63")
	selectedBg       = lipgloss.Color("33")
	selectedFg       = lipgloss.Color("230")
)

func renderLeftPanel(m *model) string {
	// Calculate responsive panel dimensions
	panelWidth := (m.termWidth - 8) / 2 // Split terminal width in half, minus padding
	panelHeight := m.termHeight - 8     // Terminal height minus margins

	// Ensure minimum sizes
	if panelWidth < 30 {
		panelWidth = 30
	}
	if panelHeight < 15 {
		panelHeight = 15
	}

	style := lipgloss.NewStyle().
		Width(panelWidth).
		Height(panelHeight).
		Border(panelBorder).
		BorderForeground(panelBorderColor).
		Background(panelBg).
		Foreground(panelFg).
		Padding(1, 2).
		AlignHorizontal(lipgloss.Left)

	var lines []string
	folderIcon := "🗂️"

	if m.loading || m.isLoading {
		lines = append(lines, "🔄 Loading Azure resources...")
	} else if m.resourceLoadErr != "" {
		lines = append(lines, "❌ "+m.resourceLoadErr)
		lines = append(lines, "")
		lines = append(lines, "📝 Running in demo mode")
		lines = append(lines, "Press ? for help")
	} else if len(m.resourceGroups) == 0 {
		lines = append(lines, "🔍 No resource groups found.")
		lines = append(lines, "")
		lines = append(lines, "📝 Running in demo mode")
		lines = append(lines, "This might indicate:")
		lines = append(lines, "• Azure CLI not configured")
		lines = append(lines, "• No subscriptions available")
		lines = append(lines, "• Network connectivity issues")
		lines = append(lines, "")
		lines = append(lines, "Press ? for keyboard shortcuts")
	} else {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Resource Groups"))

		// Show scroll indicator if needed
		totalGroups := len(m.resourceGroups)
		if totalGroups > m.maxVisibleGroups {
			scrollInfo := fmt.Sprintf("(%d of %d groups - scroll with ↑↓)",
				min(m.groupScrollOffset+m.maxVisibleGroups, totalGroups), totalGroups)
			lines = append(lines, lipgloss.NewStyle().Faint(true).Render(scrollInfo))
		} else {
			lines = append(lines, lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("(%d groups)", totalGroups)))
		}
		lines = append(lines, "")

		// Calculate visible range
		startIdx := m.groupScrollOffset
		endIdx := min(startIdx+m.maxVisibleGroups, len(m.resourceGroups))

		// Render visible resource groups with scrolling
		for i := startIdx; i < endIdx; i++ {
			rg := m.resourceGroups[i]
			prefix := "  "
			if i == m.resourceGroupIdx {
				prefix = "→ "
			}

			// Truncate long names to fit panel
			name := rg.Name
			maxNameLength := panelWidth - 10 // Account for prefix, icon, padding
			if len(name) > maxNameLength {
				name = name[:maxNameLength-3] + "..."
			}

			line := fmt.Sprintf("%s%s %s", prefix, folderIcon, name)
			if i == m.resourceGroupIdx {
				line = lipgloss.NewStyle().Background(selectedBg).Foreground(selectedFg).Render(line)
			}
			lines = append(lines, line)
		}

		// Add scroll indicators at bottom if needed
		if totalGroups > m.maxVisibleGroups {
			if m.groupScrollOffset > 0 && endIdx < totalGroups {
				lines = append(lines, lipgloss.NewStyle().Faint(true).Render("... ↑ more above, ↓ more below ..."))
			} else if m.groupScrollOffset > 0 {
				lines = append(lines, lipgloss.NewStyle().Faint(true).Render("... ↑ more above ..."))
			} else if endIdx < totalGroups {
				lines = append(lines, lipgloss.NewStyle().Faint(true).Render("... ↓ more below ..."))
			}
		}
	}

	content := strings.Join(lines, "\n")
	return style.Render(content)
}

func renderRightPanel(m *model) string {
	// Calculate responsive panel dimensions
	panelWidth := (m.termWidth - 8) / 2 // Split terminal width in half, minus padding
	panelHeight := m.termHeight - 8     // Terminal height minus margins

	// Ensure minimum sizes
	if panelWidth < 30 {
		panelWidth = 30
	}
	if panelHeight < 15 {
		panelHeight = 15
	}

	style := lipgloss.NewStyle().
		Width(panelWidth).
		Height(panelHeight).
		Border(panelBorder).
		BorderForeground(panelBorderColor).
		Background(panelBg).
		Foreground(panelFg).
		Padding(1, 2).
		AlignHorizontal(lipgloss.Left)

	azureIcons := map[string]string{
		"virtualmachines":   "🖥️",
		"keyvault":          "🔑",
		"storageaccounts":   "💾",
		"networkinterfaces": "🔌",
		"publicipaddresses": "🌐",
		"virtualnetworks":   "🔗",
		"disks":             "💽",
		"actiongroups":      "🚨",
		"metricalerts":      "📊",
		"extensions":        "🧩",
		"default":           "📦",
	}

	var lines []string

	if m.selectedGroup == "" || len(m.resourcesInGroup) == 0 {
		if m.selectedGroup != "" {
			lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Resources"))
			lines = append(lines, "")
			lines = append(lines, "Loading resources...")
		} else {
			lines = append(lines, "Select a resource group")
			lines = append(lines, "to see its resources")
		}
	} else {
		lines = append(lines, lipgloss.NewStyle().Bold(true).Render("Resources"))
		lines = append(lines, lipgloss.NewStyle().Faint(true).Render("in "+m.selectedGroup))
		lines = append(lines, "")

		for i, r := range m.resourcesInGroup {
			prefix := "  "
			if i == m.resourceIdx {
				prefix = "→ "
			}

			// Get icon by type
			icon := azureIcons["default"]
			for k, v := range azureIcons {
				if k != "default" && strings.Contains(strings.ToLower(r.Type), k) {
					icon = v
					break
				}
			}

			// Truncate name if too long
			name := r.Name
			if len(name) > 25 {
				name = name[:22] + "..."
			}

			line := fmt.Sprintf("%s%s %s", prefix, icon, name)
			if i == m.resourceIdx {
				line = lipgloss.NewStyle().Background(selectedBg).Foreground(selectedFg).Render(line)
			}
			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n")
	return style.Render(content)
}

// --- Modern Main Box with Lipgloss ---
func renderMainBoxLipgloss(content string, termWidth, termHeight int) string {
	// Use full terminal size minus small margins
	width := termWidth - 2
	height := termHeight - 2

	// Ensure minimum size
	if width < 40 {
		width = 40
	}
	if height < 10 {
		height = 10
	}

	boxStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Left, lipgloss.Top).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("252")).
		Padding(1, 2)
	return boxStyle.Render(content)
}

// centerBox centers the given string in a box using responsive dimensions.
func centerBox(content string, termWidth, termHeight int) string {
	width := termWidth - 4
	height := termHeight - 4

	// Ensure minimum size
	if width < 40 {
		width = 40
	}
	if height < 10 {
		height = 10
	}

	style := lipgloss.NewStyle().Width(width).Height(height).Align(lipgloss.Center, lipgloss.Center)
	return style.Render(content)
}

// fetchAzureSubsAndTenants uses Azure CLI to get subscriptions and tenants as a quick cross-platform solution.
func fetchAzureSubsAndTenants() ([]Subscription, []Tenant, error) {
	subs := []Subscription{}
	tenants := []Tenant{}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get subscriptions from Azure CLI
	subCmd := exec.CommandContext(ctx, "az", "account", "list", "--output", "json")
	subOut, err := subCmd.Output()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get Azure subscriptions: %v. Make sure you're logged in with 'az login'", err)
	}
	if ctx.Err() != nil {
		return nil, nil, fmt.Errorf("timeout getting Azure subscriptions. Please check your connection")
	}

	if err := json.Unmarshal(subOut, &subs); err != nil {
		return nil, nil, fmt.Errorf("failed to parse subscription data: %v", err)
	}

	if len(subs) == 0 {
		return nil, nil, fmt.Errorf("no Azure subscriptions found. Please check your access")
	}

	// Get tenants from Azure CLI
	ctx2, cancel2 := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel2()
	tenantCmd := exec.CommandContext(ctx2, "az", "account", "tenant", "list", "--output", "json")
	tenantOut, err := tenantCmd.Output()
	if err == nil && ctx2.Err() == nil {
		_ = json.Unmarshal(tenantOut, &tenants)
	}
	if len(tenants) == 0 {
		tenants = []Tenant{{ID: "default-tenant", Name: "Default Tenant"}}
	}

	return subs, tenants, nil
}

// loadResourcesCmd loads resource groups for the given subscription.
func loadResourcesCmd(subID string) tea.Cmd {
	return func() tea.Msg {
		groups, err := fetchResourceGroups(subID)
		if err != nil {
			return resourceLoadErrMsg(err.Error())
		}
		return resourcesMsg(groups)
	}
}

// fetchResourceGroups uses Azure CLI directly to get resource groups for the current subscription.
func fetchResourceGroups(subID string) ([]ResourceGroup, error) {
	// Try Azure CLI first with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "group", "list", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("Azure CLI timeout after 10 seconds. Check your internet connection")
		}
		return nil, fmt.Errorf("failed to get resource groups via Azure CLI: %v. Make sure you're logged in with 'az login'", err)
	}

	var azGroups []struct {
		Name     string `json:"name"`
		Location string `json:"location"`
	}
	if err := json.Unmarshal(output, &azGroups); err != nil {
		return nil, fmt.Errorf("failed to parse resource group data: %v", err)
	}

	if len(azGroups) == 0 {
		return nil, fmt.Errorf("no resource groups found in current subscription")
	}

	var result []ResourceGroup
	for _, g := range azGroups {
		result = append(result, ResourceGroup{
			Name:     g.Name,
			Location: g.Location,
		})
	}
	return result, nil
}

// getDemoResourceGroups returns sample resource groups for demo mode
func getDemoResourceGroups() []ResourceGroup {
	return []ResourceGroup{
		{Name: "prod-webapp-rg", Location: "westeurope"},
		{Name: "dev-environment-rg", Location: "eastus"},
		{Name: "data-analytics-rg", Location: "westus2"},
		{Name: "monitoring-rg", Location: "northeurope"},
		{Name: "backup-storage-rg", Location: "centralus"},
	}
}

// getDemoResourcesForGroup returns sample resources for a given demo resource group
func getDemoResourcesForGroup(groupName string) []AzureResource {
	switch groupName {
	case "prod-webapp-rg":
		return []AzureResource{
			{ID: "demo-webapp-01", Name: "webapp-frontend", Type: "Microsoft.Web/sites", Location: "westeurope"},
			{ID: "demo-sql-01", Name: "webapp-database", Type: "Microsoft.Sql/servers", Location: "westeurope"},
			{ID: "demo-redis-01", Name: "webapp-cache", Type: "Microsoft.Cache/Redis", Location: "westeurope"},
			{ID: "demo-storage-01", Name: "webappstorageacct", Type: "Microsoft.Storage/storageAccounts", Location: "westeurope"},
			{ID: "demo-keyvault-01", Name: "webapp-secrets", Type: "Microsoft.KeyVault/vaults", Location: "westeurope"},
		}
	case "dev-environment-rg":
		return []AzureResource{
			{ID: "demo-vm-01", Name: "dev-jumpbox", Type: "Microsoft.Compute/virtualMachines", Location: "eastus"},
			{ID: "demo-aks-01", Name: "dev-k8s-cluster", Type: "Microsoft.ContainerService/managedClusters", Location: "eastus"},
			{ID: "demo-acr-01", Name: "devcontainerregistry", Type: "Microsoft.ContainerRegistry/registries", Location: "eastus"},
			{ID: "demo-vnet-01", Name: "dev-virtual-network", Type: "Microsoft.Network/virtualNetworks", Location: "eastus"},
		}
	case "data-analytics-rg":
		return []AzureResource{
			{ID: "demo-cosmos-01", Name: "analytics-cosmosdb", Type: "Microsoft.DocumentDB/databaseAccounts", Location: "westus2"},
			{ID: "demo-datafactory-01", Name: "analytics-pipeline", Type: "Microsoft.DataFactory/factories", Location: "westus2"},
			{ID: "demo-synapse-01", Name: "analytics-workspace", Type: "Microsoft.Synapse/workspaces", Location: "westus2"},
			{ID: "demo-storage-02", Name: "datalakestorage", Type: "Microsoft.Storage/storageAccounts", Location: "westus2"},
		}
	case "monitoring-rg":
		return []AzureResource{
			{ID: "demo-loganalytics-01", Name: "central-logs", Type: "Microsoft.OperationalInsights/workspaces", Location: "northeurope"},
			{ID: "demo-appinsights-01", Name: "app-monitoring", Type: "microsoft.insights/components", Location: "northeurope"},
			{ID: "demo-alerts-01", Name: "critical-alerts", Type: "microsoft.insights/actiongroups", Location: "northeurope"},
		}
	case "backup-storage-rg":
		return []AzureResource{
			{ID: "demo-vault-01", Name: "backup-vault", Type: "Microsoft.RecoveryServices/vaults", Location: "centralus"},
			{ID: "demo-storage-03", Name: "backupstorage", Type: "Microsoft.Storage/storageAccounts", Location: "centralus"},
		}
	default:
		return []AzureResource{
			{ID: "demo-resource-default", Name: "sample-resource", Type: "Microsoft.Resources/resourceGroups", Location: "westeurope"},
		}
	}
}

// loadResourcesInGroupCmd loads resources for a given resource group.
func loadResourcesInGroupCmd(groupName string) tea.Cmd {
	return func() tea.Msg {
		resources, err := fetchResourcesInGroup(groupName)
		if err != nil {
			return resourcesInGroupErrMsg{groupName, err.Error()}
		}
		return resourcesInGroupMsg{groupName, resources}
	}
}

// fetchResourcesInGroup uses Azure CLI to get resources in a resource group.
func fetchResourcesInGroup(groupName string) ([]AzureResource, error) {
	cmd := exec.Command("az", "resource", "list", "--resource-group", groupName, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get resources for group %s: %v", groupName, err)
	}

	var resources []AzureResource
	if err := json.Unmarshal(out, &resources); err != nil {
		return nil, fmt.Errorf("failed to parse resources for group %s: %v", groupName, err)
	}

	return resources, nil
}

// fetchResourceDetails uses Azure CLI to get details for a specific resource by ID.
func fetchResourceDetails(resourceID string) (string, error) {
	cmd := exec.Command("az", "resource", "show", "--ids", resourceID, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// loadAKSClustersCmd loads AKS clusters in the current subscription.
func loadAKSClustersCmd() tea.Cmd {
	return func() tea.Msg {
		clusters, err := fetchAKSClusters()
		if err != nil {
			return aksClusterErrMsg(err.Error())
		}
		return aksClustersMsg(clusters)
	}
}

// fetchAKSClusters uses Azure CLI to get AKS clusters.
func fetchAKSClusters() ([]AKSCluster, error) {
	cmd := exec.Command("az", "aks", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var clusters []AKSCluster
	if err := json.Unmarshal(out, &clusters); err != nil {
		return nil, err
	}
	return clusters, nil
}

// createAKSCluster creates a new AKS cluster with user input.
func createAKSCluster(name, group, location string) {
	exec.Command("az", "aks", "create", "--name", name, "--resource-group", group, "--location", location, "--node-count", "1", "--generate-ssh-keys").Run()
}

// deleteAKSCluster deletes an AKS cluster by name and resource group.
func deleteAKSCluster(name, group string) {
	exec.Command("az", "aks", "delete", "--name", name, "--resource-group", group, "--yes", "--no-wait").Run()
}

// loadKeyVaultsCmd loads Key Vaults in the current subscription.
func loadKeyVaultsCmd() tea.Cmd {
	return func() tea.Msg {
		vaults, err := fetchKeyVaults()
		if err != nil {
			return keyVaultErrMsg(err.Error())
		}
		return keyVaultsMsg(vaults)
	}
}

// fetchKeyVaults uses Azure CLI to get Key Vaults.
func fetchKeyVaults() ([]struct {
	Name          string
	Location      string
	ResourceGroup string
}, error) {
	cmd := exec.Command("az", "keyvault", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	type vaultRaw struct {
		Name          string `json:"name"`
		Location      string `json:"location"`
		ResourceGroup string `json:"resourceGroup"`
	}
	var raw []vaultRaw
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}
	vaults := make([]struct {
		Name          string
		Location      string
		ResourceGroup string
	}, len(raw))
	for i, v := range raw {
		vaults[i] = struct {
			Name          string
			Location      string
			ResourceGroup string
		}{v.Name, v.Location, v.ResourceGroup}
	}
	return vaults, nil
}

// createKeyVault creates a new Key Vault with user input.
func createKeyVault(name, group, location string) {
}

// deleteKeyVault deletes a Key Vault by name and resource group.
func deleteKeyVault(name, group string) {
}

// loadStorageAccountsCmd loads Storage Accounts in the current subscription.
func loadStorageAccountsCmd() tea.Cmd {
	return func() tea.Msg {
		accounts, err := fetchStorageAccounts()
		if err != nil {
			return storageErrMsg(err.Error())
		}
		return storageAccountsMsg(accounts)
	}
}

// fetchStorageAccounts uses Azure CLI to get Storage Accounts.
func fetchStorageAccounts() ([]struct {
	Name          string
	Location      string
	ResourceGroup string
}, error) {
	cmd := exec.Command("az", "storage", "account", "list", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	type accountRaw struct {
		Name          string `json:"name"`
		Location      string `json:"location"`
		ResourceGroup string `json:"resourceGroup"`
	}
	var raw []accountRaw
	if err := json.Unmarshal(out, &raw); err != nil {
		return nil, err
	}
	accounts := make([]struct {
		Name          string
		Location      string
		ResourceGroup string
	}, len(raw))
	for i, a := range raw {
		accounts[i] = struct {
			Name          string
			Location      string
			ResourceGroup string
		}{a.Name, a.Location, a.ResourceGroup}
	}
	return accounts, nil
}

// createStorageAccount creates a new Storage Account with user input.
func createStorageAccount(name, group, location string) {
}

// deleteStorageAccount deletes a Storage Account by name and resource group.
func deleteStorageAccount(name, group string) {
}

// scanIaCFilesCmd scans a directory for IaC files.
func scanIaCFilesCmd(dir string) tea.Cmd {
	return func() tea.Msg {
		files, err := tfbicep.ScanIaCFiles(dir)
		if err != nil {
			return iacFilesErrMsg(err.Error())
		}
		var out []struct{ Path, Type string }
		for _, f := range files {
			out = append(out, struct{ Path, Type string }{f.Path, f.Type})
		}
		return iacFilesMsg(out)
	}
}

// setAzureSubscription sets the active Azure subscription using the Azure CLI.
func setAzureSubscription(subID string) {
	exec.Command("az", "account", "set", "--subscription", subID).Run()
}

// setAzureTenant sets the active Azure tenant using the Azure CLI.
func setAzureTenant(tenantID string) {
	exec.Command("az", "account", "tenant", "set", "--tenant", tenantID).Run()
}

// getActiveContext returns the current active subscription and tenant from Azure CLI.
func getActiveContext() (string, string) {
	cmd := exec.Command("az", "account", "show", "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return "?", "?"
	}
	var acc struct {
		Name     string `json:"name"`
		ID       string `json:"id"`
		TenantID string `json:"tenantId"`
	}
	if err := json.Unmarshal(out, &acc); err != nil {
		return "?", "?"
	}
	return acc.Name + " (" + acc.ID + ")", acc.TenantID
}

// Example usage: Summarize resource groups with AI
func summarizeResourceGroupsWithAI(groups []ResourceGroup) (string, error) {
	var names []string
	for _, g := range groups {
		names = append(names, g.Name)
	}
	return "", nil
}

// Example CLI command: List and summarize virtual networks
func listAndSummarizeVNetsCLI(subscriptionID, resourceGroup string) error {
	netClient, err := azuresdk.NewNetworkClient()
	if err != nil {
		return err
	}
	vnets, err := netClient.ListVirtualNetworks(subscriptionID, resourceGroup)
	if err != nil {
		return err
	}
	var vnetNames []string
	for _, v := range vnets {
		if v.Name != nil {
			vnetNames = append(vnetNames, *v.Name)
		}
	}
	return nil
}

// Render usage matrix
func renderUsageMatrix(headers []string, matrix [][]string) string {
	cols := make([]table.Column, len(headers))
	for i, h := range headers {
		cols[i] = table.Column{Title: h, Width: 16}
	}
	t := table.New(table.WithColumns(cols))
	rows := make([]table.Row, len(matrix))
	for i, row := range matrix {
		rows[i] = table.Row(row)
	}
	t.SetRows(rows)
	return t.View()
}

// Render alarms
func renderAlarms(alarms []struct {
	Name    string
	Status  string
	Details string
}) string {
	var b strings.Builder
	for _, a := range alarms {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")).Render(a.Name + ": " + a.Status))
		b.WriteString("\n" + a.Details + "\n\n")
	}
	return b.String()
}

// AI-powered log error summarization
func (m *model) summarizeResourceLogErrors(logs []string) string {
	return ""
}

// Helper function to read file preview
func readFilePreview(path string, maxLines int) string {
	f, err := os.Open(path)
	if err != nil {
		return "Error opening file: " + err.Error()
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() && len(lines) < maxLines {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "Error reading file: " + err.Error()
	}
	return strings.Join(lines, "\n")
}

// Add runIaCDeployment helper (scaffold):
func runIaCDeployment(path, typ string, m *model) {
	// TODO: Run terraform or bicep deployment, stream output to m.deployOutput, update m.showDeployPopup
	// On error, call AI provider to analyze and suggest fix
}

// Add config read for naming standards (scaffold):
func getNamingStandard() string {
	// TODO: Read from ~/.config/azure-tui/config.yaml
	return "demo-{{type}}-{{name}}"
}
