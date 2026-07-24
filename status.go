package main

import (
	"fmt"
	"strconv"
	"strings"
)

// statusExecArgs builds the git arguments used to gather machine-readable
// status. It always asks for porcelain v2 + branch info, and forwards any
// extra user flags (e.g. -uall, --ignored) that follow `status`.
func statusExecArgs(userExtra []string) []string {
	args := []string{"status", "--porcelain=v2", "--branch"}
	return append(args, userExtra...)
}

// repoStatus is the parsed, condensed state of one repository.
type repoStatus struct {
	name      string
	branch    string
	ahead     int
	behind    int
	staged    int
	modified  int
	deleted   int
	untracked int
	conflicts int
	detached  bool
	failed    bool
	failMsg   string
}

func (s repoStatus) clean() bool {
	return s.staged == 0 && s.modified == 0 && s.deleted == 0 &&
		s.untracked == 0 && s.conflicts == 0
}

// renderStatus runs the status results through a compact per-repo table.
// Results are re-derived from the porcelain v2 output that main injected for
// the status path; renderStatus is only reached when --full is off.
func renderStatus(results []result, opts options) int {
	p := newPalette(colorEnabled(opts))
	statuses := make([]repoStatus, len(results))
	nameWidth := 0
	for i, r := range results {
		statuses[i] = parseStatus(r)
		if len(statuses[i].name) > nameWidth {
			nameWidth = len(statuses[i].name)
		}
	}
	if nameWidth > 48 {
		nameWidth = 48
	}

	var dirty, clean, failed int
	for _, s := range statuses {
		line := formatStatusLine(s, nameWidth, p)
		fmt.Println(line)
		switch {
		case s.failed:
			failed++
		case s.clean() && s.ahead == 0 && s.behind == 0:
			clean++
		default:
			dirty++
		}
	}

	fmt.Println(summaryLine(len(statuses), clean, dirty, failed, p))
	if failed > 0 {
		return 1
	}
	return 0
}

// formatStatusLine renders a single repo as one compact line:
//
//	name                 branch     ↑2 ↓1   +3 ~1 ?2
func formatStatusLine(s repoStatus, nameWidth int, p palette) string {
	var b strings.Builder

	// Name column. Truncate on rune boundaries so multibyte names aren't
	// split into invalid UTF-8.
	nameCol := s.name
	if rs := []rune(nameCol); len(rs) > nameWidth {
		nameCol = "…" + string(rs[len(rs)-nameWidth+1:])
	}
	fmt.Fprintf(&b, "%s%-*s%s  ", p.bold, nameWidth, nameCol, p.reset)

	if s.failed {
		fmt.Fprintf(&b, "%s✖ %s%s", p.red, s.failMsg, p.reset)
		return b.String()
	}

	// Branch column (fixed-ish width for alignment).
	branch := s.branch
	if s.detached {
		branch = "(" + branch + ")"
	}
	fmt.Fprintf(&b, "%s%-16s%s", p.cyan, branch, p.reset)

	// Ahead/behind.
	ab := ""
	if s.ahead > 0 {
		ab += fmt.Sprintf("%s↑%d%s ", p.yellow, s.ahead, p.reset)
	}
	if s.behind > 0 {
		ab += fmt.Sprintf("%s↓%d%s ", p.yellow, s.behind, p.reset)
	}
	fmt.Fprintf(&b, " %-10s", stripPad(ab, 10))

	// Worktree state.
	if s.clean() {
		fmt.Fprintf(&b, " %s✔ clean%s", p.green, p.reset)
	} else {
		var parts []string
		if s.conflicts > 0 {
			parts = append(parts, fmt.Sprintf("%s✖%d%s", p.red, s.conflicts, p.reset))
		}
		if s.staged > 0 {
			parts = append(parts, fmt.Sprintf("%s+%d%s", p.green, s.staged, p.reset))
		}
		if s.modified > 0 {
			parts = append(parts, fmt.Sprintf("%s~%d%s", p.yellow, s.modified, p.reset))
		}
		if s.deleted > 0 {
			parts = append(parts, fmt.Sprintf("%s-%d%s", p.red, s.deleted, p.reset))
		}
		if s.untracked > 0 {
			parts = append(parts, fmt.Sprintf("%s?%d%s", p.dim, s.untracked, p.reset))
		}
		fmt.Fprintf(&b, " %s", strings.Join(parts, " "))
	}

	return b.String()
}

// stripPad pads s to width based on its *visible* length, ignoring ANSI
// escape codes so coloured columns still line up.
func stripPad(s string, width int) string {
	visible := visibleLen(s)
	if visible >= width {
		return s
	}
	return s + strings.Repeat(" ", width-visible)
}

func visibleLen(s string) int {
	n, inEsc := 0, false
	for _, r := range s {
		switch {
		case r == '\x1b':
			inEsc = true
		case inEsc && r == 'm':
			inEsc = false
		case inEsc:
			// skip
		default:
			n++
		}
	}
	return n
}

func summaryLine(total, clean, dirty, failed int, p palette) string {
	parts := []string{fmt.Sprintf("%d repos", total)}
	if clean > 0 {
		parts = append(parts, fmt.Sprintf("%s%d clean%s", p.green, clean, p.reset))
	}
	if dirty > 0 {
		parts = append(parts, fmt.Sprintf("%s%d need attention%s", p.yellow, dirty, p.reset))
	}
	if failed > 0 {
		parts = append(parts, fmt.Sprintf("%s%d failed%s", p.red, failed, p.reset))
	}
	return fmt.Sprintf("%s—%s %s", p.dim, p.reset, strings.Join(parts, ", "))
}

// parseStatus derives a repoStatus from a git result. rgit runs status with
// porcelain v2 + branch info (see statusExecArgs); if the output is not in
// that format (because the user forced their own flags) the repo simply
// reads as clean, which the caller can still distinguish from a failure.
func parseStatus(r result) repoStatus {
	s := repoStatus{name: sanitizeLine(r.name)}
	if r.runErr != nil {
		s.failed = true
		s.failMsg = "git not runnable"
		return s
	}
	if r.exitCode != 0 {
		s.failed = true
		s.failMsg = sanitizeLine(firstLine(r.stderr))
		if s.failMsg == "" {
			s.failMsg = "git exited " + strconv.Itoa(r.exitCode)
		}
		return s
	}

	for _, line := range strings.Split(r.stdout, "\n") {
		if line == "" {
			continue
		}
		switch line[0] {
		case '#':
			parseBranchHeader(line, &s)
		case '1', '2':
			// Changed tracked entry: field 2 is the XY status pair.
			xy := field(line, 1)
			applyXY(xy, &s)
		case 'u':
			s.conflicts++
		case '?':
			s.untracked++
		case '!':
			// ignored; skip
		}
	}
	if s.branch == "" {
		s.branch = "?"
	}
	return s
}

func parseBranchHeader(line string, s *repoStatus) {
	switch {
	case strings.HasPrefix(line, "# branch.head "):
		head := strings.TrimPrefix(line, "# branch.head ")
		if head == "(detached)" {
			s.detached = true
			s.branch = "detached"
		} else {
			s.branch = head
		}
	case strings.HasPrefix(line, "# branch.ab "):
		// Format: "# branch.ab +A -B"
		fields := strings.Fields(strings.TrimPrefix(line, "# branch.ab "))
		for _, f := range fields {
			if len(f) < 2 {
				continue
			}
			n, _ := strconv.Atoi(f[1:])
			if f[0] == '+' {
				s.ahead = n
			} else if f[0] == '-' {
				s.behind = n
			}
		}
	}
}

// applyXY interprets the two-character XY code from a porcelain v2 entry.
// X = staged (index) state, Y = unstaged (worktree) state.
func applyXY(xy string, s *repoStatus) {
	if len(xy) < 2 {
		return
	}
	x, y := xy[0], xy[1]
	if x != '.' {
		if x == 'D' {
			s.deleted++
		} else {
			s.staged++
		}
	}
	if y != '.' {
		if y == 'D' {
			s.deleted++
		} else {
			s.modified++
		}
	}
}

// field returns the i-th space-separated field of a porcelain line.
func field(line string, i int) string {
	f := strings.Fields(line)
	if i < len(f) {
		return f[i]
	}
	return ""
}

func firstLine(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return s[:idx]
	}
	return s
}
