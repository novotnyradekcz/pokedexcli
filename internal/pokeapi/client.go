// Package pokeapi is a small client for the PokeAPI (https://pokeapi.co).
package pokeapi

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/novotnyradekcz/pokedexcli/internal/pokecache"
)

const baseURL = "https://pokeapi.co/api/v2"

// Client talks to the PokeAPI and caches responses by URL.
type Client struct {
	httpClient http.Client
	cache      *pokecache.Cache
}

// NewClient creates a Client whose HTTP requests give up after timeout and
// whose cached responses expire after cacheInterval.
func NewClient(timeout, cacheInterval time.Duration) Client {
	return Client{
		httpClient: http.Client{Timeout: timeout},
		cache:      pokecache.NewCache(cacheInterval),
	}
}

// get returns the body of a successful GET request to url, or the cached body
// (with cached == true) without making a request. Bodies fetched from the
// network are not cached here; callers add them to the cache once they have
// confirmed the body is valid, so bad responses are never cached.
func (c *Client) get(url string) (body []byte, cached bool, err error) {
	if body, ok := c.cache.Get(url); ok {
		return body, true, nil
	}

	res, err := c.httpClient.Get(url)
	if err != nil {
		return nil, false, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, false, &StatusError{StatusCode: res.StatusCode, Status: res.Status}
	}

	body, err = io.ReadAll(res.Body)
	if err != nil {
		return nil, false, err
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
