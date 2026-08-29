---
page_title: "nominal_status_page Resource - terraform-provider-nominal"
description: |-
  A public Nominal status page, including listed monitors, branding, and an optional password.
---

# nominal_status_page (Resource)

A public status page with branding, listed monitors, and an optional visitor password.

The page is served at `/status/{slug}` (`path_url`). When `custom_domain` is set, Nominal also serves the page at `/` on that hostname (`public_url`). Scheme prefixes on `custom_domain` are stripped by the API.

`monitor` blocks are listed in order. `public_name` is the visitor-facing label; the real target stays hidden unless `show_targets` is true.

## Example Usage

```terraform
resource "nominal_status_page" "public" {
  name      = "Acme Status"
  slug      = "acme"
  published = true
  theme     = "Dark"
  headline  = "Acme service status"
  logo_url  = "https://example.com/logo.svg"

  monitor {
    monitor_id  = nominal_monitor.api.id
    public_name = "API"
  }

  monitor {
    monitor_id  = nominal_monitor.web.id
    public_name = "Website"
  }
}
```

### Custom domain and password

```terraform
resource "nominal_status_page" "internal" {
  name           = "Internal Status"
  slug           = "internal"
  custom_domain  = "status.acme.internal"
  published      = true
  password       = var.status_page_password
  show_targets   = false
  refresh_seconds = 15

  monitor {
    monitor_id  = nominal_monitor.api.id
    public_name = "Payments API"
  }
}
```

## Schema

### Required

- `name` (String) Display name.
- `slug` (String) Path slug served at `/status/{slug}`.

### Optional

- `custom_css` (String) Extra CSS injected into the public page.
- `custom_domain` (String) Hostname for the page at `/`. Scheme prefixes are stripped by the API. Set to null/omit on update to clear it when Terraform sends an explicit empty value.
- `description` (String) Page description.
- `favicon_url` (String) Favicon URL.
- `footer_text` (String) Footer copy.
- `headline` (String) Hero headline.
- `logo_url` (String) Logo URL.
- `monitor` (Block List) Monitors listed on the page, in order. See [nested schema](#nested-schema-for-monitor) below.
- `password` (String, Sensitive) Visitor password. Never returned by GraphQL. Omit to leave the current password unchanged on update. Set `""` to clear it.
- `published` (Boolean) Whether the page is publicly reachable. Defaults to `false`.
- `refresh_seconds` (Number) Client refresh interval. Defaults to `30`.
- `show_targets` (Boolean) When true, the public page shows monitor targets. Defaults to `false`.
- `theme` (String) `Dark` or `Light`. Defaults to `Dark`.

### Read-Only

- `id` (String) Nominal status page ID.
- `health` (String) Aggregate health: `Operational`, `Degraded`, `PartialOutage`, `MajorOutage`, or `Maintenance`.
- `password_protected` (Boolean) Whether a visitor password is currently set.
- `path_url` (String) Path on the Nominal origin, typically `/status/{slug}`.
- `public_url` (String) Preferred public URL (custom domain when set, otherwise `path_url`).

### Nested Schema for `monitor`

Required:

- `monitor_id` (String) ID of a [`nominal_monitor`](monitor.md) (or any existing monitor ID).

Optional:

- `public_name` (String) Public label. Targets stay hidden unless `show_targets` is true.

## Password behavior

GraphQL never returns the password. Terraform keeps the configured `password` in state from the last plan and refreshes `password_protected` from the API.

| Config | Effect |
| --- | --- |
| `password` omitted | Create: no password. Update: leave the existing password as-is. |
| `password = "secret"` | Set or replace the password. |
| `password = ""` | Clear the password. |

After import, `password` is unset. Do not add `password` to config unless you intend to rotate it.

Incidents on a status page are not managed by this provider.

## Import

Import is supported using the following syntax:

```shell
terraform import nominal_status_page.public {{id}}
```
