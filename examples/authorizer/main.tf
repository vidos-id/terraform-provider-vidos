terraform {
  required_version = ">= 1.6.0"

  required_providers {
    vidos = {
      source = "vidos/vidos"
    }
  }
}

provider "vidos" {
  region = var.vidos_region

  # required (or set VIDOS_API_KEY)
  api_key = var.vidos_api_key
}

resource "vidos_authorizer_configuration" "example" {
  name = "terraform-example-authorizer-config"

  values = jsonencode({
    policies = {
      verify = {
        skip = true
      }
      validate = {
        skip = true
      }
    }
  })
}

resource "vidos_authorizer_instance" "example" {
  name = "terraform-example-authorizer-instance"

  configuration_resource_id = vidos_authorizer_configuration.example.resource_id
}
