package main

import (
	"errors"
	"io"
	"strconv"
)

// options holds parsed rgit-level flags. Everything rgit does not recognise
// is passed straight through to git.
type options struct {
	root        string
	maxDepth    int  // -1 means unlimited
	jobs        int  // concurrent git invocations
	full        bool // show full git output instead of the condensed view
	noColor     bool
	showHelp    bool
	showVersion bool
}

// parseArgs splits argv into rgit options and the git command/arguments.
//
// rgit-specific flags must appear before the git command. The first
// non-flag token is treated as the git subcommand, and everything from
// there on is forwarded to git verbatim. This keeps the pass-through
// honest: `rgit -j8 status -s` runs `git status -s` with 8 workers.
func parseArgs(argv []string) (options, []string, error) {
	opts := options{
		root:     ".",
		maxDepth: -1,
		jobs:     0, // 0 -> auto (see execAll)
	}

	i := 0
	for ; i < len(argv); i++ {
		arg := argv[i]
		switch {
		case arg == "-h" || arg == "--help":
			opts.showHelp = true
			return opts, nil, nil
		case arg == "-v" || arg == "--version":
			opts.showVersion = true
			return opts, nil, nil
		case arg == "--full" || arg == "-f":
			opts.full = true
		case arg == "--no-color":
			opts.noColor = true
		case arg == "-C" || arg == "--dir":
			v, ok := next(argv, &i)
			if !ok {
				return opts, nil, errors.New(arg + " requires a directory")
			}
			opts.root = v
		case arg == "--depth":
			v, ok := next(argv, &i)
			if !ok {
				return opts, nil, errors.New("--depth requires a number")
			}
			n, err := strconv.Atoi(v)
			if err != nil {
				return opts, nil, errors.New("invalid --depth: " + v)
			}
			opts.maxDepth = n
		case arg == "-j" || arg == "--jobs":
			v, ok := next(argv, &i)
			if !ok {
				return opts, nil, errors.New(arg + " requires a number")
			}
			n, err := strconv.Atoi(v)
			if err != nil || n < 1 {
				return opts, nil, errors.New("invalid job count: " + v)
			}
			opts.jobs = n
		case arg == "--":
			// Explicit end of rgit flags; the rest is the git command.
			i++
			return opts, argv[i:], nil
		default:
			// First non-flag token: this and everything after is git's.
			return opts, argv[i:], nil
		}
	}
	return opts, nil, nil
}

// next advances the index and returns the following argument.
func next(argv []string, i *int) (string, bool) {
	if *i+1 >= len(argv) {
		return "", false
	}
	*i++
	return argv[*i], true
}

func printUsage(w io.Writer) {
	io.WriteString(w, `rgit — run a git command across every repo beneath the current directory

USAGE:
    rgit [rgit-flags] <git-command> [git-args...]

rgit is a 1:1 pass-through to git: ANY git command works and runs in every
repo found recursively below the current directory. The commands below are the
ones that make the most sense to fan out across many repos.

OVERVIEW (read-only — safe to run anywhere):
    rgit status              Compact status table: branch, ahead/behind, changes
    rgit branch              List branches (add --show-current for just the tip)
    rgit log -1              Most recent commit per repo
    rgit diff --stat         Uncommitted-change summary per repo
    rgit stash list          Find forgotten stashes in any repo
    rgit remote -v           Confirm every repo's remotes

SYNC (network):
    rgit fetch --all         Fetch everything; never touches the working tree
    rgit pull --ff-only      Fast-forward every repo (safe for bulk runs)

MAINTENANCE:
    rgit gc                  Compact/optimise every repo
    rgit remote prune origin Drop stale remote-tracking branches

Destructive or per-repo commands (commit, push, merge, rebase, reset) still
work, but think before fanning them out — see the project's
multi-repo-git-commands.md for what's safe to run in bulk.

EXAMPLES:
    rgit status              Compact status table for every repo found
    rgit fetch --all         Fetch in every repo, summarised
    rgit --full log -1       Full git output (no condensing) of last commit
    rgit -j 16 status        Use 16 concurrent workers

RGIT FLAGS (must come before the git command):
    -C, --dir <path>   Root directory to scan (default: current directory)
    --depth <n>        Max directory depth to descend (default: unlimited)
    -j, --jobs <n>     Concurrent git invocations (default: auto)
    -f, --full         Show full git output instead of the condensed view
    --no-color         Disable ANSI colours
    -h, --help         Show this help
    -v, --version      Show version

Anything that is not an rgit flag is passed straight through to git.
`)
}
