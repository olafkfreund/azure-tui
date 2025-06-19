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

type AppConfig struct {
	Naming NamingConfig `yaml:"naming"`
	Env    string       `yaml:"env"`
	AI     AIConfig     `yaml:"ai"`
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
