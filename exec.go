package main

import (
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
)

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
// auto (a sensible multiple of the CPU count, capped).
func execAll(repos []string, gitArgs []string, jobs int) []result {
	if jobs <= 0 {
		jobs = runtime.NumCPU() * 2
		if jobs > 16 {
			jobs = 16
		}
		if jobs < 1 {
			jobs = 1
		}
	}
	if jobs > len(repos) {
		jobs = len(repos)
	}

	cwd := currentDir()

	results := make([]result, len(repos))
	sem := make(chan struct{}, jobs)
	var wg sync.WaitGroup

	for idx, repo := range repos {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, repo string) {
			defer wg.Done()
			defer func() { <-sem }()
			results[idx] = runGit(repo, cwd, gitArgs)
		}(idx, repo)
	}
	wg.Wait()

	sort.Slice(results, func(i, j int) bool {
		return results[i].name < results[j].name
	})
	return results
}

// runGit executes git in a single repo. It uses `git -C <repo>` so we never
// have to change the process working directory.
func runGit(repo, cwd string, gitArgs []string) result {
	res := result{repo: repo, name: displayName(cwd, repo)}

	args := append([]string{"-C", repo}, gitArgs...)
	cmd := exec.Command("git", args...)

	var outBuf, errBuf capBuffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &errBuf

	err := cmd.Run()
	res.stdout = outBuf.String()
	res.stderr = errBuf.String()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.exitCode = exitErr.ExitCode()
		} else {
			res.runErr = err
			res.exitCode = -1
		}
	}
	return res
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
