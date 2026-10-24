# Build and install a skills shim locally

Use this when a target shim already exists and the task is to put it in
front of the real CLI on this machine, verify it for an agent, troubleshoot
it, or remove it. The worked commands below use `examples/gh/`; substitute
the target's own documented build/install commands when they differ.

## 1. Preflight: know what is already in front of the target

Record the current resolution before changing anything:

```console
type -a <target>
command -v <target>
<target> --version
```

Follow symlinks. If the first result is a script or another wrapper, read
enough of it to understand how it finds its own "real" binary. Stacked
wrappers can exec each other unless the shim implements the visited-path
loop prevention in docs/06. Do not remove or bypass an identity/token
wrapper to make installation easier.

Identify and preserve:

- the true target executable or the wrapper that must remain in the chain;
- any supported real-binary override variable;
- the shell/environment the agent actually launches commands in (an
  interactive shell profile may not be read by that agent).

## 2. Build and validate before installing

For the gh example:

```console
cd examples/gh
make test
make check
make build
```

Use only commands documented by the target's README/Makefile. If the
toolchain is missing, report that and stop; do not substitute a shell or
Python script for the compiled artifact.

Run the built artifact's own validation against the installed real target.
For gh, `make check` builds and runs `./dist/gh skills check`. A build by
itself does not validate the corpus.

## 3. Install into one explicit shim directory

For gh:

```console
make install
```

The current example installs to
`~/.local/share/skills-shims/bin/gh` unless `SHIM_DIR` is overridden.

Rules:

- The installed artifact's basename must be exactly `<target>`.
- Use one clearly named shim directory; do not drop files into system
  directories or the real target's directory.
- Never edit shell profiles or global environment files. Give the user
  the exact `export PATH="<shim-dir>:$PATH"` line to place where their
  agent launches.
- For your own verification shell, a process-scoped `PATH` prefix is the
  right test and avoids persistent mutation.

## 4. Verify resolution and proxy behavior

In a fresh shell or command environment with the shim directory first:

```console
command -v <target>
<target> --version
<target> <harmless-read-only-command>
```

Confirm:

- `command -v` is the shim artifact;
- `--version` and the read-only command are byte-for-byte the real
  target's behavior;
- a bogus command returns the same exit status through the shim as direct;
- no process or wrapper loop occurs when another wrapper was already
  present.

Use the target's diagnostics mode if it has one. For the gh artifact,
invoke the same binary under a non-`gh` name with `doctor`; do not infer
resolution from `command -v` alone.

## 5. Verify the agent-facing surface

Exercise the canonical machine paths, not only the human list:

```console
<target> skills
<target> skills --json
<target> skills <one-real-skill>
<target> skills <one-real-skill> --json
<target> skills search <one-trigger>
<target> skills <unknown-name>    # expect exit 2 and useful stderr
```

Parse JSON with a real parser. Confirm the detected version, per-skill
`version_ok` values if offered, references, and chain links. If export is
part of the task, export to a temporary/test directory first and confirm
overwrite refusal for a directory the shim did not create.

Then add the discovery pointer only where the user asked for it: a
project's `AGENTS.md`/`CLAUDE.md` should instruct the agent to run
`<target> skills` before multi-step work and load matching skills. The
snippet in `docs/examples/agents-md-snippet.md` is the starting point.
Installing a shim without that pointer does not establish that agents will
discover it spontaneously; report discovery as harnessed, not automatic.

## 6. Troubleshoot in this order

1. **Wrong binary:** `type -a <target>`; check hashed command paths in the
   shell and whether the agent inherits the same `PATH`.
2. **Shim serves but real commands fail:** compare direct absolute real
   path vs shim for stdout/stderr/exit code; check real-binary override.
3. **Loop or process explosion:** inspect for an existing wrapper, then
   verify visited-path handling. Stop the test processes you started; do
   not kill unrelated user processes.
4. **Skills empty or stale:** rebuild—embedded content changes only at
   build time—then rerun manifest generation and corpus check.
5. **Version warning:** believe the probed version. Do not serve a skill
   silently when the installed target is below its `min_version`.
6. **Target grew its own `skills`:** real behavior wins under the design's
   collision policy. Use the documented alternate/interception escape only
   if already implemented; otherwise report the collision rather than
   inventing one locally.

## 7. Remove or bypass cleanly

Removal is reversing exactly your install:

- use the target's documented uninstall command (gh example:
  `make uninstall`), or remove only `<shim-dir>/<target>` you installed;
- remove the shim directory from `PATH` where the user added it;
- rerun `command -v <target>` and `<target> --version` to prove the
  original chain is restored.

For a one-off bypass, invoke the recorded real path directly or use the
shim's documented override. Do not rename another wrapper or rewrite its
configuration as a shortcut.

## Completion report

Report artifact path, real binary/wrapper chain observed, build and check
results, exact public inputs verified (catalog, JSON, one skill, proxy),
where the user must add `PATH` if it is not already effective for their
agent, and anything that remains unverified.
