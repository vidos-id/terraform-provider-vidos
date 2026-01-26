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
  # IAM is always global; this region is for non-IAM services.
  region = var.vidos_region

  # required (or set VIDOS_API_KEY)
  api_key = var.vidos_api_key
}

resource "vidos_iam_api_key" "example" {
  name = "terraform-example"

  # Optional. If set, can scope down what this key can do.
  # inline_policy_document = jsonencode({
  #   version = "1.0"
  #   permissions = [
  #     {
  #       effect = "allow"
  #       scope  = "management"
  #       actions = ["read", "list"]
  #       resources = [
  #         {
  #           region       = "global"
  #           service      = "iam"
  #           resourceType = "*"
  #           resourceId   = "*"
  #         }
  #       ]
  #     }
  #   ]
  # })
}

# Account policy managed by Terraform.
resource "vidos_iam_policy" "account" {
  name = "terraform-example-policy"

  document = jsonencode({
    version = "1.0"
    permissions = [
      {
        effect  = "allow"
        scope   = "management"
        actions = ["read", "list"]
        resources = [
          {
            region       = "global"
            service      = "iam"
            resourceType = "*"
            resourceId   = "*"
          }
        ]
      }
    ]
  })
}

# Attach account policy to the API key.
resource "vidos_iam_api_key_policy_attachment" "account" {
  api_key_id  = vidos_iam_api_key.example.resource_id
  policy_type = "account"
  policy_id   = vidos_iam_policy.account.resource_id
}

# Attach an existing MANAGED policy to the API key (policy must already exist).
resource "vidos_iam_api_key_policy_attachment" "managed" {
  api_key_id  = vidos_iam_api_key.example.resource_id
  policy_type = "MANAGED"
  policy_id   = var.managed_policy_id
}

# Optional: account-owned service role.
resource "vidos_iam_service_role" "example" {
  name = "terraform-example-role"
}

# Attach the account policy to the service role.
resource "vidos_iam_service_role_policy_attachment" "role_account" {
  service_role_id = vidos_iam_service_role.example.resource_id
  policy_type     = "account"
  policy_id       = vidos_iam_policy.account.resource_id
}

# Attach an existing MANAGED policy to the service role.
resource "vidos_iam_service_role_policy_attachment" "role_managed" {
  service_role_id = vidos_iam_service_role.example.resource_id
  policy_type     = "MANAGED"
  policy_id       = var.managed_policy_id
}
