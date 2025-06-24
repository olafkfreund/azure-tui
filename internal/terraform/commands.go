package terraform

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/olafkfreund/azure-tui/internal/azure/tfbicep"
	"github.com/olafkfreund/azure-tui/internal/config"
)

// Additional rendering methods for TerraformTUI

func (m *TerraformTUI) renderTemplatesView() string {
	// Clean, frameless styling for consistency
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(1, 2)

	if m.activeView == ViewTemplates {
		style = style.Foreground(lipgloss.Color("#FF5F87"))
	}

	return style.Render(m.templates.View())
}

func (m *TerraformTUI) renderWorkspacesView() string {
	// Clean, frameless styling for consistency
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(1, 2)

	if m.activeView == ViewWorkspaces {
		style = style.Foreground(lipgloss.Color("#FF5F87"))
	}

	return style.Render(m.workspaces.View())
}

func (m *TerraformTUI) renderEditorView() string {
	// Clean, frameless styling for consistency
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(1, 2)

	if m.activeView == ViewEditor {
		style = style.Foreground(lipgloss.Color("#FF5F87"))
	}

	title := fmt.Sprintf("Editor - %s", m.currentFile)
	if m.currentFile == "" {
		title = "Editor - New File"
	}

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Bold(true).Render(title),
		"",
		m.editor.View(),
	)

	return style.Render(content)
}

func (m *TerraformTUI) renderOperationsView() string {
	// Clean, frameless styling for consistency
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(1, 2)

	if m.activeView == ViewOperations {
		style = style.Foreground(lipgloss.Color("#FF5F87"))
	}

	var operations []string
	operations = append(operations, lipgloss.NewStyle().Bold(true).Render("Recent Operations"))
	operations = append(operations, "")

	if len(m.operations) == 0 {
		operations = append(operations, "No operations yet")
	} else {
		// Show last 10 operations
		start := 0
		if len(m.operations) > 10 {
			start = len(m.operations) - 10
		}

		for i := start; i < len(m.operations); i++ {
			op := m.operations[i]
			status := "✓"
			color := lipgloss.Color("#50FA7B")
			if !op.Success {
				status = "✗"
				color = lipgloss.Color("#FF5555")
			}

			opLine := fmt.Sprintf("%s %s (%s)",
				lipgloss.NewStyle().Foreground(color).Render(status),
				op.Command,
				op.Duration,
			)
			operations = append(operations, opLine)
		}
	}

	content := strings.Join(operations, "\n")
	return style.Render(content)
}

func (m *TerraformTUI) renderStateView() string {
	// Clean, frameless styling for consistency
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(1, 2)

	if m.activeView == ViewState {
		style = style.Foreground(lipgloss.Color("#FF5F87"))
	}

	var stateInfo []string
	stateInfo = append(stateInfo, lipgloss.NewStyle().Bold(true).Render("Terraform State"))
	stateInfo = append(stateInfo, "")

	if m.manager == nil {
		stateInfo = append(stateInfo, "No workspace selected")
	} else {
		state, err := m.manager.GetState()
		if err != nil {
			stateInfo = append(stateInfo, fmt.Sprintf("Error: %s", err))
		} else {
			stateInfo = append(stateInfo, fmt.Sprintf("Version: %d", state.Version))
			stateInfo = append(stateInfo, fmt.Sprintf("Serial: %d", state.Serial))
			stateInfo = append(stateInfo, fmt.Sprintf("Resources: %d", len(state.Resources)))
			stateInfo = append(stateInfo, fmt.Sprintf("Outputs: %d", len(state.Outputs)))
			stateInfo = append(stateInfo, "")

			if len(state.Resources) > 0 {
				stateInfo = append(stateInfo, "Resources:")
				for _, resource := range state.Resources {
					stateInfo = append(stateInfo, fmt.Sprintf("  • %s.%s (%s)", resource.Type, resource.Name, resource.Mode))
				}
			}
		}
	}

	content := strings.Join(stateInfo, "\n")
	return style.Render(content)
}

// Enhanced view rendering methods

func (m *TerraformTUI) renderStateViewerView() string {
	// Clean, frameless styling for consistency
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(1, 2)

	if m.activeView == ViewStateViewer {
		style = style.Foreground(lipgloss.Color("#FF5F87"))
	}

	var content []string
	content = append(content, lipgloss.NewStyle().Bold(true).Render("Terraform State Resources"))
	content = append(content, "")

	if len(m.stateViewer.resources) == 0 {
		content = append(content, "No state resources found")
		content = append(content, "Press 's' to load state resources")
	} else {
		// Header
		content = append(content, fmt.Sprintf("Total Resources: %d", len(m.stateViewer.resources)))
		content = append(content, "")

		// Show resources
		for i, resource := range m.stateViewer.resources {
			prefix := "  "
			if i == m.selectedResource {
				prefix = "▶ "
			}

			statusIcon := "✓"
			if resource.Tainted {
				statusIcon = "⚠"
			}

			line := fmt.Sprintf("%s%s %s.%s (%s)",
				prefix, statusIcon, resource.Type, resource.Name, resource.Status)

			if i == m.selectedResource {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Render(line)
			}

			content = append(content, line)

			// Show dependencies if enabled and selected
			if m.showDependencies && i == m.selectedResource && len(resource.Dependencies) > 0 {
				for _, dep := range resource.Dependencies {
					content = append(content, fmt.Sprintf("    └─ depends on: %s", dep))
				}
			}
		}

		content = append(content, "")
		content = append(content, "d: toggle dependencies | ↑/↓: navigate")
	}

	return style.Render(strings.Join(content, "\n"))
}

func (m *TerraformTUI) renderPlanViewerView() string {
	// Clean, frameless styling for consistency
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(1, 2)

	if m.activeView == ViewPlanViewer {
		style = style.Foreground(lipgloss.Color("#FF5F87"))
	}

	var content []string
	content = append(content, lipgloss.NewStyle().Bold(true).Render("Terraform Plan Changes"))
	content = append(content, "")

	if len(m.planViewer.changes) == 0 {
		content = append(content, "No plan changes found")
		content = append(content, "Press 'p' to load plan changes")
	} else {
		// Filter summary
		filterText := "All changes"
		if m.planViewer.filterAction != "" {
			filterText = fmt.Sprintf("Filter: %s", m.planViewer.filterAction)
		}

		content = append(content, fmt.Sprintf("Changes: %d | %s", len(m.planViewer.changes), filterText))
		content = append(content, "")

		// Count changes by action
		actionCounts := make(map[string]int)
		for _, change := range m.planViewer.changes {
			actionCounts[change.Action]++
		}

		var summary []string
		for action, count := range actionCounts {
			icon := getActionIcon(action)
			summary = append(summary, fmt.Sprintf("%s %d %s", icon, count, action))
		}
		content = append(content, strings.Join(summary, " | "))
		content = append(content, "")

		// Show filtered changes
		filteredChanges := m.planViewer.changes
		if m.planViewer.filterAction != "" {
			var filtered []PlanChange
			for _, change := range m.planViewer.changes {
				if change.Action == m.planViewer.filterAction {
					filtered = append(filtered, change)
				}
			}
			filteredChanges = filtered
		}

		for i, change := range filteredChanges {
			prefix := "  "
			if i == m.selectedChange {
				prefix = "▶ "
			}

			icon := getActionIcon(change.Action)
			line := fmt.Sprintf("%s%s %s (%s)",
				prefix, icon, change.Resource, change.Action)

			if i == m.selectedChange {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Render(line)
			}

			content = append(content, line)

			// Show details if enabled and selected
			if m.showPlanDetails && i == m.selectedChange {
				if change.Reason != "" {
					content = append(content, fmt.Sprintf("    Reason: %s", change.Reason))
				}
				content = append(content, fmt.Sprintf("    Impact: %s", change.Impact))
			}
		}

		content = append(content, "")
		content = append(content, "f: filter toggle | a: approval mode | t: target resource")
	}

	return style.Render(strings.Join(content, "\n"))
}

func (m *TerraformTUI) renderEnvManagerView() string {
	// Clean, frameless styling for consistency
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(1, 2)

	if m.activeView == ViewEnvManager {
		style = style.Foreground(lipgloss.Color("#FF5F87"))
	}

	var content []string
	content = append(content, lipgloss.NewStyle().Bold(true).Render("Workspace Management"))
	content = append(content, "")

	if len(m.workspaceManager.workspaces) == 0 {
		content = append(content, "No workspaces found")
		content = append(content, "Press 'w' to load workspace info")
	} else {
		// Current environment info
		content = append(content, fmt.Sprintf("Current Environment: %s", m.workspaceManager.currentEnv))
		content = append(content, fmt.Sprintf("Active Workspace: %s", m.currentWorkspace))
		content = append(content, "")

		// List workspaces
		content = append(content, "Available Workspaces:")
		for i, workspace := range m.workspaceManager.workspaces {
			prefix := "  "
			if i == m.workspaceManager.selectedIndex {
				prefix = "▶ "
			}

			current := ""
			if workspace.Name == m.currentWorkspace {
				current = " (current)"
			}

			statusIcon := getWorkspaceStatusIcon(workspace.Status)
			line := fmt.Sprintf("%s%s %s (%s)%s",
				prefix, statusIcon, workspace.Name, workspace.Environment, current)

			if i == m.workspaceManager.selectedIndex {
				line = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Render(line)
			}

			content = append(content, line)

			// Show details if selected
			if i == m.workspaceManager.selectedIndex {
				content = append(content, fmt.Sprintf("    Backend: %s", workspace.Backend))
				content = append(content, fmt.Sprintf("    Path: %s", workspace.Path))
				if workspace.LastApply != "" {
					content = append(content, fmt.Sprintf("    Last Apply: %s", workspace.LastApply))
				}
			}
		}

		content = append(content, "")
		content = append(content, "Enter: switch workspace | ↑/↓: navigate")
	}

	return style.Render(strings.Join(content, "\n"))
}

// Helper functions for enhanced views

func getActionIcon(action string) string {
	switch action {
	case "create":
		return "+"
	case "update":
		return "~"
	case "delete":
		return "-"
	case "replace":
		return "±"
	default:
		return "?"
	}
}

func getWorkspaceStatusIcon(status string) string {
	switch status {
	case "clean":
		return "✓"
	case "dirty":
		return "⚠"
	case "error":
		return "✗"
	default:
		return "?"
	}
}

// Helper functions for enhanced workspace management

func detectBackendType(workingDir string) string {
	// Read terraform configuration files to detect backend type
	configFiles := []string{"main.tf", "backend.tf", "terraform.tf", "providers.tf"}

	for _, file := range configFiles {
		filePath := filepath.Join(workingDir, file)
		if content, err := os.ReadFile(filePath); err == nil {
			contentStr := string(content)

			// Check for different backend types
			if strings.Contains(contentStr, `backend "s3"`) {
				return "s3"
			}
			if strings.Contains(contentStr, `backend "azurerm"`) {
				return "azurerm"
			}
			if strings.Contains(contentStr, `backend "gcs"`) {
				return "gcs"
			}
			if strings.Contains(contentStr, `backend "remote"`) {
				return "remote"
			}
			if strings.Contains(contentStr, `backend "http"`) {
				return "http"
			}
			if strings.Contains(contentStr, `backend "consul"`) {
				return "consul"
			}
			if strings.Contains(contentStr, `backend "etcdv3"`) {
				return "etcdv3"
			}
		}
	}

	// Check for .terraform directory which might contain backend config
	terraformDir := filepath.Join(workingDir, ".terraform")
	if _, err := os.Stat(terraformDir); err == nil {
		// Check terraform.tfstate for backend info
		if stateFile, err := os.ReadFile(filepath.Join(terraformDir, "terraform.tfstate")); err == nil {
			if strings.Contains(string(stateFile), `"backend"`) {
				return "configured"
			}
		}
	}

	return "local"
}

func getWorkspaceStatus(workingDir, workspaceName string) string {
	// Check for uncommitted terraform files
	cmd := exec.Command("git", "status", "--porcelain", "*.tf", "*.tfvars")
	cmd.Dir = workingDir
	if output, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
		return "dirty"
	}

	// Check if terraform plan shows changes
	cmd = exec.Command("terraform", "plan", "-detailed-exitcode", "-no-color")
	cmd.Dir = workingDir

	// Set workspace if not default
	if workspaceName != "default" {
		// Switch to workspace temporarily for status check
		selectCmd := exec.Command("terraform", "workspace", "select", workspaceName)
		selectCmd.Dir = workingDir
		selectCmd.Run() // Ignore errors for status check
	}

	if err := cmd.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			switch exitError.ExitCode() {
			case 1:
				return "error"
			case 2:
				return "changes-pending"
			}
		}
		return "unknown"
	}

	return "clean"
}

func checkStateLock(workingDir string) (bool, string, error) {
	lockFile := filepath.Join(workingDir, ".terraform", "terraform.tfstate.lock.info")

	if _, err := os.Stat(lockFile); os.IsNotExist(err) {
		return false, "", nil
	}

	content, err := os.ReadFile(lockFile)
	if err != nil {
		return false, "", err
	}

	var lockInfo struct {
		ID        string    `json:"ID"`
		Operation string    `json:"Operation"`
		Info      string    `json:"Info"`
		Who       string    `json:"Who"`
		Version   string    `json:"Version"`
		Created   time.Time `json:"Created"`
	}

	// Try to parse JSON lock info
	if err := json.Unmarshal(content, &lockInfo); err == nil {
		return true, fmt.Sprintf("locked by %s", lockInfo.Who), nil
	}

	// If JSON parsing fails, assume it's locked
	if len(content) > 0 {
		return true, "locked", nil
	}

	return false, "", nil
}

// Command methods

func (m *TerraformTUI) loadTemplates() tea.Cmd {
	return func() tea.Msg {
		cfg := config.GetTerraformConfig()
		templates, err := tfbicep.ListTemplates(cfg.TemplatesPath)
		if err != nil {
			return errorMsg{err}
		}

		var items []list.Item
		for _, template := range templates {
			templatePath := filepath.Join(cfg.TemplatesPath, template)
			info, err := tfbicep.GetTemplateInfo(templatePath)
			if err != nil {
				continue
			}

			description := fmt.Sprintf("Template with %d files", len(info["terraform_files"].([]string)))
			if readme, ok := info["readme"]; ok {
				lines := strings.Split(readme.(string), "\n")
				if len(lines) > 0 {
					description = strings.TrimSpace(lines[0])
					if len(description) > 60 {
						description = description[:60] + "..."
					}
				}
			}

			items = append(items, templateItem{
				title:       template,
				description: description,
				path:        templatePath,
			})
		}

		return templatesLoadedMsg{items}
	}
}

func (m *TerraformTUI) loadWorkspaces() tea.Cmd {
	return func() tea.Msg {
		cfg := config.GetTerraformConfig()
		workspaces := []string{}

		// Scan workspace directory
		if _, err := os.Stat(cfg.WorkspacePath); err == nil {
			entries, err := os.ReadDir(cfg.WorkspacePath)
			if err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						workspaces = append(workspaces, entry.Name())
					}
				}
			}
		}

		var items []list.Item
		for _, workspace := range workspaces {
			workspacePath := filepath.Join(cfg.WorkspacePath, workspace)

			// Check if it's a valid Terraform workspace
			if _, err := os.Stat(filepath.Join(workspacePath, "main.tf")); err == nil {
				description := "Terraform workspace"

				// Try to get resource count
				tm := tfbicep.NewTerraformManager(workspacePath)
				if count, err := tm.GetResourceCount(); err == nil {
					description = fmt.Sprintf("%d resources", count)
				}

				items = append(items, workspaceItem{
					title:       workspace,
					description: description,
					path:        workspacePath,
				})
			}
		}

		return workspacesLoadedMsg{items}
	}
}

func (m *TerraformTUI) selectTemplate(templatePath string) tea.Cmd {
	return func() tea.Msg {
		// Load template and set up workspace
		m.currentTemplate = templatePath
		if m.manager == nil {
			m.manager = tfbicep.NewTerraformManager(templatePath)
		}
		m.status = fmt.Sprintf("Selected template: %s", filepath.Base(templatePath))
		return nil
	}
}

func (m *TerraformTUI) selectWorkspace(workspacePath string) tea.Cmd {
	return func() tea.Msg {
		// Set up workspace manager
		m.currentWorkspace = filepath.Base(workspacePath)
		m.manager = tfbicep.NewTerraformManager(workspacePath)
		m.status = fmt.Sprintf("Selected workspace: %s", m.currentWorkspace)
		return nil
	}
}

func (m *TerraformTUI) terraformInit() tea.Cmd {
	if m.manager == nil {
		return func() tea.Msg {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}
	}

	return func() tea.Msg {
		m.status = "Running terraform init..."
		op, err := m.manager.Init()
		if err != nil {
			return errorMsg{err}
		}
		return operationCompletedMsg{*op}
	}
}

func (m *TerraformTUI) terraformPlan() tea.Cmd {
	if m.manager == nil {
		return func() tea.Msg {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}
	}

	return func() tea.Msg {
		m.status = "Running terraform plan..."
		op, err := m.manager.Plan()
		if err != nil {
			return errorMsg{err}
		}
		return operationCompletedMsg{*op}
	}
}

func (m *TerraformTUI) terraformApply() tea.Cmd {
	if m.manager == nil {
		return func() tea.Msg {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}
	}

	return func() tea.Msg {
		m.status = "Running terraform apply..."
		op, err := m.manager.Apply()
		if err != nil {
			return errorMsg{err}
		}
		return operationCompletedMsg{*op}
	}
}

func (m *TerraformTUI) terraformDestroy() tea.Cmd {
	if m.manager == nil {
		return func() tea.Msg {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}
	}

	return func() tea.Msg {
		m.status = "Running terraform destroy..."
		op, err := m.manager.Destroy()
		if err != nil {
			return errorMsg{err}
		}
		return operationCompletedMsg{*op}
	}
}

func (m *TerraformTUI) terraformValidate() tea.Cmd {
	if m.manager == nil {
		return func() tea.Msg {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}
	}

	return func() tea.Msg {
		m.status = "Running terraform validate..."
		op, err := m.manager.Validate()
		if err != nil {
			return errorMsg{err}
		}
		return operationCompletedMsg{*op}
	}
}

func (m *TerraformTUI) terraformFormat() tea.Cmd {
	if m.manager == nil {
		return func() tea.Msg {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}
	}

	return func() tea.Msg {
		m.status = "Running terraform format..."
		op, err := m.manager.Format()
		if err != nil {
			return errorMsg{err}
		}
		return operationCompletedMsg{*op}
	}
}

func (m *TerraformTUI) loadState() tea.Cmd {
	if m.manager == nil {
		return func() tea.Msg {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}
	}

	return func() tea.Msg {
		// State is loaded in the view rendering
		return nil
	}
}

func (m *TerraformTUI) saveCurrentFile() tea.Cmd {
	if m.currentFile == "" {
		return func() tea.Msg {
			return errorMsg{fmt.Errorf("no file selected")}
		}
	}

	return func() tea.Msg {
		content := m.editor.Value()
		if err := os.WriteFile(m.currentFile, []byte(content), 0644); err != nil {
			return errorMsg{err}
		}
		return fileSavedMsg{m.currentFile}
	}
}

func (m *TerraformTUI) newTerraformFile() tea.Cmd {
	return func() tea.Msg {
		m.currentFile = ""
		m.editor.SetValue("")
		m.activeView = ViewEditor
		return nil
	}
}

func (m *TerraformTUI) openExternalEditor() tea.Cmd {
	if m.currentFile == "" {
		return func() tea.Msg {
			return errorMsg{fmt.Errorf("no file to edit")}
		}
	}

	return func() tea.Msg {
		cfg := config.GetEditorConfig()

		var cmd *exec.Cmd
		if len(cfg.EditorArgs) > 0 {
			args := append(cfg.EditorArgs, m.currentFile)
			cmd = exec.Command(cfg.DefaultEditor, args...)
		} else {
			cmd = exec.Command(cfg.DefaultEditor, m.currentFile)
		}

		if err := cmd.Run(); err != nil {
			return errorMsg{err}
		}

		// Reload file content
		if content, err := os.ReadFile(m.currentFile); err == nil {
			m.editor.SetValue(string(content))
		}

		return fileEditedMsg{m.currentFile}
	}
}

// Enhanced State Management Commands

func (m *TerraformTUI) loadStateResources() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}

		// Execute terraform state list to get resources
		cmd := exec.Command("terraform", "state", "list")
		cmd.Dir = m.manager.WorkingDir
		output, err := cmd.Output()
		if err != nil {
			return errorMsg{fmt.Errorf("failed to list state resources: %v", err)}
		}

		// Parse output and create StateResource objects
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		var resources []StateResource

		for _, line := range lines {
			if line == "" {
				continue
			}

			// Parse resource address (e.g., "azurerm_resource_group.main")
			parts := strings.Split(line, ".")
			if len(parts) >= 2 {
				resourceType := parts[0]
				resourceName := strings.Join(parts[1:], ".")

				resource := StateResource{
					Type:     resourceType,
					Name:     resourceName,
					Address:  line,
					Provider: strings.Split(resourceType, "_")[0], // Extract provider from type
					Status:   "ok",
				}

				// Get detailed resource information
				detailCmd := exec.Command("terraform", "state", "show", line)
				detailCmd.Dir = m.manager.WorkingDir
				if detailOutput, detailErr := detailCmd.Output(); detailErr == nil {
					// Parse terraform state show output for attributes
					resource.Attributes = parseStateShowOutput(string(detailOutput))
				}

				resources = append(resources, resource)
			}
		}

		return stateResourcesLoadedMsg{resources: resources}
	}
}

func (m *TerraformTUI) loadPlanChanges() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}

		// Execute terraform plan with JSON output
		cmd := exec.Command("terraform", "plan", "-json", "-out=tfplan")
		cmd.Dir = m.manager.WorkingDir
		output, err := cmd.Output()
		if err != nil {
			return errorMsg{fmt.Errorf("failed to generate plan: %v", err)}
		}

		// Parse JSON plan output
		changes := parseEnhancedPlanOutput(string(output))

		return planChangesLoadedMsg{changes: changes}
	}
}

func (m *TerraformTUI) loadWorkspaceInfo() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}

		// Get current workspace
		cmd := exec.Command("terraform", "workspace", "show")
		cmd.Dir = m.manager.WorkingDir
		currentOutput, err := cmd.Output()
		if err != nil {
			return errorMsg{fmt.Errorf("failed to get current workspace: %v", err)}
		}

		currentWorkspace := strings.TrimSpace(string(currentOutput))

		// List all workspaces
		cmd = exec.Command("terraform", "workspace", "list")
		cmd.Dir = m.manager.WorkingDir
		listOutput, err := cmd.Output()
		if err != nil {
			return errorMsg{fmt.Errorf("failed to list workspaces: %v", err)}
		}

		// Parse workspace list
		var workspaces []WorkspaceInfo
		lines := strings.Split(strings.TrimSpace(string(listOutput)), "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Remove asterisk from current workspace
			workspaceName := strings.TrimPrefix(line, "* ")
			workspaceName = strings.TrimSpace(workspaceName)

			workspace := WorkspaceInfo{
				Name:        workspaceName,
				Path:        m.manager.WorkingDir,
				Environment: inferEnvironment(workspaceName),
				Backend:     detectBackendType(m.manager.WorkingDir),
				Status:      getWorkspaceStatus(m.manager.WorkingDir, workspaceName),
			}

			workspaces = append(workspaces, workspace)
		}

		return workspaceInfoLoadedMsg{
			workspaces: workspaces,
			current:    currentWorkspace,
		}
	}
}

func (m *TerraformTUI) togglePlanFilter() {
	// Cycle through filter options: all -> create -> update -> delete -> all
	switch m.planViewer.filterAction {
	case "":
		m.planViewer.filterAction = "create"
	case "create":
		m.planViewer.filterAction = "update"
	case "update":
		m.planViewer.filterAction = "delete"
	case "delete":
		m.planViewer.filterAction = ""
	}

	m.status = fmt.Sprintf("Filter: %s", m.planViewer.filterAction)
}

func (m *TerraformTUI) targetResource(resourceAddress string) tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}

		// Execute terraform apply with target
		cmd := exec.Command("terraform", "apply", "-target="+resourceAddress, "-auto-approve")
		cmd.Dir = m.manager.WorkingDir
		output, err := cmd.CombinedOutput()

		success := err == nil
		operation := tfbicep.TerraformOperation{
			Command:  fmt.Sprintf("apply -target=%s", resourceAddress),
			Success:  success,
			Output:   string(output),
			Duration: time.Duration(0), // Initialize with zero duration
		}

		if err != nil {
			operation.Error = err.Error()
		}

		return operationCompletedMsg{operation: operation}
	}
}

// Workspace management functions

func (m *TerraformTUI) switchWorkspace(workspaceName string) tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}

		// Don't switch if already current
		if workspaceName == m.currentWorkspace {
			return workspaceSwitchedMsg{
				workspace: workspaceName,
				success:   true,
				message:   fmt.Sprintf("Already in workspace '%s'", workspaceName),
			}
		}

		// Execute terraform workspace select
		cmd := exec.Command("terraform", "workspace", "select", workspaceName)
		cmd.Dir = m.manager.WorkingDir
		output, err := cmd.CombinedOutput()

		if err != nil {
			return workspaceSwitchedMsg{
				workspace: workspaceName,
				success:   false,
				message:   fmt.Sprintf("Failed to switch workspace: %v", err),
				error:     err,
			}
		}

		return workspaceSwitchedMsg{
			workspace: workspaceName,
			success:   true,
			message:   fmt.Sprintf("Switched to workspace '%s'", workspaceName),
			output:    string(output),
		}
	}
}

func (m *TerraformTUI) createWorkspace(workspaceName string) tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}

		// Execute terraform workspace new
		cmd := exec.Command("terraform", "workspace", "new", workspaceName)
		cmd.Dir = m.manager.WorkingDir
		output, err := cmd.CombinedOutput()

		success := err == nil
		message := fmt.Sprintf("Created workspace '%s'", workspaceName)
		if err != nil {
			message = fmt.Sprintf("Failed to create workspace: %v", err)
		}

		return workspaceCreatedMsg{
			workspace: workspaceName,
			success:   success,
			message:   message,
			output:    string(output),
			error:     err,
		}
	}
}

func (m *TerraformTUI) deleteWorkspace(workspaceName string) tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}

		// Cannot delete current workspace
		if workspaceName == m.currentWorkspace {
			return workspaceDeletedMsg{
				workspace: workspaceName,
				success:   false,
				message:   "Cannot delete current workspace",
			}
		}

		// Execute terraform workspace delete
		cmd := exec.Command("terraform", "workspace", "delete", workspaceName)
		cmd.Dir = m.manager.WorkingDir
		output, err := cmd.CombinedOutput()

		success := err == nil
		message := fmt.Sprintf("Deleted workspace '%s'", workspaceName)
		if err != nil {
			message = fmt.Sprintf("Failed to delete workspace: %v", err)
		}

		return workspaceDeletedMsg{
			workspace: workspaceName,
			success:   success,
			message:   message,
			output:    string(output),
			error:     err,
		}
	}
}

// Variable Management Functions

func (m *TerraformTUI) loadTerraformVariables() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}

		variables := make(map[string]string)

		// Try to read terraform.tfvars file
		tfvarsPath := filepath.Join(m.manager.WorkingDir, "terraform.tfvars")
		if content, err := os.ReadFile(tfvarsPath); err == nil {
			variables = parseTerraformVariables(string(content))
		}

		// Try to read *.auto.tfvars files
		if files, err := filepath.Glob(filepath.Join(m.manager.WorkingDir, "*.auto.tfvars")); err == nil {
			for _, file := range files {
				if content, err := os.ReadFile(file); err == nil {
					autoVars := parseTerraformVariables(string(content))
					for k, v := range autoVars {
						variables[k] = v
					}
				}
			}
		}

		// Add environment-specific variables
		if m.currentEnv != "" {
			envVarsPath := filepath.Join(m.manager.WorkingDir, fmt.Sprintf("%s.tfvars", m.currentEnv))
			if content, err := os.ReadFile(envVarsPath); err == nil {
				envVars := parseTerraformVariables(string(content))
				for k, v := range envVars {
					variables[k] = v
				}
			}
		}

		return variablesLoadedMsg{
			variables: variables,
		}
	}
}

func (m *TerraformTUI) updateTerraformVariable(name, value string) tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return variableUpdatedMsg{
				name:    name,
				success: false,
				message: "No workspace selected",
			}
		}

		// Update variable in terraform.tfvars
		tfvarsPath := filepath.Join(m.manager.WorkingDir, "terraform.tfvars")

		// Read existing content
		var content string
		if existing, err := os.ReadFile(tfvarsPath); err == nil {
			content = string(existing)
		}

		// Update or add the variable
		lines := strings.Split(content, "\n")
		updated := false

		for i, line := range lines {
			if strings.HasPrefix(strings.TrimSpace(line), name+" =") {
				lines[i] = fmt.Sprintf(`%s = "%s"`, name, value)
				updated = true
				break
			}
		}

		if !updated {
			lines = append(lines, fmt.Sprintf(`%s = "%s"`, name, value))
		}

		// Write back to file
		newContent := strings.Join(lines, "\n")
		if err := os.WriteFile(tfvarsPath, []byte(newContent), 0644); err != nil {
			return variableUpdatedMsg{
				name:    name,
				success: false,
				message: fmt.Sprintf("Failed to write variable: %v", err),
				error:   err,
			}
		}

		return variableUpdatedMsg{
			name:    name,
			value:   value,
			success: true,
			message: fmt.Sprintf("Variable '%s' updated successfully", name),
		}
	}
}

func (m *TerraformTUI) loadTerraformOutputs() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}

		// Execute terraform output -json
		cmd := exec.Command("terraform", "output", "-json")
		cmd.Dir = m.manager.WorkingDir
		output, err := cmd.Output()

		if err != nil {
			return outputsLoadedMsg{
				outputs: make(map[string]interface{}),
				content: "No outputs available or terraform not initialized",
			}
		}

		// Parse JSON output
		var outputs map[string]interface{}
		if err := json.Unmarshal(output, &outputs); err != nil {
			return outputsLoadedMsg{
				outputs: make(map[string]interface{}),
				content: fmt.Sprintf("Failed to parse outputs: %v", err),
			}
		}

		// Format outputs for display
		var contentLines []string
		contentLines = append(contentLines, "Terraform Outputs:")
		contentLines = append(contentLines, "")

		for name, output := range outputs {
			if outputMap, ok := output.(map[string]interface{}); ok {
				value := outputMap["value"]
				sensitive := false
				if s, exists := outputMap["sensitive"]; exists {
					sensitive = s.(bool)
				}

				if sensitive {
					contentLines = append(contentLines, fmt.Sprintf("%s = <sensitive>", name))
				} else {
					valueStr := fmt.Sprintf("%v", value)
					contentLines = append(contentLines, fmt.Sprintf("%s = %s", name, valueStr))
				}
			}
		}

		if len(outputs) == 0 {
			contentLines = append(contentLines, "No outputs defined")
		}

		return outputsLoadedMsg{
			outputs: outputs,
			content: strings.Join(contentLines, "\n"),
		}
	}
}

// Helper functions for parsing Terraform output

func parseStateShowOutput(output string) map[string]interface{} {
	attributes := make(map[string]interface{})

	// Simple parsing - in a real implementation, you'd use terraform's JSON output
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, " = ") {
			parts := strings.SplitN(line, " = ", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				attributes[key] = value
			}
		}
	}

	return attributes
}

// Enhanced JSON Plan Parsing
func parseEnhancedPlanOutput(jsonOutput string) []PlanChange {
	var changes []PlanChange

	// Parse JSON plan output properly
	var planData struct {
		ResourceChanges []struct {
			Address      string `json:"address"`
			Type         string `json:"type"`
			Name         string `json:"name"`
			ProviderName string `json:"provider_name"`
			Change       struct {
				Actions []string               `json:"actions"`
				Before  map[string]interface{} `json:"before"`
				After   map[string]interface{} `json:"after"`
				Reason  string                 `json:"action_reason"`
			} `json:"change"`
		} `json:"resource_changes"`
	}

	if err := json.Unmarshal([]byte(jsonOutput), &planData); err != nil {
		// Fallback to simplified parsing if JSON parsing fails
		return []PlanChange{
			{
				Action:   "unknown",
				Resource: "parsing_error",
				Type:     "error",
				Name:     "json_parse_failed",
				Impact:   "unknown",
				Reason:   fmt.Sprintf("Failed to parse JSON: %v", err),
			},
		}
	}

	for _, resourceChange := range planData.ResourceChanges {
		if len(resourceChange.Change.Actions) == 0 {
			continue
		}

		action := resourceChange.Change.Actions[0]
		if len(resourceChange.Change.Actions) > 1 {
			// Handle complex actions like ["delete", "create"] for replace
			if contains(resourceChange.Change.Actions, "delete") && contains(resourceChange.Change.Actions, "create") {
				action = "replace"
			}
		}

		// Determine impact level
		impact := determineChangeImpact(action, resourceChange.Type, resourceChange.Change.Before, resourceChange.Change.After)

		change := PlanChange{
			Action:    action,
			Resource:  resourceChange.Address,
			Type:      resourceChange.Type,
			Name:      resourceChange.Name,
			Before:    resourceChange.Change.Before,
			After:     resourceChange.Change.After,
			Reason:    resourceChange.Change.Reason,
			Sensitive: containsSensitiveData(resourceChange.Change.After),
			Impact:    impact,
		}

		changes = append(changes, change)
	}

	return changes
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// Helper function to determine change impact
func determineChangeImpact(action, resourceType string, before, after map[string]interface{}) string {
	switch action {
	case "create":
		// New resources are generally low impact unless they're critical infrastructure
		if isCriticalResource(resourceType) {
			return "medium"
		}
		return "low"
	case "delete":
		// Deletions are always high impact
		return "high"
	case "replace":
		// Replacements are high impact as they involve downtime
		return "high"
	case "update":
		// Updates vary based on what's changing
		if hasHighImpactChanges(before, after, resourceType) {
			return "high"
		}
		if hasMediumImpactChanges(before, after, resourceType) {
			return "medium"
		}
		return "low"
	default:
		return "unknown"
	}
}

// Helper function to check if resource type is critical
func isCriticalResource(resourceType string) bool {
	criticalTypes := []string{
		"azurerm_virtual_machine",
		"azurerm_kubernetes_cluster",
		"azurerm_sql_database",
		"azurerm_storage_account",
		"azurerm_application_gateway",
		"azurerm_firewall",
	}

	for _, criticalType := range criticalTypes {
		if resourceType == criticalType {
			return true
		}
	}
	return false
}

// Helper function to check for high impact changes
func hasHighImpactChanges(before, after map[string]interface{}, resourceType string) bool {
	// Check for size/sku changes
	if before["size"] != after["size"] || before["sku"] != after["sku"] {
		return true
	}

	// Check for location changes (always high impact)
	if before["location"] != after["location"] {
		return true
	}

	// Resource-specific high impact checks
	switch resourceType {
	case "azurerm_virtual_machine":
		return before["vm_size"] != after["vm_size"]
	case "azurerm_kubernetes_cluster":
		return before["node_count"] != after["node_count"] || before["vm_size"] != after["vm_size"]
	}

	return false
}

// Helper function to check for medium impact changes
func hasMediumImpactChanges(before, after map[string]interface{}, resourceType string) bool {
	// Check for tag changes
	if !mapsEqual(getMapFromInterface(before["tags"]), getMapFromInterface(after["tags"])) {
		return true
	}

	// Check for network-related changes
	if before["subnet_id"] != after["subnet_id"] || before["public_ip"] != after["public_ip"] {
		return true
	}

	return false
}

// Helper function to check if data contains sensitive information
func containsSensitiveData(data map[string]interface{}) bool {
	sensitiveKeys := []string{"password", "secret", "key", "token", "credential"}

	for key := range data {
		keyLower := strings.ToLower(key)
		for _, sensitiveKey := range sensitiveKeys {
			if strings.Contains(keyLower, sensitiveKey) {
				return true
			}
		}
	}
	return false
}

// Helper functions for map comparison
func getMapFromInterface(i interface{}) map[string]interface{} {
	if m, ok := i.(map[string]interface{}); ok {
		return m
	}
	return make(map[string]interface{})
}

func mapsEqual(a, b map[string]interface{}) bool {
	if len(a) != len(b) {
		return false
	}

	for k, v := range a {
		if b[k] != v {
			return false
		}
	}
	return true
}

// Helper functions for parsing Terraform variables
func parseTerraformVariables(content string) map[string]string {
	variables := make(map[string]string)
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "//") {
			continue
		}

		// Parse variable assignment: var_name = "value"
		if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])

			// Remove quotes from value
			value = strings.Trim(value, `"'`)

			variables[key] = value
		}
	}

	return variables
}

func inferEnvironment(workspaceName string) string {
	workspaceLower := strings.ToLower(workspaceName)
	if strings.Contains(workspaceLower, "prod") || strings.Contains(workspaceLower, "production") {
		return "prod"
	}
	if strings.Contains(workspaceLower, "stag") || strings.Contains(workspaceLower, "staging") {
		return "staging"
	}
	if strings.Contains(workspaceLower, "dev") || strings.Contains(workspaceLower, "development") {
		return "dev"
	}
	return "dev" // Default to dev
}

// Message types for the enhanced features

type stateResourcesLoadedMsg struct {
	resources []StateResource
}

type planChangesLoadedMsg struct {
	changes []PlanChange
}

type workspaceInfoLoadedMsg struct {
	workspaces []WorkspaceInfo
	current    string
}

type workspaceSwitchedMsg struct {
	workspace string
	success   bool
	message   string
	output    string
	error     error
}

type workspaceCreatedMsg struct {
	workspace string
	success   bool
	message   string
	output    string
	error     error
}

type workspaceDeletedMsg struct {
	workspace string
	success   bool
	message   string
	output    string
	error     error
}

// Variable management message types
type variablesLoadedMsg struct {
	variables map[string]string
}

type variableUpdatedMsg struct {
	name    string
	value   string
	success bool
	message string
	error   error
}

// Output values message types
type outputsLoadedMsg struct {
	outputs map[string]interface{}
	content string
}

// Message types
type errorMsg struct{ error }
type templatesLoadedMsg struct{ items []list.Item }
type workspacesLoadedMsg struct{ items []list.Item }
type fileLoadedMsg struct{ path, content string }
type fileSavedMsg struct{ path string }
type fileEditedMsg struct{ path string }
type operationCompletedMsg struct{ operation tfbicep.TerraformOperation }

// Handle messages
func (m *TerraformTUI) handleMessages(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case errorMsg:
		m.errorMsg = msg.Error()
		m.status = "Error"

	case templatesLoadedMsg:
		m.templates.SetItems(msg.items)
		m.status = "Templates loaded"

	case workspacesLoadedMsg:
		m.workspaces.SetItems(msg.items)
		m.status = "Workspaces loaded"

	case fileLoadedMsg:
		m.currentFile = msg.path
		m.editor.SetValue(msg.content)
		m.activeView = ViewEditor
		m.status = fmt.Sprintf("Loaded %s", filepath.Base(msg.path))

	case fileSavedMsg:
		m.status = fmt.Sprintf("Saved %s", filepath.Base(msg.path))

	case fileEditedMsg:
		m.status = fmt.Sprintf("Edited %s", filepath.Base(msg.path))

	case operationCompletedMsg:
		m.operations = append(m.operations, msg.operation)
		if msg.operation.Success {
			m.status = fmt.Sprintf("✓ %s completed", msg.operation.Command)
		} else {
			m.status = fmt.Sprintf("✗ %s failed", msg.operation.Command)
			m.errorMsg = msg.operation.Error
		}

		// Show operation result in popup
		m.showPopup = true
		m.popupTitle = msg.operation.Command
		if msg.operation.Success {
			m.popupContent = fmt.Sprintf("Success!\n\nDuration: %s\n\nOutput:\n%s",
				msg.operation.Duration,
				msg.operation.Output)
		} else {
			m.popupContent = fmt.Sprintf("Failed!\n\nDuration: %s\n\nError:\n%s",
				msg.operation.Duration,
				msg.operation.Error)
		}
	}

	return m, nil
}
