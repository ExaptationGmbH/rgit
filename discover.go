package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// discover walks root recursively and returns the paths of every git
// repository found, sorted for stable output.
//
// A directory is a git repo if it contains a ".git" entry (a subdirectory
// for normal repos, or a file for worktrees/submodules). Once a repo is
// found we do NOT descend into it: nested content (submodules, vendored
// checkouts, the .git dir itself) is treated as part of that repo rather
// than as separate repos. This keeps the overview about *your* top-level
// projects and avoids walking huge object stores.
//
// maxDepth limits how deep to descend; -1 means unlimited. Depth 0 is root
// itself.
func discover(root string, maxDepth int) ([]string, error) {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}

	var repos []string
	walkFn := func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// Unreadable directory: skip it rather than aborting the scan.
			if d != nil && d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if !d.IsDir() {
			return nil
		}

		if maxDepth >= 0 && depthOf(absRoot, path) > maxDepth {
			return fs.SkipDir
		}

		base := d.Name()
		// Skip common heavy directories that never contain repos we care about.
		if path != absRoot && (base == "node_modules" || base == "vendor") {
			return fs.SkipDir
		}

		if isGitRepo(path) {
			repos = append(repos, path)
			return fs.SkipDir // don't descend into a repo
		}
		return nil
	}

	if err := filepath.WalkDir(absRoot, walkFn); err != nil {
		return nil, err
	}

	sort.Strings(repos)
	return repos, nil
}

// isGitRepo reports whether dir contains a ".git" entry.
func isGitRepo(dir string) bool {
	_, err := os.Stat(filepath.Join(dir, ".git"))
	return err == nil
}

// depthOf returns how many path segments deep path is relative to root.
func depthOf(root, path string) int {
	rel, err := filepath.Rel(root, path)
	if err != nil || rel == "." {
		return 0
	}
	return strings.Count(rel, string(filepath.Separator)) + 1
}
