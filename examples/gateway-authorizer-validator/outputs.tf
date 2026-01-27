output "validator_resource_id" {
  description = "Validator instance resource ID."
  value       = vidos_validator_instance.main.resource_id
}

output "authorizer_resource_id" {
  description = "Authorizer instance resource ID."
  value       = vidos_authorizer_instance.main.resource_id
}

output "gateway_resource_id" {
  description = "Gateway instance resource ID."
  value       = vidos_gateway_instance.main.resource_id
}

output "gateway_endpoint" {
  description = "Gateway instance endpoint."
  value       = vidos_gateway_instance.main.endpoint
}
