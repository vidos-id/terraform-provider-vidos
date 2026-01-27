# Quickstart: Instance Endpoint Output

## Output an instance endpoint

```hcl
resource "vidos_verifier_instance" "example" {
  name = "example"
  # ... other required arguments ...
}

output "verifier_endpoint" {
  value = vidos_verifier_instance.example.endpoint
}

# All instance-like resources expose `endpoint` similarly.
resource "vidos_gateway_instance" "example" {
  name = "example"
  # ... other required arguments ...
}

output "gateway_endpoint" {
  value = vidos_gateway_instance.example.endpoint
}
```
