# Example: `gh` skills shim

A working example of the embedded-skills design (`../../docs/`), targeting
GitHub's `gh` CLI. Build one Go binary, install it under the name `gh` in a
directory at the front of your `PATH`, and your agent gains:

```console
$ gh skills                  # catalog of gh workflow skills
$ gh skills pr-flow          # one full workflow, verbatim SKILL.md
$ gh skills --json           # machine-readable catalog for harnesses
$ gh pr status               # ← everything else: the real gh, untouched
```

Nothing about this is gh-specific except the corpus. The binary is a
PATH-shadow shim (docs/06): it answers `skills` (and the optional leading
`--skill`) itself from skills **embedded in the binary at build time**, and
execs the real `gh` for every other invocation. No gh extensions, no plugin
dispatch, no cooperation from gh.

## What's inside

```text
examples/gh/
├── main.go  catalog.go  frontmatter.go  proxy.go  commands.go
│                            the shim (standard library only, no dependencies)
├── exec_unix.go / exec_other.go
│                            pass-through: syscall.Exec on Unix
├── cmd/genmanifest/         build-time tool: regenerates manifest.json
├── manifest.json            generated index of the corpus (do not edit)
├── skills/                  the corpus — this is the part worth copying
│   ├── issue-flow/SKILL.md      beginner
│   ├── pr-flow/SKILL.md         intermediate (+ references/body-template.md)
│   ├── review-flow/SKILL.md     intermediate
│   └── pr-merge/SKILL.md        advanced
├── Makefile                 build / install / test / check
└── main_test.go
```

Skill format and authoring rules: `../../docs/02-skill-format.md` and
`../../docs/08-skill-authoring.md`. The four bundled skills form one
chain (`issue-flow → pr-flow → review-flow → pr-merge`) and were written
against gh 2.103 (manifest: `>=2.40`).

## Install

Requires Go (to build) and a real `gh` already installed and authenticated.

```console
$ cd examples/gh
$ make install
```

That compiles `dist/gh` and copies it to
`~/.local/share/skills-shims/bin/gh` (override with `SHIM_DIR=...`).
It does **not** edit your shell profile. Add the directory to the front
of `PATH` yourself, in whatever environment your agent runs in:

```console
$ export PATH="$HOME/.local/share/skills-shims/bin:$PATH"
$ which gh        # must print the shim path, not your real gh
$ gh skills       # catalog from the shim
$ gh --version    # still the real gh, proxied
```

Then tell your agent the corpus exists — paste into the project's
`AGENTS.md` / `CLAUDE.md` (see `../../docs/examples/agents-md-snippet.md`).
Agents are not trained to run `gh skills` spontaneously; that one-paste
pointer is the discovery bootstrap (docs/07).

Uninstall: `make uninstall`, or just remove the directory from `PATH`.
The real `gh` was never modified.

## How it works

- Invoked as `gh` with first arg `skills` → answered locally from the
  embedded corpus. Read-only and offline, except `skills export`.
- Leading `--skill <name>` (optional convenience) → delimited skill
  blocks printed, then the rest of the command runs.
- Anything else → `syscall.Exec` of the real binary, so argv, env, stdio,
  signals, and exit codes are literally the real gh's.
- Real-binary resolution walks `PATH` skipping the shim itself
  (`os.SameFile` on resolved paths). Override when needed:
  `SKILLS_REAL_GH=/usr/bin/gh gh ...` (or `SKILLS_REAL_BIN=...`).
- Stacked wrappers are handled: if your `gh` is already wrapped (say, an
  identity/token script), the two shims would otherwise exec each other
  forever, since each resolves "the real gh" by scanning PATH and finds
  the other first. The shim threads a visited set of resolved paths
  through `EMBEDDED_SKILLS_SHIM_VISITED`; resolution skips anything
  already handed off to, so the chain terminates at the true binary —
  with the intermediary wrapper's environment (e.g. its token) intact.
- On the skills path only, the shim probes the real binary's version
  (`--json` reports `detected_version` and per-skill `version_ok`).
  The proxy path never probes.
- Invoked under any other name (e.g. `./dist/gh` from a path, `go run .`),
  it offers `doctor` (diagnostics) and passes `skills ...` through for
  development.

Try the machine paths your harness would use:

```console
$ gh skills --json
$ gh skills pr-flow --json
$ gh skills search merge
$ gh skills export pr-flow --to ./.claude/skills/pr-flow
$ gh skills check          # validates the corpus against your real gh
```

## Add or edit a skill

1. Create `skills/<name>/SKILL.md` (copy an existing skill as the template;
   rules in docs/08).
2. `make manifest` — regenerates the index + checksums.
3. `make test && make check` — unit tests plus corpus validation against
   your installed gh (command paths are resolved against the real binary).
4. `make install` — skills are embedded at build time, so a content change
   is a rebuild. That is deliberate: the guidance can never drift from the
   artifact you installed.

## Example-grade limitations (vs. the full design)

- The frontmatter parser is a restricted subset (scalars, flow/block
  lists), stdlib-only by choice. `skills check` rejects anything outside it.
- `check` validates command *paths* against the real binary, not every
  flag in every body (the design's per-target help adapters, docs/06).
- `--skill` is supported only as a leading run of flags, delimited-block
  form only (no JSONL). The canonical path — harness fetches
  `gh skills <name> --json`, runs the command separately — is unaffected.
- No usage tracking, `next`, or `feedback`: designed, deferred (docs/04).
- Windows: pass-through uses `os/exec` with exit-code propagation instead
  of `syscall.Exec`; treat as a fallback, Unix is the tested path.
