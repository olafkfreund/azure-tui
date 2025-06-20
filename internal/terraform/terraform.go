package terraform

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/olafkfreund/azure-tui/internal/config"
)

// TerraformConfig holds configuration for Terraform operations
type TerraformConfig struct {
	WorkspacePath   string            `yaml:"workspace_path"`
	TemplatesPath   string            `yaml:"templates_path"`
	StatePath       string            `yaml:"state_path"`
	BackendType     string            `yaml:"backend_type"`
	BackendConfig   map[string]string `yaml:"backend_config"`
	DefaultTags     map[string]string `yaml:"default_tags"`
	AutoApprove     bool              `yaml:"auto_approve"`
	ParallelismFlag int               `yaml:"parallelism"`
}

// TerraformOperation represents a Terraform operation
type TerraformOperation struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"` // init, plan, apply, destroy
	WorkingDir string            `json:"working_dir"`
	Status     string            `json:"status"` // running, completed, failed
	StartTime  time.Time         `json:"start_time"`
	EndTime    time.Time         `json:"end_time"`
	Output     string            `json:"output"`
	Error      string            `json:"error"`
	Variables  map[string]string `json:"variables"`
}

// TerraformState represents Terraform state information
type TerraformState struct {
	Version          int                    `json:"version"`
	TerraformVersion string                 `json:"terraform_version"`
	Serial           int                    `json:"serial"`
	Lineage          string                 `json:"lineage"`
	Resources        []TerraformResource    `json:"resources"`
	Outputs          map[string]interface{} `json:"outputs"`
}

// TerraformResource represents a resource in Terraform state
type TerraformResource struct {
	Mode       string                 `json:"mode"`
	Type       string                 `json:"type"`
	Name       string                 `json:"name"`
	Provider   string                 `json:"provider"`
	Instances  []TerraformInstance    `json:"instances"`
	Attributes map[string]interface{} `json:"attributes"`
}

// TerraformInstance represents an instance of a Terraform resource
type TerraformInstance struct {
	SchemaVersion int                    `json:"schema_version"`
	Attributes    map[string]interface{} `json:"attributes"`
	Dependencies  []string               `json:"dependencies"`
}

// TerraformTemplate represents a Terraform template
type TerraformTemplate struct {
	Name        string             `json:"name"`
	Path        string             `json:"path"`
	Description string             `json:"description"`
	Category    string             `json:"category"`
	Variables   []TemplateVariable `json:"variables"`
	Outputs     []TemplateOutput   `json:"outputs"`
	Resources   []string           `json:"resources"`
}

// TemplateVariable represents a template variable
type TemplateVariable struct {
	Name         string      `json:"name"`
	Type         string      `json:"type"`
	Description  string      `json:"description"`
	DefaultValue interface{} `json:"default_value"`
	Required     bool        `json:"required"`
}

// TemplateOutput represents a template output
type TemplateOutput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Sensitive   bool   `json:"sensitive"`
}

// LoadTerraformConfig loads Terraform configuration from the app config
func LoadTerraformConfig() (*TerraformConfig, error) {
	appConfig, err := config.LoadConfig()
	if err != nil {
		// Return default config if no config file exists
		return &TerraformConfig{
			WorkspacePath: filepath.Join(os.Getenv("HOME"), ".config", "azure-tui", "terraform", "workspaces"),
			TemplatesPath: "./terraform/templates",
			StatePath:     filepath.Join(os.Getenv("HOME"), ".config", "azure-tui", "terraform", "state"),
			BackendType:   "local",
			DefaultTags: map[string]string{
				"ManagedBy": "Azure-TUI",
				"CreatedBy": "Terraform",
			},
			AutoApprove:     false,
			ParallelismFlag: 10,
		}, nil
	}

	// Extract terraform config if it exists in the app config
	// For now, return default config since we haven't extended the config struct yet
	return &TerraformConfig{
		WorkspacePath: filepath.Join(os.Getenv("HOME"), ".config", "azure-tui", "terraform", "workspaces"),
		TemplatesPath: "./terraform/templates",
		StatePath:     filepath.Join(os.Getenv("HOME"), ".config", "azure-tui", "terraform", "state"),
		BackendType:   "local",
		DefaultTags: map[string]string{
			"ManagedBy":   "Azure-TUI",
			"CreatedBy":   "Terraform",
			"Environment": appConfig.Env,
		},
		AutoApprove:     false,
		ParallelismFlag: 10,
	}, nil
}

// InitWorkspace initializes a Terraform workspace
func InitWorkspace(workspaceDir string) error {
	cmd := exec.Command("terraform", "init")
	cmd.Dir = workspaceDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init failed: %v\nOutput: %s", err, string(output))
	}
	return nil
}

// PlanWorkspace runs terraform plan in a workspace
func PlanWorkspace(workspaceDir string, varFile string) (string, error) {
	args := []string{"plan", "-detailed-exitcode"}
	if varFile != "" {
		args = append(args, "-var-file", varFile)
	}

	cmd := exec.Command("terraform", args...)
	cmd.Dir = workspaceDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			// Exit code 2 means there are changes to apply
			if exitError.ExitCode() == 2 {
				return string(output), nil
			}
		}
		return string(output), fmt.Errorf("terraform plan failed: %v", err)
	}
	return string(output), nil
}

// ApplyWorkspace applies Terraform changes in a workspace
func ApplyWorkspace(workspaceDir string, varFile string, autoApprove bool) (string, error) {
	args := []string{"apply"}
	if autoApprove {
		args = append(args, "-auto-approve")
	}
	if varFile != "" {
		args = append(args, "-var-file", varFile)
	}

	cmd := exec.Command("terraform", args...)
	cmd.Dir = workspaceDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("terraform apply failed: %v", err)
	}
	return string(output), nil
}

// DestroyWorkspace destroys Terraform-managed infrastructure
func DestroyWorkspace(workspaceDir string, varFile string, autoApprove bool) (string, error) {
	args := []string{"destroy"}
	if autoApprove {
		args = append(args, "-auto-approve")
	}
	if varFile != "" {
		args = append(args, "-var-file", varFile)
	}

	cmd := exec.Command("terraform", args...)
	cmd.Dir = workspaceDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("terraform destroy failed: %v", err)
	}
	return string(output), nil
}

// ListTemplates lists available Terraform templates
func ListTemplates(templatesPath string) ([]TerraformTemplate, error) {
	var templates []TerraformTemplate

	err := filepath.Walk(templatesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for main.tf files to identify templates
		if info.Name() == "main.tf" {
			templateDir := filepath.Dir(path)
			relPath, _ := filepath.Rel(templatesPath, templateDir)

			// Parse template information
			template := TerraformTemplate{
				Name:        filepath.Base(templateDir),
				Path:        templateDir,
				Category:    strings.Split(relPath, string(filepath.Separator))[0],
				Description: fmt.Sprintf("Terraform template for %s", filepath.Base(templateDir)),
			}

			templates = append(templates, template)
		}

		return nil
	})

	return templates, err
}

// CreateWorkspaceFromTemplate creates a new workspace from a template
func CreateWorkspaceFromTemplate(templatePath, workspacePath, workspaceName string) error {
	workspaceDir := filepath.Join(workspacePath, workspaceName)

	// Create workspace directory
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %v", err)
	}

	// Copy template files to workspace
	return copyDir(templatePath, workspaceDir)
}

// copyDir recursively copies a directory
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = dstFile.ReadFrom(srcFile)
	return err
}
