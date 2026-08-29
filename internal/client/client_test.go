package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestQueryMapsGraphQLErrorsOnHTTP200(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing bearer token")
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"data":null,"errors":[{"message":"Unauthenticated."}]}`))
	}))
	defer server.Close()

	client := New(server.URL, "test-token")
	var out json.RawMessage
	err := client.Query(context.Background(), "{ monitors { id } }", nil, &out)

	if err == nil {
		t.Fatal("expected GraphQL errors to fail the client")
	}

	if err.Error() != "nominal graphql: Unauthenticated." {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestQuerySendsCustomHeadersAndKeepsReserved(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("CF-Access-Client-Id") != "client-id" {
			t.Errorf("CF-Access-Client-Id: %q", r.Header.Get("CF-Access-Client-Id"))
		}
		if r.Header.Get("CF-Access-Client-Secret") != "client-secret" {
			t.Errorf("CF-Access-Client-Secret: %q", r.Header.Get("CF-Access-Client-Secret"))
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Authorization: %q", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type: %q", r.Header.Get("Content-Type"))
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"ok":true}}`))
	}))
	defer server.Close()

	api := New(server.URL, "test-token").WithHeaders(map[string]string{
		"CF-Access-Client-Id":     "client-id",
		"CF-Access-Client-Secret": "client-secret",
		"Authorization":           "Bearer hijack",
		"Content-Type":            "text/plain",
	})

	if err := api.Query(context.Background(), "{ __typename }", nil, nil); err != nil {
		t.Fatal(err)
	}
}

func TestSanitizeHeadersDropsReservedAndEmptyKeys(t *testing.T) {
	t.Parallel()

	got := SanitizeHeaders(map[string]string{
		"CF-Access-Client-Id": "client-id",
		"Authorization":       "Bearer hijack",
		"content-type":        "text/plain",
		"Accept":              "*/*",
		"":                    "nope",
	})

	if len(got) != 1 || got["CF-Access-Client-Id"] != "client-id" {
		t.Fatalf("headers: %#v", got)
	}

	if !ReservedHeader("Authorization") || !ReservedHeader("CONTENT-TYPE") || ReservedHeader("CF-Access-Client-Id") {
		t.Fatal("reserved header matching is wrong")
	}
}
