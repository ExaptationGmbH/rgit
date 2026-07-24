# Multi-Repo Git Commands

Commands that make sense to run across **many repos at once** — e.g. a parent
folder holding several independent repositories:

```
projects/
├── repo-a/   (git repo)
├── repo-b/   (git repo)
├── repo-c/   (git repo)
└── repo-d/   (git repo)
```

The goal of every command below is **getting an overview** or **doing routine
maintenance** across all repos in one shot, instead of `cd`-ing into each one.

---

## The mental model

A command belongs on this list if running it in 4 repos and reading 4 outputs
together tells you something useful *at a glance*. Two families qualify:

1. **Read-only overview** — "what's the state of everything?" (status, branch, ahead/behind, stashes, recent commits)
2. **Bulk sync / maintenance** — "keep everything current / tidy" (fetch, pull, prune, gc)

Anything that needs interactive input, targets one specific repo, or rewrites
history is a *bad* fit for fan-out and is listed at the bottom.

---

## 1. Overview commands (read-only, safe)

These never change anything — run them freely across every repo.

| Command | What it tells you across all repos |
| --- | --- |
| `git status -sb` | Short status + branch + ahead/behind in a few lines each. The single best overview command. |
| `git status` | Full status. Verbose but complete. |
| `git branch --show-current` | Which branch each repo is currently on. |
| `git rev-parse --abbrev-ref HEAD` | Same as above, scriptable (works on older git too). |
| `git log --oneline -5` | Last 5 commits per repo — spot which repos moved recently. |
| `git log --oneline -1` | Just the tip commit of each repo. |
| `git stash list` | Find forgotten stashes hiding in any repo. |
| `git remote -v` | Confirm every repo points at the remote you expect. |
| `git diff --stat` | Summary of uncommitted (unstaged) changes per repo. |
| `git diff --cached --stat` | Summary of staged-but-uncommitted changes. |
| `git tag --sort=-creatordate \| head` | Newest tags per repo (release overview). |
| `git describe --tags --always` | Human-readable version marker per repo. |
| `git shortlog -sn --all` | Contributor counts — quick "who touched this repo". |
| `git for-each-ref --format='%(refname:short) %(upstream:track)' refs/heads` | Ahead/behind for **every** local branch, not just current. |

### The "am I ahead / behind the remote?" check

Two steps, because ahead/behind is only accurate *after* fetching:

```bash
git fetch --quiet          # update remote-tracking refs (see §2)
git status -sb             # now the ↑/↓ counts are trustworthy
```

---

## 2. Sync & update commands (network / mutating, mostly safe)

| Command | Effect across all repos |
| --- | --- |
| `git fetch --all --prune` | Download new commits + delete refs for branches deleted on the remote. **Does not touch your working tree** — safest way to refresh. |
| `git fetch --quiet` | Quiet fetch, ideal before a bulk `status -sb`. |
| `git pull --ff-only` | Update current branch, but **only** if it can fast-forward. Refuses to create merge commits — safe for fan-out. |
| `git remote update` | Refresh all remotes' tracking refs. |
| `git submodule update --init --recursive` | Sync submodules across repos that have them. |

> ⚠️ Plain `git pull` (without `--ff-only`) can trigger merges/conflicts and
> should **not** be blasted across many repos unattended. Prefer `--ff-only`.

---

## 3. Maintenance commands (housekeeping)

| Command | Effect across all repos |
| --- | --- |
| `git gc` | Compact/optimize the repo. `git gc --auto` only runs when needed. |
| `git prune` | Drop unreachable objects (usually done by `gc`). |
| `git remote prune origin` | Remove stale remote-tracking branches. |
| `git maintenance run` | Modern umbrella for gc/prune/commit-graph (git ≥ 2.30). |
| `git count-objects -vH` | Report repo size — find the bloated one. |
| `git fsck` | Integrity check across all repos. |

---

## 4. Running a command across every subfolder

### One-liner (bash/zsh)

```bash
for d in */; do
  [ -d "$d/.git" ] || continue
  echo "=== $d ==="
  git -C "$d" status -sb
done
```

`git -C <path>` runs a command as if from inside that repo — no `cd` needed.

### Skip repos that aren't git repos safely

```bash
for d in */; do
  git -C "$d" rev-parse --is-inside-work-tree >/dev/null 2>&1 || continue
  printf '\n=== %s ===\n' "$d"
  git -C "$d" status -sb
done
```

### Reusable shell function

```bash
# in ~/.bashrc / ~/.zshrc — run any git command in every subrepo
gitall() {
  for d in */; do
    git -C "$d" rev-parse --git-dir >/dev/null 2>&1 || continue
    printf '\n=== %s ===\n' "${d%/}"
    git -C "$d" "$@"
  done
}
# usage:  gitall status -sb   |   gitall fetch --all --prune   |   gitall log --oneline -3
```

### Purpose-built tools

If you do this a lot, dedicated tools are nicer than shell loops:

- **[`mr`](https://myrepos.branchable.com/)** (myrepos) — declare repos in `.mrconfig`, run `mr status`, `mr update`, etc.
- **[`gita`](https://github.com/nosarthur/gita)** — side-by-side status of many repos in one colored table.
- **[`ghorg`](https://github.com/gabrie30/ghorg)** — clone/pull a whole GitHub/GitLab org at once.
- **`git submodule` / `git subtree`** — if the repos are genuinely a single project.
- **`vcstool` / `wstool`** — used in the ROS world to manage many repos from a manifest.

---

## 5. What does NOT belong on a fan-out list

Avoid blasting these across many repos — they're per-repo, interactive, or destructive:

| Command | Why it's a bad fit |
| --- | --- |
| `git commit` | Needs a message + intentional, repo-specific staging. |
| `git push` | Publishes — do it deliberately per repo, not in bulk. |
| `git merge` / plain `git pull` | Can create conflicts that need hands-on resolution. |
| `git rebase` | History rewrite; conflict-prone. |
| `git checkout <branch>` / `git switch` | The right branch differs per repo. |
| `git reset --hard` / `git clean -fd` | **Destructive** — mass-running risks data loss. |
| `git filter-branch` / `git filter-repo` | Rewrites history irreversibly. |
| `git rm` / `git mv` | Path-specific to one repo. |

> Rule of thumb: **read-only overview and idempotent maintenance = safe to fan out.
> Anything that writes commits, publishes, or rewrites history = do per repo.**

---

## Quick reference — the daily overview

```bash
gitall fetch --quiet         # refresh remote-tracking refs everywhere
gitall status -sb            # one clean overview of all repos
gitall stash list            # anything left uncommitted-and-stashed?
```
