---
page_title: "nominal_notification_channel Resource - terraform-provider-nominal"
description: |-
  A Nominal notification channel (Mail, Slack, MicrosoftTeams, Discord, Webhook, or Pagerduty).
---

# nominal_notification_channel (Resource)

An alert destination. Attach channels to monitors with [`nominal_monitor.channel_ids`](monitor.md).

`config` is a list of key/value pairs. Nominal validates and normalizes keys per `type`. Unknown keys are dropped. Values are sensitive in Terraform state.

## Example Usage

### Webhook

```terraform
resource "nominal_notification_channel" "ops" {
  name = "Ops webhook"
  type = "Webhook"

  config {
    key   = "url"
    value = "https://example.com/hooks/nominal"
  }
}
```

### Email

```terraform
resource "nominal_notification_channel" "email" {
  name = "On-call inbox"
  type = "Mail"

  config {
    key   = "to"
    value = "ops@example.com"
  }
}
```

### Slack

```terraform
resource "nominal_notification_channel" "slack" {
  name = "Slack #alerts"
  type = "Slack"

  config {
    key   = "webhook_url"
    value = var.slack_webhook_url
  }
}
```

### PagerDuty

```terraform
resource "nominal_notification_channel" "pagerduty" {
  name = "PagerDuty"
  type = "Pagerduty"

  config {
    key   = "routing_key"
    value = var.pagerduty_routing_key
  }
}
```

## Schema

### Required

- `name` (String) Display name.
- `type` (String) `Mail`, `Slack`, `MicrosoftTeams`, `Discord`, `Webhook`, or `Pagerduty`.

### Optional

- `config` (Block List) Channel settings as key/value pairs. See [nested schema](#nested-schema-for-config) below and [config keys](#config-keys).

### Read-Only

- `id` (String) Nominal notification channel ID.

### Nested Schema for `config`

Required:

- `key` (String) Canonical key or alias (`url`, `webhook_url`, `to`, `routing_key`, `integration_key`).
- `value` (String, Sensitive) Setting value.

## Config keys

Nominal stores one canonical key per type. Aliases are accepted and rewritten.

| Type | Canonical key | Aliases | Value |
| --- | --- | --- | --- |
| `Mail` | `to` | — | One recipient email. Uses the Nominal app mailer (`MAIL_*`). |
| `Slack` | `webhook_url` | `url` | Incoming webhook URL bound to a Slack channel. |
| `MicrosoftTeams` | `webhook_url` | `url` | Incoming webhook or Workflows URL. |
| `Discord` | `webhook_url` | `url` | Discord channel webhook. |
| `Webhook` | `url` | `webhook_url` | Endpoint that receives a JSON POST (`event`, `headline`, `monitor`, `result`). |
| `Pagerduty` | `routing_key` | `integration_key` | Events API v2 integration key. Recoveries send a resolve for the monitor. |

```terraform
# These two Slack configs are equivalent after Nominal normalizes them.
config {
  key   = "url"
  value = "https://hooks.slack.com/services/T/B/xxx"
}

config {
  key   = "webhook_url"
  value = "https://hooks.slack.com/services/T/B/xxx"
}
```

Missing or invalid values (bad email, non-URL webhook) fail the GraphQL mutation and therefore the apply.

## Import

Import is supported using the following syntax:

```shell
terraform import nominal_notification_channel.ops {{id}}
```
