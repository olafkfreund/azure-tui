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

# Generate random suffix for unique naming
resource "random_id" "main" {
  byte_length = 4
}

# Create resource group
resource "azurerm_resource_group" "main" {
  name     = var.resource_group_name
  location = var.location

  tags = var.tags
}

# Create Log Analytics Workspace for monitoring
resource "azurerm_log_analytics_workspace" "main" {
  count               = var.enable_monitoring ? 1 : 0
  name                = "${var.cluster_name}-logs-${random_id.main.hex}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "PerGB2018"
  retention_in_days   = var.log_retention_days

  tags = var.tags
}

# Create virtual network
resource "azurerm_virtual_network" "main" {
  name                = "${var.cluster_name}-vnet"
  address_space       = ["10.0.0.0/16"]
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = var.tags
}

# Create AKS subnet
resource "azurerm_subnet" "aks" {
  name                 = "${var.cluster_name}-aks-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.1.0/24"]
}

# Create Application Gateway subnet (if enabled)
resource "azurerm_subnet" "appgw" {
  count                = var.enable_application_gateway ? 1 : 0
  name                 = "${var.cluster_name}-appgw-subnet"
  resource_group_name  = azurerm_resource_group.main.name
  virtual_network_name = azurerm_virtual_network.main.name
  address_prefixes     = ["10.0.2.0/24"]
}

# Create managed identity for AKS
resource "azurerm_user_assigned_identity" "aks" {
  name                = "${var.cluster_name}-identity"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name

  tags = var.tags
}

# Create AKS cluster
resource "azurerm_kubernetes_cluster" "main" {
  name                = var.cluster_name
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  dns_prefix          = "${var.cluster_name}-${random_id.main.hex}"
  kubernetes_version  = var.kubernetes_version

  # Default node pool
  default_node_pool {
    name                = "system"
    node_count          = var.system_node_count
    vm_size             = var.system_node_vm_size
    vnet_subnet_id      = azurerm_subnet.aks.id
    type                = "VirtualMachineScaleSets"
    zones               = var.availability_zones
    enable_auto_scaling = var.enable_auto_scaling
    min_count           = var.enable_auto_scaling ? var.system_node_min_count : null
    max_count           = var.enable_auto_scaling ? var.system_node_max_count : null
    os_disk_size_gb     = var.os_disk_size_gb
    os_disk_type        = var.os_disk_type

    tags = var.tags
  }

  # Identity configuration
  identity {
    type         = "UserAssigned"
    identity_ids = [azurerm_user_assigned_identity.aks.id]
  }

  # Network configuration
  network_profile {
    network_plugin     = var.network_plugin
    network_policy     = var.network_policy
    dns_service_ip     = "10.1.0.10"
    service_cidr       = "10.1.0.0/16"
    load_balancer_sku  = "standard"
  }

  # Azure AD integration
  azure_active_directory_role_based_access_control {
    admin_group_object_ids = var.admin_group_object_ids
    azure_rbac_enabled     = var.azure_rbac_enabled
  }

  # Monitoring and logging
  dynamic "oms_agent" {
    for_each = var.enable_monitoring ? [1] : []
    content {
      log_analytics_workspace_id = azurerm_log_analytics_workspace.main[0].id
    }
  }

  # Add-ons
  dynamic "ingress_application_gateway" {
    for_each = var.enable_application_gateway ? [1] : []
    content {
      gateway_name = "${var.cluster_name}-appgw"
      subnet_id    = azurerm_subnet.appgw[0].id
    }
  }

  dynamic "key_vault_secrets_provider" {
    for_each = var.enable_key_vault_secrets_provider ? [1] : []
    content {
      secret_rotation_enabled = true
    }
  }

  # Auto-upgrade
  automatic_channel_upgrade = var.automatic_channel_upgrade

  # Maintenance window
  dynamic "maintenance_window" {
    for_each = var.maintenance_window != null ? [var.maintenance_window] : []
    content {
      allowed {
        day   = maintenance_window.value.day
        hours = maintenance_window.value.hours
      }
    }
  }

  tags = var.tags

  depends_on = [
    azurerm_user_assigned_identity.aks
  ]
}

# Additional node pools
resource "azurerm_kubernetes_cluster_node_pool" "user" {
  count                 = length(var.additional_node_pools)
  name                  = var.additional_node_pools[count.index].name
  kubernetes_cluster_id = azurerm_kubernetes_cluster.main.id
  vm_size               = var.additional_node_pools[count.index].vm_size
  node_count            = var.additional_node_pools[count.index].node_count
  vnet_subnet_id        = azurerm_subnet.aks.id
  zones                 = var.availability_zones
  enable_auto_scaling   = var.additional_node_pools[count.index].enable_auto_scaling
  min_count             = var.additional_node_pools[count.index].enable_auto_scaling ? var.additional_node_pools[count.index].min_count : null
  max_count             = var.additional_node_pools[count.index].enable_auto_scaling ? var.additional_node_pools[count.index].max_count : null
  os_disk_size_gb       = var.additional_node_pools[count.index].os_disk_size_gb
  os_disk_type          = var.additional_node_pools[count.index].os_disk_type
  os_type               = var.additional_node_pools[count.index].os_type
  node_labels           = var.additional_node_pools[count.index].node_labels

  tags = var.tags
}

# Container Registry (if enabled)
resource "azurerm_container_registry" "main" {
  count               = var.create_container_registry ? 1 : 0
  name                = "${replace(var.cluster_name, "-", "")}acr${random_id.main.hex}"
  resource_group_name = azurerm_resource_group.main.name
  location            = azurerm_resource_group.main.location
  sku                 = var.container_registry_sku
  admin_enabled       = false

  # Private endpoint (for Premium SKU)
  dynamic "network_rule_set" {
    for_each = var.container_registry_sku == "Premium" ? [1] : []
    content {
      default_action = "Deny"
    }
  }

  tags = var.tags
}

# Assign AcrPull role to AKS managed identity
resource "azurerm_role_assignment" "aks_acr_pull" {
  count                = var.create_container_registry ? 1 : 0
  scope                = azurerm_container_registry.main[0].id
  role_definition_name = "AcrPull"
  principal_id         = azurerm_kubernetes_cluster.main.kubelet_identity[0].object_id
}

# Key Vault for secrets (if enabled)
resource "azurerm_key_vault" "main" {
  count               = var.create_key_vault ? 1 : 0
  name                = "${var.cluster_name}-kv-${random_id.main.hex}"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"

  soft_delete_retention_days = 7
  purge_protection_enabled   = false

  network_acls {
    default_action = "Allow"
    bypass         = "AzureServices"
  }

  tags = var.tags
}

# Get current Azure client configuration
data "azurerm_client_config" "current" {}

# Key Vault access policy for AKS
resource "azurerm_key_vault_access_policy" "aks" {
  count        = var.create_key_vault && var.enable_key_vault_secrets_provider ? 1 : 0
  key_vault_id = azurerm_key_vault.main[0].id
  tenant_id    = data.azurerm_client_config.current.tenant_id
  object_id    = azurerm_kubernetes_cluster.main.key_vault_secrets_provider[0].secret_identity[0].object_id

  secret_permissions = [
    "Get", "List"
  ]
}
