# Pokedex CLI

A command-line Pokedex for exploring Pokémon location areas, catching Pokémon, and inspecting your collection. It uses the [PokéAPI](https://pokeapi.co/) for location and Pokémon data.

This project was originally created as part of the [Build a Pokedex CLI in Go course on boot.dev](https://www.boot.dev/courses/build-pokedex-cli-golang).

## Features

- Browse location areas 20 at a time with `map` and `mapb`.
- Explore a location area to see the Pokémon found there.
- Try to catch Pokémon; higher base experience makes a Pokémon harder to catch.
- Inspect the details of Pokémon you have caught.
- List your caught Pokémon in catch order.
- Use the up and down arrow keys to navigate command history in an interactive terminal.
- Cache API responses in memory to avoid repeated requests.

## Requirements

- Go 1.26.4 or newer, as declared in `go.mod`.
- Network access to [pokeapi.co](https://pokeapi.co/) when requesting data that is not cached.

## Getting started

Clone the repository, change into the project directory, and start the CLI:

```sh
git clone https://github.com/novotnyradekcz/pokedexcli.git
cd pokedexcli
go run .
```

Alternatively, build an executable and run it:

```sh
go build -o pokedex .
./pokedex
```

## Commands

| Command | Description |
| --- | --- |
| `help` | Show available commands. |
| `map` | Display the next 20 location areas. Repeated calls page forward. |
| `mapb` | Display the previous 20 location areas. |
| `explore <area_name>` | Show Pokémon found in a location area, for example `explore pastoria-city-area`. |
| `catch <pokemon_name>` | Try to catch a Pokémon, for example `catch pikachu`. |
| `inspect <pokemon_name>` | Show details for a Pokémon you have caught, for example `inspect pikachu`. |
| `pokedex` | List caught Pokémon in the order they were first caught. |
| `exit` | Quit the CLI. |

Commands and arguments are case-insensitive. In an interactive terminal, use the up and down arrow keys to recall earlier commands. History is held for the current session only.

### Example session

```text
Pokedex > map
canalave-city-area
eterna-city-area
...
Pokedex > explore pastoria-city-area
Exploring pastoria-city-area...
Found Pokemon:
 - tentacool
 - tentacruel
...
Pokedex > catch pikachu
Throwing a Pokeball at pikachu...
pikachu was caught!
You may now inspect it with the inspect command.
Pokedex > inspect pikachu
Name: pikachu
Height: 4
Weight: 60
Stats:
  -hp: 35
  -attack: 55
  ...
Types:
  - electric
Pokedex > pokedex
Your Pokedex:
 - pikachu
```

## Data and behavior

- Location-area pages, location-area details, and Pokémon data are fetched from the [PokéAPI v2](https://pokeapi.co/docs/v2).
- Successful API response bodies are cached in memory for five minutes. The cache is shared by the API client and is safe for concurrent use. The client has a five-second request timeout.
- Caught Pokémon and command history are not persisted. Both are lost when the process exits.
- Inspecting a Pokémon does not make an API request; it uses the data stored when the Pokémon was caught.

## Project layout

```text
.
├── main.go                 # Application wiring
└── internal/
    ├── pokeapi/            # PokéAPI client and response types
    ├── pokecache/          # Concurrent in-memory expiring cache
    ├── pokedex/            # Caught-Pokémon collection and catch order
    └── repl/               # REPL, terminal input, and commands
```

## Tests and checks

Run the tests, including the race detector:

```sh
go test -race ./...
```

Run static analysis:

```sh
go vet ./...
```

The API client tests use a fake HTTP transport and do not require network access.

## License

No license is currently specified in this repository. Contact the project owner before redistributing or reusing the code if you need explicit licensing terms.
