package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/returnearly/terraform-provider-nominal/internal/client"
)

func TestListProbes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": map[string]any{
				"probes": []map[string]any{
					{
						"id":         "probe-local",
						"slug":       "local",
						"name":       "Local",
						"queue":      "checks.local",
						"enabled":    true,
						"is_default": true,
					},
				},
			},
		})
	}))
	defer server.Close()

	probes, err := listProbes(context.Background(), client.New(server.URL, "token"))
	if err != nil {
		t.Fatal(err)
	}

	if len(probes) != 1 || probes[0].Slug != "local" || !probes[0].IsDefault {
		t.Fatalf("probes: %#v", probes)
	}
}
