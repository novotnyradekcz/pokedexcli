package repl

import (
	"context"
	"io"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
	"github.com/novotnyradekcz/pokedexcli/internal/pokedex"
)

// API is the part of the PokeAPI client that the commands use. It is defined
// here, where it is consumed, so tests can substitute a fake.
type API interface {
	ListLocationAreas(ctx context.Context, pageURL *string) (pokeapi.LocationAreaList, error)
	GetLocationArea(ctx context.Context, name string) (pokeapi.LocationAreaDetail, error)
	GetPokemon(ctx context.Context, name string) (pokeapi.Pokemon, error)
}

// config is the state shared by all commands.
type config struct {
	out           io.Writer      // where commands print to
	roll          func() float64 // returns a random number in [0.0, 1.0)
	commands      map[string]cliCommand
	pokeapiClient API
	pokedex       *pokedex.Pokedex
	next          *string // next page of location areas, nil if there is none
	previous      *string // previous page of location areas, nil if there is none
}
