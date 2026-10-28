package repl

import (
	"errors"
	"fmt"
)

// catchDifficulty controls how hard catching is. The chance of catching a
// Pokemon is catchDifficulty / (catchDifficulty + baseExperience), so a higher
// base experience means a lower chance, but the chance is never zero.
// For reference: base experience 64 -> ~70%, 112 (pikachu) -> ~57%,
// 306 (mewtwo) -> ~33%, 608 (blissey) -> ~20%.
const catchDifficulty = 150.0

// catchChance returns the probability (0 < p <= 1) of catching a Pokemon with
// the given base experience.
func catchChance(baseExperience int) float64 {
	if baseExperience < 0 {
		baseExperience = 0
	}
	return catchDifficulty / (catchDifficulty + float64(baseExperience))
}

func commandCatch(cfg *config, args ...string) error {
	if len(args) != 1 {
		return errors.New("usage: catch <pokemon_name>")
	}
	name := args[0]

	fmt.Fprintf(cfg.out, "Throwing a Pokeball at %s...\n", name)
	pokemon, err := cfg.pokeapiClient.GetPokemon(name)
	if err != nil {
		return describeAPIError(err, "pokemon", name)
	}

	if cfg.roll() >= catchChance(pokemon.BaseExperience) {
		fmt.Fprintf(cfg.out, "%s escaped!\n", pokemon.Name)
		return nil
	}

	fmt.Fprintf(cfg.out, "%s was caught!\n", pokemon.Name)
	fmt.Fprintln(cfg.out, "You may now inspect it with the inspect command.")
	cfg.pokedex.Add(pokemon)
	return nil
}
