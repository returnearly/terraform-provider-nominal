---
page_title: "nominal_maintenance_window Resource - terraform-provider-nominal"
description: |-
  A maintenance window that suppresses alerts. Dates use YYYY-MM-DD HH:MM:SS.
---

# nominal_maintenance_window (Resource)

A maintenance window that suppresses alerts for selected monitors (or every monitor). Covered monitors report status `Maintenance` while the window is active.

Timestamps use Nominal's GraphQL `DateTime` format: `YYYY-MM-DD HH:MM:SS`.

Set either `applies_to_all = true` or a non-empty `monitor_ids` list. `monitor_ids` is required unless `applies_to_all` is true.

## Example Usage

### Specific monitors

```terraform
resource "nominal_maintenance_window" "db" {
  title       = "Database upgrade"
  message     = "Upgrading Postgres."
  starts_at   = "2026-08-21 02:00:00"
  ends_at     = "2026-08-21 04:00:00"
  monitor_ids = [nominal_monitor.api.id]
}
```

### Every monitor, starting now

```terraform
resource "nominal_maintenance_window" "fleet" {
  title          = "Platform maintenance"
  message        = "All checks paused."
  applies_to_all = true
  ends_at        = "2026-08-22 06:00:00"
}
```

Omitting `starts_at` defaults the window start to now. Omitting `ends_at` leaves an open-ended window.

## Schema

### Required

- `title` (String) Window title.

### Optional

- `applies_to_all` (Boolean) When true, the window covers every monitor. Defaults to `false`. Otherwise set `monitor_ids`.
- `ends_at` (String) Window end. Omit for an open-ended window. Format: `YYYY-MM-DD HH:MM:SS`.
- `message` (String) Visitor- or operator-facing explanation.
- `monitor_ids` (List of String) Monitors covered by this window. Required unless `applies_to_all` is true.
- `starts_at` (String) Window start. Defaults to now when omitted. Format: `YYYY-MM-DD HH:MM:SS`.

### Read-Only

- `id` (String) Nominal maintenance window ID.
- `phase` (String) `scheduled`, `active`, or `ended`.

## Notes

- Delete calls `deleteMaintenanceWindow`. That removes the window; it is not the same as ending it early in the UI.
- `phase` is computed by Nominal from `starts_at`, `ends_at`, and the current time.

## Import

Import is supported using the following syntax:

```shell
terraform import nominal_maintenance_window.db {{id}}
```
