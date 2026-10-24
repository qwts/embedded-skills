# Target: `agent-bot` skills shim

A PATH-shadow shim named `agent-bot` that serves an embedded skill corpus for
the `agent-bot` CLI (the qwts agent identity runtime) and proxies everything
else to the real `agent-bot`. Design: `../../docs/06-shim-pattern.md`,
surface: `../../docs/01-cli-surface.md`.

The shim logic lives in the shared [`../../shim/`](../../shim) package; this
target is a thin wrapper that embeds the corpus and calls `shim.Run`.

## What's inside

```text
targets/agent-bot/
├── main.go              thin wrapper: embeds corpus, calls shim.Run
├── main_test.go         corpus validity + checksum tests
├── go.mod               module with replace directive to ../../shim
├── Makefile             build / install / test / check / manifest
├── manifest.json        generated index of the corpus (do not edit)
├── skills/
│   ├── verified-publish/SKILL.md   advanced, destructive
│   └── join-and-bind/SKILL.md      intermediate, mutating
└── README.md            this file
```

Two skills, chosen by failure frequency from agent session mining:

- **`verified-publish`** — the `agent-bot signed-commit` flow: identity
  preconditions, dry-run, publish, recovery, verifying GitHub shows the
  signed bot commit.
- **`join-and-bind`** — `agent-bot join`, `setup-worktree`, MCP bind,
  `doctor`, and the failure modes that fail closed (missing App, missing
  binding, wrong checkout).

Both skills reconcile with the existing `agent-bot skill` surface: they
present the same workflows as standalone, discoverable procedures without
duplicating the bundled `--for` references.

## Prerequisites

- Go 1.23+ (to build)
- `agent-bot` v0.12.0 installed and on `PATH` (tested range: `>=0.12.0 <0.13.0`)

## Build

```console
$ cd targets/agent-bot
$ make build
```

Compiles `dist/agent-bot` — a binary named `agent-bot`. That name is the
mechanism.

## Install

```console
$ make install
```

Copies `dist/agent-bot` to `~/.local/share/skills-shims/bin/agent-bot`
(override with `SHIM_DIR=...`). It does **not** edit your shell profile.
Add the directory to the front of `PATH` yourself:

```console
$ export PATH="$HOME/.local/share/skills-shims/bin:$PATH"
$ which agent-bot        # must print the shim path, not your real agent-bot
$ agent-bot skills       # catalog from the shim
$ agent-bot --version    # still the real agent-bot, proxied
```

Then tell your agent the corpus exists — paste into the project's
`AGENTS.md` / `CLAUDE.md` (see `../../docs/examples/agents-md-snippet.md`).

Uninstall: `make uninstall`, or just remove the directory from `PATH`.
The real `agent-bot` was never modified.

## Verify

```console
$ agent-bot skills                  # catalog: join-and-bind, verified-publish
$ agent-bot skills --json           # machine-readable, with detected_version
$ agent-bot skills verified-publish # one full workflow, verbatim
$ agent-bot skills search publish   # search by keyword
$ agent-bot skills check            # validate corpus against real agent-bot
$ agent-bot --version              # proxy: real agent-bot, untouched
$ agent-bot doctor                 # proxy: real agent-bot doctor
```

## How it works

- Invoked as `agent-bot` with first arg `skills` → answered locally from the
  embedded corpus. Read-only and offline, except `skills export`.
- Leading `--skill <name>` (optional convenience) → delimited skill
  blocks printed, then the rest of the command runs.
- Anything else → `syscall.Exec` of the real binary, so argv, env, stdio,
  signals, and exit codes are literally the real agent-bot's.
- Real-binary resolution walks `PATH` skipping the shim itself
  (`os.SameFile` on resolved paths). Override when needed:
  `SKILLS_REAL_AGENT_BOT=/usr/local/bin/agent-bot agent-bot ...`
  (or `SKILLS_REAL_BIN=...`).
- Stacked wrappers are handled via the visited-set mechanism
  (`EMBEDDED_SKILLS_SHIM_VISITED`).
- On the skills path only, the shim probes the real binary's version
  (`--json` reports `detected_version` and per-skill `version_ok`).
  The proxy path never probes.
- Invoked under any other name (e.g. `./dist/agent-bot` from a path,
  `go run .`), it offers `doctor` (diagnostics) and passes `skills ...`
  through for development.

## Collision policy

`agent-bot` already has a `skill` subcommand (singular). The shim intercepts
`skills` (plural), which does not collide today. The shim does not probe the
real binary before serving — it intercepts a leading `skills` unconditionally.
If `agent-bot` ever grows a `skills` subcommand, the shim will shadow it;
the design's collision policy (real behavior wins, via a `<real> skills --help`
probe) is documented in `../../docs/06-shim-pattern.md` but not yet implemented
in this shim.

## Add or edit a skill

1. Create `skills/<name>/SKILL.md` (copy an existing skill as the template;
   rules in `../../docs/08-skill-authoring.md`).
2. `make manifest` — regenerates the index + checksums.
3. `make test && make check` — unit tests plus corpus validation against
   your installed agent-bot (command paths are resolved against the real
   binary).
4. `make install` — skills are embedded at build time, so a content change
   is a rebuild.

## Known limitations

- The frontmatter parser is a restricted subset (scalars, flow/block
  lists), stdlib-only by choice. `skills check` rejects anything outside it.
- `check` validates command *paths* against the real binary, not every
  flag in every body (the design's per-target help adapters, docs/06).
- `--skill` is supported only as a leading run of flags, delimited-block
  form only (no JSONL). The canonical path — harness fetches
  `agent-bot skills <name> --json`, runs the command separately — is
  unaffected.
- No usage tracking, `next`, or `feedback`: designed, deferred (docs/04).
- Windows: pass-through uses `os/exec` with exit-code propagation instead
  of `syscall.Exec`; treat as a fallback, Unix is the tested path.
