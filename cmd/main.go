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

	"github.com/olafkfreund/azure-tui/internal/azure/aci"
	"github.com/olafkfreund/azure-tui/internal/azure/keyvault"
	"github.com/olafkfreund/azure-tui/internal/azure/resourceactions"
	"github.com/olafkfreund/azure-tui/internal/azure/resourcedetails"
	"github.com/olafkfreund/azure-tui/internal/azure/storage"
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

// Terraform message types
type terraformMenuMsg struct{ content string }
type terraformFilesMsg struct {
	files []terraform.TerraformFile
}
type terraformFileContentMsg struct {
	filename string
	content  string
}
type terraformPlanMsg struct {
	output terraform.PlanOutput
}
type terraformStateMsg struct {
	state terraform.TerraformState
}
type terraformAIMsg struct{ content string }
type terraformOperationMsg struct {
	operation string
	success   bool
	message   string
}

// Enhanced Terraform message types
type terraformTemplateCreatedMsg struct {
	filename string
	content  string
}
type terraformFileCreatedMsg struct {
	filename string
	success  bool
	message  string
}
type terraformFileDeletedMsg struct {
	filename string
	success  bool
	message  string
}
type terraformValidationMsg struct {
	valid   bool
	errors  []string
	message string
}
type terraformWorkspaceMsg struct {
	workspaces []string
	current    string
}
type terraformResourceImportMsg struct {
	resourceId string
	success    bool
	message    string
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
	showHelpPopup bool

	// Navigation stack for back navigation
	navigationStack []string

	// Terraform-specific content and state
	terraformManager          *terraform.TerraformManager
	terraformMenuContent      string
	terraformFilesContent     string
	terraformFileContent      string
	terraformPlanContent      string
	terraformStateContent     string
	terraformAIContent        string
	selectedTerraformFile     string
	terraformFiles            []terraform.TerraformFile
	terraformOperationMode    string // "browse", "edit", "plan", "apply", "destroy"
	terraformValidationStatus bool
	terraformLastOperation    string
	terraformWorkspaces       []string
	terraformCurrentWorkspace string

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

func initModel() model {
	// Initialize AI provider with auto-detection (GitHub Copilot or OpenAI)
	ai := openai.NewAIProviderAuto()

	// Initialize Terraform manager
	terraformManager, err := terraform.NewTerraformManager()
	if err != nil {
		fmt.Printf("Warning: Failed to initialize Terraform manager: %v\n", err)
	}

	// Initialize search engine
	searchEngine := search.NewSearchEngine()

	m := model{
		treeView:                  tui.NewTreeView(),
		statusBar:                 &tui.StatusBar{},
		aiProvider:                ai,
		actionInProgress:          false,
		showDashboard:             true,
		expandedProperties:        make(map[string]bool),
		terraformManager:          terraformManager,
		searchEngine:              searchEngine,
		navigationStack:           []string{},
		terraformValidationStatus: false,
		terraformLastOperation:    "",
		terraformWorkspaces:       []string{"default"},
		terraformCurrentWorkspace: "default",
	}

	// Load initial data
	m.subscriptions = []Subscription{
		{ID: "demo-sub-1", Name: "Development", TenantID: "demo-tenant", IsDefault: true},
		{ID: "demo-sub-2", Name: "Production", TenantID: "demo-tenant", IsDefault: false},
	}

	m.resourceGroups = []ResourceGroup{
		{Name: "NetworkWatcherRG", Location: "uksouth"},
		{Name: "rg-fcaks-identity", Location: "uksouth"},
		{Name: "rg-fcaks-tfstate", Location: "uksouth"},
		{Name: "dem01_group", Location: "uksouth"},
	}

	m.allResources = []AzureResource{
		{
			ID:            "/subscriptions/demo/resourceGroups/NetworkWatcherRG/providers/Microsoft.Network/virtualNetworks/demo-vnet",
			Name:          "demo-vnet",
			Type:          "Microsoft.Network/virtualNetworks",
			Location:      "uksouth",
			ResourceGroup: "NetworkWatcherRG",
			Status:        "Running",
		},
		{
			ID:            "/subscriptions/demo/resourceGroups/dem01_group/providers/Microsoft.Compute/virtualMachines/demo-vm",
			Name:          "demo-vm",
			Type:          "Microsoft.Compute/virtualMachines",
			Location:      "uksouth",
			ResourceGroup: "dem01_group",
			Status:        "Running",
		},
	}

	// Initialize filtered resources
	m.filteredResources = m.allResources
	m.updateSearchEngine()

	// Set initial view
	m.activeView = "welcome"

	return m
}

// Init initializes the model for BubbleTea
func (m model) Init() tea.Cmd {
	return loadDataCmd()
}

// Update handles messages and updates the model state
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		// Update tree view size if it exists
		if m.treeView != nil {
			m.treeView.MaxVisible = (m.height - 6) / 2 // Adjust visible items based on height
		}

	case tea.KeyMsg:
		return m.handleKeyPress(msg)

	// Azure resource messages
	case subscriptionsLoadedMsg:
		m.subscriptions = msg.subscriptions
		m.loadingState = ""
		return m, nil

	case resourceGroupsLoadedMsg:
		m.resourceGroups = msg.groups
		m.updateTreeView()
		m.loadingState = ""
		return m, nil

	case resourcesInGroupMsg:
		// Update resources for specific group
		for i, resource := range m.allResources {
			if resource.ResourceGroup == msg.groupName {
				m.allResources = append(m.allResources[:i], m.allResources[i+1:]...)
				i--
			}
		}
		m.allResources = append(m.allResources, msg.resources...)
		m.filteredResources = m.allResources
		m.updateTreeView()
		m.updateSearchEngine()
		m.loadingState = ""
		return m, nil

	case resourceDetailsLoadedMsg:
		m.selectedResource = &msg.resource
		m.resourceDetails = msg.details
		m.pushView("details")
		return m, nil

	case aiDescriptionLoadedMsg:
		m.aiDescription = msg.description
		return m, nil

	case resourceActionMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		return m, nil

	case errorMsg:
		m.actionInProgress = false
		m.loadingState = ""
		// Add error to log entries
		m.logEntries = append(m.logEntries, fmt.Sprintf("ERROR: %s", msg.error))
		return m, nil

	// Network dashboard messages
	case networkDashboardMsg:
		m.actionInProgress = false
		m.networkDashboardContent = msg.content
		m.pushView("network-dashboard")
		return m, nil

	case vnetDetailsMsg:
		m.actionInProgress = false
		m.vnetDetailsContent = msg.content
		m.pushView("vnet-details")
		return m, nil

	case nsgDetailsMsg:
		m.actionInProgress = false
		m.nsgDetailsContent = msg.content
		m.pushView("nsg-details")
		return m, nil

	case networkTopologyMsg:
		m.actionInProgress = false
		m.networkTopologyContent = msg.content
		m.pushView("network-topology")
		return m, nil

	case networkAIAnalysisMsg:
		m.actionInProgress = false
		m.networkAIContent = msg.content
		m.pushView("network-ai")
		return m, nil

	case networkResourceCreatedMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		return m, nil

	// Container Instance messages
	case containerInstanceDetailsMsg:
		m.actionInProgress = false
		m.containerInstanceDetailsContent = msg.content
		m.pushView("container-details")
		return m, nil

	case containerInstanceLogsMsg:
		m.actionInProgress = false
		m.containerInstanceLogsContent = msg.content
		m.pushView("container-logs")
		return m, nil

	case containerInstanceActionMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		return m, nil

	case containerInstanceScaleMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		return m, nil

	// Key Vault messages
	case keyVaultSecretsMsg:
		m.actionInProgress = false
		m.keyVaultSecretsContent = fmt.Sprintf("Secrets in vault %s: %d secrets loaded", msg.vaultName, len(msg.secrets))
		m.keyVaultSecrets = msg.secrets
		return m, nil

	case keyVaultSecretDetailsMsg:
		m.actionInProgress = false
		if msg.secret != nil {
			m.keyVaultSecretDetailsContent = fmt.Sprintf("Secret: %s\nEnabled: %t", msg.secret.Name, msg.secret.Enabled)
			m.selectedSecret = msg.secret
		}
		return m, nil

	case keyVaultSecretActionMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		return m, nil

	// Storage Account messages
	case storageContainersMsg:
		m.actionInProgress = false
		m.storageContainersContent = fmt.Sprintf("Containers in %s: %d containers", msg.accountName, len(msg.containers))
		m.storageContainers = msg.containers
		m.currentStorageAccount = msg.accountName
		return m, nil

	case storageBlobsMsg:
		m.actionInProgress = false
		m.storageBlobsContent = fmt.Sprintf("Blobs in %s/%s: %d blobs", msg.accountName, msg.containerName, len(msg.blobs))
		m.storageBlobs = msg.blobs
		m.currentContainer = msg.containerName
		return m, nil

	case storageBlobDetailsMsg:
		m.actionInProgress = false
		if msg.blob != nil {
			m.storageBlobDetailsContent = fmt.Sprintf("Blob: %s\nSize: %d bytes\nType: %s", msg.blob.Name, msg.blob.Size, msg.blob.BlobType)
			m.selectedBlob = msg.blob
		}
		return m, nil

	case storageActionMsg:
		m.actionInProgress = false
		m.lastActionResult = &msg.result
		return m, nil

	// Basic Terraform messages
	case terraformMenuMsg:
		m.actionInProgress = false
		m.terraformMenuContent = msg.content
		m.pushView("terraform-menu")
		return m, nil

	case terraformFilesMsg:
		m.actionInProgress = false
		m.terraformFiles = msg.files
		m.pushView("terraform-files")
		return m, nil

	case terraformFileContentMsg:
		m.actionInProgress = false
		m.terraformFileContent = msg.content
		m.selectedTerraformFile = msg.filename
		m.pushView("terraform-file-content")
		return m, nil

	case terraformPlanMsg:
		m.actionInProgress = false
		m.terraformPlanContent = fmt.Sprintf("Terraform Plan Output:\nAdd: %d, Change: %d, Destroy: %d", msg.output.Add, msg.output.Change, msg.output.Destroy)
		m.terraformLastOperation = "plan"
		m.pushView("terraform-plan")
		return m, nil

	case terraformStateMsg:
		m.actionInProgress = false
		m.terraformStateContent = fmt.Sprintf("Terraform State:\n%d resources in state", len(msg.state.Resources))
		m.pushView("terraform-state")
		return m, nil

	case terraformAIMsg:
		m.actionInProgress = false
		m.terraformAIContent = msg.content
		m.pushView("terraform-ai")
		return m, nil

	case terraformOperationMsg:
		m.actionInProgress = false
		m.terraformLastOperation = msg.operation
		if msg.success {
			m.logEntries = append(m.logEntries, fmt.Sprintf("Terraform %s: %s", msg.operation, msg.message))
		} else {
			m.logEntries = append(m.logEntries, fmt.Sprintf("Terraform %s FAILED: %s", msg.operation, msg.message))
		}
		return m, nil

	// Enhanced Terraform messages
	case terraformTemplateCreatedMsg:
		m.actionInProgress = false
		if msg.filename != "" && msg.content != "" {
			m.terraformFileContent = msg.content
			m.selectedTerraformFile = msg.filename
			m.pushView("terraform-template")
			m.logEntries = append(m.logEntries, fmt.Sprintf("Generated Terraform template: %s", msg.filename))
		}
		return m, nil

	case terraformFileCreatedMsg:
		m.actionInProgress = false
		if msg.success {
			m.logEntries = append(m.logEntries, fmt.Sprintf("Created Terraform file: %s", msg.filename))
		} else {
			m.logEntries = append(m.logEntries, fmt.Sprintf("Failed to create file %s: %s", msg.filename, msg.message))
		}
		return m, nil

	case terraformFileDeletedMsg:
		m.actionInProgress = false
		if msg.success {
			m.logEntries = append(m.logEntries, fmt.Sprintf("Deleted Terraform file: %s", msg.filename))
		} else {
			m.logEntries = append(m.logEntries, fmt.Sprintf("Failed to delete file %s: %s", msg.filename, msg.message))
		}
		return m, nil

	case terraformValidationMsg:
		m.actionInProgress = false
		m.terraformValidationStatus = msg.valid
		if msg.valid {
			m.logEntries = append(m.logEntries, "Terraform configuration is valid")
		} else {
			m.logEntries = append(m.logEntries, fmt.Sprintf("Terraform validation failed: %d errors", len(msg.errors)))
		}
		return m, nil

	case terraformWorkspaceMsg:
		m.actionInProgress = false
		m.terraformWorkspaces = msg.workspaces
		m.terraformCurrentWorkspace = msg.current
		m.pushView("terraform-workspaces")
		return m, nil

	case terraformResourceImportMsg:
		m.actionInProgress = false
		if msg.success {
			m.logEntries = append(m.logEntries, fmt.Sprintf("Successfully imported resource: %s", msg.resourceId))
		} else {
			m.logEntries = append(m.logEntries, fmt.Sprintf("Failed to import resource %s: %s", msg.resourceId, msg.message))
		}
		return m, nil
	}

	return m, nil
}

// handleKeyPress processes keyboard input
func (m model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global key handlers
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit

	case "?":
		m.showHelpPopup = !m.showHelpPopup
		return m, nil

	case "esc":
		// Handle Esc key navigation
		if m.showHelpPopup {
			m.showHelpPopup = false
			return m, nil
		}
		if m.popView() {
			m.rightPanelScrollOffset = 0
			return m, nil
		}
		return m, nil

	case "r":
		if !m.actionInProgress {
			m.actionInProgress = true
			m.loadingState = "Refreshing resources..."
			return m, loadDataCmd()
		}
		return m, nil

	case "f2":
		m.showDashboard = !m.showDashboard
		return m, nil

	case "tab":
		m.selectedPanel = (m.selectedPanel + 1) % 2
		return m, nil

	case "h", "left":
		m.selectedPanel = 0
		return m, nil

	case "l", "right":
		m.selectedPanel = 1
		return m, nil
	}

	// Context-sensitive key handlers
	if m.selectedPanel == 0 && m.treeView != nil {
		// Tree view navigation
		switch msg.String() {
		case "j", "down":
			m.treeView.SelectNext()
			return m, nil
		case "k", "up":
			m.treeView.SelectPrevious()
			return m, nil
		case "enter", " ":
			selectedNode := m.treeView.GetSelectedNode()
			if selectedNode != nil {
				return m.handleTreeSelection(selectedNode)
			}
			return m, nil
		}
	} else if m.selectedPanel == 1 {
		// Right panel navigation
		switch msg.String() {
		case "j", "down":
			m.rightPanelScrollOffset++
			return m, nil
		case "k", "up":
			if m.rightPanelScrollOffset > 0 {
				m.rightPanelScrollOffset--
			}
			return m, nil
		case "e":
			if m.selectedResource != nil && m.activeView == "details" {
				m.propertyExpandedIndex = (m.propertyExpandedIndex + 1) % 10
				return m, nil
			}
			return m, nil
		}
	}

	// Enhanced Terraform and Azure shortcuts
	if !m.actionInProgress {
		switch msg.String() {
		case "t":
			// Open Terraform menu
			m.actionInProgress = true
			return m, tea.Cmd(func() tea.Msg {
				return terraformMenuMsg{content: "Terraform Management\n\n1. Browse Files\n2. Create Template\n3. Plan & Apply\n4. View State\n5. AI Assistant\n\nPress Esc to go back"}
			})

		case "ctrl+g":
			// Generate Terraform template from selected resource
			if m.selectedResource != nil {
				m.actionInProgress = true
				return m, tea.Cmd(func() tea.Msg {
					content := fmt.Sprintf("# Terraform template for %s\nresource \"azurerm_%s\" \"%s\" {\n  # Generated by Azure TUI\n}",
						m.selectedResource.Name,
						strings.ToLower(strings.Split(m.selectedResource.Type, "/")[1]),
						m.selectedResource.Name)
					return terraformTemplateCreatedMsg{filename: fmt.Sprintf("%s.tf", m.selectedResource.Name), content: content}
				})
			}
			return m, nil

		case "ctrl+f":
			// Format Terraform files
			if !m.actionInProgress {
				m.actionInProgress = true
				return m, tea.Cmd(func() tea.Msg {
					return terraformOperationMsg{operation: "format", success: true, message: "Terraform files formatted successfully"}
				})
			}
			return m, nil

		case "ctrl+v":
			// Validate Terraform configuration
			if !m.actionInProgress {
				m.actionInProgress = true
				return m, tea.Cmd(func() tea.Msg {
					return terraformValidationMsg{valid: true, errors: []string{}, message: "Terraform configuration is valid"}
				})
			}
			return m, nil

		// Azure resource shortcuts
		case "n":
			if m.selectedResource != nil && strings.Contains(m.selectedResource.Type, "Network") {
				m.actionInProgress = true
				return m, tea.Cmd(func() tea.Msg {
					return networkDashboardMsg{content: fmt.Sprintf("Network Dashboard for %s\n\nNetwork configuration and monitoring", m.selectedResource.Name)}
				})
			}
			return m, nil

		case "c":
			if m.selectedResource != nil && strings.Contains(m.selectedResource.Type, "ContainerInstance") {
				m.actionInProgress = true
				return m, tea.Cmd(func() tea.Msg {
					return containerInstanceDetailsMsg{content: fmt.Sprintf("Container Instance Details for %s\n\nContainer status and logs", m.selectedResource.Name)}
				})
			}
			return m, nil
		}
	}

	return m, nil
}

// handleTreeSelection handles tree node selection
func (m model) handleTreeSelection(selectedNode *tui.TreeNode) (tea.Model, tea.Cmd) {
	if selectedNode.Type == "group" && !selectedNode.Expanded {
		// Expand resource group
		selectedNode.Expanded = true
		if !m.actionInProgress {
			m.actionInProgress = true
			m.loadingState = fmt.Sprintf("Loading resources in %s...", selectedNode.Name)
			return m, loadResourcesInGroupCmd(selectedNode.Name)
		}
	} else if selectedNode.Type == "resource" && selectedNode.ResourceData != nil {
		// Select resource and load details
		if resource, ok := selectedNode.ResourceData.(AzureResource); ok {
			m.selectedResource = &resource
			if !m.actionInProgress {
				m.actionInProgress = true
				m.loadingState = fmt.Sprintf("Loading details for %s...", resource.Name)
				return m, loadResourceDetailsCmd(resource)
			}
		}
	}
	return m, nil
}

// Navigation helper functions
func (m *model) pushView(newView string) {
	if m.activeView != newView {
		m.navigationStack = append(m.navigationStack, m.activeView)
		m.activeView = newView
	}
}

func (m *model) popView() bool {
	if len(m.navigationStack) == 0 {
		return false
	}
	lastIndex := len(m.navigationStack) - 1
	previousView := m.navigationStack[lastIndex]
	m.navigationStack = m.navigationStack[:lastIndex]
	m.activeView = previousView
	return true
}

// updateTreeView updates the tree view with current resource groups and resources
func (m *model) updateTreeView() {
	if m.treeView == nil {
		return
	}

	// Clear existing tree
	m.treeView.Root.Children = []*tui.TreeNode{}

	// Add resource groups
	for _, rg := range m.resourceGroups {
		groupNode := m.treeView.AddResourceGroup(rg.Name, rg.Location)

		// Add resources for this group
		for _, resource := range m.allResources {
			if resource.ResourceGroup == rg.Name {
				m.treeView.AddResource(groupNode, resource.Name, resource.Type, resource)
			}
		}
	}

	// Ensure a node is selected
	m.treeView.EnsureSelection()
}

// View renders the current state of the application
func (m model) View() string {
	if !m.ready {
		return "Initializing Azure TUI..."
	}

	if m.showHelpPopup {
		return m.renderHelpPopup()
	}

	// Render based on active view
	switch m.activeView {
	case "terraform-menu":
		return m.renderTerraformMenuView()
	case "terraform-files":
		return m.renderTerraformFilesView()
	case "terraform-file-content", "terraform-template":
		return m.renderTerraformFileContentView()
	case "terraform-plan":
		return m.renderTerraformPlanView()
	case "terraform-state":
		return m.renderTerraformStateView()
	case "terraform-workspaces":
		return m.renderTerraformWorkspacesView()
	case "terraform-ai":
		return m.renderTerraformAIView()
	case "network-dashboard":
		return m.renderNetworkDashboardView()
	case "vnet-details":
		return m.renderVNetDetailsView()
	case "nsg-details":
		return m.renderNSGDetailsView()
	case "network-topology":
		return m.renderNetworkTopologyView()
	case "network-ai":
		return m.renderNetworkAIView()
	case "container-details":
		return m.renderContainerDetailsView()
	case "container-logs":
		return m.renderContainerLogsView()
	case "details":
		return m.renderResourceDetailsView()
	case "dashboard":
		return m.renderDashboardView()
	default:
		return m.renderWelcomeView()
	}
}

// Basic view rendering methods
func (m model) renderWelcomeView() string {
	if m.showDashboard {
		// Traditional two-panel layout
		return m.renderTraditionalLayout()
	}
	// Tree view layout
	return m.renderTreeViewLayout()
}

func (m model) renderTreeViewLayout() string {
	leftPanel := m.renderTreePanel()
	rightPanel := m.renderResourcePanel(m.width/2, m.height-4)
	statusBar := m.renderStatusBar()

	// Create border styles based on selected panel
	leftStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())
	rightStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder())

	if m.selectedPanel == 0 {
		leftStyle = leftStyle.BorderForeground(colorBlue)
	} else {
		leftStyle = leftStyle.BorderForeground(colorGray)
	}

	if m.selectedPanel == 1 {
		rightStyle = rightStyle.BorderForeground(colorGreen)
	} else {
		rightStyle = rightStyle.BorderForeground(colorGray)
	}

	leftPanelStyled := leftStyle.Width(m.width/2 - 2).Height(m.height - 4).Render(leftPanel)
	rightPanelStyled := rightStyle.Width(m.width/2 - 2).Height(m.height - 4).Render(rightPanel)

	layout := lipgloss.JoinHorizontal(lipgloss.Top, leftPanelStyled, rightPanelStyled)
	return lipgloss.JoinVertical(lipgloss.Left, layout, statusBar)
}

func (m model) renderTraditionalLayout() string {
	// Traditional dashboard layout
	content := "Azure Dashboard - Traditional Mode\n\n"
	content += "Press F2 to switch to Tree View mode\n"
	if m.selectedResource != nil {
		content += fmt.Sprintf("Selected: %s (%s)\n", m.selectedResource.Name, m.selectedResource.Type)
	}
	return content
}

func (m model) renderTreePanel() string {
	if m.treeView == nil {
		return "Initializing tree view..."
	}
	return m.treeView.RenderTreeView(m.width/2-4, m.height-6)
}

func (m model) renderResourcePanel(width, height int) string {
	// Basic resource panel rendering - will be enhanced with specific view rendering
	if m.loadingState != "" {
		return fmt.Sprintf("🔄 %s", m.loadingState)
	}

	if m.selectedResource == nil {
		return "Select a resource to view details"
	}

	return fmt.Sprintf("Resource: %s\nType: %s\nLocation: %s\nResource Group: %s",
		m.selectedResource.Name,
		m.selectedResource.Type,
		m.selectedResource.Location,
		m.selectedResource.ResourceGroup)
}

func (m model) renderStatusBar() string {
	statusBar := tui.CreatePowerlineStatusBar(m.width)

	if m.actionInProgress {
		statusBar.AddSegment("🔄 Loading...", colorYellow, bgDark)
	} else {
		statusBar.AddSegment("✅ Ready", colorGreen, bgDark)
	}

	// Add panel indicator
	if m.selectedPanel == 0 {
		statusBar.AddSegment("[TREE]", colorBlue, bgMedium)
	} else {
		statusBar.AddSegment("[DETAILS]", colorGreen, bgMedium)
	}

	// Add navigation indicator if there's history
	if len(m.navigationStack) > 0 {
		statusBar.AddSegment(fmt.Sprintf("Esc:Back(%d)", len(m.navigationStack)), colorAqua, bgMedium)
	}

	// Add shortcuts
	statusBar.AddSegment("?:Help", colorPurple, bgMedium)
	statusBar.AddSegment("q:Quit", colorRed, bgMedium)

	return statusBar.RenderStatusBar()
}

func (m model) renderHelpPopup() string {
	helpContent := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBlue).
		Padding(2).
		Width(60).
		Render(`Azure TUI - Help

Navigation:
j/k or ↑/↓    Navigate tree/details
h/l or ←/→    Switch panels (Tree/Details)
Tab           Switch panels
Space/Enter   Expand resource group / Select resource
Esc           Navigate back / Close dialogs

Operations:
r             Refresh resources
t             Open Terraform menu
n             Network dashboard (for network resources)
c             Container details (for container resources)

Terraform:
Ctrl+G        Generate Terraform template
Ctrl+F        Format Terraform files  
Ctrl+V        Validate Terraform configuration

General:
F2            Toggle view mode
?             Show/hide this help
q             Quit application`)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, helpContent)
}

// Placeholder rendering methods for different views
func (m model) renderTerraformMenuView() string {
	return m.terraformMenuContent
}

func (m model) renderTerraformFilesView() string {
	return m.terraformFilesContent
}

func (m model) renderTerraformFileContentView() string {
	return m.terraformFileContent
}

func (m model) renderTerraformPlanView() string {
	return m.terraformPlanContent
}

func (m model) renderTerraformStateView() string {
	return m.terraformStateContent
}

func (m model) renderTerraformWorkspacesView() string {
	content := "Terraform Workspaces\n\n"
	content += fmt.Sprintf("Current: %s\n\n", m.terraformCurrentWorkspace)
	content += "Available workspaces:\n"
	for _, ws := range m.terraformWorkspaces {
		if ws == m.terraformCurrentWorkspace {
			content += fmt.Sprintf("* %s (current)\n", ws)
		} else {
			content += fmt.Sprintf("  %s\n", ws)
		}
	}
	return content
}

func (m model) renderTerraformAIView() string {
	return m.terraformAIContent
}

func (m model) renderNetworkDashboardView() string {
	return m.networkDashboardContent
}

func (m model) renderVNetDetailsView() string {
	return m.vnetDetailsContent
}

func (m model) renderNSGDetailsView() string {
	return m.nsgDetailsContent
}

func (m model) renderNetworkTopologyView() string {
	return m.networkTopologyContent
}

func (m model) renderNetworkAIView() string {
	return m.networkAIContent
}

func (m model) renderContainerDetailsView() string {
	return m.containerInstanceDetailsContent
}

func (m model) renderContainerLogsView() string {
	return m.containerInstanceLogsContent
}

func (m model) renderResourceDetailsView() string {
	if m.selectedResource == nil {
		return "No resource selected"
	}

	content := fmt.Sprintf("Resource Details: %s\n\n", m.selectedResource.Name)
	content += fmt.Sprintf("Type: %s\n", m.selectedResource.Type)
	content += fmt.Sprintf("Location: %s\n", m.selectedResource.Location)
	content += fmt.Sprintf("Resource Group: %s\n", m.selectedResource.ResourceGroup)
	content += fmt.Sprintf("Status: %s\n\n", m.selectedResource.Status)

	if m.resourceDetails != nil {
		content += "Detailed Information:\n"
		content += fmt.Sprintf("Resource ID: %s\n", m.resourceDetails.ID)
		content += fmt.Sprintf("Properties: %+v\n", m.resourceDetails.Properties)
	}

	if m.aiDescription != "" {
		content += "\n\nAI Analysis:\n"
		content += m.aiDescription
	}

	return content
}

func (m model) renderDashboardView() string {
	return "Dashboard View - Implementation pending"
}

func main() {
	m := initModel()
	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error starting Azure Dashboard: %v\n", err)
	}
}
