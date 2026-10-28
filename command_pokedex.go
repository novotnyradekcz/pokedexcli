package main

import (
	"fmt"
	"strings"
)

func commandPokedex(cfg *config, args ...string) error {
	fmt.Print(formatPokedex(cfg.caughtOrder))
	return nil
}

func formatPokedex(names []string) string {
	if len(names) == 0 {
		return "Your Pokedex is empty. Use the catch command to add Pokemon.\n"
	}

	var b strings.Builder
	b.WriteString("Your Pokedex:\n")
	for _, name := range names {
		fmt.Fprintf(&b, " - %s\n", name)
	}
	return b.String()
}
