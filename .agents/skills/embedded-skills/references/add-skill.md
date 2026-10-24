# Add or edit a skill for an existing target

Use this when the CLI already has a runnable corpus. Today that means
`examples/gh/` and `targets/agent-bot/`. If the requested CLI has no corpus
or shim, stop and load `add-target.md` first.

Read these design references before authoring:

- `docs/02-skill-format.md`
- `docs/08-skill-authoring.md`
- `docs/04-progressive-learning.md` (levels and chains)

## 1. Justify the skill from evidence

Do not add a skill merely because a command exists. Identify a recurring,
multi-step workflow that agents get wrong or must reconstruct through
trial and error. Prefer evidence in this order:

1. A user-provided request or failure transcript.
2. A repeated retry/error loop in an available agent transcript.
3. A destructive or easily-confused workflow where `--help` gives flags
   but not safe ordering, checks, or recovery.

Write the intended outcome in one sentence before creating files. One
skill has one job and one observable definition of done.

## 2. Establish the real command surface

From `examples/gh/` (or the relevant existing target directory), inspect the
installed target. For `gh`, for example:

```console
gh --version
gh <command> --help
```

When a local shim is installed, `gh` may be the shim. Verify which binary
is being observed. Use its documented doctor/override mechanism or an
absolute path to the real target when you need target facts rather than
shim behavior.

For every command in the planned workflow, verify:

- command path;
- non-interactive form;
- flags and required arguments;
- machine-readable output and field names;
- expected success output or state;
- known failure text where the procedure branches on it.

Never write flags, fields, prompts, or output text from memory.

## 3. Name and place the skill

Use a verb-object, lowercase-hyphen slug (`safe-rebase`, `review-flow`).
Create exactly:

```text
skills/<name>/SKILL.md
```

Add `skills/<name>/references/<topic>.md` only for detail that the main
workflow should load conditionally. Do not put task-critical steps only
in a reference.

The directory name and frontmatter `name` must match.

## 4. Write frontmatter

Use only the field set and grammar in `docs/02-skill-format.md`.

- `description` is the routing text: state **what** and **when**, keep it
  one line and within the documented length limit.
- `triggers` use the user's vocabulary, not only the CLI's nouns.
- Assign `level` by judgment and composition, not by how advanced the
  topic sounds.
- Set `risk` honestly. Anything whose main action is hard to undo is
  destructive even if individual commands look ordinary.
- `commands` contains every target command path the body relies on.
- `requires` contains checkable preconditions, written as observations an
  agent can make.
- Set `min_version` from the version actually inspected and tested.
- Link only existing skills in `prerequisites` and `next`.
- Set `completes_with` only when one command path unambiguously marks the
  workflow complete; in v1 it is reserved metadata, not behavior.

Do not invent new frontmatter fields. If the format genuinely cannot
express the need, stop and raise a design change rather than adding a
field silently.

## 5. Write the body as an executable procedure

Use the fixed sections in `docs/02-skill-format.md`. Each step provides:

1. the exact command or filesystem check;
2. the expected observation;
3. the branch when the observation differs.

Requirements:

- Start with checks, not assumptions.
- Prefer stdout-stable, non-interactive, machine-readable target forms.
- Include negative knowledge: deprecated spellings, destructive variants,
  and commands that look right but are wrong for this target/version.
- Mark the irreversible step explicitly.
- End with a verification command and the state that proves success.
- Link conditional recovery or depth as `references/<topic>.md`.
- Do not copy flag reference material out of `--help`.
- Do not include repo-, org-, or user-specific policy. That belongs in the
  user's project instructions, not a target corpus.
- Do not embed secrets, internal URLs, invented reviewers/assignees, or
  reproduction claims that were not observed.

## 6. Regenerate and validate

For a target with a Makefile (`examples/gh/`, `targets/agent-bot/`):

```console
cd <target-dir>
make manifest
make test
make check
```

Fix the skill or manifest source, never hand-edit `manifest.json` beyond
running its generator. `make check` must validate the corpus against the
currently installed real target. If validation cannot run because the
target or toolchain is missing, report that exact missing evidence; do
not call the skill validated.

## 7. Exercise the served skill

Build and install/test the shim using `install-shim.md`, then verify the
public surface rather than reading only the source file:

```console
<target> skills
<target> skills <name>
<target> skills <name> --json
<target> skills search <trigger>
```

Confirm the description makes the skill discoverable, chain links resolve,
references are listed, and the byte output is the authored document.

## 8. Evaluate before calling it ready

Use at least one realistic task fixture and, when an agent run is
available, compare:

- `--help`-only: completed, retries, destructive mistakes;
- with the skill loaded: completed, retries, destructive mistakes, and
  whether the agent discovered the skill without being given its name.

Store or report the evidence the repository's current eval practice
requires. If no eval harness exists yet, report the manual scenario and
observed result plainly, and name automated evaluation as missing. A
skill that does not improve on `--help`-only should be revised or not
added.

## Completion report

Report the skill path, target version observed, generator/test/check
commands and results, the public-surface exercise performed, and any
validation or eval steps not completed.
