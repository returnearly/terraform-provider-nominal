---
page_title: "nominal_monitor Resource - terraform-provider-nominal"
description: |-
  A Nominal monitor. Types: Http, GraphQL, Ping, Tcp, Dns, Tls, Heartbeat, Udp, WebSocket, Mysql, Redis, Postgres.
---

# nominal_monitor (Resource)

A Nominal uptime or heartbeat monitor. After create or update the provider reads the monitor back, including computed badge URLs, heartbeat URLs, status, and rolling uptime.

Notification channels are attached with a separate `syncMonitorChannels` mutation after the monitor is saved. See `channel_ids`.

See [Monitor types](../guides/monitor-types.md) and [Conditions](../guides/conditions.md) for type-specific fields and expressions.

## Example Usage

### HTTP check

```terraform
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

  request_headers {
    key   = "X-Health-Token"
    value = var.health_token
  }
}
```

### DNS

```terraform
resource "nominal_monitor" "resolver" {
  name           = "example.com A"
  type           = "Dns"
  target         = "1.1.1.1"
  dns_query_name = "example.com"
  dns_query_type = "A"
  conditions     = ["[DNS_RCODE] == NOERROR"]
}
```

### Heartbeat

```terraform
resource "nominal_monitor" "backup" {
  name   = "Nightly backup"
  type   = "Heartbeat"
  target = "backup-job"
}

output "backup_heartbeat_url" {
  value     = nominal_monitor.backup.heartbeat_url
  sensitive = false
}
```

### PostgreSQL

```terraform
resource "nominal_monitor" "db" {
  name         = "Primary Postgres"
  type         = "Postgres"
  target       = "postgres://monitor:${var.db_password}@db.example.com:5432/app"
  request_body = "SELECT 1"
  conditions   = ["[CONNECTED] == true"]
}
```

## Schema

### Required

- `name` (String) Display name.
- `target` (String) URL, host, or connection URL. Database monitors take `mysql://`, `postgres://`, or `redis://` URLs. See [Monitor types](../guides/monitor-types.md).
- `type` (String) `Http`, `GraphQL`, `Ping`, `Tcp`, `Dns`, `Tls`, `Heartbeat`, `Udp`, `WebSocket`, `Mysql`, `Redis`, or `Postgres`.

### Optional

- `channel_ids` (List of String) Notification channel IDs. Synced via `syncMonitorChannels` after create and update. Omit to leave attachments unchanged. Set `[]` to detach every channel.
- `conditions` (List of String) Gatus-style expressions such as `[STATUS] == 200`. Omitted conditions use the type defaults. See [Conditions](../guides/conditions.md).
- `description` (String) Longer notes shown in the Nominal UI. Omitting the argument on update sends `null` and clears the stored description.
- `dns_query_name` (String) Name to resolve for `Dns` monitors.
- `dns_query_type` (String) `A`, `AAAA`, `CNAME`, `MX`, `NS`, `PTR`, `SRV`, or `TXT`. `Dns` monitors default to `A`.
- `enabled` (Boolean) Whether the monitor is scheduled. Defaults to `true`.
- `follow_redirects` (Boolean) Follow HTTP redirects. Defaults to `true`. Applies to `Http` and `GraphQL`.
- `group` (String, Deprecated) Removed from the Nominal API. Set `tags` instead. If `tags` is omitted, `group` is sent as a single tag so older configs still apply. `group` is never sent as its own GraphQL field.
- `interval_seconds` (Number) Check interval, or expected heartbeat period. Defaults to `60`. Monitors with a `[DOMAIN_EXPIRATION]` condition must use at least `300`.
- `ip_family` (String) `Ipv4`, `Ipv6`, or `Any`. Defaults to `Any`.
- `method` (String) HTTP method: `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, or `HEAD`. GraphQL monitors default to `POST`.
- `probe_ids` (List of String) Probe IDs that run this monitor. Omitted IDs attach Nominal's default probes (`is_default = true`). Heartbeat monitors have no probes.
- `proxy_url` (String, Sensitive) HTTP or SOCKS proxy URL for `Http`, `GraphQL`, `Tcp`, `Tls`, `WebSocket`, and `Redis`.
- `request_body` (String) HTTP body, GraphQL query, UDP/WebSocket payload, custom SQL, or Redis command.
- `request_headers` (Block List) HTTP or WebSocket request headers. See [nested schema](#nested-schema-for-request_headers) below.
- `retention_days` (Number) How long check results are kept. Defaults to `30`.
- `tags` (List of String) Labels used for grouping and for [`nominal_monitors`](../data-sources/monitors.md). Nominal normalizes tags (trim, case-insensitive unique, max 32 tags, 64 characters each).
- `timeout_seconds` (Number) Probe timeout. Defaults to `10`. Unused for heartbeat monitors.
- `verify_tls` (Boolean) Verify TLS certificates. Defaults to `true`. Applies to `Http`, `GraphQL`, `Tls`, `WebSocket`, `Mysql`, `Redis`, and `Postgres`.

### Read-Only

- `id` (String) Nominal monitor ID.
- `badge_markdown` (String) Markdown snippet that embeds the status badge.
- `heartbeat_error_url` (String) URL to signal a failed heartbeat job.
- `heartbeat_finish_url` (String) URL to signal a successful heartbeat job.
- `heartbeat_start_url` (String) URL to signal that a heartbeat job started.
- `heartbeat_token` (String, Sensitive) Secret token embedded in heartbeat URLs.
- `heartbeat_url` (String) Default heartbeat finish URL (`/api/heartbeat/{token}`).
- `latency_badge_json_url` (String) JSON latency badge URL.
- `latency_badge_url` (String) SVG latency badge URL.
- `status` (String) Effective status: `Pending`, `Up`, `Down`, `Paused`, or `Maintenance`.
- `status_badge_json_url` (String) JSON status badge URL.
- `status_badge_url` (String) SVG status badge URL.
- `uptime` (Attributes) Rolling uptime percentages from check results. See [nested schema](#nested-schema-for-uptime) below.
- `uptime_badge_json_url` (String) JSON uptime badge URL.
- `uptime_badge_url` (String) SVG uptime badge URL.

### Nested Schema for `request_headers`

Required:

- `key` (String) Header name.
- `value` (String, Sensitive) Header value.

### Nested Schema for `uptime`

Read-Only:

- `one_hour` (Number)
- `twenty_four_hours` (Number)
- `seven_days` (Number)
- `thirty_days` (Number)

## Notes

- Create uses `createMonitor`, then `syncMonitorChannels` when `channel_ids` is set, then a `monitor(id:)` read.
- Update uses `updateMonitor` the same way. A missing monitor on refresh is removed from state.
- Heartbeat URLs and badge URLs are computed by Nominal and stored with `UseStateForUnknown` so they do not churn every plan.
- `uptime` is computed from check results and may be null for a brand-new monitor.

## Import

Import is supported using the following syntax:

```shell
terraform import nominal_monitor.api {{id}}
```
