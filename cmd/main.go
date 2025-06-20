package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/olafkfreund/azure-tui/internal/azure/aci"
	"github.com/olafkfreund/azure-tui/internal/azure/keyvault"
	"github.com/olafkfreund/azure-tui/internal/azure/network"
	"github.com/olafkfreund/azure-tui/internal/azure/resourceactions"
	"github.com/olafkfreund/azure-tui/internal/azure/resourcedetails"
	"github.com/olafkfreund/azure-tui/internal/azure/storage"
	"github.com/olafkfreund/azure-tui/internal/azure/tfbicep"
	"github.com/olafkfreund/azure-tui/internal/config"
	"github.com/olafkfreund/azure-tui/internal/openai"
	"github.com/olafkfreund/azure-tui/internal/search"
	"github.com/olafkfreund/azure-tui/internal/terraform"
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

// Network loading progress message types
type networkLoadingProgressMsg struct {
	progress network.NetworkLoadingProgress
}
type networkDashboardLoadedMsg struct {
	dashboard *network.NetworkDashboard
	content   string
}

// Network topology loading progress message types
type networkTopologyLoadingProgressMsg struct {
	progress network.NetworkLoadingProgress
}
type networkTopologyLoadingProgressWithContinuationMsg struct {
	progress         network.NetworkLoadingProgress
	remainingUpdates []network.NetworkLoadingProgress
	finalTopology    *network.NetworkTopology
	finalError       error
}

// Dashboard loading progress message types
type dashboardLoadingProgressMsg struct {
	progress resourcedetails.DashboardLoadingProgress
}
type dashboardDataLoadedMsg struct {
	data    *resourcedetails.ComprehensiveDashboardData
	content string
}
type dashboardLoadingProgressWithContinuationMsg struct {
	progress         resourcedetails.DashboardLoadingProgress
	remainingUpdates []resourcedetails.DashboardLoadingProgress
	finalData        *resourcedetails.ComprehensiveDashboardData
	finalError       error
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

// Key Vault message types
type keyVaultSecretsMsg struct {
	vaultName string
	secrets   []keyvault.Secret
}
type keyVaultSecretDetailsMsg struct {
	secret *keyvault.Secret
}
type keyVaultSecretActionMsg struct {
	action string
	result resourceactions.ActionResult
}

// Terraform message types
type terraformFoldersLoadedMsg struct {
	folders []string
}
type terraformAnalysisMsg struct {
	analysis string
	path     string
}
type terraformOperationMsg struct {
	operation string
	result    string
	success   bool
}

// Settings message types
type settingsFoldersLoadedMsg struct {
	folders []string
}
type settingsConfigLoadedMsg struct {
	config  *config.AppConfig
	content string
}
type settingsConfigSavedMsg struct {
	success bool
	message string
}

// Storage Account message types
type storageContainersMsg struct {
	accountName string
	containers  []storage.Container
}
type storageBlobsMsg struct {
	accountName   string
	containerName string
	blobs         []storage.Blob
}
type storageBlobDetailsMsg struct {
	blob *storage.Blob
}
type storageActionMsg struct {
	action string
	result resourceactions.ActionResult
}

// Storage progress message types
type storageLoadingStartMsg struct {
	operation   string
	accountName string
}

type storageLoadingProgressMsg struct {
	progress storage.StorageLoadingProgress
}

type storageLoadingCompleteMsg struct {
	operation string
	success   bool
	data      interface{}
	error     error
}

// Subscription selection message types
type currentSubscriptionMsg struct {
	subscription *Subscription
}
type subscriptionMenuMsg struct {
	subscriptions []Subscription
}
type subscriptionSelectedMsg struct {
	subscription Subscription
	success      bool
	message      string
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

	// Network loading progress tracking
	networkLoadingInProgress bool
	networkLoadingStartTime  time.Time

	// Network topology loading progress tracking
	topologyLoadingInProgress bool
	topologyLoadingStartTime  time.Time

	// Dashboard loading progress tracking
	dashboardLoadingInProgress bool
	dashboardLoadingStartTime  time.Time
	dashboardData              *resourcedetails.ComprehensiveDashboardData

	// Container Instance-specific content
	containerInstanceDetailsContent string
	containerInstanceLogsContent    string

	// Key Vault-specific content
	keyVaultSecretsContent       string
	keyVaultSecretDetailsContent string
	keyVaultSecrets              []keyvault.Secret
	selectedSecret               *keyvault.Secret

	// Storage Account-specific content
	storageContainersContent  string
	storageBlobsContent       string
	storageBlobDetailsContent string
	storageContainers         []storage.Container
	storageBlobs              []storage.Blob
	selectedBlob              *storage.Blob
	currentStorageAccount     string
	currentContainer          string

	// Help popup state
	showHelpPopup    bool
	helpScrollOffset int // For scrolling through help content

	// Navigation stack for back navigation
	navigationStack []string

	// Search functionality
	searchEngine      *search.SearchEngine
	searchMode        bool
	searchQuery       string
	searchResults     []search.SearchResult
	searchResultIndex int
	searchSuggestions []string
	showSearchResults bool
	searchHistory     []string
	filteredResources []AzureResource

	// Terraform integration
	showTerraformPopup    bool
	terraformMenuIndex    int
	terraformMode         string // "menu", "folder-select", "templates", "workspaces", "operations"
	terraformFolderPath   string
	terraformMenuOptions  []string
	terraformFolders      []string
	terraformAnalysis     string
	terraformMenuAction   string // Track the original menu action for folder selection
	terraformScrollOffset int    // For scrolling through long analysis text

	// Settings menu functionality
	showSettingsPopup     bool
	settingsMenuIndex     int
	settingsMode          string // "menu", "config-view", "folder-browser", "edit-setting"
	settingsCurrentPath   string
	settingsFolders       []string
	settingsConfigContent string
	settingsEditKey       string
	settingsEditValue     string
	settingsCurrentConfig *config.AppConfig

	// Subscription selection functionality
	currentSubscription    *Subscription
	showSubscriptionPopup  bool
	subscriptionMenuIndex  int
	availableSubscriptions []Subscription
	subscriptionMenuMode   string // "menu" or "loading"
}

// Helper functions for search functionality

// convertAzureResourceToSearchResource converts AzureResource to search.Resource
func convertAzureResourceToSearchResource(azResource AzureResource) search.Resource {
	return search.Resource{
		ID:            azResource.ID,
		Name:          azResource.Name,
		Type:          azResource.Type,
		Location:      azResource.Location,
		ResourceGroup: azResource.ResourceGroup,
		Status:        azResource.Status,
		Tags:          azResource.Tags,
		Properties:    azResource.Properties,
	}
}

// updateSearchEngine updates the search engine with current resources
func (m *model) updateSearchEngine() {
	searchResources := make([]search.Resource, len(m.allResources))
	for i, azResource := range m.allResources {
		searchResources[i] = convertAzureResourceToSearchResource(azResource)
	}
	m.searchEngine.SetResources(searchResources)
}

// performSearch executes a search and updates results
func (m *model) performSearch() {
	if m.searchQuery == "" {
		m.searchResults = []search.SearchResult{}
		m.filteredResources = m.allResources
		m.showSearchResults = false
		return
	}

	results, err := m.searchEngine.Search(m.searchQuery)
	if err != nil {
		m.searchResults = []search.SearchResult{}
		m.filteredResources = []AzureResource{}
		return
	}

	m.searchResults = results
	m.searchResultIndex = 0

	// Create filtered resources list from search results
	resourceMap := make(map[string]AzureResource)
	for _, resource := range m.allResources {
		resourceMap[resource.ID] = resource
	}

	m.filteredResources = []AzureResource{}
	seenIDs := make(map[string]bool)
	for _, result := range results {
		if !seenIDs[result.ResourceID] {
			if resource, exists := resourceMap[result.ResourceID]; exists {
				m.filteredResources = append(m.filteredResources, resource)
				seenIDs[result.ResourceID] = true
			}
		}
	}

	m.showSearchResults = len(results) > 0
}

// addToSearchHistory adds a query to search history
func (m *model) addToSearchHistory(query string) {
	if query == "" {
		return
	}

	// Remove duplicates
	for i, h := range m.searchHistory {
		if h == query {
			m.searchHistory = append(m.searchHistory[:i], m.searchHistory[i+1:]...)
			break
		}
	}

	// Add to front
	m.searchHistory = append([]string{query}, m.searchHistory...)

	// Limit history size
	if len(m.searchHistory) > 20 {
		m.searchHistory = m.searchHistory[:20]
	}
}

// navigateSearchResults moves to next/previous search result
func (m *model) navigateSearchResults(direction int) {
	if len(m.searchResults) == 0 {
		return
	}

	m.searchResultIndex += direction
	if m.searchResultIndex < 0 {
		m.searchResultIndex = len(m.searchResults) - 1
	} else if m.searchResultIndex >= len(m.searchResults) {
		m.searchResultIndex = 0
	}

	// Auto-select the resource in tree view
	if m.searchResultIndex < len(m.searchResults) {
		result := m.searchResults[m.searchResultIndex]
		for _, resource := range m.allResources {
			if resource.ID == result.ResourceID {
				// Set the selected resource and load details
				m.selectedResource = &resource
				break
			}
		}
	}
}

// enterSearchMode activates search mode
func (m *model) enterSearchMode() {
	m.searchMode = true
	m.searchQuery = ""
	m.searchResults = []search.SearchResult{}
	m.searchResultIndex = 0
	m.showSearchResults = false
}

// exitSearchMode deactivates search mode and resets filters
func (m *model) exitSearchMode() {
	m.searchMode = false
	m.searchQuery = ""
	m.searchResults = []search.SearchResult{}
	m.searchResultIndex = 0
	m.showSearchResults = false
	m.filteredResources = m.allResources
}

// updateSearchSuggestions updates search suggestions based on current query
func (m *model) updateSearchSuggestions() {
	if len(m.searchQuery) >= 2 {
		m.searchSuggestions = m.searchEngine.GetSuggestions(m.searchQuery)
	} else {
		m.searchSuggestions = []string{}
	}
}

// renderSearchInput renders the search input bar
func (m *model) renderSearchInput(width int) string {
	if !m.searchMode {
		return ""
	}

	searchStyle := lipgloss.NewStyle().
		Foreground(fgLight).
		Background(bgMedium).
		Padding(0, 1).
		Width(width)

	prompt := "🔍 Search: "
	cursor := ""
	if len(m.searchQuery) == 0 {
		cursor = "█"
	}

	content := prompt + m.searchQuery + cursor

	// Show suggestions if available
	if len(m.searchSuggestions) > 0 {
		content += "\n" + lipgloss.NewStyle().Faint(true).Render("Suggestions: "+strings.Join(m.searchSuggestions[:min(3, len(m.searchSuggestions))], ", "))
	}

	// Show search results count
	if m.showSearchResults {
		resultCount := fmt.Sprintf(" (%d results)", len(m.searchResults))
		if len(m.searchResults) > 0 {
			resultCount += fmt.Sprintf(" [%d/%d]", m.searchResultIndex+1, len(m.searchResults))
		}
		content += lipgloss.NewStyle().Foreground(colorAqua).Render(resultCount)
	}

	return searchStyle.Render(content)
}

// renderSearchResults renders the search results view
func (m *model) renderSearchResults(width, height int) string {
	if !m.showSearchResults || len(m.searchResults) == 0 {
		return ""
	}

	var content strings.Builder
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(colorBlue)
	content.WriteString(headerStyle.Render(fmt.Sprintf("🔍 Search Results (%d found)", len(m.searchResults))))
	content.WriteString("\n\n")

	// Group results by resource for better display
	resourceResults := make(map[string][]search.SearchResult)
	for _, result := range m.searchResults {
		resourceResults[result.ResourceID] = append(resourceResults[result.ResourceID], result)
	}

	maxResults := min(height-5, len(resourceResults))
	resultCount := 0

	for resourceID, results := range resourceResults {
		if resultCount >= maxResults {
			break
		}

		// Find the resource for display
		var resource AzureResource
		for _, r := range m.allResources {
			if r.ID == resourceID {
				resource = r
				break
			}
		}

		// Highlight current selection
		isSelected := resultCount == m.searchResultIndex
		nameStyle := lipgloss.NewStyle().Foreground(colorGreen)
		if isSelected {
			nameStyle = nameStyle.Background(bgLight).Bold(true)
		}

		content.WriteString(nameStyle.Render(fmt.Sprintf("📦 %s", resource.Name)))
		content.WriteString(fmt.Sprintf(" (%s)\n", lipgloss.NewStyle().Foreground(colorGray).Render(resource.Type)))

		// Show match details
		for _, result := range results {
			matchStyle := lipgloss.NewStyle().Foreground(colorYellow).Faint(true)
			content.WriteString(fmt.Sprintf("   %s: %s\n", matchStyle.Render(result.MatchType), result.MatchValue))
		}

		content.WriteString("\n")
		resultCount++
	}

	if len(resourceResults) > maxResults {
		moreStyle := lipgloss.NewStyle().Faint(true).Foreground(colorGray)
		content.WriteString(moreStyle.Render(fmt.Sprintf("... and %d more results", len(resourceResults)-maxResults)))
	}

	return content.String()
}

// handleTerraformMenuSelection handles the selection from the Terraform menu
func (m *model) handleTerraformMenuSelection() (tea.Model, tea.Cmd) {
	switch m.terraformMode {
	case "menu":
		switch m.terraformMenuIndex {
		case 0: // Browse Folders
			m.terraformMode = "folder-select"
			m.terraformMenuIndex = 0
			m.terraformMenuAction = "browse"
			// If folders aren't loaded yet, load them
			if len(m.terraformFolders) == 0 {
				return *m, loadTerraformFoldersCmd()
			}
		case 1: // Create from Template
			m.terraformMode = "folder-select"
			m.terraformMenuIndex = 0
			m.terraformMenuAction = "template"
			// If folders aren't loaded yet, load them
			if len(m.terraformFolders) == 0 {
				return *m, loadTerraformFoldersCmd()
			}
		case 2: // Analyze Code
			m.terraformMode = "folder-select"
			m.terraformMenuIndex = 0
			m.terraformMenuAction = "analyze"
			// If folders aren't loaded yet, load them
			if len(m.terraformFolders) == 0 {
				return *m, loadTerraformFoldersCmd()
			}
		case 3: // Terraform Operations
			m.terraformMode = "folder-select"
			m.terraformMenuIndex = 0
			m.terraformMenuAction = "operations"
			// If folders aren't loaded yet, load them
			if len(m.terraformFolders) == 0 {
				return *m, loadTerraformFoldersCmd()
			}
		case 4: // Open External Editor
			m.terraformMode = "folder-select"
			m.terraformMenuIndex = 0
			m.terraformMenuAction = "editor"
			// If folders aren't loaded yet, load them
			if len(m.terraformFolders) == 0 {
				return *m, loadTerraformFoldersCmd()
			}
		}
	case "folder-select":
		if m.terraformMenuIndex < len(m.terraformFolders) {
			selectedFolder := m.terraformFolders[m.terraformMenuIndex]
			m.showTerraformPopup = false

			// Perform different actions based on original menu selection
			switch m.terraformMenuAction {
			case "analyze":
				return *m, analyzeTerraformCodeCmd(selectedFolder)
			case "operations":
				// Show operations submenu or execute default operation
				return *m, executeTerraformOperationCmd("validate", selectedFolder)
			case "editor":
				// Open external editor (e.g., code, vim)
				return *m, openTerraformEditorCmd(selectedFolder)
			case "template":
				// Implement template creation workflow
				return *m, createFromTemplateCmd(selectedFolder)
			default:
				return *m, analyzeTerraformCodeCmd(selectedFolder)
			}
		}
	}
	return *m, nil
}

// handleTerraformDeploymentSelection handles the selection from the deployment menu
func (m *model) handleTerraformDeploymentSelection() (tea.Model, tea.Cmd) {
	switch m.terraformMenuIndex {
	case 0: // Initialize Terraform (terraform init)
		m.showTerraformPopup = false
		return *m, executeTerraformOperationCmd("init", m.terraformFolderPath)
	case 1: // Plan Deployment (terraform plan)
		m.showTerraformPopup = false
		return *m, executeTerraformOperationCmd("plan", m.terraformFolderPath)
	case 2: // Deploy Infrastructure (terraform apply)
		m.showTerraformPopup = false
		return *m, executeTerraformOperationCmd("apply", m.terraformFolderPath)
	case 3: // Edit Template Files
		m.showTerraformPopup = false
		return *m, analyzeTerraformCodeCmd(m.terraformFolderPath)
	case 4: // Open in External Editor
		m.showTerraformPopup = false
		return *m, openTerraformEditorCmd(m.terraformFolderPath)
	case 5: // Return to Main Menu
		m.terraformMode = "menu"
		m.terraformMenuIndex = 0
		return *m, nil
	}
	return *m, nil
}

// handleSettingsMenuSelection handles the selection from the Settings menu
func (m *model) handleSettingsMenuSelection() (tea.Model, tea.Cmd) {
	switch m.settingsMode {
	case "menu":
		switch m.settingsMenuIndex {
		case 0: // View Configuration
			m.settingsMode = "config-view"
			return *m, nil
		case 1: // Edit Terraform Directory
			m.settingsMode = "folder-browser"
			m.settingsMenuIndex = 0
			return *m, loadSettingsFoldersCmd()
		case 2: // Edit UI Settings
			if m.settingsCurrentConfig != nil {
				m.settingsMode = "edit-setting"
				m.settingsEditKey = "UI Settings"
				m.settingsEditValue = fmt.Sprintf("PopupWidth: %d, PopupHeight: %d",
					m.settingsCurrentConfig.UI.PopupWidth,
					m.settingsCurrentConfig.UI.PopupHeight)
			}
			return *m, nil
		case 3: // Edit Editor Settings
			if m.settingsCurrentConfig != nil {
				m.settingsMode = "edit-setting"
				m.settingsEditKey = "Editor Settings"
				m.settingsEditValue = m.settingsCurrentConfig.Editor.DefaultEditor
			}
			return *m, nil
		case 4: // Save Configuration
			if m.settingsCurrentConfig != nil {
				m.showSettingsPopup = false
				return *m, saveSettingsConfigCmd(m.settingsCurrentConfig)
			}
		}
	case "folder-browser":
		if m.settingsMenuIndex < len(m.settingsFolders) {
			selectedPath := m.settingsFolders[m.settingsMenuIndex]
			m.settingsCurrentPath = selectedPath
			if m.settingsCurrentConfig != nil {
				m.settingsCurrentConfig.Terraform.WorkspacePath = selectedPath
			}
			m.settingsMode = "menu"
			m.settingsMenuIndex = 0
			return *m, nil
		}
	}
	return *m, nil
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
			return aiDescriptionLoadedMsg{description: "AI provider not configured. Set GITHUB_TOKEN or OPENAI_API_KEY environment variable."}
		}

		detailsStr := fmt.Sprintf("Resource: %s\nType: %s\nLocation: %s\nStatus: %s",
			resource.Name, resource.Type, resource.Location, resource.Status)

		if details != nil {
			detailsStr += fmt.Sprintf("\nProperties: %v", details.Properties)
		}

		description, err := ai.DescribeResource(resource.Type, resource.Name, detailsStr)
		if err != nil {
			errorMsg := err.Error()
			// Check for common API errors and provide helpful messages
			if strings.Contains(errorMsg, "insufficient_quota") {
				if ai.ProviderType == "github_copilot" {
					return aiDescriptionLoadedMsg{description: "❌ GitHub Copilot quota exceeded. Using fallback to OpenAI."}
				} else {
					return aiDescriptionLoadedMsg{description: "❌ AI quota exceeded. Please check your billing details or try GitHub Copilot."}
				}
			} else if strings.Contains(errorMsg, "invalid_api_key") || strings.Contains(errorMsg, "401") {
				return aiDescriptionLoadedMsg{description: "❌ Invalid API key. Please check your GITHUB_TOKEN or OPENAI_API_KEY environment variable."}
			} else if strings.Contains(errorMsg, "rate_limit") {
				return aiDescriptionLoadedMsg{description: "❌ AI rate limit exceeded. Please try again in a moment."}
			} else if strings.Contains(errorMsg, "403") || strings.Contains(errorMsg, "forbidden") {
				if ai.ProviderType == "github_copilot" {
					return aiDescriptionLoadedMsg{description: "❌ GitHub Copilot access forbidden. Check your subscription or use OPENAI_API_KEY instead."}
				} else {
					return aiDescriptionLoadedMsg{description: "❌ API access forbidden. Please check your credentials."}
				}
			} else {
				providerInfo := fmt.Sprintf(" (Provider: %s)", ai.ProviderType)
				return aiDescriptionLoadedMsg{description: fmt.Sprintf("❌ AI analysis failed: %v%s", err, providerInfo)}
			}
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

// showNetworkDashboardCmd displays comprehensive network dashboard with progress
func showNetworkDashboardCmd() tea.Cmd {
	return func() tea.Msg {
		// Return initial progress message immediately
		return networkLoadingProgressMsg{progress: network.NetworkLoadingProgress{
			CurrentOperation:       "Initializing network dashboard...",
			TotalOperations:        7,
			CompletedOperations:    0,
			ProgressPercentage:     0.0,
			ResourceProgress:       make(map[string]network.ResourceProgress),
			Errors:                 []string{},
			StartTime:              time.Now(),
			EstimatedTimeRemaining: "Calculating...",
		}}
	}
}

// loadNetworkDashboardWithProgressCmd loads the network dashboard with real-time progress updates
func loadNetworkDashboardWithProgressCmd() tea.Cmd {
	return func() tea.Msg {
		// Start async loading with real-time progress
		return startNetworkLoadingCmd()
	}
}

// startNetworkLoadingCmd starts the async network loading process
func startNetworkLoadingCmd() tea.Msg {
	// Create a command that will load the dashboard async and send progress updates
	return tea.Batch(
		// Start the actual loading process
		loadNetworkDashboardAsyncWithProgressCmd(),
		// Start a ticker for smooth progress animation
		tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return progressTickMsg{}
		}),
	)()
}

// loadNetworkDashboardAsyncWithProgressCmd loads dashboard with streaming progress
func loadNetworkDashboardAsyncWithProgressCmd() tea.Cmd {
	return func() tea.Msg {
		// This will take time, so we'll simulate progress
		// In a real implementation, this would stream progress updates
		dashboard, err := network.GetNetworkDashboardWithProgress("", nil)

		// Return final result
		var dashboardContent string
		if err != nil && dashboard == nil {
			dashboardContent = fmt.Sprintf("Error loading network dashboard: %v", err)
		} else {
			dashboardContent = network.RenderNetworkDashboard()
		}

		return networkDashboardMsg{content: dashboardContent}
	}
}

// progressTickMsg is sent by the progress ticker
type progressTickMsg struct{}

// New message type for progress with continuation
type networkLoadingProgressWithContinuationMsg struct {
	progress         network.NetworkLoadingProgress
	remainingUpdates []network.NetworkLoadingProgress
	finalDashboard   *network.NetworkDashboard
	finalError       error
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

// showNetworkTopologyCmd displays network topology view with progress
func showNetworkTopologyCmd() tea.Cmd {
	return func() tea.Msg {
		// Return initial progress message immediately
		return networkTopologyLoadingProgressMsg{progress: network.NetworkLoadingProgress{
			CurrentOperation:       "Initializing network topology loading...",
			TotalOperations:        7,
			CompletedOperations:    0,
			ProgressPercentage:     0.0,
			ResourceProgress:       make(map[string]network.ResourceProgress),
			Errors:                 []string{},
			StartTime:              time.Now(),
			EstimatedTimeRemaining: "Calculating...",
		}}
	}
}

// loadNetworkTopologyWithProgressCmd loads the network topology with real-time progress updates
func loadNetworkTopologyWithProgressCmd() tea.Cmd {
	return func() tea.Msg {
		// Start async loading with real-time progress
		return startNetworkTopologyLoadingCmd()
	}
}

// startNetworkTopologyLoadingCmd starts the async network topology loading process
func startNetworkTopologyLoadingCmd() tea.Msg {
	// Create a command that will load the topology async and send progress updates
	return tea.Batch(
		// Start the actual loading process
		loadNetworkTopologyAsyncWithProgressCmd(),
		// Start a ticker for smooth progress animation
		tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return progressTickMsg{}
		}),
	)()
}

// loadNetworkTopologyAsyncWithProgressCmd loads topology with streaming progress
func loadNetworkTopologyAsyncWithProgressCmd() tea.Cmd {
	return func() tea.Msg {
		// This will take time, so we'll simulate progress
		// In a real implementation, this would stream progress updates
		topologyContent, err := network.GetNetworkTopologyWithProgress("", nil)

		// Return final result
		var finalContent string
		if err != nil {
			finalContent = fmt.Sprintf("Error loading network topology: %v", err)
		} else {
			finalContent = topologyContent
		}

		return networkTopologyMsg{content: finalContent}
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
// RESOURCE DASHBOARD COMMANDS
// =============================================================================

// showEnhancedDashboardCmd displays enhanced dashboard with progress
func showEnhancedDashboardCmd(resourceID string) tea.Cmd {
	return func() tea.Msg {
		// Add safety check to prevent crashes
		if resourceID == "" {
			return errorMsg{error: "Cannot load dashboard: resource ID is empty"}
		}

		// Return initial progress message immediately
		return dashboardLoadingProgressMsg{progress: resourcedetails.DashboardLoadingProgress{
			CurrentOperation:       "Initializing resource dashboard...",
			TotalOperations:        5,
			CompletedOperations:    0,
			ProgressPercentage:     0.0,
			DataProgress:           make(map[string]resourcedetails.DataProgress),
			Errors:                 []string{},
			StartTime:              time.Now(),
			EstimatedTimeRemaining: "Calculating...",
		}}
	}
}

// loadDashboardWithProgressCmd loads the dashboard with real-time progress updates
func loadDashboardWithProgressCmd(resourceID string) tea.Cmd {
	return func() tea.Msg {
		// Start async loading with real-time progress
		return startDashboardLoadingCmd(resourceID)
	}
}

// startDashboardLoadingCmd starts the async dashboard loading process
func startDashboardLoadingCmd(resourceID string) tea.Msg {
	// Create a command that will load the dashboard async and send progress updates
	return tea.Batch(
		// Start the actual loading process
		loadDashboardAsyncWithProgressCmd(resourceID),
		// Start a ticker for smooth progress animation
		tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
			return progressTickMsg{}
		}),
	)()
}

// loadDashboardAsyncWithProgressCmd loads dashboard with streaming progress
func loadDashboardAsyncWithProgressCmd(resourceID string) tea.Cmd {
	return func() tea.Msg {
		// Add safety check
		if resourceID == "" {
			return errorMsg{error: "Cannot load dashboard: resource ID is empty"}
		}

		// This will take time, so we'll simulate progress
		// In a real implementation, this would stream progress updates
		dashboardData, err := resourcedetails.GetComprehensiveDashboardDataWithProgress(resourceID, nil)

		// Return final result with better error handling
		var dashboardContent string
		if err != nil && dashboardData == nil {
			dashboardContent = fmt.Sprintf("Error loading dashboard: %v", err)
			return dashboardDataLoadedMsg{data: nil, content: dashboardContent}
		} else if dashboardData == nil {
			dashboardContent = "Dashboard data is unavailable"
			return dashboardDataLoadedMsg{data: nil, content: dashboardContent}
		} else {
			// Extract resource name from ID if possible, or use ID as fallback
			resourceName := resourceID
			if parts := strings.Split(resourceID, "/"); len(parts) > 0 {
				resourceName = parts[len(parts)-1]
			}
			dashboardContent = tui.RenderComprehensiveDashboard(resourceName, dashboardData)
		}

		return dashboardDataLoadedMsg{data: dashboardData, content: dashboardContent}
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

// =============================================================================
// KEY VAULT SECRET MANAGEMENT COMMANDS
// =============================================================================

// listKeyVaultSecretsCmd lists all secrets in a Key Vault
func listKeyVaultSecretsCmd(vaultName string) tea.Cmd {
	return func() tea.Msg {
		secrets, err := keyvault.ListSecrets(vaultName)
		if err != nil {
			return errorMsg{error: fmt.Sprintf("Failed to list secrets: %v", err)}
		}
		return keyVaultSecretsMsg{vaultName: vaultName, secrets: secrets}
	}
}

// showKeyVaultSecretDetailsCmd shows detailed information about a specific secret
func showKeyVaultSecretDetailsCmd(vaultName, secretName string) tea.Cmd {
	return func() tea.Msg {
		secret, err := keyvault.GetSecretMetadata(vaultName, secretName)
		if err != nil {
			return errorMsg{error: fmt.Sprintf("Failed to get secret details: %v", err)}
		}
		return keyVaultSecretDetailsMsg{secret: secret}
	}
}

// createKeyVaultSecretCmd creates a new secret in a Key Vault
func createKeyVaultSecretCmd(vaultName, secretName, secretValue string, tags map[string]string) tea.Cmd {
	return func() tea.Msg {
		err := keyvault.CreateSecret(vaultName, secretName, secretValue, tags)
		var result resourceactions.ActionResult

		if err != nil {
			result = resourceactions.ActionResult{
				Success: false,
				Message: fmt.Sprintf("Failed to create secret '%s': %v", secretName, err),
				Output:  "",
			}
		} else {
			result = resourceactions.ActionResult{
				Success: true,
				Message: fmt.Sprintf("Successfully created secret '%s' in Key Vault '%s'", secretName, vaultName),
				Output:  fmt.Sprintf("Secret '%s' is now available in Key Vault", secretName),
			}
		}

		return keyVaultSecretActionMsg{action: "create", result: result}
	}
}

// deleteKeyVaultSecretCmd deletes a secret from a Key Vault
func deleteKeyVaultSecretCmd(vaultName, secretName string) tea.Cmd {
	return func() tea.Msg {
		err := keyvault.DeleteSecret(vaultName, secretName)
		var result resourceactions.ActionResult

		if err != nil {
			result = resourceactions.ActionResult{
				Success: false,
				Message: fmt.Sprintf("Failed to delete secret '%s': %v", secretName, err),
				Output:  "",
			}
		} else {
			result = resourceactions.ActionResult{
				Success: true,
				Message: fmt.Sprintf("Successfully deleted secret '%s' from Key Vault '%s'", secretName, vaultName),
				Output:  fmt.Sprintf("Secret '%s' has been removed from Key Vault", secretName),
			}
		}

		return keyVaultSecretActionMsg{action: "delete", result: result}
	}
}

// =============================================================================
// STORAGE ACCOUNT MANAGEMENT COMMANDS
// =============================================================================

// listStorageContainersCmd lists all containers in a storage account with progress
func listStorageContainersCmd(accountName string) tea.Cmd {
	return func() tea.Msg {
		// Start with progress tracking
		return storageLoadingStartMsg{
			operation:   "containers",
			accountName: accountName,
		}
	}
}

// listStorageContainersWithProgressCmd performs the actual container listing with progress
func listStorageContainersWithProgressCmd(accountName string) tea.Cmd {
	return func() tea.Msg {
		containers, err := storage.ListContainersWithProgress(accountName, nil)
		if err != nil {
			return storageLoadingCompleteMsg{
				operation: "containers",
				success:   false,
				error:     err,
			}
		}
		return storageLoadingCompleteMsg{
			operation: "containers",
			success:   true,
			data:      containers,
		}
	}
}

// listStorageBlobsCmd lists all blobs in a container with progress
func listStorageBlobsCmd(accountName, containerName string) tea.Cmd {
	return func() tea.Msg {
		// Start with progress tracking
		return storageLoadingStartMsg{
			operation:   "blobs",
			accountName: accountName,
		}
	}
}

// listStorageBlobsWithProgressCmd performs the actual blob listing with progress
func listStorageBlobsWithProgressCmd(accountName, containerName string) tea.Cmd {
	return func() tea.Msg {
		blobs, err := storage.ListBlobsWithProgress(accountName, containerName, nil)
		if err != nil {
			return storageLoadingCompleteMsg{
				operation: "blobs",
				success:   false,
				error:     err,
			}
		}
		return storageLoadingCompleteMsg{
			operation: "blobs",
			success:   true,
			data:      blobs,
		}
	}
}

// showBlobDetailsCmd shows detailed information about a specific blob
func showBlobDetailsCmd(accountName, containerName, blobName string) tea.Cmd {
	return func() tea.Msg {
		blob, err := storage.GetBlobProperties(accountName, containerName, blobName)
		if err != nil {
			return errorMsg{error: fmt.Sprintf("Failed to get blob details: %v", err)}
		}
		return storageBlobDetailsMsg{blob: blob}
	}
}

// createStorageContainerCmd creates a new container in a storage account
func createStorageContainerCmd(accountName, containerName string) tea.Cmd {
	return func() tea.Msg {
		err := storage.CreateContainer(accountName, containerName)
		var result resourceactions.ActionResult

		if err != nil {
			result = resourceactions.ActionResult{
				Success: false,
				Message: fmt.Sprintf("Failed to create container '%s': %v", containerName, err),
				Output:  "",
			}
		} else {
			result = resourceactions.ActionResult{
				Success: true,
				Message: fmt.Sprintf("Successfully created container '%s' in storage account '%s'", containerName, accountName),
				Output:  fmt.Sprintf("Container '%s' is now available in storage account", containerName),
			}
		}

		return storageActionMsg{action: "create-container", result: result}
	}
}

// deleteStorageContainerCmd deletes a container from a storage account
func deleteStorageContainerCmd(accountName, containerName string) tea.Cmd {
	return func() tea.Msg {
		err := storage.DeleteContainer(accountName, containerName)
		var result resourceactions.ActionResult

		if err != nil {
			result = resourceactions.ActionResult{
				Success: false,
				Message: fmt.Sprintf("Failed to delete container '%s': %v", containerName, err),
				Output:  "",
			}
		} else {
			result = resourceactions.ActionResult{
				Success: true,
				Message: fmt.Sprintf("Successfully deleted container '%s' from storage account '%s'", containerName, accountName),
				Output:  fmt.Sprintf("Container '%s' has been removed from storage account", containerName),
			}
		}

		return storageActionMsg{action: "delete-container", result: result}
	}
}

// uploadBlobCmd uploads a file to a blob container
func uploadBlobCmd(accountName, containerName, blobName, filePath string) tea.Cmd {
	return func() tea.Msg {
		err := storage.UploadBlob(accountName, containerName, blobName, filePath)
		var result resourceactions.ActionResult

		if err != nil {
			result = resourceactions.ActionResult{
				Success: false,
				Message: fmt.Sprintf("Failed to upload blob '%s': %v", blobName, err),
				Output:  "",
			}
		} else {
			result = resourceactions.ActionResult{
				Success: true,
				Message: fmt.Sprintf("Successfully uploaded blob '%s' to container '%s'", blobName, containerName),
				Output:  fmt.Sprintf("Blob '%s' is now available in container", blobName),
			}
		}

		return storageActionMsg{action: "upload-blob", result: result}
	}
}

// deleteBlobCmd deletes a blob from a container
func deleteBlobCmd(accountName, containerName, blobName string) tea.Cmd {
	return func() tea.Msg {
		err := storage.DeleteBlob(accountName, containerName, blobName)
		var result resourceactions.ActionResult

		if err != nil {
			result = resourceactions.ActionResult{
				Success: false,
				Message: fmt.Sprintf("Failed to delete blob '%s': %v", blobName, err),
				Output:  "",
			}
		} else {
			result = resourceactions.ActionResult{
				Success: true,
				Message: fmt.Sprintf("Successfully deleted blob '%s' from container '%s'", blobName, containerName),
				Output:  fmt.Sprintf("Blob '%s' has been removed from container", blobName),
			}
		}

		return storageActionMsg{action: "delete-blob", result: result}
	}
}

// getContextualShortcuts returns relevant shortcuts based on the selected resource and current view
func (m model) getContextualShortcuts() string {
	var shortcuts []string

	// Always available shortcuts
	baseShortcuts := []string{"Tab:Switch", "?:Help", "q:Quit"}

	// Context-specific shortcuts based on selected resource
	if m.selectedResource != nil {
		switch m.selectedResource.Type {
		case "Microsoft.Compute/virtualMachines":
			shortcuts = append(shortcuts, []string{
				"s:Start", "S:Stop", "r:Restart",
				"c:SSH", "b:Bastion", "d:Dashboard",
			}...)

		case "Microsoft.ContainerService/managedClusters":
			shortcuts = append(shortcuts, []string{
				"s:Start", "S:Stop", "p:Pods",
				"y:Deployments", "n:Nodes", "v:Services",
			}...)

		case "Microsoft.ContainerInstance/containerGroups":
			shortcuts = append(shortcuts, []string{
				"s:Start", "S:Stop", "r:Restart",
				"L:Logs", "E:Exec", "a:Attach", "u:Scale", "I:Details",
			}...)

		case "Microsoft.Network/virtualNetworks":
			shortcuts = append(shortcuts, []string{
				"V:VNet Details", "N:Network Dashboard",
				"Z:Topology", "A:AI Analysis", "C:Create VNet",
			}...)

		case "Microsoft.Network/networkSecurityGroups":
			shortcuts = append(shortcuts, []string{
				"G:NSG Details", "N:Network Dashboard",
				"Z:Topology", "A:AI Analysis", "Ctrl+N:Create NSG",
			}...)

		case "Microsoft.Storage/storageAccounts":
			shortcuts = append(shortcuts, []string{
				"T:List Containers", "Shift+T:Create Container", "B:List Blobs",
				"U:Upload Blob", "Ctrl+X:Delete Item", "d:Dashboard", "R:Refresh",
			}...)

		case "Microsoft.KeyVault/vaults":
			shortcuts = append(shortcuts, []string{
				"K:List Secrets", "Shift+K:Create Secret", "Ctrl+D:Delete Secret",
				"d:Dashboard", "R:Refresh",
			}...)

		default:
			// Generic resource shortcuts
			shortcuts = append(shortcuts, []string{
				"d:Dashboard", "R:Refresh",
			}...)
		}
	} else {
		// No resource selected - show navigation shortcuts
		shortcuts = append(shortcuts, []string{
			"N:Network Dashboard", "Z:Topology", "A:AI Analysis",
			"Space:Expand", "Enter:Select", "R:Refresh",
		}...)
	}

	// Add base shortcuts
	shortcuts = append(shortcuts, baseShortcuts...)

	return strings.Join(shortcuts, " ")
}

// getTerraformShortcuts returns relevant shortcuts based on the current Terraform mode
func (m model) getTerraformShortcuts() string {
	var shortcuts []string

	switch m.terraformMode {
	case "menu":
		shortcuts = append(shortcuts, []string{
			"↑/↓:Navigate", "Enter:Select", "Esc:Close",
		}...)

	case "folder-select":
		shortcuts = append(shortcuts, []string{
			"↑/↓:Navigate", "Enter:Select", "Esc:Back",
		}...)

	case "analysis":
		shortcuts = append(shortcuts, []string{
			"↑/↓:Scroll", "Enter:Back", "Esc:Close",
		}...)

	case "deployment":
		shortcuts = append(shortcuts, []string{
			"↑/↓:Navigate", "Enter:Select", "Esc:Back",
		}...)
	}

	// Always available shortcuts
	baseShortcuts := []string{"Ctrl+T:Menu", "?:Help"}
	shortcuts = append(shortcuts, baseShortcuts...)

	return strings.Join(shortcuts, " ")
}

// getSettingsShortcuts returns relevant shortcuts based on the current Settings mode
func (m model) getSettingsShortcuts() string {
	var shortcuts []string

	switch m.settingsMode {
	case "menu":
		shortcuts = append(shortcuts, []string{
			"↑/↓:Navigate", "Enter:Select", "Esc:Close",
		}...)

	case "config-view":
		shortcuts = append(shortcuts, []string{
			"Enter:Back", "Esc:Close",
		}...)

	case "folder-browser":
		shortcuts = append(shortcuts, []string{
			"↑/↓:Navigate", "Enter:Select", "Esc:Back",
		}...)

	case "edit-setting":
		shortcuts = append(shortcuts, []string{
			"Esc:Back",
		}...)
	}

	// Always available shortcuts
	baseShortcuts := []string{"Ctrl+,:Menu", "?:Help"}
	shortcuts = append(shortcuts, baseShortcuts...)

	return strings.Join(shortcuts, " ")
}

func initModel() model {
	// Initialize AI provider with auto-detection (GitHub Copilot or OpenAI)
	ai := openai.NewAIProviderAuto()

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
		helpScrollOffset:       0,
		navigationStack:        []string{}, // Initialize navigation stack
		// Initialize search functionality
		searchEngine:      search.NewSearchEngine(),
		searchMode:        false,
		searchQuery:       "",
		searchResults:     []search.SearchResult{},
		searchResultIndex: 0,
		searchSuggestions: []string{},
		showSearchResults: false,
		searchHistory:     []string{},
		filteredResources: []AzureResource{},
		// Initialize Terraform functionality
		showTerraformPopup:  false,
		terraformMenuIndex:  0,
		terraformMode:       "menu",
		terraformFolderPath: "",
		terraformMenuOptions: []string{
			"Browse Folders",
			"Create from Template",
			"Analyze Code",
			"Terraform Operations",
			"Open External Editor",
		},
		terraformFolders:      []string{},
		terraformAnalysis:     "",
		terraformMenuAction:   "",
		terraformScrollOffset: 0,
		// Initialize Settings functionality
		showSettingsPopup:     false,
		settingsMenuIndex:     0,
		settingsMode:          "menu",
		settingsCurrentPath:   "",
		settingsFolders:       []string{},
		settingsConfigContent: "",
		settingsEditKey:       "",
		settingsEditValue:     "",
		settingsCurrentConfig: nil,
		// Initialize Subscription selection functionality
		currentSubscription:    nil,
		showSubscriptionPopup:  false,
		subscriptionMenuIndex:  0,
		availableSubscriptions: []Subscription{},
		subscriptionMenuMode:   "menu",
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		loadDataCmd(),
		getCurrentSubscriptionCmd(),
	)
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
		// Update search engine with new resources
		m.updateSearchEngine()

	case resourceDetailsLoadedMsg:
		m.selectedResource = &msg.resource
		m.resourceDetails = msg.details
		// AI analysis is now manual-only by default - users must press 'a' to trigger
		// Auto-analysis can be enabled by setting AZURE_TUI_AUTO_AI="true"
		autoAI := os.Getenv("AZURE_TUI_AUTO_AI") == "true" // Default to false - manual trigger only
		if m.aiProvider != nil && autoAI {
			m.actionInProgress = true
			return m, loadAIDescriptionCmd(m.aiProvider, msg.resource, msg.details)
		}

	case aiDescriptionLoadedMsg:
		m.actionInProgress = false
		m.aiDescription = msg.description

	case resourceActionMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		if msg.result.Success && m.selectedResource != nil {
			return m, loadResourceDetailsCmd(*m.selectedResource)
		}

	case networkDashboardMsg:
		// Final dashboard loaded - stop progress and show result
		m.actionInProgress = false
		m.networkLoadingInProgress = false
		m.networkDashboardContent = msg.content
		m.pushView("network-dashboard")
		// Add debug logging
		m.logEntries = append(m.logEntries, "DEBUG: Network Dashboard loaded successfully")

	case networkLoadingProgressMsg:
		// Handle network loading progress updates
		m.networkDashboardContent = network.RenderNetworkLoadingProgress(msg.progress)
		m.pushView("network-dashboard")

		// If this is the initial progress message (0%), start the async loading
		if msg.progress.ProgressPercentage == 0.0 && msg.progress.CompletedOperations == 0 {
			m.networkLoadingInProgress = true
			m.networkLoadingStartTime = time.Now()
			return m, loadNetworkDashboardWithProgressCmd()
		}

	case progressTickMsg:
		// Handle progress animation ticks during network loading
		if m.networkLoadingInProgress {
			// Create simulated progress update
			elapsed := time.Since(m.networkLoadingStartTime).Seconds()
			estimatedTotal := 15.0 // Estimated total time in seconds

			// Calculate realistic progress based on elapsed time
			simulatedProgress := (elapsed / estimatedTotal) * 95.0 // Cap at 95% until real completion
			if simulatedProgress > 95.0 {
				simulatedProgress = 95.0
			}

			simulatedCompletedOps := int(simulatedProgress / 100.0 * 7.0)

			progress := network.NetworkLoadingProgress{
				CurrentOperation:       fmt.Sprintf("Loading network resources... (%.1fs elapsed)", elapsed),
				TotalOperations:        7,
				CompletedOperations:    simulatedCompletedOps,
				ProgressPercentage:     simulatedProgress,
				ResourceProgress:       make(map[string]network.ResourceProgress),
				Errors:                 []string{},
				StartTime:              m.networkLoadingStartTime,
				EstimatedTimeRemaining: fmt.Sprintf("%.1fs remaining", estimatedTotal-elapsed),
			}

			// Add resource-specific progress simulation
			resourceTypes := []string{"VirtualNetworks", "NetworkSecurityGroups", "RouteTables", "PublicIPs", "NetworkInterfaces", "LoadBalancers", "Firewalls"}
			for i, resType := range resourceTypes {
				var status string
				if i < simulatedCompletedOps {
					status = "completed"
				} else if i == simulatedCompletedOps {
					status = "loading"
				} else {
					status = "pending"
				}

				progress.ResourceProgress[resType] = network.ResourceProgress{
					ResourceType: resType,
					Status:       status,
					StartTime:    m.networkLoadingStartTime.Add(time.Duration(i) * time.Second * 2),
					Count:        0,
				}
			}

			m.networkDashboardContent = network.RenderNetworkLoadingProgress(progress)

			// Continue ticking if still in progress
			if m.networkLoadingInProgress {
				return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
					return progressTickMsg{}
				})
			}
		}

		// Handle progress animation ticks during topology loading
		if m.topologyLoadingInProgress {
			// Create simulated progress update
			elapsed := time.Since(m.topologyLoadingStartTime).Seconds()
			estimatedTotal := 12.0 // Estimated total time in seconds for topology

			// Calculate realistic progress based on elapsed time
			simulatedProgress := (elapsed / estimatedTotal) * 95.0 // Cap at 95% until real completion
			if simulatedProgress > 95.0 {
				simulatedProgress = 95.0
			}

			simulatedCompletedOps := int(simulatedProgress / 100.0 * 6.0)

			progress := network.NetworkLoadingProgress{
				CurrentOperation:       fmt.Sprintf("Building network topology... (%.1fs elapsed)", elapsed),
				TotalOperations:        6,
				CompletedOperations:    simulatedCompletedOps,
				ProgressPercentage:     simulatedProgress,
				ResourceProgress:       make(map[string]network.ResourceProgress),
				Errors:                 []string{},
				StartTime:              m.topologyLoadingStartTime,
				EstimatedTimeRemaining: fmt.Sprintf("%.1fs remaining", estimatedTotal-elapsed),
			}

			// Add topology-specific progress simulation
			topologySteps := []string{"VirtualNetworks", "Subnets", "NetworkInterfaces", "PublicIPs", "RouteTables", "SecurityGroups"}
			for i, stepType := range topologySteps {
				var status string
				if i < simulatedCompletedOps {
					status = "completed"
				} else if i == simulatedCompletedOps {
					status = "loading"
				} else {
					status = "pending"
				}

				progress.ResourceProgress[stepType] = network.ResourceProgress{
					ResourceType: stepType,
					Status:       status,
					StartTime:    m.topologyLoadingStartTime.Add(time.Duration(i) * time.Second * 2),
					Count:        0,
				}
			}

			m.networkTopologyContent = network.RenderNetworkTopologyLoadingProgress(progress)

			// Continue ticking if still in progress
			if m.topologyLoadingInProgress {
				return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
					return progressTickMsg{}
				})
			}
		}

		// Handle progress animation ticks during dashboard loading
		if m.dashboardLoadingInProgress {
			// Create simulated progress update
			elapsed := time.Since(m.dashboardLoadingStartTime).Seconds()
			estimatedTotal := 8.0 // Estimated total time in seconds for dashboard

			// Calculate realistic progress based on elapsed time
			simulatedProgress := (elapsed / estimatedTotal) * 95.0 // Cap at 95% until real completion
			if simulatedProgress > 95.0 {
				simulatedProgress = 95.0
			}

			simulatedCompletedOps := int(simulatedProgress / 100.0 * 5.0)

			progress := resourcedetails.DashboardLoadingProgress{
				CurrentOperation:       fmt.Sprintf("Loading dashboard data... (%.1fs elapsed)", elapsed),
				TotalOperations:        5,
				CompletedOperations:    simulatedCompletedOps,
				ProgressPercentage:     simulatedProgress,
				DataProgress:           make(map[string]resourcedetails.DataProgress),
				Errors:                 []string{},
				StartTime:              m.dashboardLoadingStartTime,
				EstimatedTimeRemaining: fmt.Sprintf("%.1fs remaining", estimatedTotal-elapsed),
			}

			// Add dashboard-specific progress simulation
			dashboardSteps := []string{"ResourceDetails", "Metrics", "UsageMetrics", "Alarms", "LogEntries"}
			for i, stepType := range dashboardSteps {
				var status string
				if i < simulatedCompletedOps {
					status = "completed"
				} else if i == simulatedCompletedOps {
					status = "loading"
				} else {
					status = "pending"
				}

				progress.DataProgress[stepType] = resourcedetails.DataProgress{
					DataType:  stepType,
					Status:    status,
					StartTime: m.dashboardLoadingStartTime.Add(time.Duration(i) * time.Second * 2),
					Count:     0,
				}
			}

			// Render progress and continue if still loading
			progressContent := tui.RenderDashboardLoadingProgress(progress)
			// Store rendered progress content for dashboard view
			_ = progressContent // Mark as used

			// Continue ticking if still in progress
			if m.dashboardLoadingInProgress {
				return m, tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
					return progressTickMsg{}
				})
			}
		}

	case networkLoadingProgressWithContinuationMsg:
		// Handle progress updates with continuation
		m.networkDashboardContent = network.RenderNetworkLoadingProgress(msg.progress)
		m.pushView("network-dashboard")

		// If there are more progress updates, continue with the next one
		if len(msg.remainingUpdates) > 0 {
			// Add a small delay to make progress visible
			return m, tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
				return networkLoadingProgressWithContinuationMsg{
					progress:         msg.remainingUpdates[0],
					remainingUpdates: msg.remainingUpdates[1:],
					finalDashboard:   msg.finalDashboard,
					finalError:       msg.finalError,
				}
			})
		}

		// No more progress updates, show final result
		var dashboardContent string
		if msg.finalError != nil && msg.finalDashboard == nil {
			dashboardContent = fmt.Sprintf("Error loading network dashboard: %v", msg.finalError)
		} else {
			dashboardContent = network.RenderNetworkDashboard()
		}

		// Set final content and transition to dashboard view
		m.actionInProgress = false
		m.networkDashboardContent = dashboardContent
		return m, nil

	case networkTopologyMsg:
		// Final topology loaded - stop progress and show result
		m.actionInProgress = false
		m.topologyLoadingInProgress = false
		m.networkTopologyContent = msg.content
		m.pushView("network-topology")
		// Add debug logging
		m.logEntries = append(m.logEntries, "DEBUG: Network Topology loaded successfully")

	case networkTopologyLoadingProgressMsg:
		// Handle network topology loading progress updates
		m.networkTopologyContent = network.RenderNetworkTopologyLoadingProgress(msg.progress)
		m.pushView("network-topology")

		// If this is the initial progress message (0%), start the async loading
		if msg.progress.ProgressPercentage == 0.0 && msg.progress.CompletedOperations == 0 {
			m.topologyLoadingInProgress = true
			m.topologyLoadingStartTime = time.Now()
			return m, loadNetworkTopologyWithProgressCmd()
		}

	case networkTopologyLoadingProgressWithContinuationMsg:
		// Handle topology progress updates with continuation
		m.networkTopologyContent = network.RenderNetworkTopologyLoadingProgress(msg.progress)
		m.pushView("network-topology")

		// If there are more progress updates, continue with the next one
		if len(msg.remainingUpdates) > 0 {
			// Add a small delay to make progress visible
			return m, tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
				return networkTopologyLoadingProgressWithContinuationMsg{
					progress:         msg.remainingUpdates[0],
					remainingUpdates: msg.remainingUpdates[1:],
					finalTopology:    msg.finalTopology,
					finalError:       msg.finalError,
				}
			})
		}

		// No more progress updates, show final result
		var topologyContent string
		if msg.finalError != nil && msg.finalTopology == nil {
			topologyContent = fmt.Sprintf("Error loading network topology: %v", msg.finalError)
		} else {
			topologyContent = network.RenderNetworkTopology()
		}

		// Set final content and transition to topology view
		m.actionInProgress = false
		m.networkTopologyContent = topologyContent
		return m, nil

	case vnetDetailsMsg:
		m.actionInProgress = false
		m.vnetDetailsContent = msg.content
		m.pushView("vnet-details")

	case nsgDetailsMsg:
		m.actionInProgress = false
		m.nsgDetailsContent = msg.content
		m.pushView("nsg-details")

	// Dashboard loading progress message handlers
	case dashboardLoadingProgressMsg:
		// Handle dashboard loading progress updates
		m.pushView("dashboard")

		// If this is the initial progress message (0%), start the async loading
		if msg.progress.ProgressPercentage == 0.0 && msg.progress.CompletedOperations == 0 {
			m.dashboardLoadingInProgress = true
			m.dashboardLoadingStartTime = time.Now()
			// Store the resource ID for the dashboard
			if m.selectedResource != nil {
				return m, loadDashboardWithProgressCmd(m.selectedResource.ID)
			}
		}

	case dashboardDataLoadedMsg:
		// Final dashboard loaded - stop progress and show result
		m.actionInProgress = false
		m.dashboardLoadingInProgress = false
		m.dashboardData = msg.data
		m.pushView("dashboard")
		// Add debug logging
		m.logEntries = append(m.logEntries, "DEBUG: Enhanced Dashboard loaded successfully")

	case dashboardLoadingProgressWithContinuationMsg:
		// Handle dashboard progress updates with continuation
		m.pushView("dashboard")

		// If there are more progress updates, continue with the next one
		if len(msg.remainingUpdates) > 0 {
			// Add a small delay to make progress visible
			return m, tea.Tick(time.Millisecond*200, func(t time.Time) tea.Msg {
				return dashboardLoadingProgressWithContinuationMsg{
					progress:         msg.remainingUpdates[0],
					remainingUpdates: msg.remainingUpdates[1:],
					finalData:        msg.finalData,
					finalError:       msg.finalError,
				}
			})
		}

		// No more progress updates, show final result
		var dashboardContent string
		if msg.finalError != nil && msg.finalData == nil {
			dashboardContent = fmt.Sprintf("Error loading dashboard: %v", msg.finalError)
		} else if m.selectedResource != nil {
			dashboardContent = tui.RenderComprehensiveDashboard(m.selectedResource.Name, msg.finalData)
		} else {
			dashboardContent = "No resource selected"
		}

		// Set final content and transition to dashboard view
		m.actionInProgress = false
		m.dashboardData = msg.finalData
		// Store the rendered content if needed for the view
		_ = dashboardContent // Use the content (mark as used)
		return m, nil

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

	case keyVaultSecretsMsg:
		m.keyVaultSecrets = msg.secrets
		m.keyVaultSecretsContent = fmt.Sprintf("Secrets in Vault '%s':\n", msg.vaultName)
		for _, secret := range msg.secrets {
			m.keyVaultSecretsContent += fmt.Sprintf("- %s\n", secret.Name)
		}
		m.pushView("keyvault-secrets")

	case keyVaultSecretDetailsMsg:
		m.selectedSecret = msg.secret
		m.keyVaultSecretDetailsContent = keyvault.RenderSecretDetails(msg.secret)
		m.pushView("keyvault-secret-details")

	case keyVaultSecretActionMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		if msg.result.Success {
			return m, listKeyVaultSecretsCmd(m.selectedResource.Name)
		}

	// Storage Progress message handlers
	case storageLoadingStartMsg:
		m.actionInProgress = true
		switch msg.operation {
		case "containers":
			// Start container loading with progress
			progress := storage.StorageLoadingProgress{
				CurrentOperation:    "Starting container listing...",
				TotalOperations:     1,
				CompletedOperations: 0,
				ProgressPercentage:  0.0,
				StartTime:           time.Now(),
			}
			m.storageContainersContent = storage.RenderStorageLoadingProgress(progress)
			m.pushView("storage-containers")
			return m, listStorageContainersWithProgressCmd(msg.accountName)
		case "blobs":
			// Start blob loading with progress
			progress := storage.StorageLoadingProgress{
				CurrentOperation:    "Starting blob listing...",
				TotalOperations:     1,
				CompletedOperations: 0,
				ProgressPercentage:  0.0,
				StartTime:           time.Now(),
			}
			m.storageBlobsContent = storage.RenderStorageLoadingProgress(progress)
			m.pushView("storage-blobs")
			return m, listStorageBlobsWithProgressCmd(msg.accountName, m.currentContainer)
		}

	case storageLoadingProgressMsg:
		// Update progress display
		switch m.activeView {
		case "storage-containers":
			m.storageContainersContent = storage.RenderStorageLoadingProgress(msg.progress)
		case "storage-blobs":
			m.storageBlobsContent = storage.RenderStorageLoadingProgress(msg.progress)
		}

	case storageLoadingCompleteMsg:
		m.actionInProgress = false
		if msg.success {
			switch msg.operation {
			case "containers":
				if containers, ok := msg.data.([]storage.Container); ok {
					m.storageContainers = containers
					m.storageContainersContent = storage.RenderStorageContainersView(m.currentStorageAccount, containers)
				}
			case "blobs":
				if blobs, ok := msg.data.([]storage.Blob); ok {
					m.storageBlobs = blobs
					m.storageBlobsContent = storage.RenderStorageBlobsView(m.currentStorageAccount, m.currentContainer, blobs)
				}
			}
		} else {
			return m, func() tea.Msg {
				return errorMsg{error: fmt.Sprintf("Storage operation failed: %v", msg.error)}
			}
		}

	// Storage Account message handlers
	case storageContainersMsg:
		m.actionInProgress = false
		m.storageContainers = msg.containers
		m.currentStorageAccount = msg.accountName
		m.storageContainersContent = storage.RenderStorageContainersView(msg.accountName, msg.containers)
		m.pushView("storage-containers")

	case storageBlobsMsg:
		m.actionInProgress = false
		m.storageBlobs = msg.blobs
		m.currentContainer = msg.containerName
		m.storageBlobsContent = storage.RenderStorageBlobsView(msg.accountName, msg.containerName, msg.blobs)
		m.pushView("storage-blobs")

	case storageBlobDetailsMsg:
		m.actionInProgress = false
		m.selectedBlob = msg.blob
		m.storageBlobDetailsContent = storage.RenderBlobDetails(msg.blob)
		m.pushView("storage-blob-details")

	case storageActionMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		if msg.result.Success {
			// Refresh the appropriate view based on the action
			switch msg.action {
			case "create-container", "delete-container":
				if m.currentStorageAccount != "" {
					return m, listStorageContainersCmd(m.currentStorageAccount)
				}
			case "upload-blob", "delete-blob":
				if m.currentStorageAccount != "" && m.currentContainer != "" {
					return m, listStorageBlobsCmd(m.currentStorageAccount, m.currentContainer)
				}
			}
		}

	// Terraform message handlers
	case terraformFoldersLoadedMsg:
		m.terraformFolders = msg.folders

	case terraformAnalysisMsg:
		m.terraformAnalysis = msg.analysis
		m.terraformFolderPath = msg.path
		m.terraformMode = "analysis"
		m.showTerraformPopup = true
		m.terraformScrollOffset = 0 // Reset scroll when loading new analysis

	case terraformOperationMsg:
		m.actionInProgress = false
		status := "Success"
		if !msg.success {
			status = "Failed"
		}
		m.logEntries = append(m.logEntries, fmt.Sprintf("Terraform %s: %s - %s", msg.operation, status, msg.result))

		// Handle template creation success - start deployment workflow
		if msg.operation == "template" && msg.success {
			// Switch to deployment mode and show deployment options
			m.terraformMode = "deployment"
			m.terraformMenuIndex = 0
			// Keep the popup open to show deployment options
		}

	// Settings message handlers
	case settingsFoldersLoadedMsg:
		m.settingsFolders = msg.folders

	case settingsConfigLoadedMsg:
		m.settingsCurrentConfig = msg.config
		m.settingsConfigContent = msg.content
		if msg.config != nil {
			m.settingsMode = "config-view"
		}

	case settingsConfigSavedMsg:
		m.actionInProgress = false
		if msg.success {
			m.logEntries = append(m.logEntries, "Settings: "+msg.message)
		} else {
			m.logEntries = append(m.logEntries, "Settings Error: "+msg.message)
		}

	case currentSubscriptionMsg:
		m.currentSubscription = msg.subscription

	case subscriptionMenuMsg:
		m.availableSubscriptions = msg.subscriptions
		m.subscriptionMenuMode = "menu"

	case subscriptionSelectedMsg:
		m.actionInProgress = false
		if msg.success {
			m.currentSubscription = &msg.subscription
			m.logEntries = append(m.logEntries, "Subscription: "+msg.message)
			// Reload resource groups for the new subscription
			return m, loadDataCmd()
		} else {
			m.logEntries = append(m.logEntries, "Subscription Error: "+msg.message)
		}

	case errorMsg:
		m.loadingState = "error"

	case tea.KeyMsg:
		// Handle popups first (they should take priority over search mode)

		// Handle Terraform popup navigation
		if m.showTerraformPopup {
			switch msg.String() {
			case "escape":
				if m.terraformMode == "analysis" || m.terraformMode == "deployment" {
					// Go back to menu from analysis or deployment
					m.terraformMode = "menu"
					m.terraformMenuIndex = 0
					m.terraformScrollOffset = 0 // Reset scroll when going back to menu
				} else {
					m.showTerraformPopup = false
				}
			case "enter":
				if m.terraformMode == "analysis" {
					// Go back to menu from analysis
					m.terraformMode = "menu"
					m.terraformMenuIndex = 0
					m.terraformScrollOffset = 0 // Reset scroll when going back to menu
				} else if m.terraformMode == "deployment" {
					return m.handleTerraformDeploymentSelection()
				} else {
					return m.handleTerraformMenuSelection()
				}
			case "j", "down":
				if m.terraformMode == "menu" {
					m.terraformMenuIndex = (m.terraformMenuIndex + 1) % len(m.terraformMenuOptions)
				} else if m.terraformMode == "folder-select" {
					m.terraformMenuIndex = (m.terraformMenuIndex + 1) % len(m.terraformFolders)
				} else if m.terraformMode == "deployment" {
					// Navigate deployment options (6 options)
					m.terraformMenuIndex = (m.terraformMenuIndex + 1) % 6
				} else if m.terraformMode == "analysis" {
					// Scroll down in analysis text
					m.terraformScrollOffset += 1
				}
			case "k", "up":
				if m.terraformMode == "menu" {
					m.terraformMenuIndex = (m.terraformMenuIndex - 1 + len(m.terraformMenuOptions)) % len(m.terraformMenuOptions)
				} else if m.terraformMode == "folder-select" {
					m.terraformMenuIndex = (m.terraformMenuIndex - 1 + len(m.terraformFolders)) % len(m.terraformFolders)
				} else if m.terraformMode == "deployment" {
					// Navigate deployment options (6 options)
					m.terraformMenuIndex = (m.terraformMenuIndex - 1 + 6) % 6
				} else if m.terraformMode == "analysis" {
					// Scroll up in analysis text (prevent negative scroll)
					if m.terraformScrollOffset > 0 {
						m.terraformScrollOffset -= 1
					}
				}
			}
			return m, nil
		}

		// Handle Settings popup navigation
		if m.showSettingsPopup {
			switch msg.String() {
			case "escape":
				if m.settingsMode == "config-view" || m.settingsMode == "folder-browser" {
					// Go back to main settings menu
					m.settingsMode = "menu"
					m.settingsMenuIndex = 0
				} else {
					m.showSettingsPopup = false
				}
			case "enter":
				if m.settingsMode == "config-view" {
					// Go back to menu from config view
					m.settingsMode = "menu"
					m.settingsMenuIndex = 0
				} else {
					return m.handleSettingsMenuSelection()
				}
			case "j", "down":
				if m.settingsMode == "menu" {
					settingsMenuOptions := []string{"View Config", "Edit Config", "Browse Folders", "Reset to Defaults"}
					m.settingsMenuIndex = (m.settingsMenuIndex + 1) % len(settingsMenuOptions)
				} else if m.settingsMode == "folder-browser" {
					m.settingsMenuIndex = (m.settingsMenuIndex + 1) % len(m.settingsFolders)
				}
			case "k", "up":
				if m.settingsMode == "menu" {
					settingsMenuOptions := []string{"View Config", "Edit Config", "Browse Folders", "Reset to Defaults"}
					m.settingsMenuIndex = (m.settingsMenuIndex - 1 + len(settingsMenuOptions)) % len(settingsMenuOptions)
				} else if m.settingsMode == "folder-browser" {
					m.settingsMenuIndex = (m.settingsMenuIndex - 1 + len(m.settingsFolders)) % len(m.settingsFolders)
				}
			}
			return m, nil
		}

		// Handle Subscription popup navigation
		if m.showSubscriptionPopup {
			switch msg.String() {
			case "escape":
				m.showSubscriptionPopup = false
			case "enter":
				if m.subscriptionMenuMode == "menu" && len(m.availableSubscriptions) > 0 {
					selectedSubscription := m.availableSubscriptions[m.subscriptionMenuIndex]
					m.showSubscriptionPopup = false
					return m, selectSubscriptionCmd(selectedSubscription.ID)
				}
			case "j", "down":
				if m.subscriptionMenuMode == "menu" && len(m.availableSubscriptions) > 0 {
					m.subscriptionMenuIndex = (m.subscriptionMenuIndex + 1) % len(m.availableSubscriptions)
				}
			case "k", "up":
				if m.subscriptionMenuMode == "menu" && len(m.availableSubscriptions) > 0 {
					m.subscriptionMenuIndex = (m.subscriptionMenuIndex - 1 + len(m.availableSubscriptions)) % len(m.availableSubscriptions)
				}
			}
			return m, nil
		}

		// Handle Help popup navigation
		if m.showHelpPopup {
			switch msg.String() {
			case "escape", "?":
				m.showHelpPopup = false
				m.helpScrollOffset = 0 // Reset scroll when closing
			case "j", "down":
				// Scroll down in help content
				m.helpScrollOffset += 1
			case "k", "up":
				// Scroll up in help content (prevent negative scroll)
				if m.helpScrollOffset > 0 {
					m.helpScrollOffset -= 1
				}
			}
			return m, nil
		}

		// Handle search mode input after popups
		if m.searchMode {
			switch msg.String() {
			case "escape":
				m.exitSearchMode()
			case "enter":
				// Execute search and add to history
				if m.searchQuery != "" {
					m.addToSearchHistory(m.searchQuery)
					m.performSearch()
				}
			case "backspace":
				// Remove last character from search query
				if len(m.searchQuery) > 0 {
					m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
					m.performSearch()
					m.updateSearchSuggestions()
				}
			case "tab":
				// Accept first suggestion if available
				if len(m.searchSuggestions) > 0 {
					m.searchQuery = m.searchSuggestions[0]
					m.performSearch()
				}
			case "down", "ctrl+j":
				// Navigate to next search result
				if m.showSearchResults {
					m.navigateSearchResults(1)
					if m.selectedResource != nil {
						return m, loadResourceDetailsCmd(*m.selectedResource)
					}
				}
			case "up", "ctrl+k":
				// Navigate to previous search result
				if m.showSearchResults {
					m.navigateSearchResults(-1)
					if m.selectedResource != nil {
						return m, loadResourceDetailsCmd(*m.selectedResource)
					}
				}
			default:
				// Add character to search query
				if len(msg.String()) == 1 && msg.String() >= " " && msg.String() <= "~" {
					m.searchQuery += msg.String()
					m.performSearch()
					m.updateSearchSuggestions()
				}
			}
			return m, nil
		}

		// Regular key handling when not in search mode
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.selectedPanel = (m.selectedPanel + 1) % 2

		// Terraform Integration - Primary Access Key
		case "ctrl+t":
			if !m.showTerraformPopup {
				m.showTerraformPopup = true
				m.terraformMenuIndex = 0
				return m, loadTerraformFoldersCmd()
			} else {
				m.showTerraformPopup = false
			}

		// Settings Menu - Primary Access Key
		case "ctrl+,":
			if !m.showSettingsPopup {
				m.showSettingsPopup = true
				m.settingsMode = "menu"
				m.settingsMenuIndex = 0
				return m, loadSettingsConfigCmd()
			} else {
				m.showSettingsPopup = false
			}

		// Subscription Selection Menu - Primary Access Key
		case "ctrl+a":
			if !m.showSubscriptionPopup {
				m.showSubscriptionPopup = true
				m.subscriptionMenuMode = "loading"
				m.subscriptionMenuIndex = 0
				return m, loadSubscriptionMenuCmd()
			} else {
				m.showSubscriptionPopup = false
			}

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
			switch m.selectedPanel {
			case 0:
				// Left panel scrolling up
				if m.leftPanelScrollOffset > 0 {
					m.leftPanelScrollOffset--
				}
			case 1:
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
		case "/":
			// Enter search mode
			if !m.searchMode {
				m.enterSearchMode()
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
		case "D", "shift+d":
			// Enhanced dashboard with progress and real data (Shift+D)
			if m.selectedResource != nil && !m.actionInProgress {
				// Additional safety checks to prevent crashes
				if m.selectedResource.ID == "" {
					m.logEntries = append(m.logEntries, "ERROR: Cannot load dashboard - resource ID is empty")
					return m, nil
				}
				m.actionInProgress = true
				m.dashboardLoadingInProgress = true
				m.dashboardLoadingStartTime = time.Now()
				m.dashboardData = nil // Clear any existing data
				m.pushView("dashboard")
				return m, showEnhancedDashboardCmd(m.selectedResource.ID)
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
		case "y":
			// AKS deployments (moved from 'D' to avoid conflict with enhanced dashboard)
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.ContainerService/managedClusters" {
				m.actionInProgress = true
				return m, executeResourceActionCmd("deployments", *m.selectedResource)
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
			// Key Vault: Create secret (with demo values for now)
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.KeyVault/vaults" {
				m.actionInProgress = true
				// For demo purposes, create a secret with sample name and value
				// In a real implementation, this would open a form dialog
				return m, createKeyVaultSecretCmd(m.selectedResource.Name, "demo-secret", "demo-value", map[string]string{"created-by": "azure-tui"})
			}
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
			// AI Analysis for selected resource (general case)
			if m.selectedResource != nil && !m.actionInProgress && m.aiProvider != nil {
				m.actionInProgress = true
				return m, loadAIDescriptionCmd(m.aiProvider, *m.selectedResource, m.resourceDetails)
			}
			// Attach to container (only for container instances) - fallback if no AI provider
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

		// Key Vault Management Actions
		case "K":
			// List Key Vault secrets
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.KeyVault/vaults" {
				m.actionInProgress = true
				return m, listKeyVaultSecretsCmd(m.selectedResource.Name)
			}
		case "shift+k":
			// Create Key Vault secret
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.KeyVault/vaults" {
				m.actionInProgress = true
				// For demo purposes, create a secret with sample name and value
				// In a real implementation, this would open a form dialog
				return m, createKeyVaultSecretCmd(m.selectedResource.Name, "demo-secret", "demo-value", map[string]string{"created-by": "azure-tui"})
			}
		case "ctrl+d":
			// Delete Key Vault secret (demo - would need secret selection in real implementation)
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.KeyVault/vaults" {
				m.actionInProgress = true
				// For demo purposes, delete a known secret name
				// In a real implementation, this would show a list to select from
				return m, deleteKeyVaultSecretCmd(m.selectedResource.Name, "demo-secret")
			}

		// Storage Account Management Actions
		case "T":
			// List Storage Containers (using T for sTroage containers)
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.Storage/storageAccounts" {
				m.actionInProgress = true
				return m, listStorageContainersCmd(m.selectedResource.Name)
			}
		case "shift+t":
			// Create Storage Container
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.Storage/storageAccounts" {
				m.actionInProgress = true
				// For demo purposes, create a container with a sample name
				// In a real implementation, this would open a form dialog
				return m, createStorageContainerCmd(m.selectedResource.Name, "demo-container")
			}
		case "B":
			// List Blobs in Container (only available when viewing containers)
			if m.selectedResource != nil && !m.actionInProgress &&
				m.selectedResource.Type == "Microsoft.Storage/storageAccounts" &&
				m.activeView == "storage-containers" && len(m.storageContainers) > 0 {
				m.actionInProgress = true
				// For demo purposes, use the first container
				// In a real implementation, this would allow container selection
				containerName := m.storageContainers[0].Name
				return m, listStorageBlobsCmd(m.selectedResource.Name, containerName)
			}
		case "U":
			// Upload Blob (only available when viewing blobs)
			if m.selectedResource != nil && !m.actionInProgress &&
				m.selectedResource.Type == "Microsoft.Storage/storageAccounts" &&
				m.activeView == "storage-blobs" && m.currentContainer != "" {
				m.actionInProgress = true
				// For demo purposes, simulate uploading a file
				// In a real implementation, this would open a file dialog
				return m, uploadBlobCmd(m.selectedResource.Name, m.currentContainer, "demo-blob.txt", "/tmp/demo-file.txt")
			}
		case "ctrl+x":
			// Delete Storage Item (Container or Blob depending on current view)
			if m.selectedResource != nil && !m.actionInProgress && m.selectedResource.Type == "Microsoft.Storage/storageAccounts" {
				m.actionInProgress = true
				if m.activeView == "storage-containers" && len(m.storageContainers) > 0 {
					// Delete first container for demo
					return m, deleteStorageContainerCmd(m.selectedResource.Name, m.storageContainers[0].Name)
				} else if m.activeView == "storage-blobs" && len(m.storageBlobs) > 0 && m.currentContainer != "" {
					// Delete first blob for demo
					return m, deleteBlobCmd(m.selectedResource.Name, m.currentContainer, m.storageBlobs[0].Name)
				}
			}

		case "R":
			return m, loadDataCmd()
		case "?":
			// Toggle help popup
			m.showHelpPopup = !m.showHelpPopup
		case "escape":
			// Handle escape key for search mode, help popup, or navigation
			if m.searchMode {
				m.exitSearchMode()
			} else if m.showHelpPopup {
				m.showHelpPopup = false
				m.helpScrollOffset = 0 // Reset scroll when closing
			} else {
				// Try to go back to previous view
				if !m.popView() {
					// If no previous view, try to reset to welcome view
					if m.activeView != "welcome" {
						m.activeView = "welcome"
						m.showDashboard = false
						m.selectedResource = nil
						m.rightPanelScrollOffset = 0
						m.leftPanelScrollOffset = 0
					}
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

		// Show current subscription info instead of generic "Azure Dashboard"
		if m.currentSubscription != nil {
			m.statusBar.AddSegment(fmt.Sprintf("☁️ %s", m.currentSubscription.Name), colorBlue, bgDark)
		} else {
			m.statusBar.AddSegment("☁️ Azure Dashboard", colorBlue, bgDark)
		}

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

		// Add search indicators
		if m.searchMode {
			m.statusBar.AddSegment("🔍 Search Mode", colorYellow, bgMedium)
			if m.showSearchResults {
				m.statusBar.AddSegment(fmt.Sprintf("%d Results", len(m.searchResults)), colorGreen, bgMedium)
				if len(m.searchResults) > 0 {
					m.statusBar.AddSegment(fmt.Sprintf("Result %d/%d", m.searchResultIndex+1, len(m.searchResults)), colorPurple, bgMedium)
				}
				m.statusBar.AddSegment("Enter:Select", colorGray, bgLight)
			}
		} else {
			m.statusBar.AddSegment("/:Search", colorGray, bgLight)
		}

		// Add contextual shortcuts
		m.statusBar.AddSegment(m.getContextualShortcuts(), colorGray, bgLight)
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
		// ALWAYS apply left panel scroll offset to maintain independent position
		treeContent = m.renderScrollableContentWithOffset(treeContentRaw, m.height-6, m.leftPanelScrollOffset)
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
		treeContent = "🔍 " + strings.ReplaceAll(treeContent, "\n", "\n   ")
	}

	leftPanel := leftPanelStyle.Render(treeContent)

	// Details panel with scrolling support and STRICT width constraints
	rightContentRaw := ""
	if m.searchMode && m.showSearchResults {
		// Show search results in right panel when in search mode
		rightContentRaw = m.renderSearchResults(rightWidth-4, m.height-2)
	} else {
		rightContentRaw = m.renderResourcePanel(rightWidth-4, m.height-2)
	}

	// Ensure content is properly wrapped to prevent layout breaking
	rightContentWrapped := wrapText(rightContentRaw, rightWidth-8)

	// ALWAYS apply right panel scroll offset to maintain independent position
	rightContent := m.renderScrollableContentWithOffset(rightContentWrapped, m.height-6, m.rightPanelScrollOffset)

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
		rightContent = "📊 " + strings.ReplaceAll(rightContent, "\n", "\n   ")
	}

	rightPanel := rightPanelStyle.Render(rightContent)

	// Join everything
	statusBarContent := ""
	if m.statusBar != nil {
		statusBarContent = m.statusBar.RenderStatusBar()
	}

	// Add search input bar if in search mode
	searchInput := ""
	if m.searchMode {
		searchInput = m.renderSearchInput(m.width)
	}

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Combine status bar, search input, and main content
	var fullView string
	if searchInput != "" {
		fullView = lipgloss.JoinVertical(lipgloss.Left, statusBarContent, searchInput, mainContent)
	} else {
		fullView = lipgloss.JoinVertical(lipgloss.Left, statusBarContent, mainContent)
	}

	// Render help popup if active
	if m.showHelpPopup {
		// Create a comprehensive help content with better table formatting
		var helpContent strings.Builder
		helpContent.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("⌨️  Azure TUI - Keyboard Shortcuts"))
		helpContent.WriteString("\n\n")

		// Create structured table data for better formatting
		var allSections []string

		// Navigation section
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("🧭 Navigation:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("j/k ↑/↓", "Navigate up/down in tree"))
		allSections = append(allSections, renderShortcutRow("h/l ←/→", "Switch between panels"))
		allSections = append(allSections, renderShortcutRow("Space", "Expand/collapse resource groups"))
		allSections = append(allSections, renderShortcutRow("Enter", "Open resource in details panel"))
		allSections = append(allSections, renderShortcutRow("Tab", "Switch between panels"))
		allSections = append(allSections, renderShortcutRow("e", "Expand/collapse complex properties"))
		allSections = append(allSections, renderShortcutRow("Ctrl+j/k", "Scroll up/down in current panel"))
		allSections = append(allSections, "")

		// Search section
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorYellow).Render("🔍 Search:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("/", "Enter search mode"))
		allSections = append(allSections, renderShortcutRow("Enter", "Execute search / Accept suggestion"))
		allSections = append(allSections, renderShortcutRow("Tab", "Accept first suggestion"))
		allSections = append(allSections, renderShortcutRow("↑/↓", "Navigate search results"))
		allSections = append(allSections, renderShortcutRow("Escape", "Exit search mode"))
		allSections = append(allSections, renderShortcutRow("Advanced", "type:vm location:eastus tag:env=prod"))
		allSections = append(allSections, "")

		// Resource Actions section
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorAqua).Render("⚡ Resource Actions:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("s", "Start resource (VMs, Containers)"))
		allSections = append(allSections, renderShortcutRow("S", "Stop resource (VMs, Containers)"))
		allSections = append(allSections, renderShortcutRow("r", "Restart resource (VMs, Containers)"))
		allSections = append(allSections, renderShortcutRow("Shift+D", "Enhanced dashboard with real data"))
		allSections = append(allSections, renderShortcutRow("R", "Refresh all data"))
		allSections = append(allSections, "")

		// Network Management section
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("🌐 Network Management:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("N", "Network Dashboard"))
		allSections = append(allSections, renderShortcutRow("V", "VNet Details (for VNets)"))
		allSections = append(allSections, renderShortcutRow("G", "NSG Details (for NSGs)"))
		allSections = append(allSections, renderShortcutRow("Z", "Network Topology"))
		allSections = append(allSections, renderShortcutRow("A", "AI Network Analysis"))
		allSections = append(allSections, renderShortcutRow("C", "Create VNet"))
		allSections = append(allSections, renderShortcutRow("Ctrl+N", "Create NSG"))
		allSections = append(allSections, "")

		// Terraform Management section
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorAqua).Render("🏗️  Terraform Management:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("Ctrl+T", "Open Terraform Manager"))
		allSections = append(allSections, renderShortcutRow("", "• Browse Terraform projects"))
		allSections = append(allSections, renderShortcutRow("", "• Analyze code"))
		allSections = append(allSections, renderShortcutRow("", "• Execute operations"))
		allSections = append(allSections, renderShortcutRow("", "• Create from templates"))
		allSections = append(allSections, "")

		// Container Management section
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorPurple).Render("🐳 Container Management:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("L", "Get Container Logs"))
		allSections = append(allSections, renderShortcutRow("E", "Exec into Container"))
		allSections = append(allSections, renderShortcutRow("a", "Attach to Container"))
		allSections = append(allSections, renderShortcutRow("u", "Scale Container Resources"))
		allSections = append(allSections, renderShortcutRow("I", "Container Instance Details"))
		allSections = append(allSections, "")

		// SSH & AKS section
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorYellow).Render("🔐 SSH & AKS:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("c", "SSH Connect (VMs)"))
		allSections = append(allSections, renderShortcutRow("b", "Bastion Connect (VMs)"))
		allSections = append(allSections, renderShortcutRow("p", "List Pods (AKS)"))
		allSections = append(allSections, renderShortcutRow("y", "List Deployments (AKS)"))
		allSections = append(allSections, renderShortcutRow("n", "List Nodes (AKS)"))
		allSections = append(allSections, renderShortcutRow("v", "List Services (AKS)"))
		allSections = append(allSections, "")

		// Key Vault Management section
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorGray).Render("🔑 Key Vault Management:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("K", "List Secrets"))
		allSections = append(allSections, renderShortcutRow("Shift+K", "Create Secret"))
		allSections = append(allSections, renderShortcutRow("Ctrl+D", "Delete Secret"))
		allSections = append(allSections, "")

		// Subscription Management section
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorAqua).Render("☁️ Subscription Management:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("Ctrl+A", "Open Subscription Manager"))
		allSections = append(allSections, renderShortcutRow("", "• Switch Azure subscriptions"))
		allSections = append(allSections, renderShortcutRow("", "• View tenant information"))
		allSections = append(allSections, renderShortcutRow("", "• Change active context"))
		allSections = append(allSections, "")

		// Interface section
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorGray).Render("🎮 Interface:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("?", "Show/hide this help"))
		allSections = append(allSections, renderShortcutRow("Ctrl+,", "Open Settings Manager"))
		allSections = append(allSections, renderShortcutRow("Esc", "Navigate back / Close dialogs"))
		allSections = append(allSections, renderShortcutRow("q", "Quit application"))
		allSections = append(allSections, "")

		// Add scroll navigation instructions
		allSections = append(allSections, lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("📜 Help Navigation:"))
		allSections = append(allSections, "")
		allSections = append(allSections, renderShortcutRow("j/k ↑/↓", "Scroll help content"))
		allSections = append(allSections, renderShortcutRow("? / Esc", "Close this help"))

		// Join all sections
		fullHelpContent := strings.Join(allSections, "\n")

		// Apply scrolling to help content
		visibleLines := 20 // Number of lines visible in the popup
		scrolledContent := m.renderScrollableContentWithOffset(fullHelpContent, visibleLines, m.helpScrollOffset)

		helpContent.WriteString(scrolledContent)

		// Create popup style without frames or background
		popupStyle := lipgloss.NewStyle().
			Foreground(fgLight).
			Padding(1, 2).
			Width(78). // Slightly wider for better table formatting
			Align(lipgloss.Left, lipgloss.Top)

		styledPopup := popupStyle.Render(helpContent.String())

		// Create a simple centered layout
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, styledPopup)
	}

	// Render Terraform popup if active
	if m.showTerraformPopup {
		return m.renderTerraformPopup(fullView)
	}

	// Render Settings popup if active
	if m.showSettingsPopup {
		return m.renderSettingsPopup(fullView)
	}

	// Render Subscription popup if active
	if m.showSubscriptionPopup {
		return m.renderSubscriptionPopup(fullView)
	}

	return lipgloss.NewStyle().Background(bgDark).Render(fullView)
}

func (m model) renderTerraformPopup(background string) string {
	var content strings.Builder

	// Title
	title := lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("🏗️  Terraform Manager")
	content.WriteString(title)
	content.WriteString("\n\n")

	switch m.terraformMode {
	case "menu":
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("Select an option:"))
		content.WriteString("\n\n")

		for i, option := range m.terraformMenuOptions {
			style := lipgloss.NewStyle().Foreground(fgMedium)
			if i == m.terraformMenuIndex {
				style = style.Foreground(fgLight).Bold(true)
			}
			prefix := "  "
			if i == m.terraformMenuIndex {
				prefix = "▶ "
			}
			content.WriteString(style.Render(prefix + option))
			content.WriteString("\n")
		}

	case "folder-select":
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("Select a Terraform project folder:"))
		content.WriteString("\n\n")

		if len(m.terraformFolders) == 0 {
			content.WriteString(lipgloss.NewStyle().Foreground(colorYellow).Render("No Terraform projects found (.tf files)"))
			content.WriteString("\n")
			content.WriteString(lipgloss.NewStyle().Faint(true).Render("Create a .tf file in any directory to get started"))
		} else {
			for i, folder := range m.terraformFolders {
				style := lipgloss.NewStyle().Foreground(fgMedium)
				if i == m.terraformMenuIndex {
					style = style.Foreground(fgLight).Bold(true)
				}
				prefix := "  "
				if i == m.terraformMenuIndex {
					prefix = "▶ "
				}

				// Show relative path for better display
				displayPath := folder
				if folder == "." {
					displayPath = "current directory"
				}

				content.WriteString(style.Render(prefix + "📁 " + displayPath))
				content.WriteString("\n")
			}
		}

	case "analysis":
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render(fmt.Sprintf("Code Analysis: %s", m.terraformFolderPath)))
		content.WriteString("\n\n")

		// Apply scrolling to the analysis text
		analysisLines := strings.Split(m.terraformAnalysis, "\n")

		// Calculate visible lines (popup height - header - footer - status bar)
		visibleLines := 15 // Reasonable default for popup content

		// Apply scroll offset
		startLine := m.terraformScrollOffset
		endLine := min(startLine+visibleLines, len(analysisLines))

		// Ensure we don't scroll past the beginning
		if startLine >= len(analysisLines) {
			startLine = max(0, len(analysisLines)-visibleLines)
			m.terraformScrollOffset = startLine
		}

		// Add scroll indicators
		if startLine > 0 {
			content.WriteString(lipgloss.NewStyle().Foreground(colorGray).Render("↑ (more content above - use k/↑ to scroll up)\n"))
		}

		// Render visible lines
		for i := startLine; i < endLine; i++ {
			if i < len(analysisLines) {
				content.WriteString(analysisLines[i])
				content.WriteString("\n")
			}
		}

		// Add scroll indicator at bottom
		if endLine < len(analysisLines) {
			content.WriteString(lipgloss.NewStyle().Foreground(colorGray).Render("↓ (more content below - use j/↓ to scroll down)"))
		}

	case "deployment":
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("🚀 Deploy Template"))
		content.WriteString("\n\n")

		content.WriteString(lipgloss.NewStyle().Foreground(colorGreen).Render("✅ Template created successfully!"))
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("Choose deployment action:"))
		content.WriteString("\n\n")

		deploymentOptions := []string{
			"🔧 Initialize Terraform (terraform init)",
			"📋 Plan Deployment (terraform plan)",
			"🚀 Deploy Infrastructure (terraform apply)",
			"📝 Edit Template Files",
			"📁 Open in External Editor",
			"🏠 Return to Main Menu",
		}

		for i, option := range deploymentOptions {
			style := lipgloss.NewStyle().Foreground(fgMedium)
			if i == m.terraformMenuIndex {
				style = style.Foreground(fgLight).Bold(true)
			}
			prefix := "  "
			if i == m.terraformMenuIndex {
				prefix = "▶ "
			}
			content.WriteString(style.Render(prefix + option))
			content.WriteString("\n")
		}
	}

	content.WriteString("\n\n")

	// Add statusbar with contextual shortcuts
	shortcuts := m.getTerraformShortcuts()
	statusbarStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Padding(0, 1).
		Width(58)

	content.WriteString(statusbarStyle.Render("Terraform: " + shortcuts))
	content.WriteString("\n")

	// Different footer text based on mode (kept for additional context)
	switch m.terraformMode {
	case "analysis":
		content.WriteString(lipgloss.NewStyle().Italic(true).Foreground(colorGray).Render("Press Enter or Esc to return to menu"))
	default:
		content.WriteString(lipgloss.NewStyle().Italic(true).Foreground(colorGray).Render("Navigate: ↑/↓  Select: Enter  Back: Esc"))
	}

	// Create popup style - clean, no borders or backgrounds
	popupStyle := lipgloss.NewStyle().
		Foreground(fgLight).
		Padding(1, 2).
		Width(60).
		Align(lipgloss.Center, lipgloss.Top)

	styledPopup := popupStyle.Render(content.String())

	// Overlay on background
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, styledPopup)
}

func (m model) renderSettingsPopup(background string) string {
	var content strings.Builder

	// Title
	title := lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("⚙️  Settings Manager")
	content.WriteString(title)
	content.WriteString("\n\n")

	switch m.settingsMode {
	case "menu":
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("Select an option:"))
		content.WriteString("\n\n")

		menuOptions := []string{
			"📋 View Configuration",
			"📁 Edit Terraform Directory",
			"🎨 Edit UI Settings",
			"📝 Edit Editor Settings",
			"💾 Save Configuration",
		}

		for i, option := range menuOptions {
			style := lipgloss.NewStyle().Foreground(fgMedium)
			if i == m.settingsMenuIndex {
				style = style.Foreground(fgLight).Bold(true)
			}
			prefix := "  "
			if i == m.settingsMenuIndex {
				prefix = "▶ "
			}
			content.WriteString(style.Render(prefix + option))
			content.WriteString("\n")
		}

	case "config-view":
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("Current Configuration:"))
		content.WriteString("\n\n")
		content.WriteString(m.settingsConfigContent)

	case "folder-browser":
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("Select Terraform Directory:"))
		content.WriteString("\n\n")

		for i, folder := range m.settingsFolders {
			style := lipgloss.NewStyle().Foreground(fgMedium)
			if i == m.settingsMenuIndex {
				style = style.Foreground(fgLight).Bold(true)
			}
			prefix := "  "
			if i == m.settingsMenuIndex {
				prefix = "▶ "
				content.WriteString(style.Render(prefix + folder))
				content.WriteString("\n")
			}
		}

	case "edit-setting":
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render(fmt.Sprintf("Editing: %s", m.settingsEditKey)))
		content.WriteString("\n\n")
		content.WriteString("Current Value:\n")
		content.WriteString(lipgloss.NewStyle().Foreground(colorAqua).Render(m.settingsEditValue))
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().Italic(true).Foreground(colorGray).Render("Note: Direct editing not yet implemented"))
	}

	content.WriteString("\n\n")

	// Add status bar with contextual shortcuts
	shortcuts := m.getSettingsShortcuts()
	statusbarStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Padding(0, 1).
		Width(58)

	content.WriteString(statusbarStyle.Render("Settings: " + shortcuts))
	content.WriteString("\n")

	// Footer text based on mode
	switch m.settingsMode {
	case "config-view":
		content.WriteString(lipgloss.NewStyle().Italic(true).Foreground(colorGray).Render("Press Enter or Esc to return to menu"))
	case "edit-setting":
		content.WriteString(lipgloss.NewStyle().Italic(true).Foreground(colorGray).Render("Press Esc to return to menu"))
	default:
		content.WriteString(lipgloss.NewStyle().Italic(true).Foreground(colorGray).Render("Navigate: ↑/↓  Select: Enter  Back: Esc"))
	}

	// Create popup style
	popupStyle := lipgloss.NewStyle().
		Foreground(fgLight).
		Padding(1, 2).
		Width(60).
		Align(lipgloss.Center, lipgloss.Top)

	styledPopup := popupStyle.Render(content.String())

	// Overlay on background
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, styledPopup)
}

func (m model) renderSubscriptionPopup(background string) string {
	var content strings.Builder

	// Title
	title := lipgloss.NewStyle().Bold(true).Foreground(colorBlue).Render("☁️  Azure Subscription Manager")
	content.WriteString(title)
	content.WriteString("\n\n")

	switch m.subscriptionMenuMode {
	case "loading":
		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorYellow).Render("Loading subscriptions..."))
		content.WriteString("\n\n")
		content.WriteString(lipgloss.NewStyle().Faint(true).Render("Please wait while we fetch your Azure subscriptions"))

	case "menu":
		// Show current subscription at the top
		if m.currentSubscription != nil {
			content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("Current Subscription:"))
			content.WriteString("\n")
			content.WriteString(lipgloss.NewStyle().Foreground(colorAqua).Render(fmt.Sprintf("🎯 %s", m.currentSubscription.Name)))
			content.WriteString("\n")
			content.WriteString(lipgloss.NewStyle().Faint(true).Render(fmt.Sprintf("   Tenant: %s", m.currentSubscription.TenantID)))
			content.WriteString("\n\n")
		}

		content.WriteString(lipgloss.NewStyle().Bold(true).Foreground(colorGreen).Render("Available Subscriptions:"))
		content.WriteString("\n\n")

		if len(m.availableSubscriptions) == 0 {
			content.WriteString(lipgloss.NewStyle().Foreground(colorYellow).Render("No subscriptions found"))
			content.WriteString("\n")
			content.WriteString(lipgloss.NewStyle().Faint(true).Render("Check your Azure CLI login status"))
		} else {
			for i, subscription := range m.availableSubscriptions {
				style := lipgloss.NewStyle().Foreground(fgMedium)
				if i == m.subscriptionMenuIndex {
					style = style.Foreground(fgLight).Bold(true)
				}
				prefix := "  "
				if i == m.subscriptionMenuIndex {
					prefix = "▶ "
				}

				// Highlight current subscription
				icon := "📋"
				if m.currentSubscription != nil && subscription.ID == m.currentSubscription.ID {
					icon = "✅"
					style = style.Foreground(colorGreen)
				}

				content.WriteString(style.Render(prefix + icon + " " + subscription.Name))
				content.WriteString("\n")
				if i == m.subscriptionMenuIndex {
					content.WriteString(style.Render(fmt.Sprintf("   Tenant: %s", subscription.TenantID)))
					content.WriteString("\n")
					content.WriteString(style.Render(fmt.Sprintf("   ID: %s", subscription.ID)))
					content.WriteString("\n")
				}
			}
		}
	}

	content.WriteString("\n\n")

	// Add status bar with contextual shortcuts
	shortcuts := "Navigate: ↑/↓  Select: Enter  Back: Esc"
	statusbarStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("4")).
		Foreground(lipgloss.Color("15")).
		Bold(true).
		Padding(0, 1).
		Width(58)

	content.WriteString(statusbarStyle.Render("Subscriptions: " + shortcuts))
	content.WriteString("\n")

	// Footer text based on mode
	switch m.subscriptionMenuMode {
	case "loading":
		content.WriteString(lipgloss.NewStyle().Italic(true).Foreground(colorGray).Render("Please wait..."))
	case "menu":
		content.WriteString(lipgloss.NewStyle().Italic(true).Foreground(colorGray).Render("Select a subscription to switch context"))
	}

	// Create popup style - clean, no borders or backgrounds
	popupStyle := lipgloss.NewStyle().
		Foreground(fgLight).
		Padding(1, 2).
		Width(70).
		Align(lipgloss.Center, lipgloss.Top)

	styledPopup := popupStyle.Render(content.String())

	// Overlay on background
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, styledPopup)
}

func (m model) renderResourcePanel(width, height int) string {
	switch m.activeView {
	case "container-details":
		return m.containerInstanceDetailsContent
	case "container-logs":
		return m.containerInstanceLogsContent
	case "keyvault-secrets":
		return m.keyVaultSecretsContent
	case "keyvault-secret-details":
		return m.keyVaultSecretDetailsContent
	case "storage-containers":
		return m.storageContainersContent
	case "storage-blobs":
		return m.storageBlobsContent
	case "storage-blob-details":
		return m.storageBlobDetailsContent
	}

	// Handle regular resource views
	if m.selectedResource == nil {
		return m.renderWelcomePanel(width, height)
	}

	// Check if enhanced dashboard loading is in progress
	if m.dashboardLoadingInProgress && m.activeView == "dashboard" {
		// Show dashboard loading progress
		progress := resourcedetails.DashboardLoadingProgress{
			CurrentOperation:       "Loading comprehensive dashboard...",
			TotalOperations:        5,
			CompletedOperations:    0,
			ProgressPercentage:     0.0,
			DataProgress:           make(map[string]resourcedetails.DataProgress),
			Errors:                 []string{},
			StartTime:              m.dashboardLoadingStartTime,
			EstimatedTimeRemaining: "Calculating...",
		}
		return tui.RenderDashboardLoadingProgress(progress)
	}

	// Check if enhanced dashboard data is loaded
	if m.dashboardData != nil && m.activeView == "dashboard" {
		return tui.RenderComprehensiveDashboard(m.selectedResource.Name, m.dashboardData)
	}

	// Original dashboard view
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

	// Actions Section for Key Vaults
	if resource.Type == "Microsoft.KeyVault/vaults" {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("🔑 Key Vault Management"))
		content.WriteString("\n")

		actionStyle := lipgloss.NewStyle().Foreground(colorBlue)
		content.WriteString(fmt.Sprintf("%s List Secrets\n", actionStyle.Render("[K]")))
		content.WriteString(fmt.Sprintf("%s Create Secret\n", actionStyle.Render("[Shift+K]")))
		content.WriteString(fmt.Sprintf("%s Delete Secret\n", actionStyle.Render("[Ctrl+D]")))

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

	// Actions Section for Storage Accounts
	if resource.Type == "Microsoft.Storage/storageAccounts" {
		content.WriteString("\n")
		content.WriteString(sectionStyle.Render("💾 Storage Management"))
		content.WriteString("\n")

		actionStyle := lipgloss.NewStyle().Foreground(colorBlue)
		content.WriteString(fmt.Sprintf("%s List Containers\n", actionStyle.Render("[T]")))
		content.WriteString(fmt.Sprintf("%s Create Container\n", actionStyle.Render("[Shift+T]")))
		content.WriteString(fmt.Sprintf("%s List Blobs\n", actionStyle.Render("[B]")))
		content.WriteString(fmt.Sprintf("%s Upload Blob\n", actionStyle.Render("[U]")))
		content.WriteString(fmt.Sprintf("%s Delete Storage Item\n", actionStyle.Render("[Ctrl+X]")))

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

// renderShortcutRow formats a keyboard shortcut row with proper alignment
func renderShortcutRow(shortcut, description string) string {
	if shortcut == "" {
		// For sub-items or descriptions without shortcuts
		return fmt.Sprintf("           %s", description)
	}

	// Format with proper padding for alignment
	shortcutStyle := lipgloss.NewStyle().Foreground(colorAqua).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(fgLight)

	// Ensure consistent spacing: shortcut gets 12 characters, description follows
	paddedShortcut := fmt.Sprintf("%-12s", shortcut)
	return fmt.Sprintf("%s %s",
		shortcutStyle.Render(paddedShortcut),
		descStyle.Render(description))
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

		// Search
		"/":      "Enter search mode",
		"Escape": "Exit search mode",
		"↑/↓":    "Navigate search results (in search mode)",

		// Resource Actions
		"s":       "Start resource (VMs, Containers)",
		"S":       "Stop resource (VMs, Containers)",
		"r":       "Restart resource (VMs, Containers)",
		"shift+d": "Enhanced dashboard with real data",
		"R":       "Refresh all data",

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
		"y": "List Deployments (AKS)",
		"n": "List Nodes (AKS)",
		"v": "List Services (AKS)",

		// Key Vault Management
		"K":       "List Secrets (Key Vault)",
		"shift+k": "Create Secret (Key Vault)",
		"ctrl+d":  "Delete Secret (Key Vault)",

		// Interface
		"?":   "Show/hide this help",
		"Esc": "Navigate back / Close dialogs",
		"q":   "Quit application",
	}
}

// Terraform Commands
func loadTerraformFoldersCmd() tea.Cmd {
	return func() tea.Msg {
		// Scan for folders with .tf files
		folders, err := scanTerraformProjects(".")
		if err != nil {
			return errorMsg{error: fmt.Sprintf("Failed to load Terraform folders: %v", err)}
		}
		return terraformFoldersLoadedMsg{folders: folders}
	}
}

func scanTerraformProjects(rootDir string) ([]string, error) {
	var folders []string

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if strings.HasSuffix(path, ".tf") {
			dir := filepath.Dir(path)
			// Check if we already added this directory
			found := false
			for _, folder := range folders {
				if folder == dir {
					found = true
					break
				}
			}
			if !found {
				folders = append(folders, dir)
			}
		}
		return nil
	})

	return folders, err
}

func analyzeTerraformCodeCmd(folderPath string) tea.Cmd {
	return func() tea.Msg {
		analysis, err := analyzeTerraformCode(folderPath)
		if err != nil {
			return errorMsg{error: fmt.Sprintf("Failed to analyze Terraform code: %v", err)}
		}
		return terraformAnalysisMsg{analysis: analysis, path: folderPath}
	}
}

func analyzeTerraformCode(folderPath string) (string, error) {
	// Start with basic project analysis
	analysis := fmt.Sprintf("📁 Terraform Project Analysis: %s\n\n", folderPath)

	// Check for main files and gather content
	files := []string{"main.tf", "variables.tf", "outputs.tf", "terraform.tf"}
	foundFiles := make(map[string]bool)
	terraformContent := strings.Builder{}

	for _, file := range files {
		filePath := filepath.Join(folderPath, file)
		if _, err := os.Stat(filePath); err == nil {
			analysis += fmt.Sprintf("✅ %s found\n", file)
			foundFiles[file] = true

			// Read file content for AI analysis
			if content, readErr := os.ReadFile(filePath); readErr == nil {
				terraformContent.WriteString(fmt.Sprintf("\n--- %s ---\n", file))
				// Limit content size for AI analysis (first 2000 chars per file)
				contentStr := string(content)
				if len(contentStr) > 2000 {
					contentStr = contentStr[:2000] + "... (truncated)"
				}
				terraformContent.WriteString(contentStr)
			}
		} else {
			analysis += fmt.Sprintf("❌ %s missing\n", file)
		}
	}

	// Try to get AI-powered analysis if available
	aiProvider := openai.NewAIProviderAuto()
	if aiProvider != nil && terraformContent.Len() > 0 {
		analysis += "\n🤖 AI Analysis:\n"
		analysis += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"

		prompt := fmt.Sprintf(`As a senior Azure infrastructure expert, analyze this Terraform project and provide:

1. **Code Quality Assessment:**
   - Syntax and structure quality
   - Terraform best practices compliance
   - Resource naming conventions

2. **Security Analysis:**
   - Security vulnerabilities or risks
   - Missing security configurations
   - Recommended security improvements

3. **Azure-Specific Recommendations:**
   - Optimal Azure resource configurations
   - Cost optimization opportunities
   - Performance and scalability considerations

4. **Best Practices & Improvements:**
   - Missing essential configurations
   - State management recommendations
   - Testing and validation suggestions

5. **Next Steps:**
   - Prioritized action items
   - Quick wins for improvement

Project Files:
%s

Provide specific, actionable recommendations focused on Azure cloud best practices. Keep analysis professional and concise.`, terraformContent.String())

		aiAnalysis, err := aiProvider.Ask(prompt, "Azure Terraform Expert Analysis")
		if err != nil {
			analysis += fmt.Sprintf("⚠️ AI analysis unavailable: %v\n", err)
		} else {
			analysis += aiAnalysis + "\n"
		}
		analysis += "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"
	} else if aiProvider == nil {
		analysis += "\n💡 For enhanced AI analysis, set OPENAI_API_KEY or GITHUB_TOKEN environment variable.\n"
	}

	// Enhanced project health assessment
	analysis += "\n📊 Project Health Assessment:\n"
	score := 0
	total := 4
	for _, found := range foundFiles {
		if found {
			score++
		}
	}

	// Detailed file assessment
	analysis += "Files Status:\n"
	for _, file := range files {
		if foundFiles[file] {
			analysis += fmt.Sprintf("  ✅ %s - Present\n", file)
		} else {
			analysis += fmt.Sprintf("  ❌ %s - Missing\n", file)
		}
	}

	analysis += fmt.Sprintf("\nOverall Completeness: %d/%d files (%d%%)\n", score, total, (score*100)/total)

	// Enhanced health recommendations
	if score >= 3 {
		analysis += "✅ Good project structure - Ready for deployment\n"
		if score == 4 {
			analysis += "🎉 Complete Terraform project structure!\n"
		}
	} else if score >= 2 {
		analysis += "⚠️ Basic project structure - Consider adding missing files:\n"
		if !foundFiles["variables.tf"] {
			analysis += "  • variables.tf: Define input variables for flexibility\n"
		}
		if !foundFiles["outputs.tf"] {
			analysis += "  • outputs.tf: Export important resource information\n"
		}
		if !foundFiles["terraform.tf"] {
			analysis += "  • terraform.tf: Define provider requirements and versions\n"
		}
	} else {
		analysis += "❌ Incomplete project structure - Missing critical files\n"
		analysis += "  💡 Consider using 'Create from Template' to generate a complete structure\n"
	}

	// Additional checks for common files
	additionalFiles := []string{"README.md", "terraform.tfvars.example", ".gitignore"}
	foundAdditional := false
	for _, file := range additionalFiles {
		if _, err := os.Stat(filepath.Join(folderPath, file)); err == nil {
			if !foundAdditional {
				analysis += "\n📋 Additional Files Found:\n"
				foundAdditional = true
			}
			analysis += fmt.Sprintf("  ✅ %s\n", file)
		}
	}

	analysis += "\n🔧 Available Actions:\n"
	analysis += "  • Terraform Operations: validate, plan, apply, destroy\n"
	analysis += "  • Open External Editor: Edit files in VS Code/vim\n"
	analysis += "  • Create from Template: Generate missing structure\n"

	return analysis, nil
}

func executeTerraformOperationCmd(operation string, workspacePath string) tea.Cmd {
	return func() tea.Msg {
		var result string
		var err error

		// Ensure we're in the correct directory
		if _, statErr := os.Stat(workspacePath); os.IsNotExist(statErr) {
			return terraformOperationMsg{
				operation: operation,
				result:    fmt.Sprintf("Directory not found: %s", workspacePath),
				success:   false,
			}
		}

		switch operation {
		case "init":
			// Use the terraform package function
			err = terraform.InitWorkspace(workspacePath)
			if err == nil {
				result = "✅ Terraform initialized successfully"
			}
		case "plan":
			// Use the terraform package function
			result, err = terraform.PlanWorkspace(workspacePath, "")
			if err == nil && result != "" {
				result = "✅ Terraform plan completed successfully:\n" + result
			}
		case "apply":
			// Use the terraform package function
			result, err = terraform.ApplyWorkspace(workspacePath, "", true)
			if err == nil && result != "" {
				result = "✅ Terraform apply completed successfully:\n" + result
			}
		case "destroy":
			// Use the terraform package function
			result, err = terraform.DestroyWorkspace(workspacePath, "", true)
			if err == nil && result != "" {
				result = "✅ Terraform destroy completed successfully:\n" + result
			}
		case "validate":
			// Use the enhanced tfbicep package for better error handling
			valid, issues, validationErr := tfbicep.ValidateTerraformConfig(workspacePath)
			if validationErr != nil {
				err = validationErr
			} else if !valid {
				result = "❌ Terraform validation failed:\n" + strings.Join(issues, "\n")
				err = fmt.Errorf("validation failed")
			} else {
				result = "✅ Terraform configuration is valid"
			}
		case "format":
			// Use the enhanced tfbicep package for formatting
			formatErr := tfbicep.FormatTerraformFiles(workspacePath)
			if formatErr != nil {
				err = formatErr
			} else {
				result = "✅ Terraform files formatted successfully"
			}
		case "show":
			// Show current state or plan
			op, showErr := tfbicep.TerraformShow(workspacePath)
			if showErr != nil {
				err = showErr
			} else {
				result = "📋 Terraform show output:\n" + op.Output
			}
		case "state":
			// Show state list
			op, stateErr := tfbicep.TerraformState(workspacePath, "list", []string{})
			if stateErr != nil {
				err = stateErr
			} else {
				result = "📊 Terraform state resources:\n" + op.Output
			}
		default:
			err = fmt.Errorf("unknown operation: %s", operation)
		}

		success := err == nil
		if err != nil {
			result = fmt.Sprintf("❌ %s failed: %v", operation, err)
		}

		return terraformOperationMsg{
			operation: operation,
			result:    result,
			success:   success,
		}
	}
}

func openTerraformEditorCmd(folderPath string) tea.Cmd {
	return func() tea.Msg {
		// Try to open with VS Code first, then fall back to other editors
		editors := []string{"code", "vim", "nvim", "nano"}

		for _, editor := range editors {
			cmd := exec.Command(editor, folderPath)
			if err := cmd.Start(); err == nil {
				// Successfully started editor
				return terraformOperationMsg{
					operation: "editor",
					result:    fmt.Sprintf("Opened %s in %s", folderPath, editor),
					success:   true,
				}
			}
		}

		return errorMsg{error: "No suitable editor found (tried: code, vim, nvim, nano)"}
	}
}

func createFromTemplateCmd(folderPath string) tea.Cmd {
	return func() tea.Msg {
		// Use the existing template system to create a project from a template
		templatesPath := "./terraform/templates"

		// Check if templates directory exists
		if _, err := os.Stat(templatesPath); os.IsNotExist(err) {
			return terraformOperationMsg{
				operation: "template",
				result:    "❌ Templates directory not found. Please ensure terraform/templates exists.",
				success:   false,
			}
		}

		// For now, use the linux-vm template as default (can be enhanced to show selection menu)
		templateSource := filepath.Join(templatesPath, "vm", "linux-vm")

		// Check if the specific template exists
		if _, err := os.Stat(templateSource); os.IsNotExist(err) {
			// Fallback to creating a basic template
			return createBasicTemplate(folderPath)
		}

		// Copy template files to target directory
		err := copyTemplateFiles(templateSource, folderPath)
		if err != nil {
			return terraformOperationMsg{
				operation: "template",
				result:    fmt.Sprintf("❌ Failed to create template: %v", err),
				success:   false,
			}
		}

		return terraformOperationMsg{
			operation: "template",
			result:    fmt.Sprintf("✅ Linux VM template created successfully in %s\n\nFiles created:\n- main.tf\n- variables.tf\n- outputs.tf\n- terraform.tf\n\nNext steps:\n1. Customize variables in variables.tf\n2. Run 'terraform init'\n3. Run 'terraform plan'\n4. Run 'terraform apply'", folderPath),
			success:   true,
		}
	}
}

// createBasicTemplate creates a basic template when no predefined templates are available
func createBasicTemplate(folderPath string) terraformOperationMsg {
	templateContent := `# Basic Azure Resource Group and Storage Account Template
# Generated by Azure TUI

terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location

  tags = {
    Environment = var.environment
    Project     = var.project_name
  }
}

resource "azurerm_storage_account" "main" {
  name                     = var.storage_account_name
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "LRS"

  tags = {
    Environment = var.environment
    Project     = var.project_name
  }
}
`

	variablesContent := `# Variables for basic Azure template

variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "azure-tui-rg"
}

variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "East US"
}

variable "storage_account_name" {
  description = "Name of the storage account (must be globally unique)"
  type        = string
  default     = "azuretui${random_integer.suffix.result}"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "dev"
}

variable "project_name" {
  description = "Project name for tagging"
  type        = string
  default     = "azure-tui-project"
}

resource "random_integer" "suffix" {
  min = 1000
  max = 9999
}
`

	outputsContent := `# Outputs for basic Azure template

output "resource_group_name" {
  description = "Name of the created resource group"
  value       = azurerm_resource_group.main.name
}

output "resource_group_id" {
  description = "ID of the created resource group"
  value       = azurerm_resource_group.main.id
}

output "storage_account_name" {
  description = "Name of the created storage account"
  value       = azurerm_storage_account.main.name
}

output "storage_account_primary_endpoint" {
  description = "Primary blob endpoint of the storage account"
  value       = azurerm_storage_account.main.primary_blob_endpoint
}
`

	// Create main.tf
	if err := os.WriteFile(filepath.Join(folderPath, "main.tf"), []byte(templateContent), 0644); err != nil {
		return terraformOperationMsg{
			operation: "template",
			result:    fmt.Sprintf("❌ Failed to create main.tf: %v", err),
			success:   false,
		}
	}

	// Create variables.tf
	if err := os.WriteFile(filepath.Join(folderPath, "variables.tf"), []byte(variablesContent), 0644); err != nil {
		return terraformOperationMsg{
			operation: "template",
			result:    fmt.Sprintf("❌ Failed to create variables.tf: %v", err),
			success:   false,
		}
	}

	// Create outputs.tf
	if err := os.WriteFile(filepath.Join(folderPath, "outputs.tf"), []byte(outputsContent), 0644); err != nil {
		return terraformOperationMsg{
			operation: "template",
			result:    fmt.Sprintf("❌ Failed to create outputs.tf: %v", err),
			success:   false,
		}
	}

	return terraformOperationMsg{
		operation: "template",
		result:    fmt.Sprintf("✅ Basic template created successfully in %s\n\nFiles created:\n- main.tf (Resource Group & Storage Account)\n- variables.tf (Configurable variables)\n- outputs.tf (Resource outputs)\n\nNext steps:\n1. Customize variables in variables.tf\n2. Run 'terraform init'\n3. Run 'terraform plan'\n4. Run 'terraform apply'", folderPath),
		success:   true,
	}
}

// copyTemplateFiles copies template files from source to destination
func copyTemplateFiles(src, dst string) error {
	// Ensure destination directory exists
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Skip subdirectories for now (can be enhanced later)
			continue
		}

		// Only copy .tf files and README
		if strings.HasSuffix(entry.Name(), ".tf") || entry.Name() == "README.md" {
			content, err := os.ReadFile(srcPath)
			if err != nil {
				return fmt.Errorf("failed to read %s: %v", srcPath, err)
			}

			if err := os.WriteFile(dstPath, content, 0644); err != nil {
				return fmt.Errorf("failed to write %s: %v", dstPath, err)
			}
		}
	}

	return nil
}

// =============================================================================
// SETTINGS MANAGEMENT COMMANDS
// =============================================================================

// loadSettingsConfigCmd loads the current configuration
func loadSettingsConfigCmd() tea.Cmd {
	return func() tea.Msg {
		cfg, err := config.LoadConfig()
		if err != nil {
			return settingsConfigLoadedMsg{
				config:  nil,
				content: fmt.Sprintf("Error loading config: %v", err),
			}
		}

		// Format config content for display
		content := fmt.Sprintf("Current Configuration:\n\n")
		content += fmt.Sprintf("🔧 Terraform:\n")
		content += fmt.Sprintf("  Workspace Path: %s\n", cfg.Terraform.WorkspacePath)
		content += fmt.Sprintf("  Templates Path: %s\n", cfg.Terraform.TemplatesPath)
		content += fmt.Sprintf("  Auto Format: %t\n", cfg.Terraform.AutoFormat)
		content += fmt.Sprintf("\n🎨 UI:\n")
		content += fmt.Sprintf("  Show Terraform Menu: %t\n", cfg.UI.ShowTerraformMenu)
		content += fmt.Sprintf("  Popup Width: %d\n", cfg.UI.PopupWidth)
		content += fmt.Sprintf("  Popup Height: %d\n", cfg.UI.PopupHeight)
		content += fmt.Sprintf("  Enable Mouse Support: %t\n", cfg.UI.EnableMouseSupport)
		content += fmt.Sprintf("\n📝 Editor:\n")
		content += fmt.Sprintf("  Default Editor: %s\n", cfg.Editor.DefaultEditor)
		content += fmt.Sprintf("  Temp Directory: %s\n", cfg.Editor.TempDir)

		return settingsConfigLoadedMsg{
			config:  cfg,
			content: content,
		}
	}
}

// loadSettingsFoldersCmd loads available configuration paths for folder browser
func loadSettingsFoldersCmd() tea.Cmd {
	return func() tea.Msg {
		folders := []string{
			"~/.config/azure-tui",
			"./terraform",
			"./config",
			"/tmp",
			"~/Documents",
		}

		// Expand home directory
		homeDir, err := os.UserHomeDir()
		if err == nil {
			for i, folder := range folders {
				if strings.HasPrefix(folder, "~/") {
					folders[i] = strings.Replace(folder, "~", homeDir, 1)
				}
			}
		}

		return settingsFoldersLoadedMsg{
			folders: folders,
		}
	}
}

// saveSettingsConfigCmd saves the current configuration
func saveSettingsConfigCmd(cfg *config.AppConfig) tea.Cmd {
	return func() tea.Msg {
		err := config.SaveConfig(cfg)
		if err != nil {
			return settingsConfigSavedMsg{
				success: false,
				message: fmt.Sprintf("Failed to save config: %v", err),
			}
		}

		return settingsConfigSavedMsg{
			success: true,
			message: "Configuration saved successfully",
		}
	}
}

func getCurrentSubscription() (*Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "account", "show", "--output", "json")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current subscription: %v", err)
	}

	var azSub struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		TenantID  string `json:"tenantId"`
		IsDefault bool   `json:"isDefault"`
	}

	if err := json.Unmarshal(output, &azSub); err != nil {
		return nil, fmt.Errorf("failed to parse current subscription data: %v", err)
	}

	return &Subscription{
		ID:        azSub.ID,
		Name:      azSub.Name,
		TenantID:  azSub.TenantID,
		IsDefault: azSub.IsDefault,
	}, nil
}

func setCurrentSubscription(subscriptionID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "account", "set", "--subscription", subscriptionID)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to set current subscription: %v", err)
	}

	return nil
}

func getCurrentSubscriptionCmd() tea.Cmd {
	return func() tea.Msg {
		sub, err := getCurrentSubscription()
		if err != nil {
			return errorMsg{error: err.Error()}
		}
		return currentSubscriptionMsg{subscription: sub}
	}
}

func loadSubscriptionMenuCmd() tea.Cmd {
	return func() tea.Msg {
		subs, err := fetchSubscriptions()
		if err != nil {
			return errorMsg{error: err.Error()}
		}
		return subscriptionMenuMsg{subscriptions: subs}
	}
}

func selectSubscriptionCmd(subscriptionID string) tea.Cmd {
	return func() tea.Msg {
		err := setCurrentSubscription(subscriptionID)
		if err != nil {
			return subscriptionSelectedMsg{
				success: false,
				message: fmt.Sprintf("Failed to switch subscription: %v", err),
			}
		}

		// Get the updated current subscription
		sub, err := getCurrentSubscription()
		if err != nil {
			return subscriptionSelectedMsg{
				success: false,
				message: fmt.Sprintf("Subscription changed but failed to get updated info: %v", err),
			}
		}

		return subscriptionSelectedMsg{
			subscription: *sub,
			success:      true,
			message:      fmt.Sprintf("Successfully switched to subscription: %s", sub.Name),
		}
	}
}

func main() {
	m := initModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting Azure Dashboard: %v\n", err)
	}
}
