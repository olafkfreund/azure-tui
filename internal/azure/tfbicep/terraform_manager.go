package tfbicep

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// TerraformManager provides comprehensive Terraform operations
type TerraformManager struct {
	WorkingDir string
	Timeout    time.Duration
}

// NewTerraformManager creates a new Terraform manager
func NewTerraformManager(workingDir string) *TerraformManager {
	return &TerraformManager{
		WorkingDir: workingDir,
		Timeout:    5 * time.Minute,
	}
}

// TerraformOperation represents a Terraform operation result
type TerraformOperation struct {
	Command   string        `json:"command"`
	Directory string        `json:"directory"`
	Output    string        `json:"output"`
	Error     string        `json:"error"`
	Duration  time.Duration `json:"duration"`
	ExitCode  int           `json:"exit_code"`
	Success   bool          `json:"success"`
	StartTime time.Time     `json:"start_time"`
	EndTime   time.Time     `json:"end_time"`
}

// TerraformStateInfo represents Terraform state information
type TerraformStateInfo struct {
	Version   int                    `json:"version"`
	Serial    int                    `json:"serial"`
	Lineage   string                 `json:"lineage"`
	Resources []TerraformResource    `json:"resources"`
	Outputs   map[string]interface{} `json:"outputs"`
}

// TerraformResource represents a resource in Terraform state
type TerraformResource struct {
	Mode      string              `json:"mode"`
	Type      string              `json:"type"`
	Name      string              `json:"name"`
	Provider  string              `json:"provider"`
	Instances []TerraformInstance `json:"instances"`
}

// TerraformInstance represents a resource instance
type TerraformInstance struct {
	SchemaVersion int                    `json:"schema_version"`
	Attributes    map[string]interface{} `json:"attributes"`
	Status        string                 `json:"status"`
}

// TerraformPlanResult represents a Terraform plan
type TerraformPlanResult struct {
	ResourceChanges []ResourceChange `json:"resource_changes"`
	Configuration   interface{}      `json:"configuration"`
	PlannedValues   interface{}      `json:"planned_values"`
}

// ResourceChange represents a planned resource change
type ResourceChange struct {
	Address string `json:"address"`
	Mode    string `json:"mode"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Change  Change `json:"change"`
}

// Change represents the actual change details
type Change struct {
	Actions []string    `json:"actions"`
	Before  interface{} `json:"before"`
	After   interface{} `json:"after"`
}

// WorkspaceInfo represents Terraform workspace information
type WorkspaceInfo struct {
	Name      string `json:"name"`
	Current   bool   `json:"current"`
	Directory string `json:"directory"`
}

// Init initializes a Terraform working directory
func (tm *TerraformManager) Init() (*TerraformOperation, error) {
	return tm.runCommand("init", []string{})
}

// Plan creates an execution plan
func (tm *TerraformManager) Plan() (*TerraformOperation, error) {
	return tm.runCommand("plan", []string{})
}

// PlanWithOutput creates an execution plan and saves it to a file
func (tm *TerraformManager) PlanWithOutput(planFile string) (*TerraformOperation, error) {
	return tm.runCommand("plan", []string{"-out", planFile})
}

// PlanJSON creates an execution plan in JSON format
func (tm *TerraformManager) PlanJSON() (*TerraformPlanResult, error) {
	op, err := tm.runCommand("plan", []string{"-json"})
	if err != nil {
		return nil, err
	}

	var plan TerraformPlanResult
	if err := json.Unmarshal([]byte(op.Output), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	return &plan, nil
}

// Apply applies the changes
func (tm *TerraformManager) Apply() (*TerraformOperation, error) {
	return tm.runCommand("apply", []string{"-auto-approve"})
}

// ApplyPlan applies a specific plan file
func (tm *TerraformManager) ApplyPlan(planFile string) (*TerraformOperation, error) {
	return tm.runCommand("apply", []string{planFile})
}

// Destroy destroys the infrastructure
func (tm *TerraformManager) Destroy() (*TerraformOperation, error) {
	return tm.runCommand("destroy", []string{"-auto-approve"})
}

// Validate validates the configuration
func (tm *TerraformManager) Validate() (*TerraformOperation, error) {
	return tm.runCommand("validate", []string{})
}

// Format formats the configuration files
func (tm *TerraformManager) Format() (*TerraformOperation, error) {
	return tm.runCommand("fmt", []string{"-recursive"})
}

// Show shows the current state or plan
func (tm *TerraformManager) Show() (*TerraformOperation, error) {
	return tm.runCommand("show", []string{})
}

// StateList lists resources in the state
func (tm *TerraformManager) StateList() (*TerraformOperation, error) {
	return tm.runCommand("state", []string{"list"})
}

// StateShow shows a specific resource in the state
func (tm *TerraformManager) StateShow(address string) (*TerraformOperation, error) {
	return tm.runCommand("state", []string{"show", address})
}

// StateRm removes a resource from the state
func (tm *TerraformManager) StateRm(address string) (*TerraformOperation, error) {
	return tm.runCommand("state", []string{"rm", address})
}

// Output gets output values
func (tm *TerraformManager) Output() (*TerraformOperation, error) {
	return tm.runCommand("output", []string{"-json"})
}

// OutputValue gets a specific output value
func (tm *TerraformManager) OutputValue(name string) (*TerraformOperation, error) {
	return tm.runCommand("output", []string{name})
}

// Refresh refreshes the state
func (tm *TerraformManager) Refresh() (*TerraformOperation, error) {
	return tm.runCommand("refresh", []string{})
}

// Import imports existing resources
func (tm *TerraformManager) Import(address, id string) (*TerraformOperation, error) {
	return tm.runCommand("import", []string{address, id})
}

// WorkspaceList lists all workspaces
func (tm *TerraformManager) WorkspaceList() (*TerraformOperation, error) {
	return tm.runCommand("workspace", []string{"list"})
}

// WorkspaceNew creates a new workspace
func (tm *TerraformManager) WorkspaceNew(name string) (*TerraformOperation, error) {
	return tm.runCommand("workspace", []string{"new", name})
}

// WorkspaceSelect switches to a workspace
func (tm *TerraformManager) WorkspaceSelect(name string) (*TerraformOperation, error) {
	return tm.runCommand("workspace", []string{"select", name})
}

// WorkspaceDelete deletes a workspace
func (tm *TerraformManager) WorkspaceDelete(name string) (*TerraformOperation, error) {
	return tm.runCommand("workspace", []string{"delete", name})
}

// GetState reads and parses the current Terraform state
func (tm *TerraformManager) GetState() (*TerraformStateInfo, error) {
	statePath := filepath.Join(tm.WorkingDir, "terraform.tfstate")

	data, err := os.ReadFile(statePath)
	if err != nil {
		if os.IsNotExist(err) {
			return &TerraformStateInfo{}, nil // Empty state if file doesn't exist
		}
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state TerraformStateInfo
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state JSON: %w", err)
	}

	return &state, nil
}

// GetResourceCount returns the number of resources in the state
func (tm *TerraformManager) GetResourceCount() (int, error) {
	state, err := tm.GetState()
	if err != nil {
		return 0, err
	}
	return len(state.Resources), nil
}

// ValidateConfig validates the Terraform configuration
func (tm *TerraformManager) ValidateConfig() (bool, []string, error) {
	op, err := tm.Validate()
	if err != nil {
		return false, nil, err
	}

	var issues []string
	if !op.Success {
		scanner := bufio.NewScanner(strings.NewReader(op.Error))
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				issues = append(issues, line)
			}
		}
	}

	return op.Success, issues, nil
}

// FormatFiles formats all Terraform files in the working directory
func (tm *TerraformManager) FormatFiles() error {
	op, err := tm.Format()
	if err != nil {
		return fmt.Errorf("terraform format failed: %s", op.Error)
	}
	return nil
}

// runCommand executes a Terraform command and captures detailed output
func (tm *TerraformManager) runCommand(command string, args []string) (*TerraformOperation, error) {
	start := time.Now()

	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command("terraform", cmdArgs...)
	cmd.Dir = tm.WorkingDir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	end := time.Now()
	duration := end.Sub(start)

	operation := &TerraformOperation{
		Command:   fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
		Directory: tm.WorkingDir,
		Output:    stdout.String(),
		Error:     stderr.String(),
		Duration:  duration,
		Success:   err == nil,
		StartTime: start,
		EndTime:   end,
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		operation.ExitCode = exitError.ExitCode()
	}

	return operation, err
}

// CheckTerraformInstalled checks if Terraform is installed and gets version
func CheckTerraformInstalled() (string, error) {
	cmd := exec.Command("terraform", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("terraform not found or not executable: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) > 0 {
		return strings.TrimSpace(lines[0]), nil
	}

	return "", fmt.Errorf("unable to parse terraform version")
}

// InitializeWorkspace initializes a new Terraform workspace with backend configuration
func InitializeWorkspace(dir string, backend map[string]string) error {
	// Ensure directory exists
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Create backend configuration if specified
	if len(backend) > 0 {
		backendConfig := generateBackendConfig(backend)
		backendPath := filepath.Join(dir, "backend.tf")
		if err := os.WriteFile(backendPath, []byte(backendConfig), 0644); err != nil {
			return fmt.Errorf("failed to write backend config: %w", err)
		}
	}

	// Initialize the workspace
	tm := NewTerraformManager(dir)
	op, err := tm.Init()
	if err != nil {
		return fmt.Errorf("terraform init failed: %s", op.Error)
	}

	return nil
}

// generateBackendConfig generates Terraform backend configuration
func generateBackendConfig(backend map[string]string) string {
	backendType := backend["type"]
	if backendType == "" {
		backendType = "local"
	}

	config := fmt.Sprintf(`terraform {
  backend "%s" {`, backendType)

	for key, value := range backend {
		if key != "type" {
			config += fmt.Sprintf(`
    %s = "%s"`, key, value)
		}
	}

	config += `
  }
}
`
	return config
}

// CopyTemplate copies a Terraform template to a workspace
func CopyTemplate(templatePath, workspacePath string) error {
	// Ensure workspace directory exists
	if err := os.MkdirAll(workspacePath, 0755); err != nil {
		return fmt.Errorf("failed to create workspace directory: %w", err)
	}

	// Walk through template directory
	return filepath.Walk(templatePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			return nil
		}

		// Calculate relative path
		relPath, err := filepath.Rel(templatePath, path)
		if err != nil {
			return err
		}

		// Create destination path
		destPath := filepath.Join(workspacePath, relPath)

		// Ensure destination directory exists
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return err
		}

		// Copy file
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		return os.WriteFile(destPath, data, info.Mode())
	})
}

// ListTemplates returns available Terraform templates
func ListTemplates(templatesPath string) ([]string, error) {
	var templates []string

	err := filepath.Walk(templatesPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() && path != templatesPath {
			relPath, err := filepath.Rel(templatesPath, path)
			if err != nil {
				return err
			}
			templates = append(templates, relPath)
		}

		return nil
	})

	return templates, err
}

// GetTemplateInfo returns information about a specific template
func GetTemplateInfo(templatePath string) (map[string]interface{}, error) {
	info := make(map[string]interface{})

	// Check if README exists
	readmePath := filepath.Join(templatePath, "README.md")
	if _, err := os.Stat(readmePath); err == nil {
		readme, err := os.ReadFile(readmePath)
		if err == nil {
			info["readme"] = string(readme)
		}
	}

	// List Terraform files
	var tfFiles []string
	err := filepath.Walk(templatePath, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fileInfo.IsDir() && strings.HasSuffix(path, ".tf") {
			relPath, err := filepath.Rel(templatePath, path)
			if err != nil {
				return err
			}
			tfFiles = append(tfFiles, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	info["terraform_files"] = tfFiles
	info["path"] = templatePath

	return info, nil
}
