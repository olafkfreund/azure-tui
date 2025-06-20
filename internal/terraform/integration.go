package terraform

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/olafkfreund/azure-tui/internal/config"
)

// Integration functions for adding Terraform support to the main TUI

// LaunchTerraformTUI launches the standalone Terraform TUI
func LaunchTerraformTUI() error {
	// Ensure Terraform directories exist
	if err := config.EnsureTerraformDirectories(); err != nil {
		return err
	}

	tui := NewTerraformTUI()
	p := tea.NewProgram(tui, tea.WithAltScreen())

	_, err := p.Run()
	return err
}

// TerraformMenuOption represents a Terraform menu option for integration
type TerraformMenuOption struct {
	Title       string
	Description string
	Action      func() tea.Cmd
}

// GetTerraformMenuOptions returns menu options for Terraform integration
func GetTerraformMenuOptions() []TerraformMenuOption {
	return []TerraformMenuOption{
		{
			Title:       "🏗️  Terraform Manager",
			Description: "Full Terraform workspace management",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					go LaunchTerraformTUI()
					return nil
				}
			},
		},
		{
			Title:       "📝  Create from Template",
			Description: "Create new infrastructure from templates",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					// TODO: Implement quick template creation
					return nil
				}
			},
		},
		{
			Title:       "⚡  Quick Deploy",
			Description: "Deploy common Azure resources quickly",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					// TODO: Implement quick deploy
					return nil
				}
			},
		},
		{
			Title:       "📊  State Viewer",
			Description: "View and manage Terraform state",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					// TODO: Implement state viewer
					return nil
				}
			},
		},
	}
}

// TerraformShortcuts returns keyboard shortcuts for Terraform operations
func TerraformShortcuts() map[string]string {
	cfg := config.GetUIConfig()
	return cfg.TerraformShortcuts
}

// HandleTerraformShortcut handles a Terraform keyboard shortcut
func HandleTerraformShortcut(shortcut string) tea.Cmd {
	switch shortcut {
	case "terraform_menu":
		return func() tea.Msg {
			go LaunchTerraformTUI()
			return nil
		}
	case "new_terraform_file":
		return func() tea.Msg {
			// TODO: Implement new file creation
			return nil
		}
	case "terraform_plan":
		return func() tea.Msg {
			// TODO: Implement quick plan
			return nil
		}
	case "terraform_apply":
		return func() tea.Msg {
			// TODO: Implement quick apply
			return nil
		}
	default:
		return nil
	}
}
