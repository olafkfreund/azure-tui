# Variables for Azure TUI Terraform Configuration

variable "location" {
  description = "The Azure region where resources will be created"
  type        = string
  default     = "uksouth"
}

variable "resource_group_name" {
  description = "The name of the resource group"
  type        = string
  default     = "rg-azure-tui"
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "dev"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "azure-tui"
}

# VM Configuration Variables
variable "vm_size" {
  description = "Size of the Virtual Machine"
  type        = string
  default     = "Standard_B1s"
}

variable "admin_username" {
  description = "Admin username for the Virtual Machine"
  type        = string
  default     = "azureuser"
}

# AKS Configuration Variables
variable "aks_node_count" {
  description = "Number of nodes in the AKS cluster"
  type        = number
  default     = 1
}

variable "aks_node_size" {
  description = "Size of the AKS nodes"
  type        = string
  default     = "Standard_B2s"
}

variable "kubernetes_version" {
  description = "Kubernetes version for AKS cluster"
  type        = string
  default     = "1.27.7"
}

# Container Instance Variables
variable "container_cpu" {
  description = "CPU allocation for container instances"
  type        = string
  default     = "0.5"
}

variable "container_memory" {
  description = "Memory allocation for container instances"
  type        = string
  default     = "1.5"
}

# Networking Variables
variable "vnet_address_space" {
  description = "Address space for the virtual network"
  type        = list(string)
  default     = ["10.0.0.0/16"]
}

variable "subnet_address_prefix" {
  description = "Address prefix for the subnet"
  type        = string
  default     = "10.0.1.0/24"
}

variable "aks_subnet_address_prefix" {
  description = "Address prefix for the AKS subnet"
  type        = string
  default     = "10.0.2.0/24"
}

# Tags
variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    Environment = "dev"
    Project     = "azure-tui"
    ManagedBy   = "terraform"
    CreatedBy   = "azure-tui-app"
  }
}
