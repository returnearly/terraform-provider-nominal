package provider

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProviderHeadersIgnoresReserved(t *testing.T) {
	t.Parallel()

	raw, diags := types.MapValueFrom(context.Background(), types.StringType, map[string]string{
		"CF-Access-Client-Id": "client-id",
		"Authorization":       "Bearer hijack",
		"Accept":              "*/*",
	})
	if diags.HasError() {
		t.Fatalf("map: %v", diags)
	}

	headers, reserved := providerHeaders(context.Background(), raw)
	if headers["CF-Access-Client-Id"] != "client-id" || len(headers) != 1 {
		t.Fatalf("headers: %#v", headers)
	}

	if !reflect.DeepEqual(reserved, []string{"Accept", "Authorization"}) {
		t.Fatalf("reserved: %#v", reserved)
	}
}

func TestProviderHeadersNull(t *testing.T) {
	t.Parallel()

	headers, reserved := providerHeaders(context.Background(), types.MapNull(types.StringType))
	if headers != nil || reserved != nil {
		t.Fatalf("headers: %#v reserved: %#v", headers, reserved)
	}
}
