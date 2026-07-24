package main

import (
	"fmt"
	"strings"
)

// renderGeneric prints a condensed per-repo view for any git command other
// than the specially-handled ones. For each repo it shows a status glyph,
// the repo name, and a one-line summary of git's output. With --full the
// complete output is printed under each repo instead.
func renderGeneric(results []result, opts options) int {
	p := newPalette(colorEnabled(opts))

	var ok, failed int
	for _, r := range results {
		if opts.full {
			printFull(r, p)
		} else {
			printCondensed(r, p)
		}
		if r.exitCode == 0 && r.runErr == nil {
			ok++
		} else {
			failed++
		}
	}

	// Summary footer.
	parts := []string{fmt.Sprintf("%d repos", len(results))}
	if ok > 0 {
		parts = append(parts, fmt.Sprintf("%s%d ok%s", p.green, ok, p.reset))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%s%d failed%s", p.red, failed, p.reset))
	}
	fmt.Printf("%s—%s %s\n", p.dim, p.reset, strings.Join(parts, ", "))

	if failed > 0 {
		return 1
	}
	return 0
}

// printCondensed emits one line per repo: glyph + name + summary. Both the
// repo name and the summarised git output are sanitized because they come
// from repositories rgit discovered and may contain terminal escapes.
func printCondensed(r result, p palette) {
	glyph, color := statusGlyph(r, p)

	summary := sanitizeLine(summarise(r))
	nameField := fmt.Sprintf("%s%s%s", p.bold, sanitizeLine(r.name), p.reset)

	if summary == "" {
		fmt.Printf("%s%s%s %s\n", color, glyph, p.reset, nameField)
		return
	}
	fmt.Printf("%s%s%s %s  %s%s%s\n",
		color, glyph, p.reset, nameField, p.dim, summary, p.reset)
}

// printFull prints a header per repo followed by git's complete output.
func printFull(r result, p palette) {
	glyph, color := statusGlyph(r, p)
	fmt.Printf("%s%s%s %s%s%s", color, glyph, p.reset, p.bold, r.name, p.reset)
	if r.exitCode != 0 {
		fmt.Printf(" %s(exit %d)%s", p.red, r.exitCode, p.reset)
	}
	fmt.Println()

	out := strings.TrimRight(r.stdout, "\n")
	errOut := strings.TrimRight(r.stderr, "\n")
	for _, block := range []string{out, errOut} {
		if block == "" {
			continue
		}
		for _, line := range strings.Split(block, "\n") {
			fmt.Printf("    %s\n", line)
		}
	}
}

func statusGlyph(r result, p palette) (string, string) {
	if r.runErr != nil || r.exitCode < 0 {
		return "✖", p.red
	}
	if r.exitCode != 0 {
		return "✖", p.red
	}
	return "✔", p.green
}

// summarise reduces git's output to a single informative line.
//
// The heuristic: prefer the last non-empty line of stdout (git's most
// meaningful line for many commands — "Already up to date.", the summary of
// a pull, the new HEAD of a commit); fall back to stderr; note how many
// more lines were hidden.
func summarise(r result) string {
	if r.runErr != nil {
		return r.runErr.Error()
	}

	// On failure, the most useful line is git's diagnostic, so prefer a
	// line beginning with fatal:/error: from stderr.
	if r.exitCode != 0 {
		if msg := diagnosticLine(r.stderr); msg != "" {
			return truncate(msg, 100)
		}
	}

	lines := nonEmptyLines(r.stdout)
	source := "out"
	if len(lines) == 0 {
		lines = nonEmptyLines(r.stderr)
		source = "err"
	}
	if len(lines) == 0 {
		if r.exitCode == 0 {
			return "" // clean, nothing to say
		}
		return fmt.Sprintf("exit %d", r.exitCode)
	}

	last := lines[len(lines)-1]
	last = truncate(strings.TrimSpace(last), 100)

	if len(lines) > 1 {
		return fmt.Sprintf("%s  (+%d more %s lines)", last, len(lines)-1, source)
	}
	return last
}

// diagnosticLine returns the first stderr line that looks like a git error
// ("fatal: ..." or "error: ..."), or "" if none is found.
func diagnosticLine(stderr string) string {
	for _, l := range strings.Split(stderr, "\n") {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "fatal:") || strings.HasPrefix(t, "error:") {
			return t
		}
	}
	return ""
}

func nonEmptyLines(s string) []string {
	var out []string
	for _, l := range strings.Split(s, "\n") {
		if strings.TrimSpace(l) != "" {
			out = append(out, l)
		}
	}
	return out
}

func truncate(s string, max int) string {
	if len([]rune(s)) <= max {
		return s
	}
	r := []rune(s)
	return string(r[:max-1]) + "…"
}
