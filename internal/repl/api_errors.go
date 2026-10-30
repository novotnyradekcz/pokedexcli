package repl

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
)

// notFoundError is the user-facing message for a lookup the API answered with
// a 404. It keeps the original error reachable through Unwrap.
type notFoundError struct {
	kind string // what was looked up, e.g. "pokemon"
	name string
	err  error
}

func (e *notFoundError) Error() string {
	return fmt.Sprintf("%s %q not found", e.kind, e.name)
}

func (e *notFoundError) Unwrap() error { return e.err }

// describeAPIError turns a 404 from the API into a friendlier message about
// the thing (e.g. "pokemon") that was looked up; other errors pass through.
func describeAPIError(err error, kind, name string) error {
	var statusErr *pokeapi.StatusError
	if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound {
		return &notFoundError{kind: kind, name: name, err: err}
	}
	return err
}
