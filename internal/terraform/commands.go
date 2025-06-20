package terraform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/olafkfreund/azure-tui/internal/azure/tfbicep"
	"github.com/olafkfreund/azure-tui/internal/config"
)

// Additional rendering methods for TerraformTUI

func (m *TerraformTUI) renderTemplatesView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	if m.activeView == ViewTemplates {
		style = style.BorderForeground(lipgloss.Color("#FF5F87"))
	}

	return style.Render(m.templates.View())
}

func (m *TerraformTUI) renderWorkspacesView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	if m.activeView == ViewWorkspaces {
		style = style.BorderForeground(lipgloss.Color("#FF5F87"))
	}

	return style.Render(m.workspaces.View())
}

func (m *TerraformTUI) renderEditorView() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	if m.activeView == ViewEditor {
		style = style.BorderForeground(lipgloss.Color("#FF5F87"))
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
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	if m.activeView == ViewOperations {
		style = style.BorderForeground(lipgloss.Color("#FF5F87"))
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
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2)

	if m.activeView == ViewState {
		style = style.BorderForeground(lipgloss.Color("#FF5F87"))
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
		m.currentTemplate = templatePath

		// Load main.tf if exists
		mainTfPath := filepath.Join(templatePath, "main.tf")
		if content, err := os.ReadFile(mainTfPath); err == nil {
			return fileLoadedMsg{
				path:    mainTfPath,
				content: string(content),
			}
		}

		return nil
	}
}

func (m *TerraformTUI) selectWorkspace(workspacePath string) tea.Cmd {
	return func() tea.Msg {
		m.currentWorkspace = workspacePath
		m.manager = tfbicep.NewTerraformManager(workspacePath)

		// Load main.tf if exists
		mainTfPath := filepath.Join(workspacePath, "main.tf")
		if content, err := os.ReadFile(mainTfPath); err == nil {
			return fileLoadedMsg{
				path:    mainTfPath,
				content: string(content),
			}
		}

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
