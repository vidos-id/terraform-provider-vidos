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

resource "vidos_resolver_configuration" "example" {
  name = "terraform-example-resolver-config"

  values = jsonencode({})
}

resource "vidos_resolver_instance" "example" {
  name = "terraform-example-resolver-instance"

  configuration_resource_id = vidos_resolver_configuration.example.resource_id
}

resource "vidos_verifier_configuration" "example" {
  name = "terraform-example-verifier-config"

  values = jsonencode({
    cors = {
      enabled = true
      origin  = ["*"]
    }

    policies = {
      proof = {
        skip = true
      }
    }

    resolver = {
      type       = "instance"
      resourceId = vidos_resolver_instance.example.resource_id

      serviceRole = {
        owner      = "managed"
        resourceId = "resolver_all_actions"
      }
    }
  })
}

resource "vidos_verifier_instance" "example" {
  name = "terraform-example-verifier-instance"

  configuration_resource_id = vidos_verifier_configuration.example.resource_id
}
