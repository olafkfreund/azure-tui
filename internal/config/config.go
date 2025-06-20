package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type NamingConfig struct {
	VM      string `yaml:"vm"`
	Storage string `yaml:"storage"`
	VNet    string `yaml:"vnet"`
	Default string `yaml:"default"`
}

type AIConfig struct {
	Provider            string `yaml:"provider"`
	Model               string `yaml:"model"`
	ConfirmBeforeWrite  bool   `yaml:"confirm_before_write"`
	ConfirmBeforeDeploy bool   `yaml:"confirm_before_deploy"`
}

type TerraformConfig struct {
	WorkspacePath  string            `yaml:"workspace_path"`
	TemplatesPath  string            `yaml:"templates_path"`
	StatePath      string            `yaml:"state_path"`
	DefaultEditor  string            `yaml:"default_editor"`
	AutoSave       bool              `yaml:"auto_save"`
	AutoFormat     bool              `yaml:"auto_format"`
	ValidateOnSave bool              `yaml:"validate_on_save"`
	BackendType    string            `yaml:"backend_type"`
	BackendConfig  map[string]string `yaml:"backend_config"`
	ModuleSources  []string          `yaml:"module_sources"`
	VariableFiles  []string          `yaml:"variable_files"`
}

type EditorConfig struct {
	DefaultEditor  string            `yaml:"default_editor"`
	EditorArgs     []string          `yaml:"editor_args"`
	TempDir        string            `yaml:"temp_dir"`
	FileExtensions map[string]string `yaml:"file_extensions"`
}

type UIConfig struct {
	ShowTerraformMenu  bool              `yaml:"show_terraform_menu"`
	TerraformShortcuts map[string]string `yaml:"terraform_shortcuts"`
	PopupWidth         int               `yaml:"popup_width"`
	PopupHeight        int               `yaml:"popup_height"`
	EnableMouseSupport bool              `yaml:"enable_mouse_support"`
	ColorScheme        string            `yaml:"color_scheme"`
}

type AppConfig struct {
	Naming    NamingConfig    `yaml:"naming"`
	Env       string          `yaml:"env"`
	AI        AIConfig        `yaml:"ai"`
	Terraform TerraformConfig `yaml:"terraform"`
	Editor    EditorConfig    `yaml:"editor"`
	UI        UIConfig        `yaml:"ui"`
}

var loadedConfig *AppConfig

func LoadConfig() (*AppConfig, error) {
	if loadedConfig != nil {
		return loadedConfig, nil
	}
	path := filepath.Join(os.Getenv("HOME"), ".config", "azure-tui", "config.yaml")
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := f.Close(); closeErr != nil {
			// Log the error but don't override the main error
			fmt.Printf("Warning: failed to close config file: %v\n", closeErr)
		}
	}()
	var cfg AppConfig
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}
	loadedConfig = &cfg
	return &cfg, nil
}

func GetNamingStandard(resourceType string) string {
	cfg, err := LoadConfig()
	if err != nil {
		return "{{env}}-{{type}}-{{name}}"
	}
	switch resourceType {
	case "vm":
		if cfg.Naming.VM != "" {
			return cfg.Naming.VM
		}
	case "storage":
		if cfg.Naming.Storage != "" {
			return cfg.Naming.Storage
		}
	case "vnet":
		if cfg.Naming.VNet != "" {
			return cfg.Naming.VNet
		}
	}
	if cfg.Naming.Default != "" {
		return cfg.Naming.Default
	}
	return "{{env}}-{{type}}-{{name}}"
}

// GetTerraformConfig returns the Terraform configuration with defaults
func GetTerraformConfig() TerraformConfig {
	cfg, err := LoadConfig()
	if err != nil {
		return getDefaultTerraformConfig()
	}

	// Apply defaults for empty values
	if cfg.Terraform.WorkspacePath == "" {
		cfg.Terraform.WorkspacePath = filepath.Join(os.Getenv("HOME"), ".config", "azure-tui", "terraform", "workspaces")
	}
	if cfg.Terraform.TemplatesPath == "" {
		cfg.Terraform.TemplatesPath = "./terraform/templates"
	}
	if cfg.Terraform.StatePath == "" {
		cfg.Terraform.StatePath = filepath.Join(os.Getenv("HOME"), ".config", "azure-tui", "terraform", "state")
	}
	if cfg.Terraform.DefaultEditor == "" {
		cfg.Terraform.DefaultEditor = getDefaultEditor()
	}
	if cfg.Terraform.BackendType == "" {
		cfg.Terraform.BackendType = "local"
	}

	return cfg.Terraform
}

// GetEditorConfig returns the editor configuration with defaults
func GetEditorConfig() EditorConfig {
	cfg, err := LoadConfig()
	if err != nil {
		return getDefaultEditorConfig()
	}

	if cfg.Editor.DefaultEditor == "" {
		cfg.Editor.DefaultEditor = getDefaultEditor()
	}
	if cfg.Editor.TempDir == "" {
		cfg.Editor.TempDir = os.TempDir()
	}
	if len(cfg.Editor.FileExtensions) == 0 {
		cfg.Editor.FileExtensions = map[string]string{
			"terraform": ".tf",
			"variables": ".tfvars",
			"output":    ".tf",
		}
	}

	return cfg.Editor
}

// GetUIConfig returns the UI configuration with defaults
func GetUIConfig() UIConfig {
	cfg, err := LoadConfig()
	if err != nil {
		return getDefaultUIConfig()
	}

	if cfg.UI.PopupWidth == 0 {
		cfg.UI.PopupWidth = 80
	}
	if cfg.UI.PopupHeight == 0 {
		cfg.UI.PopupHeight = 24
	}
	if len(cfg.UI.TerraformShortcuts) == 0 {
		cfg.UI.TerraformShortcuts = getDefaultTerraformShortcuts()
	}
	if cfg.UI.ColorScheme == "" {
		cfg.UI.ColorScheme = "azure"
	}

	return cfg.UI
}

func getDefaultTerraformConfig() TerraformConfig {
	return TerraformConfig{
		WorkspacePath:  filepath.Join(os.Getenv("HOME"), ".config", "azure-tui", "terraform", "workspaces"),
		TemplatesPath:  "./terraform/templates",
		StatePath:      filepath.Join(os.Getenv("HOME"), ".config", "azure-tui", "terraform", "state"),
		DefaultEditor:  getDefaultEditor(),
		AutoSave:       true,
		AutoFormat:     true,
		ValidateOnSave: true,
		BackendType:    "local",
		BackendConfig:  make(map[string]string),
		ModuleSources:  []string{},
		VariableFiles:  []string{},
	}
}

func getDefaultEditorConfig() EditorConfig {
	return EditorConfig{
		DefaultEditor: getDefaultEditor(),
		EditorArgs:    []string{},
		TempDir:       os.TempDir(),
		FileExtensions: map[string]string{
			"terraform": ".tf",
			"variables": ".tfvars",
			"output":    ".tf",
		},
	}
}

func getDefaultUIConfig() UIConfig {
	return UIConfig{
		ShowTerraformMenu:  true,
		TerraformShortcuts: getDefaultTerraformShortcuts(),
		PopupWidth:         80,
		PopupHeight:        24,
		EnableMouseSupport: true,
		ColorScheme:        "azure",
	}
}

func getDefaultTerraformShortcuts() map[string]string {
	return map[string]string{
		"ctrl+t": "terraform_menu",
		"ctrl+n": "new_terraform_file",
		"ctrl+e": "edit_terraform_file",
		"ctrl+p": "terraform_plan",
		"ctrl+a": "terraform_apply",
		"ctrl+d": "terraform_destroy",
		"ctrl+s": "terraform_state",
		"ctrl+i": "terraform_init",
		"ctrl+f": "terraform_format",
		"ctrl+v": "terraform_validate",
	}
}

func getDefaultEditor() string {
	// Check common environment variables
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}

	// Check for common editors
	editors := []string{"nvim", "vim", "nano", "code", "emacs"}
	for _, editor := range editors {
		if _, err := os.Stat("/usr/bin/" + editor); err == nil {
			return editor
		}
		if _, err := os.Stat("/usr/local/bin/" + editor); err == nil {
			return editor
		}
	}

	return "vi" // fallback
}

// SaveConfig saves the current configuration to file
func SaveConfig(cfg *AppConfig) error {
	configDir := filepath.Join(os.Getenv("HOME"), ".config", "azure-tui")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "config.yaml")
	f, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer f.Close()

	encoder := yaml.NewEncoder(f)
	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	// Reset loaded config to force reload
	loadedConfig = nil

	return nil
}

// EnsureTerraformDirectories creates necessary Terraform directories
func EnsureTerraformDirectories() error {
	tfConfig := GetTerraformConfig()

	dirs := []string{
		tfConfig.WorkspacePath,
		tfConfig.StatePath,
		filepath.Dir(tfConfig.TemplatesPath),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
