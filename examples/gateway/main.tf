terraform {
  required_version = ">= 1.6.0"

  required_providers {
    vidos = {
      source  = "registry.terraform.io/vidos-id/vidos"
      version = "~> 0.1"
    }
  }
}

provider "vidos" {
  region = var.vidos_region

  # required (or set VIDOS_API_KEY)
  api_key = var.vidos_api_key
}

resource "vidos_gateway_configuration" "example" {
  name   = "terraform-example-gateway-config"
  values = jsonencode({})
}

resource "vidos_gateway_instance" "example" {
  name = "terraform-example-gateway-instance"

  configuration_resource_id = vidos_gateway_configuration.example.resource_id
}
