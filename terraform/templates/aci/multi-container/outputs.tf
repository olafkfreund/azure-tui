output "resource_group_name" {
  description = "Name of the created resource group"
  value       = azurerm_resource_group.main.name
}

output "virtual_machine_id" {
  description = "ID of the created virtual machine"
  value       = azurerm_linux_virtual_machine.main.id
}

output "virtual_machine_name" {
  description = "Name of the created virtual machine"
  value       = azurerm_linux_virtual_machine.main.name
}

output "public_ip_address" {
  description = "Public IP address of the virtual machine"
  value       = azurerm_public_ip.main.ip_address
}

output "private_ip_address" {
  description = "Private IP address of the virtual machine"
  value       = azurerm_network_interface.main.private_ip_address
}

output "admin_username" {
  description = "Admin username for SSH access"
  value       = azurerm_linux_virtual_machine.main.admin_username
}

output "ssh_private_key" {
  description = "Private SSH key for connecting to the VM"
  value       = tls_private_key.ssh.private_key_pem
  sensitive   = true
}

output "ssh_public_key" {
  description = "Public SSH key used for the VM"
  value       = tls_private_key.ssh.public_key_openssh
}

output "ssh_connection_command" {
  description = "SSH command to connect to the VM"
  value       = "ssh -i private_key.pem ${azurerm_linux_virtual_machine.main.admin_username}@${azurerm_public_ip.main.ip_address}"
}

output "network_security_group_id" {
  description = "ID of the network security group"
  value       = azurerm_network_security_group.main.id
}

output "virtual_network_id" {
  description = "ID of the virtual network"
  value       = azurerm_virtual_network.main.id
}

output "subnet_id" {
  description = "ID of the subnet"
  value       = azurerm_subnet.main.id
}
