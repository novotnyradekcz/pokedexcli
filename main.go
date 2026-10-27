package main

import (
	"time"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
)

func main() {
	cfg := &config{
		commands:      getCommands(),
		pokeapiClient: pokeapi.NewClient(5 * time.Second),
	}
	startRepl(cfg)
}
