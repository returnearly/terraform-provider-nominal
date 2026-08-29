package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type Client struct {
	endpoint   string
	token      string
	headers    map[string]string
	httpClient *http.Client
}

type gqlError struct {
	Message string `json:"message"`
}

type gqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []gqlError      `json:"errors"`
}

func New(endpoint, token string) *Client {
	return &Client{
		endpoint: strings.TrimRight(endpoint, "/"),
		token:    token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// WithHeaders copies extra HTTP headers onto every GraphQL request.
// Authorization, Content-Type, and Accept are reserved and ignored.
func (c *Client) WithHeaders(headers map[string]string) *Client {
	if len(headers) == 0 {
		return c
	}

	c.headers = SanitizeHeaders(headers)
	return c
}

func SanitizeHeaders(headers map[string]string) map[string]string {
	if len(headers) == 0 {
		return nil
	}

	out := make(map[string]string, len(headers))
	for key, value := range headers {
		if key == "" || ReservedHeader(key) {
			continue
		}

		out[key] = value
	}

	if len(out) == 0 {
		return nil
	}

	return out
}

func ReservedHeader(key string) bool {
	switch strings.ToLower(key) {
	case "authorization", "content-type", "accept":
		return true
	default:
		return false
	}
}

func (c *Client) Query(ctx context.Context, query string, variables map[string]any, out any) error {
	payload, err := json.Marshal(map[string]any{
		"query":     query,
		"variables": variables,
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var parsed gqlResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		if resp.StatusCode >= 400 {
			return fmt.Errorf("nominal graphql http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		return fmt.Errorf("nominal graphql: invalid JSON: %w", err)
	}

	if len(parsed.Errors) > 0 {
		messages := make([]string, 0, len(parsed.Errors))
		for _, item := range parsed.Errors {
			messages = append(messages, item.Message)
		}

		return fmt.Errorf("nominal graphql: %s", strings.Join(messages, "; "))
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("nominal graphql http %d", resp.StatusCode)
	}

	if out == nil || len(parsed.Data) == 0 || string(parsed.Data) == "null" {
		return nil
	}

	return json.Unmarshal(parsed.Data, out)
}
