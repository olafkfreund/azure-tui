package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// This package will contain the Bubble Tea TUI logic and models.
// TODO: Implement main menu, environment dashboard, and navigation.

// PopupMsg is a message for showing popups (alarms/errors)
type PopupMsg struct {
	Title   string
	Content string
	Level   string // "error", "alarm", "info"
}

// MatrixGraphMsg is a message for showing a matrix/graph in the TUI
type MatrixGraphMsg struct {
	Title  string
	Rows   [][]string // 2D matrix of values
	Labels []string   // Row/column labels
}

// RenderPopup renders a popup window for alarms/errors
func RenderPopup(msg PopupMsg) string {
	border := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).Padding(1, 2)
	title := msg.Title
	if msg.Level == "error" {
		title = "❌ " + title
	} else if msg.Level == "alarm" {
		title = "⚠️  " + title
	}
	return border.Render(fmt.Sprintf("%s\n\n%s", title, msg.Content))
}

// RenderMatrixGraph renders a simple ASCII matrix/graph
func RenderMatrixGraph(msg MatrixGraphMsg) string {
	var b strings.Builder
	b.WriteString(msg.Title + "\n\n")
	if len(msg.Labels) > 0 {
		b.WriteString("    ")
		for _, label := range msg.Labels {
			b.WriteString(fmt.Sprintf("%8s", label))
		}
		b.WriteString("\n")
	}
	for i, row := range msg.Rows {
		if len(msg.Labels) > 0 && i < len(msg.Labels) {
			b.WriteString(fmt.Sprintf("%4s", msg.Labels[i]))
		} else {
			b.WriteString("    ")
		}
		for _, val := range row {
			b.WriteString(fmt.Sprintf("%8s", val))
		}
		b.WriteString("\n")
	}
	return b.String()
}
