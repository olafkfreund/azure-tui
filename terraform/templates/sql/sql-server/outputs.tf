output "resource_group_name" {
  description = "Name of the created resource group"
  value       = azurerm_resource_group.main.name
}

output "sql_server_id" {
  description = "ID of the SQL Server"
  value       = azurerm_mssql_server.main.id
}

output "sql_server_name" {
  description = "Name of the SQL Server"
  value       = azurerm_mssql_server.main.name
}

output "sql_server_fqdn" {
  description = "Fully qualified domain name of the SQL Server"
  value       = azurerm_mssql_server.main.fully_qualified_domain_name
}

output "database_id" {
  description = "ID of the SQL Database"
  value       = azurerm_mssql_database.main.id
}

output "database_name" {
  description = "Name of the SQL Database"
  value       = azurerm_mssql_database.main.name
}

output "sql_admin_username" {
  description = "SQL Server administrator username"
  value       = azurerm_mssql_server.main.administrator_login
}

output "sql_admin_password" {
  description = "SQL Server administrator password"
  value       = random_password.sql_admin_password.result
  sensitive   = true
}

output "connection_string" {
  description = "SQL Server connection string"
  value       = "Server=tcp:${azurerm_mssql_server.main.fully_qualified_domain_name},1433;Initial Catalog=${azurerm_mssql_database.main.name};Persist Security Info=False;User ID=${azurerm_mssql_server.main.administrator_login};Password=${random_password.sql_admin_password.result};MultipleActiveResultSets=False;Encrypt=True;TrustServerCertificate=False;Connection Timeout=30;"
  sensitive   = true
}

output "key_vault_id" {
  description = "ID of the Key Vault (if created)"
  value       = var.create_key_vault ? azurerm_key_vault.main[0].id : null
}

output "key_vault_name" {
  description = "Name of the Key Vault (if created)"
  value       = var.create_key_vault ? azurerm_key_vault.main[0].name : null
}

output "virtual_network_id" {
  description = "ID of the Virtual Network (if created)"
  value       = var.create_virtual_network ? azurerm_virtual_network.main[0].id : null
}

output "subnet_id" {
  description = "ID of the SQL subnet (if created)"
  value       = var.create_virtual_network ? azurerm_subnet.sql[0].id : null
}

output "audit_storage_account_name" {
  description = "Name of the audit storage account (if auditing enabled)"
  value       = var.enable_auditing ? azurerm_storage_account.audit[0].name : null
}

output "sql_server_identity_principal_id" {
  description = "Principal ID of the SQL Server managed identity"
  value       = azurerm_mssql_server.main.identity[0].principal_id
}

output "sql_server_identity_tenant_id" {
  description = "Tenant ID of the SQL Server managed identity"
  value       = azurerm_mssql_server.main.identity[0].tenant_id
}
