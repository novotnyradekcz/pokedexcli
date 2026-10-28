package repl

import "testing"

func TestFormatPokedex(t *testing.T) {
	want := "Your Pokedex:\n - pidgey\n - caterpie\n"
	if got := formatPokedex([]string{"pidgey", "caterpie"}); got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	want = "Your Pokedex is empty. Use the catch command to add Pokemon.\n"
	if got := formatPokedex(nil); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
