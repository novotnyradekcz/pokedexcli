// Package pokeapi is a small client for the PokeAPI (https://pokeapi.co).
package pokeapi

import (
	"net/http"
	"time"
)

const baseURL = "https://pokeapi.co/api/v2"

// Client talks to the PokeAPI.
type Client struct {
	httpClient http.Client
}

// NewClient creates a Client whose HTTP requests give up after timeout.
func NewClient(timeout time.Duration) Client {
	return Client{
		httpClient: http.Client{Timeout: timeout},
	}
}
