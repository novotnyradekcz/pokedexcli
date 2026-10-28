package main

import (
	"encoding/json"
	"testing"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
)

const pidgeyJSON = `{
  "id": 16, "name": "pidgey", "base_experience": 50, "height": 3, "weight": 18,
  "stats": [
    {"base_stat": 40, "effort": 0, "stat": {"name": "hp"}},
    {"base_stat": 45, "effort": 0, "stat": {"name": "attack"}},
    {"base_stat": 40, "effort": 0, "stat": {"name": "defense"}},
    {"base_stat": 35, "effort": 0, "stat": {"name": "special-attack"}},
    {"base_stat": 35, "effort": 0, "stat": {"name": "special-defense"}},
    {"base_stat": 56, "effort": 1, "stat": {"name": "speed"}}
  ],
  "types": [
    {"slot": 1, "type": {"name": "normal"}},
    {"slot": 2, "type": {"name": "flying"}}
  ]
}`

func TestFormatPokemon(t *testing.T) {
	var p pokeapi.Pokemon
	if err := json.Unmarshal([]byte(pidgeyJSON), &p); err != nil {
		t.Fatal(err)
	}

	want := `Name: pidgey
Height: 3
Weight: 18
Stats:
  -hp: 40
  -attack: 45
  -defense: 40
  -special-attack: 35
  -special-defense: 35
  -speed: 56
Types:
  - normal
  - flying
`
	if got := formatPokemon(p); got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}
