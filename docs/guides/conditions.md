---
page_title: "Conditions"
description: |-
  Gatus-style condition expressions used by nominal_monitor.
---

# Conditions

`nominal_monitor.conditions` is a list of Gatus-style expressions. Each expression is `PLACEHOLDER COMPARATOR VALUE`. When you omit `conditions`, Nominal stores the type defaults.

## Syntax

```text
[STATUS] == 200
[STATUS] >= 200
[BODY].status == ok
[RESPONSE_TIME] < 500
[CERTIFICATE_EXPIRATION] > 48h
[DOMAIN_EXPIRATION] > 720h
[DNS_RCODE] == NOERROR
[CONNECTED] == true
```

Comparators: `==`, `!=`, `<`, `<=`, `>`, `>=`.

`[BODY]` can take a JSON path: `[BODY].foo`, `[BODY][0].id`. Other placeholders do not use a path.

Durations for `[CERTIFICATE_EXPIRATION]` and `[DOMAIN_EXPIRATION]` use values such as `48h` or `720h`.

## Placeholders by type

| Placeholder | HTTP / GraphQL | Ping | TCP / UDP / WebSocket / DB | DNS | TLS | Heartbeat |
| --- | --- | --- | --- | --- | --- | --- |
| `[STATUS]` | yes | — | — | — | — | — |
| `[BODY]` | yes | — | yes | yes | yes | — |
| `[REDIRECT]` | yes | — | — | — | — | — |
| `[CONNECTED]` | yes | yes | yes | yes | yes | — |
| `[RESPONSE_TIME]` | yes | yes | yes | yes | yes | — |
| `[IP]` | yes | yes | yes | yes | yes | — |
| `[CERTIFICATE_EXPIRATION]` | yes | — | — | — | yes | — |
| `[DOMAIN_EXPIRATION]` | yes | yes | yes | — | yes | — |
| `[DNS_RCODE]` | — | — | — | yes | — | — |

Heartbeat monitors have no conditions. Nominal deletes any that are sent.

`[CONNECTED]`, `[IP]`, `[DNS_RCODE]`, and `[REDIRECT]` only support `==` and `!=`.

## Type defaults

| Type | Default expressions |
| --- | --- |
| `Http`, `GraphQL` | `[STATUS] >= 200`, `[STATUS] <= 299` |
| `Ping`, `Tcp`, `Udp`, `WebSocket`, `Mysql`, `Redis`, `Postgres` | `[CONNECTED] == true`, `[RESPONSE_TIME] < 50` |
| `Tls` | `[CONNECTED] == true`, `[CERTIFICATE_EXPIRATION] > 48h` |
| `Dns` | `[DNS_RCODE] == NOERROR` |
| `Heartbeat` | none |

`[RESPONSE_TIME]` is measured in milliseconds.

## Domain expiration

`[DOMAIN_EXPIRATION]` looks up WHOIS/RDAP for the target hostname. It is not valid on `Dns` or `Heartbeat` monitors.

Monitors that include a `[DOMAIN_EXPIRATION]` condition **must** use `interval_seconds` of at least `300`. The API rejects shorter intervals.

```terraform
resource "nominal_monitor" "site" {
  name             = "Marketing site"
  type             = "Http"
  target           = "https://example.com"
  interval_seconds = 300
  conditions = [
    "[STATUS] == 200",
    "[DOMAIN_EXPIRATION] > 720h",
  ]
}
```

## Examples

```terraform
# HTTP status and JSON body
conditions = [
  "[STATUS] == 200",
  "[BODY].ok == true",
]

# TLS handshake plus certificate lifetime
conditions = [
  "[CONNECTED] == true",
  "[CERTIFICATE_EXPIRATION] > 168h",
]

# DNS answer
conditions = [
  "[DNS_RCODE] == NOERROR",
  "[BODY] == 93.184.216.34",
]
```
