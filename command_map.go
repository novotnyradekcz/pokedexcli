package main

import "fmt"

func commandMap(cfg *config) error {
	if cfg.next == nil && cfg.previous != nil {
		// A previous page exists but there is no next one: we are at the end.
		fmt.Println("you're on the last page")
		return nil
	}

	return printLocationAreas(cfg, cfg.next)
}

func commandMapb(cfg *config) error {
	if cfg.previous == nil {
		fmt.Println("you're on the first page")
		return nil
	}

	return printLocationAreas(cfg, cfg.previous)
}

func printLocationAreas(cfg *config, pageURL *string) error {
	list, err := cfg.pokeapiClient.ListLocationAreas(pageURL)
	if err != nil {
		return err
	}

	cfg.next = list.Next
	cfg.previous = list.Previous

	for _, area := range list.Results {
		fmt.Println(area.Name)
	}
	return nil
}
