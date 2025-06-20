# Configure the Azure Provider
terraform {
  required_version = ">= 1.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

# Generate random password for SQL Server admin
resource "random_password" "sql_admin_password" {
  length  = 16
  special = true
  upper   = true
  lower   = true
  numeric = true
}

# Create resource group
resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location

  tags = var.tags
}

# Create SQL Server
resource "azurerm_mssql_server" "main" {
  name                         = var.sql_server_name
  resource_group_name          = azurerm_resource_group.main.name
  location                     = azurerm_resource_group.main.location
  version                      = var.sql_server_version
  administrator_login          = var.sql_admin_username
  administrator_login_password = random_password.sql_admin_password.result
  minimum_tls_version          = "1.2"
  
  # Azure AD authentication
  azuread_administrator {
    login_username = var.azuread_admin_login
    object_id      = var.azuread_admin_object_id
    tenant_id      = data.azurerm_client_config.current.tenant_id
  }

  # Security features
  public_network_access_enabled = var.public_network_access_enabled
  
  # Advanced security
  identity {
    type = "SystemAssigned"
  }

  tags = var.tags
}

# Get current Azure client configuration
data "azurerm_client_config" "current" {}

# Create SQL Database
resource "azurerm_mssql_database" "main" {
  name           = var.database_name
  server_id      = azurerm_mssql_server.main.id
  collation      = var.database_collation
  license_type   = var.license_type
  max_size_gb    = var.max_size_gb
  sku_name       = var.database_sku

  # Backup and retention
  short_term_retention_policy {
    retention_days = var.backup_retention_days
  }

  long_term_retention_policy {
    weekly_retention  = var.weekly_backup_retention
    monthly_retention = var.monthly_backup_retention
    yearly_retention  = var.yearly_backup_retention
    week_of_year      = 1
  }

  # Security
  transparent_data_encryption_enabled = true

  tags = var.tags
}

# Create firewall rules
resource "azurerm_mssql_firewall_rule" "azure_services" {
  count            = var.allow_azure_services ? 1 : 0
  name             = "AllowAzureServices"
  server_id        = azurerm_mssql_server.main.id
  start_ip_address = "0.0.0.0"
  end_ip_address   = "0.0.0.0"
}

resource "azurerm_mssql_firewall_rule" "client_ips" {
  count            = length(var.allowed_client_ips)
  name             = "ClientIP-${count.index + 1}"
  server_id        = azurerm_mssql_server.main.id
  start_ip_address = var.allowed_client_ips[count.index]
  end_ip_address   = var.allowed_client_ips[count.index]
}

# Create virtual network if specified
resource "azurerm_virtual_network" "main" {
  count               = var.create_virtual_network ? 1 : 0
  name                = "${var.sql_server_name}-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = var.tags
}

resource "azurerm_subnet" "sql" {
  count                = var.create_virtual_network ? 1 : 0
  name                 = "${var.sql_server_name}-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main[0].name
  address_prefixes     = ["10.0.1.0/24"]

  service_endpoints = ["Microsoft.Sql"]
}

# Create virtual network rule
resource "azurerm_mssql_virtual_network_rule" "main" {
  count     = var.create_virtual_network ? 1 : 0
  name      = "${var.sql_server_name}-vnet-rule"
  server_id = azurerm_mssql_server.main.id
  subnet_id = azurerm_subnet.sql[0].id
}

# Enable auditing
resource "azurerm_mssql_server_extended_auditing_policy" "main" {
  count                       = var.enable_auditing ? 1 : 0
  server_id                   = azurerm_mssql_server.main.id
  storage_endpoint            = azurerm_storage_account.audit[0].primary_blob_endpoint
  storage_account_access_key  = azurerm_storage_account.audit[0].primary_access_key
  storage_account_access_key_is_secondary = false
  retention_in_days           = var.audit_retention_days
}

# Storage account for auditing
resource "azurerm_storage_account" "audit" {
  count                    = var.enable_auditing ? 1 : 0
  name                     = "${replace(var.sql_server_name, "-", "")}audit"
  resource_group_name      = azurerm_resource_group.main.name
  location                 = azurerm_resource_group.main.location
  account_tier             = "Standard"
  account_replication_type = "LRS"
  min_tls_version          = "TLS1_2"

  tags = var.tags
}

# Enable Microsoft Defender for SQL
resource "azurerm_mssql_server_security_alert_policy" "main" {
  count                      = var.enable_threat_detection ? 1 : 0
  resource_group_name        = azurerm_resource_group.main.name
  server_name                = azurerm_mssql_server.main.name
  state                      = "Enabled"
  storage_endpoint           = azurerm_storage_account.audit[0].primary_blob_endpoint
  storage_account_access_key = azurerm_storage_account.audit[0].primary_access_key
  
  disabled_alerts = []
  retention_days  = var.threat_detection_retention_days
  
  email_account_admins = var.email_admins_on_alert
  email_addresses      = var.alert_email_addresses
}

# Key Vault for storing secrets (optional)
resource "azurerm_key_vault" "main" {
  count               = var.create_key_vault ? 1 : 0
  name                = "${var.sql_server_name}-kv"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"

  soft_delete_retention_days = 7
  purge_protection_enabled   = false

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_client_config.current.object_id

    secret_permissions = [
      "Get", "List", "Set", "Delete", "Recover", "Backup", "Restore"
    ]
  }

  tags = var.tags
}

# Store SQL admin password in Key Vault
resource "azurerm_key_vault_secret" "sql_admin_password" {
  count        = var.create_key_vault ? 1 : 0
  name         = "sql-admin-password"
  value        = random_password.sql_admin_password.result
  key_vault_id = azurerm_key_vault.main[0].id

  depends_on = [
    azurerm_key_vault.main
  ]
}
