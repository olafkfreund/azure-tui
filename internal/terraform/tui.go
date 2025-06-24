package terraform

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/olafkfreund/azure-tui/internal/azure/tfbicep"
)

// Enhanced State Management Types
type StateResource struct {
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Provider     string                 `json:"provider"`
	Address      string                 `json:"address"`
	Attributes   map[string]interface{} `json:"attributes"`
	Dependencies []string               `json:"dependencies"`
	Tainted      bool                   `json:"tainted"`
	Status       string                 `json:"status"` // "ok", "tainted", "error"
}

type StateViewer struct {
	resources     []StateResource
	selectedIndex int
	viewMode      string // "list", "tree", "graph"
	filterText    string
	showDetails   bool
}

// Enhanced Plan Management Types
type PlanChange struct {
	Action    string                 `json:"action"`    // "create", "update", "delete", "replace"
	Resource  string                 `json:"resource"`  // resource address
	Type      string                 `json:"type"`      // resource type
	Name      string                 `json:"name"`      // resource name
	Before    map[string]interface{} `json:"before"`    // current values
	After     map[string]interface{} `json:"after"`     // planned values
	Reason    string                 `json:"reason"`    // reason for change
	Sensitive bool                   `json:"sensitive"` // contains sensitive data
	Impact    string                 `json:"impact"`    // "low", "medium", "high"
}

type PlanViewer struct {
	changes       []PlanChange
	selectedIndex int
	showDetails   bool
	filterAction  string // filter by action type
	groupByType   bool   // group changes by resource type
}

// Enhanced Workspace Management Types
type WorkspaceManager struct {
	workspaces    []WorkspaceInfo
	selectedIndex int
	currentEnv    string
	envVars       map[string]map[string]string // env -> var -> value
}

type WorkspaceInfo struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Environment string            `json:"environment"` // dev, staging, prod
	Backend     string            `json:"backend"`
	Variables   map[string]string `json:"variables"`
	LastApply   string            `json:"last_apply"`
	Status      string            `json:"status"` // "clean", "dirty", "error"
}

// Enhanced Variable Management Types
type VariableEditor struct {
	variables     map[string]string
	selectedIndex int
	editMode      bool
	editingVar    string
	editingValue  string
	originalValue string
	showConfirm   bool
}

// Progress indicator for performance tracking
type ProgressIndicator struct {
	isActive   bool
	current    int
	total      int
	stage      string
	lastUpdate time.Time
}

// Performance optimization settings
type PerformanceConfig struct {
	enableProgressIndicators bool
	batchSize                int
	cacheSize                int
	streamingThreshold       int // bytes
}

// ParseProgressMsg represents progress while parsing large plan files
type ParseProgressMsg struct {
	current int
	total   int
	stage   string
}

// Message types for variable management and progress tracking
type variableEditCompletedMsg struct {
	name  string
	value string
}

// Additional message types for Terraform operations
type errorMsg struct{ error }
type templatesLoadedMsg struct{ items []list.Item }
type workspacesLoadedMsg struct{ items []list.Item }
type fileLoadedMsg struct{ path, content string }
type fileSavedMsg struct{ path string }
type fileEditedMsg struct{ path string }
type operationCompletedMsg struct{ operation tfbicep.TerraformOperation }

// Enhanced feature message types
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

	// Enhanced State Management
	stateViewer      StateViewer
	stateResources   []StateResource
	selectedResource int
	showDependencies bool

	// Enhanced Plan Management
	planViewer      PlanViewer
	planChanges     []PlanChange
	selectedChange  int
	showPlanDetails bool
	approvalMode    bool

	// Workspace Management
	workspaceManager WorkspaceManager
	currentEnv       string // dev, staging, prod
	envVariables     map[string]string

	// Interactive Variable Editing
	variableEditor VariableEditor
	showVarEditor  bool

	// Performance and Progress Tracking
	progressIndicator ProgressIndicator
	perfConfig        PerformanceConfig
}

// Views
const (
	ViewTemplates   = "templates"
	ViewWorkspaces  = "workspaces"
	ViewEditor      = "editor"
	ViewOperations  = "operations"
	ViewState       = "state"
	ViewStateViewer = "state_viewer"
	ViewPlanViewer  = "plan_viewer"
	ViewEnvManager  = "env_manager"
	ViewVarEditor   = "var_editor"
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

	// Enhanced features
	StateViewer    key.Binding
	PlanViewer     key.Binding
	EnvManager     key.Binding
	ShowDeps       key.Binding
	FilterToggle   key.Binding
	ApprovalMode   key.Binding
	ResourceTarget key.Binding

	// Additional enhanced features
	VariableManager key.Binding
	OutputViewer    key.Binding
	EditVariable    key.Binding
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

	// Enhanced features
	StateViewer: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "state viewer"),
	),
	PlanViewer: key.NewBinding(
		key.WithKeys("p"),
		key.WithHelp("p", "plan viewer"),
	),
	EnvManager: key.NewBinding(
		key.WithKeys("w"),
		key.WithHelp("w", "workspace manager"),
	),
	ShowDeps: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "show dependencies"),
	),
	FilterToggle: key.NewBinding(
		key.WithKeys("f"),
		key.WithHelp("f", "toggle filter"),
	),
	ApprovalMode: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "approval mode"),
	),
	ResourceTarget: key.NewBinding(
		key.WithKeys("t"),
		key.WithHelp("t", "target resource"),
	),
	VariableManager: key.NewBinding(
		key.WithKeys("v"),
		key.WithHelp("v", "variable manager"),
	),
	OutputViewer: key.NewBinding(
		key.WithKeys("o"),
		key.WithHelp("o", "show outputs"),
	),
	EditVariable: key.NewBinding(
		key.WithKeys("e"),
		key.WithHelp("e", "edit variable"),
	),
}

// NewTerraformTUI creates a new Terraform TUI
func NewTerraformTUI() *TerraformTUI {
	tui := &TerraformTUI{
		activeView: ViewTemplates,
		status:     "Ready",
		operations: make([]tfbicep.TerraformOperation, 0),

		// Initialize enhanced components
		stateViewer: StateViewer{
			viewMode: "list",
		},
		planViewer: PlanViewer{
			showDetails: false,
		},
		workspaceManager: WorkspaceManager{
			currentEnv: "dev",
			envVars:    make(map[string]map[string]string),
		},
		envVariables: make(map[string]string),
		variableEditor: VariableEditor{
			variables: make(map[string]string),
		},
		progressIndicator: ProgressIndicator{
			isActive: false,
			current:  0,
			total:    0,
			stage:    "",
		},
		perfConfig: PerformanceConfig{
			enableProgressIndicators: true,
			batchSize:                100,
			cacheSize:                1024,
			streamingThreshold:       4096,
		},
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

		// Enhanced feature key bindings
		case key.Matches(msg, keys.StateViewer):
			m.activeView = ViewStateViewer
			if m.perfConfig.enableProgressIndicators {
				return m, m.loadStateResourcesWithProgress()
			} else {
				return m, m.loadStateResources()
			}

		case key.Matches(msg, keys.PlanViewer):
			m.activeView = ViewPlanViewer
			if m.perfConfig.enableProgressIndicators {
				return m, m.loadPlanChangesWithProgress()
			} else {
				return m, m.loadPlanChanges()
			}

		case key.Matches(msg, keys.EnvManager):
			m.activeView = ViewEnvManager
			return m, m.loadWorkspaceInfo()

		case key.Matches(msg, keys.ShowDeps):
			if m.activeView == ViewStateViewer {
				m.showDependencies = !m.showDependencies
			}

		case key.Matches(msg, keys.FilterToggle):
			if m.activeView == ViewPlanViewer {
				m.togglePlanFilter()
			}

		case key.Matches(msg, keys.ApprovalMode):
			if m.activeView == ViewPlanViewer {
				m.approvalMode = !m.approvalMode
			}

		case key.Matches(msg, keys.ResourceTarget):
			if m.activeView == ViewPlanViewer && len(m.planChanges) > 0 {
				return m, m.targetResource(m.planChanges[m.selectedChange].Resource)
			}

		case key.Matches(msg, keys.VariableManager):
			m.activeView = ViewVarEditor
			return m, m.loadTerraformVariables()

		case key.Matches(msg, keys.OutputViewer):
			return m, m.loadTerraformOutputs()

		case key.Matches(msg, keys.EditVariable):
			if m.activeView == ViewVarEditor && !m.variableEditor.editMode {
				return m, m.startVariableEdit()
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

		case ViewVarEditor:
			// Handle variable editor specific key bindings
			if m.variableEditor.editMode {
				// In edit mode
				switch {
				case key.Matches(msg, keys.Enter):
					cmd = m.saveVariableEdit()
				case key.Matches(msg, keys.Escape):
					m.cancelVariableEdit()
				default:
					// Handle text input
					if len(msg.String()) == 1 {
						m.variableEditor.editingValue += msg.String()
					} else if msg.Type == tea.KeyBackspace && len(m.variableEditor.editingValue) > 0 {
						m.variableEditor.editingValue = m.variableEditor.editingValue[:len(m.variableEditor.editingValue)-1]
					}
				}
			} else {
				// In navigation mode
				switch {
				case key.Matches(msg, keys.Up):
					if m.variableEditor.selectedIndex > 0 {
						m.variableEditor.selectedIndex--
					}
				case key.Matches(msg, keys.Down):
					if m.variableEditor.selectedIndex < len(m.variableEditor.variables)-1 {
						m.variableEditor.selectedIndex++
					}
				case key.Matches(msg, keys.EditVariable):
					cmd = m.startVariableEdit()
				}
			}
		}

		cmds = append(cmds, cmd)

	// Handle enhanced feature messages
	case stateResourcesLoadedMsg:
		m.stateViewer.resources = msg.resources
		m.status = "State resources loaded"
		m.progressIndicator.isActive = false

	case planChangesLoadedMsg:
		m.planViewer.changes = msg.changes
		m.status = "Plan changes loaded"
		m.progressIndicator.isActive = false

	case workspaceInfoLoadedMsg:
		m.workspaceManager.workspaces = msg.workspaces
		m.currentWorkspace = msg.current
		m.status = "Workspace info loaded"

	// Progress handling for performance optimization
	case ParseProgressMsg:
		m.progressIndicator.isActive = true
		m.progressIndicator.current = msg.current
		m.progressIndicator.total = msg.total
		m.progressIndicator.stage = msg.stage
		m.progressIndicator.lastUpdate = time.Now()
		m.status = fmt.Sprintf("%s (%d%%)", msg.stage, int(float64(msg.current)/float64(msg.total)*100))

	// Workspace management messages
	case workspaceSwitchedMsg:
		if msg.success {
			m.currentWorkspace = msg.workspace
			m.status = msg.message
			// Reload workspace info after switching
			cmds = append(cmds, m.loadWorkspaceInfo())
		} else {
			m.errorMsg = msg.message
		}

	case workspaceCreatedMsg:
		if msg.success {
			m.status = msg.message
			// Reload workspace info after creation
			cmds = append(cmds, m.loadWorkspaceInfo())
		} else {
			m.errorMsg = msg.message
		}

	case workspaceDeletedMsg:
		if msg.success {
			m.status = msg.message
			// Reload workspace info after deletion
			cmds = append(cmds, m.loadWorkspaceInfo())
		} else {
			m.errorMsg = msg.message
			// Variable management messages
		}

	case variablesLoadedMsg:
		m.envVariables = msg.variables
		m.variableEditor.variables = msg.variables
		m.variableEditor.selectedIndex = 0
		m.status = "Variables loaded"

	case variableUpdatedMsg:
		if msg.success {
			m.status = fmt.Sprintf("Variable '%s' updated", msg.name)
			// Reload variables
			cmds = append(cmds, m.loadTerraformVariables())
		} else {
			m.errorMsg = msg.message
		}

	case variableEditCompletedMsg:
		m.status = fmt.Sprintf("Variable '%s' successfully updated to '%s'", msg.name, msg.value)

	// Output values messages
	case outputsLoadedMsg:
		m.status = "Output values loaded"
		m.showPopup = true
		m.popupTitle = "Terraform Outputs"
		m.popupContent = msg.content

	default:
		// For now, just return the model as-is for unhandled messages
		// This could be enhanced to handle additional message types
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
	case ViewStateViewer:
		content = m.renderStateViewerView()
	case ViewPlanViewer:
		content = m.renderPlanViewerView()
	case ViewEnvManager:
		content = m.renderEnvManagerView()
	case ViewVarEditor:
		content = m.renderVarEditorView()
	default:
		content = m.renderTemplatesView()
	}

	// Add progress indicator overlay if active
	if m.progressIndicator.isActive && m.perfConfig.enableProgressIndicators {
		progressOverlay := m.renderProgressIndicator(ParseProgressMsg{
			current: m.progressIndicator.current,
			total:   m.progressIndicator.total,
			stage:   m.progressIndicator.stage,
		})

		content = lipgloss.JoinVertical(
			lipgloss.Left,
			content,
			"",
			progressOverlay,
		)
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
	views := []string{ViewTemplates, ViewWorkspaces, ViewEditor, ViewOperations, ViewState, ViewStateViewer, ViewPlanViewer, ViewEnvManager, ViewVarEditor}
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
	views := []string{ViewTemplates, ViewWorkspaces, ViewEditor, ViewOperations, ViewState, ViewStateViewer, ViewPlanViewer, ViewEnvManager, ViewVarEditor}
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

	// Clean, frameless popup style for consistency with main help popup
	popup := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
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
