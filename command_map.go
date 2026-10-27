package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const locationAreaURL = "https://pokeapi.co/api/v2/location-area/"

type locationAreaResponse struct {
	Count    int     `json:"count"`
	Next     *string `json:"next"`
	Previous *string `json:"previous"`
	Results  []struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"results"`
}

func getLocationAreas(url string) (locationAreaResponse, error) {
	res, err := http.Get(url)
	if err != nil {
		return locationAreaResponse{}, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return locationAreaResponse{}, fmt.Errorf("unexpected response status: %s", res.Status)
	}

	var data locationAreaResponse
	if err := json.NewDecoder(res.Body).Decode(&data); err != nil {
		return locationAreaResponse{}, err
	}
	return data, nil
}

func commandMap(cfg *config) error {
	url := locationAreaURL
	if cfg.next != nil {
		url = *cfg.next
	} else if cfg.previous != nil {
		// A previous page exists but there is no next one: we are at the end.
		fmt.Println("you're on the last page")
		return nil
	}

	return printLocationAreas(cfg, url)
}

func commandMapb(cfg *config) error {
	if cfg.previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}

	return printLocationAreas(cfg, *cfg.previous)
}

func printLocationAreas(cfg *config, url string) error {
	data, err := getLocationAreas(url)
	if err != nil {
		return err
	}

	cfg.next = data.Next
	cfg.previous = data.Previous

	for _, area := range data.Results {
		fmt.Println(area.Name)
	}
	return nil
}
