# Configure the Azure Provider
terraform {
  required_version = ">= 1.0"
  required_providers {
    azurerm = {
      source  = "hashicorp/azurerm"
      version = "~> 3.0"
    }
  }
}

provider "azurerm" {
  features {}
}

# Create resource group
resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location

  tags = var.tags
}

# Multi-Container Instance
resource "azurerm_container_group" "main" {
  name                = var.container_group_name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  ip_address_type     = var.ip_address_type
  dns_name_label      = var.dns_name_label
  os_type             = var.os_type
  restart_policy      = var.restart_policy

  # Web container (nginx)
  container {
    name   = "web"
    image  = var.web_image
    cpu    = var.web_cpu
    memory = var.web_memory

    ports {
      port     = 80
      protocol = "TCP"
    }

    environment_variables = var.web_environment_variables
  }

  # API container (httpd)
  container {
    name   = "api"
    image  = var.api_image
    cpu    = var.api_cpu
    memory = var.api_memory

    ports {
      port     = 8080
      protocol = "TCP"
    }

    environment_variables = var.api_environment_variables

    liveness_probe {
      failure_threshold     = 3
      initial_delay_seconds = 30
      period_seconds        = 10
      success_threshold     = 1
      timeout_seconds       = 5
      
      http_get {
        path   = "/health"
        port   = 8080
        scheme = "HTTP"
      }
    }

    readiness_probe {
      failure_threshold     = 3
      initial_delay_seconds = 5
      period_seconds        = 10
      success_threshold     = 1
      timeout_seconds       = 5
      
      http_get {
        path   = "/ready"
        port   = 8080
        scheme = "HTTP"
      }
    }
  }

  # Exposed ports for the container group
  exposed_port {
    port     = 80
    protocol = "TCP"
  }

  exposed_port {
    port     = 8080
    protocol = "TCP"
  }

  tags = var.tags
}

# Optional: Log Analytics Workspace for diagnostics
resource "azurerm_log_analytics_workspace" "main" {
  count               = var.enable_diagnostics ? 1 : 0
  name                = "${var.container_group_name}-logs"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "PerGB2018"
  retention_in_days   = var.log_retention_days

  tags = var.tags
}
