# Implementation and Evals

Scope: the native integration (doc 03) and the shim (doc 06) share one
implementation core — an embedded corpus, a renderer, and (for the shim) a
faithful proxy. This doc covers what makes the proxy trustworthy and the
corpus measurable. Sized for a small v1: standard library first,
dependencies only for YAML.

## Proxy fidelity — the trust foundation

Prefer `syscall.Exec` on Unix: it replaces the process image entirely, so
signals, TTY, exit status, and file descriptors need no forwarding code. On
Windows (no `exec`), use `os/exec` with explicit stdin/stdout/stderr wiring,
signal forwarding, and exit-code propagation, and test that path separately.

The shim adds zero observable behavior on the proxy path: no banners, no
logging by default, no env mutation except removing its own bookkeeping vars.

## Real-binary resolution

- Walk `PATH` in order; skip the entry that resolves to the shim itself.
  Compare canonical paths (`os.SameFile`), not strings — symlinks and
  relative PATH entries otherwise bite.
- If the only candidate is the shim itself, error clearly instead of looping.
- Honor a per-target override for managed environments
  (`SKILLS_REAL_BIN=/usr/bin/gh`, or per-target `SKILLS_REAL_GH`).
- Never resolve via a shell; never invoke through `sh -c`.

## Version probing

Only on the skills path, never the proxy path — probing would tax every real
command. Per-target config in the manifest:

```json
"version_probe": {"args": ["--version"], "regex": "(\\d+\\.\\d+(\\.\\d+)?)"}
```

Run with a short timeout; failure → `detected_version: null`, never an
error. If the detected version is below a skill's `min_version`, mark
`version_ok: false` per skill in `--json` and print a warning line on stderr
in text mode. Never silently serve guidance written for a newer target than
the one installed.

## Validation thresholds (`tools/lint-skills`, CI)

- YAML parses; required fields present; `name` matches the directory and the
  `^[a-z0-9]+(-[a-z0-9]+)*$` grammar; `description` ≤ 200 chars.
- `risk` in the known set; `min_version` parses as a version when present.
- `SKILL.md` exists; `references/` links resolve; no orphan references.
- Budgets: `SKILL.md` ≤ 200 lines and ≤ 20 KB; rendered catalog per target
  ≤ 8 KB (~2k tokens).
- Lint: no absolute paths to author machines, no obvious secrets patterns,
  no "ignore previous instructions"-style phrasing.

## Testing — three layers, increasing cost

1. **Unit/golden tests.** Catalog parsing, rendering, resolver against
   synthetic PATH layouts, interception truth table. Golden files for
   rendered catalog/JSON compared byte-for-byte, so renderer changes are
   reviewable diffs.
2. **Proxy-fidelity tests.** Wrap a test binary that echoes argv/env, reads
   stdin, exits 42, dies by signal; assert the shim is indistinguishable.
   This suite is the product's trust foundation — invest here first.
3. **Skill evals (manual → automated).** Per skill, a recorded task (repo
   fixture + user request) run by an agent twice: with `--help` only, and
   with the skill loaded. Score: completed? destructive mistakes? retries?
   discovery — did the agent find the skill without being told the exact
   commands? Start as checked-in transcripts reviewed by a human; automate
   once the fixture format settles. **A skill that doesn't measurably beat
   `--help`-only doesn't ship.**

## Build and release

- Static binaries per GOOS/GOARCH (`CGO_ENABLED=0`); embed the full corpus;
  stamp version + git SHA via `-ldflags`; `-trimpath` and a pinned toolchain
  for reproducible builds.
- Sign releases (cosign/minisign) and publish checksums. A PATH shim in front
  of a real CLI is a high-trust install; users must be able to verify it.
- Install creates the shim directory + per-target links and prints the exact
  PATH line to add — never edits shell rc files itself.

## Explicit non-features for v1

- No telemetry and no usage tracking. Doc 04's tracking/`next`/`feedback`
  design is explicitly deferred behind eval evidence; v1 writes nothing
  except explicit `export`.
- No runtime skill downloading or updates (upgrade = new binary).
- No bundled executables/scripts in skills (extraction = attack surface).
- No user overlay skills directory yet — design the catalog loader so a
  second FS layer *could* be stacked later, but ship embedded-only.
