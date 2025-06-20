variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "rg-container-instances"
}

variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "East US"
}

variable "container_group_name" {
  description = "Name of the container group"
  type        = string
  
  validation {
    condition     = can(regex("^[a-z0-9-]{1,63}$", var.container_group_name))
    error_message = "Container group name must be 1-63 characters, lowercase letters, numbers, and hyphens only."
  }
}

variable "container_name" {
  description = "Name of the main container"
  type        = string
  default     = "app"
}

variable "container_image" {
  description = "Container image to run"
  type        = string
  default     = "nginx:latest"
}

variable "cpu_cores" {
  description = "Number of CPU cores for the container"
  type        = number
  default     = 1
  
  validation {
    condition     = var.cpu_cores >= 0.1 && var.cpu_cores <= 4
    error_message = "CPU cores must be between 0.1 and 4."
  }
}

variable "memory_gb" {
  description = "Memory in GB for the container"
  type        = number
  default     = 1.5
  
  validation {
    condition     = var.memory_gb >= 0.1 && var.memory_gb <= 28
    error_message = "Memory must be between 0.1 and 28 GB."
  }
}

variable "ip_address_type" {
  description = "IP address type for the container group"
  type        = string
  default     = "Public"
  
  validation {
    condition     = contains(["Public", "Private"], var.ip_address_type)
    error_message = "IP address type must be either Public or Private."
  }
}

variable "dns_name_label" {
  description = "DNS name label for the container group"
  type        = string
  default     = ""
  
  validation {
    condition     = var.dns_name_label == "" || can(regex("^[a-z0-9-]{1,63}$", var.dns_name_label))
    error_message = "DNS name label must be 1-63 characters, lowercase letters, numbers, and hyphens only."
  }
}

variable "os_type" {
  description = "Operating system type"
  type        = string
  default     = "Linux"
  
  validation {
    condition     = contains(["Linux", "Windows"], var.os_type)
    error_message = "OS type must be either Linux or Windows."
  }
}

variable "restart_policy" {
  description = "Restart policy for the container group"
  type        = string
  default     = "Always"
  
  validation {
    condition     = contains(["Always", "Never", "OnFailure"], var.restart_policy)
    error_message = "Restart policy must be Always, Never, or OnFailure."
  }
}

variable "ports" {
  description = "List of ports to expose"
  type = list(object({
    port     = number
    protocol = string
  }))
  default = [
    {
      port     = 80
      protocol = "TCP"
    }
  ]
}

variable "environment_variables" {
  description = "Environment variables for the container"
  type        = map(string)
  default     = {}
}

variable "secure_environment_variables" {
  description = "Secure environment variables for the container"
  type        = map(string)
  default     = {}
  sensitive   = true
}

variable "volume_mounts" {
  description = "Volume mounts for the container"
  type = list(object({
    name       = string
    mount_path = string
    read_only  = bool
  }))
  default = []
}

variable "volumes" {
  description = "Volumes for the container group"
  type = list(object({
    name                 = string
    storage_account_name = string
    storage_account_key  = string
    share_name          = string
    quota_gb            = number
  }))
  default = []
}

variable "additional_containers" {
  description = "Additional containers in the group"
  type = list(object({
    name                  = string
    image                 = string
    cpu                   = number
    memory                = number
    environment_variables = map(string)
    ports = list(object({
      port     = number
      protocol = string
    }))
    volume_mounts = list(object({
      name       = string
      mount_path = string
      read_only  = bool
    }))
  }))
  default = []
}

variable "liveness_probe" {
  description = "Liveness probe configuration"
  type = object({
    exec                     = list(string)
    initial_delay_seconds    = number
    period_seconds          = number
    failure_threshold       = number
    success_threshold       = number
    timeout_seconds         = number
    http_get = object({
      path   = string
      port   = number
      scheme = string
    })
  })
  default = null
}

variable "readiness_probe" {
  description = "Readiness probe configuration"
  type = object({
    exec                     = list(string)
    initial_delay_seconds    = number
    period_seconds          = number
    failure_threshold       = number
    success_threshold       = number
    timeout_seconds         = number
    http_get = object({
      path   = string
      port   = number
      scheme = string
    })
  })
  default = null
}

variable "image_registry_credentials" {
  description = "Image registry credentials"
  type = list(object({
    server   = string
    username = string
    password = string
  }))
  default   = []
  sensitive = true
}

variable "enable_log_analytics" {
  description = "Enable Log Analytics for container logging"
  type        = bool
  default     = false
}

variable "log_analytics_workspace_id" {
  description = "Log Analytics workspace ID (if not provided, a new one will be created)"
  type        = string
  default     = ""
}

variable "log_analytics_workspace_key" {
  description = "Log Analytics workspace key"
  type        = string
  default     = ""
  sensitive   = true
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

variable "enable_system_assigned_identity" {
  description = "Enable system-assigned managed identity"
  type        = bool
  default     = false
}

variable "create_storage_account" {
  description = "Create storage account for Azure File Share volumes"
  type        = bool
  default     = true
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
