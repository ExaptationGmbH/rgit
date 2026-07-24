// Command rgit runs a git command across every git repository found
// recursively beneath the current directory, and prints a condensed,
// at-a-glance overview of the results.
//
// It is a 1:1 pass-through to git: `rgit status` runs `git status` in each
// repo, `rgit fetch` runs `git fetch`, and so on. The value rgit adds is
// discovery (find all the repos) and summarisation (show less, not more).
package main

import (
	"fmt"
	"os"
)

// version is overridden at build time via -ldflags "-X main.version=...".
var version = "dev"

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(argv []string) int {
	opts, gitArgs, err := parseArgs(argv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "rgit: "+err.Error())
		return 2
	}
	// `rgit help` (as a bare command) is help, not `git help` per repo.
	if opts.showHelp || (len(gitArgs) > 0 && gitArgs[0] == "help") {
		printUsage(os.Stdout)
		return 0
	}
	if opts.showVersion {
		fmt.Println("rgit " + version)
		return 0
	}
	if len(gitArgs) == 0 {
		printUsage(os.Stderr)
		return 2
	}

	repos, err := discover(opts.root, opts.maxDepth)
	if err != nil {
		fmt.Fprintln(os.Stderr, "rgit: "+err.Error())
		return 1
	}
	if len(repos) == 0 {
		fmt.Fprintf(os.Stderr, "rgit: no git repositories found under %s\n", opts.root)
		return 1
	}

	// For the compact status table we need machine-readable output, so we
	// run porcelain v2 under the hood. --full keeps the user's literal args.
	execArgs := gitArgs
	compactStatus := gitArgs[0] == "status" && !opts.full
	if compactStatus {
		execArgs = statusExecArgs(gitArgs[1:])
	}

	results := execAll(repos, execArgs, opts.jobs)

	// The status subcommand gets a bespoke compact table; everything else
	// gets the generic condensed renderer.
	if compactStatus {
		return renderStatus(results, opts)
	}
	return renderGeneric(results, opts)
}
