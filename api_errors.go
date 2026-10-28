package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
)

// describeAPIError turns a 404 from the API into a friendlier message about
// the thing (e.g. "pokemon") that was looked up; other errors pass through.
func describeAPIError(err error, kind, name string) error {
	var statusErr *pokeapi.StatusError
	if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound {
		return fmt.Errorf("%s %q not found", kind, name)
	}
	return err
}
