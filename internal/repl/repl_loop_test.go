package repl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/novotnyradekcz/pokedexcli/internal/pokeapi"
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

	loop(t.Context(), cfg, &scriptedReader{lines: []string{"map", "  ", "MAPB", "explore pastoria-city-area"}})

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

	loop(t.Context(), cfg, &scriptedReader{lines: []string{"dance", "explore", "catch nobody"}})

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

	loop(t.Context(), cfg, reader)

	if got, want := out.String(), "Closing the Pokedex... Goodbye!\n"; got != want {
		t.Errorf("printed %q, want %q", got, want)
	}
	if len(reader.lines) != 2 {
		t.Errorf("expected the loop to stop reading after exit, %d lines left", len(reader.lines))
	}
}

func TestLoopPrintsReaderErrors(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())

	loop(t.Context(), cfg, &scriptedReader{err: errors.New("disk on fire")})

	if got, want := out.String(), "\ndisk on fire\n"; got != want {
		t.Errorf("printed %q, want %q", got, want)
	}
}

func TestLoopPassesContextToCommands(t *testing.T) {
	api := newFakeAPI()
	cfg, _ := newTestConfig(api)
	ctx := context.WithValue(t.Context(), ctxKey{}, "from-the-loop")

	loop(ctx, cfg, &scriptedReader{lines: []string{"catch pikachu"}})

	if api.ctxValue != "from-the-loop" {
		t.Errorf("the API saw context value %v, want the loop's context", api.ctxValue)
	}
}

func TestLoopReportsCancelledCommands(t *testing.T) {
	api := newFakeAPI()
	api.err = context.Canceled // as if the user pressed Ctrl-C during the request
	cfg, out := newTestConfig(api)

	loop(t.Context(), cfg, &scriptedReader{lines: []string{"map", "help"}})

	if got := out.String(); !strings.HasPrefix(got, "Canceled\n") {
		t.Errorf("printed %q, want it to start with %q", got, "Canceled\n")
	}
	if got := out.String(); !strings.Contains(got, "Welcome to the Pokedex!") {
		t.Errorf("the loop must keep running after a cancelled command; printed %q", got)
	}
}

func TestLoopStopsWhenContextIsCancelled(t *testing.T) {
	cfg, out := newTestConfig(newFakeAPI())
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	reader := &scriptedReader{lines: []string{"map"}}

	loop(ctx, cfg, reader)

	if got := out.String(); got != "" {
		t.Errorf("printed %q, want nothing", got)
	}
	if len(reader.lines) != 1 {
		t.Error("the loop must not read input once its context is cancelled")
	}
}

// Sends a real interrupt signal to this process while a command is running.
// The command's context must be cancelled, and the process must survive.
func TestRunCommandIsCancelledByInterrupt(t *testing.T) {
	started := make(chan struct{})
	cmd := cliCommand{callback: func(ctx context.Context, _ *config, _ ...string) error {
		close(started)
		<-ctx.Done()
		return ctx.Err()
	}}

	errc := make(chan error, 1)
	go func() { // ends when the command returns, which the interrupt below causes
		errc <- runCommand(t.Context(), &config{}, cmd, nil)
	}()

	// runCommand starts listening for interrupts before calling the command, so
	// once the command has started the signal cannot be missed.
	<-started
	proc, err := os.FindProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	if err := proc.Signal(os.Interrupt); err != nil {
		t.Skipf("cannot send an interrupt on this platform: %v", err)
	}

	select {
	case err := <-errc:
		if !errors.Is(err, context.Canceled) {
			t.Errorf("error = %v, want context.Canceled", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the command was not cancelled by the interrupt")
	}
}

func TestDescribeAPIErrorKeepsTheCause(t *testing.T) {
	cause := &pokeapi.StatusError{StatusCode: http.StatusNotFound, Status: "404 Not Found"}

	err := describeAPIError(fmt.Errorf("get: %w", cause), "pokemon", "missingno")

	if err.Error() != `pokemon "missingno" not found` {
		t.Errorf("error = %q", err)
	}
	var statusErr *pokeapi.StatusError
	if !errors.As(err, &statusErr) || statusErr != cause {
		t.Error("expected the original StatusError to stay reachable with errors.As")
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
