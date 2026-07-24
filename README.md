# rgit — recursive git

> A friend of mine showed me a screenshot where they could do a `git status`
> for multiple repos. So I took the idea and packaged it up into this.
> Credits to [nanocat.net](https://nanocat.net/).

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
```

This uses the custom tap at `github.com/exaptation/homebrew-tap` (see
[Releasing](#releasing) for how the tap is maintained). To try a formula from a
local checkout without a tap:

```sh
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
make install                       # installs to /usr/local/bin
```

`/usr/local/bin` needs root on many systems (e.g. Apple Silicon macOS). If you
hit a permission error, either:

```sh
sudo make install                  # into /usr/local/bin
make install PREFIX=$HOME/.local   # no sudo; add ~/.local/bin to your PATH
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

| Flag               | Meaning                                             |
| ------------------ | --------------------------------------------------- |
| `-C, --dir <path>` | Root directory to scan (default: current directory) |
| `--depth <n>`      | Max directory depth to descend (default: unlimited) |
| `-j, --jobs <n>`   | Concurrent git invocations (default: auto, max 128) |
| `--timeout <dur>`  | Per-repo git timeout, e.g. `30s`, `2m` (default: none) |
| `-f, --full`       | Show full git output instead of the condensed view  |
| `--no-color`       | Disable ANSI colours (also honours `NO_COLOR`)      |
| `-h, --help`       | Help                                                |
| `-v, --version`    | Version                                             |

## Reading the status table

```
repo-name            branch     ↑ahead ↓behind   state
```

| Symbol      | Meaning                               |
| ----------- | ------------------------------------- |
| `↑N` / `↓N` | commits ahead / behind the upstream   |
| `✔ clean`   | nothing to commit, working tree clean |
| `+N`        | staged changes                        |
| `~N`        | modified (unstaged)                   |
| `-N`        | deleted                               |
| `?N`        | untracked files                       |
| `✖N`        | merge conflicts                       |

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

## Security

rgit runs git in **every** repo it finds, so only point it at trees you trust:
running git in an untrusted repo can execute code from that repo's config
(fsmonitor, hooks, aliases). rgit hardens the subprocesses (no interactive
credential prompts, SSH batch mode, stdin from `/dev/null`, output sanitized of
terminal escapes, concurrency/output/`--timeout` bounds). See
[SECURITY.md](SECURITY.md) for the full threat model and how to report issues.

## Development

```sh
make test       # go test ./...
make vet
make build      # -> bin/rgit
```

## Releasing

Releases are automated with
[release-please](https://github.com/googleapis/release-please). Commit using
[Conventional Commits](https://www.conventionalcommits.org/) and the rest
happens on merge to `main`:

- `feat: …` → minor bump, `fix: …` → patch bump, `feat!:` / `BREAKING CHANGE`
  → major bump. `chore:`/`docs:`/`refactor:` don't trigger a release.
- release-please opens and maintains a **release PR** that updates
  `CHANGELOG.md` and the version. Merging it tags `vX.Y.Z` and publishes a
  GitHub Release.
- A follow-up CI job then runs `scripts/update-formula.sh` to pin
  `Formula/rgit.rb` to the new release tarball's `sha256` and commits it back.

Version numbers are injected into the binary at build time via
`-ldflags -X main.version=…`, so `rgit --version` reports the release tag.

### Setting up the Homebrew tap (one-time)

`brew install exaptation/tap/rgit` needs a tap repo named
`github.com/exaptation/homebrew-tap` containing the formula:

```sh
# in the tap repo
mkdir -p Formula
cp /path/to/rgit/Formula/rgit.rb Formula/rgit.rb
git add Formula/rgit.rb && git commit -m "rgit 0.1.0" && git push
```

After each release, copy the freshly-pinned `Formula/rgit.rb` from this repo
into the tap (or automate it with a CI step that pushes to the tap using a PAT).
To pin the formula manually for a tag:

```sh
scripts/update-formula.sh v0.1.0
```
