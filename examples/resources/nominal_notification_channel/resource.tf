resource "nominal_notification_channel" "ops" {
  name = "Ops webhook"
  type = "Webhook"

  config {
    key   = "url"
    value = "https://example.com/hooks/nominal"
  }
}

resource "nominal_notification_channel" "email" {
  name = "On-call inbox"
  type = "Mail"

  config {
    key   = "to"
    value = "ops@example.com"
  }
}

resource "nominal_notification_channel" "pagerduty" {
  name = "PagerDuty"
  type = "Pagerduty"

  config {
    key   = "routing_key"
    value = var.pagerduty_routing_key
  }
}
