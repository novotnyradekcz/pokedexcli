// Package repl implements the Pokedex's interactive command loop and the
// commands it runs.
package repl

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"

	"github.com/novotnyradekcz/pokedexcli/internal/pokedex"
)

// Run starts the REPL on the terminal (or piped stdin) and returns when the
// user exits or the input ends. api is used for all PokeAPI lookups.
func Run(api API) {
	cfg := &config{
		out:           os.Stdout,
		roll:          rand.Float64,
		commands:      getCommands(),
		pokeapiClient: api,
		pokedex:       pokedex.New(),
	}
	loop(cfg, newLineReader(cfg.out))
}

// loop reads and runs commands from reader until the input ends or a command
// returns errQuit.
func loop(cfg *config, reader lineReader) {
	for {
		line, err := reader.ReadLine()
		if err != nil {
			// End of input (Ctrl-D / Ctrl-C, or piped input ran out).
			fmt.Fprintln(cfg.out)
			if err != io.EOF {
				fmt.Fprintln(cfg.out, err)
			}
			return
		}

		words := cleanInput(line)
		if len(words) == 0 {
			continue
		}

		command, exists := cfg.commands[words[0]]
		if !exists {
			fmt.Fprintln(cfg.out, "Unknown command")
			continue
		}

		if err := command.callback(cfg, words[1:]...); err != nil {
			if errors.Is(err, errQuit) {
				return
			}
			fmt.Fprintln(cfg.out, err)
		}
	}
}

func cleanInput(text string) []string {
	output := strings.ToLower(text)
	words := strings.Fields(output)
	return words
}
