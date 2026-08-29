---
page_title: "nominal_monitors Data Source - terraform-provider-nominal"
description: |-
  List Nominal monitors, optionally filtered by tag.
---

# nominal_monitors (Data Source)

Lists monitors, optionally filtered by a single tag. Useful for wiring existing monitors into a status page or maintenance window without importing them.

The filter is Nominal's `monitors(tag:)` query (`whereJsonContains` on `tags`).

## Example Usage

```terraform
data "nominal_monitors" "prod" {
  tag = "prod"
}

resource "nominal_maintenance_window" "fleet" {
  title       = "Production window"
  starts_at   = "2026-08-21 02:00:00"
  ends_at     = "2026-08-21 04:00:00"
  monitor_ids = data.nominal_monitors.prod.monitors[*].id
}
```

All monitors:

```terraform
data "nominal_monitors" "all" {}

output "down" {
  value = [
    for monitor in data.nominal_monitors.all.monitors : monitor.name
    if monitor.status == "Down"
  ]
}
```

## Schema

### Optional

- `tag` (String) When set, only monitors with this tag are returned.

### Read-Only

- `monitors` (Attributes List) Matching monitors. See [nested schema](#nested-schema-for-monitors) below.

### Nested Schema for `monitors`

Read-Only:

- `enabled` (Boolean)
- `id` (String)
- `name` (String)
- `status` (String) Effective status: `Pending`, `Up`, `Down`, `Paused`, or `Maintenance`.
- `tags` (List of String)
- `target` (String)
- `type` (String)
