// Package pokeapi is a small client for the PokeAPI (https://pokeapi.co).
package pokeapi

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/novotnyradekcz/pokedexcli/internal/pokecache"
)

const defaultBaseURL = "https://pokeapi.co/api/v2"

// Client talks to the PokeAPI and caches responses by URL.
type Client struct {
	baseURL    string
	httpClient http.Client
	cache      *pokecache.Cache
}

// Option customizes a Client created by NewClient.
type Option func(*Client)

// WithBaseURL makes the client call the API at baseURL instead of the real
// PokeAPI, for example a local test server.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) { c.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithTransport makes the client send its requests through rt instead of
// http.DefaultTransport, for example a fake in tests.
func WithTransport(rt http.RoundTripper) Option {
	return func(c *Client) { c.httpClient.Transport = rt }
}

// NewClient creates a Client whose HTTP requests give up after timeout and
// whose cached responses expire after cacheInterval. Call Close when the
// client is no longer needed.
func NewClient(timeout, cacheInterval time.Duration, opts ...Option) *Client {
	c := &Client{
		baseURL:    defaultBaseURL,
		httpClient: http.Client{Timeout: timeout},
		cache:      pokecache.NewCache(cacheInterval),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// Close releases the resources held by the client (its cache's background
// goroutine). The client must not be used afterwards.
func (c *Client) Close() {
	c.cache.Stop()
}

// get returns the body of a successful GET request to url, or the cached body
// (with cached == true) without making a request. Cancelling ctx aborts a
// request that is in flight. Bodies fetched from the
// network are not cached here; callers add them to the cache once they have
// confirmed the body is valid, so bad responses are never cached.
func (c *Client) get(ctx context.Context, url string) (body []byte, cached bool, err error) {
	if body, ok := c.cache.Get(url); ok {
		return body, true, nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("creating request: %w", err)
	}

	// The error from Do already names the method and URL.
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	// Closing a fully read (or rejected) response body cannot lose data.
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, false, &StatusError{StatusCode: res.StatusCode, Status: res.Status}
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, false, fmt.Errorf("reading response from %s: %w", url, err)
	}
	return body, false, nil
}

// StatusError is returned when the API responds with a non-200 status.
type StatusError struct {
	StatusCode int
	Status     string
}

func (e *StatusError) Error() string {
	return fmt.Sprintf("unexpected response status: %s", e.Status)
}
