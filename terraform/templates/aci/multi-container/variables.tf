variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "rg-linux-vm"
}

variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "East US"
  
  validation {
    condition = contains([
      "East US", "East US 2", "West US", "West US 2", "West US 3",
      "Central US", "North Central US", "South Central US", "West Central US",
      "Canada Central", "Canada East", "Brazil South", "UK South", "UK West",
      "North Europe", "West Europe", "France Central", "Germany West Central",
      "Norway East", "Switzerland North", "UAE North", "South Africa North",
      "Australia East", "Australia Southeast", "Central India", "South India",
      "Japan East", "Japan West", "Korea Central", "Southeast Asia", "East Asia"
    ], var.location)
    error_message = "The location must be a valid Azure region."
  }
}

variable "vm_name" {
  description = "Name of the virtual machine"
  type        = string
  default     = "vm-linux-01"
  
  validation {
    condition     = can(regex("^[a-zA-Z0-9-]{1,64}$", var.vm_name))
    error_message = "VM name must be between 1-64 characters and contain only alphanumeric characters and hyphens."
  }
}

variable "vm_size" {
  description = "Size of the virtual machine"
  type        = string
  default     = "Standard_B2s"
  
  validation {
    condition = contains([
      "Standard_B1s", "Standard_B1ms", "Standard_B2s", "Standard_B2ms", "Standard_B4ms",
      "Standard_D2s_v3", "Standard_D4s_v3", "Standard_D8s_v3", "Standard_D16s_v3",
      "Standard_E2s_v3", "Standard_E4s_v3", "Standard_E8s_v3", "Standard_E16s_v3"
    ], var.vm_size)
    error_message = "VM size must be a valid Azure VM size."
  }
}

variable "admin_username" {
  description = "Admin username for the virtual machine"
  type        = string
  default     = "azureuser"
  
  validation {
    condition     = can(regex("^[a-z_][a-z0-9_-]*[$]?$", var.admin_username)) && length(var.admin_username) >= 1 && length(var.admin_username) <= 32
    error_message = "Admin username must be between 1-32 characters, start with a letter or underscore, and contain only lowercase letters, numbers, hyphens, and underscores."
  }
}

variable "os_disk_type" {
  description = "Type of OS disk (Standard_LRS, Premium_LRS, StandardSSD_LRS)"
  type        = string
  default     = "Standard_LRS"
  
  validation {
    condition     = contains(["Standard_LRS", "Premium_LRS", "StandardSSD_LRS", "UltraSSD_LRS"], var.os_disk_type)
    error_message = "OS disk type must be Standard_LRS, Premium_LRS, StandardSSD_LRS, or UltraSSD_LRS."
  }
}

variable "install_docker" {
  description = "Whether to install Docker on the VM"
  type        = bool
  default     = false
}

variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    Environment = "Development"
    Project     = "Azure-TUI"
    ManagedBy   = "Terraform"
    CreatedBy   = "Azure-TUI"
  }
}

variable "ssh_source_addresses" {
  description = "List of source IP addresses allowed for SSH access"
  type        = list(string)
  default     = ["*"]
  
  validation {
    condition     = length(var.ssh_source_addresses) > 0
    error_message = "At least one source address must be specified for SSH access."
  }
}
