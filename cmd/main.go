package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/olafkfreund/azure-tui/internal/azure/azuresdk"
	"github.com/olafkfreund/azure-tui/internal/azure/resourceactions"
	"github.com/olafkfreund/azure-tui/internal/azure/resourcedetails"
	"github.com/olafkfreund/azure-tui/internal/tui"
)

// Azure SDK client for resource group listing
var azureClient *azuresdk.AzureClient

// Initialize Azure client lazily to avoid blocking startup
func getAzureClient() *azuresdk.AzureClient {
	if azureClient == nil {
		var err error
		azureClient, err = azuresdk.NewAzureClient()
		if err != nil {
			// Log error but don't panic - continue with Azure CLI fallback
			fmt.Printf("Warning: Failed to initialize Azure SDK client: %v\n", err)
		}
	}
	return azureClient
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

// Enhanced resource status and metadata
type ResourceStatus struct {
	Health     string    `json:"health"` // "Healthy", "Warning", "Critical", "Unknown"
	State      string    `json:"state"`  // "Running", "Stopped", "Starting", etc.
	LastUpdate time.Time `json:"lastUpdate"`
}

type EnhancedAzureResource struct {
	AzureResource
	Status       ResourceStatus         `json:"status"`
	Metadata     map[string]interface{} `json:"metadata"`
	Tags         map[string]string      `json:"tags"`
	Dependencies []string               `json:"dependencies"`
}

// Real-time Resource Operations Enhancement
type ResourceOperationManager struct {
	operationsInProgress map[string]*ResourceOperation
	bulkSelections       map[string]bool // resource ID -> selected
	batchOperationChan   chan BatchOperation
	mutex                sync.RWMutex
}

type ResourceOperation struct {
	ResourceID   string
	ResourceName string
	Operation    string
	Status       string // "pending", "running", "completed", "failed"
	StartTime    time.Time
	Progress     int // 0-100
	Output       string
	Error        error
}

type BatchOperation struct {
	Operation   string
	ResourceIDs []string
	Parameters  map[string]interface{}
	CallbackMsg tea.Msg
}

type ResourceOperationResult struct {
	ResourceID   string
	ResourceName string
	Success      bool
	Output       string
	Error        error
}

// Real-time resource status tracking
type ResourceStatusUpdate struct {
	ResourceID  string
	Status      string
	Health      string
	Metrics     map[string]float64
	LastUpdated time.Time
}

// Live resource expansion with caching
type ResourceExpansionCache struct {
	cache      map[string][]AzureResource // group name -> resources
	lastUpdate map[string]time.Time       // group name -> last update time
	ttl        time.Duration
	mutex      sync.RWMutex
}

// Message types for real-time operations
type resourceOperationStartedMsg struct {
	operationID string
	operation   *ResourceOperation
}

type resourceOperationProgressMsg struct {
	operationID string
	progress    int
	output      string
}

type resourceOperationCompletedMsg struct {
	operationID string
	success     bool
	output      string
	error       error
}

type resourceStatusUpdateMsg struct {
	resourceID string
	status     ResourceStatusUpdate
}

type bulkOperationStartedMsg struct {
	operationCount int
	operation      string
}

type bulkOperationProgressMsg struct {
	completed int
	total     int
	current   string
}

type bulkOperationCompletedMsg struct {
	successful int
	failed     int
	results    []ResourceOperationResult
}

type resourceStatusUpdatesMsg struct {
	updates []ResourceStatusUpdate
}

type resourceStatusMonitoringTickMsg struct {
	time time.Time
}

// Message types for UI communication
type resourcesLoadingMsg struct {
	groupName string
}

type resourcesInGroupErrMsg struct {
	groupName string
	error     string
}

type resourcesInGroupMsg struct {
	groupName string
	resources []AzureResource
}

// Main application model
type model struct {
	treeView          *tui.TreeView
	tabManager        *tui.TabManager
	statusBar         *tui.StatusBar
	width             int
	height            int
	ready             bool
	resourcesInGroup  []AzureResource
	operationManager  *ResourceOperationManager
	expansionCache    *ResourceExpansionCache
	bulkSelectionMode bool
	healthMonitor     *ResourceHealthMonitor
}

// Resource health monitoring
type ResourceHealthMonitor struct {
	resources map[string]*resourcedetails.ResourceDetails
	mutex     sync.RWMutex
}

func NewResourceHealthMonitor() *ResourceHealthMonitor {
	return &ResourceHealthMonitor{
		resources: make(map[string]*resourcedetails.ResourceDetails),
	}
}

func (rhm *ResourceHealthMonitor) UpdateResourceHealth(resourceID string) *EnhancedAzureResource {
	// This would typically fetch real health data
	// For now, return nil as a placeholder
	return nil
}

// Initialize resource operation manager
func NewResourceOperationManager() *ResourceOperationManager {
	return &ResourceOperationManager{
		operationsInProgress: make(map[string]*ResourceOperation),
		bulkSelections:       make(map[string]bool),
		batchOperationChan:   make(chan BatchOperation, 10),
	}
}

// Initialize resource expansion cache
func NewResourceExpansionCache(ttl time.Duration) *ResourceExpansionCache {
	return &ResourceExpansionCache{
		cache:      make(map[string][]AzureResource),
		lastUpdate: make(map[string]time.Time),
		ttl:        ttl,
	}
}

// Real-time resource expansion with smart caching
func (rec *ResourceExpansionCache) GetResources(groupName string, forceRefresh bool) ([]AzureResource, bool, error) {
	rec.mutex.RLock()
	cached, exists := rec.cache[groupName]
	lastUpdate, hasTime := rec.lastUpdate[groupName]
	rec.mutex.RUnlock()

	// Check if cache is valid
	if exists && hasTime && !forceRefresh && time.Since(lastUpdate) < rec.ttl {
		return cached, false, nil // returned from cache
	}

	// Load resources from Azure
	resources, err := fetchResourcesInGroupWithTimeout(groupName, 10*time.Second)
	if err != nil {
		// Return cached data if available, even if expired
		if exists {
			return cached, false, err
		}
		return nil, false, err
	}

	// Update cache
	rec.mutex.Lock()
	rec.cache[groupName] = resources
	rec.lastUpdate[groupName] = time.Now()
	rec.mutex.Unlock()

	return resources, true, nil // loaded fresh
}

// Execute resource operation with real-time feedback
func (rom *ResourceOperationManager) ExecuteResourceOperation(resourceID, resourceName, resourceType, operation string, params map[string]interface{}) tea.Cmd {
	operationID := fmt.Sprintf("%s-%s-%d", resourceID, operation, time.Now().Unix())

	rom.mutex.Lock()
	rom.operationsInProgress[operationID] = &ResourceOperation{
		ResourceID:   resourceID,
		ResourceName: resourceName,
		Operation:    operation,
		Status:       "pending",
		StartTime:    time.Now(),
		Progress:     0,
	}
	rom.mutex.Unlock()

	return func() tea.Msg {
		// Send operation started message
		tea.NewProgram(nil).Send(resourceOperationStartedMsg{
			operationID: operationID,
			operation:   rom.operationsInProgress[operationID],
		})

		// Extract resource group from resource ID
		resourceGroup := extractResourceGroupFromID(resourceID)

		// Update operation status
		rom.mutex.Lock()
		rom.operationsInProgress[operationID].Status = "running"
		rom.operationsInProgress[operationID].Progress = 25
		rom.mutex.Unlock()

		// Send progress update
		tea.NewProgram(nil).Send(resourceOperationProgressMsg{
			operationID: operationID,
			progress:    25,
			output:      fmt.Sprintf("Starting %s operation on %s...", operation, resourceName),
		})

		// Execute the actual operation
		result := resourceactions.ExecuteResourceAction(operation, resourceType, resourceName, resourceGroup, params)

		// Update final status
		rom.mutex.Lock()
		if result.Success {
			rom.operationsInProgress[operationID].Status = "completed"
			rom.operationsInProgress[operationID].Progress = 100
		} else {
			rom.operationsInProgress[operationID].Status = "failed"
		}
		rom.operationsInProgress[operationID].Output = result.Output
		rom.mutex.Unlock()

		// Send completion message
		return resourceOperationCompletedMsg{
			operationID: operationID,
			success:     result.Success,
			output:      result.Output,
			error:       nil,
		}
	}
}

// Execute bulk operations on multiple resources
func (rom *ResourceOperationManager) ExecuteBulkOperation(operation string, resourceIDs []string, params map[string]interface{}) tea.Cmd {
	return func() tea.Msg {
		// Send bulk operation started message
		tea.NewProgram(nil).Send(bulkOperationStartedMsg{
			operationCount: len(resourceIDs),
			operation:      operation,
		})

		results := make([]ResourceOperationResult, 0, len(resourceIDs))
		successful := 0
		failed := 0

		for i, resourceID := range resourceIDs {
			// Send progress update
			tea.NewProgram(nil).Send(bulkOperationProgressMsg{
				completed: i,
				total:     len(resourceIDs),
				current:   resourceID,
			})

			// Extract resource info
			resourceGroup := extractResourceGroupFromID(resourceID)
			resourceName := extractResourceNameFromID(resourceID)
			resourceType := extractResourceTypeFromID(resourceID)

			// Execute operation
			result := resourceactions.ExecuteResourceAction(operation, resourceType, resourceName, resourceGroup, params)

			opResult := ResourceOperationResult{
				ResourceID:   resourceID,
				ResourceName: resourceName,
				Success:      result.Success,
				Output:       result.Output,
			}

			if result.Success {
				successful++
			} else {
				failed++
				opResult.Error = fmt.Errorf("operation failed: %s", result.Message)
			}

			results = append(results, opResult)

			// Small delay between operations to avoid overwhelming Azure
			time.Sleep(1 * time.Second)
		}

		// Send completion message
		return bulkOperationCompletedMsg{
			successful: successful,
			failed:     failed,
			results:    results,
		}
	}
}

// Real-time resource status monitoring
func (m *model) startResourceStatusMonitoring() tea.Cmd {
	return tea.Tick(15*time.Second, func(t time.Time) tea.Msg {
		// Update status for currently visible resources
		updates := make([]ResourceStatusUpdate, 0)

		for _, resource := range m.resourcesInGroup {
			// Get enhanced resource status
			if enhanced := m.healthMonitor.UpdateResourceHealth(resource.ID); enhanced != nil {
				update := ResourceStatusUpdate{
					ResourceID:  resource.ID,
					Status:      enhanced.Status.State,
					Health:      enhanced.Status.Health,
					LastUpdated: time.Now(),
					Metrics:     make(map[string]float64),
				}

				// Extract metrics if available
				if cpuUsage, ok := enhanced.Metadata["cpuUsage"].(float64); ok {
					update.Metrics["cpu"] = cpuUsage
				}
				if memUsage, ok := enhanced.Metadata["memoryUsage"].(float64); ok {
					update.Metrics["memory"] = memUsage
				}

				updates = append(updates, update)
			}
		}

		if len(updates) > 0 {
			return resourceStatusUpdatesMsg{updates}
		}

		return resourceStatusMonitoringTickMsg{t}
	})
}

// Toggle bulk selection mode
func (m *model) toggleBulkSelection() {
	if m.operationManager == nil {
		m.operationManager = NewResourceOperationManager()
	}
	m.bulkSelectionMode = !m.bulkSelectionMode

	if !m.bulkSelectionMode {
		// Clear selections when exiting bulk mode
		m.operationManager.bulkSelections = make(map[string]bool)
	}
}

// Toggle resource selection for bulk operations
func (m *model) toggleResourceSelection(resourceID string) {
	if m.operationManager == nil {
		m.operationManager = NewResourceOperationManager()
	}

	m.operationManager.mutex.Lock()
	defer m.operationManager.mutex.Unlock()

	if m.operationManager.bulkSelections[resourceID] {
		delete(m.operationManager.bulkSelections, resourceID)
	} else {
		m.operationManager.bulkSelections[resourceID] = true
	}
}

// Get selected resources for bulk operations
func (m *model) getSelectedResourcesForBulk() []string {
	if m.operationManager == nil {
		return []string{}
	}

	m.operationManager.mutex.RLock()
	defer m.operationManager.mutex.RUnlock()

	selected := make([]string, 0, len(m.operationManager.bulkSelections))
	for resourceID := range m.operationManager.bulkSelections {
		selected = append(selected, resourceID)
	}
	return selected
}

// Enhanced resource loading with real-time feedback
func loadResourcesInGroupCmdEnhanced(groupName string, cache *ResourceExpansionCache) tea.Cmd {
	return tea.Batch(
		// Show loading state immediately
		func() tea.Msg {
			return resourcesLoadingMsg{groupName}
		},
		// Load resources with caching
		func() tea.Msg {
			resources, fromCache, err := cache.GetResources(groupName, false)
			if err != nil {
				return resourcesInGroupErrMsg{groupName, err.Error()}
			}

			// If loaded from cache, also trigger background refresh
			if fromCache {
				go func() {
					time.Sleep(500 * time.Millisecond) // Delay to show cached data first
					freshResources, _, refreshErr := cache.GetResources(groupName, true)
					if refreshErr == nil && len(freshResources) != len(resources) {
						// Data changed, send update
						tea.NewProgram(nil).Send(resourcesInGroupMsg{groupName, freshResources})
					}
				}()
			}

			return resourcesInGroupMsg{groupName, resources}
		},
	)
}

// Helper functions for extracting information from resource IDs
func extractResourceGroupFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i, part := range parts {
		if part == "resourceGroups" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func extractResourceNameFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}

func extractResourceTypeFromID(resourceID string) string {
	parts := strings.Split(resourceID, "/")
	for i := 0; i < len(parts)-1; i += 2 {
		if i+1 < len(parts) && parts[i] == "providers" && i+2 < len(parts) {
			return parts[i+1] + "/" + parts[i+2]
		}
	}
	return ""
}

// fetchResourcesInGroupWithTimeout fetches resources in a resource group with timeout
func fetchResourcesInGroupWithTimeout(groupName string, timeout time.Duration) ([]AzureResource, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "az", "resource", "list",
		"--resource-group", groupName,
		"--output", "json")

	output, err := cmd.Output()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("timeout fetching resources for group %s", groupName)
		}
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

// Enhanced real-time updates integration

// Initialize the model
func initModel() model {
	treeView := tui.NewTreeView()
	tabManager := tui.NewTabManager()
	statusBar := tui.CreatePowerlineStatusBar(80)
	expansionCache := NewResourceExpansionCache(5 * time.Minute)
	healthMonitor := NewResourceHealthMonitor()

	// Add a default tab
	tabManager.AddTab(tui.Tab{
		Title:    "Azure Resources",
		Content:  "Welcome to Azure TUI",
		Type:     "main",
		Closable: false,
	})

	return model{
		treeView:          treeView,
		tabManager:        tabManager,
		statusBar:         statusBar,
		expansionCache:    expansionCache,
		healthMonitor:     healthMonitor,
		bulkSelectionMode: false,
	}
}

// BubbleTea methods
func (m model) Init() tea.Cmd {
	return nil
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

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "j", "down":
			if m.treeView != nil {
				m.treeView.SelectNext()
			}
			return m, nil
		case "k", "up":
			if m.treeView != nil {
				m.treeView.SelectPrevious()
			}
			return m, nil
		case " ":
			if m.treeView != nil {
				m.treeView.ToggleExpansion()
			}
			return m, nil
		case "r":
			// Refresh current view
			return m, nil
		}

	case resourcesLoadingMsg:
		// Handle resource loading
		return m, nil

	case resourcesInGroupMsg:
		// Handle loaded resources
		m.resourcesInGroup = msg.resources
		return m, nil

	case resourcesInGroupErrMsg:
		// Handle resource loading error
		return m, nil
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
		m.statusBar.AddSegment("Ready", lipgloss.Color("33"), lipgloss.Color("15"))
	}

	// Render tree view
	treeContent := ""
	if m.treeView != nil {
		treeContent = m.treeView.RenderTreeView(m.width/3, m.height-3)
	}

	// Render tabs content
	tabsContent := ""
	if m.tabManager != nil {
		tabsContent = tui.RenderTabs(m.tabManager, "Status: Ready")
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
