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

func (m *TerraformTUI) renderVarEditorView() string {
	// Clean, frameless styling for consistency
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Padding(1, 2)

	if m.activeView == ViewVarEditor {
		style = style.Foreground(lipgloss.Color("#FF5F87"))
	}

	var content []string
	content = append(content, lipgloss.NewStyle().Bold(true).Render("Interactive Variable Editor"))
	content = append(content, "")

	if len(m.variableEditor.variables) == 0 {
		content = append(content, "No variables loaded")
		content = append(content, "Press 'v' to load variables first")
	} else {
		// Show editing instructions
		if m.variableEditor.editMode {
			content = append(content, fmt.Sprintf("Editing: %s", m.variableEditor.editingVar))
			content = append(content, fmt.Sprintf("Original: %s", m.variableEditor.originalValue))
			content = append(content, fmt.Sprintf("New Value: %s", m.variableEditor.editingValue))
			content = append(content, "")
			content = append(content, "Enter: save | Esc: cancel | Type to edit")
		} else {
			content = append(content, fmt.Sprintf("Variables: %d", len(m.variableEditor.variables)))
			content = append(content, "")

			// List variables with selection indicator
			i := 0
			for name, value := range m.variableEditor.variables {
				prefix := "  "
				if i == m.variableEditor.selectedIndex {
					prefix = "▶ "
				}

				line := fmt.Sprintf("%s%s = %s", prefix, name, value)
				if i == m.variableEditor.selectedIndex {
					line = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5F87")).Render(line)
				}

				content = append(content, line)
				i++
			}

			content = append(content, "")
			content = append(content, "e: edit | ↑/↓: navigate | Esc: back")
		}
	}

	return style.Render(strings.Join(content, "\n"))
}

// Interactive Variable Editing Methods

func (m *TerraformTUI) startVariableEdit() tea.Cmd {
	if len(m.variableEditor.variables) == 0 {
		return nil
	}

	// Get the selected variable
	i := 0
	for name, value := range m.variableEditor.variables {
		if i == m.variableEditor.selectedIndex {
			m.variableEditor.editMode = true
			m.variableEditor.editingVar = name
			m.variableEditor.editingValue = value
			m.variableEditor.originalValue = value
			break
		}
		i++
	}

	return nil
}

func (m *TerraformTUI) saveVariableEdit() tea.Cmd {
	if !m.variableEditor.editMode {
		return nil
	}

	// Update the variable
	name := m.variableEditor.editingVar
	value := m.variableEditor.editingValue

	// Exit edit mode
	m.variableEditor.editMode = false
	m.variableEditor.editingVar = ""
	m.variableEditor.editingValue = ""
	m.variableEditor.originalValue = ""

	// Update the variable in the file and reload
	return tea.Batch(
		m.updateTerraformVariable(name, value),
		func() tea.Msg { return variableEditCompletedMsg{name: name, value: value} },
	)
}

func (m *TerraformTUI) cancelVariableEdit() {
	m.variableEditor.editMode = false
	m.variableEditor.editingVar = ""
	m.variableEditor.editingValue = ""
	m.variableEditor.originalValue = ""
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

// Enhanced Error Handling Types and Functions

type TerraformError struct {
	Operation   string
	Message     string
	Details     string
	Severity    string // "low", "medium", "high", "critical"
	Suggestions []string
}

func (e TerraformError) Error() string {
	return fmt.Sprintf("[%s] %s: %s", e.Severity, e.Operation, e.Message)
}

// Enhanced error handler for Terraform operations
func (m *TerraformTUI) handleTerraformError(operation string, err error) TerraformError {
	if err == nil {
		return TerraformError{}
	}

	errStr := err.Error()
	terraformErr := TerraformError{
		Operation: operation,
		Message:   errStr,
		Severity:  "medium",
	}

	// Analyze error patterns and provide suggestions
	switch {
	case strings.Contains(errStr, "terraform not found") || strings.Contains(errStr, "command not found"):
		terraformErr.Severity = "critical"
		terraformErr.Details = "Terraform CLI is not installed or not in PATH"
		terraformErr.Suggestions = []string{
			"Install Terraform CLI from https://terraform.io/downloads",
			"Ensure terraform binary is in your PATH",
			"Verify installation with: terraform version",
		}

	case strings.Contains(errStr, "not initialized"):
		terraformErr.Severity = "high"
		terraformErr.Details = "Terraform working directory not initialized"
		terraformErr.Suggestions = []string{
			"Run 'terraform init' to initialize the working directory",
			"Ensure .terraform directory exists",
			"Check backend configuration",
		}

	case strings.Contains(errStr, "state lock"):
		terraformErr.Severity = "high"
		terraformErr.Details = "Terraform state is locked by another process"
		terraformErr.Suggestions = []string{
			"Wait for the current operation to complete",
			"Force unlock if the lock is stale: terraform force-unlock <lock-id>",
			"Check for running Terraform processes",
		}

	case strings.Contains(errStr, "no configuration files"):
		terraformErr.Severity = "high"
		terraformErr.Details = "No Terraform configuration files found"
		terraformErr.Suggestions = []string{
			"Create .tf files in the current directory",
			"Navigate to a directory with Terraform configuration",
			"Use templates to create a new configuration",
		}

	case strings.Contains(errStr, "provider"):
		terraformErr.Severity = "medium"
		terraformErr.Details = "Provider configuration or authentication issue"
		terraformErr.Suggestions = []string{
			"Check provider configuration",
			"Verify authentication credentials",
			"Run 'terraform init' to download providers",
		}

	case strings.Contains(errStr, "invalid syntax") || strings.Contains(errStr, "syntax error"):
		terraformErr.Severity = "medium"
		terraformErr.Details = "Terraform configuration syntax error"
		terraformErr.Suggestions = []string{
			"Run 'terraform validate' to check syntax",
			"Use 'terraform fmt' to format configuration",
			"Check for missing quotes, brackets, or commas",
		}

	case strings.Contains(errStr, "permission denied"):
		terraformErr.Severity = "high"
		terraformErr.Details = "Insufficient permissions"
		terraformErr.Suggestions = []string{
			"Check file permissions in the working directory",
			"Verify cloud provider permissions",
			"Ensure terraform can write to .terraform directory",
		}

	case strings.Contains(errStr, "network") || strings.Contains(errStr, "timeout"):
		terraformErr.Severity = "medium"
		terraformErr.Details = "Network connectivity issue"
		terraformErr.Suggestions = []string{
			"Check internet connection",
			"Verify firewall settings",
			"Try again after a few moments",
		}

	default:
		terraformErr.Severity = "low"
		terraformErr.Details = "Unknown error occurred"
		terraformErr.Suggestions = []string{
			"Check Terraform logs for more details",
			"Try running the command directly in terminal",
			"Verify working directory and configuration",
		}
	}

	return terraformErr
}

// Format error for display in TUI
func (m *TerraformTUI) formatErrorForDisplay(terraformErr TerraformError) string {
	if terraformErr.Message == "" {
		return ""
	}

	var lines []string

	// Severity indicator
	severityIcon := "⚠️"
	switch terraformErr.Severity {
	case "critical":
		severityIcon = "🚨"
	case "high":
		severityIcon = "🔴"
	case "medium":
		severityIcon = "🟡"
	case "low":
		severityIcon = "🔵"
	}

	lines = append(lines, fmt.Sprintf("%s %s Error in %s", severityIcon, strings.Title(terraformErr.Severity), terraformErr.Operation))
	lines = append(lines, "")
	lines = append(lines, "Message:")
	lines = append(lines, terraformErr.Message)

	if terraformErr.Details != "" {
		lines = append(lines, "")
		lines = append(lines, "Details:")
		lines = append(lines, terraformErr.Details)
	}

	if len(terraformErr.Suggestions) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Suggestions:")
		for _, suggestion := range terraformErr.Suggestions {
			lines = append(lines, fmt.Sprintf("• %s", suggestion))
		}
	}

	return strings.Join(lines, "\n")
}

// Enhanced progress indicator with better visual styling
func (m *TerraformTUI) renderEnhancedProgressIndicator(progress ParseProgressMsg) string {
	if progress.total == 0 {
		return ""
	}

	percentage := float64(progress.current) / float64(progress.total) * 100
	filledBlocks := int(percentage / 5)

	// Create gradient progress bar
	progressBar := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Render(strings.Repeat("█", filledBlocks)) +
		lipgloss.NewStyle().
			Foreground(lipgloss.Color("#44475A")).
			Render(strings.Repeat("░", 20-filledBlocks))

	// Stage indicator with icon
	stageIcon := "⚡"
	switch {
	case strings.Contains(progress.stage, "Loading"):
		stageIcon = "📂"
	case strings.Contains(progress.stage, "Parsing"):
		stageIcon = "⚙️"
	case strings.Contains(progress.stage, "Generating"):
		stageIcon = "🔨"
	case strings.Contains(progress.stage, "Analyzing"):
		stageIcon = "🔍"
	}

	progressStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#6272A4")).
		Padding(1, 2).
		Margin(1, 0)

	content := fmt.Sprintf("%s %s\n\n[%s] %.1f%%\n\nETA: %s",
		stageIcon, progress.stage, progressBar, percentage, m.estimateETA(progress))

	return progressStyle.Render(content)
}

// Estimate completion time for progress indicator
func (m *TerraformTUI) estimateETA(progress ParseProgressMsg) string {
	if progress.current == 0 || !m.progressIndicator.isActive {
		return "calculating..."
	}

	elapsed := time.Since(m.progressIndicator.lastUpdate)
	if elapsed.Seconds() < 1 {
		return "< 1s"
	}

	rate := float64(progress.current) / elapsed.Seconds()
	if rate <= 0 {
		return "unknown"
	}

	remaining := float64(progress.total-progress.current) / rate
	if remaining < 60 {
		return fmt.Sprintf("%.0fs", remaining)
	}
	return fmt.Sprintf("%.1fm", remaining/60)
}

// Enhanced error display with better styling and categorization
func (m *TerraformTUI) renderEnhancedError(terraformErr TerraformError) string {
	if terraformErr.Message == "" {
		return ""
	}

	// Choose color scheme based on severity
	var borderColor, titleColor, iconColor lipgloss.Color
	var severityIcon string

	switch terraformErr.Severity {
	case "critical":
		borderColor = lipgloss.Color("#FF5555")
		titleColor = lipgloss.Color("#FF5555")
		iconColor = lipgloss.Color("#FF5555")
		severityIcon = "🚨"
	case "high":
		borderColor = lipgloss.Color("#FFB86C")
		titleColor = lipgloss.Color("#FFB86C")
		iconColor = lipgloss.Color("#FFB86C")
		severityIcon = "⚠️"
	case "medium":
		borderColor = lipgloss.Color("#F1FA8C")
		titleColor = lipgloss.Color("#F1FA8C")
		iconColor = lipgloss.Color("#F1FA8C")
		severityIcon = "⚡"
	default:
		borderColor = lipgloss.Color("#8BE9FD")
		titleColor = lipgloss.Color("#8BE9FD")
		iconColor = lipgloss.Color("#8BE9FD")
		severityIcon = "ℹ️"
	}

	errorStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Margin(1, 0).
		Width(80)

	titleStyle := lipgloss.NewStyle().
		Foreground(titleColor).
		Bold(true)

	iconStyle := lipgloss.NewStyle().
		Foreground(iconColor)

	var content strings.Builder

	// Title with icon
	content.WriteString(titleStyle.Render(fmt.Sprintf("%s %s Error in %s",
		iconStyle.Render(severityIcon),
		strings.Title(terraformErr.Severity),
		terraformErr.Operation)))
	content.WriteString("\n\n")

	// Error message
	content.WriteString("📝 Message:\n")
	content.WriteString(terraformErr.Message)

	// Details if available
	if terraformErr.Details != "" {
		content.WriteString("\n\n🔍 Details:\n")
		content.WriteString(terraformErr.Details)
	}

	// Suggestions if available
	if len(terraformErr.Suggestions) > 0 {
		content.WriteString("\n\n💡 Suggestions:\n")
		for i, suggestion := range terraformErr.Suggestions {
			content.WriteString(fmt.Sprintf("%d. %s\n", i+1, suggestion))
		}
	}

	return errorStyle.Render(content.String())
}

// Enhanced variable editor with syntax highlighting
func (m *TerraformTUI) renderEnhancedVarEditor() string {
	baseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(1, 2)

	if m.activeView == ViewVarEditor {
		baseStyle = baseStyle.Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF79C6"))
	}

	var content []string

	// Header with enhanced styling
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Bold(true)
	content = append(content, headerStyle.Render("🔧 Interactive Variable Editor"))
	content = append(content, "")

	if len(m.variableEditor.variables) == 0 {
		noVarsStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true)
		content = append(content, noVarsStyle.Render("No variables loaded"))
		content = append(content, "Press 'v' to load variables first")
	} else {
		if m.variableEditor.editMode {
			// Edit mode with enhanced styling
			editHeaderStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFB86C")).
				Bold(true)
			content = append(content, editHeaderStyle.Render("✏️ Editing Variable"))
			content = append(content, "")

			varNameStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#8BE9FD")).
				Bold(true)
			content = append(content, fmt.Sprintf("Variable: %s",
				varNameStyle.Render(m.variableEditor.editingVar)))

			originalStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272A4"))
			content = append(content, fmt.Sprintf("Original: %s",
				originalStyle.Render(m.variableEditor.originalValue)))

			newValueStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#50FA7B")).
				Background(lipgloss.Color("#44475A")).
				Padding(0, 1)
			content = append(content, fmt.Sprintf("New Value: %s",
				newValueStyle.Render(m.variableEditor.editingValue)))

			content = append(content, "")
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272A4"))
			content = append(content, helpStyle.Render("Enter: save | Esc: cancel | Type to edit"))
		} else {
			// List mode with enhanced styling
			content = append(content, fmt.Sprintf("Variables: %s",
				lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")).Render(fmt.Sprintf("%d", len(m.variableEditor.variables)))))
			content = append(content, "")

			// Enhanced variable list
			i := 0
			for name, value := range m.variableEditor.variables {
				isSelected := i == m.variableEditor.selectedIndex

				var nameStyle, valueStyle, equalsStyle lipgloss.Style
				if isSelected {
					nameStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#50FA7B")).
						Bold(true)
					equalsStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#FF79C6"))
					valueStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#F1FA8C"))
				} else {
					nameStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#8BE9FD"))
					equalsStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#6272A4"))
					valueStyle = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#F8F8F2"))
				}

				prefix := "  "
				if isSelected {
					prefix = lipgloss.NewStyle().
						Foreground(lipgloss.Color("#FF79C6")).
						Render("▶ ")
				}

				line := fmt.Sprintf("%s%s %s %s",
					prefix,
					nameStyle.Render(name),
					equalsStyle.Render("="),
					valueStyle.Render(fmt.Sprintf("\"%s\"", value)))

				content = append(content, line)
				i++
			}

			content = append(content, "")
			helpStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272A4"))
			content = append(content, helpStyle.Render("e: edit | ↑/↓: navigate | Esc: back"))
		}
	}

	return baseStyle.Render(strings.Join(content, "\n"))
}

// Enhanced plan viewer with better action icons and colors
func (m *TerraformTUI) renderEnhancedPlanViewer() string {
	baseStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F8F8F2")).
		Padding(1, 2)

	if m.activeView == ViewPlanViewer {
		baseStyle = baseStyle.Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF79C6"))
	}

	var content []string

	// Enhanced header
	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#50FA7B")).
		Bold(true)
	content = append(content, headerStyle.Render("📊 Terraform Plan Analysis"))
	content = append(content, "")

	if len(m.planViewer.changes) == 0 {
		noChangesStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Italic(true)
		content = append(content, noChangesStyle.Render("No plan changes found"))
		content = append(content, "Press 'p' to load plan changes")
	} else {
		// Enhanced filter summary
		filterText := "All changes"
		if m.planViewer.filterAction != "" {
			filterText = fmt.Sprintf("Filter: %s", m.planViewer.filterAction)
		}

		summaryStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FA8C"))
		content = append(content, summaryStyle.Render(fmt.Sprintf("Changes: %d | %s", len(m.planViewer.changes), filterText)))
		content = append(content, "")

		// Enhanced action summary with icons and colors
		actionCounts := make(map[string]int)
		for _, change := range m.planViewer.changes {
			actionCounts[change.Action]++
		}

		var summary []string
		actionStyles := map[string]lipgloss.Style{
			"create":  lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")),
			"update":  lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")),
			"delete":  lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")),
			"replace": lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")),
		}

		for action, count := range actionCounts {
			icon := getEnhancedActionIcon(action)
			style := actionStyles[action]
			if style.GetForeground() == nil {
				style = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
			}
			summary = append(summary, style.Render(fmt.Sprintf("%s %d %s", icon, count, action)))
		}
		content = append(content, strings.Join(summary, " | "))
		content = append(content, "")

		// Enhanced change list
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
			isSelected := i == m.selectedChange

			icon := getEnhancedActionIcon(change.Action)
			actionStyle := actionStyles[change.Action]
			if actionStyle.GetForeground() == nil {
				actionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4"))
			}

			var resourceStyle lipgloss.Style
			if isSelected {
				resourceStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FF79C6")).
					Bold(true)
			} else {
				resourceStyle = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#8BE9FD"))
			}

			prefix := "  "
			if isSelected {
				prefix = lipgloss.NewStyle().
					Foreground(lipgloss.Color("#FF79C6")).
					Render("▶ ")
			}

			impactIcon := getImpactIcon(change.Impact)
			line := fmt.Sprintf("%s%s %s %s %s",
				prefix,
				actionStyle.Render(icon),
				resourceStyle.Render(change.Resource),
				lipgloss.NewStyle().Foreground(lipgloss.Color("#6272A4")).Render(fmt.Sprintf("(%s)", change.Action)),
				impactIcon)

			content = append(content, line)

			// Enhanced details for selected item
			if m.showPlanDetails && isSelected {
				detailStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color("#6272A4")).
					MarginLeft(4)

				if change.Reason != "" {
					content = append(content, detailStyle.Render(fmt.Sprintf("Reason: %s", change.Reason)))
				}
				content = append(content, detailStyle.Render(fmt.Sprintf("Impact: %s %s", getImpactIcon(change.Impact), change.Impact)))

				if change.Sensitive {
					sensitiveStyle := lipgloss.NewStyle().
						Foreground(lipgloss.Color("#FF5555")).
						Bold(true)
					content = append(content, detailStyle.Render(sensitiveStyle.Render("⚠️ Contains sensitive data")))
				}
			}
		}

		content = append(content, "")
		helpStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))
		content = append(content, helpStyle.Render("f: filter toggle | a: approval mode | t: target resource | d: toggle details"))
	}

	return baseStyle.Render(strings.Join(content, "\n"))
}

// Enhanced action icons with better visual representation
func getEnhancedActionIcon(action string) string {
	switch action {
	case "create":
		return "🟢"
	case "update":
		return "🟡"
	case "delete":
		return "🔴"
	case "replace":
		return "🟠"
	default:
		return "❓"
	}
}

// Impact icons for better visual feedback
func getImpactIcon(impact string) string {
	switch impact {
	case "high":
		return "🔥"
	case "medium":
		return "⚡"
	case "low":
		return "💡"
	default:
		return "❓"
	}
}

// Basic TUI command methods implementation

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
			terraformErr := m.handleTerraformError("init", err)
			m.status = m.formatErrorForDisplay(terraformErr)
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
			terraformErr := m.handleTerraformError("plan", err)
			m.status = m.formatErrorForDisplay(terraformErr)
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
			terraformErr := m.handleTerraformError("apply", err)
			m.status = m.formatErrorForDisplay(terraformErr)
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
			terraformErr := m.handleTerraformError("destroy", err)
			m.status = m.formatErrorForDisplay(terraformErr)
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
			terraformErr := m.handleTerraformError("validate", err)
			m.status = m.formatErrorForDisplay(terraformErr)
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
			terraformErr := m.handleTerraformError("format", err)
			m.status = m.formatErrorForDisplay(terraformErr)
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

// Helper method for variable updates
func (m *TerraformTUI) updateTerraformVariable(name, value string) tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no terraform manager available")}
		}

		workingDir := m.manager.WorkingDir
		tfvarsPath := filepath.Join(workingDir, "terraform.tfvars")

		// Read existing file or create new one
		variables := make(map[string]string)
		if content, err := os.ReadFile(tfvarsPath); err == nil {
			// Temporary simple parsing until helper functions are fixed
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && strings.Contains(line, "=") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(strings.Trim(parts[1], `"'`))
						variables[key] = value
					}
				}
			}
		}

		// Update the variable
		variables[name] = value

		// Write back to file
		var content strings.Builder
		for k, v := range variables {
			content.WriteString(fmt.Sprintf("%s = \"%s\"\n", k, v))
		}

		if err := os.WriteFile(tfvarsPath, []byte(content.String()), 0644); err != nil {
			return errorMsg{err}
		}

		return variableUpdatedMsg{
			name:    name,
			value:   value,
			success: true,
			message: "Variable updated successfully",
		}
	}
}

// Enhanced feature methods implementation

// loadStateResources loads Terraform state resources
func (m *TerraformTUI) loadStateResources() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no terraform manager available")}
		}

		// Get state list from terraform manager
		operation, err := m.manager.StateList()
		if err != nil {
			return errorMsg{fmt.Errorf("failed to load state: %v", err)}
		}

		if !operation.Success {
			return errorMsg{fmt.Errorf("terraform state list failed: %s", operation.Error)}
		}

		// Parse state list output
		var resources []StateResource
		lines := strings.Split(operation.Output, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			// Parse terraform state list output format
			parts := strings.Split(line, ".")
			if len(parts) >= 2 {
				resourceType := parts[0]
				resourceName := strings.Join(parts[1:], ".")

				resource := StateResource{
					Type:     resourceType,
					Name:     resourceName,
					Address:  line,
					Provider: strings.Split(resourceType, "_")[0], // Simple provider extraction
					Status:   "ok",                                // Default status
					Tainted:  false,
				}

				// Check if resource is tainted
				if strings.Contains(line, "(tainted)") {
					resource.Tainted = true
					resource.Status = "tainted"
				}

				resources = append(resources, resource)
			}
		}

		return stateResourcesLoadedMsg{resources: resources}
	}
}

// loadStateResourcesWithProgress loads state resources with progress indicators
func (m *TerraformTUI) loadStateResourcesWithProgress() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			return ParseProgressMsg{current: 0, total: 100, stage: "Loading state..."}
		},
		func() tea.Msg {
			return ParseProgressMsg{current: 25, total: 100, stage: "Parsing resources..."}
		},
		func() tea.Msg {
			return ParseProgressMsg{current: 75, total: 100, stage: "Processing dependencies..."}
		},
		m.loadStateResources(),
		func() tea.Msg {
			return ParseProgressMsg{current: 100, total: 100, stage: "Complete"}
		},
	)
}

// loadPlanChanges loads Terraform plan changes
func (m *TerraformTUI) loadPlanChanges() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no terraform manager available")}
		}

		// Get plan JSON from terraform manager
		planResult, err := m.manager.PlanJSON()
		if err != nil {
			return errorMsg{fmt.Errorf("failed to load plan: %v", err)}
		}

		// Parse the plan JSON resource changes
		var changes []PlanChange
		for _, rc := range planResult.ResourceChanges {
			action := "unknown"
			if len(rc.Change.Actions) > 0 {
				action = rc.Change.Actions[0]
			}

			change := PlanChange{
				Action:    action,
				Resource:  rc.Address,
				Type:      rc.Type,
				Name:      rc.Name,
				Before:    rc.Change.Before.(map[string]interface{}),
				After:     rc.Change.After.(map[string]interface{}),
				Reason:    "",       // Would need to be determined from change details
				Sensitive: false,    // Would need to be determined from plan output
				Impact:    "medium", // Default impact, could be enhanced
			}

			changes = append(changes, change)
		}

		return planChangesLoadedMsg{changes: changes}
	}
}

// loadPlanChangesWithProgress loads plan changes with progress indicators
func (m *TerraformTUI) loadPlanChangesWithProgress() tea.Cmd {
	return tea.Sequence(
		func() tea.Msg {
			return ParseProgressMsg{current: 0, total: 100, stage: "Loading plan..."}
		},
		func() tea.Msg {
			return ParseProgressMsg{current: 33, total: 100, stage: "Parsing changes..."}
		},
		func() tea.Msg {
			return ParseProgressMsg{current: 66, total: 100, stage: "Analyzing impact..."}
		},
		m.loadPlanChanges(),
		func() tea.Msg {
			return ParseProgressMsg{current: 100, total: 100, stage: "Complete"}
		},
	)
}

// loadWorkspaceInfo loads workspace information
func (m *TerraformTUI) loadWorkspaceInfo() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no terraform manager available")}
		}

		// Get workspace list using the tfbicep package
		operation, err := tfbicep.TerraformWorkspace(m.manager.WorkingDir, "list", []string{})
		if err != nil {
			return errorMsg{fmt.Errorf("failed to load workspaces: %v", err)}
		}

		if !operation.Success {
			return errorMsg{fmt.Errorf("terraform workspace list failed: %s", operation.Error)}
		}

		// Parse workspace list output
		var workspaces []WorkspaceInfo
		currentWorkspace := "default"
		lines := strings.Split(operation.Output, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}

			workspace := WorkspaceInfo{
				Status: "clean",
			}

			// Check if this is the current workspace (marked with *)
			if strings.HasPrefix(line, "*") {
				workspace.Name = strings.TrimSpace(strings.TrimPrefix(line, "*"))
				workspace.Status = "current"
				currentWorkspace = workspace.Name
			} else {
				workspace.Name = line
			}

			// Detect environment from workspace name
			if strings.Contains(workspace.Name, "dev") {
				workspace.Environment = "dev"
			} else if strings.Contains(workspace.Name, "staging") || strings.Contains(workspace.Name, "stage") {
				workspace.Environment = "staging"
			} else if strings.Contains(workspace.Name, "prod") || strings.Contains(workspace.Name, "production") {
				workspace.Environment = "prod"
			} else {
				workspace.Environment = "dev"
			}

			// Detect backend type for the workspace
			if m.manager.WorkingDir != "" {
				workspace.Backend = detectBackendType(m.manager.WorkingDir)
			}

			workspaces = append(workspaces, workspace)
		}

		return workspaceInfoLoadedMsg{workspaces: workspaces, current: currentWorkspace}
	}
}

// togglePlanFilter toggles the plan filter between different action types
func (m *TerraformTUI) togglePlanFilter() {
	filters := []string{"", "create", "update", "delete", "replace"}
	current := ""
	for i, filter := range filters {
		if filter == m.planViewer.filterAction {
			current = filters[(i+1)%len(filters)]
			break
		}
	}
	if current == "" && m.planViewer.filterAction == "" {
		current = "create"
	}
	m.planViewer.filterAction = current
}

// targetResource sets a target resource for terraform operations
func (m *TerraformTUI) targetResource(resourceAddress string) tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no terraform manager available")}
		}

		// Set target resource for future operations
		m.status = fmt.Sprintf("Targeting resource: %s", resourceAddress)

		// This would typically be used with terraform plan/apply -target=resource
		// For now, just update the status to show the targeting
		return nil
	}
}

// loadTerraformVariables loads Terraform variables from various sources
func (m *TerraformTUI) loadTerraformVariables() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no terraform manager available")}
		}

		variables := make(map[string]string)
		workingDir := m.manager.WorkingDir

		// Load from terraform.tfvars
		tfvarsPath := filepath.Join(workingDir, "terraform.tfvars")
		if content, err := os.ReadFile(tfvarsPath); err == nil {
			// Simple variable parsing
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if line != "" && strings.Contains(line, "=") {
					parts := strings.SplitN(line, "=", 2)
					if len(parts) == 2 {
						key := strings.TrimSpace(parts[0])
						value := strings.TrimSpace(strings.Trim(parts[1], `"'`))
						variables[key] = value
					}
				}
			}
		}

		// Load from terraform.tfvars.json
		tfvarsJSONPath := filepath.Join(workingDir, "terraform.tfvars.json")
		if content, err := os.ReadFile(tfvarsJSONPath); err == nil {
			var jsonVars map[string]interface{}
			if json.Unmarshal(content, &jsonVars) == nil {
				for k, v := range jsonVars {
					variables[k] = fmt.Sprintf("%v", v)
				}
			}
		}

		// Load from environment-specific files
		envFiles := []string{
			"dev.tfvars",
			"staging.tfvars",
			"prod.tfvars",
		}

		for _, envFile := range envFiles {
			envPath := filepath.Join(workingDir, envFile)
			if content, err := os.ReadFile(envPath); err == nil {
				// Simple variable parsing for environment files
				lines := strings.Split(string(content), "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line != "" && strings.Contains(line, "=") {
						parts := strings.SplitN(line, "=", 2)
						if len(parts) == 2 {
							key := strings.TrimSpace(parts[0])
							value := strings.TrimSpace(strings.Trim(parts[1], `"'`))
							variables[key] = value // Environment files override defaults
						}
					}
				}
			}
		}

		return variablesLoadedMsg{variables: variables}
	}
}

// loadTerraformOutputs loads Terraform output values
func (m *TerraformTUI) loadTerraformOutputs() tea.Cmd {
	return func() tea.Msg {
		if m.manager == nil {
			return errorMsg{fmt.Errorf("no terraform manager available")}
		}

		// Get outputs using terraform output -json
		operation, err := tfbicep.TerraformOutput(m.manager.WorkingDir)
		if err != nil {
			return errorMsg{fmt.Errorf("failed to load outputs: %v", err)}
		}

		if !operation.Success {
			return errorMsg{fmt.Errorf("terraform output failed: %s", operation.Error)}
		}

		// Parse JSON output
		var outputs map[string]interface{}
		if err := json.Unmarshal([]byte(operation.Output), &outputs); err != nil {
			return errorMsg{fmt.Errorf("failed to parse output JSON: %v", err)}
		}

		return outputsLoadedMsg{outputs: outputs, content: operation.Output}
	}
}

// renderProgressIndicator renders a fallback progress indicator
func (m *TerraformTUI) renderProgressIndicator(progress ParseProgressMsg) string {
	percentage := float64(progress.current) / float64(progress.total) * 100

	progressBar := strings.Repeat("█", int(percentage/4))
	progressBar += strings.Repeat("░", 25-int(percentage/4))

	return fmt.Sprintf("Progress: [%s] %.1f%% - %s",
		progressBar, percentage, progress.stage)
}
