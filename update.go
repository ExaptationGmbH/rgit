package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// goModulePath is where `go install` fetches rgit from.
const goModulePath = "github.com/ExaptationGmbH/rgit"

// selfUpdate upgrades the installed rgit binary using whatever installed it:
// Homebrew if the binary lives under a brew Cellar, otherwise `go install`.
// If neither is detected it prints the manual options and fails.
func selfUpdate(stdout, stderr io.Writer) int {
	exe := resolvedExecutable()

	switch {
	case isHomebrew(exe):
		fmt.Fprintln(stdout, "rgit: updating via Homebrew…")
		return runUpdater(stdout, stderr,
			[]string{"brew", "update"},
			[]string{"brew", "upgrade", "ExaptationGmbH/tap/rgit"})
	case hasExec("go"):
		fmt.Fprintln(stdout, "rgit: updating via `go install`…")
		return runUpdater(stdout, stderr,
			[]string{"go", "install", goModulePath + "@latest"})
	default:
		fmt.Fprintln(stderr, "rgit: can't detect how this copy was installed. Update manually:")
		fmt.Fprintln(stderr, "  Homebrew:  brew upgrade ExaptationGmbH/tap/rgit")
		fmt.Fprintln(stderr, "  Go:        go install "+goModulePath+"@latest")
		fmt.Fprintln(stderr, "  Binaries:  https://github.com/ExaptationGmbH/rgit/releases/latest")
		return 1
	}
}

// resolvedExecutable returns the real path of the running binary, following
// symlinks (Homebrew installs a symlink in <prefix>/bin -> ../Cellar/…).
func resolvedExecutable() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if real, err := filepath.EvalSymlinks(exe); err == nil {
		return real
	}
	return exe
}

// isHomebrew reports whether exe is a Homebrew-managed binary and brew is
// available to drive the upgrade.
func isHomebrew(exe string) bool {
	if !strings.Contains(exe, "/Cellar/") && !strings.Contains(exe, "/homebrew/") {
		return false
	}
	return hasExec("brew")
}

func hasExec(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// runUpdater runs each command in order, streaming output, and stops on the
// first failure.
func runUpdater(stdout, stderr io.Writer, cmds ...[]string) int {
	for _, c := range cmds {
		fmt.Fprintf(stdout, "  $ %s\n", strings.Join(c, " "))
		cmd := exec.Command(c[0], c[1:]...)
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(stderr, "rgit: update failed: %v\n", err)
			return 1
		}
	}
	fmt.Fprintln(stdout, "rgit: up to date.")
	return 0
}
