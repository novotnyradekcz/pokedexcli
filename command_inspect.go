package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
)

func commandInspect(cfg *config, args ...string) error {
	if len(args) != 1 {
		return errors.New("usage: inspect <pokemon_name>")
	}

	pokemon, ok := cfg.caughtPokemon[args[0]]
	if !ok {
		fmt.Println("you have not caught that pokemon")
		return nil
	}

	fmt.Print(formatPokemon(pokemon))
	return nil
}

func formatPokemon(p pokeapi.Pokemon) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Name: %s\n", p.Name)
	fmt.Fprintf(&b, "Height: %d\n", p.Height)
	fmt.Fprintf(&b, "Weight: %d\n", p.Weight)
	b.WriteString("Stats:\n")
	for _, s := range p.Stats {
		fmt.Fprintf(&b, "  -%s: %d\n", s.Stat.Name, s.BaseStat)
	}
	b.WriteString("Types:\n")
	for _, t := range p.Types {
		fmt.Fprintf(&b, "  - %s\n", t.Type.Name)
	}
	return b.String()
}
