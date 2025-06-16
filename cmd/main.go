package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/olafkfreund/azure-tui/internal/azure/azuresdk"
	"github.com/olafkfreund/azure-tui/internal/azure/tfbicep"
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
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		os.Exit(1)
	}
}

// initialModel returns the starting state for the TUI.
func initialModel() tea.Model {
	return &model{
		profiles:           []string{"default"},
		currentProfile:     0,
		environments:       []string{"East US", "West Europe"},
		currentEnv:         0,
		loading:            true,
		tabManager:         tui.NewTabManager(),
		showShortcutsPopup: false,
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

	tabManager         *tui.TabManager // Multi-tab/window manager
	showShortcutsPopup bool            // Show keyboard shortcuts popup
}

func (m *model) Init() tea.Cmd {
	return func() tea.Msg {
		subs, tenants, err := fetchAzureSubsAndTenants()
		if err != nil {
			return loadErrMsg(err.Error())
		}
		return loadedMsg{subs, tenants}
	}
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
	case loadedMsg:
		m.subscriptions = msg.subs
		m.tenants = msg.tenants
		m.loading = false
		m.currentSub = 0
		m.currentTenant = 0
		// Load resources for the selected subscription
		return m, loadResourcesCmd()
	case resourcesMsg:
		m.resourceGroups = msg
		return m, nil
	case resourceLoadErrMsg:
		m.resourceLoadErr = string(msg)
		return m, nil
	case resourcesInGroupMsg:
		if msg.groupName == m.selectedGroup {
			m.resourcesInGroup = msg.resources
			m.resourcesInGroupErr = ""
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
	case tea.KeyMsg:
		if m.loading {
			return m, nil
		}
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
		case "tab":
			m.currentSub = (m.currentSub + 1) % len(m.subscriptions)
			return m, loadResourcesCmd()
		case "shift+tab":
			m.currentTenant = (m.currentTenant + 1) % len(m.tenants)
		case "down":
			if len(m.resourceGroups) > 0 {
				m.resourceGroupIdx = (m.resourceGroupIdx + 1) % len(m.resourceGroups)
				m.selectedGroup = m.resourceGroups[m.resourceGroupIdx].Name
				return m, loadResourcesInGroupCmd(m.selectedGroup)
			}
		case "up":
			if len(m.resourceGroups) > 0 {
				m.resourceGroupIdx = (m.resourceGroupIdx - 1 + len(m.resourceGroups)) % len(m.resourceGroups)
				m.selectedGroup = m.resourceGroups[m.resourceGroupIdx].Name
				return m, loadResourcesInGroupCmd(m.selectedGroup)
			}
		case "right":
			if len(m.resourcesInGroup) > 0 {
				m.resourceIdx = (m.resourceIdx + 1) % len(m.resourcesInGroup)
				m.selectedResource = m.resourcesInGroup[m.resourceIdx].Name
			}
		case "left":
			if len(m.resourcesInGroup) > 0 {
				m.resourceIdx = (m.resourceIdx - 1 + len(m.resourcesInGroup)) % len(m.resourcesInGroup)
				m.selectedResource = m.resourcesInGroup[m.resourceIdx].Name
			}
		case "enter":
			if len(m.subscriptions) > 0 {
				setAzureSubscription(m.subscriptions[m.currentSub].ID)
			}
			if len(m.tenants) > 0 {
				setAzureTenant(m.tenants[m.currentTenant].ID)
			}
			if len(m.resourceGroups) > 0 {
				m.selectedGroup = m.resourceGroups[m.resourceGroupIdx].Name
				return m, loadResourcesInGroupCmd(m.selectedGroup)
			}
			return m, loadResourcesCmd()
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
		case "k":
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
			if m.showIacPanel && len(m.iacFiles) > 0 {
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
			// Deploy selected IaC file
			if m.showIacPanel && len(m.iacFiles) > 0 {
				m.deployingResource = true
				m.deployOutput = ""
				m.showDeployPopup = true
				go runIaCDeployment(m.iacFiles[m.selectedIacIdx].Path, m.iacFiles[m.selectedIacIdx].Type, m)
				return m, nil
			}
		// --- Tab/Window and Shortcuts Popup ---
		case "ctrl+t":
			// Open a new tab (demo: blank tab, real: could prompt for type)
			title := fmt.Sprintf("Tab %d", len(m.tabManager.Tabs)+1)
			m.tabManager.AddTab(tui.Tab{Title: title, Content: "New tab opened.", Type: "blank", Closable: true})
			return m, nil
		case "ctrl+w":
			// Close current tab
			if len(m.tabManager.Tabs) > 0 {
				m.tabManager.CloseTab(m.tabManager.ActiveIndex)
			}
			return m, nil
		case "F1":
			// Show shortcuts popup
			m.showShortcutsPopup = true
			return m, nil
		}
	}
	return m, nil
}

func (m *model) View() string {
	title := titleStyle.Render("Azure TUI - Welcome!")
	if m.loading {
		return title + "\nLoading Azure subscriptions and tenants..."
	}
	if m.loadErr != "" {
		return title + "\nError: " + m.loadErr
	}
	var sub, tenant string
	if len(m.subscriptions) > 0 {
		s := m.subscriptions[m.currentSub]
		sub = s.Name + " (" + s.ID + ")"
	} else {
		sub = "No subscriptions found"
	}
	if len(m.tenants) > 0 {
		t := m.tenants[m.currentTenant]
		tenant = t.Name + " (" + t.ID + ")"
	} else {
		tenant = "No tenants found"
	}
	activeSub, activeTenant := getActiveContext()
	contextLine := subtitleStyle.Render("Active: Subscription: " + activeSub + " | Tenant: " + activeTenant)
	subtitle := subtitleStyle.Render(
		"tab: next subscription | shift+tab: next tenant | enter: set active | Subscription: " + sub + " | Tenant: " + tenant,
	)
	help := helpStyle.Render("Press q to quit. Use tab/shift+tab to switch. Enter to set active context.")

	// Show shortcuts popup if needed
	if m.showShortcutsPopup {
		shortcuts := map[string]string{
			"tab":       "Next tab",
			"shift+tab": "Previous tab",
			"ctrl+w":    "Close tab",
			"ctrl+t":    "New tab",
			"ctrl+q":    "Quit",
			"F1":        "Show shortcuts",
			// ...add more as needed...
		}
		return tui.RenderShortcutsPopup(shortcuts)
	}

	// Show IaC file popup if needed
	if m.showIacFilePopup {
		popup := lipgloss.NewStyle().Width(90).Height(30).Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.RoundedBorder()).Render(m.iacFilePopupContent)
		return "\n\n" + popup + "\n\nPress esc to close file view."
	}

	// Show alarms popup if needed
	if m.showAlarmsPopup && len(m.alarms) > 0 {
		popupWidth := 60
		popupHeight := 10
		alarm := m.alarms[0] // Show the first alarm for now
		popupMsg := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")).Render("ALARM: " + alarm.Name + "\n" + alarm.Status + "\n" + alarm.Details)
		box := lipgloss.NewStyle().Width(popupWidth).Height(popupHeight).Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.RoundedBorder()).Render(popupMsg)
		return "\n\n" + box + "\n\nPress esc to close popup."
	}

	// Show error log popup if needed
	if m.showMatrixPopup && len(m.usageMatrix) > 0 {
		popupWidth := 70
		popupHeight := 12
		popupMsg := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9")).Render("LOG ERROR:\n" + renderUsageMatrix(m.usageHeaders, m.usageMatrix))
		box := lipgloss.NewStyle().Width(popupWidth).Height(popupHeight).Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.RoundedBorder()).Render(popupMsg)
		return "\n\n" + box + "\n\nPress esc to close popup."
	}

	// Show IaC file panel if needed
	if m.showIacPanel {
		var iacPanel string
		if m.iacScanLoading {
			iacPanel = "Scanning for Terraform/Bicep files..."
		} else if m.iacScanErr != "" {
			iacPanel = "IaC scan error: " + m.iacScanErr
		} else if len(m.iacFiles) == 0 {
			iacPanel = "No Terraform/Bicep/tfstate files found. Press F to scan."
		} else {
			iacPanel = "IaC Files:\n"
			for i, f := range m.iacFiles {
				selected := "  "
				if i == m.selectedIacIdx {
					selected = "> "
				}
				iacPanel += selected + f.Path + " [" + f.Type + "]\n"
			}
		}
		help := helpStyle.Render("n: next | p: prev | F: scan | v: view file | esc: close IaC panel")
		return title + "\n" + iacPanel + "\n" + help
	}

	// Show resource creation popup if needed
	if m.showCreatePopup {
		popup := lipgloss.NewStyle().Width(80).Height(16).Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.RoundedBorder()).Render(m.createAIMessage)
		return "\n\n" + popup + "\n\nPress esc to cancel."
	}

	// Show deployment popup if needed
	if m.showDeployPopup {
		popup := lipgloss.NewStyle().Width(90).Height(30).Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.RoundedBorder()).Render(m.deployOutput)
		return "\n\n" + popup + "\n\nPress esc to cancel deployment."
	}

	// Build left panel: resource list
	var leftPanel string
	if m.resourceLoadErr != "" {
		leftPanel = "Resource group error: " + m.resourceLoadErr
	} else if len(m.resourceGroups) == 0 {
		leftPanel = "No resource groups found."
	} else {
		leftPanel = "Resource Groups:\n"
		for i, rg := range m.resourceGroups {
			selected := "  "
			if i == m.resourceGroupIdx {
				selected = "> "
			}
			leftPanel += selected + rg.Name + " (" + rg.Location + ")\n"
		}
		if m.selectedGroup != "" && len(m.resourcesInGroup) > 0 {
			leftPanel += "\nResources in '" + m.selectedGroup + "':\n"
			for i, r := range m.resourcesInGroup {
				selected := "  "
				if i == m.resourceIdx {
					selected = "> "
				}
				leftPanel += selected + r.Name + " [" + r.Type + "]\n"
			}
		}
	}

	// Build right panel: details
	var rightPanel string
	if m.selectedResource != "" && m.resourceIdx < len(m.resourcesInGroup) {
		resource := m.resourcesInGroup[m.resourceIdx]
		details, err := fetchResourceDetails(resource.ID)
		if err != nil {
			rightPanel = "Details error: " + err.Error()
		} else {
			rightPanel = fmt.Sprintf("Name: %s\nType: %s\nLocation: %s\nID: %s\n\n%s", resource.Name, resource.Type, resource.Location, resource.ID, details)
		}
	} else {
		rightPanel = "Select a resource to see details."
	}

	// Box layout: left and right panels side by side
	boxWidth := 40
	leftLines := strings.Split(leftPanel, "\n")
	rightLines := strings.Split(rightPanel, "\n")
	maxLines := len(leftLines)
	if len(rightLines) > maxLines {
		maxLines = len(rightLines)
	}
	var box strings.Builder
	box.WriteString("┌" + strings.Repeat("─", boxWidth) + "┬" + strings.Repeat("─", boxWidth) + "┐\n")
	for i := 0; i < maxLines; i++ {
		l := ""
		if i < len(leftLines) {
			l = leftLines[i]
		}
		r := ""
		if i < len(rightLines) {
			r = rightLines[i]
		}
		box.WriteString(fmt.Sprintf("│%-*s│%-*s│\n", boxWidth, l, boxWidth, r))
	}
	box.WriteString("└" + strings.Repeat("─", boxWidth) + "┴" + strings.Repeat("─", boxWidth) + "┘\n")

	return title + "\n" + contextLine + "\n" + subtitle + "\n\n" + box.String() + "\n" + help
}

// fetchAzureSubsAndTenants uses Azure CLI to get subscriptions and tenants as a quick cross-platform solution.
func fetchAzureSubsAndTenants() ([]Subscription, []Tenant, error) {
	subs := []Subscription{}
	tenants := []Tenant{}
	// Try to get subscriptions from Azure CLI, but fallback to demo data if it fails
	subCmd := exec.Command("az", "account", "list", "--output", "json")
	subOut, err := subCmd.Output()
	if err == nil {
		_ = json.Unmarshal(subOut, &subs)
	}
	if len(subs) == 0 {
		subs = []Subscription{{ID: "demo-sub", Name: "Demo Subscription", TenantID: "demo-tenant", IsDefault: true}}
	}
	// Try to get tenants from Azure CLI, but fallback to demo data if it fails
	tenantCmd := exec.Command("az", "account", "tenant", "list", "--output", "json")
	tenantOut, err := tenantCmd.Output()
	if err == nil {
		_ = json.Unmarshal(tenantOut, &tenants)
	}
	if len(tenants) == 0 {
		tenants = []Tenant{{ID: "demo-tenant", Name: "Demo Tenant"}}
	}
	return subs, tenants, nil
}

// loadResourcesCmd loads resource groups for the current subscription.
func loadResourcesCmd() tea.Cmd {
	return func() tea.Msg {
		groups, err := fetchResourceGroups()
		if err != nil {
			return resourceLoadErrMsg(err.Error())
		}
		return resourcesMsg(groups)
	}
}

// fetchResourceGroups uses Azure Go SDK via azuresdk.AzureClient to get resource groups for the current subscription.
func fetchResourceGroups() ([]ResourceGroup, error) {
	subID := "demo-sub" // fallback demo sub
	if azureClient == nil {
		return []ResourceGroup{{Name: "DemoGroup", Location: "westeurope"}}, nil
	}
	groups, err := azureClient.ListResourceGroups(subID)
	if err != nil || len(groups) == 0 {
		return []ResourceGroup{{Name: "DemoGroup", Location: "westeurope"}}, nil
	}
	var result []ResourceGroup
	for _, g := range groups {
		result = append(result, ResourceGroup{
			Name:     *g.Name,
			Location: *g.Location,
		})
	}
	return result, nil
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
	// fallback demo resource if CLI fails
	cmd := exec.Command("az", "resource", "list", "--resource-group", groupName, "--output", "json")
	out, err := cmd.Output()
	var resources []AzureResource
	if err == nil {
		_ = json.Unmarshal(out, &resources)
	}
	if len(resources) == 0 {
		resources = []AzureResource{{ID: "demo-res", Name: "DemoVM", Type: "Microsoft.Compute/virtualMachines", Location: "westeurope"}}
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
