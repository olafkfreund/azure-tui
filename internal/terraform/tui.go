package terraform

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/olafkfreund/azure-tui/internal/azure/tfbicep"
)

// TerraformTUI represents the Terraform TUI interface
type TerraformTUI struct {
	width            int
	height           int
	activeView       string
	templates        list.Model
	workspaces       list.Model
	editor           textarea.Model
	manager          *tfbicep.TerraformManager
	currentFile      string
	currentTemplate  string
	currentWorkspace string
	status           string
	errorMsg         string
	operations       []tfbicep.TerraformOperation
	showPopup        bool
	popupContent     string
	popupTitle       string
}

// Views
const (
	ViewTemplates  = "templates"
	ViewWorkspaces = "workspaces"
	ViewEditor     = "editor"
	ViewOperations = "operations"
	ViewState      = "state"
)

// Key bindings
type keyMap struct {
	Up         key.Binding
	Down       key.Binding
	Left       key.Binding
	Right      key.Binding
	Enter      key.Binding
	Escape     key.Binding
	Tab        key.Binding
	ShiftTab   key.Binding
	Quit       key.Binding
	Help       key.Binding
	NewFile    key.Binding
	SaveFile   key.Binding
	OpenEditor key.Binding
	Plan       key.Binding
	Apply      key.Binding
	Destroy    key.Binding
	Init       key.Binding
	Validate   key.Binding
	Format     key.Binding
	State      key.Binding
	Refresh    key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Escape, k.Tab, k.ShiftTab},
		{k.NewFile, k.SaveFile, k.OpenEditor},
		{k.Plan, k.Apply, k.Destroy, k.Init},
		{k.Validate, k.Format, k.State, k.Refresh},
		{k.Help, k.Quit},
	}
}

var keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "move left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "move right"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next view"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev view"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	NewFile: key.NewBinding(
		key.WithKeys("ctrl+n"),
		key.WithHelp("ctrl+n", "new file"),
	),
	SaveFile: key.NewBinding(
		key.WithKeys("ctrl+s"),
		key.WithHelp("ctrl+s", "save file"),
	),
	OpenEditor: key.NewBinding(
		key.WithKeys("ctrl+e"),
		key.WithHelp("ctrl+e", "external editor"),
	),
	Plan: key.NewBinding(
		key.WithKeys("ctrl+p"),
		key.WithHelp("ctrl+p", "terraform plan"),
	),
	Apply: key.NewBinding(
		key.WithKeys("ctrl+a"),
		key.WithHelp("ctrl+a", "terraform apply"),
	),
	Destroy: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "terraform destroy"),
	),
	Init: key.NewBinding(
		key.WithKeys("ctrl+i"),
		key.WithHelp("ctrl+i", "terraform init"),
	),
	Validate: key.NewBinding(
		key.WithKeys("ctrl+v"),
		key.WithHelp("ctrl+v", "terraform validate"),
	),
	Format: key.NewBinding(
		key.WithKeys("ctrl+f"),
		key.WithHelp("ctrl+f", "terraform format"),
	),
	State: key.NewBinding(
		key.WithKeys("ctrl+t"),
		key.WithHelp("ctrl+t", "terraform state"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("ctrl+r"),
		key.WithHelp("ctrl+r", "refresh"),
	),
}

// NewTerraformTUI creates a new Terraform TUI
func NewTerraformTUI() *TerraformTUI {
	tui := &TerraformTUI{
		activeView: ViewTemplates,
		status:     "Ready",
		operations: make([]tfbicep.TerraformOperation, 0),
	}

	// Initialize templates list
	tui.templates = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	tui.templates.Title = "Terraform Templates"
	tui.templates.SetShowStatusBar(false)
	tui.templates.SetFilteringEnabled(true)

	// Initialize workspaces list
	tui.workspaces = list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	tui.workspaces.Title = "Terraform Workspaces"
	tui.workspaces.SetShowStatusBar(false)
	tui.workspaces.SetFilteringEnabled(true)

	// Initialize editor
	tui.editor = textarea.New()
	tui.editor.Placeholder = "Start typing your Terraform configuration..."
	tui.editor.Focus()

	return tui
}

// Init implements the bubbletea.Model interface
func (m *TerraformTUI) Init() tea.Cmd {
	return tea.Batch(
		m.loadTemplates(),
		m.loadWorkspaces(),
	)
}

// Update implements the bubbletea.Model interface
func (m *TerraformTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateSizes()

	case tea.KeyMsg:
		if m.showPopup {
			if key.Matches(msg, keys.Escape) || key.Matches(msg, keys.Enter) {
				m.showPopup = false
				return m, nil
			}
			return m, nil
		}

		switch {
		case key.Matches(msg, keys.Quit):
			return m, tea.Quit

		case key.Matches(msg, keys.Tab):
			m.nextView()

		case key.Matches(msg, keys.ShiftTab):
			m.prevView()

		case key.Matches(msg, keys.Init):
			return m, m.terraformInit()

		case key.Matches(msg, keys.Plan):
			return m, m.terraformPlan()

		case key.Matches(msg, keys.Apply):
			return m, m.terraformApply()

		case key.Matches(msg, keys.Destroy):
			return m, m.terraformDestroy()

		case key.Matches(msg, keys.Validate):
			return m, m.terraformValidate()

		case key.Matches(msg, keys.Format):
			return m, m.terraformFormat()

		case key.Matches(msg, keys.State):
			m.activeView = ViewState
			return m, m.loadState()

		case key.Matches(msg, keys.OpenEditor):
			return m, m.openExternalEditor()

		case key.Matches(msg, keys.SaveFile):
			return m, m.saveCurrentFile()

		case key.Matches(msg, keys.NewFile):
			return m, m.newTerraformFile()

		case key.Matches(msg, keys.Escape):
			if m.activeView == ViewEditor {
				m.activeView = ViewTemplates
			}
		}

		// Handle view-specific updates
		switch m.activeView {
		case ViewTemplates:
			m.templates, cmd = m.templates.Update(msg)
			if key.Matches(msg, keys.Enter) {
				if item, ok := m.templates.SelectedItem().(templateItem); ok {
					return m, m.selectTemplate(item.path)
				}
			}

		case ViewWorkspaces:
			m.workspaces, cmd = m.workspaces.Update(msg)
			if key.Matches(msg, keys.Enter) {
				if item, ok := m.workspaces.SelectedItem().(workspaceItem); ok {
					return m, m.selectWorkspace(item.path)
				}
			}

		case ViewEditor:
			m.editor, cmd = m.editor.Update(msg)
		}

		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View implements the bubbletea.Model interface
func (m *TerraformTUI) View() string {
	if m.showPopup {
		return m.renderWithPopup()
	}

	var content string
	switch m.activeView {
	case ViewTemplates:
		content = m.renderTemplatesView()
	case ViewWorkspaces:
		content = m.renderWorkspacesView()
	case ViewEditor:
		content = m.renderEditorView()
	case ViewOperations:
		content = m.renderOperationsView()
	case ViewState:
		content = m.renderStateView()
	default:
		content = m.renderTemplatesView()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		m.renderHeader(),
		content,
		m.renderFooter(),
	)
}

// Helper methods
func (m *TerraformTUI) updateSizes() {
	listHeight := m.height - 6 // Header + footer + padding
	listWidth := m.width - 4   // Side padding

	m.templates.SetSize(listWidth, listHeight)
	m.workspaces.SetSize(listWidth, listHeight)
	m.editor.SetWidth(listWidth)
	m.editor.SetHeight(listHeight)
}

func (m *TerraformTUI) nextView() {
	views := []string{ViewTemplates, ViewWorkspaces, ViewEditor, ViewOperations, ViewState}
	current := 0
	for i, view := range views {
		if view == m.activeView {
			current = i
			break
		}
	}
	m.activeView = views[(current+1)%len(views)]
}

func (m *TerraformTUI) prevView() {
	views := []string{ViewTemplates, ViewWorkspaces, ViewEditor, ViewOperations, ViewState}
	current := 0
	for i, view := range views {
		if view == m.activeView {
			current = i
			break
		}
	}
	m.activeView = views[(current-1+len(views))%len(views)]
}

func (m *TerraformTUI) renderHeader() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(0, 1).
		Render("Azure TUI - Terraform Manager")

	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(fmt.Sprintf("Status: %s | View: %s", m.status, m.activeView))

	return lipgloss.JoinVertical(lipgloss.Left, title, status)
}

func (m *TerraformTUI) renderFooter() string {
	if m.errorMsg != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)
		return errorStyle.Render(fmt.Sprintf("Error: %s", m.errorMsg))
	}

	helpText := "Tab: Switch views • Ctrl+P: Plan • Ctrl+A: Apply • Ctrl+D: Destroy • Q: Quit • ?: Help"
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render(helpText)
}

func (m *TerraformTUI) renderWithPopup() string {
	// base := m.View()

	popup := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFD")).
		Padding(1, 2).
		Width(60).
		Height(20).
		Render(fmt.Sprintf("%s\n\n%s", m.popupTitle, m.popupContent))

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, popup, lipgloss.WithWhitespaceChars("•"))
}

// Template item for the list
type templateItem struct {
	title       string
	description string
	path        string
}

func (i templateItem) Title() string       { return i.title }
func (i templateItem) Description() string { return i.description }
func (i templateItem) FilterValue() string { return i.title }

// Workspace item for the list
type workspaceItem struct {
	title       string
	description string
	path        string
}

func (i workspaceItem) Title() string       { return i.title }
func (i workspaceItem) Description() string { return i.description }
func (i workspaceItem) FilterValue() string { return i.title }

// Additional view rendering methods will be implemented in the next part...
