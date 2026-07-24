# Security policy

## Reporting a vulnerability

Please report security issues privately to
**rgit.subsiding036@passmail.net**. Do not open a public issue for
undisclosed vulnerabilities. We aim to acknowledge reports within a few days.

## Threat model

`rgit` is a thin, 1:1 wrapper around `git`. It discovers git repositories
recursively beneath the current directory and runs the git command you give it
in each one. Its trust boundary is therefore the same as running git yourself —
**with one important amplification**: rgit runs git in *every* repository it
finds, including ones you may not have deliberately targeted.

### Running git in a repository can execute code from that repository

This is a property of git, not a bug in rgit, but rgit makes it easier to
trigger unintentionally. A repository's `.git/config` (and related files) can
point git at programs that run as part of ordinary commands — for example
`core.fsmonitor`, `core.pager`, `core.sshCommand`, hooks, or aliases. If you
run `rgit` in a directory tree that contains an **untrusted** repository (say,
an extracted archive or a freshly cloned third-party repo), running git in it
can execute attacker-controlled code.

**Guidance:** only run `rgit` over trees of repositories you trust, the same
way you would only `cd` into and run git in a repo you trust. git's
`safe.directory` protection still applies (git refuses to operate on repos
owned by another user).

### Hardening rgit already applies

- **No interactive prompts / no hangs.** git subprocesses run with
  `GIT_TERMINAL_PROMPT=0`, stdin connected to `/dev/null`, and SSH in batch
  mode (`ssh -oBatchMode=yes`, unless you set your own `GIT_SSH_COMMAND`). A
  repo that needs credentials fails fast instead of hanging the whole run or
  prompting you to type a password into a bulk operation.
- **Terminal-escape sanitization.** Repository names and the summarised git
  output shown in the condensed views are stripped of ANSI escape sequences
  and control characters, so a maliciously named repo can't inject terminal
  escapes (cursor moves, title changes) into rgit's output. The opt-in
  `--full` view shows verbatim git output and is intentionally not sanitized.
- **Read-only friendliness.** `GIT_OPTIONAL_LOCKS=0` avoids taking the index
  lock for read-only operations like `status`.
- **Resource bounds.** Concurrency is capped (max 128 parallel git
  processes), per-stream output is capped at 1 MiB, and `--timeout` can bound
  each git invocation.

### Out of scope

- Vulnerabilities in `git` itself (report those upstream).
- Arbitrary code execution that results from you choosing to run git in a repo
  you control or trust.
