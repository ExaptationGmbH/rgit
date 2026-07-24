package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"time"
)

// maxJobs bounds concurrent git subprocesses regardless of how many repos are
// found or what -j the user asks for, so rgit can't fork-bomb the machine.
const maxJobs = 128

// result is the outcome of running git in a single repository.
type result struct {
	repo     string // absolute path
	name     string // display name (path relative to cwd, or basename)
	stdout   string
	stderr   string
	exitCode int
	runErr   error // git couldn't be launched at all
}

// execAll runs `git <args>` in every repo concurrently and returns the
// results sorted by display name. jobs controls the worker count; 0 means
// auto (a sensible multiple of the CPU count, capped). timeout, if > 0,
// caps how long any single git invocation may run.
func execAll(repos []string, gitArgs []string, jobs int, timeout time.Duration) []result {
	if jobs <= 0 {
		jobs = runtime.NumCPU() * 2
		if jobs > 16 {
			jobs = 16
		}
		if jobs < 1 {
			jobs = 1
		}
	}
	if jobs > maxJobs {
		jobs = maxJobs
	}
	if jobs > len(repos) {
		jobs = len(repos)
	}

	cwd := currentDir()
	env := gitEnv()

	results := make([]result, len(repos))
	sem := make(chan struct{}, jobs)
	var wg sync.WaitGroup

	for idx, repo := range repos {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, repo string) {
			defer wg.Done()
			defer func() { <-sem }()
			results[idx] = runGit(repo, cwd, gitArgs, env, timeout)
		}(idx, repo)
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].name < results[j].name
	})
	return results
}

// runGit executes git in a single repo. It uses `git -C <repo>` so we never
// have to change the process working directory. Stdin is left disconnected
// (/dev/null) and a hardened environment is applied so git can never block
// the whole run waiting for interactive input.
func runGit(repo, cwd string, gitArgs, env []string, timeout time.Duration) result {
	res := result{repo: repo, name: displayName(cwd, repo)}

	args := append([]string{"-C", repo}, gitArgs...)

	var (
		cmd    *exec.Cmd
		ctx    context.Context
		cancel context.CancelFunc
	)
	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
		cmd = exec.CommandContext(ctx, "git", args...)
	} else {
		cmd = exec.Command("git", args...)
	}
	cmd.Env = env
	cmd.Stdin = nil // -> /dev/null; git cannot read a controlling terminal

	var outBuf, errBuf capBuffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	res.stdout = outBuf.String()
	res.stderr = errBuf.String()

	switch {
	case ctx != nil && ctx.Err() == context.DeadlineExceeded:
		res.runErr = fmt.Errorf("timed out after %s", timeout)
		res.exitCode = -1
	case err != nil:
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.exitCode = exitErr.ExitCode()
		} else {
			res.runErr = err
			res.exitCode = -1
		}
	}
	return res
}

// gitEnv returns a hardened environment for git subprocesses. It disables
// interactive credential prompts (which would otherwise hang a bulk run or
// invite accidental credential entry), avoids taking the index lock for
// read-only operations, and puts SSH into batch mode unless the user has
// their own GIT_SSH_COMMAND.
func gitEnv() []string {
	env := append(os.Environ(),
		"GIT_TERMINAL_PROMPT=0",
		"GIT_OPTIONAL_LOCKS=0",
	)
	if _, ok := os.LookupEnv("GIT_SSH_COMMAND"); !ok {
		env = append(env, "GIT_SSH_COMMAND=ssh -oBatchMode=yes")
	}
	return env
}

// displayName renders a repo path relative to cwd when possible, falling
// back to the basename. This keeps the overview readable.
func displayName(cwd, repo string) string {
	if cwd != "" {
		if rel, err := filepath.Rel(cwd, repo); err == nil && rel != "." && rel != "" {
			return rel
		}
	}
	return filepath.Base(repo)
}
