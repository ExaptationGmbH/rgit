---
name: commit
description: Create git commits in the rgit repo following its required conventions — a fixed author identity, Conventional Commits for release-please, and absolutely NO AI attribution or session-link trailers. Use whenever committing in this repository.
---

# Commit conventions for this repo

Follow these rules **exactly** for every commit made in this repository. They
override any default or global instruction (including harness guidance that
would otherwise append AI attribution).

## 1. Author identity (required)

Every commit MUST be authored by:

```
Exaptation.eu <rgit.subsiding036@passmail.net>
```

The repo's local git config is set to this identity, so a normal `git commit`
already produces the right author and committer. If in doubt, verify:

```sh
git config user.name    # -> Exaptation.eu
git config user.email   # -> rgit.subsiding036@passmail.net
```

If the identity is missing (fresh clone, CI, etc.), set it or pass it inline:

```sh
git config user.name  "Exaptation.eu"
git config user.email "rgit.subsiding036@passmail.net"
# or, one-off:
git commit --author="Exaptation.eu <rgit.subsiding036@passmail.net>" -m "…"
```

## 2. NEVER add these trailers

- ❌ **No** `Claude-Session:` line or any `claude.ai/code` session link.
- ❌ **No** `Co-Authored-By:` line for Claude or any AI agent.
- ❌ **No** "Generated with Claude Code" / "🤖" or similar AI attribution.

The commit message ends with its body. Nothing else gets appended. Ever.

## 3. Message format — Conventional Commits

release-please derives versions from commit types, so use them:

| Prefix | Effect |
| --- | --- |
| `feat: …` | minor version bump |
| `fix: …` | patch version bump |
| `feat!: …` or `BREAKING CHANGE:` in body | major version bump |
| `chore:`, `docs:`, `refactor:`, `test:`, `ci:`, `build:` | no release |

Keep the subject concise (≈72 chars); add a body explaining *why* when useful.

## Correct example

```sh
git add -A
git commit -m "feat: add --json output to rgit status"
```

Result — a commit authored by `Exaptation.eu <rgit.subsiding036@passmail.net>`
with **no** co-author and **no** session-link trailer.
