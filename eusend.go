// Package eusend is the official Go SDK for the Eusend API — the EU-native
// transactional email platform. Its shape mirrors github.com/resend/resend-go,
// so migrating from Resend is largely a `resend` → `eusend` rename.
//
//	client := eusend.NewClient("eu_live_...")
//	sent, err := client.Emails.Send(&eusend.SendEmailRequest{
//	    From:    "Acme <you@yourdomain.com>",
//	    To:      []string{"user@example.com"},
//	    Subject: "Hello",
//	    Html:    "<p>Hello world</p>",
//	})
package eusend

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	version     = "0.1.0"
	userAgent   = "eusend-go/" + version
	contentType = "application/json"
)

var defaultBaseURL = getEnv("EUSEND_BASE_URL", "https://api.eusend.dev/")

var defaultHTTPClient = &http.Client{Timeout: time.Minute}

// Options is implemented by per-call option structs (e.g. SendEmailOptions).
type Options interface {
	GetIdempotencyKey() string
}

// Client is the entry point to the Eusend API. Create one with NewClient and
// share it across goroutines — it is safe for concurrent use.
type Client struct {
	client    *http.Client
	ApiKey    string
	BaseURL   *url.URL
	UserAgent string
	headers   map[string]string

	Emails     EmailsSvc
	Batch      BatchSvc
	ApiKeys    ApiKeysSvc
	Domains    DomainsSvc
	Audiences  AudiencesSvc
	Templates  TemplatesSvc
	Webhooks   WebhooksSvc
	Broadcasts BroadcastsSvc
}

// NewClient creates a Client with the given API key. If apiKey is empty, the
// EUSEND_API_KEY environment variable is used.
func NewClient(apiKey string) *Client {
	if apiKey == "" {
		apiKey = os.Getenv("EUSEND_API_KEY")
	}
	key := strings.Trim(strings.TrimSpace(apiKey), "'")
	return NewCustomClient(defaultHTTPClient, key)
}

// NewCustomClient creates a Client with a custom *http.Client (for custom
// timeouts, proxies, etc.).
func NewCustomClient(httpClient *http.Client, apiKey string) *Client {
	if httpClient == nil {
		httpClient = defaultHTTPClient
	}
	baseURL, _ := url.Parse(defaultBaseURL)
	c := &Client{
		client:    httpClient,
		BaseURL:   baseURL,
		UserAgent: userAgent,
		ApiKey:    apiKey,
		headers:   make(map[string]string),
	}
	c.Emails = &EmailsSvcImpl{client: c}
	c.Batch = &BatchSvcImpl{client: c}
	c.ApiKeys = &ApiKeysSvcImpl{client: c}
	c.Domains = &DomainsSvcImpl{client: c}
	c.Audiences = &AudiencesSvcImpl{client: c}
	c.Templates = &TemplatesSvcImpl{client: c}
	c.Webhooks = &WebhooksSvcImpl{client: c}
	c.Broadcasts = &BroadcastsSvcImpl{client: c}
	return c
}

// NewRequest builds an *http.Request against the API, JSON-encoding params when
// non-nil.
func (c *Client) NewRequest(ctx context.Context, method, path string, params any) (*http.Request, error) {
	u, err := c.BaseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, u.String(), nil)
	if err != nil {
		return nil, err
	}

	if params != nil {
		buf := new(bytes.Buffer)
		if err := json.NewEncoder(buf).Encode(params); err != nil {
			return nil, err
		}
		req.Body = io.NopCloser(buf)
		req.Header.Set("Content-Type", contentType)
	}

	for k, v := range c.headers {
		req.Header.Add(k, v)
	}
	req.Header.Set("Accept", contentType)
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Authorization", "Bearer "+c.ApiKey)
	return req, nil
}

// NewRequestWithOptions is NewRequest plus per-call options such as an
// idempotency key.
func (c *Client) NewRequestWithOptions(ctx context.Context, method, path string, params any, options Options) (*http.Request, error) {
	req, err := c.NewRequest(ctx, method, path, params)
	if err != nil {
		return nil, err
	}
	if options != nil && options.GetIdempotencyKey() != "" && method == http.MethodPost {
		req.Header.Set("Idempotency-Key", options.GetIdempotencyKey())
	}
	return req, nil
}

// Perform sends req and decodes a 2xx JSON body into ret (which may be nil).
// Non-2xx responses are converted to *Error via handleError.
func (c *Client) Perform(req *http.Request, ret any) (*http.Response, error) {
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, &Error{Message: "network request failed: the request could not be resolved", Code: CodeApplicationError}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return resp, handleError(resp)
	}

	if resp.StatusCode != http.StatusNoContent && ret != nil {
		if err := json.NewDecoder(resp.Body).Decode(ret); err != nil && err != io.EOF {
			return resp, &Error{Message: "failed to decode response: " + err.Error(), Code: CodeApplicationError, StatusCode: resp.StatusCode}
		}
	}
	return resp, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// String returns a pointer to v — a convenience for optional *string fields.
func String(v string) *string { return &v }

// Bool returns a pointer to v — a convenience for optional *bool fields.
func Bool(v bool) *bool { return &v }
