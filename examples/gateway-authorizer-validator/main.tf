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

locals {
  # Example PEM-encoded root certificates for validator trust anchors.
  # Replace these with real roots for your environment.
  valera_test_certificate = trimspace(<<-PEM
-----BEGIN CERTIFICATE-----
MIICGzCCAcCgAwIBAgIUb9GJdqQMdwXaoO61uxoBlg+jhbYwCgYIKoZIzj0EAwIw
LDELMAkGA1UEBhMCQVQxDjAMBgNVBAoMBUEtU0lUMQ0wCwYDVQQDDARJQUNBMB4X
DTI1MDQwNzA5NDQ1N1oXDTI2MDQwNzA5NDQ1N1owLDELMAkGA1UEBhMCQVQxDjAM
BgNVBAoMBUEtU0lUMQ0wCwYDVQQDDARJQUNBMFkwEwYHKoZIzj0CAQYIKoZIzj0D
AQcDQgAElIXOzb+iF+zGutygdIVOBnC4R6OvhYo5TGWhrH0idmqs56IVwJWYzQYz
K4CbYePcxpMQY3lKBa5O0MAZe+EogKOBvzCBvDASBgNVHRMBAf8ECDAGAQH/AgEA
MA4GA1UdDwEB/wQEAwIBBjAiBgNVHRIEGzAZhhdodHRwczovL3dhbGxldC5hLXNp
dC5hdDAyBgNVHR8EKzApMCegJaAjhiFodHRwczovL3dhbGxldC5hLXNpdC5hdC9j
cmwvMS5jcmwwHwYDVR0jBBgwFoAUDQF5K46YVgzLpfV5stoutBezK6QwHQYDVR0O
BBYEFA0BeSuOmFYMy6X1ebLaLrQXsyukMAoGCCqGSM49BAMCA0kAMEYCIQCz0i9G
A24ZOf3Wk+w8+09J6ARAHKLuBuepszBxVZdaZAIhAJlgzKBhHw8+Bwr+wLGQVjMC
5e9BWWaUga8ZP9dRYhHJ
-----END CERTIFICATE-----
  PEM
  )

  multipaz_certificate = trimspace(<<-PEM
-----BEGIN CERTIFICATE-----
MIICpjCCAi2gAwIBAgIQiiieDKBRbQvx4FJgTHQFbTAKBggqhkjOPQQDAzAuMR8wHQYDVQQDDBZP
V0YgTXVsdGlwYXogVEVTVCBJQUNBMQswCQYDVQQGDAJVUzAeFw0yNDEyMDEwMDAwMDBaFw0zNDEy
MDEwMDAwMDBaMC4xHzAdBgNVBAMMFk9XRiBNdWx0aXBheiBURVNUIElBQ0ExCzAJBgNVBAYMAlVT
MHYwEAYHKoZIzj0CAQYFK4EEACIDYgAE+QDye70m2O0llPXMjVjxVZz3m5k6agT+wih+L79b7jyq
Ul99sbeUnpxaLD+cmB3HK3twkA7fmVJSobBc+9CDhkh3mx6n+YoH5RulaSWThWBfMyRjsfVODkos
HLCDnbPVo4IBDjCCAQowDgYDVR0PAQH/BAQDAgEGMBIGA1UdEwEB/wQIMAYBAf8CAQAwTAYDVR0S
BEUwQ4ZBaHR0cHM6Ly9naXRodWIuY29tL29wZW53YWxsZXQtZm91bmRhdGlvbi1sYWJzL2lkZW50
aXR5LWNyZWRlbnRpYWwwVgYDVR0fBE8wTTBLoEmgR4ZFaHR0cHM6Ly9naXRodWIuY29tL29wZW53
YWxsZXQtZm91bmRhdGlvbi1sYWJzL2lkZW50aXR5LWNyZWRlbnRpYWwvY3JsMB0GA1UdDgQWBBSr
ZRvgVsKQU/Hdf2zkh75o3mDJ9TAfBgNVHSMEGDAWgBSrZRvgVsKQU/Hdf2zkh75o3mDJ9TAKBggq
hkjOPQQDAwNnADBkAjAtTLS7FfsbUe/SKlIhYgnDcD6fDgiUaUR4htNhFVHPA4d8OlUGqmof76xi
eBjEc9MCMGKk27tss0KCk93qaRsZ7NuAGWMSun6mraePJ7PUpaYz2/6zztu51kYK6NftObq4fw==
-----END CERTIFICATE-----
  PEM
  )
}

resource "vidos_validator_instance" "main" {
  name = "terraform-example-validator-instance"

  inline_configuration = jsonencode({
    policies = {
      # Enforce that issuers chain to one of these roots.
      trustedIssuer = {
        skip = false
        trustedIssuerRootCertificates = [
          { type = "predefined", tag = "vidos" },
          { type = "pem", pem = local.valera_test_certificate },
          { type = "pem", pem = local.multipaz_certificate },
        ]
      }
    }
  })
}

resource "vidos_authorizer_instance" "main" {
  name = "terraform-example-authorizer-instance"

  inline_configuration = jsonencode({
    policies = {
      validate = {
        skip = false
        validator = {
          type       = "instance"
          resourceId = vidos_validator_instance.main.resource_id
          serviceRole = {
            owner      = "managed"
            resourceId = "validator_all_actions"
          }
        }
      }
  })
}

resource "vidos_gateway_instance" "main" {
  name = "terraform-example-gateway-instance"

  inline_configuration = jsonencode({
    cors = {
      enabled      = true
      allowHeaders = ["*"]
      origin       = ["*"]
    }

    paths = {
      # Requests to /auth/* are forwarded to the authorizer instance.
      auth = {
        type       = "instance"
        service    = "authorizer"
        resourceId = vidos_authorizer_instance.main.resource_id
        serviceRole = {
          owner      = "managed"
          resourceId = "authorizer_all_actions"
        }
      }
    }
  })
}
