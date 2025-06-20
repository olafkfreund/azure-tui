variable "resource_group_name" {
  description = "Name of the resource group"
  type        = string
  default     = "test-rg"
}

variable "location" {
  description = "Azure region for resources"
  type        = string
  default     = "East US"
}

variable "storage_account_name" {
  description = "Name of the storage account"
  type        = string
  default     = "teststorage12345"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "test"
}
