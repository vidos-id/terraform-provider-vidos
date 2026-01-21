variable "vidos_api_key" {
  type        = string
  sensitive   = true
  description = "Vidos IAM API secret (64 hex)."
}

variable "vidos_region" {
  type        = string
  description = "Region for resolver management endpoint (e.g. eu)."
  default     = "eu"
}
