package repl

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/term"
)

const (
	prompt         = "Pokedex > "
	maxHistorySize = 100
)

// lineReader reads one line of user input. It returns io.EOF when the user
// ends the session (Ctrl-D / Ctrl-C in a terminal, or end of piped input).
type lineReader interface {
	ReadLine() (string, error)
}

// newLineReader returns a reader with arrow-key history when stdin and stdout
// are a terminal, and a plain line reader otherwise (e.g. piped input).
func newLineReader(out io.Writer) lineReader {
	stdin := int(os.Stdin.Fd())
	if term.IsTerminal(stdin) && term.IsTerminal(int(os.Stdout.Fd())) {
		return &terminalReader{
			fd: stdin,
			t: newTerminal(struct {
				io.Reader
				io.Writer
			}{os.Stdin, os.Stdout}),
		}
	}
	return &scannerReader{scanner: bufio.NewScanner(os.Stdin), out: out}
}

// newTerminal builds a line editor on rw. Up/down arrows cycle through the
// lines previously entered on it.
func newTerminal(rw io.ReadWriter) *term.Terminal {
	t := term.NewTerminal(rw, prompt)
	t.History = &history{}
	return t
}

// terminalReader reads lines from an interactive terminal.
type terminalReader struct {
	fd int
	t  *term.Terminal
}

func (r *terminalReader) ReadLine() (string, error) {
	// Raw mode is only needed while editing a line. Leaving it on would break
	// the output of commands, which print plain "\n" line endings.
	oldState, err := term.MakeRaw(r.fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(r.fd, oldState)

	// Some terminals report a 0x0 size; keep the default size in that case.
	if width, height, err := term.GetSize(r.fd); err == nil && width > 0 && height > 0 {
		r.t.SetSize(width, height)
	}
	return r.t.ReadLine()
}

// scannerReader reads lines from a non-interactive input and prints the prompt
// itself.
type scannerReader struct {
	scanner *bufio.Scanner
	out     io.Writer
}

func (r *scannerReader) ReadLine() (string, error) {
	fmt.Fprint(r.out, prompt)
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return "", err
		}
		return "", io.EOF
	}
	return r.scanner.Text(), nil
}

// history implements term.History. It ignores empty lines and repeats of the
// previous line, and keeps only the most recent maxHistorySize lines.
type history struct {
	entries []string // oldest first
}

func (h *history) Add(entry string) {
	if entry == "" {
		return
	}
	if n := len(h.entries); n > 0 && h.entries[n-1] == entry {
		return
	}
	h.entries = append(h.entries, entry)
	if len(h.entries) > maxHistorySize {
		h.entries = h.entries[len(h.entries)-maxHistorySize:]
	}
}

func (h *history) Len() int {
	return len(h.entries)
}

// At returns the idx-th most recent entry; At(0) is the newest.
func (h *history) At(idx int) string {
	return h.entries[len(h.entries)-1-idx]
}
