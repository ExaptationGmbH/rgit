package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseStatusPorcelainV2(t *testing.T) {
	out := "# branch.head main\n" +
		"# branch.ab +2 -1\n" +
		"1 M. N... 100644 100644 100644 aaa bbb staged.txt\n" +
		"1 .M N... 100644 100644 100644 aaa bbb modified.txt\n" +
		"1 D. N... 100644 000000 000000 aaa bbb deleted.txt\n" +
		"u UU N... ...\n" +
		"? untracked.txt\n" +
		"! ignored.txt\n"

	s := parseStatus(result{name: "r", stdout: out})
	if s.branch != "main" {
		t.Errorf("branch = %q, want main", s.branch)
	}
	if s.ahead != 2 || s.behind != 1 {
		t.Errorf("ahead/behind = %d/%d, want 2/1", s.ahead, s.behind)
	}
	if s.staged != 1 {
		t.Errorf("staged = %d, want 1", s.staged)
	}
	if s.modified != 1 {
		t.Errorf("modified = %d, want 1", s.modified)
	}
	if s.deleted != 1 {
		t.Errorf("deleted = %d, want 1", s.deleted)
	}
	if s.conflicts != 1 {
		t.Errorf("conflicts = %d, want 1", s.conflicts)
	}
	if s.untracked != 1 {
		t.Errorf("untracked = %d, want 1", s.untracked)
	}
	if s.clean() {
		t.Error("clean() = true, want false")
	}
}

func TestParseStatusCleanDetached(t *testing.T) {
	out := "# branch.head (detached)\n# branch.oid abc123\n"
	s := parseStatus(result{name: "r", stdout: out})
	if !s.detached {
		t.Error("expected detached")
	}
	if !s.clean() {
		t.Error("expected clean working tree")
	}
}

func TestParseStatusFailure(t *testing.T) {
	s := parseStatus(result{name: "r", exitCode: 128, stderr: "fatal: not a git repository"})
	if !s.failed {
		t.Fatal("expected failed")
	}
	if s.failMsg == "" {
		t.Error("expected a failure message")
	}
}

func TestDiscoverPrunesNestedAndSkipsNonRepos(t *testing.T) {
	root := t.TempDir()
	mkRepo := func(rel string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Join(p, ".git"), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	mkRepo("a")
	mkRepo("nested/b")
	mkRepo("a/submodule") // inside repo a → must NOT be reported separately
	os.MkdirAll(filepath.Join(root, "plain"), 0o755)

	repos, err := discover(root, -1)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 2 {
		t.Fatalf("found %d repos, want 2: %v", len(repos), repos)
	}
}

func TestDiscoverDepthLimit(t *testing.T) {
	root := t.TempDir()
	os.MkdirAll(filepath.Join(root, "top", ".git"), 0o755)
	os.MkdirAll(filepath.Join(root, "a/b/deep", ".git"), 0o755)

	repos, err := discover(root, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(repos) != 1 {
		t.Fatalf("depth 1 found %d repos, want 1: %v", len(repos), repos)
	}
}

func TestParseArgsSplitsRgitAndGit(t *testing.T) {
	opts, gitArgs, err := parseArgs([]string{"-j", "4", "--no-color", "status", "-s"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.jobs != 4 || !opts.noColor {
		t.Errorf("opts not parsed: %+v", opts)
	}
	if len(gitArgs) != 2 || gitArgs[0] != "status" || gitArgs[1] != "-s" {
		t.Errorf("gitArgs = %v, want [status -s]", gitArgs)
	}
}

func TestHelpCommandDoesNotRunGit(t *testing.T) {
	// `rgit help` must render usage (exit 0) rather than shelling out to
	// git in every repo. run() with "help" should not touch the filesystem.
	if code := run([]string{"help"}); code != 0 {
		t.Errorf("run([help]) = %d, want 0", code)
	}
}

func TestVisibleLenIgnoresANSI(t *testing.T) {
	if got := visibleLen("\x1b[31mred\x1b[0m"); got != 3 {
		t.Errorf("visibleLen = %d, want 3", got)
	}
}

func TestDiagnosticLine(t *testing.T) {
	in := "hint: something\nfatal: No configured push destination.\nhint: more\n"
	if got := diagnosticLine(in); got != "fatal: No configured push destination." {
		t.Errorf("diagnosticLine = %q", got)
	}
}
