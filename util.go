package main

import (
	"os"
)

// currentDir returns the process working directory, or "" if unavailable.
func currentDir() string {
	if d, err := os.Getwd(); err == nil {
		return d
	}
	return ""
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
