package repl

import (
	"errors"
	"fmt"
)

// errQuit is returned by the exit command to tell the REPL loop to stop.
var errQuit = errors.New("quit")

func commandExit(cfg *config, args ...string) error {
	fmt.Fprintln(cfg.out, "Closing the Pokedex... Goodbye!")
	return errQuit
}
