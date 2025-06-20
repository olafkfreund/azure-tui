variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "rg-sql-server"
}

variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "East US"
}

variable "sql_server_name" {
  description = "Name of the SQL Server (must be globally unique)"
  type        = string
  
  validation {
    condition     = can(regex("^[a-z0-9-]{3,63}$", var.sql_server_name))
    error_message = "SQL Server name must be 3-63 characters, lowercase letters, numbers, and hyphens only."
  }
}

variable "sql_server_version" {
  description = "Version of SQL Server"
  type        = string
  default     = "12.0"
  
  validation {
    condition     = contains(["12.0"], var.sql_server_version)
    error_message = "SQL Server version must be 12.0."
  }
}

variable "sql_admin_username" {
  description = "SQL Server administrator username"
  type        = string
  default     = "sqladmin"
  
  validation {
    condition     = can(regex("^[a-zA-Z][a-zA-Z0-9]{2,127}$", var.sql_admin_username))
    error_message = "Admin username must start with a letter, be 3-128 characters, and contain only letters and numbers."
  }
}

variable "database_name" {
  description = "Name of the SQL Database"
  type        = string
  default     = "myapp-db"
  
  validation {
    condition     = can(regex("^[a-zA-Z0-9_-]{1,128}$", var.database_name))
    error_message = "Database name must be 1-128 characters and contain only letters, numbers, underscores, and hyphens."
  }
}

variable "database_sku" {
  description = "SKU for the SQL Database"
  type        = string
  default     = "S0"
  
  validation {
    condition = contains([
      "Basic", "S0", "S1", "S2", "S3", "S4", "S6", "S7", "S9", "S12",
      "P1", "P2", "P4", "P6", "P11", "P15",
      "GP_Gen5_2", "GP_Gen5_4", "GP_Gen5_8", "GP_Gen5_16", "GP_Gen5_32",
      "BC_Gen5_2", "BC_Gen5_4", "BC_Gen5_8", "BC_Gen5_16", "BC_Gen5_32"
    ], var.database_sku)
    error_message = "Database SKU must be a valid Azure SQL Database SKU."
  }
}

variable "database_collation" {
  description = "Collation for the SQL Database"
  type        = string
  default     = "SQL_Latin1_General_CP1_CI_AS"
}

variable "license_type" {
  description = "License type for the database"
  type        = string
  default     = "LicenseIncluded"
  
  validation {
    condition     = contains(["LicenseIncluded", "BasePrice"], var.license_type)
    error_message = "License type must be either LicenseIncluded or BasePrice."
  }
}

variable "max_size_gb" {
  description = "Maximum size of the database in GB"
  type        = number
  default     = 2
  
  validation {
    condition     = var.max_size_gb >= 1 && var.max_size_gb <= 4096
    error_message = "Database size must be between 1 and 4096 GB."
  }
}

variable "public_network_access_enabled" {
  description = "Whether public network access is enabled"
  type        = bool
  default     = true
}

variable "allow_azure_services" {
  description = "Allow Azure services to access the SQL server"
  type        = bool
  default     = true
}

variable "allowed_client_ips" {
  description = "List of client IP addresses allowed to access the SQL server"
  type        = list(string)
  default     = []
}

variable "create_virtual_network" {
  description = "Whether to create a virtual network for private access"
  type        = bool
  default     = false
}

variable "backup_retention_days" {
  description = "Short-term backup retention in days"
  type        = number
  default     = 7
  
  validation {
    condition     = var.backup_retention_days >= 1 && var.backup_retention_days <= 35
    error_message = "Backup retention days must be between 1 and 35."
  }
}

variable "weekly_backup_retention" {
  description = "Weekly backup retention (ISO 8601 format)"
  type        = string
  default     = "P1W"
}

variable "monthly_backup_retention" {
  description = "Monthly backup retention (ISO 8601 format)"
  type        = string
  default     = "P1M"
}

variable "yearly_backup_retention" {
  description = "Yearly backup retention (ISO 8601 format)"
  type        = string
  default     = "P1Y"
}

variable "enable_auditing" {
  description = "Enable SQL Server auditing"
  type        = bool
  default     = true
}

variable "audit_retention_days" {
  description = "Audit log retention in days"
  type        = number
  default     = 90
  
  validation {
    condition     = var.audit_retention_days >= 0 && var.audit_retention_days <= 3285
    error_message = "Audit retention days must be between 0 and 3285."
  }
}

variable "enable_threat_detection" {
  description = "Enable Microsoft Defender for SQL"
  type        = bool
  default     = true
}

variable "threat_detection_retention_days" {
  description = "Threat detection log retention in days"
  type        = number
  default     = 30
}

variable "email_admins_on_alert" {
  description = "Email admins when security alert is triggered"
  type        = bool
  default     = true
}

variable "alert_email_addresses" {
  description = "List of email addresses for security alerts"
  type        = list(string)
  default     = []
}

variable "azuread_admin_login" {
  description = "Azure AD admin login name"
  type        = string
  default     = ""
}

variable "azuread_admin_object_id" {
  description = "Azure AD admin object ID"
  type        = string
  default     = ""
}

variable "create_key_vault" {
  description = "Create Key Vault for storing secrets"
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
