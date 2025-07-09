package bicep

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewBicepManager(t *testing.T) {
	manager := NewBicepManager()
	
	if manager == nil {
		t.Fatal("NewBicepManager returned nil")
	}
	
	if manager.timeout != 120*time.Second {
		t.Errorf("Expected timeout 120s, got %v", manager.timeout)
	}
	
	if manager.bicepPath == "" {
		t.Error("Bicep path is empty")
	}
	
	if manager.tempDir != "/tmp/azure-tui-bicep" {
		t.Errorf("Expected temp dir '/tmp/azure-tui-bicep', got '%s'", manager.tempDir)
	}
}

func TestSetters(t *testing.T) {
	manager := NewBicepManager()
	
	// Test SetTimeout
	newTimeout := 60 * time.Second
	manager.SetTimeout(newTimeout)
	if manager.timeout != newTimeout {
		t.Errorf("Expected timeout %v, got %v", newTimeout, manager.timeout)
	}
	
	// Test SetBicepPath
	newPath := "/custom/bicep/path"
	manager.SetBicepPath(newPath)
	if manager.bicepPath != newPath {
		t.Errorf("Expected bicep path '%s', got '%s'", newPath, manager.bicepPath)
	}
	
	// Test SetTempDir
	newTempDir := "/custom/temp/dir"
	manager.SetTempDir(newTempDir)
	if manager.tempDir != newTempDir {
		t.Errorf("Expected temp dir '%s', got '%s'", newTempDir, manager.tempDir)
	}
}

func TestBicepTemplateStructures(t *testing.T) {
	// Test BicepTemplate
	template := &BicepTemplate{
		FilePath:   "/test/path.bicep",
		Content:    "param test string = 'value'",
		Parameters: make(map[string]Parameter),
		Variables:  make(map[string]interface{}),
		Resources:  []Resource{},
		Outputs:    make(map[string]Output),
	}
	
	if template.FilePath != "/test/path.bicep" {
		t.Errorf("Expected file path '/test/path.bicep', got '%s'", template.FilePath)
	}
	
	// Test Parameter
	param := Parameter{
		Type:         "string",
		DefaultValue: "test",
		Description:  "Test parameter",
	}
	
	if param.Type != "string" {
		t.Errorf("Expected parameter type 'string', got '%s'", param.Type)
	}
	
	// Test Resource
	resource := Resource{
		Type:       "Microsoft.Storage/storageAccounts",
		APIVersion: "2021-04-01",
		Name:       "testStorage",
		Location:   "eastus",
		Properties: make(map[string]interface{}),
		Tags:       make(map[string]string),
	}
	
	if resource.Type != "Microsoft.Storage/storageAccounts" {
		t.Errorf("Expected resource type 'Microsoft.Storage/storageAccounts', got '%s'", resource.Type)
	}
	
	// Test Output
	output := Output{
		Type:  "string",
		Value: "storageAccount.id",
	}
	
	if output.Type != "string" {
		t.Errorf("Expected output type 'string', got '%s'", output.Type)
	}
}

func TestParseBicepContent(t *testing.T) {
	manager := NewBicepManager()
	
	testCases := []struct {
		name           string
		content        string
		expectedParams int
		expectedVars   int
		expectedRes    int
		expectedOut    int
	}{
		{
			name: "Simple parameter",
			content: `param storageAccountName string = 'mystore'
param location string = resourceGroup().location`,
			expectedParams: 2,
			expectedVars:   0,
			expectedRes:    0,
			expectedOut:    0,
		},
		{
			name: "Variable definition",
			content: `var uniqueName = '${storageAccountName}${uniqueString(resourceGroup().id)}'
var location = 'eastus'`,
			expectedParams: 0,
			expectedVars:   2,
			expectedRes:    0,
			expectedOut:    0,
		},
		{
			name: "Resource definition",
			content: `resource storageAccount 'Microsoft.Storage/storageAccounts@2021-04-01' = {
  name: storageAccountName
}`,
			expectedParams: 0,
			expectedVars:   0,
			expectedRes:    1,
			expectedOut:    0,
		},
		{
			name: "Output definition",
			content: `output storageAccountId string = storageAccount.id
output storageAccountName string = storageAccount.name`,
			expectedParams: 0,
			expectedVars:   0,
			expectedRes:    0,
			expectedOut:    2,
		},
		{
			name: "Complete template",
			content: `param name string = 'test'
var uniqueName = '${name}unique'
resource storage 'Microsoft.Storage/storageAccounts@2021-04-01' = {}
output id string = storage.id`,
			expectedParams: 1,
			expectedVars:   1,
			expectedRes:    1,
			expectedOut:    1,
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			template := &BicepTemplate{
				Content:    tc.content,
				Parameters: make(map[string]Parameter),
				Variables:  make(map[string]interface{}),
				Resources:  []Resource{},
				Outputs:    make(map[string]Output),
			}
			
			err := manager.parseBicepContent(template)
			if err != nil {
				t.Errorf("Failed to parse Bicep content: %v", err)
			}
			
			if len(template.Parameters) != tc.expectedParams {
				t.Errorf("Expected %d parameters, got %d", tc.expectedParams, len(template.Parameters))
			}
			
			if len(template.Variables) != tc.expectedVars {
				t.Errorf("Expected %d variables, got %d", tc.expectedVars, len(template.Variables))
			}
			
			if len(template.Resources) != tc.expectedRes {
				t.Errorf("Expected %d resources, got %d", tc.expectedRes, len(template.Resources))
			}
			
			if len(template.Outputs) != tc.expectedOut {
				t.Errorf("Expected %d outputs, got %d", tc.expectedOut, len(template.Outputs))
			}
		})
	}
}

func TestGenerateStorageAccountTemplate(t *testing.T) {
	manager := NewBicepManager()
	
	template := manager.GenerateStorageAccountTemplate("teststorage", "eastus", "Standard_LRS")
	
	if template == nil {
		t.Fatal("Generated template is nil")
	}
	
	if len(template.Parameters) != 3 {
		t.Errorf("Expected 3 parameters, got %d", len(template.Parameters))
	}
	
	if len(template.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(template.Resources))
	}
	
	if len(template.Outputs) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(template.Outputs))
	}
	
	// Check parameter types
	if param, exists := template.Parameters["storageAccountName"]; !exists || param.Type != "string" {
		t.Error("Missing or incorrect storageAccountName parameter")
	}
	
	// Check resource type
	if len(template.Resources) > 0 {
		resource := template.Resources[0]
		if resource.Type != "Microsoft.Storage/storageAccounts" {
			t.Errorf("Expected resource type 'Microsoft.Storage/storageAccounts', got '%s'", resource.Type)
		}
	}
	
	// Check that content is generated
	if template.Content == "" {
		t.Error("Template content is empty")
	}
	
	// Check content contains expected elements
	if !strings.Contains(template.Content, "param storageAccountName") {
		t.Error("Template content missing storageAccountName parameter")
	}
	
	if !strings.Contains(template.Content, "Microsoft.Storage/storageAccounts") {
		t.Error("Template content missing storage account resource")
	}
}

func TestGenerateVirtualMachineTemplate(t *testing.T) {
	manager := NewBicepManager()
	
	template := manager.GenerateVirtualMachineTemplate("testvm", "eastus", "Standard_B1s", "azureuser")
	
	if template == nil {
		t.Fatal("Generated template is nil")
	}
	
	if len(template.Parameters) != 5 {
		t.Errorf("Expected 5 parameters, got %d", len(template.Parameters))
	}
	
	if len(template.Resources) != 5 {
		t.Errorf("Expected 5 resources, got %d", len(template.Resources))
	}
	
	if len(template.Outputs) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(template.Outputs))
	}
	
	// Check that all required VM resources are present
	expectedResources := []string{
		"Microsoft.Network/publicIPAddresses",
		"Microsoft.Network/networkSecurityGroups",
		"Microsoft.Network/virtualNetworks",
		"Microsoft.Network/networkInterfaces",
		"Microsoft.Compute/virtualMachines",
	}
	
	for _, expectedType := range expectedResources {
		found := false
		for _, resource := range template.Resources {
			if resource.Type == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing resource type: %s", expectedType)
		}
	}
}

func TestGenerateAKSTemplate(t *testing.T) {
	manager := NewBicepManager()
	
	template := manager.GenerateAKSTemplate("testaks", "eastus", 3)
	
	if template == nil {
		t.Fatal("Generated template is nil")
	}
	
	if len(template.Parameters) != 4 {
		t.Errorf("Expected 4 parameters, got %d", len(template.Parameters))
	}
	
	if len(template.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(template.Resources))
	}
	
	if len(template.Outputs) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(template.Outputs))
	}
	
	// Check AKS resource
	if len(template.Resources) > 0 {
		resource := template.Resources[0]
		if resource.Type != "Microsoft.ContainerService/managedClusters" {
			t.Errorf("Expected AKS resource type, got '%s'", resource.Type)
		}
	}
}

func TestGenerateKeyVaultTemplate(t *testing.T) {
	manager := NewBicepManager()
	
	template := manager.GenerateKeyVaultTemplate("testkv", "eastus", "test-tenant-id")
	
	if template == nil {
		t.Fatal("Generated template is nil")
	}
	
	if len(template.Parameters) != 4 {
		t.Errorf("Expected 4 parameters, got %d", len(template.Parameters))
	}
	
	if len(template.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(template.Resources))
	}
	
	if len(template.Outputs) != 2 {
		t.Errorf("Expected 2 outputs, got %d", len(template.Outputs))
	}
	
	// Check Key Vault resource
	if len(template.Resources) > 0 {
		resource := template.Resources[0]
		if resource.Type != "Microsoft.KeyVault/vaults" {
			t.Errorf("Expected Key Vault resource type, got '%s'", resource.Type)
		}
	}
}

func TestGenerateCompleteInfrastructureTemplate(t *testing.T) {
	manager := NewBicepManager()
	
	template := manager.GenerateCompleteInfrastructureTemplate("testproject", "eastus")
	
	if template == nil {
		t.Fatal("Generated template is nil")
	}
	
	if len(template.Parameters) != 4 {
		t.Errorf("Expected 4 parameters, got %d", len(template.Parameters))
	}
	
	if len(template.Resources) != 3 {
		t.Errorf("Expected 3 resources, got %d", len(template.Resources))
	}
	
	if len(template.Outputs) != 3 {
		t.Errorf("Expected 3 outputs, got %d", len(template.Outputs))
	}
	
	// Check that all infrastructure components are present
	expectedResources := []string{
		"Microsoft.Storage/storageAccounts",
		"Microsoft.KeyVault/vaults",
		"Microsoft.ContainerService/managedClusters",
	}
	
	for _, expectedType := range expectedResources {
		found := false
		for _, resource := range template.Resources {
			if resource.Type == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Missing resource type: %s", expectedType)
		}
	}
}

func TestSaveBicepTemplate(t *testing.T) {
	manager := NewBicepManager()
	
	// Create a temporary directory
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bicep")
	
	template := &BicepTemplate{
		Content:  "param test string = 'value'\nresource test 'Microsoft.Storage/storageAccounts@2021-04-01' = {}",
		FilePath: testFile,
	}
	
	err := manager.SaveBicepTemplate(template, testFile)
	if err != nil {
		t.Errorf("Failed to save Bicep template: %v", err)
	}
	
	// Verify file was created
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Error("Bicep template file was not created")
	}
	
	// Verify file content
	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Errorf("Failed to read saved file: %v", err)
	}
	
	if string(content) != template.Content {
		t.Error("Saved file content does not match template content")
	}
}

func TestParseBicepFile(t *testing.T) {
	manager := NewBicepManager()
	
	// Create a temporary Bicep file
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.bicep")
	
	testContent := `// Test Bicep file
param storageAccountName string = 'mystore'
param location string = resourceGroup().location

var uniqueName = '${storageAccountName}${uniqueString(resourceGroup().id)}'

resource storageAccount 'Microsoft.Storage/storageAccounts@2021-04-01' = {
  name: uniqueName
  location: location
}

output storageAccountId string = storageAccount.id`
	
	err := os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	
	template, err := manager.ParseBicepFile(testFile)
	if err != nil {
		t.Errorf("Failed to parse Bicep file: %v", err)
	}
	
	if template == nil {
		t.Fatal("Parsed template is nil")
	}
	
	if template.FilePath != testFile {
		t.Errorf("Expected file path '%s', got '%s'", testFile, template.FilePath)
	}
	
	if len(template.Parameters) != 2 {
		t.Errorf("Expected 2 parameters, got %d", len(template.Parameters))
	}
	
	if len(template.Variables) != 1 {
		t.Errorf("Expected 1 variable, got %d", len(template.Variables))
	}
	
	if len(template.Resources) != 1 {
		t.Errorf("Expected 1 resource, got %d", len(template.Resources))
	}
	
	if len(template.Outputs) != 1 {
		t.Errorf("Expected 1 output, got %d", len(template.Outputs))
	}
}

func TestGenerateBicepContent(t *testing.T) {
	manager := NewBicepManager()
	
	template := &BicepTemplate{
		Parameters: map[string]Parameter{
			"name": {
				Type:        "string",
				DefaultValue: "test",
				Description: "Test parameter",
			},
		},
		Variables: map[string]interface{}{
			"uniqueName": "${name}unique",
		},
		Resources: []Resource{
			{
				Type:       "Microsoft.Storage/storageAccounts",
				APIVersion: "2021-04-01",
				Name:       "storage",
				Location:   "location",
				Properties: map[string]interface{}{
					"supportsHttpsTrafficOnly": true,
				},
				Tags: map[string]string{
					"Environment": "Test",
				},
			},
		},
		Outputs: map[string]Output{
			"id": {
				Type:  "string",
				Value: "storage.id",
			},
		},
		Metadata: Metadata{
			Description: "Test template",
			Author:      "Test",
			Created:     time.Now(),
		},
	}
	
	content := manager.generateBicepContent(template)
	
	if content == "" {
		t.Error("Generated content is empty")
	}
	
	// Check that content contains expected elements
	expectedElements := []string{
		"// Test template",
		"param name string = 'test'",
		"var uniqueName = '${name}unique'",
		"resource storage 'Microsoft.Storage/storageAccounts@2021-04-01' = {",
		"output id string = 'storage.id'",
	}
	
	for _, element := range expectedElements {
		if !strings.Contains(content, element) {
			t.Errorf("Generated content missing: %s", element)
		}
	}
}

// Benchmark tests
func BenchmarkGenerateStorageAccountTemplate(b *testing.B) {
	manager := NewBicepManager()
	
	for i := 0; i < b.N; i++ {
		template := manager.GenerateStorageAccountTemplate("test", "eastus", "Standard_LRS")
		_ = template
	}
}

func BenchmarkGenerateVMTemplate(b *testing.B) {
	manager := NewBicepManager()
	
	for i := 0; i < b.N; i++ {
		template := manager.GenerateVirtualMachineTemplate("testvm", "eastus", "Standard_B1s", "azureuser")
		_ = template
	}
}

func BenchmarkParseBicepContent(b *testing.B) {
	manager := NewBicepManager()
	
	testContent := `param name string = 'test'
var unique = '${name}unique'
resource storage 'Microsoft.Storage/storageAccounts@2021-04-01' = {}
output id string = storage.id`
	
	template := &BicepTemplate{
		Content:    testContent,
		Parameters: make(map[string]Parameter),
		Variables:  make(map[string]interface{}),
		Resources:  []Resource{},
		Outputs:    make(map[string]Output),
	}
	
	for i := 0; i < b.N; i++ {
		// Reset template for each iteration
		template.Parameters = make(map[string]Parameter)
		template.Variables = make(map[string]interface{})
		template.Resources = []Resource{}
		template.Outputs = make(map[string]Output)
		
		err := manager.parseBicepContent(template)
		if err != nil {
			b.Errorf("Parse failed: %v", err)
		}
	}
}