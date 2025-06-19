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
	SourceFolder    string `yaml:"source_folder"`
	DefaultLocation string `yaml:"default_location"`
	AIProvider      string `yaml:"ai_provider"`
	AutoFormat      bool   `yaml:"auto_format"`
	ValidateOnSave  bool   `yaml:"validate_on_save"`
	StateBackend    string `yaml:"state_backend"`
	AutoInit        bool   `yaml:"auto_init"`
	ConfirmDestroy  bool   `yaml:"confirm_destroy"`
}

type AppConfig struct {
	Naming    NamingConfig    `yaml:"naming"`
	Env       string          `yaml:"env"`
	AI        AIConfig        `yaml:"ai"`
	Terraform TerraformConfig `yaml:"terraform"`
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

func GetTerraformConfig() (*TerraformConfig, error) {
	cfg, err := LoadConfig()
	if err != nil {
		// Return default config if no config file exists
		return &TerraformConfig{
			SourceFolder:    "./terraform",
			DefaultLocation: "uksouth",
			AIProvider:      "openai",
			AutoFormat:      true,
			ValidateOnSave:  true,
			StateBackend:    "local",
			AutoInit:        true,
			ConfirmDestroy:  true,
		}, nil
	}
	return &cfg.Terraform, nil
}

func GetTerraformSourceFolder() string {
	cfg, err := GetTerraformConfig()
	if err != nil || cfg.SourceFolder == "" {
		return "./terraform"
	}
	return cfg.SourceFolder
}

func GetDefaultLocation() string {
	cfg, err := GetTerraformConfig()
	if err != nil || cfg.DefaultLocation == "" {
		return "uksouth"
	}
	return cfg.DefaultLocation
}

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
	defer encoder.Close()

	if err := encoder.Encode(cfg); err != nil {
		return fmt.Errorf("failed to encode config: %w", err)
	}

	loadedConfig = cfg // Update the loaded config cache
	return nil
}
