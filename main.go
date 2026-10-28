package main

import (
	"time"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
	"github.com/novotnyradekcz/pokedexcli/internal/repl"
)

func main() {
	client := pokeapi.NewClient(5*time.Second, 5*time.Minute)
	defer client.Close()

	repl.Run(client)
}
