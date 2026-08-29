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
}

resource "nominal_monitor" "resolver" {
  name           = "example.com A"
  type           = "Dns"
  target         = "1.1.1.1"
  dns_query_name = "example.com"
  dns_query_type = "A"
  conditions     = ["[DNS_RCODE] == NOERROR"]
}

resource "nominal_monitor" "backup" {
  name   = "Nightly backup"
  type   = "Heartbeat"
  target = "backup-job"
}
