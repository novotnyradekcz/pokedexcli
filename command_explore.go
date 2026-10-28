package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
)

func commandExplore(cfg *config, args ...string) error {
	if len(args) != 1 {
		return errors.New("usage: explore <area_name>")
	}
	name := args[0]

	fmt.Printf("Exploring %s...\n", name)
	area, err := cfg.pokeapiClient.GetLocationArea(name)
	if err != nil {
		var statusErr *pokeapi.StatusError
		if errors.As(err, &statusErr) && statusErr.StatusCode == http.StatusNotFound {
			return fmt.Errorf("location area %q not found", name)
		}
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, encounter := range area.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}
	return nil
}
