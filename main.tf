# Generate random resource group name
resource "random_pet" "rg_name" {
  prefix = var.resource_group_name_prefix
}

resource "azurerm_resource_group" "rg" {
  location = var.resource_group_location
  name     = random_pet.rg_name.id
}

resource "random_pet" "azurerm_kubernetes_cluster_name" {
  prefix = "cluster"
}

resource "random_pet" "azurerm_kubernetes_cluster_dns_prefix" {
  prefix = "dns"
}

resource "azurerm_kubernetes_cluster" "k8s" {
  location            = azurerm_resource_group.rg.location
  name                = random_pet.azurerm_kubernetes_cluster_name.id
  resource_group_name = azurerm_resource_group.rg.name
  dns_prefix          = random_pet.azurerm_kubernetes_cluster_dns_prefix.id

  identity {
    type = "SystemAssigned"
  }

  default_node_pool {
    name    = "agentpool"
    vm_size = "Standard_B4ms"
    //    vm_size    = "Standard_D2_v2"
    node_count = var.node_count
  }
  linux_profile {
    admin_username = "ubuntu"

    ssh_key {
      key_data = jsondecode(azapi_resource_action.ssh_public_key_gen.output).publicKey
    }
  }
  key_vault_secrets_provider {
    secret_rotation_enabled = true
  }
  network_profile {
    network_plugin    = "kubenet"
    load_balancer_sku = "standard"
  }
}


resource "azurerm_kubernetes_cluster_extension" "flux" {
  name           = "flux"
  cluster_id     = azurerm_kubernetes_cluster.k8s.id
  extension_type = "microsoft.flux"
}

resource "azurerm_kubernetes_flux_configuration" "flux-config" {
  name       = "gitops-flux-aks-demo"
  cluster_id = azurerm_kubernetes_cluster.k8s.id
  namespace  = "aks-demo"
  scope      = "cluster"

  git_repository {
    url             = "https://github.com/denyslietnikov/tf-aks-demo"
    reference_type  = "branch"
    reference_value = "main"
  }

  kustomizations {
    name               = "infra"
    path               = "./clusters/flux-system/infra"
    recreating_enabled = true
  }
  kustomizations {
    name               = "log"
    path               = "./clusters/flux-system/log"
    recreating_enabled = true
    depends_on         = ["infra"]
  }
  kustomizations {
    name               = "bot"
    path               = "./clusters/flux-system/bot"
    recreating_enabled = true
    depends_on         = ["job"]
  }
  kustomizations {
    name               = "job"
    path               = "./clusters/flux-system/job"
    recreating_enabled = true
    depends_on         = ["log"]
  }

  depends_on = [
    azurerm_kubernetes_cluster_extension.flux
  ]
}

# Key Vault

provider "azurerm" {
  features {
    key_vault {
      purge_soft_deleted_secrets_on_destroy = true
      recover_soft_deleted_secrets          = false
    }
  }
}

data "azurerm_client_config" "current" {}


resource "azurerm_key_vault" "aks-demo-kv" {
  name                = "aks-demo-kv"
  location            = azurerm_resource_group.rg.location
  resource_group_name = azurerm_resource_group.rg.name
  tenant_id           = data.azurerm_client_config.current.tenant_id
  sku_name            = "standard"
  #soft_delete_retention_days = 7

  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_client_config.current.object_id

    key_permissions = [
      "Create",
      "Get"
    ]

    secret_permissions = [
      "Set",
      "Get",
      "Delete",
      "Purge",
      "Recover",
      "List"
    ]
  }
  access_policy {
    tenant_id = data.azurerm_client_config.current.tenant_id
    object_id = data.azurerm_user_assigned_identity.identity.principal_id

    secret_permissions = [
      "Get",
      "List"
    ]
  }
}


resource "azurerm_key_vault_secret" "aks-demo-kv-tg-token" {
  name         = "aks-demo-kv-tg-token"
  value        = var.aks-demo-kv-tg-token
  key_vault_id = azurerm_key_vault.aks-demo-kv.id
}

resource "azurerm_key_vault_secret" "aks-demo-kv-user" {
  name         = "aks-demo-kv-user"
  value        = var.aks-demo-sql-server-login
  key_vault_id = azurerm_key_vault.aks-demo-kv.id
}

resource "azurerm_key_vault_secret" "aks-demo-kv-password" {
  name         = "aks-demo-kv-password"
  value        = var.aks-demo-sql-server-password
  key_vault_id = azurerm_key_vault.aks-demo-kv.id
}


# Azure Key Vault Role Assignment for Managed Identity for AKS
resource "azurerm_role_assignment" "assign_kv_admin_role_for_yourself" {
  scope                = "/subscriptions/${data.azurerm_client_config.current.subscription_id}/resourceGroups/${azurerm_resource_group.rg.name}/providers/Microsoft.KeyVault/vaults/aks-demo-kv"
  role_definition_name = "Key Vault Administrator"
  principal_id         = data.azurerm_client_config.current.object_id
}

data "azurerm_user_assigned_identity" "identity" {
  name                = "azurekeyvaultsecretsprovider-${azurerm_kubernetes_cluster.k8s.name}"
  resource_group_name = "MC_${azurerm_resource_group.rg.name}_${azurerm_kubernetes_cluster.k8s.name}_${var.resource_group_location}"
}

resource "azurerm_role_assignment" "assign_kv_admin_role" {
  scope                = "/subscriptions/${data.azurerm_client_config.current.subscription_id}/resourceGroups/${azurerm_resource_group.rg.name}/providers/Microsoft.KeyVault/vaults/aks-demo-kv"
  role_definition_name = "Key Vault Administrator"
  principal_id         = data.azurerm_user_assigned_identity.identity.principal_id
}



### SQL Server and Database ###

resource "azurerm_mssql_server" "mssqlserver" {
  name                         = var.aks-demo-sql-server-name
  resource_group_name          = azurerm_resource_group.rg.name
  location                     = azurerm_resource_group.rg.location
  version                      = "12.0"
  administrator_login          = var.aks-demo-sql-server-login
  administrator_login_password = var.aks-demo-sql-server-password
  minimum_tls_version          = "1.2"

  identity {
    type = "SystemAssigned"
  }

  tags = {
    env = "test"
  }

}

resource "azurerm_mssql_database" "db" {
  name         = var.aks-demo-sql-server-dbname
  server_id    = azurerm_mssql_server.mssqlserver.id
  collation    = "SQL_Latin1_General_CP1_CI_AS"
  license_type = "LicenseIncluded"
  #  max_size_gb    = 4
  #  read_scale     = true
  sku_name       = "Basic"
  zone_redundant = false

}

resource "azurerm_mssql_firewall_rule" "mssqlfirewallrule" {
  name             = "Allow access to Azure services"
  server_id        = azurerm_mssql_server.mssqlserver.id
  start_ip_address = "0.0.0.0"
  end_ip_address   = "0.0.0.0"
}

data "azurerm_user_assigned_identity" "aks_identity" {
  name                = "${azurerm_kubernetes_cluster.k8s.name}-agentpool"
  resource_group_name = "MC_${azurerm_resource_group.rg.name}_${azurerm_kubernetes_cluster.k8s.name}_${var.resource_group_location}"
}

resource "azurerm_role_assignment" "assign_mssql_contributor_role" {
  scope                = azurerm_mssql_server.mssqlserver.id
  role_definition_name = "Contributor"
  principal_id         = data.azurerm_user_assigned_identity.aks_identity.principal_id
}