# Target: `agent-bot` skills shim

A PATH-shadow shim for the `agent-bot` CLI (standalone agent identity
runtime, tested against 0.12.0). Build one Go binary, install it under the
name `agent-bot` in a directory at the front of your `PATH`, and your agent
gains:

```console
$ agent-bot skills                  # catalog of agent-bot workflow skills
$ agent-bot skills verified-publish # one full workflow, verbatim SKILL.md
$ agent-bot skills --json           # machine-readable catalog for harnesses
$ agent-bot doctor                 # ← everything else: the real agent-bot, untouched
```

The binary is a PATH-shadow shim (docs/06): it answers `skills` (and the
optional leading `--skill`) itself from skills **embedded in the binary at
build time**, and execs the real `agent-bot` for every other invocation.
No plugin or extension integration — `agent-bot` has no plugin dispatch,
and the design forbids it anyway.

This is the first target built on the shared shim module in
[`../../shim/`](../../shim/) (the generic code extracted from the
[`examples/gh/`](../../examples/gh/) reference implementation, per
docs/06 "one shim codebase, many targets"). This directory contributes
only the target name and the corpus.

## Two discovery surfaces, one story

`agent-bot` ships its own bundled skill surface, and this shim adds the
workflow layer on top. They cross-link; neither duplicates the other:

| Surface | Command | What it serves |
|---|---|---|
| Bundled (ships with the CLI) | `agent-bot skill <name>` | Router skill (`agent-bot`, `agent-space`, `thread-orders`) read offline from the release; other names resolve through the SOP catalog at a pinned commit |
| Bundled | `agent-bot skill agent-bot --for <subcommand>` | One reference file per subcommand (static table) |
| Bundled | `agent-bot skill path` | The bundle directory and its source commit |
| This shim | `agent-bot skills` | Workflow skills: `join-and-bind`, `verified-publish` |

The bundled router skill routes to deep references (operations, verified
publishing, execution identities); the shim's corpus carries the executable
procedures agents get wrong under trial and error — the join/bind/setup
sequence and the signed-commit publish flow. Each shim skill's "See also"
points back at the bundled surface, so an agent that lands on either entry
point finds the other. `agent-bot` 0.12.0 has no native `skills` subcommand
(verified: `agent-bot skills` exits 1 with `unknown command: skills`), so
the shim shadows nothing the CLI already answers.

## What's inside

```text
targets/agent-bot/
├── main.go                  thin target: embeds the corpus, calls shim.Run
├── main_test.go             corpus validity + manifest checksum gate
├── go.mod                   requires embedded-skills/shim (local replace)
├── skills/                  the corpus — this is the part worth copying
│   ├── join-and-bind/SKILL.md      intermediate: join → MCP bind → setup-worktree → doctor
│   └── verified-publish/SKILL.md   advanced: signed-commit preconditions → dry-run → publish → verify
├── manifest.json            generated index of the corpus (do not edit)
├── Makefile                 build / install / test / check / manifest
└── dist/agent-bot           build artifact (gitignored)
```

Skill format and authoring rules: `../../docs/02-skill-format.md` and
`../../docs/08-skill-authoring.md`. The two skills form one chain
(`join-and-bind` → `verified-publish`) and were written against agent-bot
0.12.0 (manifest: `>=0.12.0 <0.13.0`, matching the CLI's own declared
version contract for this release line).

## Install

Requires Go (to build) and a real `agent-bot` already installed.

```console
$ cd targets/agent-bot
$ make install
```

That compiles `dist/agent-bot` and copies it to
`~/.local/share/skills-shims/bin/agent-bot` (override with `SHIM_DIR=...`).
It does **not** edit your shell profile. Add the directory to the front
of `PATH` yourself, in whatever environment your agent runs in:

```console
$ export PATH="$HOME/.local/share/skills-shims/bin:$PATH"
$ which agent-bot        # must print the shim path, not your real agent-bot
$ agent-bot skills       # catalog from the shim
$ agent-bot --version    # still the real agent-bot, proxied
```

Then tell your agent the corpus exists — paste into the project's
`AGENTS.md` / `CLAUDE.md` (see `../../docs/examples/agents-md-snippet.md`;
mention both `agent-bot skills` and the bundled `agent-bot skill`).
Agents are not trained to run `agent-bot skills` spontaneously; that
one-paste pointer is the discovery bootstrap (docs/07).

Uninstall: `make uninstall`, or just remove the directory from `PATH`.
The real `agent-bot` was never modified.

## How it works

- Invoked as `agent-bot` with first arg `skills` → answered locally from
  the embedded corpus. Read-only and offline, except `skills export`.
- Leading `--skill <name>` (optional convenience) → delimited skill
  blocks printed, then the rest of the command runs.
- Anything else → `syscall.Exec` of the real binary, so argv, env, stdio,
  signals, and exit codes are literally the real agent-bot's.
- Real-binary resolution walks `PATH` skipping the shim itself
  (`os.SameFile` on resolved paths). Override when needed:
  `SKILLS_REAL_AGENT_BOT=/path/to/agent-bot agent-bot ...` (or
  `SKILLS_REAL_BIN=...`).
- Stacked wrappers are handled: a visited set of resolved paths is
  threaded through `EMBEDDED_SKILLS_SHIM_VISITED`, so a chain of wrappers
  terminates at the true binary with the intermediary wrapper's
  environment intact.
- On the skills path only, the shim probes the real binary's version
  (`--json` reports `detected_version` and per-skill `version_ok`).
  The proxy path never probes.
- Invoked under any other name (e.g. `./dist/agent-bot` from a path,
  `go run .`), it offers `doctor` (diagnostics) and passes `skills ...`
  through for development.

Try the machine paths your harness would use:

```console
$ agent-bot skills --json
$ agent-bot skills verified-publish --json
$ agent-bot skills search publish
$ agent-bot skills export verified-publish --to ./.claude/skills/verified-publish
$ agent-bot skills check          # validates the corpus against your real agent-bot
```

## Add or edit a skill

1. Create `skills/<name>/SKILL.md` (copy an existing skill as the template;
   rules in docs/08).
2. `make manifest` — regenerates the index + checksums (the generator is
   the shared `embedded-skills/shim/cmd/genmanifest`).
3. `make test && make check` — unit tests plus corpus validation against
   your installed agent-bot (command paths are resolved against the real
   binary).
4. `make install` — skills are embedded at build time, so a content change
   is a rebuild. That is deliberate: the guidance can never drift from the
   artifact you installed.

## Known limitations

- The frontmatter parser is a restricted subset (scalars, flow/block
  lists), stdlib-only by choice. `skills check` rejects anything outside
  it.
- `check` validates command *paths* against the real binary, not every
  flag in every body (the design's per-target help adapters, docs/06).
- Collision policy (docs/06): the shim does not probe `<real> skills
  --help` on every call — the skills path is the agent's hot path, and
  agent-bot is a node-backed CLI. Instead, `skills check` fails if the
  real target ever answers `skills --help`, so the collision is caught at
  validation time rather than silently shadowing a native surface.
- `--skill` is supported only as a leading run of flags, delimited-block
  form only (no JSONL). The canonical path — harness fetches
  `agent-bot skills <name> --json`, runs the command separately — is
  unaffected.
- No usage tracking, `next`, or `feedback`: designed, deferred (docs/04).
- Windows: pass-through uses `os/exec` with exit-code propagation instead
  of `syscall.Exec`; treat as a fallback, Unix is the tested path.
- `examples/gh/` still carries its own copy of the shim code (it is the
  self-contained reference example). Migrating it onto `shim/` is the
  follow-up that retires that duplication; until then, changes to shim
  behavior must be made in `shim/` and backported to `examples/gh/`.
