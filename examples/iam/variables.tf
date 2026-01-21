variable "vidos_api_key" {
  type        = string
  sensitive   = true
  description = "Vidos IAM API secret (64 hex)."
}

variable "vidos_region" {
  type        = string
  description = "Default region for non-IAM management endpoints."
  default     = "eu"
}

variable "managed_policy_id" {
  type        = string
  description = "Resource ID of an existing MANAGED policy to attach."
}
