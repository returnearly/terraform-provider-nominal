# terraform-provider-nominal

Terraform provider for [Nominal](https://github.com/returnearly/nominal). Registry address: `returnearly/nominal`. This GitHub repo is named `terraform-provider-nominal` because HashiCorp requires that prefix.

Talks to Nominal over **GraphQL only**. HTTP 200 with `errors[]` is a failed apply.

Resources: `nominal_monitor`, `nominal_notification_channel`, `nominal_status_page`, `nominal_maintenance_window`.

Data sources: `nominal_probe`, `nominal_probes`, `nominal_monitors`.

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

data "nominal_probe" "local" {
  slug = "local"
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
  name        = "API health"
  description = "Public HTTPS check. Page payments if this is down."
  tags        = ["prod", "critical"]
  type        = "Http"
  target      = "https://example.com/health"
  method      = "GET"
  proxy_url   = "socks5h://127.0.0.1:1080"
  conditions  = ["[STATUS] == 200"]
  probe_ids   = [data.nominal_probe.local.id]
  channel_ids = [nominal_notification_channel.ops.id]
}

resource "nominal_monitor" "resolver" {
  name           = "example.com A"
  type           = "Dns"
  target         = "1.1.1.1"
  dns_query_name = "example.com"
  dns_query_type = "A"
  conditions     = ["[DNS_RCODE] == NOERROR"]
}

resource "nominal_monitor" "backup" {
  name   = "Nightly backup"
  type   = "Heartbeat"
  target = "backup-job"
}

resource "nominal_status_page" "public" {
  name      = "Acme Status"
  slug      = "acme"
  published = true
  theme     = "Dark"

  monitor {
    monitor_id  = nominal_monitor.api.id
    public_name = "API"
  }
}

resource "nominal_maintenance_window" "db" {
  title       = "Database upgrade"
  message     = "Upgrading Postgres."
  starts_at   = "2026-08-21 02:00:00"
  ends_at     = "2026-08-21 04:00:00"
  monitor_ids = [nominal_monitor.api.id]
}
```

Monitor types: `Http`, `GraphQL`, `Ping`, `Tcp`, `Dns`, `Tls`, `Heartbeat`, `Udp`, `WebSocket`, `Mysql`, `Redis`, `Postgres`.

Heartbeat monitors expose `heartbeat_url`, `heartbeat_start_url`, `heartbeat_finish_url`, and `heartbeat_error_url` after apply. Badge URLs and rolling uptime are computed on every monitor.

`group` is deprecated. Use `tags`. If `tags` is omitted, `group` is sent as a single tag so older configs still apply.

Maintenance window timestamps use Nominal's GraphQL DateTime format: `YYYY-MM-DD HH:MM:SS`.

Build locally:

```bash
go build -o terraform-provider-nominal
```

## Releasing

Push a SemVer tag on the commit you want published. GitHub Actions builds signed binaries and creates the GitHub Release. The Terraform Registry then ingresses [returnearly/nominal](https://registry.terraform.io/providers/returnearly/nominal).

```bash
git tag v0.1.0
git push origin v0.1.0
```

Do not create the GitHub Release in the UI first. The workflow owns that.

One-time Registry setup (after the first GitHub Release exists):

1. Add the public GPG key from the 1Password note **Terraform Registry GPG Key** at [Registry signing keys](https://registry.terraform.io/settings/gpg-keys).
2. [Publish → Provider](https://registry.terraform.io/publish/provider) and select `returnearly/terraform-provider-nominal`.

