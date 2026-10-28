// Package pokedex keeps track of the Pokemon a user has caught.
package pokedex

import "github.com/novotnyradekcz/pokedexcli/internal/pokeapi"

// Pokedex is a collection of caught Pokemon that remembers the order in which
// they were first caught. The zero value is not ready for use; call New.
type Pokedex struct {
	pokemon map[string]pokeapi.Pokemon
	order   []string
}

// New returns an empty Pokedex.
func New() *Pokedex {
	return &Pokedex{pokemon: make(map[string]pokeapi.Pokemon)}
}

// Add records p. Adding a Pokemon that is already present replaces its data
// but keeps its original position in the catch order.
func (d *Pokedex) Add(p pokeapi.Pokemon) {
	if _, exists := d.pokemon[p.Name]; !exists {
		d.order = append(d.order, p.Name)
	}
	d.pokemon[p.Name] = p
}

// Get returns the Pokemon with the given name and whether it has been caught.
func (d *Pokedex) Get(name string) (pokeapi.Pokemon, bool) {
	p, ok := d.pokemon[name]
	return p, ok
}

// Names returns the names of all caught Pokemon, in the order they were first
// caught. The returned slice is a copy and is safe to modify.
func (d *Pokedex) Names() []string {
	return append([]string(nil), d.order...)
}

// Len returns the number of caught Pokemon.
func (d *Pokedex) Len() int {
	return len(d.order)
}
