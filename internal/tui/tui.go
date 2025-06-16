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

// Tab represents a TUI tab/window (e.g., resource, connection, monitor, etc.)
type Tab struct {
	Title    string
	Content  string            // Could be a rendered panel, connection, or monitoring view
	Type     string            // e.g. "resource", "aks", "vm", "monitor", "health", "shell"
	Meta     map[string]string // Extra info (e.g. resource ID, connection params)
	Closable bool
}

// TabManager manages multiple tabs/windows in the TUI
// Supports tab open/close, switching, and nested tabs
// (Stateful, to be embedded in the main model)
type TabManager struct {
	Tabs        []Tab
	ActiveIndex int
}

func NewTabManager() *TabManager {
	return &TabManager{Tabs: []Tab{}, ActiveIndex: 0}
}

func (tm *TabManager) AddTab(tab Tab) {
	tm.Tabs = append(tm.Tabs, tab)
	tm.ActiveIndex = len(tm.Tabs) - 1
}

func (tm *TabManager) CloseTab(idx int) {
	if idx < 0 || idx >= len(tm.Tabs) {
		return
	}
	tm.Tabs = append(tm.Tabs[:idx], tm.Tabs[idx+1:]...)
	if tm.ActiveIndex >= len(tm.Tabs) {
		tm.ActiveIndex = len(tm.Tabs) - 1
	}
	if tm.ActiveIndex < 0 {
		tm.ActiveIndex = 0
	}
}

func (tm *TabManager) SwitchTab(delta int) {
	if len(tm.Tabs) == 0 {
		tm.ActiveIndex = 0
		return
	}
	tm.ActiveIndex = (tm.ActiveIndex + delta + len(tm.Tabs)) % len(tm.Tabs)
}

func (tm *TabManager) ActiveTab() *Tab {
	if len(tm.Tabs) == 0 {
		return nil
	}
	return &tm.Tabs[tm.ActiveIndex]
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

// RenderTabs renders the tab bar and the active tab's content
func RenderTabs(tm *TabManager, status string) string {
	if tm == nil || len(tm.Tabs) == 0 {
		return "No tabs open."
	}
	var tabBar strings.Builder
	for i, tab := range tm.Tabs {
		if i == tm.ActiveIndex {
			tabBar.WriteString(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10")).Render("[" + tab.Title + "] "))
		} else {
			tabBar.WriteString(lipgloss.NewStyle().Faint(true).Render(tab.Title + " "))
		}
	}
	content := tm.Tabs[tm.ActiveIndex].Content
	statusLine := lipgloss.NewStyle().Faint(true).Render(status)
	return tabBar.String() + "\n" + content + "\n" + statusLine
}

// RenderShortcutsPopup renders a popup with all keyboard shortcuts
func RenderShortcutsPopup(shortcuts map[string]string) string {
	var b strings.Builder
	b.WriteString("Keyboard Shortcuts:\n\n")
	for k, v := range shortcuts {
		b.WriteString(fmt.Sprintf("%-8s : %s\n", k, v))
	}
	return lipgloss.NewStyle().Width(50).Height(20).Align(lipgloss.Center, lipgloss.Center).Border(lipgloss.RoundedBorder()).Render(b.String())
}
