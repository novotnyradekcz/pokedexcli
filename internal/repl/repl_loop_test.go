package repl

import (
	"errors"
	"io"
	"testing"
)

// scriptedReader is a lineReader that returns canned lines, then err (io.EOF
// by default) once they run out.
type scriptedReader struct {
	lines []string
	err   error
}

func (r *scriptedReader) ReadLine() (string, error) {
	if len(r.lines) == 0 {
		if r.err != nil {
			return "", r.err
		}
		return "", io.EOF
	}
	line := r.lines[0]
	r.lines = r.lines[1:]
	return line, nil
}

func TestLoopRunsCommandsAndEndsAtEOF(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	loop(cfg, &scriptedReader{lines: []string{"map", "  ", "MAPB", "explore pastoria-city-area"}})

	want := "area-1\narea-2\n" + // map
		"you're on the first page\n" + // MAPB is normalised to lower case
		"Exploring pastoria-city-area...\nFound Pokemon:\n - tentacool\n - magikarp\n" +
		"\n" // newline printed at end of input
	if got := out.String(); got != want {
		t.Errorf("printed %q, want %q", got, want)
	}
}

func TestLoopReportsUnknownCommandsAndErrors(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	loop(cfg, &scriptedReader{lines: []string{"dance", "explore", "catch nobody"}})

	want := "Unknown command\n" +
		"usage: explore <area_name>\n" +
		"Throwing a Pokeball at nobody...\n" +
		"pokemon \"nobody\" not found\n" +
		"\n"
	if got := out.String(); got != want {
		t.Errorf("printed %q, want %q", got, want)
	}
}

func TestLoopStopsAtExit(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())
	reader := &scriptedReader{lines: []string{"exit", "map", "map"}}

	loop(cfg, reader)

	if got, want := out.String(), "Closing the Pokedex... Goodbye!\n"; got != want {
		t.Errorf("printed %q, want %q", got, want)
	}
	if len(reader.lines) != 2 {
		t.Errorf("expected the loop to stop reading after exit, %d lines left", len(reader.lines))
	}
}

func TestLoopPrintsReaderErrors(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	loop(cfg, &scriptedReader{err: errors.New("disk on fire")})

	if got, want := out.String(), "\ndisk on fire\n"; got != want {
		t.Errorf("printed %q, want %q", got, want)
	}
}

func TestExitReturnsErrQuit(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	if err := run(t, cfg, "exit"); !errors.Is(err, errQuit) {
		t.Errorf("error = %v, want errQuit", err)
	}
	if got := out.String(); got != "Closing the Pokedex... Goodbye!\n" {
		t.Errorf("printed %q", got)
	}
}
