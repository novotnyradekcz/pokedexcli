// Package pokeapi is a small client for the PokeAPI (https://pokeapi.co).
package pokeapi

import (
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
