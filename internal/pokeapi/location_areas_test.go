package pokeapi

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

const testBaseURL = "https://pokeapi.test/api/v2"

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func response(r *http.Request, status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Body:       io.NopCloser(strings.NewReader(body)),
		Header:     make(http.Header),
		Request:    r,
	}
}

// newTestClient returns a client that sends every request to handler instead
// of the network, and counts how many requests were made.
func newTestClient(t *testing.T, handler func(*http.Request) *http.Response) (*Client, *atomic.Int32) {
	t.Helper()
	var requests atomic.Int32
	client := NewClient(5*time.Second, time.Minute,
		WithBaseURL(testBaseURL),
		WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			requests.Add(1)
			return handler(r), nil
		})),
	)
	t.Cleanup(client.Close)
	return client, &requests
}

// fixedResponse returns a handler that always answers with status and body.
func fixedResponse(status int, body string) func(*http.Request) *http.Response {
	return func(r *http.Request) *http.Response { return response(r, status, body) }
}

func TestListLocationAreasUsesCache(t *testing.T) {
	client, requests := newTestClient(t, fixedResponse(http.StatusOK,
		`{"count":1,"next":null,"previous":null,"results":[{"name":"canalave-city-area","url":"u"}]}`))
	url := "https://example.com/location-area/"

	for i := 0; i < 3; i++ {
		list, err := client.ListLocationAreas(&url)
		if err != nil {
			t.Fatalf("request %d: unexpected error: %v", i, err)
		}
		if len(list.Results) != 1 || list.Results[0].Name != "canalave-city-area" {
			t.Fatalf("request %d: unexpected results: %+v", i, list.Results)
		}
	}

	if got := requests.Load(); got != 1 {
		t.Errorf("expected 1 HTTP request, got %d", got)
	}
}

func TestListLocationAreasFirstPageURL(t *testing.T) {
	var gotURL string
	client, _ := newTestClient(t, func(r *http.Request) *http.Response {
		gotURL = r.URL.String()
		return response(r, http.StatusOK, `{"results":[]}`)
	})

	if _, err := client.ListLocationAreas(nil); err != nil {
		t.Fatal(err)
	}

	if want := testBaseURL + "/location-area/?offset=0&limit=20"; gotURL != want {
		t.Errorf("requested %q, want %q", gotURL, want)
	}
}

// Mirrors the real API: page 2's Previous URL must be the same URL that was
// used to fetch page 1, so going back is served from the cache.
func TestPreviousPageIsServedFromCache(t *testing.T) {
	const page1 = testBaseURL + "/location-area/?offset=0&limit=20"
	const page2 = testBaseURL + "/location-area/?offset=20&limit=20"
	pages := map[string]string{
		page1: `{"next":"` + page2 + `","previous":null,"results":[{"name":"a"}]}`,
		page2: `{"next":null,"previous":"` + page1 + `","results":[{"name":"b"}]}`,
	}

	client, requests := newTestClient(t, func(r *http.Request) *http.Response {
		body, ok := pages[r.URL.String()]
		if !ok {
			t.Errorf("unexpected request to %s", r.URL)
		}
		return response(r, http.StatusOK, body)
	})

	first, err := client.ListLocationAreas(nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := client.ListLocationAreas(first.Next)
	if err != nil {
		t.Fatal(err)
	}
	back, err := client.ListLocationAreas(second.Previous)
	if err != nil {
		t.Fatal(err)
	}

	if back.Results[0].Name != "a" {
		t.Errorf("expected to go back to page 1, got %+v", back.Results)
	}
	if got := requests.Load(); got != 2 {
		t.Errorf("expected 2 HTTP requests (going back should hit the cache), got %d", got)
	}
}

func TestGetLocationAreaUsesCache(t *testing.T) {
	var gotURL string
	client, requests := newTestClient(t, func(r *http.Request) *http.Response {
		gotURL = r.URL.String()
		return response(r, http.StatusOK,
			`{"name":"pastoria-city-area","pokemon_encounters":[`+
				`{"pokemon":{"name":"tentacool","url":"u1"}},`+
				`{"pokemon":{"name":"magikarp","url":"u2"}}]}`)
	})

	for i := 0; i < 2; i++ {
		area, err := client.GetLocationArea("pastoria-city-area")
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if len(area.PokemonEncounters) != 2 ||
			area.PokemonEncounters[0].Pokemon.Name != "tentacool" ||
			area.PokemonEncounters[1].Pokemon.Name != "magikarp" {
			t.Fatalf("call %d: unexpected encounters: %+v", i, area.PokemonEncounters)
		}
	}

	if want := testBaseURL + "/location-area/pastoria-city-area"; gotURL != want {
		t.Errorf("requested %q, want %q", gotURL, want)
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("expected 1 HTTP request, got %d", got)
	}
}

func TestGetPokemonUsesCache(t *testing.T) {
	var gotURL string
	client, requests := newTestClient(t, func(r *http.Request) *http.Response {
		gotURL = r.URL.String()
		return response(r, http.StatusOK, `{"id":25,"name":"pikachu","base_experience":112,"height":4,"weight":60,`+
			`"stats":[{"base_stat":35,"stat":{"name":"hp"}},{"base_stat":55,"stat":{"name":"attack"}}],`+
			`"types":[{"type":{"name":"electric"}}]}`)
	})

	for i := 0; i < 2; i++ {
		p, err := client.GetPokemon("pikachu")
		if err != nil {
			t.Fatalf("call %d: unexpected error: %v", i, err)
		}
		if p.ID != 25 || p.Name != "pikachu" || p.BaseExperience != 112 || p.Height != 4 || p.Weight != 60 {
			t.Errorf("call %d: unexpected Pokemon fields: %+v", i, p)
		}
		if len(p.Stats) != 2 || p.Stats[0].BaseStat != 35 || p.Stats[0].Stat.Name != "hp" ||
			p.Stats[1].BaseStat != 55 || p.Stats[1].Stat.Name != "attack" {
			t.Errorf("call %d: unexpected stats: %+v", i, p.Stats)
		}
		if len(p.Types) != 1 || p.Types[0].Type.Name != "electric" {
			t.Errorf("call %d: unexpected types: %+v", i, p.Types)
		}
	}

	if want := testBaseURL + "/pokemon/pikachu"; gotURL != want {
		t.Errorf("requested %q, want %q", gotURL, want)
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("expected 1 HTTP request, got %d", got)
	}
}

func TestGetLocationAreaNotFound(t *testing.T) {
	client, _ := newTestClient(t, fixedResponse(http.StatusNotFound, "Not Found"))

	_, err := client.GetLocationArea("nowhere")
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected a 404 StatusError, got %v", err)
	}
}

func TestListLocationAreasDoesNotCacheErrors(t *testing.T) {
	client, requests := newTestClient(t, fixedResponse(http.StatusInternalServerError, "boom"))
	url := "https://example.com/location-area/"

	for i := 0; i < 2; i++ {
		if _, err := client.ListLocationAreas(&url); err == nil {
			t.Fatal("expected an error for a 500 response")
		}
	}

	if got := requests.Load(); got != 2 {
		t.Errorf("expected 2 HTTP requests, got %d", got)
	}
}

func TestInvalidJSONIsNotCached(t *testing.T) {
	client, requests := newTestClient(t, fixedResponse(http.StatusOK, "not json"))

	for i := 0; i < 2; i++ {
		if _, err := client.GetPokemon("pikachu"); err == nil {
			t.Fatal("expected a decoding error")
		}
	}

	if got := requests.Load(); got != 2 {
		t.Errorf("expected 2 HTTP requests (bad bodies must not be cached), got %d", got)
	}
}

func TestWithBaseURLTrimsTrailingSlash(t *testing.T) {
	var gotURL string
	client := NewClient(time.Second, time.Minute,
		WithBaseURL("https://pokeapi.test/api/v2/"),
		WithTransport(roundTripFunc(func(r *http.Request) (*http.Response, error) {
			gotURL = r.URL.String()
			return response(r, http.StatusOK, `{}`), nil
		})),
	)
	t.Cleanup(client.Close)

	if _, err := client.GetPokemon("pikachu"); err != nil {
		t.Fatal(err)
	}
	if want := "https://pokeapi.test/api/v2/pokemon/pikachu"; gotURL != want {
		t.Errorf("requested %q, want %q", gotURL, want)
	}
}

func TestDefaultBaseURL(t *testing.T) {
	client := NewClient(time.Second, time.Minute)
	t.Cleanup(client.Close)

	if client.baseURL != "https://pokeapi.co/api/v2" {
		t.Errorf("baseURL = %q, want the real PokeAPI", client.baseURL)
	}
}
