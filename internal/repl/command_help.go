package repl

import (
	"context"
	"fmt"
	"sort"
)

func commandHelp(_ context.Context, cfg *config, args ...string) error {
	// Sort so the output is the same on every run (map order is random).
	keys := make([]string, 0, len(cfg.commands))
	for key := range cfg.commands {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	fmt.Fprintln(cfg.out)
	fmt.Fprintln(cfg.out, "Welcome to the Pokedex!")
	fmt.Fprintln(cfg.out, "Usage:")
	fmt.Fprintln(cfg.out)
	for _, key := range keys {
		cmd := cfg.commands[key]
		fmt.Fprintf(cfg.out, "%s: %s\n", cmd.name, cmd.description)
	}
	fmt.Fprintln(cfg.out)
	return nil
}
