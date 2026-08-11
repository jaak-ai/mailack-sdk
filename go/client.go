package mailack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Client talks to one mailack API deployment on behalf of a customer.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithAPIKey sets the Bearer token (mlk_… secret shown once at key creation).
func WithAPIKey(key string) Option {
	return func(c *Client) { c.apiKey = strings.TrimSpace(key) }
}

// WithHTTPClient replaces the default *http.Client (30s timeout).
func WithHTTPClient(hc *http.Client) Option {
	return func(c *Client) { c.httpClient = hc }
}

// NewClient builds a Client for baseURL (e.g. "https://api.mailack.com").
// Prefer WithAPIKey in production; without a key only public endpoints work
// (signup status).
func NewClient(baseURL string, opts ...Option) *Client {
	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
	for _, o := range opts {
		o(c)
	}
	return c
}

func (c *Client) do(ctx context.Context, method, path, idempotencyKey, contentType string, body any) (*http.Response, error) {
	var rdr io.Reader
	switch b := body.(type) {
	case nil:
	case []byte:
		rdr = bytes.NewReader(b)
	default:
		raw, err := json.Marshal(b)
		if err != nil {
			return nil, fmt.Errorf("mailack: encoding body: %w", err)
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, rdr)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	if idempotencyKey != "" {
		req.Header.Set("Idempotency-Key", idempotencyKey)
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	return c.httpClient.Do(req)
}

func checkStatus(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	apiErr := &APIError{Status: resp.StatusCode, Code: "http_error"}
	raw, err := io.ReadAll(resp.Body)
	if err == nil && len(raw) > 0 {
		var env struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if json.Unmarshal(raw, &env) == nil && env.Error.Code != "" {
			apiErr.Code = env.Error.Code
			apiErr.Message = env.Error.Message
		} else {
			apiErr.Message = strings.TrimSpace(string(raw))
		}
	}
	return apiErr
}

func decodeJSON[T any](resp *http.Response, out *T) error {
	defer resp.Body.Close()
	if err := checkStatus(resp); err != nil {
		return err
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("mailack: decoding response: %w", err)
	}
	return nil
}

// query builds path?k=v.
func query(path string, kv map[string]string) string {
	if len(kv) == 0 {
		return path
	}
	q := url.Values{}
	for k, v := range kv {
		if v != "" {
			q.Set(k, v)
		}
	}
	return path + "?" + q.Encode()
}
