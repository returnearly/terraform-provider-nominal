---
page_title: "nominal_probe Data Source - terraform-provider-nominal"
description: |-
  Look up a Nominal probe by slug or id. Probes are created by Nominal itself; this data source is read-only.
---

# nominal_probe (Data Source)

Look up one probe by `slug` or `id`. Probes are created by Nominal; this provider cannot create them.

Set **either** `slug` or `id`. If both are set, `id` is matched first.

Use the resulting `id` in [`nominal_monitor.probe_ids`](../resources/monitor.md). When `is_default` is true, new monitors attach this probe unless `probe_ids` is set.

## Example Usage

```terraform
data "nominal_probe" "local" {
  slug = "local"
}

resource "nominal_monitor" "api" {
  name      = "API health"
  type      = "Http"
  target    = "https://example.com/health"
  probe_ids = [data.nominal_probe.local.id]
}
```

```terraform
data "nominal_probe" "by_id" {
  id = "01HZXEXAMPLE"
}
```

## Schema

### Optional

- `id` (String) Probe ID. Computed when you look up by `slug`.
- `slug` (String) Probe slug such as `local` or `us-east`. Computed when you look up by `id`.

### Read-Only

- `enabled` (Boolean) Whether the probe is accepting work.
- `is_default` (Boolean) When true, new monitors attach this probe unless `probe_ids` is set.
- `name` (String) Display name.
- `queue` (String) Queue name the probe consumes, for example `checks.local`.

## Errors

| Situation | Message |
| --- | --- |
| Neither `slug` nor `id` | `Missing probe lookup` — `Set slug or id.` |
| No match | `Probe not found` |
