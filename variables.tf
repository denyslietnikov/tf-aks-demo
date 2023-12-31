variable "resource_group_location" {
  type        = string
  default     = "eastus"
  description = "Location of the resource group."
}

variable "resource_group_name_prefix" {
  type        = string
  default     = "rg"
  description = "Prefix of the resource group name that's combined with a random ID so name is unique in your Azure subscription."
}

variable "node_count" {
  type        = number
  description = "The initial quantity of nodes for the node pool."
  default     = 2
}

variable "msi_id" {
  type        = string
  description = "The Managed Service Identity ID. Set this value if you're running this example using Managed Identity as the authentication method."
  default     = null
}


variable "aks-demo-kv-tg-token" {
  type        = string
  description = ""
  default     = "token"
}

#variable "aks-demo-kv-user" {
#  type        = string
#  description = ""
#  default     = "user"
#}

#variable "aks-demo-kv-password" {
#  type        = string
#  description = ""
#  default     = "password"
#}

//SQL
variable "aks-demo-sql-server-name" {
  type        = string
  description = ""
  default     = "sqlservername"
}

variable "aks-demo-sql-server-login" {
  type        = string
  description = ""
  default     = "login"
}

variable "aks-demo-sql-server-password" {
  type        = string
  description = ""
  default     = "password"
}

variable "aks-demo-sql-server-dbname" {
  type        = string
  description = ""
  default     = "dbname"
}

variable "aks-demo-sql-server-port" {
  type        = string
  description = ""
  default     = "1433"
}

