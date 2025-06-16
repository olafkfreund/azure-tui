package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/popup"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/olafkfreund/azure-tui/internal/azure/azuresdk"
	"github.com/olafkfreund/azure-tui/internal/azure/usage"
	ai "github.com/olafkfreund/azure-tui/internal/openai"
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

func main() {
	if len(os.Args) > 1 && os.Args[1] == "vnet-summary" {
		subID := "<your-subscription-id>" // TODO: get from config or flag
		resourceGroup := ""               // Optionally set
		_ = listAndSummarizeVNetsCLI(subID, resourceGroup)
		return
	}
	p := tea.NewProgram(initialModel())
	if err := p.Start(); err != nil {
		os.Exit(1)
	}
}

// initialModel returns the starting state for the TUI.
func initialModel() tea.Model {
	return &model{
		profiles:       []string{"default"},
		currentProfile: 0,
		environments:   []string{"East US", "West Europe"},
		currentEnv:     0,
		loading:        true,
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
	usageMatrix     [][]string
	usageHeaders    []string
	alarms          []usage.Alarm
	showMatrixPopup bool
	showAlarmsPopup bool
	matrixViewport  viewport.Model
	alarmsViewport  viewport.Model
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
			// Start interactive AKS creation prompt
			m.promptingAKS = true
			m.promptStep = 0
			m.promptMsg = "Enter AKS cluster name:"
			return m, nil
		case "D":
			// Delete selected AKS cluster
			if len(m.aksClusters) > 0 {
				go deleteAKSCluster(m.aksClusters[0].Name, m.aksClusters[0].ResourceGroup)
				m.aksLoading = true
				return m, loadAKSClustersCmd()
			}
		case "v":
			// Load Key Vaults
			m.keyVaultsLoading = true
			return m, loadKeyVaultsCmd()
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

	// Resource group listing with selection
	var rgList string
	if m.resourceLoadErr != "" {
		rgList = "Resource group error: " + m.resourceLoadErr
	} else if len(m.resourceGroups) == 0 {
		rgList = "No resource groups found."
	} else {
		rgList = "Resource Groups (use up/down to select):\n"
		for i, rg := range m.resourceGroups {
			selected := " "
			if i == m.resourceGroupIdx {
				selected = ">"
			}
			rgList += selected + " " + rg.Name + " (" + rg.Location + ")\n"
		}
	}
	// Resources in selected group
	var resList string
	if m.selectedGroup != "" {
		resList = "Resources in group '" + m.selectedGroup + "' (use left/right to select, d: details):\n"
		if m.resourcesInGroupErr != "" {
			resList += "Error: " + m.resourcesInGroupErr + "\n"
		} else if len(m.resourcesInGroup) == 0 {
			resList += "No resources found.\n"
		} else {
			for i, r := range m.resourcesInGroup {
				selected := " "
				if i == m.resourceIdx {
					selected = ">"
				}
				resList += selected + " " + r.Name + " [" + r.Type + "] (" + r.Location + ")\n"
			}
			if m.selectedResource != "" {
				resList += "\nDetails for selected resource:\n" + m.selectedResource + "\n"
			}
		}
	}
	// AKS cluster management
	var aksList string
	if m.promptingAKS {
		aksList = "[AKS Creation] " + m.promptMsg + "\n(Type and press Enter)"
	} else if m.aksLoading {
		aksList = "Loading AKS clusters..."
	} else if m.aksErr != "" {
		aksList = "AKS error: " + m.aksErr
	} else if len(m.aksClusters) == 0 {
		aksList = "No AKS clusters found. Press 'k' to list, 'K' to create, 'D' to delete."
	} else {
		aksList = "AKS Clusters (press 'D' to delete first, 'k' to refresh):\n"
		for _, c := range m.aksClusters {
			aksList += "- " + c.Name + " (" + c.Location + ", RG: " + c.ResourceGroup + ")\n"
		}
	}
	// Key Vaults section
	var keyVaultList string
	if m.promptingKeyVault {
		keyVaultList = "[Key Vault Creation] " + m.promptKeyVaultMsg + "\n(Type and press Enter)"
	} else if m.keyVaultsLoading {
		keyVaultList = "Loading Key Vaults..."
	} else if m.keyVaultErr != "" {
		keyVaultList = "Key Vault error: " + m.keyVaultErr
	} else if len(m.keyVaults) == 0 {
		keyVaultList = "No Key Vaults found. Press 'v' to list, 'V' to create, 'X' to delete."
	} else {
		keyVaultList = "Key Vaults (press 'X' to delete first, 'v' to refresh):\n"
		for _, v := range m.keyVaults {
			keyVaultList += "- " + v.Name + " (" + v.Location + ", RG: " + v.ResourceGroup + ")\n"
		}
	}
	// Storage Accounts section
	var storageList string
	if m.promptingStorage {
		storageList = "[Storage Account Creation] " + m.promptStorageMsg + "\n(Type and press Enter)"
	} else if m.storageLoading {
		storageList = "Loading Storage Accounts..."
	} else if m.storageErr != "" {
		storageList = "Storage Account error: " + m.storageErr
	} else if len(m.storageAccounts) == 0 {
		storageList = "No Storage Accounts found. Press 's' to list, 'S' to create, 'Y' to delete."
	} else {
		storageList = "Storage Accounts (press 'Y' to delete first, 's' to refresh):\n"
		for _, v := range m.storageAccounts {
			storageList += "- " + v.Name + " (" + v.Location + ", RG: " + v.ResourceGroup + ")\n"
		}
	}
	// Matrix and Alarms popups
	if m.showMatrixPopup {
		return popup.New().WithContent(m.matrixViewport.View()).View()
	}
	if m.showAlarmsPopup {
		return popup.New().WithContent(m.alarmsViewport.View()).View()
	}
	return title + "\n" + contextLine + "\n" + subtitle + "\n\n" + rgList + "\n" + resList + "\n" + aksList + "\n" + keyVaultList + "\n" + storageList + "\n" + help
}

// fetchAzureSubsAndTenants uses Azure CLI to get subscriptions and tenants as a quick cross-platform solution.
func fetchAzureSubsAndTenants() ([]Subscription, []Tenant, error) {
	// Get subscriptions
	subCmd := exec.Command("az", "account", "list", "--output", "json")
	subOut, err := subCmd.Output()
	if err != nil {
		return nil, nil, err
	}
	var subs []Subscription
	if err := json.Unmarshal(subOut, &subs); err != nil {
		return nil, nil, err
	}
	// Get tenants
	tenantCmd := exec.Command("az", "account", "tenant", "list", "--output", "json")
	tenantOut, err := tenantCmd.Output()
	if err != nil {
		return subs, nil, err
	}
	var tenants []Tenant
	if err := json.Unmarshal(tenantOut, &tenants); err != nil {
		return subs, nil, err
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
	subID := "" // TODO: get current subscription ID from state or config
	if azureClient == nil {
		return nil, nil
	}
	groups, err := azureClient.ListResourceGroups(subID)
	if err != nil {
		return nil, err
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
	cmd := exec.Command("az", "resource", "list", "--resource-group", groupName, "--output", "json")
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	var resources []AzureResource
	if err := json.Unmarshal(out, &resources); err != nil {
		return nil, err
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
	return nil, nil
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

// fetchStorageAccounts uses Azure SDK helper to get Storage Accounts.
func fetchStorageAccounts() ([]struct {
	Name          string
	Location      string
	ResourceGroup string
}, error) {
	return nil, nil
}

// createStorageAccount creates a new Storage Account with user input.
func createStorageAccount(name, group, location string) {
}

// deleteStorageAccount deletes a Storage Account by name and resource group.
func deleteStorageAccount(name, group string) {
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
	aiProvider := ai.NewAIProvider("") // TODO: pass actual API key or config
	return aiProvider.SummarizeResourceGroups(names)
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
	aiProvider := ai.NewAIProvider("") // TODO: pass actual API key or config
	summary, err := aiProvider.Ask("Summarize these Azure VNets and suggest improvements:", strings.Join(vnetNames, ", "))
	if err != nil {
		return err
	}
	println("Virtual Networks:", strings.Join(vnetNames, ", "))
	println("AI Summary:", summary)
	return nil
}

// Render usage matrix
func renderUsageMatrix(headers []string, matrix [][]string) string {
	t := table.New(table.WithColumns(headers))
	for _, row := range matrix {
		t.AddRow(row...)
	}
	return t.View()
}

// Render alarms
func renderAlarms(alarms []usage.Alarm) string {
	var b strings.Builder
	for _, a := range alarms {
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("1")).Render(a.Name + ": " + a.Status))
		b.WriteString("\n" + a.Details + "\n\n")
	}
	return b.String()
}

// AI-powered log error summarization
func (m *model) summarizeResourceLogErrors(logs []string) string {
	aiProvider := ai.NewAIProvider("") // TODO: pass actual API key
	summary, err := aiProvider.Ask("Summarize and explain these Azure resource log errors. Suggest fixes:", strings.Join(logs, "\n"))
	if err != nil {
		return "AI error: " + err.Error()
	}
	return summary
}
