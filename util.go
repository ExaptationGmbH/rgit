package main

import (
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

// termWidth returns the terminal width in columns for stdout, falling back to
// the COLUMNS env var and then 80 when stdout isn't a terminal.
func termWidth() int {
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 0 {
		return w
	}
	if c, err := strconv.Atoi(os.Getenv("COLUMNS")); err == nil && c > 0 {
		return c
	}
	return 80
}

// currentDir returns the process working directory, or "" if unavailable.
func currentDir() string {
	if d, err := os.Getwd(); err == nil {
		return d
	}
	return ""
}

// sanitizeLine strips control characters and ANSI escape sequences from a
// string so that untrusted content — repo directory names and git output
// from repositories rgit discovered — cannot inject terminal escape codes
// (cursor moves, title changes, colour resets) into the condensed views.
// Tabs become single spaces; C0/C1 controls and DEL are dropped. The raw
// --full view intentionally does not go through this (it's an explicit
// opt-in to see verbatim git output, like running git yourself).
func sanitizeLine(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		switch {
		case r == '\t':
			b.WriteByte(' ')
		case r == 0x1b: // ESC — drop so escape sequences can't form
			continue
		case r < 0x20: // other C0 controls (incl. CR/LF)
			continue
		case r == 0x7f: // DEL
			continue
		case r >= 0x80 && r <= 0x9f: // C1 controls
			continue
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// capBuffer is a byte buffer that stops growing past a cap. git output for a
// whole tree of repos can be large; since rgit only ever shows a condensed
// view we never need to retain megabytes per repo.
type capBuffer struct {
	buf     []byte
	dropped bool
}

const capBufferMax = 1 << 20 // 1 MiB per stream per repo

func (c *capBuffer) Write(p []byte) (int, error) {
	if len(c.buf) < capBufferMax {
		room := capBufferMax - len(c.buf)
		if room >= len(p) {
			c.buf = append(c.buf, p...)
		} else {
			c.buf = append(c.buf, p[:room]...)
			c.dropped = true
		}
	} else {
		c.dropped = true
	}
	// Always report a full write so git never blocks or errors.
	return len(p), nil
}

func (c *capBuffer) String() string { return string(c.buf) }

// --- colour helpers -------------------------------------------------------

// palette holds the ANSI codes in use, or empty strings when colour is off.
type palette struct {
	reset, bold, dim, red, green, yellow, blue, cyan string
}

func newPalette(enabled bool) palette {
	if !enabled {
		return palette{}
	}
	return palette{
		reset:  "\x1b[0m",
		bold:   "\x1b[1m",
		dim:    "\x1b[2m",
		red:    "\x1b[31m",
		green:  "\x1b[32m",
		yellow: "\x1b[33m",
		blue:   "\x1b[34m",
		cyan:   "\x1b[36m",
	}
}

// colorEnabled decides whether to emit ANSI colour, honouring --no-color,
// the NO_COLOR convention, and whether stdout is a terminal.
func colorEnabled(opts options) bool {
	if opts.noColor {
		return false
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return isTerminal(os.Stdout)
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}
