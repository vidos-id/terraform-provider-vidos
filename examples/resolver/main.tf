terraform {
  required_version = ">= 1.6.0"

  required_providers {
    vidos = {
      source = "vidos/vidos"
    }
  }
}

provider "vidos" {
  region  = var.vidos_region
  api_key = var.vidos_api_key

}

resource "vidos_resolver_configuration" "example" {
  name = "terraform-resolver-config"

  # Replace with resolver configuration.
  values = jsonencode({
    "methods" : {
      "cheqd" : {
        "enabled" : false
      }
    }
  })
}

resource "vidos_resolver_instance" "example" {
  name = "terraform-resolver-instance"

  configuration_resource_id = vidos_resolver_configuration.example.resource_id

  # Optional alternative: inline configuration (must be valid JSON).
  # inline_configuration = jsonencode({
  #   
  # })
}
