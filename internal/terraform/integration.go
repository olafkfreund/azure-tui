package terraform

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/lipgloss"
	"github.com/olafkfreund/azure-tui/internal/config"
	"github.com/olafkfreund/azure-tui/internal/azure/tfbicep"
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
					return quickTemplateCreationMsg{}
				}
			},
		},
		{
			Title:       "⚡  Quick Deploy",
			Description: "Deploy common Azure resources quickly",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return quickDeployMsg{}
				}
			},
		},
		{
			Title:       "📊  State Viewer",
			Description: "View and manage Terraform state",
			Action: func() tea.Cmd {
				return func() tea.Msg {
					return stateViewerMsg{}
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
			return newFileCreationMsg{}
		}
	case "terraform_plan":
		return func() tea.Msg {
			return quickPlanMsg{}
		}
	case "terraform_apply":
		return func() tea.Msg {
			return quickApplyMsg{}
		}
	default:
		return nil
	}
}

// Message types for quick operations
type quickTemplateCreationMsg struct{}
type quickDeployMsg struct{}
type stateViewerMsg struct{}
type newFileCreationMsg struct{}
type quickPlanMsg struct{}
type quickApplyMsg struct{}

// QuickTemplateCreator handles quick template creation
type QuickTemplateCreator struct {
	templates     list.Model
	nameInput     textinput.Model
	selectedTemplate string
	step          int // 0: select template, 1: enter name, 2: create
	status        string
	error         string
}

// NewQuickTemplateCreator creates a new quick template creator
func NewQuickTemplateCreator() *QuickTemplateCreator {
	// Create templates list
	templates := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 10)
	templates.Title = "Select Template"

	// Create name input
	nameInput := textinput.New()
	nameInput.Placeholder = "Enter workspace name..."
	nameInput.CharLimit = 50
	nameInput.Width = 30

	return &QuickTemplateCreator{
		templates: templates,
		nameInput: nameInput,
		step:      0,
	}
}

// Init initializes the quick template creator
func (qtc *QuickTemplateCreator) Init() tea.Cmd {
	return tea.Batch(
		qtc.loadTemplates(),
		textinput.Blink,
	)
}

// Update handles updates for the quick template creator
func (qtc *QuickTemplateCreator) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return qtc, tea.Quit
		case "enter":
			switch qtc.step {
			case 0: // Template selection
				if item, ok := qtc.templates.SelectedItem().(templateItem); ok {
					qtc.selectedTemplate = item.path
					qtc.step = 1
					qtc.nameInput.Focus()
				}
			case 1: // Name input
				if qtc.nameInput.Value() != "" {
					qtc.step = 2
					return qtc, qtc.createWorkspace()
				}
			}
		case "esc":
			if qtc.step > 0 {
				qtc.step--
				if qtc.step == 0 {
					qtc.nameInput.Blur()
				}
			}
		}
	case templatesLoadedMsg:
		qtc.templates.SetItems(msg.items)
	case workspaceCreatedMsg:
		qtc.status = msg.message
		qtc.step = 3 // Show completion
	case errorMsg:
		qtc.error = msg.Error()
	}

	// Update components based on current step
	switch qtc.step {
	case 0:
		qtc.templates, cmd = qtc.templates.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		qtc.nameInput, cmd = qtc.nameInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return qtc, tea.Batch(cmds...)
}

// View renders the quick template creator
func (qtc *QuickTemplateCreator) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#874BFE")).
		Padding(1, 2).
		Margin(1, 0)

	switch qtc.step {
	case 0:
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("🚀 Quick Template Creation"),
			"",
			qtc.templates.View(),
			"",
			"Enter: Select | q: Quit",
		)
		return style.Render(content)

	case 1:
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("🚀 Quick Template Creation"),
			"",
			fmt.Sprintf("Selected Template: %s", filepath.Base(qtc.selectedTemplate)),
			"",
			"Enter workspace name:",
			qtc.nameInput.View(),
			"",
			"Enter: Create | Esc: Back | q: Quit",
		)
		return style.Render(content)

	case 2:
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("🚀 Quick Template Creation"),
			"",
			"Creating workspace...",
			"",
			"Please wait...",
		)
		return style.Render(content)

	case 3:
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("🚀 Quick Template Creation"),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("✓ " + qtc.status),
			"",
			"q: Quit",
		)
		return style.Render(content)
	}

	if qtc.error != "" {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("🚀 Quick Template Creation"),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("✗ Error: " + qtc.error),
			"",
			"q: Quit",
		)
		return style.Render(content)
	}

	return style.Render("Loading...")
}

// loadTemplates loads available templates
func (qtc *QuickTemplateCreator) loadTemplates() tea.Cmd {
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

// createWorkspace creates a new workspace from the selected template
func (qtc *QuickTemplateCreator) createWorkspace() tea.Cmd {
	return func() tea.Msg {
		cfg := config.GetTerraformConfig()
		workspaceName := qtc.nameInput.Value()
		workspacePath := filepath.Join(cfg.WorkspacePath, workspaceName)

		// Check if workspace already exists
		if _, err := os.Stat(workspacePath); err == nil {
			return errorMsg{fmt.Errorf("workspace '%s' already exists", workspaceName)}
		}

		// Copy template to workspace
		if err := tfbicep.CopyTemplate(qtc.selectedTemplate, workspacePath); err != nil {
			return errorMsg{fmt.Errorf("failed to copy template: %v", err)}
		}

		// Initialize the workspace
		if err := tfbicep.InitializeWorkspace(workspacePath, nil); err != nil {
			return errorMsg{fmt.Errorf("failed to initialize workspace: %v", err)}
		}

		return workspaceCreatedMsg{
			workspace: workspaceName,
			success:   true,
			message:   fmt.Sprintf("Workspace '%s' created successfully", workspaceName),
		}
	}
}

// Note: templatesLoadedMsg, workspaceCreatedMsg, errorMsg, templateItem are already defined in tui.go

// QuickTemplateCreationProgram creates a standalone program for quick template creation
func QuickTemplateCreationProgram() *tea.Program {
	creator := NewQuickTemplateCreator()
	return tea.NewProgram(creator, tea.WithAltScreen())
}

// QuickDeployManager handles quick deployment of common Azure resources
type QuickDeployManager struct {
	resourceTypes list.Model
	parameters    map[string]string
	paramInputs   map[string]textinput.Model
	currentParam  string
	paramKeys     []string
	paramIndex    int
	step          int // 0: select resource, 1: enter parameters, 2: deploy
	status        string
	error         string
	deployment    *TerraformDeployment
}

// TerraformDeployment represents a deployment operation
type TerraformDeployment struct {
	ResourceType string
	Name         string
	Parameters   map[string]string
	Status       string
	WorkingDir   string
	Manager      *tfbicep.TerraformManager
}

// NewQuickDeployManager creates a new quick deploy manager
func NewQuickDeployManager() *QuickDeployManager {
	// Create resource types list
	resourceTypes := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 10)
	resourceTypes.Title = "Select Resource Type"

	return &QuickDeployManager{
		resourceTypes: resourceTypes,
		parameters:    make(map[string]string),
		paramInputs:   make(map[string]textinput.Model),
		step:          0,
	}
}

// Init initializes the quick deploy manager
func (qdm *QuickDeployManager) Init() tea.Cmd {
	return tea.Batch(
		qdm.loadResourceTypes(),
		textinput.Blink,
	)
}

// Update handles updates for the quick deploy manager
func (qdm *QuickDeployManager) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return qdm, tea.Quit
		case "enter":
			switch qdm.step {
			case 0: // Resource selection
				if item, ok := qdm.resourceTypes.SelectedItem().(resourceTypeItem); ok {
					qdm.setupParameters(item.resourceType)
					qdm.step = 1
					qdm.focusCurrentParameter()
				}
			case 1: // Parameter input
				if qdm.paramIndex < len(qdm.paramKeys) {
					key := qdm.paramKeys[qdm.paramIndex]
					qdm.parameters[key] = qdm.paramInputs[key].Value()
					qdm.paramIndex++
					if qdm.paramIndex < len(qdm.paramKeys) {
						qdm.focusCurrentParameter()
					} else {
						qdm.step = 2
						return qdm, qdm.deployResource()
					}
				}
			}
		case "esc":
			if qdm.step > 0 {
				if qdm.step == 1 && qdm.paramIndex > 0 {
					qdm.paramIndex--
					qdm.focusCurrentParameter()
				} else {
					qdm.step--
					if qdm.step == 0 {
						qdm.blurAllInputs()
					}
				}
			}
		}
	case resourceTypesLoadedMsg:
		qdm.resourceTypes.SetItems(msg.items)
	case deploymentCompletedMsg:
		qdm.status = msg.message
		qdm.step = 3 // Show completion
	case errorMsg:
		qdm.error = msg.Error()
	}

	// Update components based on current step
	switch qdm.step {
	case 0:
		qdm.resourceTypes, cmd = qdm.resourceTypes.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		if qdm.paramIndex < len(qdm.paramKeys) {
			key := qdm.paramKeys[qdm.paramIndex]
			if input, exists := qdm.paramInputs[key]; exists {
				input, cmd = input.Update(msg)
				qdm.paramInputs[key] = input
				cmds = append(cmds, cmd)
			}
		}
	}

	return qdm, tea.Batch(cmds...)
}

// View renders the quick deploy manager
func (qdm *QuickDeployManager) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF6B6B")).
		Padding(1, 2).
		Margin(1, 0)

	switch qdm.step {
	case 0:
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("⚡ Quick Deploy"),
			"",
			qdm.resourceTypes.View(),
			"",
			"Enter: Select | q: Quit",
		)
		return style.Render(content)

	case 1:
		var content []string
		content = append(content, lipgloss.NewStyle().Bold(true).Render("⚡ Quick Deploy"))
		content = append(content, "")
		content = append(content, fmt.Sprintf("Resource Type: %s", qdm.resourceTypes.SelectedItem().(resourceTypeItem).title))
		content = append(content, "")
		content = append(content, "Enter parameters:")
		content = append(content, "")

		for i, key := range qdm.paramKeys {
			value := qdm.parameters[key]
			if i == qdm.paramIndex {
				content = append(content, fmt.Sprintf("▶ %s: %s", key, qdm.paramInputs[key].View()))
			} else if value != "" {
				content = append(content, fmt.Sprintf("  %s: %s", key, value))
			} else {
				content = append(content, fmt.Sprintf("  %s: <pending>", key))
			}
		}

		content = append(content, "")
		content = append(content, "Enter: Next | Esc: Back | q: Quit")

		return style.Render(lipgloss.JoinVertical(lipgloss.Left, content...))

	case 2:
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("⚡ Quick Deploy"),
			"",
			"Deploying resource...",
			"",
			"Please wait...",
		)
		return style.Render(content)

	case 3:
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("⚡ Quick Deploy"),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("✓ " + qdm.status),
			"",
			"q: Quit",
		)
		return style.Render(content)
	}

	if qdm.error != "" {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("⚡ Quick Deploy"),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("✗ Error: " + qdm.error),
			"",
			"q: Quit",
		)
		return style.Render(content)
	}

	return style.Render("Loading...")
}

// loadResourceTypes loads available resource types for quick deployment
func (qdm *QuickDeployManager) loadResourceTypes() tea.Cmd {
	return func() tea.Msg {
		resourceTypes := []resourceTypeItem{
			{
				title:        "Storage Account",
				description:  "Azure Storage Account with blob containers",
				resourceType: "storage_account",
			},
			{
				title:        "Virtual Machine",
				description:  "Linux/Windows virtual machine with networking",
				resourceType: "virtual_machine",
			},
			{
				title:        "AKS Cluster",
				description:  "Azure Kubernetes Service cluster",
				resourceType: "aks_cluster",
			},
			{
				title:        "Key Vault",
				description:  "Azure Key Vault for secrets management",
				resourceType: "key_vault",
			},
			{
				title:        "SQL Database",
				description:  "Azure SQL Database with server",
				resourceType: "sql_database",
			},
			{
				title:        "App Service",
				description:  "Azure App Service with service plan",
				resourceType: "app_service",
			},
		}

		items := make([]list.Item, len(resourceTypes))
		for i, rt := range resourceTypes {
			items[i] = rt
		}

		return resourceTypesLoadedMsg{items}
	}
}

// setupParameters sets up parameter inputs for the selected resource type
func (qdm *QuickDeployManager) setupParameters(resourceType string) {
	qdm.paramInputs = make(map[string]textinput.Model)
	qdm.parameters = make(map[string]string)
	qdm.paramIndex = 0

	switch resourceType {
	case "storage_account":
		qdm.paramKeys = []string{"name", "location", "sku"}
		qdm.setupInput("name", "Storage account name (3-24 chars, lowercase)", "mystorageaccount")
		qdm.setupInput("location", "Azure region", "East US")
		qdm.setupInput("sku", "Storage SKU (Standard_LRS, Standard_GRS, etc.)", "Standard_LRS")

	case "virtual_machine":
		qdm.paramKeys = []string{"name", "location", "size", "admin_username"}
		qdm.setupInput("name", "VM name", "myvm")
		qdm.setupInput("location", "Azure region", "East US")
		qdm.setupInput("size", "VM size", "Standard_B1s")
		qdm.setupInput("admin_username", "Admin username", "azureuser")

	case "aks_cluster":
		qdm.paramKeys = []string{"name", "location", "node_count"}
		qdm.setupInput("name", "AKS cluster name", "myakscluster")
		qdm.setupInput("location", "Azure region", "East US")
		qdm.setupInput("node_count", "Number of nodes", "3")

	case "key_vault":
		qdm.paramKeys = []string{"name", "location", "tenant_id"}
		qdm.setupInput("name", "Key Vault name", "mykeyvault")
		qdm.setupInput("location", "Azure region", "East US")
		qdm.setupInput("tenant_id", "Azure tenant ID", "")

	case "sql_database":
		qdm.paramKeys = []string{"name", "location", "server_name", "admin_login"}
		qdm.setupInput("name", "Database name", "mydatabase")
		qdm.setupInput("location", "Azure region", "East US")
		qdm.setupInput("server_name", "SQL server name", "myserver")
		qdm.setupInput("admin_login", "Admin login", "sqladmin")

	case "app_service":
		qdm.paramKeys = []string{"name", "location", "plan_name", "plan_sku"}
		qdm.setupInput("name", "App Service name", "myappservice")
		qdm.setupInput("location", "Azure region", "East US")
		qdm.setupInput("plan_name", "App Service plan name", "myplan")
		qdm.setupInput("plan_sku", "Plan SKU", "B1")

	default:
		qdm.paramKeys = []string{"name", "location"}
		qdm.setupInput("name", "Resource name", "myresource")
		qdm.setupInput("location", "Azure region", "East US")
	}
}

// setupInput creates and configures a text input for a parameter
func (qdm *QuickDeployManager) setupInput(key, placeholder, defaultValue string) {
	input := textinput.New()
	input.Placeholder = placeholder
	input.CharLimit = 100
	input.Width = 50
	if defaultValue != "" {
		input.SetValue(defaultValue)
	}
	qdm.paramInputs[key] = input
}

// focusCurrentParameter focuses the current parameter input
func (qdm *QuickDeployManager) focusCurrentParameter() {
	if qdm.paramIndex < len(qdm.paramKeys) {
		key := qdm.paramKeys[qdm.paramIndex]
		if input, exists := qdm.paramInputs[key]; exists {
			input.Focus()
			qdm.paramInputs[key] = input
		}
	}
}

// blurAllInputs blurs all parameter inputs
func (qdm *QuickDeployManager) blurAllInputs() {
	for key, input := range qdm.paramInputs {
		input.Blur()
		qdm.paramInputs[key] = input
	}
}

// deployResource deploys the selected resource with parameters
func (qdm *QuickDeployManager) deployResource() tea.Cmd {
	return func() tea.Msg {
		resourceType := qdm.resourceTypes.SelectedItem().(resourceTypeItem).resourceType
		cfg := config.GetTerraformConfig()
		
		// Create temporary workspace for deployment
		workspaceName := fmt.Sprintf("quick-deploy-%s-%d", resourceType, time.Now().Unix())
		workspacePath := filepath.Join(cfg.WorkspacePath, workspaceName)

		// Create workspace directory
		if err := os.MkdirAll(workspacePath, 0755); err != nil {
			return errorMsg{fmt.Errorf("failed to create workspace: %v", err)}
		}

		// Generate Terraform configuration
		tfConfig, err := qdm.generateTerraformConfig(resourceType)
		if err != nil {
			return errorMsg{fmt.Errorf("failed to generate config: %v", err)}
		}

		// Write main.tf
		mainTfPath := filepath.Join(workspacePath, "main.tf")
		if err := os.WriteFile(mainTfPath, []byte(tfConfig), 0644); err != nil {
			return errorMsg{fmt.Errorf("failed to write main.tf: %v", err)}
		}

		// Write variables.tf
		variablesTfPath := filepath.Join(workspacePath, "variables.tf")
		variablesConfig := qdm.generateVariablesConfig(resourceType)
		if err := os.WriteFile(variablesTfPath, []byte(variablesConfig), 0644); err != nil {
			return errorMsg{fmt.Errorf("failed to write variables.tf: %v", err)}
		}

		// Write terraform.tfvars
		tfvarsPath := filepath.Join(workspacePath, "terraform.tfvars")
		tfvarsContent := qdm.generateTfvarsContent()
		if err := os.WriteFile(tfvarsPath, []byte(tfvarsContent), 0644); err != nil {
			return errorMsg{fmt.Errorf("failed to write terraform.tfvars: %v", err)}
		}

		// Initialize and apply
		manager := tfbicep.NewTerraformManager(workspacePath)
		
		// Init
		if _, err := manager.Init(); err != nil {
			return errorMsg{fmt.Errorf("terraform init failed: %v", err)}
		}

		// Apply
		if _, err := manager.Apply(); err != nil {
			return errorMsg{fmt.Errorf("terraform apply failed: %v", err)}
		}

		return deploymentCompletedMsg{
			message: fmt.Sprintf("Resource deployed successfully in workspace: %s", workspaceName),
		}
	}
}

// generateTerraformConfig generates Terraform configuration for the resource type
func (qdm *QuickDeployManager) generateTerraformConfig(resourceType string) (string, error) {
	switch resourceType {
	case "storage_account":
		return `terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = "${var.name}-rg"
  location = var.location
}

resource "azurerm_storage_account" "main" {
  name                     = var.name
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = var.sku
}

output "storage_account_name" {
  value = azurerm_storage_account.main.name
}

output "storage_account_id" {
  value = azurerm_storage_account.main.id
}`, nil

	case "virtual_machine":
		return `terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = "${var.name}-rg"
  location = var.location
}

resource "azurerm_virtual_network" "main" {
  name                = "${var.name}-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}

resource "azurerm_subnet" "main" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.2.0/24"]
}

resource "azurerm_public_ip" "main" {
  name                = "${var.name}-pip"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  allocation_method   = "Static"
}

resource "azurerm_network_security_group" "main" {
  name                = "${var.name}-nsg"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  security_rule {
    name                       = "SSH"
    priority                   = 1001
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

resource "azurerm_network_interface" "main" {
  name                = "${var.name}-nic"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.main.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.main.id
  }
}

resource "azurerm_network_interface_security_group_association" "main" {
  network_interface_id      = azurerm_network_interface.main.id
  network_security_group_id = azurerm_network_security_group.main.id
}

resource "azurerm_linux_virtual_machine" "main" {
  name                = var.name
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  size                = var.size
  admin_username      = var.admin_username

  disable_password_authentication = true

  network_interface_ids = [
    azurerm_network_interface.main.id,
  ]

  admin_ssh_key {
    username   = var.admin_username
    public_key = file("~/.ssh/id_rsa.pub")
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-focal"
    sku       = "20_04-lts-gen2"
    version   = "latest"
  }
}

output "vm_public_ip" {
  value = azurerm_public_ip.main.ip_address
}

output "vm_id" {
  value = azurerm_linux_virtual_machine.main.id
}`, nil

	default:
		return "", fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// generateVariablesConfig generates variables.tf content
func (qdm *QuickDeployManager) generateVariablesConfig(resourceType string) string {
	switch resourceType {
	case "storage_account":
		return `variable "name" {
  description = "Storage account name"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "sku" {
  description = "Storage account SKU"
  type        = string
  default     = "Standard_LRS"
}`

	case "virtual_machine":
		return `variable "name" {
  description = "VM name"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}

variable "size" {
  description = "VM size"
  type        = string
  default     = "Standard_B1s"
}

variable "admin_username" {
  description = "Admin username"
  type        = string
}`

	default:
		return `variable "name" {
  description = "Resource name"
  type        = string
}

variable "location" {
  description = "Azure region"
  type        = string
}`
	}
}

// generateTfvarsContent generates terraform.tfvars content
func (qdm *QuickDeployManager) generateTfvarsContent() string {
	var lines []string
	for key, value := range qdm.parameters {
		lines = append(lines, fmt.Sprintf("%s = \"%s\"", key, value))
	}
	return strings.Join(lines, "\n")
}

// Additional message types
type resourceTypesLoadedMsg struct {
	items []list.Item
}

type deploymentCompletedMsg struct {
	message string
}

// resourceTypeItem represents a resource type in the list
type resourceTypeItem struct {
	title        string
	description  string
	resourceType string
}

func (i resourceTypeItem) FilterValue() string { return i.title }
func (i resourceTypeItem) Title() string       { return i.title }
func (i resourceTypeItem) Description() string { return i.description }

// QuickDeployProgram creates a standalone program for quick deployment
func QuickDeployProgram() *tea.Program {
	manager := NewQuickDeployManager()
	return tea.NewProgram(manager, tea.WithAltScreen())
}

// StateViewerTUI handles viewing and managing Terraform state
type StateViewerTUI struct {
	workspaces        list.Model
	resources         list.Model
	selectedWorkspace string
	selectedResource  string
	stateData         *tfbicep.TerraformStateInfo
	view              int // 0: workspaces, 1: resources, 2: resource details
	status            string
	error             string
}

// NewStateViewerTUI creates a new state viewer
func NewStateViewerTUI() *StateViewerTUI {
	// Create workspaces list
	workspaces := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 10)
	workspaces.Title = "Select Workspace"

	// Create resources list
	resources := list.New([]list.Item{}, list.NewDefaultDelegate(), 50, 10)
	resources.Title = "State Resources"

	return &StateViewerTUI{
		workspaces: workspaces,
		resources:  resources,
		view:       0,
	}
}

// Init initializes the state viewer
func (sv *StateViewerTUI) Init() tea.Cmd {
	return sv.loadWorkspaces()
}

// Update handles updates for the state viewer
func (sv *StateViewerTUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return sv, tea.Quit
		case "enter":
			switch sv.view {
			case 0: // Workspace selection
				if item, ok := sv.workspaces.SelectedItem().(workspaceItem); ok {
					sv.selectedWorkspace = item.path
					sv.view = 1
					return sv, sv.loadStateResources()
				}
			case 1: // Resource selection
				if item, ok := sv.resources.SelectedItem().(stateResourceItem); ok {
					sv.selectedResource = item.address
					sv.view = 2
					return sv, sv.loadResourceDetails()
				}
			}
		case "esc":
			if sv.view > 0 {
				sv.view--
				if sv.view == 0 {
					sv.selectedWorkspace = ""
					sv.selectedResource = ""
				}
			}
		case "r":
			// Refresh current view
			switch sv.view {
			case 0:
				return sv, sv.loadWorkspaces()
			case 1:
				return sv, sv.loadStateResources()
			case 2:
				return sv, sv.loadResourceDetails()
			}
		}
	case workspacesLoadedMsg:
		sv.workspaces.SetItems(msg.items)
	case stateResourcesLoadedMsg:
		// Convert StateResource to list items
		var items []list.Item
		for _, resource := range msg.resources {
			items = append(items, stateResourceItem{
				address:      resource.Address,
				resourceType: resource.Type,
				name:         resource.Name,
				status:       resource.Status,
			})
		}
		sv.resources.SetItems(items)
	case stateLoadedMsg:
		sv.stateData = msg.state
		sv.status = "State loaded successfully"
	case errorMsg:
		sv.error = msg.Error()
	}

	// Update components based on current view
	switch sv.view {
	case 0:
		sv.workspaces, cmd = sv.workspaces.Update(msg)
		cmds = append(cmds, cmd)
	case 1:
		sv.resources, cmd = sv.resources.Update(msg)
		cmds = append(cmds, cmd)
	}

	return sv, tea.Batch(cmds...)
}

// View renders the state viewer
func (sv *StateViewerTUI) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#8BE9FD")).
		Padding(1, 2).
		Margin(1, 0)

	switch sv.view {
	case 0:
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("📊 Terraform State Viewer"),
			"",
			sv.workspaces.View(),
			"",
			"Enter: Select | r: Refresh | q: Quit",
		)
		return style.Render(content)

	case 1:
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("📊 Terraform State Viewer"),
			"",
			fmt.Sprintf("Workspace: %s", filepath.Base(sv.selectedWorkspace)),
			"",
			sv.resources.View(),
			"",
			"Enter: View Details | Esc: Back | r: Refresh | q: Quit",
		)
		return style.Render(content)

	case 2:
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("📊 Terraform State Viewer"),
			"",
			fmt.Sprintf("Resource: %s", sv.selectedResource),
			"",
			sv.renderResourceDetails(),
			"",
			"Esc: Back | r: Refresh | q: Quit",
		)
		return style.Render(content)
	}

	if sv.error != "" {
		content := lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.NewStyle().Bold(true).Render("📊 Terraform State Viewer"),
			"",
			lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Render("✗ Error: " + sv.error),
			"",
			"q: Quit",
		)
		return style.Render(content)
	}

	return style.Render("Loading...")
}

// loadWorkspaces loads available workspaces
func (sv *StateViewerTUI) loadWorkspaces() tea.Cmd {
	return func() tea.Msg {
		cfg := config.GetTerraformConfig()
		workspaces := []string{}

		// Scan workspace directory
		if _, err := os.Stat(cfg.WorkspacePath); err == nil {
			entries, err := os.ReadDir(cfg.WorkspacePath)
			if err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						// Check if it's a valid Terraform workspace
						workspacePath := filepath.Join(cfg.WorkspacePath, entry.Name())
						if _, err := os.Stat(filepath.Join(workspacePath, "main.tf")); err == nil {
							workspaces = append(workspaces, entry.Name())
						}
					}
				}
			}
		}

		var items []list.Item
		for _, workspace := range workspaces {
			workspacePath := filepath.Join(cfg.WorkspacePath, workspace)
			
			description := "Terraform workspace"
			// Try to get resource count
			manager := tfbicep.NewTerraformManager(workspacePath)
			if count, err := manager.GetResourceCount(); err == nil {
				description = fmt.Sprintf("%d resources", count)
			}

			items = append(items, workspaceItem{
				title:       workspace,
				description: description,
				path:        workspacePath,
			})
		}

		return workspacesLoadedMsg{items}
	}
}

// loadStateResources loads resources from the selected workspace state
func (sv *StateViewerTUI) loadStateResources() tea.Cmd {
	return func() tea.Msg {
		if sv.selectedWorkspace == "" {
			return errorMsg{fmt.Errorf("no workspace selected")}
		}

		manager := tfbicep.NewTerraformManager(sv.selectedWorkspace)
		operation, err := manager.StateList()
		if err != nil {
			return errorMsg{fmt.Errorf("failed to load state: %v", err)}
		}

		if !operation.Success {
			return errorMsg{fmt.Errorf("terraform state list failed: %s", operation.Error)}
		}

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

				status := "ok"
				tainted := false
				if strings.Contains(line, "(tainted)") {
					status = "tainted"
					tainted = true
				}

				resources = append(resources, StateResource{
					Address:  line,
					Type:     resourceType,
					Name:     resourceName,
					Status:   status,
					Tainted:  tainted,
					Provider: strings.Split(resourceType, "_")[0],
				})
			}
		}

		return stateResourcesLoadedMsg{resources: resources}
	}
}

// loadResourceDetails loads detailed information about the selected resource
func (sv *StateViewerTUI) loadResourceDetails() tea.Cmd {
	return func() tea.Msg {
		if sv.selectedWorkspace == "" || sv.selectedResource == "" {
			return errorMsg{fmt.Errorf("no resource selected")}
		}

		manager := tfbicep.NewTerraformManager(sv.selectedWorkspace)
		operation, err := manager.StateShow(sv.selectedResource)
		if err != nil {
			return errorMsg{fmt.Errorf("failed to load resource details: %v", err)}
		}

		if !operation.Success {
			return errorMsg{fmt.Errorf("terraform state show failed: %s", operation.Error)}
		}

		return resourceDetailsLoadedMsg{
			address: sv.selectedResource,
			details: operation.Output,
		}
	}
}

// renderResourceDetails renders detailed resource information
func (sv *StateViewerTUI) renderResourceDetails() string {
	if sv.selectedResource == "" {
		return "No resource selected"
	}

	// This would normally parse the detailed state output
	// For now, show a simple representation
	var details []string
	details = append(details, fmt.Sprintf("Address: %s", sv.selectedResource))
	details = append(details, "")
	details = append(details, "Detailed state information would appear here")
	details = append(details, "This includes resource attributes, dependencies, and metadata")

	return strings.Join(details, "\n")
}

// Additional message types for state viewer (extending tui.go types)
type stateLoadedMsg struct {
	state *tfbicep.TerraformStateInfo
}

type resourceDetailsLoadedMsg struct {
	address string
	details string
}

// stateResourceItem represents a resource in the state list
type stateResourceItem struct {
	address      string
	resourceType string
	name         string
	status       string
}

func (i stateResourceItem) FilterValue() string { return i.address }
func (i stateResourceItem) Title() string       { return i.address }
func (i stateResourceItem) Description() string { return i.status }

// Note: workspaceItem is already defined in tui.go

// StateViewerProgram creates a standalone program for state viewing
func StateViewerProgram() *tea.Program {
	viewer := NewStateViewerTUI()
	return tea.NewProgram(viewer, tea.WithAltScreen())
}
