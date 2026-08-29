resource "nominal_maintenance_window" "db" {
  title       = "Database upgrade"
  message     = "Upgrading Postgres."
  starts_at   = "2026-08-21 02:00:00"
  ends_at     = "2026-08-21 04:00:00"
  monitor_ids = [nominal_monitor.api.id]
}
