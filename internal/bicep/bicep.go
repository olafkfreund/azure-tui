package bicep

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// BicepTemplate represents a Bicep template structure
type BicepTemplate struct {
	FilePath        string                 `json:"filePath"`
	Content         string                 `json:"content"`
	Parameters      map[string]Parameter   `json:"parameters"`
	Variables       map[string]interface{} `json:"variables"`
	Resources       []Resource             `json:"resources"`
	Outputs         map[string]Output      `json:"outputs"`
	Metadata        Metadata               `json:"metadata"`
}

// Parameter represents a Bicep parameter
type Parameter struct {
	Type         string        `json:"type"`
	DefaultValue interface{}   `json:"defaultValue,omitempty"`
	AllowedValues []interface{} `json:"allowedValues,omitempty"`
	MinValue     *int          `json:"minValue,omitempty"`
	MaxValue     *int          `json:"maxValue,omitempty"`
	MinLength    *int          `json:"minLength,omitempty"`
	MaxLength    *int          `json:"maxLength,omitempty"`
	Description  string        `json:"description,omitempty"`
	Metadata     interface{}   `json:"metadata,omitempty"`
}

// Resource represents a Bicep resource
type Resource struct {
	Type       string                 `json:"type"`
	APIVersion string                 `json:"apiVersion"`
	Name       string                 `json:"name"`
	Location   string                 `json:"location,omitempty"`
	Properties map[string]interface{} `json:"properties,omitempty"`
	SKU        map[string]interface{} `json:"sku,omitempty"`
	Tags       map[string]string      `json:"tags,omitempty"`
	DependsOn  []string               `json:"dependsOn,omitempty"`
	Condition  string                 `json:"condition,omitempty"`
}

// Output represents a Bicep output
type Output struct {
	Type        string      `json:"type"`
	Value       interface{} `json:"value"`
	Description string      `json:"description,omitempty"`
}

// Metadata represents Bicep template metadata
type Metadata struct {
	Description string            `json:"description,omitempty"`
	Author      string            `json:"author,omitempty"`
	Version     string            `json:"version,omitempty"`
	Created     time.Time         `json:"created,omitempty"`
	Modified    time.Time         `json:"modified,omitempty"`
	Custom      map[string]string `json:"custom,omitempty"`
}

// BicepManager manages Bicep operations
type BicepManager struct {
	timeout     time.Duration
	bicepPath   string
	tempDir     string
}

// DeploymentResult represents the result of a Bicep deployment
type DeploymentResult struct {
	Success      bool              `json:"success"`
	DeploymentID string            `json:"deploymentId"`
	Message      string            `json:"message"`
	Output       map[string]string `json:"output"`
	Duration     time.Duration     `json:"duration"`
	Error        string            `json:"error,omitempty"`
}

// ValidationResult represents Bicep validation results
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Errors  []string `json:"errors"`
	Warnings []string `json:"warnings"`
}

// NewBicepManager creates a new Bicep manager
func NewBicepManager() *BicepManager {
	return &BicepManager{
		timeout:   120 * time.Second,
		bicepPath: findBicepExecutable(),
		tempDir:   "/tmp/azure-tui-bicep",
	}
}

// findBicepExecutable finds the Bicep CLI executable
func findBicepExecutable() string {
	// Try common locations
	paths := []string{
		"bicep",
		"/usr/local/bin/bicep",
		"/usr/bin/bicep",
		"/opt/bicep/bicep",
	}
	
	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path
		}
	}
	
	return "az bicep" // Fallback to Azure CLI bicep
}

// ensureTempDir ensures the temporary directory exists
func (bm *BicepManager) ensureTempDir() error {
	return os.MkdirAll(bm.tempDir, 0755)
}

// ParseBicepFile parses a Bicep file and extracts its structure
func (bm *BicepManager) ParseBicepFile(filePath string) (*BicepTemplate, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read Bicep file: %v", err)
	}
	
	template := &BicepTemplate{
		FilePath:   filePath,
		Content:    string(content),
		Parameters: make(map[string]Parameter),
		Variables:  make(map[string]interface{}),
		Resources:  []Resource{},
		Outputs:    make(map[string]Output),
	}
	
	// Parse the Bicep content
	if err := bm.parseBicepContent(template); err != nil {
		return nil, fmt.Errorf("failed to parse Bicep content: %v", err)
	}
	
	return template, nil
}

// parseBicepContent parses the Bicep file content to extract components
func (bm *BicepManager) parseBicepContent(template *BicepTemplate) error {
	lines := strings.Split(template.Content, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		
		// Parse parameters
		if strings.HasPrefix(line, "param ") {
			if err := bm.parseParameter(line, template); err != nil {
				return err
			}
		}
		
		// Parse variables
		if strings.HasPrefix(line, "var ") {
			if err := bm.parseVariable(line, template); err != nil {
				return err
			}
		}
		
		// Parse resources
		if strings.HasPrefix(line, "resource ") {
			if err := bm.parseResource(line, template); err != nil {
				return err
			}
		}
		
		// Parse outputs
		if strings.HasPrefix(line, "output ") {
			if err := bm.parseOutput(line, template); err != nil {
				return err
			}
		}
	}
	
	return nil
}

// parseParameter parses a Bicep parameter line
func (bm *BicepManager) parseParameter(line string, template *BicepTemplate) error {
	// Example: param storageAccountName string = 'mystorageaccount'
	// Example: param location string = resourceGroup().location
	re := regexp.MustCompile(`param\s+(\w+)\s+(\w+)\s*(?:=\s*(.+))?`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 3 {
		return nil // Skip malformed parameters
	}
	
	paramName := matches[1]
	paramType := matches[2]
	
	param := Parameter{
		Type: paramType,
	}
	
	if len(matches) > 3 && matches[3] != "" {
		defaultValue := strings.Trim(matches[3], "'\"")
		param.DefaultValue = defaultValue
	}
	
	template.Parameters[paramName] = param
	return nil
}

// parseVariable parses a Bicep variable line
func (bm *BicepManager) parseVariable(line string, template *BicepTemplate) error {
	// Example: var uniqueStorageName = '${storagePrefix}${uniqueString(resourceGroup().id)}'
	re := regexp.MustCompile(`var\s+(\w+)\s*=\s*(.+)`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 3 {
		return nil
	}
	
	varName := matches[1]
	varValue := strings.Trim(matches[2], "'\"")
	
	template.Variables[varName] = varValue
	return nil
}

// parseResource parses a Bicep resource line
func (bm *BicepManager) parseResource(line string, template *BicepTemplate) error {
	// Example: resource storageAccount 'Microsoft.Storage/storageAccounts@2021-04-01' = {
	re := regexp.MustCompile(`resource\s+(\w+)\s+'([^@]+)@([^']+)'`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 4 {
		return nil
	}
	
	resourceName := matches[1]
	resourceType := matches[2]
	apiVersion := matches[3]
	
	resource := Resource{
		Type:       resourceType,
		APIVersion: apiVersion,
		Name:       resourceName,
		Properties: make(map[string]interface{}),
		Tags:       make(map[string]string),
	}
	
	template.Resources = append(template.Resources, resource)
	return nil
}

// parseOutput parses a Bicep output line
func (bm *BicepManager) parseOutput(line string, template *BicepTemplate) error {
	// Example: output storageAccountId string = storageAccount.id
	re := regexp.MustCompile(`output\s+(\w+)\s+(\w+)\s*=\s*(.+)`)
	matches := re.FindStringSubmatch(line)
	
	if len(matches) < 4 {
		return nil
	}
	
	outputName := matches[1]
	outputType := matches[2]
	outputValue := strings.Trim(matches[3], "'\"")
	
	output := Output{
		Type:  outputType,
		Value: outputValue,
	}
	
	template.Outputs[outputName] = output
	return nil
}

// CompileBicep compiles a Bicep file to ARM template
func (bm *BicepManager) CompileBicep(ctx context.Context, bicepFilePath string) (string, error) {
	if err := bm.ensureTempDir(); err != nil {
		return "", err
	}
	
	outputPath := filepath.Join(bm.tempDir, "compiled.json")
	
	var cmd *exec.Cmd
	if strings.Contains(bm.bicepPath, "az bicep") {
		cmd = exec.CommandContext(ctx, "az", "bicep", "build", "--file", bicepFilePath, "--outfile", outputPath)
	} else {
		cmd = exec.CommandContext(ctx, bm.bicepPath, "build", bicepFilePath, "--outfile", outputPath)
	}
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("bicep compilation failed: %v, output: %s", err, string(output))
	}
	
	return outputPath, nil
}

// ValidateBicep validates a Bicep file
func (bm *BicepManager) ValidateBicep(ctx context.Context, bicepFilePath string) (*ValidationResult, error) {
	var cmd *exec.Cmd
	if strings.Contains(bm.bicepPath, "az bicep") {
		cmd = exec.CommandContext(ctx, "az", "bicep", "build", "--file", bicepFilePath, "--stdout")
	} else {
		cmd = exec.CommandContext(ctx, bm.bicepPath, "build", bicepFilePath, "--stdout")
	}
	
	output, err := cmd.CombinedOutput()
	result := &ValidationResult{
		Valid:    err == nil,
		Errors:   []string{},
		Warnings: []string{},
	}
	
	if err != nil {
		result.Errors = append(result.Errors, string(output))
	}
	
	return result, nil
}

// DeployBicep deploys a Bicep template to Azure
func (bm *BicepManager) DeployBicep(ctx context.Context, resourceGroup, deploymentName, bicepFilePath string, parameters map[string]string) (*DeploymentResult, error) {
	startTime := time.Now()
	
	result := &DeploymentResult{
		DeploymentID: deploymentName,
		Output:       make(map[string]string),
	}
	
	// Build the deployment command
	args := []string{
		"deployment", "group", "create",
		"--resource-group", resourceGroup,
		"--name", deploymentName,
		"--template-file", bicepFilePath,
	}
	
	// Add parameters
	if len(parameters) > 0 {
		paramStr := ""
		for key, value := range parameters {
			if paramStr != "" {
				paramStr += " "
			}
			paramStr += fmt.Sprintf("%s=%s", key, value)
		}
		args = append(args, "--parameters", paramStr)
	}
	
	cmd := exec.CommandContext(ctx, "az", args...)
	output, err := cmd.CombinedOutput()
	
	result.Duration = time.Since(startTime)
	
	if err != nil {
		result.Success = false
		result.Error = string(output)
		result.Message = fmt.Sprintf("Deployment failed: %v", err)
		return result, err
	}
	
	result.Success = true
	result.Message = "Deployment completed successfully"
	
	// Parse deployment output if it's JSON
	var deploymentOutput map[string]interface{}
	if err := json.Unmarshal(output, &deploymentOutput); err == nil {
		if outputs, ok := deploymentOutput["properties"].(map[string]interface{})["outputs"].(map[string]interface{}); ok {
			for key, value := range outputs {
				if valueMap, ok := value.(map[string]interface{}); ok {
					if val, ok := valueMap["value"].(string); ok {
						result.Output[key] = val
					}
				}
			}
		}
	}
	
	return result, nil
}

// GenerateBicepFromResource generates a Bicep template from an existing Azure resource
func (bm *BicepManager) GenerateBicepFromResource(ctx context.Context, resourceID string) (*BicepTemplate, error) {
	// Export the resource as ARM template first
	cmd := exec.CommandContext(ctx, "az", "resource", "show", "--ids", resourceID)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to get resource: %v", err)
	}
	
	var resourceData map[string]interface{}
	if err := json.Unmarshal(output, &resourceData); err != nil {
		return nil, fmt.Errorf("failed to parse resource data: %v", err)
	}
	
	// Convert ARM resource to Bicep
	template := &BicepTemplate{
		Parameters: make(map[string]Parameter),
		Variables:  make(map[string]interface{}),
		Resources:  []Resource{},
		Outputs:    make(map[string]Output),
		Metadata: Metadata{
			Description: "Generated from existing Azure resource",
			Author:      "Azure TUI",
			Created:     time.Now(),
		},
	}
	
	// Extract resource information
	if resourceType, ok := resourceData["type"].(string); ok {
		if apiVersion, ok := resourceData["apiVersion"].(string); ok {
			if name, ok := resourceData["name"].(string); ok {
				resource := Resource{
					Type:       resourceType,
					APIVersion: apiVersion,
					Name:       name,
					Properties: make(map[string]interface{}),
					Tags:       make(map[string]string),
				}
				
				// Extract location
				if location, ok := resourceData["location"].(string); ok {
					resource.Location = location
				}
				
				// Extract properties
				if properties, ok := resourceData["properties"].(map[string]interface{}); ok {
					resource.Properties = properties
				}
				
				// Extract tags
				if tags, ok := resourceData["tags"].(map[string]interface{}); ok {
					for key, value := range tags {
						if strValue, ok := value.(string); ok {
							resource.Tags[key] = strValue
						}
					}
				}
				
				template.Resources = append(template.Resources, resource)
			}
		}
	}
	
	// Generate Bicep content
	template.Content = bm.generateBicepContent(template)
	
	return template, nil
}

// generateBicepContent generates Bicep content from template structure
func (bm *BicepManager) generateBicepContent(template *BicepTemplate) string {
	var content strings.Builder
	
	// Add metadata
	if template.Metadata.Description != "" {
		content.WriteString(fmt.Sprintf("// %s\n", template.Metadata.Description))
		content.WriteString(fmt.Sprintf("// Generated by: %s\n", template.Metadata.Author))
		content.WriteString(fmt.Sprintf("// Created: %s\n\n", template.Metadata.Created.Format("2006-01-02 15:04:05")))
	}
	
	// Add parameters
	for name, param := range template.Parameters {
		line := fmt.Sprintf("param %s %s", name, param.Type)
		if param.DefaultValue != nil {
			if strValue, ok := param.DefaultValue.(string); ok {
				line += fmt.Sprintf(" = '%s'", strValue)
			} else {
				line += fmt.Sprintf(" = %v", param.DefaultValue)
			}
		}
		content.WriteString(line + "\n")
	}
	
	if len(template.Parameters) > 0 {
		content.WriteString("\n")
	}
	
	// Add variables
	for name, value := range template.Variables {
		if strValue, ok := value.(string); ok {
			content.WriteString(fmt.Sprintf("var %s = '%s'\n", name, strValue))
		} else {
			content.WriteString(fmt.Sprintf("var %s = %v\n", name, value))
		}
	}
	
	if len(template.Variables) > 0 {
		content.WriteString("\n")
	}
	
	// Add resources
	for _, resource := range template.Resources {
		content.WriteString(fmt.Sprintf("resource %s '%s@%s' = {\n", resource.Name, resource.Type, resource.APIVersion))
		content.WriteString(fmt.Sprintf("  name: '%s'\n", resource.Name))
		
		if resource.Location != "" {
			content.WriteString(fmt.Sprintf("  location: '%s'\n", resource.Location))
		}
		
		if len(resource.Properties) > 0 {
			content.WriteString("  properties: {\n")
			for key, value := range resource.Properties {
				if strValue, ok := value.(string); ok {
					content.WriteString(fmt.Sprintf("    %s: '%s'\n", key, strValue))
				} else {
					content.WriteString(fmt.Sprintf("    %s: %v\n", key, value))
				}
			}
			content.WriteString("  }\n")
		}
		
		if len(resource.Tags) > 0 {
			content.WriteString("  tags: {\n")
			for key, value := range resource.Tags {
				content.WriteString(fmt.Sprintf("    %s: '%s'\n", key, value))
			}
			content.WriteString("  }\n")
		}
		
		content.WriteString("}\n\n")
	}
	
	// Add outputs
	for name, output := range template.Outputs {
		if strValue, ok := output.Value.(string); ok {
			content.WriteString(fmt.Sprintf("output %s %s = '%s'\n", name, output.Type, strValue))
		} else {
			content.WriteString(fmt.Sprintf("output %s %s = %v\n", name, output.Type, output.Value))
		}
	}
	
	return content.String()
}

// SaveBicepTemplate saves a Bicep template to a file
func (bm *BicepManager) SaveBicepTemplate(template *BicepTemplate, filePath string) error {
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %v", err)
	}
	
	return os.WriteFile(filePath, []byte(template.Content), 0644)
}

// ListBicepFiles finds all Bicep files in a directory
func (bm *BicepManager) ListBicepFiles(directory string) ([]string, error) {
	var bicepFiles []string
	
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && strings.HasSuffix(path, ".bicep") {
			bicepFiles = append(bicepFiles, path)
		}
		
		return nil
	})
	
	return bicepFiles, err
}

// GetBicepVersion gets the version of Bicep CLI
func (bm *BicepManager) GetBicepVersion(ctx context.Context) (string, error) {
	var cmd *exec.Cmd
	if strings.Contains(bm.bicepPath, "az bicep") {
		cmd = exec.CommandContext(ctx, "az", "bicep", "version")
	} else {
		cmd = exec.CommandContext(ctx, bm.bicepPath, "--version")
	}
	
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get Bicep version: %v", err)
	}
	
	return strings.TrimSpace(string(output)), nil
}

// SetTimeout sets the operation timeout
func (bm *BicepManager) SetTimeout(timeout time.Duration) {
	bm.timeout = timeout
}

// SetBicepPath sets the path to Bicep executable
func (bm *BicepManager) SetBicepPath(path string) {
	bm.bicepPath = path
}

// SetTempDir sets the temporary directory for Bicep operations
func (bm *BicepManager) SetTempDir(dir string) {
	bm.tempDir = dir
}

// =============================================================================
// TEMPLATE GENERATION FUNCTIONS
// =============================================================================

// GenerateStorageAccountTemplate generates a Bicep template for a Storage Account
func (bm *BicepManager) GenerateStorageAccountTemplate(name, location, sku string) *BicepTemplate {
	template := &BicepTemplate{
		Parameters: map[string]Parameter{
			"storageAccountName": {
				Type:        "string",
				DefaultValue: name,
				Description: "Name of the storage account",
			},
			"location": {
				Type:        "string",
				DefaultValue: location,
				Description: "Location for the storage account",
			},
			"skuName": {
				Type:        "string",
				DefaultValue: sku,
				AllowedValues: []interface{}{"Standard_LRS", "Standard_GRS", "Standard_RAGRS", "Premium_LRS"},
				Description: "SKU for the storage account",
			},
		},
		Variables: map[string]interface{}{},
		Resources: []Resource{
			{
				Type:       "Microsoft.Storage/storageAccounts",
				APIVersion: "2021-04-01",
				Name:       "storageAccount",
				Location:   "location",
				Properties: map[string]interface{}{
					"supportsHttpsTrafficOnly": true,
					"allowBlobPublicAccess":     false,
					"minimumTlsVersion":         "TLS1_2",
				},
				Tags: map[string]string{
					"Environment": "Development",
					"CreatedBy":   "Azure TUI",
				},
			},
		},
		Outputs: map[string]Output{
			"storageAccountId": {
				Type:  "string",
				Value: "storageAccount.id",
			},
			"storageAccountName": {
				Type:  "string",
				Value: "storageAccount.name",
			},
		},
		Metadata: Metadata{
			Description: "Storage Account Bicep Template",
			Author:      "Azure TUI",
			Created:     time.Now(),
		},
	}
	
	template.Content = bm.generateBicepContent(template)
	return template
}

// GenerateVirtualMachineTemplate generates a Bicep template for a Virtual Machine
func (bm *BicepManager) GenerateVirtualMachineTemplate(vmName, location, vmSize, adminUsername string) *BicepTemplate {
	template := &BicepTemplate{
		Parameters: map[string]Parameter{
			"vmName": {
				Type:        "string",
				DefaultValue: vmName,
				Description: "Name of the virtual machine",
			},
			"location": {
				Type:        "string",
				DefaultValue: location,
				Description: "Location for the virtual machine",
			},
			"vmSize": {
				Type:        "string",
				DefaultValue: vmSize,
				Description: "Size of the virtual machine",
			},
			"adminUsername": {
				Type:        "string",
				DefaultValue: adminUsername,
				Description: "Admin username for the virtual machine",
			},
			"sshPublicKey": {
				Type:        "string",
				Description: "SSH public key for authentication",
			},
		},
		Variables: map[string]interface{}{
			"vnetName":        fmt.Sprintf("%s-vnet", vmName),
			"subnetName":      "default",
			"nicName":         fmt.Sprintf("%s-nic", vmName),
			"publicIPName":    fmt.Sprintf("%s-pip", vmName),
			"nsgName":         fmt.Sprintf("%s-nsg", vmName),
		},
		Resources: []Resource{
			{
				Type:       "Microsoft.Network/publicIPAddresses",
				APIVersion: "2021-02-01",
				Name:       "publicIP",
				Location:   "location",
				Properties: map[string]interface{}{
					"publicIPAllocationMethod": "Dynamic",
				},
			},
			{
				Type:       "Microsoft.Network/networkSecurityGroups",
				APIVersion: "2021-02-01",
				Name:       "networkSecurityGroup",
				Location:   "location",
				Properties: map[string]interface{}{
					"securityRules": []map[string]interface{}{
						{
							"name": "SSH",
							"properties": map[string]interface{}{
								"priority":                 1001,
								"access":                   "Allow",
								"direction":                "Inbound",
								"destinationPortRange":     "22",
								"protocol":                 "Tcp",
								"sourceAddressPrefix":      "*",
								"sourcePortRange":          "*",
								"destinationAddressPrefix": "*",
							},
						},
					},
				},
			},
			{
				Type:       "Microsoft.Network/virtualNetworks",
				APIVersion: "2021-02-01",
				Name:       "virtualNetwork",
				Location:   "location",
				Properties: map[string]interface{}{
					"addressSpace": map[string]interface{}{
						"addressPrefixes": []string{"10.0.0.0/16"},
					},
					"subnets": []map[string]interface{}{
						{
							"name": "default",
							"properties": map[string]interface{}{
								"addressPrefix": "10.0.0.0/24",
								"networkSecurityGroup": map[string]interface{}{
									"id": "networkSecurityGroup.id",
								},
							},
						},
					},
				},
			},
			{
				Type:       "Microsoft.Network/networkInterfaces",
				APIVersion: "2021-02-01",
				Name:       "networkInterface",
				Location:   "location",
				DependsOn:  []string{"publicIP", "virtualNetwork"},
				Properties: map[string]interface{}{
					"ipConfigurations": []map[string]interface{}{
						{
							"name": "internal",
							"properties": map[string]interface{}{
								"privateIPAllocationMethod": "Dynamic",
								"publicIPAddress": map[string]interface{}{
									"id": "publicIP.id",
								},
								"subnet": map[string]interface{}{
									"id": "virtualNetwork.properties.subnets[0].id",
								},
							},
						},
					},
				},
			},
			{
				Type:       "Microsoft.Compute/virtualMachines",
				APIVersion: "2021-03-01",
				Name:       "virtualMachine",
				Location:   "location",
				DependsOn:  []string{"networkInterface"},
				Properties: map[string]interface{}{
					"hardwareProfile": map[string]interface{}{
						"vmSize": "vmSize",
					},
					"osProfile": map[string]interface{}{
						"computerName":  "vmName",
						"adminUsername": "adminUsername",
						"linuxConfiguration": map[string]interface{}{
							"disablePasswordAuthentication": true,
							"ssh": map[string]interface{}{
								"publicKeys": []map[string]interface{}{
									{
										"path":    "/home/adminUsername/.ssh/authorized_keys",
										"keyData": "sshPublicKey",
									},
								},
							},
						},
					},
					"storageProfile": map[string]interface{}{
						"imageReference": map[string]interface{}{
							"publisher": "Canonical",
							"offer":     "0001-com-ubuntu-server-focal",
							"sku":       "20_04-lts-gen2",
							"version":   "latest",
						},
						"osDisk": map[string]interface{}{
							"createOption": "FromImage",
							"managedDisk": map[string]interface{}{
								"storageAccountType": "Standard_LRS",
							},
						},
					},
					"networkProfile": map[string]interface{}{
						"networkInterfaces": []map[string]interface{}{
							{
								"id": "networkInterface.id",
							},
						},
					},
				},
			},
		},
		Outputs: map[string]Output{
			"vmId": {
				Type:  "string",
				Value: "virtualMachine.id",
			},
			"publicIPAddress": {
				Type:  "string",
				Value: "publicIP.properties.ipAddress",
			},
		},
		Metadata: Metadata{
			Description: "Virtual Machine Bicep Template",
			Author:      "Azure TUI",
			Created:     time.Now(),
		},
	}
	
	template.Content = bm.generateBicepContent(template)
	return template
}

// GenerateAKSTemplate generates a Bicep template for an AKS cluster
func (bm *BicepManager) GenerateAKSTemplate(clusterName, location string, nodeCount int) *BicepTemplate {
	template := &BicepTemplate{
		Parameters: map[string]Parameter{
			"clusterName": {
				Type:        "string",
				DefaultValue: clusterName,
				Description: "Name of the AKS cluster",
			},
			"location": {
				Type:        "string",
				DefaultValue: location,
				Description: "Location for the AKS cluster",
			},
			"nodeCount": {
				Type:        "int",
				DefaultValue: nodeCount,
				MinValue:    &[]int{1}[0],
				MaxValue:    &[]int{100}[0],
				Description: "Number of nodes in the cluster",
			},
			"vmSize": {
				Type:        "string",
				DefaultValue: "Standard_DS2_v2",
				Description: "Size of the VMs for the node pool",
			},
		},
		Variables: map[string]interface{}{},
		Resources: []Resource{
			{
				Type:       "Microsoft.ContainerService/managedClusters",
				APIVersion: "2021-05-01",
				Name:       "aksCluster",
				Location:   "location",
				Properties: map[string]interface{}{
					"dnsPrefix": "clusterName",
					"agentPoolProfiles": []map[string]interface{}{
						{
							"name":         "nodepool1",
							"count":        "nodeCount",
							"vmSize":       "vmSize",
							"osType":       "Linux",
							"mode":         "System",
							"maxPods":      110,
							"type":         "VirtualMachineScaleSets",
							"osDiskSizeGB": 128,
						},
					},
					"servicePrincipalProfile": map[string]interface{}{
						"clientId": "msi",
					},
					"nodeResourceGroup": fmt.Sprintf("MC_%s_%s_%s", clusterName, clusterName, location),
				},
				Tags: map[string]string{
					"Environment": "Development",
					"CreatedBy":   "Azure TUI",
				},
			},
		},
		Outputs: map[string]Output{
			"clusterFQDN": {
				Type:  "string",
				Value: "aksCluster.properties.fqdn",
			},
			"clusterId": {
				Type:  "string",
				Value: "aksCluster.id",
			},
		},
		Metadata: Metadata{
			Description: "AKS Cluster Bicep Template",
			Author:      "Azure TUI",
			Created:     time.Now(),
		},
	}
	
	template.Content = bm.generateBicepContent(template)
	return template
}

// GenerateKeyVaultTemplate generates a Bicep template for a Key Vault
func (bm *BicepManager) GenerateKeyVaultTemplate(vaultName, location, tenantId string) *BicepTemplate {
	template := &BicepTemplate{
		Parameters: map[string]Parameter{
			"keyVaultName": {
				Type:        "string",
				DefaultValue: vaultName,
				Description: "Name of the Key Vault",
			},
			"location": {
				Type:        "string",
				DefaultValue: location,
				Description: "Location for the Key Vault",
			},
			"tenantId": {
				Type:        "string",
				DefaultValue: tenantId,
				Description: "Tenant ID for the Key Vault",
			},
			"objectId": {
				Type:        "string",
				Description: "Object ID of the user or service principal",
			},
		},
		Variables: map[string]interface{}{},
		Resources: []Resource{
			{
				Type:       "Microsoft.KeyVault/vaults",
				APIVersion: "2021-04-01-preview",
				Name:       "keyVault",
				Location:   "location",
				Properties: map[string]interface{}{
					"tenantId": "tenantId",
					"sku": map[string]interface{}{
						"family": "A",
						"name":   "standard",
					},
					"accessPolicies": []map[string]interface{}{
						{
							"tenantId": "tenantId",
							"objectId": "objectId",
							"permissions": map[string]interface{}{
								"keys":         []string{"get", "list", "update", "create", "import", "delete", "recover", "backup", "restore"},
								"secrets":      []string{"get", "list", "set", "delete", "recover", "backup", "restore"},
								"certificates": []string{"get", "list", "update", "create", "import", "delete", "recover", "backup", "restore", "managecontacts", "manageissuers", "getissuers", "listissuers", "setissuers", "deleteissuers"},
							},
						},
					},
					"enabledForDeployment":         true,
					"enabledForDiskEncryption":     true,
					"enabledForTemplateDeployment": true,
					"enableSoftDelete":             true,
					"softDeleteRetentionInDays":    90,
				},
				Tags: map[string]string{
					"Environment": "Development",
					"CreatedBy":   "Azure TUI",
				},
			},
		},
		Outputs: map[string]Output{
			"keyVaultId": {
				Type:  "string",
				Value: "keyVault.id",
			},
			"keyVaultUri": {
				Type:  "string",
				Value: "keyVault.properties.vaultUri",
			},
		},
		Metadata: Metadata{
			Description: "Key Vault Bicep Template",
			Author:      "Azure TUI",
			Created:     time.Now(),
		},
	}
	
	template.Content = bm.generateBicepContent(template)
	return template
}

// GenerateResourceGroupTemplate generates a Bicep template for a Resource Group
func (bm *BicepManager) GenerateResourceGroupTemplate(rgName, location string) *BicepTemplate {
	template := &BicepTemplate{
		Parameters: map[string]Parameter{
			"resourceGroupName": {
				Type:        "string",
				DefaultValue: rgName,
				Description: "Name of the resource group",
			},
			"location": {
				Type:        "string",
				DefaultValue: location,
				Description: "Location for the resource group",
			},
		},
		Variables: map[string]interface{}{},
		Resources: []Resource{
			{
				Type:       "Microsoft.Resources/resourceGroups",
				APIVersion: "2021-04-01",
				Name:       "resourceGroup",
				Location:   "location",
				Properties: map[string]interface{}{},
				Tags: map[string]string{
					"Environment": "Development",
					"CreatedBy":   "Azure TUI",
				},
			},
		},
		Outputs: map[string]Output{
			"resourceGroupId": {
				Type:  "string",
				Value: "resourceGroup.id",
			},
		},
		Metadata: Metadata{
			Description: "Resource Group Bicep Template",
			Author:      "Azure TUI",
			Created:     time.Now(),
		},
	}
	
	template.Content = bm.generateBicepContent(template)
	return template
}

// GenerateCompleteInfrastructureTemplate generates a comprehensive infrastructure template
func (bm *BicepManager) GenerateCompleteInfrastructureTemplate(projectName, location string) *BicepTemplate {
	template := &BicepTemplate{
		Parameters: map[string]Parameter{
			"projectName": {
				Type:        "string",
				DefaultValue: projectName,
				Description: "Name of the project (used as prefix)",
			},
			"location": {
				Type:        "string",
				DefaultValue: location,
				Description: "Location for all resources",
			},
			"adminUsername": {
				Type:        "string",
				DefaultValue: "azureuser",
				Description: "Admin username for VMs",
			},
			"sshPublicKey": {
				Type:        "string",
				Description: "SSH public key for VM authentication",
			},
		},
		Variables: map[string]interface{}{
			"storageAccountName": fmt.Sprintf("%ssa", projectName),
			"vmName":             fmt.Sprintf("%s-vm", projectName),
			"aksName":            fmt.Sprintf("%s-aks", projectName),
			"keyVaultName":       fmt.Sprintf("%s-kv", projectName),
		},
		Resources: []Resource{
			{
				Type:       "Microsoft.Storage/storageAccounts",
				APIVersion: "2021-04-01",
				Name:       "storageAccount",
				Location:   "location",
				Properties: map[string]interface{}{
					"supportsHttpsTrafficOnly": true,
					"allowBlobPublicAccess":     false,
					"minimumTlsVersion":         "TLS1_2",
				},
			},
			{
				Type:       "Microsoft.KeyVault/vaults",
				APIVersion: "2021-04-01-preview",
				Name:       "keyVault",
				Location:   "location",
				Properties: map[string]interface{}{
					"tenantId": "subscription().tenantId",
					"sku": map[string]interface{}{
						"family": "A",
						"name":   "standard",
					},
					"enabledForDeployment":         true,
					"enabledForDiskEncryption":     true,
					"enabledForTemplateDeployment": true,
					"enableSoftDelete":             true,
					"softDeleteRetentionInDays":    90,
				},
			},
			{
				Type:       "Microsoft.ContainerService/managedClusters",
				APIVersion: "2021-05-01",
				Name:       "aksCluster",
				Location:   "location",
				Properties: map[string]interface{}{
					"dnsPrefix": "projectName",
					"agentPoolProfiles": []map[string]interface{}{
						{
							"name":         "nodepool1",
							"count":        2,
							"vmSize":       "Standard_DS2_v2",
							"osType":       "Linux",
							"mode":         "System",
							"maxPods":      110,
							"type":         "VirtualMachineScaleSets",
							"osDiskSizeGB": 128,
						},
					},
					"servicePrincipalProfile": map[string]interface{}{
						"clientId": "msi",
					},
				},
			},
		},
		Outputs: map[string]Output{
			"storageAccountId": {
				Type:  "string",
				Value: "storageAccount.id",
			},
			"keyVaultUri": {
				Type:  "string",
				Value: "keyVault.properties.vaultUri",
			},
			"aksClusterFQDN": {
				Type:  "string",
				Value: "aksCluster.properties.fqdn",
			},
		},
		Metadata: Metadata{
			Description: "Complete Infrastructure Bicep Template",
			Author:      "Azure TUI",
			Created:     time.Now(),
		},
	}
	
	template.Content = bm.generateBicepContent(template)
	return template
}