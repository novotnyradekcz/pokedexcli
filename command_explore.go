package main

import (
	"errors"
	"fmt"
)

func commandExplore(cfg *config, args ...string) error {
	if len(args) != 1 {
		return errors.New("usage: explore <area_name>")
	}
	name := args[0]

	fmt.Printf("Exploring %s...\n", name)
	area, err := cfg.pokeapiClient.GetLocationArea(name)
	if err != nil {
		return describeAPIError(err, "location area", name)
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range area.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}
	return nil
}
