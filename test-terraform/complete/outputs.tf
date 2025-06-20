output "resource_group_name" {
  description = "Name of the created resource group"
  value       = azurerm_resource_group.test.name
}

output "storage_account_name" {
  description = "Name of the created storage account"
  value       = azurerm_storage_account.test.name
}
