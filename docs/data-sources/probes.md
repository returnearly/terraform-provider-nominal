---
page_title: "nominal_probes Data Source - terraform-provider-nominal"
description: |-
  All Nominal probes.
---

# nominal_probes (Data Source)

Lists every probe on the Nominal instance. Use this when you want to attach several probes, or to find default probes without knowing slugs in advance.

There are no arguments.

## Example Usage

```terraform
data "nominal_probes" "all" {}

resource "nominal_monitor" "api" {
  name      = "API health"
  type      = "Http"
  target    = "https://example.com/health"
  probe_ids = [for probe in data.nominal_probes.all.probes : probe.id if probe.enabled]
}
```

Default probes only:

```terraform
data "nominal_probes" "all" {}

locals {
  default_probe_ids = [
    for probe in data.nominal_probes.all.probes : probe.id
    if probe.is_default && probe.enabled
  ]
}
```

## Schema

### Read-Only

- `probes` (Attributes List) Every probe. See [nested schema](#nested-schema-for-probes) below.

### Nested Schema for `probes`

Read-Only:

- `enabled` (Boolean)
- `id` (String)
- `is_default` (Boolean)
- `name` (String)
- `queue` (String)
- `slug` (String)
