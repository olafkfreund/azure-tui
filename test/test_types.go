package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Enhanced types for testing
type EnhancedAzureResource struct {
	ID           string
	Name         string
	Type         string
	Location     string
	HealthStatus string
	LastUpdated  time.Time
	Metadata     map[string]interface{}
	Dependencies []string
	Status       string
	Tags         map[string]string
	Cost         ResourceCost
	Metrics      ResourceMetrics
}

type ResourceCost struct {
	DailyCost   float64
	MonthlyCost float64
	Currency    string
}

type ResourceMetrics struct {
	CPUUtilization    float64
	MemoryUtilization float64
	NetworkIn         float64
	NetworkOut        float64
	DiskIO            float64
	LastCollected     time.Time
}

type ResourceHealthMonitor struct {
	isRunning      bool
	resources      map[string]*EnhancedAzureResource
	isMonitoring   bool
	lastUpdate     time.Time
	warningCount   int
	criticalCount  int
	unknownCount   int
	totalResources int
	updateInterval time.Duration
	healthyCount   int
}

type LoadingProgress struct {
	message     string
	progress    float64
	totalSteps  int
	completed   bool
	isLoading   bool
	startTime   time.Time
	timeout     time.Duration
	currentStep int
}

// AzureResource represents a basic Azure resource (already defined in main.go)
type AzureResource struct {
	ID       string
	Name     string
	Type     string
	Location string
	Status   string
}

// ResourceGroup represents an Azure resource group (already defined in main.go)
type ResourceGroup struct {
	Name     string
	Location string
}

// Model represents the TUI model with basic fields for testing
type Model struct {
	resourceStatusCache map[string]*EnhancedAzureResource
	autoRefreshEnabled  bool
	keyBindings         map[string]bool
	healthMonitor       *ResourceHealthMonitor
	loadingProgress     *LoadingProgress
	currentInterface    string
	treeView            *TreeView
}

type TreeView struct {
	selectedNode *TreeNode
}

type TreeNode struct {
	Type string
	Data interface{}
}

func (tv *TreeView) GetSelectedNode() *TreeNode {
	return tv.selectedNode
}

func (tv *TreeView) AddResource(parent interface{}, name, resourceType string, resource interface{}) {
	// Mock implementation
}

// Helper functions for tests
func initModel() Model {
	return Model{
		resourceStatusCache: make(map[string]*EnhancedAzureResource),
		autoRefreshEnabled:  false,
		keyBindings:         make(map[string]bool),
		healthMonitor:       &ResourceHealthMonitor{},
		loadingProgress:     &LoadingProgress{},
		currentInterface:    "default",
		treeView:            &TreeView{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	return "Test view"
}

func (m Model) renderStatusBar(args ...interface{}) string {
	return "Status bar"
}

// Helper function to calculate health counts
func calculateHealthCounts(resources []EnhancedAzureResource) (healthy, warning, critical, unknown int) {
	for _, resource := range resources {
		switch resource.HealthStatus {
		case "Healthy":
			healthy++
		case "Warning":
			warning++
		case "Critical":
			critical++
		default:
			unknown++
		}
	}
	return
}
