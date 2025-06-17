package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/olafkfreund/azure-tui/internal/tui"
)

type testModel struct {
	statusBar *tui.StatusBar
	width     int
	height    int
	ready     bool
}

func (m testModel) Init() tea.Cmd {
	return nil
}

func (m testModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		}
	}
	return m, nil
}

func (m testModel) View() string {
	if !m.ready {
		return "Loading..."
	}

	// Update status bar
	if m.statusBar != nil {
		m.statusBar.Segments = []tui.PowerlineSegment{}
		m.statusBar.AddSegment("🚀 Azure TUI Test", lipgloss.Color("39"), lipgloss.Color("15"))
		m.statusBar.AddSegment("✅ Basic TUI Working", lipgloss.Color("33"), lipgloss.Color("15"))
	}

	content := "Azure TUI - Basic Test\n\n"
	content += "✅ BubbleTea is working\n"
	content += "✅ TUI components loading\n"
	content += "✅ Lipgloss styling active\n\n"
	content += "Press 'q' to quit"

	boxStyle := lipgloss.NewStyle().
		Width(m.width-4).
		Height(m.height-4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(2, 4).
		Align(lipgloss.Center, lipgloss.Center)

	view := boxStyle.Render(content)

	if m.statusBar != nil {
		statusBarContent := m.statusBar.RenderStatusBar()
		view = statusBarContent + "\n" + view
	}

	return view
}

func main() {
	statusBar := tui.CreatePowerlineStatusBar(80)

	m := testModel{
		statusBar: statusBar,
		width:     80,
		height:    24,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error starting test TUI: %v\n", err)
		os.Exit(1)
	}
}
