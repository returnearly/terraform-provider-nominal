---
page_title: "nominal Provider"
description: |-
  Manage Nominal monitors, notification channels, status pages, and maintenance windows through GraphQL.
---

# nominal Provider

The [Nominal](https://github.com/returnearly/nominal) provider manages uptime monitors, notification channels, public status pages, and maintenance windows on a self-hosted Nominal instance.

The provider talks to Nominal over **GraphQL only**. Registry address: `returnearly/nominal`. Protocol version 6.0 (Terraform Plugin Framework).

~> GraphQL responses that include `errors[]` fail the apply even when HTTP status is 200. The client also fails on HTTP 4xx/5xx.

Use this provider to:

- Create and update HTTP, GraphQL, DNS, TLS, heartbeat, database, and other probe monitors
- Attach notification channels (email, Slack, Teams, Discord, webhook, PagerDuty)
- Publish branded status pages with listed monitors
- Schedule maintenance windows that suppress alerts

Incidents are part of the Nominal GraphQL API but are **not** managed by this provider.

## Example Usage

```terraform
terraform {
  required_providers {
    nominal = {
      source  = "returnearly/nominal"
      version = ">= 0.1.0"
    }
  }
}

provider "nominal" {
  endpoint = "https://nominal.example.com/graphql"
  token    = var.nominal_token

  # Optional. Needed when GraphQL sits behind Cloudflare Access or a similar proxy.
  headers = {
    CF-Access-Client-Id     = var.cf_access_client_id
    CF-Access-Client-Secret = var.cf_access_client_secret
  }
}
```

Token via environment variable (omit `token` in the provider block):

```bash
export NOMINAL_TOKEN="1|your-sanctum-token"
```

## Schema

### Required

- `endpoint` (String) GraphQL endpoint, for example `https://nominal.example.com/graphql`. A trailing slash is stripped.

### Optional

- `token` (String, Sensitive) Laravel Sanctum bearer token. When omitted, the provider reads `NOMINAL_TOKEN`. The provider sends `Authorization: Bearer <token>` on every request.
- `headers` (Map of String, Sensitive) Extra HTTP headers on every GraphQL request. Use this for [Cloudflare Access](guides/authentication.md#cloudflare-access) service tokens (`CF-Access-Client-Id`, `CF-Access-Client-Secret`) or similar proxies. `Authorization`, `Content-Type`, and `Accept` are reserved and ignored.

## Authentication

Create a token on the Nominal host, then pass it to Terraform:

```bash
php artisan nominal:token ops@example.com --name=terraform
```

See [Authentication](guides/authentication.md) for environment variables, CI, and common GraphQL errors.

## Resources

| Resource | Purpose |
| --- | --- |
| [`nominal_monitor`](resources/monitor.md) | Uptime or heartbeat check |
| [`nominal_notification_channel`](resources/notification_channel.md) | Alert destination |
| [`nominal_status_page`](resources/status_page.md) | Public status page |
| [`nominal_maintenance_window`](resources/maintenance_window.md) | Alert-suppressing window |

## Data Sources

| Data source | Purpose |
| --- | --- |
| [`nominal_probe`](data-sources/probe.md) | Look up one probe by slug or id |
| [`nominal_probes`](data-sources/probes.md) | List every probe |
| [`nominal_monitors`](data-sources/monitors.md) | List monitors, optionally filtered by tag |

Probes are created by Nominal itself. This provider only reads them.

## Guides

- [Authentication](guides/authentication.md) (including [Cloudflare Access](guides/authentication.md#cloudflare-access))
- [Monitor types](guides/monitor-types.md)
- [Conditions](guides/conditions.md)
- [Import](guides/import.md)

## Behavior notes

- HTTP client timeout is 30 seconds.
- `headers` are sent on every GraphQL request. `Authorization`, `Content-Type`, and `Accept` cannot be overridden.
- Monitor `channel_ids` are applied after create/update with `syncMonitorChannels`. Omitting `channel_ids` leaves existing attachments unchanged. Set `channel_ids = []` to detach every channel.
- Monitor `group` is deprecated. Prefer `tags`. If `tags` is omitted and `group` is set, the provider sends that value as a single tag.
- Maintenance window timestamps use Nominal's GraphQL `DateTime` format: `YYYY-MM-DD HH:MM:SS`.
- Status page `password` is write-only from Terraform's point of view: GraphQL never returns it. Omit the argument to leave the current password unchanged; set `password = ""` to clear it.
