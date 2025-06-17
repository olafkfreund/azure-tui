package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type debugModel struct {
	loading    bool
	loaded     bool
	message    string
	termWidth  int
	termHeight int
}

func (m debugModel) Init() tea.Cmd {
	return func() tea.Msg {
		return loadingComplete{"Demo data loaded successfully!"}
	}
}

type loadingComplete struct {
	msg string
}

func (m debugModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		return m, nil
	case loadingComplete:
		m.loading = false
		m.loaded = true
		m.message = msg.msg
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m debugModel) View() string {
	if m.loading {
		content := "🔄 Loading Azure TUI...\n\nInitializing demo data..."
		return renderBox(content, m.termWidth, m.termHeight)
	}

	if m.loaded {
		content := fmt.Sprintf("✅ %s\n\nPress 'q' to quit", m.message)
		return renderBox(content, m.termWidth, m.termHeight)
	}

	return "Starting..."
}

func renderBox(content string, width, height int) string {
	if width < 40 {
		width = 40
	}
	if height < 10 {
		height = 10
	}

	style := lipgloss.NewStyle().
		Width(width-4).
		Height(height-4).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("63")).
		Background(lipgloss.Color("236")).
		Foreground(lipgloss.Color("252")).
		Padding(2, 4).
		Align(lipgloss.Center, lipgloss.Center)

	return style.Render(content)
}

func main() {
	m := debugModel{
		loading:    true,
		termWidth:  80,
		termHeight: 24,
	}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if err := p.Start(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
