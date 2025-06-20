output "resource_group_name" {
  description = "Name of the created resource group"
  value       = azurerm_resource_group.main.name
}

output "container_group_id" {
  description = "ID of the container group"
  value       = azurerm_container_group.main.id
}

output "container_group_name" {
  description = "Name of the container group"
  value       = azurerm_container_group.main.name
}

output "ip_address" {
  description = "IP address of the container group"
  value       = azurerm_container_group.main.ip_address
}

output "fqdn" {
  description = "FQDN of the container group"
  value       = azurerm_container_group.main.fqdn
}

output "container_urls" {
  description = "URLs to access the containers"
  value = [
    for port in var.ports : "http://${azurerm_container_group.main.fqdn != "" ? azurerm_container_group.main.fqdn : azurerm_container_group.main.ip_address}:${port.port}"
  ]
}

output "identity" {
  description = "Managed identity information (if enabled)"
  value = var.enable_system_assigned_identity ? {
    principal_id = azurerm_container_group.main.identity[0].principal_id
    tenant_id    = azurerm_container_group.main.identity[0].tenant_id
  } : null
}

output "log_analytics_workspace_id" {
  description = "ID of the Log Analytics workspace (if created)"
  value       = var.enable_log_analytics && var.log_analytics_workspace_id == "" ? azurerm_log_analytics_workspace.main[0].id : var.log_analytics_workspace_id
}

output "storage_account_name" {
  description = "Name of the storage account (if created)"
  value       = null
}

output "storage_account_key" {
  description = "Primary access key of the storage account (if created)"
  value       = null
  sensitive   = true
}

output "container_info" {
  description = "Information about all containers in the group"
  value = {
    main_container = {
      name   = var.container_name
      image  = var.container_image
      cpu    = var.cpu_cores
      memory = var.memory_gb
      ports  = var.ports
    }
    additional_containers = var.additional_containers
  }
}

output "connection_info" {
  description = "Connection information for the container group"
  value = {
    ip_address = azurerm_container_group.main.ip_address
    fqdn       = azurerm_container_group.main.fqdn
    ports      = var.ports
    os_type    = var.os_type
  }
}
