package terraform

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/olafkfreund/azure-tui/internal/config"
)

// TemplateData holds data for template generation
type TemplateData struct {
	ProjectName     string
	Location        string
	Environment     string
	ResourceGroup   string
	Tags            map[string]string
	VMSize          string
	AdminUsername   string
	AKSNodeCount    int
	AKSNodeSize     string
	ContainerCPU    string
	ContainerMemory string
}

// ResourceTemplate represents a Terraform resource template
type ResourceTemplate struct {
	Name        string
	Description string
	Content     string
	Variables   []string
	Outputs     []string
}

// GetDefaultTemplateData returns default template data with user preferences
func GetDefaultTemplateData() (*TemplateData, error) {
	cfg, err := config.GetTerraformConfig()
	if err != nil {
		return nil, err
	}

	return &TemplateData{
		ProjectName:   "azure-tui",
		Location:      cfg.DefaultLocation,
		Environment:   "dev",
		ResourceGroup: "rg-azure-tui",
		Tags: map[string]string{
			"Environment": "dev",
			"Project":     "azure-tui",
			"ManagedBy":   "terraform",
			"CreatedBy":   "azure-tui-app",
		},
		VMSize:          "Standard_B1s",
		AdminUsername:   "azureuser",
		AKSNodeCount:    1,
		AKSNodeSize:     "Standard_B2s",
		ContainerCPU:    "0.5",
		ContainerMemory: "1.5",
	}, nil
}

// GenerateVMTemplate generates a VM template
func GenerateVMTemplate(data *TemplateData) (*ResourceTemplate, error) {
	tmpl := `# Azure Virtual Machine Configuration
resource "azurerm_resource_group" "vm_rg" {
  name     = "{{.ResourceGroup}}-vm"
  location = "{{.Location}}"
  tags     = {
{{- range $key, $value := .Tags}}
    {{$key}} = "{{$value}}"
{{- end}}
  }
}

resource "azurerm_virtual_network" "vm_vnet" {
  name                = "{{.ProjectName}}-vm-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.vm_rg.location
  resource_group_name = azurerm_resource_group.vm_rg.name
  tags                = azurerm_resource_group.vm_rg.tags
}

resource "azurerm_subnet" "vm_subnet" {
  name                 = "{{.ProjectName}}-vm-subnet"
  resource_group_name  = azurerm_resource_group.vm_rg.name
  virtual_network_name = azurerm_virtual_network.vm_vnet.name
  address_prefixes     = ["10.0.1.0/24"]
}

resource "azurerm_public_ip" "vm_public_ip" {
  name                = "{{.ProjectName}}-vm-public-ip"
  resource_group_name = azurerm_resource_group.vm_rg.name
  location            = azurerm_resource_group.vm_rg.location
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = azurerm_resource_group.vm_rg.tags
}

resource "azurerm_network_security_group" "vm_nsg" {
  name                = "{{.ProjectName}}-vm-nsg"
  location            = azurerm_resource_group.vm_rg.location
  resource_group_name = azurerm_resource_group.vm_rg.name
  tags                = azurerm_resource_group.vm_rg.tags

  security_rule {
    name                       = "SSH"
    priority                   = 1001
    direction                  = "Inbound"
    access                     = "Allow"
    protocol                   = "Tcp"
    source_port_range          = "*"
    destination_port_range     = "22"
    source_address_prefix      = "*"
    destination_address_prefix = "*"
  }
}

resource "azurerm_network_interface" "vm_nic" {
  name                = "{{.ProjectName}}-vm-nic"
  location            = azurerm_resource_group.vm_rg.location
  resource_group_name = azurerm_resource_group.vm_rg.name
  tags                = azurerm_resource_group.vm_rg.tags

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.vm_subnet.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.vm_public_ip.id
  }
}

resource "azurerm_linux_virtual_machine" "main" {
  name                = "{{.ProjectName}}-vm"
  resource_group_name = azurerm_resource_group.vm_rg.name
  location            = azurerm_resource_group.vm_rg.location
  size                = "{{.VMSize}}"
  admin_username      = "{{.AdminUsername}}"
  tags                = azurerm_resource_group.vm_rg.tags

  disable_password_authentication = true

  network_interface_ids = [
    azurerm_network_interface.vm_nic.id,
  ]

  admin_ssh_key {
    username   = "{{.AdminUsername}}"
    public_key = file("~/.ssh/id_rsa.pub")
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }
}`

	t, err := template.New("vm").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}

	return &ResourceTemplate{
		Name:        "Azure Virtual Machine",
		Description: "Linux VM with SSH access and basic networking",
		Content:     buf.String(),
		Variables:   []string{"vm_size", "admin_username", "location"},
		Outputs:     []string{"vm_public_ip", "vm_private_ip", "ssh_connection"},
	}, nil
}

// GenerateAKSTemplate generates an AKS cluster template
func GenerateAKSTemplate(data *TemplateData) (*ResourceTemplate, error) {
	tmpl := `# Azure Kubernetes Service Configuration
resource "azurerm_resource_group" "aks_rg" {
  name     = "{{.ResourceGroup}}-aks"
  location = "{{.Location}}"
  tags     = {
{{- range $key, $value := .Tags}}
    {{$key}} = "{{$value}}"
{{- end}}
  }
}

resource "azurerm_virtual_network" "aks_vnet" {
  name                = "{{.ProjectName}}-aks-vnet"
  address_space       = ["10.1.0.0/16"]
  location            = azurerm_resource_group.aks_rg.location
  resource_group_name = azurerm_resource_group.aks_rg.name
  tags                = azurerm_resource_group.aks_rg.tags
}

resource "azurerm_subnet" "aks_subnet" {
  name                 = "{{.ProjectName}}-aks-subnet"
  resource_group_name  = azurerm_resource_group.aks_rg.name
  virtual_network_name = azurerm_virtual_network.aks_vnet.name
  address_prefixes     = ["10.1.1.0/24"]
}

resource "azurerm_log_analytics_workspace" "aks_law" {
  name                = "{{.ProjectName}}-aks-law"
  location            = azurerm_resource_group.aks_rg.location
  resource_group_name = azurerm_resource_group.aks_rg.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
  tags                = azurerm_resource_group.aks_rg.tags
}

resource "azurerm_kubernetes_cluster" "main" {
  name                = "{{.ProjectName}}-aks"
  location            = azurerm_resource_group.aks_rg.location
  resource_group_name = azurerm_resource_group.aks_rg.name
  dns_prefix          = "{{.ProjectName}}-aks"
  kubernetes_version  = "1.27.7"
  tags                = azurerm_resource_group.aks_rg.tags

  default_node_pool {
    name           = "default"
    node_count     = {{.AKSNodeCount}}
    vm_size        = "{{.AKSNodeSize}}"
    vnet_subnet_id = azurerm_subnet.aks_subnet.id
    
    enable_auto_scaling = true
    min_count          = 1
    max_count          = 3
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin    = "azure"
    load_balancer_sku = "standard"
    outbound_type     = "loadBalancer"
  }

  oms_agent {
    log_analytics_workspace_id = azurerm_log_analytics_workspace.aks_law.id
  }

  azure_policy_enabled = true
  http_application_routing_enabled = true
}`

	t, err := template.New("aks").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}

	return &ResourceTemplate{
		Name:        "Azure Kubernetes Service",
		Description: "Small AKS cluster with monitoring and auto-scaling",
		Content:     buf.String(),
		Variables:   []string{"aks_node_count", "aks_node_size", "kubernetes_version"},
		Outputs:     []string{"aks_cluster_name", "aks_cluster_fqdn", "kubeconfig_command"},
	}, nil
}

// GenerateContainerInstancesTemplate generates container instances template
func GenerateContainerInstancesTemplate(data *TemplateData) (*ResourceTemplate, error) {
	tmpl := `# Azure Container Instances Configuration
resource "azurerm_resource_group" "aci_rg" {
  name     = "{{.ResourceGroup}}-aci"
  location = "{{.Location}}"
  tags     = {
{{- range $key, $value := .Tags}}
    {{$key}} = "{{$value}}"
{{- end}}
  }
}

resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

resource "azurerm_container_group" "helloworld_1" {
  name                = "{{.ProjectName}}-hello1"
  location            = azurerm_resource_group.aci_rg.location
  resource_group_name = azurerm_resource_group.aci_rg.name
  ip_address_type     = "Public"
  dns_name_label      = "{{.ProjectName}}-hello1-${random_string.suffix.result}"
  os_type             = "Linux"
  tags                = azurerm_resource_group.aci_rg.tags

  container {
    name   = "hello-world-1"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    cpu    = "{{.ContainerCPU}}"
    memory = "{{.ContainerMemory}}"

    ports {
      port     = 80
      protocol = "TCP"
    }

    environment_variables = {
      "TITLE" = "Azure TUI Demo - Container 1"
    }
  }
}

resource "azurerm_container_group" "helloworld_2" {
  name                = "{{.ProjectName}}-hello2"
  location            = azurerm_resource_group.aci_rg.location
  resource_group_name = azurerm_resource_group.aci_rg.name
  ip_address_type     = "Public"
  dns_name_label      = "{{.ProjectName}}-hello2-${random_string.suffix.result}"
  os_type             = "Linux"
  tags                = azurerm_resource_group.aci_rg.tags

  container {
    name   = "hello-world-2"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    cpu    = "{{.ContainerCPU}}"
    memory = "{{.ContainerMemory}}"

    ports {
      port     = 80
      protocol = "TCP"
    }

    environment_variables = {
      "TITLE" = "Azure TUI Demo - Container 2"
    }
  }
}`

	t, err := template.New("aci").Parse(tmpl)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return nil, err
	}

	return &ResourceTemplate{
		Name:        "Azure Container Instances",
		Description: "Two hello-world container instances with public IPs",
		Content:     buf.String(),
		Variables:   []string{"container_cpu", "container_memory"},
		Outputs:     []string{"container_urls", "container_ips"},
	}, nil
}

// GetAvailableTemplates returns a list of available resource templates
func GetAvailableTemplates() []ResourceTemplate {
	return []ResourceTemplate{
		{
			Name:        "Azure Virtual Machine",
			Description: "Linux VM with SSH access and basic networking",
			Variables:   []string{"vm_size", "admin_username", "location"},
			Outputs:     []string{"vm_public_ip", "vm_private_ip", "ssh_connection"},
		},
		{
			Name:        "Azure Kubernetes Service",
			Description: "Small AKS cluster with monitoring and auto-scaling",
			Variables:   []string{"aks_node_count", "aks_node_size", "kubernetes_version"},
			Outputs:     []string{"aks_cluster_name", "aks_cluster_fqdn", "kubeconfig_command"},
		},
		{
			Name:        "Azure Container Instances",
			Description: "Two hello-world container instances with public IPs",
			Variables:   []string{"container_cpu", "container_memory"},
			Outputs:     []string{"container_urls", "container_ips"},
		},
		{
			Name:        "Storage Account",
			Description: "Azure Storage Account with blob containers",
			Variables:   []string{"account_tier", "account_replication_type"},
			Outputs:     []string{"storage_account_name", "primary_blob_endpoint"},
		},
		{
			Name:        "Key Vault",
			Description: "Azure Key Vault for secrets management",
			Variables:   []string{"sku_name", "enabled_for_disk_encryption"},
			Outputs:     []string{"key_vault_uri", "key_vault_name"},
		},
	}
}

// GenerateCompleteTemplate generates a complete Terraform configuration
func GenerateCompleteTemplate(data *TemplateData, includeVM, includeAKS, includeACI bool) (map[string]string, error) {
	files := make(map[string]string)

	// Always include provider and variables
	files["main.tf"] = fmt.Sprintf(`# Configure the Azure Provider
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~>3.1"
    }
  }
  required_version = ">= 1.0"
}

# Configure the Microsoft Azure Provider
provider "azurerm" {
  features {}
}

# Main resource group
resource "azurerm_resource_group" "main" {
  name     = "%s"
  location = "%s"
  tags = {
%s
  }
}`, data.ResourceGroup, data.Location, formatTags(data.Tags))

	// Generate variables.tf
	variables := generateVariablesFile(data)
	files["variables.tf"] = variables

	// Generate terraform.tfvars
	tfvars := generateTerraformTfvars(data)
	files["terraform.tfvars"] = tfvars

	if includeVM {
		vmTemplate, err := GenerateVMTemplate(data)
		if err != nil {
			return nil, err
		}
		files["vm.tf"] = vmTemplate.Content
	}

	if includeAKS {
		aksTemplate, err := GenerateAKSTemplate(data)
		if err != nil {
			return nil, err
		}
		files["aks.tf"] = aksTemplate.Content
	}

	if includeACI {
		aciTemplate, err := GenerateContainerInstancesTemplate(data)
		if err != nil {
			return nil, err
		}
		files["containers.tf"] = aciTemplate.Content
	}

	// Generate outputs.tf
	outputs := generateOutputsFile(includeVM, includeAKS, includeACI)
	files["outputs.tf"] = outputs

	return files, nil
}

// Helper functions
func formatTags(tags map[string]string) string {
	var lines []string
	for key, value := range tags {
		lines = append(lines, fmt.Sprintf("    %s = \"%s\"", key, value))
	}
	return strings.Join(lines, "\n")
}

func generateVariablesFile(data *TemplateData) string {
	return fmt.Sprintf(`# Variables for Azure TUI Terraform Configuration

variable "location" {
  description = "The Azure region where resources will be created"
  type        = string
  default     = "%s"
}

variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
  default     = "%s"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "%s"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "%s"
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
%s
  }
}`, data.Location, data.ResourceGroup, data.Environment, data.ProjectName, formatTags(data.Tags))
}

func generateTerraformTfvars(data *TemplateData) string {
	return fmt.Sprintf(`# Terraform variable values
location            = "%s"
resource_group_name = "%s"
environment         = "%s"
project_name        = "%s"

tags = {
%s
}`, data.Location, data.ResourceGroup, data.Environment, data.ProjectName, formatTags(data.Tags))
}

func generateOutputsFile(includeVM, includeAKS, includeACI bool) string {
	outputs := []string{
		`# Output values for Azure TUI Terraform

output "resource_group_name" {
  description = "Name of the resource group"
  value       = azurerm_resource_group.main.name
}

output "resource_group_location" {
  description = "Location of the resource group"
  value       = azurerm_resource_group.main.location
}`,
	}

	if includeVM {
		outputs = append(outputs, `
output "vm_public_ip" {
  description = "Public IP address of the Virtual Machine"
  value       = azurerm_public_ip.vm_public_ip.ip_address
}

output "vm_ssh_connection" {
  description = "SSH connection string for the VM"
  value       = "ssh ${var.admin_username}@${azurerm_public_ip.vm_public_ip.ip_address}"
}`)
	}

	if includeAKS {
		outputs = append(outputs, `
output "aks_cluster_name" {
  description = "Name of the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.name
}

output "aks_kubeconfig_command" {
  description = "Command to get kubeconfig for AKS cluster"
  value       = "az aks get-credentials --resource-group ${azurerm_resource_group.main.name} --name ${azurerm_kubernetes_cluster.main.name}"
}`)
	}

	if includeACI {
		outputs = append(outputs, `
output "container_instances" {
  description = "Container instance URLs"
  value = {
    helloworld_1_url = "http://${azurerm_container_group.helloworld_1.fqdn}"
    helloworld_2_url = "http://${azurerm_container_group.helloworld_2.fqdn}"
  }
}`)
	}

	return strings.Join(outputs, "\n")
}
