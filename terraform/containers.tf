# Container Group 1 - Hello World
resource "azurerm_container_group" "helloworld_1" {
  name                = "${var.project_name}-container-1"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  ip_address_type     = "Public"
  dns_name_label      = "${var.project_name}-hello1-${random_string.suffix.result}"
  os_type             = "Linux"
  tags                = var.tags

  container {
    name   = "hello-world-1"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    cpu    = var.container_cpu
    memory = var.container_memory

    ports {
      port     = 80
      protocol = "TCP"
    }

    environment_variables = {
      "TITLE" = "Azure TUI Demo - Container 1"
    }
  }
}

# Container Group 2 - Hello World
resource "azurerm_container_group" "helloworld_2" {
  name                = "${var.project_name}-container-2"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  ip_address_type     = "Public"
  dns_name_label      = "${var.project_name}-hello2-${random_string.suffix.result}"
  os_type             = "Linux"
  tags                = var.tags

  container {
    name   = "hello-world-2"
    image  = "mcr.microsoft.com/azuredocs/aci-helloworld:latest"
    cpu    = var.container_cpu
    memory = var.container_memory

    ports {
      port     = 80
      protocol = "TCP"
    }

    environment_variables = {
      "TITLE" = "Azure TUI Demo - Container 2"
    }
  }
}

# Random string for unique DNS names
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}
