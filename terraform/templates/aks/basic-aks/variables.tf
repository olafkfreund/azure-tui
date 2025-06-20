variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "rg-aks-cluster"
}

variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "East US"
}

variable "cluster_name" {
  description = "Name of the AKS cluster"
  type        = string
  
  validation {
    condition     = can(regex("^[a-zA-Z0-9-]{1,63}$", var.cluster_name))
    error_message = "Cluster name must be 1-63 characters and contain only alphanumeric characters and hyphens."
  }
}

variable "kubernetes_version" {
  description = "Version of Kubernetes to use"
  type        = string
  default     = null # Uses latest supported version
}

variable "system_node_count" {
  description = "Number of nodes in the system node pool"
  type        = number
  default     = 3
  
  validation {
    condition     = var.system_node_count >= 1 && var.system_node_count <= 100
    error_message = "System node count must be between 1 and 100."
  }
}

variable "system_node_vm_size" {
  description = "VM size for system node pool"
  type        = string
  default     = "Standard_D2s_v3"
}

variable "enable_auto_scaling" {
  description = "Enable auto-scaling for node pools"
  type        = bool
  default     = true
}

variable "system_node_min_count" {
  description = "Minimum number of nodes in system pool (when auto-scaling enabled)"
  type        = number
  default     = 1
}

variable "system_node_max_count" {
  description = "Maximum number of nodes in system pool (when auto-scaling enabled)"
  type        = number
  default     = 5
}

variable "availability_zones" {
  description = "List of availability zones for node pools"
  type        = list(string)
  default     = ["1", "2", "3"]
}

variable "os_disk_size_gb" {
  description = "OS disk size in GB"
  type        = number
  default     = 30
  
  validation {
    condition     = var.os_disk_size_gb >= 30 && var.os_disk_size_gb <= 2048
    error_message = "OS disk size must be between 30 and 2048 GB."
  }
}

variable "os_disk_type" {
  description = "Type of OS disk"
  type        = string
  default     = "Managed"
  
  validation {
    condition     = contains(["Managed", "Ephemeral"], var.os_disk_type)
    error_message = "OS disk type must be either Managed or Ephemeral."
  }
}

variable "network_plugin" {
  description = "Network plugin for AKS (azure, kubenet)"
  type        = string
  default     = "azure"
  
  validation {
    condition     = contains(["azure", "kubenet"], var.network_plugin)
    error_message = "Network plugin must be either azure or kubenet."
  }
}

variable "network_policy" {
  description = "Network policy for AKS (azure, calico)"
  type        = string
  default     = "azure"
  
  validation {
    condition     = contains(["azure", "calico"], var.network_policy)
    error_message = "Network policy must be either azure or calico."
  }
}

variable "admin_group_object_ids" {
  description = "List of Azure AD group object IDs for cluster admin access"
  type        = list(string)
  default     = []
}

variable "azure_rbac_enabled" {
  description = "Enable Azure RBAC for Kubernetes authorization"
  type        = bool
  default     = true
}

variable "enable_monitoring" {
  description = "Enable Azure Monitor for containers"
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "Log retention days for Log Analytics workspace"
  type        = number
  default     = 30
  
  validation {
    condition     = var.log_retention_days >= 30 && var.log_retention_days <= 730
    error_message = "Log retention days must be between 30 and 730."
  }
}

variable "enable_application_gateway" {
  description = "Enable Application Gateway Ingress Controller"
  type        = bool
  default     = false
}

variable "enable_key_vault_secrets_provider" {
  description = "Enable Key Vault Secrets Provider"
  type        = bool
  default     = false
}

variable "automatic_channel_upgrade" {
  description = "Automatic upgrade channel (patch, rapid, node-image, stable)"
  type        = string
  default     = "stable"
  
  validation {
    condition     = contains(["patch", "rapid", "node-image", "stable", "none"], var.automatic_channel_upgrade)
    error_message = "Automatic channel upgrade must be patch, rapid, node-image, stable, or none."
  }
}

variable "maintenance_window" {
  description = "Maintenance window configuration"
  type = object({
    day   = string
    hours = list(number)
  })
  default = null
}

variable "additional_node_pools" {
  description = "Additional node pools configuration"
  type = list(object({
    name                = string
    vm_size             = string
    node_count          = number
    enable_auto_scaling = bool
    min_count           = number
    max_count           = number
    os_disk_size_gb     = number
    os_disk_type        = string
    os_type             = string
    node_taints         = list(string)
    node_labels         = map(string)
  }))
  default = []
}

variable "create_container_registry" {
  description = "Create Azure Container Registry"
  type        = bool
  default     = false
}

variable "container_registry_sku" {
  description = "SKU for Azure Container Registry"
  type        = string
  default     = "Basic"
  
  validation {
    condition     = contains(["Basic", "Standard", "Premium"], var.container_registry_sku)
    error_message = "Container registry SKU must be Basic, Standard, or Premium."
  }
}

variable "create_key_vault" {
  description = "Create Azure Key Vault for secrets"
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
