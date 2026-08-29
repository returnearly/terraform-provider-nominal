---
page_title: "Import"
description: |-
  Import existing Nominal resources into Terraform state.
---

# Import

Every managed resource imports by Nominal GraphQL `id`. The import ID is passed through to the `id` attribute; Terraform then refreshes the rest of the schema from the API.

```bash
terraform import nominal_monitor.api 01HZX...
terraform import nominal_notification_channel.ops 01HZY...
terraform import nominal_status_page.public 01HZZ...
terraform import nominal_maintenance_window.db 01J00...
```

After import, run `terraform plan` and reconcile any drift:

- Status page `password` is never returned. After import it is unset in state. Omit `password` unless you intend to change it.
- Monitor `group` is not stored by the API. Imported monitors keep `group` empty; move the value into `tags` if you still have it in config.
- Monitor `channel_ids` and `probe_ids` are read from current attachments.
- Computed badge URLs, heartbeat URLs, uptime, `status`, `phase`, `path_url`, `public_url`, and `health` are filled on refresh.

There is no data-source import. Look up probes with [`nominal_probe`](../data-sources/probe.md) and existing monitors with [`nominal_monitors`](../data-sources/monitors.md).
