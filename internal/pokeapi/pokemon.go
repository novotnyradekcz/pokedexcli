package pokeapi

import "net/url"

// Pokemon is the subset of the pokemon endpoint's data that the Pokedex uses.
type Pokemon struct {
	ID             int    `json:"id"`
	Name           string `json:"name"`
	BaseExperience int    `json:"base_experience"`
}

// GetPokemon fetches the Pokemon with the given name (or id). Responses are
// cached by URL.
func (c *Client) GetPokemon(name string) (Pokemon, error) {
	pokemonURL := baseURL + "/pokemon/" + url.PathEscape(name)

	var pokemon Pokemon
	if err := c.getJSON(pokemonURL, &pokemon); err != nil {
		return Pokemon{}, err
	}
	return pokemon, nil
}
