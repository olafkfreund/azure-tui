# Public IP for Virtual Machine
resource "azurerm_public_ip" "vm_public_ip" {
  name                = "${var.project_name}-vm-public-ip"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  allocation_method   = "Static"
  sku                 = "Standard"
  tags                = var.tags
}

# Network Interface for Virtual Machine
resource "azurerm_network_interface" "vm_nic" {
  name                = "${var.project_name}-vm-nic"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tags                = var.tags

  ip_configuration {
    name                          = "internal"
    subnet_id                     = azurerm_subnet.main.id
    private_ip_address_allocation = "Dynamic"
    public_ip_address_id          = azurerm_public_ip.vm_public_ip.id
  }
}

# Virtual Machine
resource "azurerm_linux_virtual_machine" "main" {
  name                = "${var.project_name}-vm"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  size                = var.vm_size
  admin_username      = var.admin_username
  tags                = var.tags

  # Disable password authentication and use SSH keys
  disable_password_authentication = true

  network_interface_ids = [
    azurerm_network_interface.vm_nic.id,
  ]

  admin_ssh_key {
    username   = var.admin_username
    public_key = file("~/.ssh/id_rsa.pub") # You may need to adjust this path
  }

  os_disk {
    caching              = "ReadWrite"
    storage_account_type = "Premium_LRS"
  }

  source_image_reference {
    publisher = "Canonical"
    offer     = "0001-com-ubuntu-server-jammy"
    sku       = "22_04-lts-gen2"
    version   = "latest"
  }

  # Install Docker and run a simple web server
  custom_data = base64encode(<<-EOF
    #!/bin/bash
    apt-get update
    apt-get install -y docker.io nginx
    systemctl enable docker
    systemctl start docker
    systemctl enable nginx
    systemctl start nginx
    
    # Create a simple index page
    echo "<h1>Azure TUI Demo VM</h1><p>This VM was created via Terraform from Azure TUI!</p><p>Hostname: $(hostname)</p><p>Date: $(date)</p>" > /var/www/html/index.html
    
    # Run a hello world container
    docker run -d -p 8080:80 --name hello-world nginxdemos/hello
  EOF
  )
}
