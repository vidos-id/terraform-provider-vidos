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

resource "vidos_authorizer_instance" "main" {
  name = "terraform-example-authorizer-instance"

  configuration_resource_id = vidos_authorizer_configuration.example.resource_id
}

resource "vidos_gateway_instance" "example" {
  name = "terraform-example-gateway-instance"
  inline_configuration = jsonencode({
    cors = {
      enabled      = true
      allowHeaders = ["*"]
    }
    paths = {
      # Requests to /auth/* are forwarded to the authorizer instance.
      auth = {
        type       = "instance"
        service    = "authorizer"
        resourceId = vidos_authorizer_instance.main.resource_id

        # Managed service role used for gateway -> authorizer service-to-service auth.
        serviceRole = {
          owner      = "managed"
          resourceId = "authorizer_all_actions"
        }
      }
    }
  })
}
