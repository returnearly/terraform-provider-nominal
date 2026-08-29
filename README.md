# terraform-provider-nominal

Terraform provider for [Nominal](https://github.com/returnearly/nominal). Registry address: `returnearly/nominal`. This GitHub repo is named `terraform-provider-nominal` because HashiCorp requires that prefix.

Talks to Nominal over **GraphQL only**. HTTP 200 with `errors[]` is a failed apply.

## Documentation

Registry-style docs live in [`docs/`](docs/). After a release they publish on the [Terraform Registry](https://registry.terraform.io/providers/returnearly/nominal).

| Page | Description |
| --- | --- |
| [Provider](docs/index.md) | Endpoint, token, and overview |
| [Authentication](docs/guides/authentication.md) | Sanctum tokens, `NOMINAL_TOKEN`, and Cloudflare Access headers |
| [Monitor types](docs/guides/monitor-types.md) | Target and field matrix per type |
| [Conditions](docs/guides/conditions.md) | Gatus-style expressions and defaults |
| [Import](docs/guides/import.md) | Import existing resources by ID |
| [`nominal_monitor`](docs/resources/monitor.md) | Uptime and heartbeat monitors |
| [`nominal_notification_channel`](docs/resources/notification_channel.md) | Mail, Slack, Teams, Discord, webhook, PagerDuty |
| [`nominal_status_page`](docs/resources/status_page.md) | Public status pages |
| [`nominal_maintenance_window`](docs/resources/maintenance_window.md) | Alert-suppressing windows |
| [`nominal_probe`](docs/data-sources/probe.md) | Look up one probe |
| [`nominal_probes`](docs/data-sources/probes.md) | List probes |
| [`nominal_monitors`](docs/data-sources/monitors.md) | List monitors by tag |

Runnable snippets are under [`examples/`](examples/).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.0
- A Nominal instance with GraphQL enabled
- A Sanctum API token (`php artisan nominal:token <email> --name=terraform`)

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

  headers = {
    CF-Access-Client-Id     = var.cf_access_client_id
    CF-Access-Client-Secret = var.cf_access_client_secret
  }
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

## Development

```bash
go build -o terraform-provider-nominal
go test ./...
```

For local Terraform, install the binary into `~/.terraform.d/plugins/registry.terraform.io/returnearly/nominal/<version>/<os>_<arch>/`.

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
