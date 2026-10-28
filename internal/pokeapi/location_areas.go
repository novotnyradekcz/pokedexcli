package pokeapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// ListLocationAreas fetches a page of location areas. If pageURL is nil, the
// first page is returned; otherwise pageURL (a Next or Previous URL from an
// earlier response) is fetched. Responses are cached by URL.
func (c *Client) ListLocationAreas(pageURL *string) (LocationAreaList, error) {
	url := firstLocationAreaURL
	if pageURL != nil {
		url = *pageURL
	}

	if body, ok := c.cache.Get(url); ok {
		return decodeLocationAreaList(body)
	}

	res, err := c.httpClient.Get(url)
	if err != nil {
		return LocationAreaList{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return LocationAreaList{}, fmt.Errorf("unexpected response status: %s", res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return LocationAreaList{}, err
	}

	list, err := decodeLocationAreaList(body)
	if err != nil {
		return LocationAreaList{}, err
	}

	// Only cache responses that decoded successfully.
	c.cache.Add(url, body)
	return list, nil
}

func decodeLocationAreaList(body []byte) (LocationAreaList, error) {
	var list LocationAreaList
	if err := json.Unmarshal(body, &list); err != nil {
		return LocationAreaList{}, err
	}
	return list, nil
}
