package devops

import (
	"time"
)

// Organization represents an Azure DevOps organization
type Organization struct {
	ID          string `json:"accountId"`
	Name        string `json:"accountName"`
	URL         string `json:"accountUri"`
	Description string `json:"description"`
}

// Project represents an Azure DevOps project
type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	State       string `json:"state"`
}

// Repository represents a repository linked to a pipeline
type Repository struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
	URL  string `json:"url"`
}

// User represents a user in Azure DevOps
type User struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	UniqueName  string `json:"uniqueName"`
}

// Pipeline represents a build or release pipeline
type Pipeline struct {
	ID           int          `json:"id"`
	Name         string       `json:"name"`
	Path         string       `json:"path"`
	Repository   Repository   `json:"repository"`
	LastRun      *PipelineRun `json:"lastRun,omitempty"`
	Status       string       `json:"status"`
	Type         string       `json:"type"` // "build" or "release"
	CreatedDate  time.Time    `json:"createdDate"`
	ModifiedDate time.Time    `json:"modifiedDate"`
}

// PipelineRun represents a pipeline execution
type PipelineRun struct {
	ID           int                    `json:"id"`
	BuildNumber  string                 `json:"buildNumber"`
	Status       string                 `json:"status"`
	Result       string                 `json:"result"`
	StartTime    time.Time              `json:"startTime"`
	FinishTime   time.Time              `json:"finishTime"`
	Duration     time.Duration          `json:"-"`
	RequestedBy  User                   `json:"requestedBy"`
	SourceBranch string                 `json:"sourceBranch"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
}

// LogEntry represents a line in build/release logs
type LogEntry struct {
	LineNumber int       `json:"lineNumber"`
	Timestamp  time.Time `json:"timestamp"`
	Message    string    `json:"message"`
	Level      string    `json:"level"`
}

// DevOpsTreeNode represents a node in the UI tree structure
type DevOpsTreeNode struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Type     string            `json:"type"` // "organization", "project", "build-pipelines", "release-pipelines", "pipeline", "run"
	Status   string            `json:"status"`
	LastRun  string            `json:"lastRun"`
	Children []*DevOpsTreeNode `json:"children"`
	Parent   *DevOpsTreeNode   `json:"-"`
	Expanded bool              `json:"expanded"`
	Data     interface{}       `json:"-"` // Holds actual Pipeline, Project, etc.
	Depth    int               `json:"depth"`
}

// DevOpsConfig holds configuration for Azure DevOps integration
type DevOpsConfig struct {
	PersonalAccessToken string `json:"personal_access_token"`
	Organization        string `json:"organization"`
	Project             string `json:"project"`
	BaseURL             string `json:"base_url"`
}

// TreeRenderer handles rendering of the borderless tree UI
type TreeRenderer struct {
	nodes          []*DevOpsTreeNode
	selectedIndex  int
	scrollOffset   int
	maxVisibleRows int
	indentSize     int
	width          int
	height         int
}

// Helper methods for DevOpsTreeNode
func (n *DevOpsTreeNode) hasChildren() bool {
	return len(n.Children) > 0
}

func (n *DevOpsTreeNode) getDepth() int {
	return n.Depth
}

func (n *DevOpsTreeNode) addChild(child *DevOpsTreeNode) {
	child.Parent = n
	child.Depth = n.Depth + 1
	n.Children = append(n.Children, child)
}

func (n *DevOpsTreeNode) removeChild(child *DevOpsTreeNode) {
	for i, c := range n.Children {
		if c.ID == child.ID {
			n.Children = append(n.Children[:i], n.Children[i+1:]...)
			break
		}
	}
}

func (n *DevOpsTreeNode) toggleExpansion() {
	n.Expanded = !n.Expanded
}

// Flatten tree for rendering
func (n *DevOpsTreeNode) flattenVisible() []*DevOpsTreeNode {
	var result []*DevOpsTreeNode
	result = append(result, n)

	if n.Expanded {
		for _, child := range n.Children {
			result = append(result, child.flattenVisible()...)
		}
	}

	return result
}
