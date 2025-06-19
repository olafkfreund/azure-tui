package terraform

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/olafkfreund/azure-tui/internal/config"
)

func TestNewTerraformManager(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Set environment variable to use temp directory
	originalConfigDir := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalConfigDir)

	// Create config directory and file
	configDir := filepath.Join(tempDir, ".config", "azure-tui")
	err := os.MkdirAll(configDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create config dir: %v", err)
	}

	configContent := `
terraform:
  source_folder: "./test-terraform"
  default_location: "uksouth"
  auto_init: false
`

	configFile := filepath.Join(configDir, "config.yaml")
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Test creating TerraformManager
	tm, err := NewTerraformManager()
	if err != nil {
		t.Fatalf("Failed to create TerraformManager: %v", err)
	}

	if tm.WorkingDir != "./test-terraform" {
		t.Errorf("Expected working dir './test-terraform', got '%s'", tm.WorkingDir)
	}

	if tm.Config.DefaultLocation != "uksouth" {
		t.Errorf("Expected location 'uksouth', got '%s'", tm.Config.DefaultLocation)
	}
}

func TestTerraformManagerFileOperations(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	terraformDir := filepath.Join(tempDir, "terraform")

	// Create terraform directory
	err := os.MkdirAll(terraformDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create terraform dir: %v", err)
	}

	// Create TerraformManager with temp directory
	tm := &TerraformManager{
		WorkingDir: terraformDir,
		Config: &config.TerraformConfig{
			AutoFormat:     false, // Disable to avoid terraform fmt errors
			ValidateOnSave: false, // Disable to avoid terraform validate errors
		},
	}

	// Test CreateFile
	testContent := `resource "azurerm_resource_group" "test" {
  name     = "test-rg"
  location = "uksouth"
}`

	err = tm.CreateFile("test.tf", testContent)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Test ReadFile
	content, err := tm.ReadFile("test.tf")
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}

	if content != testContent {
		t.Errorf("File content mismatch. Expected:\n%s\nGot:\n%s", testContent, content)
	}

	// Test UpdateFile
	updatedContent := `resource "azurerm_resource_group" "test" {
  name     = "updated-test-rg"
  location = "uksouth"
}`

	err = tm.UpdateFile("test.tf", updatedContent)
	if err != nil {
		t.Fatalf("Failed to update file: %v", err)
	}

	// Verify update
	content, err = tm.ReadFile("test.tf")
	if err != nil {
		t.Fatalf("Failed to read updated file: %v", err)
	}

	if content != updatedContent {
		t.Errorf("Updated file content mismatch. Expected:\n%s\nGot:\n%s", updatedContent, content)
	}

	// Test ListFiles
	files, err := tm.ListFiles()
	if err != nil {
		t.Fatalf("Failed to list files: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 file, got %d", len(files))
	}

	if files[0].Name != "test.tf" {
		t.Errorf("Expected file name 'test.tf', got '%s'", files[0].Name)
	}

	// Test DeleteFile
	err = tm.DeleteFile("test.tf")
	if err != nil {
		t.Fatalf("Failed to delete file: %v", err)
	}

	// Verify deletion
	files, err = tm.ListFiles()
	if err != nil {
		t.Fatalf("Failed to list files after deletion: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected 0 files after deletion, got %d", len(files))
	}
}

func TestCheckTerraformInstalled(t *testing.T) {
	// This test checks if Terraform is available
	// It's informational and won't fail the test suite
	installed := CheckTerraformInstalled()
	t.Logf("Terraform installed: %v", installed)

	if installed {
		version, err := GetTerraformVersion()
		if err != nil {
			t.Logf("Failed to get Terraform version: %v", err)
		} else {
			t.Logf("Terraform version: %s", version)
		}
	}
}
