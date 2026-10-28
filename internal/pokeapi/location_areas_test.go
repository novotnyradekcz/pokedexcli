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

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func newTestClient(requests *atomic.Int32, status int, body string) Client {
	client := NewClient(5*time.Second, time.Minute)
	client.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests.Add(1)
		return &http.Response{
			StatusCode: status,
			Status:     http.StatusText(status),
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
	})
	return client
}

func TestListLocationAreasUsesCache(t *testing.T) {
	var requests atomic.Int32
	client := newTestClient(&requests, http.StatusOK,
		`{"count":1,"next":null,"previous":null,"results":[{"name":"canalave-city-area","url":"u"}]}`)
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

// Mirrors the real API: page 2's Previous URL must be the same URL that was
// used to fetch page 1, so going back is served from the cache.
func TestPreviousPageIsServedFromCache(t *testing.T) {
	const page2 = baseURL + "/location-area/?offset=20&limit=20"
	pages := map[string]string{
		firstLocationAreaURL: `{"next":"` + page2 + `","previous":null,"results":[{"name":"a"}]}`,
		page2:                `{"next":null,"previous":"` + firstLocationAreaURL + `","results":[{"name":"b"}]}`,
	}

	var requests atomic.Int32
	client := NewClient(5*time.Second, time.Minute)
	client.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests.Add(1)
		body, ok := pages[r.URL.String()]
		if !ok {
			t.Errorf("unexpected request to %s", r.URL)
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
			Request:    r,
		}, nil
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
	var requests atomic.Int32
	var gotURL string
	client := NewClient(5*time.Second, time.Minute)
	client.httpClient.Transport = roundTripFunc(func(r *http.Request) (*http.Response, error) {
		requests.Add(1)
		gotURL = r.URL.String()
		return &http.Response{
			StatusCode: http.StatusOK,
			Body: io.NopCloser(strings.NewReader(
				`{"name":"pastoria-city-area","pokemon_encounters":[` +
					`{"pokemon":{"name":"tentacool","url":"u1"}},` +
					`{"pokemon":{"name":"magikarp","url":"u2"}}]}`)),
			Header:  make(http.Header),
			Request: r,
		}, nil
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

	if want := baseURL + "/location-area/pastoria-city-area"; gotURL != want {
		t.Errorf("requested %q, want %q", gotURL, want)
	}
	if got := requests.Load(); got != 1 {
		t.Errorf("expected 1 HTTP request, got %d", got)
	}
}

func TestGetLocationAreaNotFound(t *testing.T) {
	var requests atomic.Int32
	client := newTestClient(&requests, http.StatusNotFound, "Not Found")

	_, err := client.GetLocationArea("nowhere")
	var statusErr *StatusError
	if !errors.As(err, &statusErr) || statusErr.StatusCode != http.StatusNotFound {
		t.Fatalf("expected a 404 StatusError, got %v", err)
	}
}

func TestListLocationAreasDoesNotCacheErrors(t *testing.T) {
	var requests atomic.Int32
	client := newTestClient(&requests, http.StatusInternalServerError, "boom")
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
