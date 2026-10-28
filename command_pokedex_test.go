package main

import (
	"reflect"
	"testing"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
)

func TestFormatPokedex(t *testing.T) {
	want := "Your Pokedex:\n - pidgey\n - caterpie\n"
	if got := formatPokedex([]string{"pidgey", "caterpie"}); got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	want = "Your Pokedex is empty. Use the catch command to add Pokemon.\n"
	if got := formatPokedex(nil); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestAddCaughtPokemonKeepsCatchOrderWithoutDuplicates(t *testing.T) {
	cfg := &config{caughtPokemon: make(map[string]pokeapi.Pokemon)}

	cfg.addCaughtPokemon(pokeapi.Pokemon{Name: "pidgey"})
	cfg.addCaughtPokemon(pokeapi.Pokemon{Name: "caterpie"})
	cfg.addCaughtPokemon(pokeapi.Pokemon{Name: "pidgey"}) // caught again

	if want := []string{"pidgey", "caterpie"}; !reflect.DeepEqual(cfg.caughtOrder, want) {
		t.Errorf("caughtOrder = %v, want %v", cfg.caughtOrder, want)
	}
	if len(cfg.caughtPokemon) != 2 {
		t.Errorf("expected 2 caught Pokemon, got %d", len(cfg.caughtPokemon))
	}
}
