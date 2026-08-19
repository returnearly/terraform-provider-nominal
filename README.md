# terraform-provider-nominal

Terraform provider for [Nominal](https://github.com/returnearly/nominal). Registry address: `returnearly/nominal`. This GitHub repo is named `terraform-provider-nominal` because HashiCorp requires that prefix.

Talks to Nominal over **GraphQL only**. HTTP 200 with `errors[]` is a failed apply.

## Example

```hcl
terraform {
  required_providers {
    nominal = {
      source = "returnearly/nominal"
    }
  }
}

provider "nominal" {
  endpoint = "https://nominal.example.com/graphql"
  token    = var.nominal_token
}

resource "nominal_notification_channel" "ops" {
  name = "Ops webhook"
  type = "Webhook"

  config {
    key   = "url"
    value = "https://example.com/hooks/nominal"
  }
}

resource "nominal_monitor" "api" {
  name             = "API health"
  type             = "Http"
  target           = "https://example.com/health"
  method           = "GET"
  conditions       = ["[STATUS] == 200"]
  channel_ids      = [nominal_notification_channel.ops.id]
  retention_days   = 30
}
```

Build locally:

```bash
go build -o terraform-provider-nominal
```
