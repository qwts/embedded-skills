# Add a new CLI target and shim

Use this when an arbitrary executable needs an embedded-skills corpus but
has no shim in this repository yet. The current runnable reference is
`examples/gh/`; the design target is the generic layout in `docs/06`.
Do not pretend a `targets/<cli>/` implementation exists until it is built.

Read first:

- `docs/01-cli-surface.md`
- `docs/02-skill-format.md`
- `docs/06-shim-pattern.md`
- `docs/09-implementation-and-evals.md`
- `examples/gh/README.md`

## 1. Define the target narrowly

Record facts from the installed executable:

```console
command -v <target>
<target> --version
<target> --help
```

Identify:

- executable basename (the installed shim artifact must use exactly this);
- version string and a bounded version-probe expression;
- operating systems that must be tested in this change;
- authentication/config side effects of harmless read-only commands;
- an existing wrapper already in front of the target (`type -a <target>`,
  symlinks, script shebangs).

The only hard requirement is an executable resolvable on `PATH`. Do not
choose a plugin, extension, completion, or vendor SDK mechanism instead.
Target-specific plugin dispatch is forbidden by the design even when the
first target happens to offer one.

## 2. Choose the implementation shape without duplicating blindly

For the first target beyond `gh`, stop and decide explicitly:

- extract the tested shim behavior from `examples/gh/` into reusable
  target-independent code, leaving per-target data as corpus + manifest; or
- keep a second self-contained example only if the request is explicitly
  an example/prototype, and record that duplication as debt in the PR/report.

Production targets should not accumulate independently drifting copies of
proxy, parsing, rendering, export, and validation logic. Preserve the
tested invariants from the reference implementation:

- compiled binary, embedded corpus, no runtime download;
- `skills` handled locally; all other invocations proxied;
- real-binary resolution skips self using canonical identity, never
  string-only comparison;
- visited-path loop prevention for stacked wrappers;
- clean command stdout on the canonical harness path;
- explicit real-binary override for managed environments;
- version probing only on the skills path;
- useful diagnostics available without shadowing a target invocation.

Do not rewrite these mechanisms from memory. Read the current code and
tests, then extract or adapt them with tests moving alongside behavior.

## 3. Create the target corpus before broad coverage

Author 3–5 high-value workflow skills first, chosen by failure frequency
and risk—not one per subcommand. Build a real chain only when the linked
skills exist.

For each skill, follow `add-skill.md` and verify every command/flag/field
against the binary you observed in step 1. Avoid generic skills copied from
another target with nouns replaced; expected output and failure branches
are target-specific facts.

Set the target version range from tested versions only. A skill requiring
newer behavior records its own honest `min_version`.

## 4. Build the manifest and validation path

The target needs a generated manifest/index equivalent to the reference:

- target name and tested version range;
- version probe (`args`, `regex`) proven against actual output;
- skill path, routing metadata, command paths, and chain links;
- checksums or another existing staleness gate if the shared
  implementation already has one.

Validation must at minimum reject:

- missing/mismatched names, descriptions over budget, invalid level/risk;
- duplicate skill names;
- dangling prerequisites/next/reference files;
- oversized skill bodies or aggregate catalog;
- command paths that do not resolve against the tested real target;
- stale generated manifest content.

Flag validation against help output is stronger and follows docs/06 when
the repository's current validator supports it. If an example supports
only command-path validation, document that limitation in its README
rather than claiming the full design check.

## 5. Test proxy fidelity before workflows

Test the transparent path first. Against a harmless fixture or the real
target's read-only commands, compare shim and real binary for:

- stdout and stderr bytes;
- exit status;
- argv containing spaces, `--`, and flags resembling shim flags;
- stdin pass-through where applicable;
- working directory and environment inheritance;
- behavior when another wrapper is already in front of the target;
- real-binary override and not-found failure;
- target's own `skills` subcommand collision (real behavior must win per
  docs/06).

Then exercise the skills surface: list, JSON, show, search, unknown skill,
export, check, version probe, and optional `--skill` only if implemented.
Do not use a mutating target command as a proxy smoke test.

## 6. Write the target README

The target is not done until a reader can install and remove it without
reading source. Document:

- prerequisites and tested target versions;
- exact build command and artifact location;
- install directory and PATH line (printed, never auto-edited into a shell
  profile);
- verification commands distinguishing shim and real binary;
- how agents discover it (project instruction snippet);
- bypass/override and uninstall;
- known limitations and platforms actually tested;
- how to add a skill, regenerate the manifest, and run validation.

## 7. Evaluate, then hand off to installation

Run at least one workflow evaluation per docs/09 and `add-skill.md`.
Record whether an agent discovers the catalog without being handed the
skill name. Build success and `check` success do not prove discovery.

Only after the artifact and its tests pass, load `install-shim.md` to put
it on this machine. Keep development checkouts and installed artifacts
conceptually separate: rebuilding is the only way embedded guidance changes.

## Completion report

Report target/version observed, implementation shape chosen and why,
corpus skills added, proxy-fidelity evidence, corpus validation,
public-surface tests, platform limits, and any eval or flag-validation
gaps remaining.
