// Package repl implements the Pokedex's interactive command loop and the
// commands it runs.
package repl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/signal"
	"strings"

	"github.com/novotnyradekcz/pokedexcli/internal/pokedex"
)

// Run starts the REPL on the terminal (or piped stdin) and returns when the
// user exits, the input ends, or ctx is cancelled. api is used for all PokeAPI
// lookups.
func Run(ctx context.Context, api API) {
	cfg := &config{
		out:           os.Stdout,
		roll:          rand.Float64,
		commands:      getCommands(),
		pokeapiClient: api,
		pokedex:       pokedex.New(),
	}
	loop(ctx, cfg, newLineReader(cfg.out))
}

// loop reads and runs commands from reader until the input ends, ctx is
// cancelled, or a command returns errQuit.
func loop(ctx context.Context, cfg *config, reader lineReader) {
	for ctx.Err() == nil {
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

		err = runCommand(ctx, cfg, command, words[1:])
		switch {
		case err == nil:
		case errors.Is(err, errQuit):
			return
		case errors.Is(err, context.Canceled):
			fmt.Fprintln(cfg.out, "Canceled")
		default:
			fmt.Fprintln(cfg.out, err)
		}
	}
}

// runCommand runs a single command. While it runs, Ctrl-C cancels the context
// given to the command (for example to abort a slow request) instead of
// killing the program.
func runCommand(ctx context.Context, cfg *config, command cliCommand, args []string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	return command.callback(ctx, cfg, args...)
}

func cleanInput(text string) []string {
	output := strings.ToLower(text)
	words := strings.Fields(output)
	return words
}
