package repl

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
	"github.com/novotnyradekcz/pokedexcli/internal/pokedex"
)

// fakeAPI is an in-memory API. Location area pages are keyed by URL; the
// first page is "page1". Lookups that miss return a 404, like the real API.
type fakeAPI struct {
	pages   map[string]pokeapi.LocationAreaList
	areas   map[string]pokeapi.LocationAreaDetail
	pokemon map[string]pokeapi.Pokemon
	calls   []string
}

func notFound() error {
	return &pokeapi.StatusError{StatusCode: 404, Status: "404 Not Found"}
}

func (f *fakeAPI) ListLocationAreas(pageURL *string) (pokeapi.LocationAreaList, error) {
	url := "page1"
	if pageURL != nil {
		url = *pageURL
	}
	f.calls = append(f.calls, "list "+url)
	if list, ok := f.pages[url]; ok {
		return list, nil
	}
	return pokeapi.LocationAreaList{}, notFound()
}

func (f *fakeAPI) GetLocationArea(name string) (pokeapi.LocationAreaDetail, error) {
	f.calls = append(f.calls, "area "+name)
	if area, ok := f.areas[name]; ok {
		return area, nil
	}
	return pokeapi.LocationAreaDetail{}, notFound()
}

func (f *fakeAPI) GetPokemon(name string) (pokeapi.Pokemon, error) {
	f.calls = append(f.calls, "pokemon "+name)
	if p, ok := f.pokemon[name]; ok {
		return p, nil
	}
	return pokeapi.Pokemon{}, notFound()
}

func areaList(next, previous *string, names ...string) pokeapi.LocationAreaList {
	list := pokeapi.LocationAreaList{Next: next, Previous: previous}
	for _, name := range names {
		list.Results = append(list.Results, pokeapi.LocationArea{Name: name})
	}
	return list
}

func ptr(s string) *string { return &s }

func newFakeAPI() *fakeAPI {
	return &fakeAPI{
		pages: map[string]pokeapi.LocationAreaList{
			"page1": areaList(ptr("page2"), nil, "area-1", "area-2"),
			"page2": areaList(nil, ptr("page1"), "area-3", "area-4"),
		},
		areas: map[string]pokeapi.LocationAreaDetail{
			"pastoria-city-area": {
				Name: "pastoria-city-area",
				PokemonEncounters: []pokeapi.PokemonEncounter{
					encounter("tentacool"), encounter("magikarp"),
				},
			},
		},
		pokemon: map[string]pokeapi.Pokemon{
			"pikachu":  {Name: "pikachu", BaseExperience: 112, Height: 4, Weight: 60},
			"caterpie": {Name: "caterpie", BaseExperience: 39},
		},
	}
}

func encounter(name string) pokeapi.PokemonEncounter {
	var e pokeapi.PokemonEncounter
	e.Pokemon.Name = name
	return e
}

// newTestConfig returns a config wired to api, plus the buffer its output goes
// to. By default every roll is 0, so every catch succeeds.
func newTestConfig(api API) (*config, *bytes.Buffer) {
	out := &bytes.Buffer{}
	cfg := &config{
		out:           out,
		roll:          func() float64 { return 0 },
		commands:      getCommands(),
		pokeapiClient: api,
		pokedex:       pokedex.New(),
	}
	return cfg, out
}

// run executes a command line such as "catch pikachu" and returns its error.
func run(t *testing.T, cfg *config, line string) error {
	t.Helper()
	words := cleanInput(line)
	cmd, ok := cfg.commands[words[0]]
	if !ok {
		t.Fatalf("unknown command %q", words[0])
	}
	return cmd.callback(cfg, words[1:]...)
}

// take returns what has been printed so far and clears the buffer.
func take(out *bytes.Buffer) string {
	s := out.String()
	out.Reset()
	return s
}

func TestMapAndMapbPagination(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	steps := []struct {
		command string
		want    string
	}{
		{"mapb", "you're on the first page\n"},
		{"map", "area-1\narea-2\n"},
		{"map", "area-3\narea-4\n"},
		{"map", "you're on the last page\n"},
		{"mapb", "area-1\narea-2\n"},
		{"mapb", "you're on the first page\n"},
	}
	for _, s := range steps {
		if err := run(t, cfg, s.command); err != nil {
			t.Fatalf("%s: unexpected error: %v", s.command, err)
		}
		if got := take(out); got != s.want {
			t.Errorf("%s printed %q, want %q", s.command, got, s.want)
		}
	}
}

func TestMapReportsAPIErrors(t *testing.T) {
	api := newFakeAPI()
	delete(api.pages, "page1")
	cfg, out := newTestConfig(api)

	if err := run(t, cfg, "map"); err == nil {
		t.Error("expected an error when the API call fails")
	}
	if got := out.String(); got != "" {
		t.Errorf("expected no output on error, got %q", got)
	}
}

func TestExplore(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	if err := run(t, cfg, "explore pastoria-city-area"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Exploring pastoria-city-area...\nFound Pokemon:\n - tentacool\n - magikarp\n"
	if got := take(out); got != want {
		t.Errorf("printed %q, want %q", got, want)
	}
}

func TestExploreErrors(t *testing.T) {
	cfg, _ := newTestConfig(newFakeAPI())

	cases := []struct {
		line    string
		wantErr string
	}{
		{"explore", "usage: explore <area_name>"},
		{"explore a b", "usage: explore <area_name>"},
		{"explore nowhere", `location area "nowhere" not found`},
	}
	for _, c := range cases {
		err := run(t, cfg, c.line)
		if err == nil || err.Error() != c.wantErr {
			t.Errorf("%q: error = %v, want %q", c.line, err, c.wantErr)
		}
	}
}

func TestCatchSuccess(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())
	cfg.roll = func() float64 { return 0 }

	if err := run(t, cfg, "catch pikachu"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Throwing a Pokeball at pikachu...\n" +
		"pikachu was caught!\n" +
		"You may now inspect it with the inspect command.\n"
	if got := take(out); got != want {
		t.Errorf("printed %q, want %q", got, want)
	}
	if _, ok := cfg.pokedex.Get("pikachu"); !ok {
		t.Error("expected pikachu to be in the Pokedex")
	}
}

func TestCatchEscape(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())
	cfg.roll = func() float64 { return 0.999 }

	if err := run(t, cfg, "catch pikachu"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "Throwing a Pokeball at pikachu...\npikachu escaped!\n"
	if got := take(out); got != want {
		t.Errorf("printed %q, want %q", got, want)
	}
	if cfg.pokedex.Len() != 0 {
		t.Error("an escaped Pokemon must not be added to the Pokedex")
	}
}

func TestCatchThrowMessageComesBeforeLookupError(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	err := run(t, cfg, "catch missingno")
	if err == nil || err.Error() != `pokemon "missingno" not found` {
		t.Errorf("error = %v, want pokemon not found", err)
	}
	if got := take(out); got != "Throwing a Pokeball at missingno...\n" {
		t.Errorf("printed %q", got)
	}
}

func TestCatchUsage(t *testing.T) {
	cfg, _ := newTestConfig(newFakeAPI())
	for _, line := range []string{"catch", "catch a b"} {
		err := run(t, cfg, line)
		if err == nil || err.Error() != "usage: catch <pokemon_name>" {
			t.Errorf("%q: error = %v, want usage message", line, err)
		}
	}
}

func TestCatchUsesTheCatchChance(t *testing.T) {
	// pikachu (base experience 112) has a ~57% chance: a roll just under
	// catches it, a roll just over lets it escape.
	chance := catchChance(112)

	for _, c := range []struct {
		roll   float64
		caught bool
	}{
		{chance - 0.01, true},
		{chance + 0.01, false},
	} {
		cfg, _ := newTestConfig(newFakeAPI())
		roll := c.roll
		cfg.roll = func() float64 { return roll }

		if err := run(t, cfg, "catch pikachu"); err != nil {
			t.Fatal(err)
		}
		if _, ok := cfg.pokedex.Get("pikachu"); ok != c.caught {
			t.Errorf("roll %.2f: caught = %v, want %v", c.roll, ok, c.caught)
		}
	}
}

func TestInspect(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	if err := run(t, cfg, "inspect pikachu"); err != nil {
		t.Fatal(err)
	}
	if got := take(out); got != "you have not caught that pokemon\n" {
		t.Errorf("before catching, printed %q", got)
	}

	if err := run(t, cfg, "catch pikachu"); err != nil {
		t.Fatal(err)
	}
	take(out)
	api := cfg.pokeapiClient.(*fakeAPI)
	callsBefore := len(api.calls)

	if err := run(t, cfg, "inspect pikachu"); err != nil {
		t.Fatal(err)
	}
	got := take(out)
	if !strings.HasPrefix(got, "Name: pikachu\nHeight: 4\nWeight: 60\n") {
		t.Errorf("after catching, printed %q", got)
	}
	if len(api.calls) != callsBefore {
		t.Errorf("inspect must not call the API, but made calls: %v", api.calls[callsBefore:])
	}

	err := run(t, cfg, "inspect")
	if err == nil || err.Error() != "usage: inspect <pokemon_name>" {
		t.Errorf("error = %v, want usage message", err)
	}
}

func TestPokedexListsPokemonInCatchOrder(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	if err := run(t, cfg, "pokedex"); err != nil {
		t.Fatal(err)
	}
	if got := take(out); !strings.Contains(got, "empty") {
		t.Errorf("empty Pokedex printed %q", got)
	}

	for _, line := range []string{"catch pikachu", "catch caterpie", "catch pikachu"} {
		if err := run(t, cfg, line); err != nil {
			t.Fatal(err)
		}
	}
	take(out)

	if err := run(t, cfg, "pokedex"); err != nil {
		t.Fatal(err)
	}
	want := "Your Pokedex:\n - pikachu\n - caterpie\n"
	if got := take(out); got != want {
		t.Errorf("printed %q, want %q", got, want)
	}
}

func TestHelpListsEveryCommandInOrder(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	if err := run(t, cfg, "help"); err != nil {
		t.Fatal(err)
	}
	got := take(out)

	last := -1
	for _, key := range []string{"catch", "exit", "explore", "help", "inspect", "map", "mapb", "pokedex"} {
		line := fmt.Sprintf("%s: %s\n", cfg.commands[key].name, cfg.commands[key].description)
		idx := strings.Index(got, line)
		if idx < 0 {
			t.Fatalf("help output is missing %q:\n%s", line, got)
		}
		if idx < last {
			t.Errorf("help output is not sorted by command name; %q is out of order", key)
		}
		last = idx
	}
}
