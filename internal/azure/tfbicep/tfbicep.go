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

// Enhanced Terraform Operations with better error handling and output capture

// TerraformInit initializes a Terraform working directory
func TerraformInit(dir string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "init", []string{})
}

// TerraformPlan creates an execution plan
func TerraformPlan(dir string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "plan", []string{})
}

// TerraformPlanWithOutput creates an execution plan and saves it to a file
func TerraformPlanWithOutput(dir, planFile string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "plan", []string{"-out", planFile})
}

// TerraformPlanJSON creates an execution plan in JSON format
func TerraformPlanJSON(dir string) (*TerraformPlanResult, error) {
	op, err := runTerraformCommand(dir, "plan", []string{"-json"})
	if err != nil {
		return nil, err
	}

	var plan TerraformPlanResult
	if err := json.Unmarshal([]byte(op.Output), &plan); err != nil {
		return nil, fmt.Errorf("failed to parse plan JSON: %w", err)
	}

	return &plan, nil
}

// TerraformApply applies the changes
func TerraformApply(dir string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "apply", []string{"-auto-approve"})
}

// TerraformApplyPlan applies a specific plan file
func TerraformApplyPlan(dir, planFile string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "apply", []string{planFile})
}

// TerraformDestroy destroys the infrastructure
func TerraformDestroy(dir string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "destroy", []string{"-auto-approve"})
}

// TerraformValidate validates the configuration
func TerraformValidate(dir string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "validate", []string{})
}

// TerraformFormat formats the configuration files
func TerraformFormat(dir string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "fmt", []string{"-recursive"})
}

// TerraformShow shows the current state or plan
func TerraformShow(dir string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "show", []string{})
}

// TerraformState performs state operations
func TerraformState(dir, subcommand string, args []string) (*TerraformOperation, error) {
	allArgs := append([]string{subcommand}, args...)
	return runTerraformCommand(dir, "state", allArgs)
}

// TerraformOutput gets output values
func TerraformOutput(dir string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "output", []string{"-json"})
}

// TerraformRefresh refreshes the state
func TerraformRefresh(dir string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "refresh", []string{})
}

// TerraformImport imports existing resources
func TerraformImport(dir, address, id string) (*TerraformOperation, error) {
	return runTerraformCommand(dir, "import", []string{address, id})
}

// TerraformWorkspace manages workspaces
func TerraformWorkspace(dir, subcommand string, args []string) (*TerraformOperation, error) {
	allArgs := append([]string{subcommand}, args...)
	return runTerraformCommand(dir, "workspace", allArgs)
}

// runTerraformCommand executes a Terraform command and captures detailed output
func runTerraformCommand(dir, command string, args []string) (*TerraformOperation, error) {
	start := time.Now()

	cmdArgs := append([]string{command}, args...)
	cmd := exec.Command("terraform", cmdArgs...)
	cmd.Dir = dir

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	duration := time.Since(start)

	operation := &TerraformOperation{
		Command:   fmt.Sprintf("terraform %s", strings.Join(cmdArgs, " ")),
		Directory: dir,
		Output:    stdout.String(),
		Error:     stderr.String(),
		Duration:  duration,
		Success:   err == nil,
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		operation.ExitCode = exitError.ExitCode()
	}

	return operation, err
}

// GetTerraformState reads and parses the current Terraform state
func GetTerraformState(dir string) (*TerraformStateInfo, error) {
	statePath := filepath.Join(dir, "terraform.tfstate")

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	var state TerraformStateInfo
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse state JSON: %w", err)
	}

	return &state, nil
}

// ValidateTerraformConfig validates a Terraform configuration
func ValidateTerraformConfig(dir string) (bool, []string, error) {
	op, err := TerraformValidate(dir)
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

// FormatTerraformFiles formats all Terraform files in a directory
func FormatTerraformFiles(dir string) error {
	op, err := TerraformFormat(dir)
	if err != nil {
		return fmt.Errorf("terraform format failed: %s", op.Error)
	}
	return nil
}

// CheckTerraformVersion checks if Terraform is installed and gets version
func CheckTerraformVersion() (string, error) {
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

// Bicep
func BicepBuild(file string) error {
	return exec.Command("bicep", "build", file).Run()
}

func BicepDeploy(file, group string) error {
	return exec.Command("az", "deployment", "group", "create", "--resource-group", group, "--template-file", file).Run()
}

// =============================================================================
// NETWORK RESOURCE TERRAFORM TEMPLATES
// =============================================================================

// GenerateVNetTerraformTemplate generates comprehensive VNet Terraform configuration
func GenerateVNetTerraformTemplate(vnetName, resourceGroup, location string, addressSpace []string, subnets []SubnetTemplate) string {
	template := `# Virtual Network Configuration
resource "azurerm_virtual_network" "%s" {
  name                = "%s"
  location            = "%s"
  resource_group_name = "%s"
  address_space       = [%s]

  tags = {
    Environment = "Production"
    ManagedBy   = "Terraform"
    CreatedBy   = "Azure-TUI"
  }
}

%s
`
	// Generate subnet resources
	subnetConfigs := ""
	for _, subnet := range subnets {
		subnetConfigs += fmt.Sprintf(`
resource "azurerm_subnet" "%s" {
  name                 = "%s"
  resource_group_name  = azurerm_virtual_network.%s.resource_group_name
  virtual_network_name = azurerm_virtual_network.%s.name
  address_prefixes     = ["%s"]
}
`,
			strings.ReplaceAll(subnet.Name, "-", "_"),
			subnet.Name,
			strings.ReplaceAll(vnetName, "-", "_"),
			strings.ReplaceAll(vnetName, "-", "_"),
			subnet.AddressPrefix)
	}

	addressSpaceStr := `"` + strings.Join(addressSpace, `", "`) + `"`
	return fmt.Sprintf(template,
		strings.ReplaceAll(vnetName, "-", "_"),
		vnetName,
		location,
		resourceGroup,
		addressSpaceStr,
		subnetConfigs)
}

// GenerateNSGTerraformTemplate generates NSG with security rules
func GenerateNSGTerraformTemplate(nsgName, resourceGroup, location string, rules []SecurityRuleTemplate) string {
	template := `# Network Security Group Configuration
resource "azurerm_network_security_group" "%s" {
  name                = "%s"
  location            = "%s"
  resource_group_name = "%s"

%s

  tags = {
    Environment = "Production"
    ManagedBy   = "Terraform"
    CreatedBy   = "Azure-TUI"
  }
}
`
	// Generate security rules
	ruleConfigs := ""
	for _, rule := range rules {
		ruleConfigs += fmt.Sprintf(`  security_rule {
    name                       = "%s"
    priority                   = %d
    direction                  = "%s"
    access                     = "%s"
    protocol                   = "%s"
    source_port_range          = "%s"
    destination_port_range     = "%s"
    source_address_prefix      = "%s"
    destination_address_prefix = "%s"
  }

`, rule.Name, rule.Priority, rule.Direction, rule.Access, rule.Protocol,
			rule.SourcePortRange, rule.DestinationPortRange, rule.SourceAddressPrefix, rule.DestinationAddressPrefix)
	}

	return fmt.Sprintf(template,
		strings.ReplaceAll(nsgName, "-", "_"),
		nsgName,
		location,
		resourceGroup,
		ruleConfigs)
}

// GenerateLoadBalancerTerraformTemplate generates load balancer configuration
func GenerateLoadBalancerTerraformTemplate(lbName, resourceGroup, location, sku string, publicIPName string) string {
	template := `# Load Balancer Configuration
resource "azurerm_public_ip" "%s_ip" {
  name                = "%s"
  location            = "%s"
  resource_group_name = "%s"
  allocation_method   = "Static"
  sku                 = "%s"

  tags = {
    Environment = "Production"
    ManagedBy   = "Terraform"
    CreatedBy   = "Azure-TUI"
  }
}

resource "azurerm_lb" "%s" {
  name                = "%s"
  location            = "%s"
  resource_group_name = "%s"
  sku                 = "%s"

  frontend_ip_configuration {
    name                 = "%s-frontend"
    public_ip_address_id = azurerm_public_ip.%s_ip.id
  }

  tags = {
    Environment = "Production"
    ManagedBy   = "Terraform"
    CreatedBy   = "Azure-TUI"
  }
}

resource "azurerm_lb_backend_address_pool" "%s_backend" {
  loadbalancer_id = azurerm_lb.%s.id
  name            = "%s-backend-pool"
}

resource "azurerm_lb_probe" "%s_probe" {
  loadbalancer_id = azurerm_lb.%s.id
  name            = "%s-health-probe"
  port            = 80
  protocol        = "Http"
  request_path    = "/"
}

resource "azurerm_lb_rule" "%s_rule" {
  loadbalancer_id                = azurerm_lb.%s.id
  name                          = "%s-rule"
  protocol                      = "Tcp"
  frontend_port                 = 80
  backend_port                  = 80
  frontend_ip_configuration_name = "%s-frontend"
  backend_address_pool_ids       = [azurerm_lb_backend_address_pool.%s_backend.id]
  probe_id                      = azurerm_lb_probe.%s_probe.id
}
`
	lbNameClean := strings.ReplaceAll(lbName, "-", "_")
	return fmt.Sprintf(template,
		lbNameClean, publicIPName, location, resourceGroup, sku,
		lbNameClean, lbName, location, resourceGroup, sku,
		lbName, lbNameClean,
		lbNameClean, lbNameClean, lbName,
		lbNameClean, lbNameClean, lbName,
		lbNameClean, lbNameClean, lbName, lbName, lbNameClean, lbNameClean)
}

// =============================================================================
// NETWORK RESOURCE BICEP TEMPLATES
// =============================================================================

// GenerateVNetBicepTemplate generates comprehensive VNet Bicep configuration
func GenerateVNetBicepTemplate(vnetName, location string, addressSpace []string, subnets []SubnetTemplate) string {
	template := `// Virtual Network Configuration
param location string = '%s'
param vnetName string = '%s'

resource virtualNetwork 'Microsoft.Network/virtualNetworks@2023-09-01' = {
  name: vnetName
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: [%s]
    }
    subnets: [%s]
  }
  tags: {
    Environment: 'Production'
    ManagedBy: 'Bicep'
    CreatedBy: 'Azure-TUI'
  }
}

output vnetId string = virtualNetwork.id
output vnetName string = virtualNetwork.name
%s
`
	// Generate address prefixes
	addressPrefixes := "'" + strings.Join(addressSpace, "', '") + "'"

	// Generate subnet configurations
	subnetConfigs := ""
	subnetOutputs := ""
	for i, subnet := range subnets {
		subnetConfigs += fmt.Sprintf(`
      {
        name: '%s'
        properties: {
          addressPrefix: '%s'
        }
      }`, subnet.Name, subnet.AddressPrefix)

		if i < len(subnets)-1 {
			subnetConfigs += "\n"
		}

		subnetOutputs += fmt.Sprintf("output %sSubnetId string = virtualNetwork.properties.subnets[%d].id\n",
			strings.ReplaceAll(subnet.Name, "-", "_"), i)
	}

	return fmt.Sprintf(template, location, vnetName, addressPrefixes, subnetConfigs, subnetOutputs)
}

// GenerateNSGBicepTemplate generates NSG Bicep configuration
func GenerateNSGBicepTemplate(nsgName, location string, rules []SecurityRuleTemplate) string {
	template := `// Network Security Group Configuration
param location string = '%s'
param nsgName string = '%s'

resource networkSecurityGroup 'Microsoft.Network/networkSecurityGroups@2023-09-01' = {
  name: nsgName
  location: location
  properties: {
    securityRules: [%s]
  }
  tags: {
    Environment: 'Production'
    ManagedBy: 'Bicep'
    CreatedBy: 'Azure-TUI'
  }
}

output nsgId string = networkSecurityGroup.id
output nsgName string = networkSecurityGroup.name
`
	// Generate security rules
	ruleConfigs := ""
	for i, rule := range rules {
		ruleConfigs += fmt.Sprintf(`
      {
        name: '%s'
        properties: {
          priority: %d
          direction: '%s'
          access: '%s'
          protocol: '%s'
          sourcePortRange: '%s'
          destinationPortRange: '%s'
          sourceAddressPrefix: '%s'
          destinationAddressPrefix: '%s'
        }
      }`, rule.Name, rule.Priority, rule.Direction, rule.Access, rule.Protocol,
			rule.SourcePortRange, rule.DestinationPortRange, rule.SourceAddressPrefix, rule.DestinationAddressPrefix)

		if i < len(rules)-1 {
			ruleConfigs += "\n"
		}
	}

	return fmt.Sprintf(template, location, nsgName, ruleConfigs)
}

// GenerateCompleteNetworkBicepTemplate generates a complete network infrastructure
func GenerateCompleteNetworkBicepTemplate(vnetName, nsgName, location string) string {
	return fmt.Sprintf(`// Complete Network Infrastructure
param location string = '%s'
param vnetName string = '%s'
param nsgName string = '%s'

// Network Security Group
resource networkSecurityGroup 'Microsoft.Network/networkSecurityGroups@2023-09-01' = {
  name: nsgName
  location: location
  properties: {
    securityRules: [
      {
        name: 'AllowHTTP'
        properties: {
          priority: 1000
          direction: 'Inbound'
          access: 'Allow'
          protocol: 'Tcp'
          sourcePortRange: '*'
          destinationPortRange: '80'
          sourceAddressPrefix: '*'
          destinationAddressPrefix: '*'
        }
      }
      {
        name: 'AllowHTTPS'
        properties: {
          priority: 1010
          direction: 'Inbound'
          access: 'Allow'
          protocol: 'Tcp'
          sourcePortRange: '*'
          destinationPortRange: '443'
          sourceAddressPrefix: '*'
          destinationAddressPrefix: '*'
        }
      }
      {
        name: 'AllowSSH'
        properties: {
          priority: 1020
          direction: 'Inbound'
          access: 'Allow'
          protocol: 'Tcp'
          sourcePortRange: '*'
          destinationPortRange: '22'
          sourceAddressPrefix: '*'
          destinationAddressPrefix: '*'
        }
      }
    ]
  }
  tags: {
    Environment: 'Production'
    ManagedBy: 'Bicep'
    CreatedBy: 'Azure-TUI'
  }
}

// Virtual Network
resource virtualNetwork 'Microsoft.Network/virtualNetworks@2023-09-01' = {
  name: vnetName
  location: location
  properties: {
    addressSpace: {
      addressPrefixes: ['10.0.0.0/16']
    }
    subnets: [
      {
        name: 'default'
        properties: {
          addressPrefix: '10.0.1.0/24'
          networkSecurityGroup: {
            id: networkSecurityGroup.id
          }
        }
      }
      {
        name: 'web-tier'
        properties: {
          addressPrefix: '10.0.2.0/24'
          networkSecurityGroup: {
            id: networkSecurityGroup.id
          }
        }
      }
      {
        name: 'app-tier'
        properties: {
          addressPrefix: '10.0.3.0/24'
          networkSecurityGroup: {
            id: networkSecurityGroup.id
          }
        }
      }
      {
        name: 'data-tier'
        properties: {
          addressPrefix: '10.0.4.0/24'
          networkSecurityGroup: {
            id: networkSecurityGroup.id
          }
        }
      }
    ]
  }
  tags: {
    Environment: 'Production'
    ManagedBy: 'Bicep'
    CreatedBy: 'Azure-TUI'
  }
}

// Outputs
output vnetId string = virtualNetwork.id
output nsgId string = networkSecurityGroup.id
output defaultSubnetId string = virtualNetwork.properties.subnets[0].id
output webTierSubnetId string = virtualNetwork.properties.subnets[1].id
output appTierSubnetId string = virtualNetwork.properties.subnets[2].id
output dataTierSubnetId string = virtualNetwork.properties.subnets[3].id
`, location, vnetName, nsgName)
}

// =============================================================================
// HELPER TYPES FOR TEMPLATE GENERATION
// =============================================================================

type SubnetTemplate struct {
	Name          string
	AddressPrefix string
	NSGName       string
}

type SecurityRuleTemplate struct {
	Name                     string
	Priority                 int
	Direction                string
	Access                   string
	Protocol                 string
	SourcePortRange          string
	DestinationPortRange     string
	SourceAddressPrefix      string
	DestinationAddressPrefix string
}
