# Resource Group
output "resource_group_name" {
  description = "Name of the created resource group"
  value       = azurerm_resource_group.main.name
}

output "resource_group_location" {
  description = "Location of the created resource group"
  value       = azurerm_resource_group.main.location
}

# Container Group
output "container_group_name" {
  description = "Name of the container group"
  value       = azurerm_container_group.main.name
}

output "container_group_ip_address" {
  description = "IP address of the container group"
  value       = azurerm_container_group.main.ip_address
}

output "container_group_fqdn" {
  description = "Fully qualified domain name of the container group"
  value       = azurerm_container_group.main.fqdn
}

# Web Container
output "web_container_url" {
  description = "URL to access the web container"
  value       = "http://${azurerm_container_group.main.fqdn}:80"
}

# API Container
output "api_container_url" {
  description = "URL to access the API container"
  value       = "http://${azurerm_container_group.main.fqdn}:8080"
}

# Diagnostics
output "log_analytics_workspace_id" {
  description = "Log Analytics workspace ID (if diagnostics enabled)"
  value       = var.enable_diagnostics ? azurerm_log_analytics_workspace.main[0].workspace_id : null
}

output "log_analytics_workspace_name" {
  description = "Log Analytics workspace name (if diagnostics enabled)"
  value       = var.enable_diagnostics ? azurerm_log_analytics_workspace.main[0].name : null
}
