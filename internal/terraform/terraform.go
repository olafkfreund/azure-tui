package terraform

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/olafkfreund/azure-tui/internal/config"
)

// TerraformManager handles Terraform operations
type TerraformManager struct {
	WorkingDir string
	Config     *config.TerraformConfig
}

// TerraformFile represents a Terraform file
type TerraformFile struct {
	Name     string
	Path     string
	Content  string
	Modified time.Time
	Size     int64
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
	Mode      string                   `json:"mode"`
	Type      string                   `json:"type"`
	Name      string                   `json:"name"`
	Provider  string                   `json:"provider"`
	Instances []map[string]interface{} `json:"instances"`
}

// PlanOutput represents Terraform plan output
type PlanOutput struct {
	Add     int
	Change  int
	Destroy int
	Output  string
}

// NewTerraformManager creates a new Terraform manager
func NewTerraformManager() (*TerraformManager, error) {
	cfg, err := config.GetTerraformConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load terraform config: %w", err)
	}

	// Ensure source folder exists
	if err := os.MkdirAll(cfg.SourceFolder, 0755); err != nil {
		return nil, fmt.Errorf("failed to create terraform source folder: %w", err)
	}

	tm := &TerraformManager{
		WorkingDir: cfg.SourceFolder,
		Config:     cfg,
	}

	// Auto-initialize if configured
	if cfg.AutoInit {
		if err := tm.Init(); err != nil {
			// Don't fail if init fails, just log it
			fmt.Printf("Warning: Terraform init failed: %v\n", err)
		}
	}

	return tm, nil
}

// ListFiles returns all Terraform files in the working directory
func (tm *TerraformManager) ListFiles() ([]TerraformFile, error) {
	var files []TerraformFile

	err := filepath.WalkDir(tm.WorkingDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		// Include .tf, .tfvars, and .tfstate files
		ext := filepath.Ext(path)
		if ext == ".tf" || ext == ".tfvars" || ext == ".tfstate" {
			info, err := d.Info()
			if err != nil {
				return err
			}

			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}

			relPath, err := filepath.Rel(tm.WorkingDir, path)
			if err != nil {
				relPath = path
			}

			files = append(files, TerraformFile{
				Name:     d.Name(),
				Path:     relPath,
				Content:  string(content),
				Modified: info.ModTime(),
				Size:     info.Size(),
			})
		}

		return nil
	})

	return files, err
}

// CreateFile creates a new Terraform file
func (tm *TerraformManager) CreateFile(filename, content string) error {
	filePath := filepath.Join(tm.WorkingDir, filename)

	// Ensure it has .tf extension if not specified
	if !strings.HasSuffix(filename, ".tf") && !strings.HasSuffix(filename, ".tfvars") {
		filename += ".tf"
		filePath = filepath.Join(tm.WorkingDir, filename)
	}

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("file %s already exists", filename)
	}

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	// Auto-format if configured
	if tm.Config.AutoFormat {
		if err := tm.Format(); err != nil {
			fmt.Printf("Warning: failed to format file: %v\n", err)
		}
	}

	return nil
}

// UpdateFile updates an existing Terraform file
func (tm *TerraformManager) UpdateFile(filename, content string) error {
	filePath := filepath.Join(tm.WorkingDir, filename)

	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	// Auto-format if configured
	if tm.Config.AutoFormat {
		if err := tm.Format(); err != nil {
			fmt.Printf("Warning: failed to format file: %v\n", err)
		}
	}

	// Validate if configured
	if tm.Config.ValidateOnSave {
		if err := tm.Validate(); err != nil {
			fmt.Printf("Warning: validation failed: %v\n", err)
		}
	}

	return nil
}

// DeleteFile deletes a Terraform file
func (tm *TerraformManager) DeleteFile(filename string) error {
	filePath := filepath.Join(tm.WorkingDir, filename)

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// ReadFile reads the content of a Terraform file
func (tm *TerraformManager) ReadFile(filename string) (string, error) {
	filePath := filepath.Join(tm.WorkingDir, filename)

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

// Init initializes Terraform in the working directory
func (tm *TerraformManager) Init() error {
	cmd := exec.Command("terraform", "init")
	cmd.Dir = tm.WorkingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform init failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Plan runs terraform plan and returns the output
func (tm *TerraformManager) Plan() (*PlanOutput, error) {
	cmd := exec.Command("terraform", "plan", "-detailed-exitcode")
	cmd.Dir = tm.WorkingDir

	output, err := cmd.CombinedOutput()
	outputStr := string(output)

	planOutput := &PlanOutput{
		Output: outputStr,
	}

	// Parse the output for resource counts
	lines := strings.Split(outputStr, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "Plan:") {
			// Extract numbers from "Plan: X to add, Y to change, Z to destroy."
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "add," && i > 0 {
					fmt.Sscanf(parts[i-1], "%d", &planOutput.Add)
				} else if part == "change," && i > 0 {
					fmt.Sscanf(parts[i-1], "%d", &planOutput.Change)
				} else if part == "destroy." && i > 0 {
					fmt.Sscanf(parts[i-1], "%d", &planOutput.Destroy)
				}
			}
		}
	}

	// If exit code is 2, there are changes but no error
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 2 {
			return planOutput, nil
		}
		return planOutput, fmt.Errorf("terraform plan failed: %w", err)
	}

	return planOutput, nil
}

// Apply runs terraform apply
func (tm *TerraformManager) Apply() error {
	cmd := exec.Command("terraform", "apply", "-auto-approve")
	cmd.Dir = tm.WorkingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform apply failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Destroy runs terraform destroy
func (tm *TerraformManager) Destroy() error {
	if tm.Config.ConfirmDestroy {
		// In a real implementation, this would show a confirmation dialog
		fmt.Println("Warning: This will destroy all resources!")
	}

	cmd := exec.Command("terraform", "destroy", "-auto-approve")
	cmd.Dir = tm.WorkingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform destroy failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Format runs terraform fmt
func (tm *TerraformManager) Format() error {
	cmd := exec.Command("terraform", "fmt")
	cmd.Dir = tm.WorkingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform fmt failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// Validate runs terraform validate
func (tm *TerraformManager) Validate() error {
	cmd := exec.Command("terraform", "validate")
	cmd.Dir = tm.WorkingDir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("terraform validate failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GetState reads the Terraform state file
func (tm *TerraformManager) GetState() (*TerraformState, error) {
	stateFile := filepath.Join(tm.WorkingDir, "terraform.tfstate")

	if _, err := os.Stat(stateFile); os.IsNotExist(err) {
		return &TerraformState{}, nil // Return empty state if file doesn't exist
	}

	// For now, just return basic info
	// In a real implementation, you'd parse the JSON state file
	return &TerraformState{
		Version:   4,
		Resources: []TerraformResource{},
		Outputs:   map[string]interface{}{},
	}, nil
}

// CheckTerraformInstalled checks if Terraform is installed
func CheckTerraformInstalled() bool {
	cmd := exec.Command("terraform", "version")
	err := cmd.Run()
	return err == nil
}

// GetTerraformVersion returns the installed Terraform version
func GetTerraformVersion() (string, error) {
	cmd := exec.Command("terraform", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// Parse version from output
	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "unknown", nil
}
