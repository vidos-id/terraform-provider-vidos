output "authorizer_configuration_resource_id" {
  description = "Authorizer configuration resource ID."
  value       = vidos_authorizer_configuration.example.resource_id
}

output "authorizer_resource_id" {
  description = "Authorizer instance resource ID."
  value       = vidos_authorizer_instance.main.resource_id
}

output "gateway_resource_id" {
  description = "Gateway instance resource ID."
  value       = vidos_gateway_instance.example.resource_id
}

output "gateway_endpoint" {
  description = "Gateway instance endpoint."
  value       = vidos_gateway_instance.example.endpoint
}
