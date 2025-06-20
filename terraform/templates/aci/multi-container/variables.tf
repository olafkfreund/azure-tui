# Resource Group
variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "rg-aci-multi-container"
}

variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "East US"
}

# Container Group
variable "container_group_name" {
  description = "Name of the container group"
  type        = string
  default     = "aci-multi-container"
}

variable "ip_address_type" {
  description = "IP address type for the container group"
  type        = string
  default     = "Public"
  validation {
    condition     = contains(["Public", "Private"], var.ip_address_type)
    error_message = "IP address type must be either 'Public' or 'Private'."
  }
}

variable "dns_name_label" {
  description = "DNS name label for the container group"
  type        = string
  default     = "aci-multi-app"
}

variable "os_type" {
  description = "Operating system type"
  type        = string
  default     = "Linux"
  validation {
    condition     = contains(["Linux", "Windows"], var.os_type)
    error_message = "OS type must be either 'Linux' or 'Windows'."
  }
}

variable "restart_policy" {
  description = "Restart policy for the container group"
  type        = string
  default     = "Always"
  validation {
    condition     = contains(["Always", "Never", "OnFailure"], var.restart_policy)
    error_message = "Restart policy must be one of: 'Always', 'Never', 'OnFailure'."
  }
}

# Web Container
variable "web_image" {
  description = "Docker image for the web container"
  type        = string
  default     = "nginx:alpine"
}

variable "web_cpu" {
  description = "CPU allocation for the web container"
  type        = number
  default     = 0.5
  validation {
    condition     = var.web_cpu >= 0.1 && var.web_cpu <= 4.0
    error_message = "CPU allocation must be between 0.1 and 4.0."
  }
}

variable "web_memory" {
  description = "Memory allocation for the web container (in GB)"
  type        = number
  default     = 1.5
  validation {
    condition     = var.web_memory >= 0.1 && var.web_memory <= 8.0
    error_message = "Memory allocation must be between 0.1 and 8.0 GB."
  }
}

variable "web_environment_variables" {
  description = "Environment variables for the web container"
  type        = map(string)
  default     = {
    NGINX_PORT = "80"
  }
}

# API Container
variable "api_image" {
  description = "Docker image for the API container"
  type        = string
  default     = "httpd:alpine"
}

variable "api_cpu" {
  description = "CPU allocation for the API container"
  type        = number
  default     = 0.5
  validation {
    condition     = var.api_cpu >= 0.1 && var.api_cpu <= 4.0
    error_message = "CPU allocation must be between 0.1 and 4.0."
  }
}

variable "api_memory" {
  description = "Memory allocation for the API container (in GB)"
  type        = number
  default     = 1.5
  validation {
    condition     = var.api_memory >= 0.1 && var.api_memory <= 8.0
    error_message = "Memory allocation must be between 0.1 and 8.0 GB."
  }
}

variable "api_environment_variables" {
  description = "Environment variables for the API container"
  type        = map(string)
  default     = {
    API_PORT = "8080"
  }
}

# Diagnostics
variable "enable_diagnostics" {
  description = "Enable diagnostic logging"
  type        = bool
  default     = false
}

variable "log_retention_days" {
  description = "Log retention period in days"
  type        = number
  default     = 30
  validation {
    condition     = var.log_retention_days >= 30 && var.log_retention_days <= 730
    error_message = "Log retention must be between 30 and 730 days."
  }
}

# Tags
variable "tags" {
  description = "Tags to apply to all resources"
  type        = map(string)
  default = {
    Environment = "dev"
    Project     = "azure-tui"
    Component   = "aci-multi-container"
  }
}
