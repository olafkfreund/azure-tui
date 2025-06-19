# AKS Cluster
resource "azurerm_kubernetes_cluster" "main" {
  name                = "${var.project_name}-aks"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  dns_prefix          = "${var.project_name}-aks"
  kubernetes_version  = var.kubernetes_version
  tags                = var.tags

  default_node_pool {
    name           = "default"
    node_count     = var.aks_node_count
    vm_size        = var.aks_node_size
    vnet_subnet_id = azurerm_subnet.aks.id
    
    # Enable auto-scaling for cost optimization
    enable_auto_scaling = true
    min_count          = 1
    max_count          = 3
  }

  identity {
    type = "SystemAssigned"
  }

  network_profile {
    network_plugin    = "azure"
    load_balancer_sku = "standard"
    outbound_type     = "loadBalancer"
  }

  # Enable various add-ons for better functionality
  oms_agent {
    log_analytics_workspace_id = azurerm_log_analytics_workspace.main.id
  }

  azure_policy_enabled = true
  
  # Enable HTTP application routing for easy testing
  http_application_routing_enabled = true
}

# Log Analytics Workspace for AKS monitoring
resource "azurerm_log_analytics_workspace" "main" {
  name                = "${var.project_name}-law"
  location            = azurerm_resource_group.main.location
  resource_group_name = azurerm_resource_group.main.name
  sku                 = "PerGB2018"
  retention_in_days   = 30
  tags                = var.tags
}

# Role assignment for AKS to access the subnet
resource "azurerm_role_assignment" "aks_subnet" {
  scope                = azurerm_subnet.aks.id
  role_definition_name = "Network Contributor"
  principal_id         = azurerm_kubernetes_cluster.main.identity[0].principal_id
}
