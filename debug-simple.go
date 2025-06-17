package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type simpleModel struct {
	width  int
	height int
}

func (m simpleModel) Init() tea.Cmd {
	return nil
}

func (m simpleModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		if msg.String() == "q" {
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m simpleModel) View() string {
	// Test status bar
	statusBar := "🚀 Azure TUI | ☁️ Loading..."

	// Test left panel
	leftPanel := lipgloss.NewStyle().
		Width(40).
		Height(15).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(1, 1).
		Render("☁️ Azure Resources\n\n🔄 Loading resource groups...\n\nPress ? for help")

	// Test right panel
	rightPanel := lipgloss.NewStyle().
		Width(60).
		Height(15).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(1, 1).
		Render("Welcome to Azure TUI\n\nTREE VIEW INTERFACE\n\nNavigate with:\n• j/k or ↑↓ - Navigate tree\n• Space - Expand/collapse")

	// Join panels horizontally
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	// Join with status bar vertically
	fullView := lipgloss.JoinVertical(lipgloss.Left, statusBar, mainContent)

	// Render final box
	boxStyle := lipgloss.NewStyle().
		Width(m.width-2).
		Height(m.height-2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("63")).
		Align(lipgloss.Left, lipgloss.Top).
		Background(lipgloss.Color("235")).
		Foreground(lipgloss.Color("252")).
		Padding(1, 2)

	return boxStyle.Render(fullView)
}

func main() {
	m := simpleModel{width: 120, height: 30}
	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
