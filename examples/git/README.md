# Example: `git` skills shim

A working example of the embedded-skills design (`../../docs/`), targeting
Git. Build one Go binary, install it under the name `git` in a directory at
the front of your `PATH`, and your agent gains:

```console
$ git skills                  # catalog of git workflow skills
$ git skills soul-worktree-flow   # one full workflow, verbatim SKILL.md
$ git skills --json           # machine-readable catalog for harnesses
$ git status                  # ← everything else: the real git, untouched
```

Nothing about this is git-specific except the corpus. The binary is a
PATH-shadow shim (docs/06): it answers `skills` (and the optional leading
`--skill`) itself from skills **embedded in the binary at build time**, and
execs the real `git` for every other invocation. No plugin dispatch, no
cooperation from git.

## What's inside

```text
examples/git/
├── main.go  catalog.go  frontmatter.go  proxy.go  commands.go
│                            the shim (standard library only, no dependencies)
├── exec_unix.go / exec_other.go
│                            pass-through: syscall.Exec on Unix
├── cmd/genmanifest/         build-time tool: regenerates manifest.json
├── manifest.json            generated index of the corpus (do not edit)
├── skills/                  the corpus — this is the part worth copying
│   ├── soul-worktree-flow/SKILL.md   beginner
│   └── safe-history/SKILL.md         intermediate
├── Makefile                 build / install / test / check
└── main_test.go
```

Skill format and authoring rules: `../../docs/02-skill-format.md` and
`../../docs/08-skill-authoring.md`. The two bundled skills form one chain
(`soul-worktree-flow` → `safe-history`) and were written against git 2.54
(manifest: `>=2.54`).

## Install

Requires Go (to build) and a real `git` already installed.

```console
$ cd examples/git
$ make install
```

That compiles `dist/git` and copies it to
`~/.local/share/skills-shims/bin/git` (override with `SHIM_DIR=...`).
It does **not** edit your shell profile. Add the directory to the front
of `PATH` yourself, in whatever environment your agent runs in:

```console
$ export PATH="$HOME/.local/share/skills-shims/bin:$PATH"
$ which git        # must print the shim path, not your real git
$ git skills       # catalog from the shim
$ git --version    # still the real git, proxied
```

Uninstall: `make uninstall`, or just remove the directory from `PATH`.
The real `git` was never modified.

## Add or edit a skill

1. Create `skills/<name>/SKILL.md` (copy an existing skill as the template;
   rules in docs/08).
2. `make manifest` — regenerates the index + checksums.
3. `make test && make check` — unit tests plus corpus validation against
   your installed git (command paths are resolved against the real binary).
4. `make install` — skills are embedded at build time, so a content change
   is a rebuild.

## Example-grade limitations (vs. the full design)

- The frontmatter parser is a restricted subset (scalars, flow/block
  lists), stdlib-only by choice. `skills check` rejects anything outside it.
- `check` validates command *paths* against the real binary; git help exits
  129 for subcommands that require arguments, which the validator treats as
  "command exists".
- `--skill` is supported only as a leading run of flags, delimited-block
  form only (no JSONL). The canonical path — harness fetches
  `git skills <name> --json`, runs the command separately — is unaffected.
- No usage tracking, `next`, or `feedback`: designed, deferred (docs/04).
- Windows: pass-through uses `os/exec` with exit-code propagation instead
  of `syscall.Exec`; treat as a fallback, Unix is the tested path.
