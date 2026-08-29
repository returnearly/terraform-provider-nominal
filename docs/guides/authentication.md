---
page_title: "Authentication"
description: |-
  How to authenticate the Nominal Terraform provider with a Sanctum bearer token.
---

# Authentication

The provider authenticates every GraphQL request with a Laravel Sanctum bearer token.

## Provider configuration

```terraform
provider "nominal" {
  endpoint = "https://nominal.example.com/graphql"
  token    = var.nominal_token
}
```

`endpoint` is required. `token` is optional when `NOMINAL_TOKEN` is set.

Configuration in the provider block wins over the environment variable.

## Environment variable

```bash
export NOMINAL_TOKEN="1|your-sanctum-token"
```

```terraform
provider "nominal" {
  endpoint = "https://nominal.example.com/graphql"
}
```

This is the usual pattern for CI and for keeping secrets out of `.tf` files.

## Creating a token

On the Nominal host, create a Sanctum token for an existing user:

```bash
php artisan nominal:token ops@example.com --name=terraform
```

The command prints a plain-text token (`1|...`). Store it in a Terraform variable, an environment variable, or your secret manager. Do not commit it.

## HTTP headers

`headers` is an optional map of extra HTTP headers sent on every GraphQL request. Values are sensitive in plans and state.

This is the usual way to get through an identity proxy in front of Nominal, such as Cloudflare Access.

```terraform
provider "nominal" {
  endpoint = "https://nominal.example.com/graphql"
  token    = var.nominal_token

  headers = {
    CF-Access-Client-Id     = var.cf_access_client_id
    CF-Access-Client-Secret = var.cf_access_client_secret
  }
}
```

The provider always sets these headers itself. Entries with the same name in `headers` are ignored, and Terraform prints a warning:

- `Authorization`
- `Content-Type`
- `Accept`

Empty header names are dropped.

## Cloudflare Access

When the GraphQL endpoint is behind [Cloudflare Access](https://developers.cloudflare.com/cloudflare-one/identity/service-tokens/), create a service token and pass both values:

1. Cloudflare Zero Trust → Access → Service Auth → Service Tokens → Create.
2. Copy the Client ID and Client Secret.
3. Allow that service token on the Access application that protects `/graphql` (and typically the rest of the Nominal origin).

```terraform
variable "cf_access_client_id" {
  type      = string
  sensitive = true
}

variable "cf_access_client_secret" {
  type      = string
  sensitive = true
}

provider "nominal" {
  endpoint = "https://nominal.example.com/graphql"
  token    = var.nominal_token

  headers = {
    CF-Access-Client-Id     = var.cf_access_client_id
    CF-Access-Client-Secret = var.cf_access_client_secret
  }
}
```

The Sanctum `token` is still required. Access headers get the request past Cloudflare; Sanctum authenticates to Nominal.

Other proxies that expect static request headers work the same way. Use the header names that proxy documents.

## Request details

Each request is `POST` to `endpoint` with:

- `Authorization: Bearer <token>`
- `Content-Type: application/json`
- `Accept: application/json`
- Any extra keys from `headers`

The GraphQL document and variables are the JSON body. Trailing slashes on `endpoint` are stripped.

## Errors

The provider treats these as apply failures:

| Situation | Typical message |
| --- | --- |
| Missing `endpoint` | `Missing endpoint` |
| Missing token and `NOMINAL_TOKEN` | `Missing token` |
| HTTP 200 with GraphQL `errors[]` | `nominal graphql: <messages>` |
| HTTP 4xx/5xx | `nominal graphql http <status>` |
| HTTP 200 HTML (Access login) | `nominal graphql: invalid JSON` or a Cloudflare challenge body |
| Non-JSON body | `nominal graphql: invalid JSON` |

A common GraphQL error is `Unauthenticated.` — check the token and that `endpoint` is the `/graphql` URL, not the UI origin.

If the response looks like a Cloudflare Access login page instead of GraphQL JSON, `headers` is missing or the service token is not allowed on that application.

The HTTP client times out after 30 seconds.
