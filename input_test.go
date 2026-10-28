package main

import (
	"bufio"
	"bytes"
	"io"
	"strings"
	"testing"
)

const (
	keyUp   = "\x1b[A"
	keyDown = "\x1b[B"
)

// rw feeds scripted keystrokes to a terminal and discards what it draws.
type rw struct {
	io.Reader
	io.Writer
}

func readLines(t *testing.T, keys string, n int) []string {
	t.Helper()
	term := newTerminal(rw{strings.NewReader(keys), &bytes.Buffer{}})

	var lines []string
	for i := 0; i < n; i++ {
		line, err := term.ReadLine()
		if err != nil {
			t.Fatalf("ReadLine %d: %v", i, err)
		}
		lines = append(lines, line)
	}
	return lines
}

func TestArrowKeysCycleThroughHistory(t *testing.T) {
	keys := "map\r" + // 1: map
		"mapb\r" + // 2: mapb
		keyUp + keyUp + "\r" + // 3: up, up -> map
		keyUp + keyUp + keyDown + "\r" + // 4: up, up, down -> map (history: map mapb map)
		keyUp + "\r" // 5: up -> map

	got := readLines(t, keys, 5)
	want := []string{"map", "mapb", "map", "map", "map"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q (all: %q)", i+1, got[i], want[i], got)
		}
	}
}

func TestUpArrowRecallsMostRecentThenOlder(t *testing.T) {
	keys := "help\r" + "explore pastoria-city-area\r" + keyUp + "\r" + keyUp + keyUp + "\r"
	got := readLines(t, keys, 4)

	if got[2] != "explore pastoria-city-area" {
		t.Errorf("first up should recall the most recent command, got %q", got[2])
	}
	// history is now: help, explore..., (explore... is a repeat, not re-added)
	if got[3] != "help" {
		t.Errorf("second up should recall the older command, got %q", got[3])
	}
}

func TestHistorySkipsEmptyAndRepeatedLines(t *testing.T) {
	h := &history{}
	for _, line := range []string{"map", "", "map", "mapb", "mapb"} {
		h.Add(line)
	}

	if h.Len() != 2 {
		t.Fatalf("Len = %d, want 2", h.Len())
	}
	if h.At(0) != "mapb" || h.At(1) != "map" {
		t.Errorf("At(0), At(1) = %q, %q; want mapb, map", h.At(0), h.At(1))
	}
}

func TestHistoryIsBounded(t *testing.T) {
	h := &history{}
	for i := 0; i < maxHistorySize+10; i++ {
		h.Add(strings.Repeat("x", i+1))
	}

	if h.Len() != maxHistorySize {
		t.Fatalf("Len = %d, want %d", h.Len(), maxHistorySize)
	}
	if got := len(h.At(0)); got != maxHistorySize+10 {
		t.Errorf("newest entry has length %d, want %d", got, maxHistorySize+10)
	}
}

func TestScannerReaderReturnsEOF(t *testing.T) {
	r := &scannerReader{scanner: bufio.NewScanner(strings.NewReader("one\ntwo\n"))}

	for _, want := range []string{"one", "two"} {
		got, err := r.ReadLine()
		if err != nil || got != want {
			t.Fatalf("ReadLine = %q, %v; want %q, nil", got, err, want)
		}
	}
	if _, err := r.ReadLine(); err != io.EOF {
		t.Errorf("expected io.EOF at end of input, got %v", err)
	}
}
