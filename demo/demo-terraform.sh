#!/bin/bash

# Demo script for Terraform integration in Azure TUI
# Shows how to test the Terraform functionality

set -e

echo "🏗️  Azure TUI - Terraform Integration Demo"
echo "========================================"
echo

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Check if azure-tui is built
if [ ! -f "azure-tui" ]; then
    echo -e "${YELLOW}Building azure-tui...${NC}"
    just build
    echo
fi

# Create test terraform projects for demonstration
echo -e "${BLUE}Setting up test Terraform projects...${NC}"

# Create demo terraform projects
mkdir -p demo-terraform/vm-project
mkdir -p demo-terraform/aks-project

# Create a simple VM terraform project
cat > demo-terraform/vm-project/main.tf << 'EOF'
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location
}

resource "azurerm_virtual_network" "main" {
  name                = "${var.prefix}-network"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
}

resource "azurerm_subnet" "internal" {
  name                 = "internal"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.2.0/24"]
}
EOF

cat > demo-terraform/vm-project/variables.tf << 'EOF'
variable "prefix" {
  description = "The prefix which should be used for all resources in this example"
  type        = string
  default     = "demo"
}

variable "location" {
  description = "The Azure Region in which all resources should be created"
  type        = string
  default     = "East US"
}

variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
  default     = "rg-demo-terraform"
}
EOF

cat > demo-terraform/vm-project/outputs.tf << 'EOF'
output "resource_group_name" {
  value = azurerm_resource_group.main.name
}

output "virtual_network_id" {
  value = azurerm_virtual_network.main.id
}
EOF

# Create a simple AKS terraform project
cat > demo-terraform/aks-project/main.tf << 'EOF'
terraform {
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~>3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location
}

resource "azurerm_kubernetes_cluster" "main" {
  name                = var.cluster_name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  dns_prefix          = var.dns_prefix

  default_node_pool {
    name       = "default"
    node_count = var.node_count
    vm_size    = var.vm_size
  }

  identity {
    type = "SystemAssigned"
  }
}
EOF

cat > demo-terraform/aks-project/variables.tf << 'EOF'
variable "cluster_name" {
  description = "The name of the AKS cluster"
  type        = string
  default     = "demo-aks"
}

variable "location" {
  description = "The Azure Region in which all resources should be created"
  type        = string
  default     = "East US"
}

variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
  default     = "rg-demo-aks"
}

variable "dns_prefix" {
  description = "DNS prefix for the cluster"
  type        = string
  default     = "demo-aks"
}

variable "node_count" {
  description = "Number of nodes in the cluster"
  type        = number
  default     = 2
}

variable "vm_size" {
  description = "VM size for the nodes"
  type        = string
  default     = "Standard_D2_v2"
}
EOF

echo -e "${GREEN}✅ Test Terraform projects created:${NC}"
echo "   📁 demo-terraform/vm-project/"
echo "   📁 demo-terraform/aks-project/"
echo

echo -e "${BLUE}🚀 Starting Azure TUI...${NC}"
echo
echo -e "${YELLOW}Terraform Integration Usage:${NC}"
echo "1. Press ${GREEN}Ctrl+T${NC} to open the Terraform Manager"
echo "2. Clean, minimal popup interface - no borders or backgrounds!"
echo "3. Use ↑/↓ to navigate menu options:"
echo "   • ${BLUE}Browse Folders${NC} - See available Terraform projects"
echo "   • ${BLUE}Create from Template${NC} - Create new projects from templates"
echo "   • ${BLUE}Analyze Code${NC} - Analyze Terraform code in a project"
echo "   • ${BLUE}Terraform Operations${NC} - Run terraform commands"
echo "   • ${BLUE}Open External Editor${NC} - Open project in your preferred editor"
echo "3. Press ${GREEN}Enter${NC} to select an option"
echo "4. Select a project folder when prompted"
echo "5. Press ${GREEN}Esc${NC} to close the popup"
echo
echo -e "${YELLOW}Test the following scenarios:${NC}"
echo "• Analyze the demo VM project"
echo "• Analyze the demo AKS project" 
echo "• Try opening a project in your editor"
echo "• Test Terraform operations (validate, format)"
echo
echo -e "${RED}Press any key to start Azure TUI...${NC}"
read -n 1 -s

# Start Azure TUI
./azure-tui

# Cleanup
echo
echo -e "${BLUE}Cleaning up demo files...${NC}"
rm -rf demo-terraform/

echo -e "${GREEN}✅ Demo complete!${NC}"
