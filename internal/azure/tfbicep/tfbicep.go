package tfbicep

import (
	"fmt"
	"os/exec"
	"strings"
)

// Terraform
func TerraformInit(dir string) error {
	cmd := exec.Command("terraform", "init")
	cmd.Dir = dir
	return cmd.Run()
}

func TerraformPlan(dir string) error {
	cmd := exec.Command("terraform", "plan")
	cmd.Dir = dir
	return cmd.Run()
}

func TerraformApply(dir string) error {
	cmd := exec.Command("terraform", "apply", "-auto-approve")
	cmd.Dir = dir
	return cmd.Run()
}

func TerraformDestroy(dir string) error {
	cmd := exec.Command("terraform", "destroy", "-auto-approve")
	cmd.Dir = dir
	return cmd.Run()
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
