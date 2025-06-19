# Output values for Azure TUI Terraform

# Resource Group
output "resource_group_name" {
  description = "Name of the resource group"
  value       = azurerm_resource_group.main.name
}

output "resource_group_location" {
  description = "Location of the resource group"
  value       = azurerm_resource_group.main.location
}

# Virtual Machine
output "vm_public_ip" {
  description = "Public IP address of the Virtual Machine"
  value       = azurerm_public_ip.vm_public_ip.ip_address
}

output "vm_private_ip" {
  description = "Private IP address of the Virtual Machine"
  value       = azurerm_network_interface.vm_nic.private_ip_address
}

output "vm_ssh_connection" {
  description = "SSH connection string for the VM"
  value       = "ssh ${var.admin_username}@${azurerm_public_ip.vm_public_ip.ip_address}"
}

# AKS Cluster
output "aks_cluster_name" {
  description = "Name of the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.name
}

output "aks_cluster_fqdn" {
  description = "FQDN of the AKS cluster"
  value       = azurerm_kubernetes_cluster.main.fqdn
}

output "aks_kubeconfig_command" {
  description = "Command to get kubeconfig for AKS cluster"
  value       = "az aks get-credentials --resource-group ${azurerm_resource_group.main.name} --name ${azurerm_kubernetes_cluster.main.name}"
}

# Container Instances
output "container_instances" {
  description = "Container instance details"
  value = {
    helloworld_1 = {
      fqdn = azurerm_container_group.helloworld_1.fqdn
      ip   = azurerm_container_group.helloworld_1.ip_address
      url  = "http://${azurerm_container_group.helloworld_1.fqdn}"
    }
    helloworld_2 = {
      fqdn = azurerm_container_group.helloworld_2.fqdn
      ip   = azurerm_container_group.helloworld_2.ip_address
      url  = "http://${azurerm_container_group.helloworld_2.fqdn}"
    }
  }
}

# Networking
output "virtual_network_name" {
  description = "Name of the virtual network"
  value       = azurerm_virtual_network.main.name
}

output "subnet_id" {
  description = "ID of the main subnet"
  value       = azurerm_subnet.main.id
}

output "aks_subnet_id" {
  description = "ID of the AKS subnet"
  value       = azurerm_subnet.aks.id
}

# Summary
output "deployment_summary" {
  description = "Summary of deployed resources"
  value = {
    resource_group = azurerm_resource_group.main.name
    location       = azurerm_resource_group.main.location
    vm_ip          = azurerm_public_ip.vm_public_ip.ip_address
    aks_cluster    = azurerm_kubernetes_cluster.main.name
    container_urls = [
      "http://${azurerm_container_group.helloworld_1.fqdn}",
      "http://${azurerm_container_group.helloworld_2.fqdn}"
    ]
  }
}
