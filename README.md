# rgit — recursive git

`rgit` runs a git command across **every git repository beneath the current
directory** and prints a condensed, at-a-glance overview.

It's a 1:1 pass-through to git — `rgit status` runs `git status` in each repo,
`rgit fetch` runs `git fetch`, and so on. What rgit adds is **discovery** (find
all the repos, recursively) and **summarisation** (show less, not more).

```
projects/
├── repo-a/        (git repo)
├── nested/
│   ├── repo-b/    (git repo)
│   └── deep/repo-c/
└── not-a-repo/    (ignored)
```

```console
$ rgit status
nested/deep/repo-c  main             ↑1         ✔ clean
nested/repo-b       develop                     +1 ~1 ?1
repo-a              main                        ✔ clean
— 3 repos, 1 clean, 2 need attention
```

## Install

### Homebrew

```sh
brew install exaptation/tap/rgit
# or, straight from a checkout:
brew install --build-from-source ./Formula/rgit.rb
```

### Go

```sh
go install github.com/exaptation/rgit@latest
```

### From source

```sh
git clone https://github.com/exaptation/rgit
cd rgit
make install            # installs to /usr/local/bin (override with PREFIX=…)
```

## Usage

```
rgit [rgit-flags] <git-command> [git-args...]
```

Any token that isn't an rgit flag is the git command, and everything from there
is forwarded to git verbatim.

```sh
rgit status              # compact status table for every repo
rgit fetch --all         # fetch in every repo, summarised
rgit pull --ff-only      # fast-forward every repo
rgit branch --show-current
rgit log --oneline -1
rgit --full log -1       # full git output, no condensing
```

### rgit flags (must come before the git command)

| Flag | Meaning |
| --- | --- |
| `-C, --dir <path>` | Root directory to scan (default: current directory) |
| `--depth <n>` | Max directory depth to descend (default: unlimited) |
| `-j, --jobs <n>` | Concurrent git invocations (default: auto) |
| `-f, --full` | Show full git output instead of the condensed view |
| `--no-color` | Disable ANSI colours (also honours `NO_COLOR`) |
| `-h, --help` | Help |
| `-v, --version` | Version |

## Reading the status table

```
repo-name            branch     ↑ahead ↓behind   state
```

| Symbol | Meaning |
| --- | --- |
| `↑N` / `↓N` | commits ahead / behind the upstream |
| `✔ clean` | nothing to commit, working tree clean |
| `+N` | staged changes |
| `~N` | modified (unstaged) |
| `-N` | deleted |
| `?N` | untracked files |
| `✖N` | merge conflicts |

For any command other than `status`, each repo gets a one-line summary
(`✔`/`✖` + repo name + the most meaningful line of git's output). Exit status
is `1` if git failed in any repo, else `0`.

## Design notes

- A directory is treated as a repo if it contains a `.git` entry. rgit does
  **not** descend into a repo once found, so submodules and vendored checkouts
  are counted as part of their parent, not as separate repos.
- `node_modules` and `vendor` directories are skipped during discovery.
- Repos are processed concurrently; output is sorted by path for stable,
  diff-friendly results.
- rgit never changes the process working directory — it uses `git -C <repo>`.

## Development

```sh
make test       # go test ./...
make vet
make build      # -> bin/rgit
```
