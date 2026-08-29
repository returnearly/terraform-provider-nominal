terraform {
  required_providers {
    nominal = {
      source  = "returnearly/nominal"
      version = ">= 0.1.0"
    }
  }
}

provider "nominal" {
  endpoint = "https://nominal.example.com/graphql"
  token    = var.nominal_token

  headers = {
    CF-Access-Client-Id     = var.cf_access_client_id
    CF-Access-Client-Secret = var.cf_access_client_secret
  }
}
