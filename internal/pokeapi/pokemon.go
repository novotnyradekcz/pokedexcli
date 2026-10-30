package pokeapi

import (
	"context"
	"net/url"
)

// Pokemon is the subset of the pokemon endpoint's data that the Pokedex uses.
type Pokemon struct {
	ID             int           `json:"id"`
	Name           string        `json:"name"`
	BaseExperience int           `json:"base_experience"`
	Height         int           `json:"height"`
	Weight         int           `json:"weight"`
	Stats          []PokemonStat `json:"stats"`
	Types          []PokemonType `json:"types"`
}

// PokemonStat is one base stat (hp, attack, ...) of a Pokemon.
type PokemonStat struct {
	BaseStat int `json:"base_stat"`
	Stat     struct {
		Name string `json:"name"`
	} `json:"stat"`
}

// PokemonType is one of the types (normal, flying, ...) of a Pokemon.
type PokemonType struct {
	Type struct {
		Name string `json:"name"`
	} `json:"type"`
}

// GetPokemon fetches the Pokemon with the given name (or id). Responses are
// cached by URL.
func (c *Client) GetPokemon(ctx context.Context, name string) (Pokemon, error) {
	pokemonURL := c.baseURL + "/pokemon/" + url.PathEscape(name)

	var pokemon Pokemon
	if err := c.getJSON(ctx, pokemonURL, &pokemon); err != nil {
		return Pokemon{}, err
	}
	return pokemon, nil
}
