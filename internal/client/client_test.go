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
