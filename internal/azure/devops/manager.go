package devops

import (
	"fmt"
	"strings"
	"time"
)

// DevOpsManager handles the overall DevOps functionality and tree management
type DevOpsManager struct {
	client        *DevOpsClient
	renderer      *TreeRenderer
	organizations []Organization
	projects      []Project
	pipelines     []Pipeline
	currentOrg    string
	currentProj   string
	refreshing    bool
}

// NewDevOpsManager creates a new DevOps manager
func NewDevOpsManager(config DevOpsConfig, width, height int) *DevOpsManager {
	client := NewDevOpsClient(config)
	renderer := NewTreeRenderer(width, height)

	return &DevOpsManager{
		client:      client,
		renderer:    renderer,
		currentOrg:  config.Organization,
		currentProj: config.Project,
	}
}

// Initialize loads initial data and builds the tree structure
func (dm *DevOpsManager) Initialize() error {
	// Test connection first
	if err := dm.client.TestConnection(); err != nil {
		return fmt.Errorf("failed to connect to Azure DevOps: %w", err)
	}

	// Load organizations
	orgs, err := dm.client.ListOrganizations()
	if err != nil {
		return fmt.Errorf("failed to load organizations: %w", err)
	}
	dm.organizations = orgs

	// Build initial tree
	dm.buildInitialTree()

	return nil
}

// buildInitialTree creates the initial tree structure
func (dm *DevOpsManager) buildInitialTree() {
	var rootNodes []*DevOpsTreeNode

	// Create organization nodes
	for _, org := range dm.organizations {
		orgNode := &DevOpsTreeNode{
			ID:       org.ID,
			Name:     org.Name,
			Type:     "organization",
			Status:   "",
			Children: []*DevOpsTreeNode{},
			Expanded: org.Name == dm.currentOrg, // Expand current org
			Data:     org,
			Depth:    0,
		}

		// If this is the current organization, load projects
		if org.Name == dm.currentOrg {
			dm.loadProjectsForOrg(orgNode)
		}

		rootNodes = append(rootNodes, orgNode)
	}

	dm.renderer.SetNodes(rootNodes)
}

// loadProjectsForOrg loads projects for a specific organization
func (dm *DevOpsManager) loadProjectsForOrg(orgNode *DevOpsTreeNode) {
	// Update client organization
	dm.client.organization = orgNode.Name

	projects, err := dm.client.ListProjects()
	if err != nil {
		// Add error node
		errorNode := &DevOpsTreeNode{
			ID:   "error-projects",
			Name: fmt.Sprintf("Error loading projects: %v", err),
			Type: "error",
		}
		orgNode.addChild(errorNode)
		return
	}

	dm.projects = projects

	// Create project nodes
	for _, project := range projects {
		projNode := &DevOpsTreeNode{
			ID:       project.ID,
			Name:     project.Name,
			Type:     "project",
			Status:   project.State,
			Children: []*DevOpsTreeNode{},
			Expanded: project.Name == dm.currentProj, // Expand current project
			Data:     project,
		}

		// If this is the current project, load pipelines
		if project.Name == dm.currentProj {
			dm.loadPipelinesForProject(projNode)
		}

		orgNode.addChild(projNode)
	}
}

// loadPipelinesForProject loads pipelines for a specific project
func (dm *DevOpsManager) loadPipelinesForProject(projNode *DevOpsTreeNode) {
	// Update client project
	dm.client.project = projNode.Name

	// Create pipeline category nodes
	buildPipelinesNode := &DevOpsTreeNode{
		ID:       "build-pipelines",
		Name:     "Build Pipelines",
		Type:     "build-pipelines",
		Children: []*DevOpsTreeNode{},
		Expanded: true,
	}

	releasePipelinesNode := &DevOpsTreeNode{
		ID:       "release-pipelines",
		Name:     "Release Pipelines",
		Type:     "release-pipelines",
		Children: []*DevOpsTreeNode{},
		Expanded: true,
	}

	recentActivityNode := &DevOpsTreeNode{
		ID:       "recent-activity",
		Name:     "Recent Activity",
		Type:     "recent-activity",
		Children: []*DevOpsTreeNode{},
		Expanded: false,
	}

	// Load build pipelines
	buildPipelines, err := dm.client.ListBuildPipelines()
	if err == nil {
		for _, pipeline := range buildPipelines {
			pipelineNode := dm.createPipelineNode(pipeline)
			buildPipelinesNode.addChild(pipelineNode)
		}
	} else {
		errorNode := &DevOpsTreeNode{
			ID:   "error-build",
			Name: fmt.Sprintf("Error: %v", err),
			Type: "error",
		}
		buildPipelinesNode.addChild(errorNode)
	}

	// Load release pipelines
	releasePipelines, err := dm.client.ListReleasePipelines()
	if err == nil {
		for _, pipeline := range releasePipelines {
			pipelineNode := dm.createPipelineNode(pipeline)
			releasePipelinesNode.addChild(pipelineNode)
		}
	} else {
		errorNode := &DevOpsTreeNode{
			ID:   "error-release",
			Name: fmt.Sprintf("Error: %v", err),
			Type: "error",
		}
		releasePipelinesNode.addChild(errorNode)
	}

	// Add category nodes to project
	projNode.addChild(buildPipelinesNode)
	projNode.addChild(releasePipelinesNode)
	projNode.addChild(recentActivityNode)

	// Store all pipelines
	dm.pipelines = append(buildPipelines, releasePipelines...)
}

// createPipelineNode creates a tree node for a pipeline
func (dm *DevOpsManager) createPipelineNode(pipeline Pipeline) *DevOpsTreeNode {
	// Get last run info
	lastRun := ""
	status := "Unknown"

	if pipeline.LastRun != nil {
		status = pipeline.LastRun.Status
		if pipeline.LastRun.Result != "" {
			status = pipeline.LastRun.Result
		}

		// Format last run time
		if !pipeline.LastRun.StartTime.IsZero() {
			lastRun = formatRelativeTime(pipeline.LastRun.StartTime)
		}
	}

	return &DevOpsTreeNode{
		ID:      fmt.Sprintf("pipeline-%d", pipeline.ID),
		Name:    pipeline.Name,
		Type:    "pipeline",
		Status:  status,
		LastRun: lastRun,
		Data:    pipeline,
	}
}

// formatRelativeTime formats a time as relative to now
func formatRelativeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	} else if diff < time.Hour {
		return fmt.Sprintf("%dm ago", int(diff.Minutes()))
	} else if diff < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(diff.Hours()))
	} else if diff < 7*24*time.Hour {
		return fmt.Sprintf("%dd ago", int(diff.Hours()/24))
	} else {
		return fmt.Sprintf("%dw ago", int(diff.Hours()/(24*7)))
	}
}

// Render returns the complete rendered tree
func (dm *DevOpsManager) Render() string {
	if dm.refreshing {
		return dm.renderLoadingScreen()
	}

	return dm.renderer.RenderTree()
}

// renderLoadingScreen shows a loading message
func (dm *DevOpsManager) renderLoadingScreen() string {
	var content strings.Builder

	content.WriteString("Azure DevOps Manager\n\n")
	content.WriteString("🔄 Loading Azure DevOps data...\n\n")
	content.WriteString("This may take a few moments while we:\n")
	content.WriteString("• Connect to Azure DevOps\n")
	content.WriteString("• Load organizations and projects\n")
	content.WriteString("• Retrieve pipeline information\n\n")
	content.WriteString("Please wait...")

	return content.String()
}

// Navigation methods
func (dm *DevOpsManager) MoveUp() {
	dm.renderer.moveUp()
}

func (dm *DevOpsManager) MoveDown() {
	dm.renderer.moveDown()
}

func (dm *DevOpsManager) ToggleExpansion() {
	dm.renderer.toggleExpansion()

	// If expanding a node that needs data loading, load it
	if selectedNode := dm.renderer.getSelectedNode(); selectedNode != nil {
		if selectedNode.Expanded && len(selectedNode.Children) == 0 {
			switch selectedNode.Type {
			case "organization":
				dm.loadProjectsForOrg(selectedNode)
			case "project":
				dm.loadPipelinesForProject(selectedNode)
			}
		}
	}
}

func (dm *DevOpsManager) GetSelectedNode() *DevOpsTreeNode {
	return dm.renderer.getSelectedNode()
}

// Action methods
func (dm *DevOpsManager) RunSelectedPipeline() error {
	selectedNode := dm.GetSelectedNode()
	if selectedNode == nil || selectedNode.Type != "pipeline" {
		return fmt.Errorf("no pipeline selected")
	}

	pipeline, ok := selectedNode.Data.(Pipeline)
	if !ok {
		return fmt.Errorf("invalid pipeline data")
	}

	// Run pipeline with no parameters for now
	run, err := dm.client.RunPipeline(pipeline.ID, map[string]interface{}{})
	if err != nil {
		return fmt.Errorf("failed to run pipeline: %w", err)
	}

	// Update node status
	selectedNode.Status = run.Status

	return nil
}

func (dm *DevOpsManager) RefreshData() error {
	dm.refreshing = true
	defer func() { dm.refreshing = false }()

	// Reload the tree
	return dm.Initialize()
}

// SetDimensions updates the renderer dimensions
func (dm *DevOpsManager) SetDimensions(width, height int) {
	dm.renderer.width = width
	dm.renderer.height = height
	dm.renderer.maxVisibleRows = height - 5
}
