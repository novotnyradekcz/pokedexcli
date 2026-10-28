package pokeapi

import (
	"encoding/json"
	"net/url"
)

// firstLocationAreaURL is spelled the same way the API spells the Previous
// URL of the second page, so the first page shares a cache entry whether it is
// reached via map or via mapb.
const firstLocationAreaURL = baseURL + "/location-area/?offset=0&limit=20"

// LocationAreaList is one page of results from the location-area endpoint.
// Next and Previous are nil when there is no page in that direction.
type LocationAreaList struct {
	Count    int            `json:"count"`
	Next     *string        `json:"next"`
	Previous *string        `json:"previous"`
	Results  []LocationArea `json:"results"`
}

// LocationArea is a named location area with a link to its details.
type LocationArea struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// LocationAreaDetail is the full description of a single location area.
type LocationAreaDetail struct {
	Name              string             `json:"name"`
	PokemonEncounters []PokemonEncounter `json:"pokemon_encounters"`
}

// PokemonEncounter links a Pokemon to the location area it can be found in.
type PokemonEncounter struct {
	Pokemon struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"pokemon"`
}

// ListLocationAreas fetches a page of location areas. If pageURL is nil, the
// first page is returned; otherwise pageURL (a Next or Previous URL from an
// earlier response) is fetched. Responses are cached by URL.
func (c *Client) ListLocationAreas(pageURL *string) (LocationAreaList, error) {
	url := firstLocationAreaURL
	if pageURL != nil {
		url = *pageURL
	}

	var list LocationAreaList
	if err := c.getJSON(url, &list); err != nil {
		return LocationAreaList{}, err
	}
	return list, nil
}

// GetLocationArea fetches the details of the location area with the given
// name (or id). Responses are cached by URL.
func (c *Client) GetLocationArea(name string) (LocationAreaDetail, error) {
	areaURL := baseURL + "/location-area/" + url.PathEscape(name)

	var detail LocationAreaDetail
	if err := c.getJSON(areaURL, &detail); err != nil {
		return LocationAreaDetail{}, err
	}
	return detail, nil
}

// getJSON GETs url (using the cache) and decodes the JSON body into v. The
// body is cached only after it has decoded successfully.
func (c *Client) getJSON(url string, v any) error {
	body, cached, err := c.get(url)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(body, v); err != nil {
		return err
	}

	if !cached {
		c.cache.Add(url, body)
	}
	return nil
}
