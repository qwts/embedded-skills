# The Shim Pattern: Skills for Any Executable

## The insight

This repo's delivery mechanism is the **shim**, and only the shim:

| Shape | Who ships it | Surface | Example |
|-------|--------------|---------|---------|
| Shim (this repo) | embedded-skills, for any executable | Full `skills` surface via a PATH shadow | `gh skills`, `git skills`, `kubectl skills`, `<literally-any-cli> skills` |
| Native (not this repo) | A CLI's own authors, if they ever adopt the contract | `skills` subcommand built in | the reference integration sketch (doc 03), kept as the upstreaming endgame (doc 07) |

gh is the shim case: the corpus describes gh's workflows, but the `skills`
command is delivered by our executable standing in front of gh — gh itself
knows nothing about it. The same pattern works for **any** executable: the
shim logic is identical; only the corpus and the target name change. There
is exactly one requirement on the target: it is an executable resolvable
on PATH.

**Explicitly rejected: per-CLI plugin/extension mechanisms** (gh
extensions, `git-<command>` executables, kubectl plugins, and their kin).
They exist on some CLIs, with different semantics, install paths, and arg
passing on each — and not at all on most CLIs. A delivery mechanism that
depends on the target's plugin dispatch cannot be the universal contract,
so this design does not use, require, or document any of them. The PATH
shadow below is the only placement.

## The one placement: PATH-shadow shim

An executable named exactly `gh` placed earlier on PATH than the real one:

1. Inspect argv. If the first argument is `skills`, handle it in the shim.
   If the shim includes optional `--skill` support and `--skill` appears
   anywhere before `--`, handle that too.
2. Otherwise exec the real binary with the original argv, unchanged.

Parsing follows the doc-01 contract: both `--skill <name>` and
`--skill=<name>` forms, the skills layer owns `--skill` before `--`, and
everything after `--` passes through untouched.

**Collision policy.** If the target itself grows a `skills` subcommand, the
real behavior wins: before serving, the shim probes the real binary
(`<real> skills --help`); if the target answers, the shim proxies instead. An
alternate intercepted name is available via env (`SKILLS_SUBCOMMAND`) for
environments that need both.

**`--help` augmentation (opt-in, off by default).** The shim can proxy
`--help` and append one line — `Agent skills for this CLI: run 'gh skills'.`
Agents read `--help` constantly, so this converts existing behavior into
discovery. It mutates another program's output, so it stays a conscious
opt-in, limited to help text (see `07-discovery-and-risks.md`).

This delivers the full doc-01 contract plus the optional `--skill`
convenience — including `gh --skill pr-flow pr create ...` — for every
target, with no cooperation from the target. The canonical harness path
(fetch with `gh skills <name> --json`, run the command separately) needs
only the `skills` interception; `--skill` handling is the one part a
shadow shim may omit without breaking the contract.

Rules for a correct shadow shim:

- **No recursion.** Resolve the real binary by walking PATH in order and
  skipping the entry that resolves to the shim itself — compare canonical
  paths (`os.SameFile`), not strings; symlinks and relative PATH entries
  otherwise bite. If the only candidate is the shim itself, error clearly
  instead of looping. Honor an env override for managed environments
  (`SKILLS_REAL_BIN=/usr/bin/gh gh ...`). Never resolve via a shell.
- **Pass-through fidelity.** Inherit stdio (no buffering), propagate the exit
  code, forward signals. The shim must be invisible for everything it does not
  handle — including the target's own `--help` and shell completion.
- **Documented PATH change.** The installer prepends a shim directory (e.g.
  `~/.local/share/skills-shims/bin`) to PATH and says so.

## Shim behavior contract

```text
gh skills ...            → shim handles (list / show / search /
                           export / check)
gh --skill <n> ...       → optional: shim prints delimited skill blocks, strips
                           the flag, execs real gh with remaining args
gh <anything else>       → exec the real gh, untouched
```

Invoked under its build-artifact name (i.e. not as a target name), the
shim reports diagnostics instead of shadowing: shim placement,
real-binary resolution, detected target version.

## Corpus is per-target; shim logic is generic

One shim codebase, many targets. The corpus describes the target's workflows;
the manifest records which target and which target version it was built
against:

```json
{
  "version": 1,
  "target": "gh",
  "target_version": ">=2.40",
  "skills": []
}
```

Drift now crosses two artifacts, so:

- At runtime the shim probes the real binary's version using the manifest's
  `version_probe` (`{args: [--version], regex: ...}`, short timeout, failure
  → `detected_version: null`, never an error) and warns when the installed
  target is outside `target_version`. Per-skill `min_version` (doc 02)
  tightens this: `--json` marks `version_ok: false` per skill and text mode
  prints a warning on stderr. Never silently serve guidance written for a
  newer target than the one installed. Probing happens only on the skills
  path — never on the proxy path.
- In CI, lint installs the pinned target version and runs
  `<target> skills check`, validating every `commands` path and every flag
  used in a skill body against that version's real command tree.

"The target's command tree" is extracted by a per-target help adapter: a
small, pinned piece of code that runs `<target> --help` and
`<target> <cmd> --help` and parses command paths and flags. Adapters are
fragile by nature — help output is prose — so each adapter is version-pinned
in CI and fails loudly when the target's help format changes.

## Platforms

Unix first. The proxy path prefers `syscall.Exec`, which replaces the process
image entirely — signals, TTY, exit status, and file descriptors need no
forwarding code. Windows has no `exec`: use `os/exec` with explicit
stdin/stdout/stderr wiring, signal forwarding, and exit-code propagation,
and test that path separately (see `09-implementation-and-evals.md`).

## Corpus maintenance: tracking the upstream target

The lint mechanism catches drift; this process acts on it. When the target
releases a new version:

1. A CI job installs the new target version and runs
   `<target> skills check` against the corpus. Any failure names the skills
   that drifted.
2. Corpus maintainers update the affected skills — commands, flags, examples —
   and review `prerequisites` / `next` chains for new capabilities.
3. The shim releases: new manifest with the bumped `target_version` range,
   new binaries, per the build steps below.

`target_version` is a range (e.g. `>=2.40`), so one shim release covers a
span of target versions; the runtime version check above is what warns when a
user's target falls outside it.

## Repo layout (multi-target)

```text
embedded-skills/
├── shim/                    # the generic shim — one codebase
│   └── main.go              # argv inspect, skills surface, exec passthrough
├── targets/
│   ├── gh/
│   │   ├── skills/          # gh workflow corpus (doc 02 format)
│   │   └── manifest.json    # generated; target: gh
│   ├── git/
│   │   └── skills/
│   └── ...
├── tools/
│   ├── build-manifest       # walks targets/<t>/skills/
│   └── lint-skills          # per-target validation against pinned target version
└── tests/
```

## Build and install

1. Write the corpus under `targets/<target>/skills/` (doc 02 format; a
   complete gh skill lives at `examples/gh/pr-flow/SKILL.md`).
2. Generate `targets/<target>/manifest.json` at build time.
3. Build the shim binary per target, embedding that target's corpus. The
   installed artifact is named **exactly** `<target>` — that name is the
   whole mechanism.
4. CI matrix over targets: manifest fresh, lint against the pinned target
   version, smoke test in a clean environment.
5. Install: place the per-target artifacts in one shim directory and
   prepend that directory to PATH. That is the only install step, for
   every target. No target-specific install mechanism is invoked at any
   point.

## What does not change

The skill format (doc 02), the CLI surface (doc 01), and the
progressive-learning model (doc 04) are target-independent. Adding a
target is always the same operation: a corpus, a manifest, and a shim
artifact named after the target. If a CLI's own authors later implement
the contract natively, the shim's collision policy (real CLI wins) makes
the handover automatic.
