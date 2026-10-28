package pokedex

import (
	"reflect"
	"testing"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
)

func TestNewIsEmpty(t *testing.T) {
	d := New()
	if d.Len() != 0 {
		t.Errorf("Len = %d, want 0", d.Len())
	}
	if names := d.Names(); len(names) != 0 {
		t.Errorf("Names = %v, want empty", names)
	}
	if _, ok := d.Get("pikachu"); ok {
		t.Error("expected Get on an empty Pokedex to find nothing")
	}
}

func TestAddAndGet(t *testing.T) {
	d := New()
	d.Add(pokeapi.Pokemon{Name: "pidgey", Height: 3})

	p, ok := d.Get("pidgey")
	if !ok || p.Name != "pidgey" || p.Height != 3 {
		t.Errorf("Get = %+v, %v; want pidgey with height 3", p, ok)
	}
	if _, ok := d.Get("caterpie"); ok {
		t.Error("expected an uncaught Pokemon to not be found")
	}
}

func TestNamesKeepCatchOrderWithoutDuplicates(t *testing.T) {
	d := New()
	d.Add(pokeapi.Pokemon{Name: "pidgey"})
	d.Add(pokeapi.Pokemon{Name: "caterpie"})
	d.Add(pokeapi.Pokemon{Name: "pidgey", Height: 99}) // caught again

	if want := []string{"pidgey", "caterpie"}; !reflect.DeepEqual(d.Names(), want) {
		t.Errorf("Names = %v, want %v", d.Names(), want)
	}
	if d.Len() != 2 {
		t.Errorf("Len = %d, want 2", d.Len())
	}
	if p, _ := d.Get("pidgey"); p.Height != 99 {
		t.Errorf("expected the data to be replaced, got height %d", p.Height)
	}
}

func TestNamesReturnsACopy(t *testing.T) {
	d := New()
	d.Add(pokeapi.Pokemon{Name: "pidgey"})

	d.Names()[0] = "tampered"

	if got := d.Names()[0]; got != "pidgey" {
		t.Errorf("modifying the returned slice changed the Pokedex: %q", got)
	}
}
