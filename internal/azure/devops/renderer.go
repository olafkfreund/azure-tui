package devops

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Color definitions for borderless design
var (
	colorOrgHeader    = lipgloss.Color("39")  // Bright blue
	colorProjectName  = lipgloss.Color("35")  // Bright green
	colorPipelineName = lipgloss.Color("37")  // Bright cyan
	colorStatusGood   = lipgloss.Color("32")  // Green
	colorStatusBad    = lipgloss.Color("31")  // Red
	colorStatusWarn   = lipgloss.Color("33")  // Yellow
	colorSelected     = lipgloss.Color("226") // Bright yellow
	colorMuted        = lipgloss.Color("240") // Gray
	colorRunning      = lipgloss.Color("39")  // Blue
)

// NewTreeRenderer creates a new tree renderer for borderless UI
func NewTreeRenderer(width, height int) *TreeRenderer {
	return &TreeRenderer{
		nodes:          []*DevOpsTreeNode{},
		selectedIndex:  0,
		scrollOffset:   0,
		maxVisibleRows: height - 5, // Account for header and status bar
		indentSize:     2,
		width:          width,
		height:         height,
	}
}

// RenderTree renders the complete borderless tree interface
func (tr *TreeRenderer) RenderTree() string {
	var result strings.Builder

	// Header with organization/project context (no borders)
	result.WriteString(tr.renderHeader())
	result.WriteString("\n\n")

	// Tree content without borders
	visibleNodes := tr.getVisibleNodes()
	for i, node := range visibleNodes {
		line := tr.renderTreeNode(node, i == tr.selectedIndex)
		result.WriteString(line)
		result.WriteString("\n")
	}

	// Add spacing before status bar
	result.WriteString("\n")
	result.WriteString(tr.renderStatusBar())

	return result.String()
}

// renderHeader creates the clean header without borders
func (tr *TreeRenderer) renderHeader() string {
	var header strings.Builder

	titleStyle := lipgloss.NewStyle().
		Foreground(colorOrgHeader).
		Bold(true)

	header.WriteString(titleStyle.Render("Azure DevOps Manager"))
	header.WriteString("\n\n")

	// Organization and project info
	if len(tr.nodes) > 0 {
		if orgNode := tr.findNodeByType("organization"); orgNode != nil {
			orgStyle := lipgloss.NewStyle().Foreground(colorOrgHeader)
			header.WriteString(fmt.Sprintf("Organization: %s", orgStyle.Render(orgNode.Name)))

			if projNode := tr.findSelectedProject(); projNode != nil {
				projStyle := lipgloss.NewStyle().Foreground(colorProjectName)
				header.WriteString(fmt.Sprintf("          Project: %s", projStyle.Render(projNode.Name)))
			}
		}
	}

	return header.String()
}

// renderTreeNode renders a single tree node without borders
func (tr *TreeRenderer) renderTreeNode(node *DevOpsTreeNode, selected bool) string {
	var line strings.Builder

	// Indentation based on tree depth
	indent := strings.Repeat("  ", node.getDepth())
	line.WriteString(indent)

	// Tree structure symbols (clean, no borders)
	if node.hasChildren() {
		if node.Expanded {
			line.WriteString("└─ ")
		} else {
			line.WriteString("├─ ")
		}
	} else {
		line.WriteString("   ")
	}

	// Icon based on node type
	icon := tr.getNodeIcon(node)
	line.WriteString(icon + " ")

	// Node name with selection highlight
	name := node.Name
	if selected {
		selectedStyle := lipgloss.NewStyle().
			Foreground(colorSelected).
			Bold(true)
		name = selectedStyle.Render("► " + name)
	} else {
		name = tr.applyNodeStyle(name, node)
	}
	line.WriteString(name)

	// Status indicator for pipelines (no borders/brackets)
	if node.Type == "pipeline" {
		status := tr.getStatusIndicator(node.Status)
		line.WriteString("  " + status)

		if node.LastRun != "" {
			lastRunStyle := lipgloss.NewStyle().Foreground(colorMuted)
			line.WriteString("   " + lastRunStyle.Render("Last: "+node.LastRun))
		}
	}

	return line.String()
}

// getNodeIcon returns appropriate icons for different node types
func (tr *TreeRenderer) getNodeIcon(node *DevOpsTreeNode) string {
	switch node.Type {
	case "organization":
		return "🏢"
	case "project":
		return "📁"
	case "build-pipelines":
		return "🔧"
	case "release-pipelines":
		return "🚀"
	case "pipeline":
		return "⚙️ "
	case "pipeline-run":
		return "📋"
	case "recent-activity":
		return "📊"
	default:
		return "•"
	}
}

// getStatusIndicator returns clean status indicators without borders
func (tr *TreeRenderer) getStatusIndicator(status string) string {
	switch strings.ToLower(status) {
	case "running", "inprogress":
		style := lipgloss.NewStyle().Foreground(colorRunning)
		return style.Render("Running")
	case "succeeded", "success":
		style := lipgloss.NewStyle().Foreground(colorStatusGood)
		return style.Render("Success")
	case "failed", "error":
		style := lipgloss.NewStyle().Foreground(colorStatusBad)
		return style.Render("Failed")
	case "canceled", "cancelled":
		style := lipgloss.NewStyle().Foreground(colorStatusWarn)
		return style.Render("Canceled")
	case "queued", "pending":
		style := lipgloss.NewStyle().Foreground(colorStatusWarn)
		return style.Render("Queued")
	default:
		style := lipgloss.NewStyle().Foreground(colorMuted)
		return style.Render("Unknown")
	}
}

// applyNodeStyle applies appropriate styling based on node type
func (tr *TreeRenderer) applyNodeStyle(text string, node *DevOpsTreeNode) string {
	var style lipgloss.Style

	switch node.Type {
	case "organization":
		style = lipgloss.NewStyle().Foreground(colorOrgHeader).Bold(true)
	case "project":
		style = lipgloss.NewStyle().Foreground(colorProjectName)
	case "build-pipelines", "release-pipelines":
		style = lipgloss.NewStyle().Foreground(colorPipelineName).Bold(true)
	case "pipeline":
		style = lipgloss.NewStyle().Foreground(colorPipelineName)
	case "recent-activity":
		style = lipgloss.NewStyle().Foreground(colorMuted).Italic(true)
	default:
		style = lipgloss.NewStyle().Foreground(colorMuted)
	}

	return style.Render(text)
}

// renderStatusBar creates clean status bar without borders
func (tr *TreeRenderer) renderStatusBar() string {
	shortcuts := "Navigate: j/k  Expand: Space  Select: Enter  Back: Esc"

	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("15")).
		Background(lipgloss.Color("240")).
		Padding(0, 1).
		Width(tr.width - 2)

	return statusStyle.Render("DevOps: " + shortcuts)
}

// getVisibleNodes returns nodes that should be visible based on scroll offset
func (tr *TreeRenderer) getVisibleNodes() []*DevOpsTreeNode {
	allVisible := tr.flattenAllVisible()

	start := tr.scrollOffset
	end := start + tr.maxVisibleRows

	if start >= len(allVisible) {
		return []*DevOpsTreeNode{}
	}

	if end > len(allVisible) {
		end = len(allVisible)
	}

	return allVisible[start:end]
}

// flattenAllVisible flattens all visible nodes into a single slice
func (tr *TreeRenderer) flattenAllVisible() []*DevOpsTreeNode {
	var result []*DevOpsTreeNode

	for _, root := range tr.nodes {
		result = append(result, root.flattenVisible()...)
	}

	return result
}

// Navigation helper methods
func (tr *TreeRenderer) moveUp() {
	if tr.selectedIndex > 0 {
		tr.selectedIndex--
		if tr.selectedIndex < tr.scrollOffset {
			tr.scrollOffset = tr.selectedIndex
		}
	}
}

func (tr *TreeRenderer) moveDown() {
	allVisible := tr.flattenAllVisible()
	if tr.selectedIndex < len(allVisible)-1 {
		tr.selectedIndex++
		if tr.selectedIndex >= tr.scrollOffset+tr.maxVisibleRows {
			tr.scrollOffset = tr.selectedIndex - tr.maxVisibleRows + 1
		}
	}
}

func (tr *TreeRenderer) toggleExpansion() {
	allVisible := tr.flattenAllVisible()
	if tr.selectedIndex < len(allVisible) {
		node := allVisible[tr.selectedIndex]
		if node.hasChildren() {
			node.toggleExpansion()
		}
	}
}

func (tr *TreeRenderer) getSelectedNode() *DevOpsTreeNode {
	allVisible := tr.flattenAllVisible()
	if tr.selectedIndex < len(allVisible) {
		return allVisible[tr.selectedIndex]
	}
	return nil
}

// Helper methods to find specific nodes
func (tr *TreeRenderer) findNodeByType(nodeType string) *DevOpsTreeNode {
	var findNode func([]*DevOpsTreeNode) *DevOpsTreeNode
	findNode = func(nodes []*DevOpsTreeNode) *DevOpsTreeNode {
		for _, node := range nodes {
			if node.Type == nodeType {
				return node
			}
			if found := findNode(node.Children); found != nil {
				return found
			}
		}
		return nil
	}

	return findNode(tr.nodes)
}

func (tr *TreeRenderer) findSelectedProject() *DevOpsTreeNode {
	// Find the first expanded project node
	var findProject func([]*DevOpsTreeNode) *DevOpsTreeNode
	findProject = func(nodes []*DevOpsTreeNode) *DevOpsTreeNode {
		for _, node := range nodes {
			if node.Type == "project" && node.Expanded {
				return node
			}
			if found := findProject(node.Children); found != nil {
				return found
			}
		}
		return nil
	}

	return findProject(tr.nodes)
}

// SetNodes updates the tree nodes
func (tr *TreeRenderer) SetNodes(nodes []*DevOpsTreeNode) {
	tr.nodes = nodes
	tr.selectedIndex = 0
	tr.scrollOffset = 0
}

// AddRootNode adds a root node to the tree
func (tr *TreeRenderer) AddRootNode(node *DevOpsTreeNode) {
	node.Depth = 0
	tr.nodes = append(tr.nodes, node)
}

// BuildTreeFromData converts Azure DevOps data into tree node structure
func (tr *TreeRenderer) BuildTreeFromData(organizations []Organization, projects []Project, pipelines []Pipeline, builds []PipelineRun) {
	var rootNodes []*DevOpsTreeNode

	// If we have organizations
	for _, org := range organizations {
		orgNode := &DevOpsTreeNode{
			ID:       org.ID,
			Name:     org.Name,
			Type:     "organization",
			Status:   "active",
			Depth:    0,
			Expanded: true,
			Children: []*DevOpsTreeNode{},
		}

		// Add projects under organization
		for _, proj := range projects {
			projNode := &DevOpsTreeNode{
				ID:       proj.ID,
				Name:     proj.Name,
				Type:     "project",
				Status:   "active",
				Depth:    1,
				Expanded: true,
				Children: []*DevOpsTreeNode{},
				Parent:   orgNode,
			}

			// Create pipeline folders
			buildPipelinesNode := &DevOpsTreeNode{
				ID:       "build-pipelines",
				Name:     "Build Pipelines",
				Type:     "build-pipelines",
				Status:   "active",
				Depth:    2,
				Expanded: false,
				Children: []*DevOpsTreeNode{},
				Parent:   projNode,
			}

			releasePipelinesNode := &DevOpsTreeNode{
				ID:       "release-pipelines",
				Name:     "Release Pipelines",
				Type:     "release-pipelines",
				Status:   "active",
				Depth:    2,
				Expanded: false,
				Children: []*DevOpsTreeNode{},
				Parent:   projNode,
			}

			projNode.addChild(buildPipelinesNode)
			projNode.addChild(releasePipelinesNode)

			orgNode.addChild(projNode)
		}

		rootNodes = append(rootNodes, orgNode)
	}

	tr.SetNodes(rootNodes)
}
