resource "nominal_status_page" "public" {
  name      = "Acme Status"
  slug      = "acme"
  published = true
  theme     = "Dark"

  monitor {
    monitor_id  = nominal_monitor.api.id
    public_name = "API"
  }
}
