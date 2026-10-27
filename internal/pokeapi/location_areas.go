package pokeapi

import (
	"encoding/json"
	"fmt"
	"net/http"
)

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

// ListLocationAreas fetches a page of location areas. If pageURL is nil, the
// first page is returned; otherwise pageURL (a Next or Previous URL from an
// earlier response) is fetched.
func (c *Client) ListLocationAreas(pageURL *string) (LocationAreaList, error) {
	url := baseURL + "/location-area/"
	if pageURL != nil {
		url = *pageURL
	}

	res, err := c.httpClient.Get(url)
	if err != nil {
		return LocationAreaList{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return LocationAreaList{}, fmt.Errorf("unexpected response status: %s", res.Status)
	}

	var list LocationAreaList
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil {
		return LocationAreaList{}, err
	}
	return list, nil
}
