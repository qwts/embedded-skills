# CLI Surface: `skills` and `--skill`

## Design principles

1. `skills` is the agent-facing counterpart of `--help`. Help is prose for
   humans; skills are workflows for agents. They complement each other.
2. Everything has a `--json` form. Agents parse structure; humans read tables.
   The JSON schemas are a versioned API — change them only with a major
   version bump.
3. `skills` is fast, stable, and side-effect free. It sits on the agent's hot
   path: it runs before nearly every other decision.
4. Skills never execute code and never mutate CLI state. The `skills`
   surface is read-only and offline; the sole exception is the explicit
   `export` command, which writes only where told. No telemetry, no
   tracking in v1 (see "Deferred" in `04-progressive-learning.md`).
5. Command stdout stays clean. Content goes to stdout, diagnostics to
   stderr, no color when not a TTY, no pager, deterministic ordering.
   Skill content is never interleaved with a real command's output on the
   recommended machine path (the one documented exception is the optional
   `--skill` convenience flag, below).

## `cli skills` — discovery

```text
cli skills                      list all skills
cli skills --json               list, machine-readable
cli skills --level <level>      filter by beginner | intermediate | advanced
cli skills --tag <tag>          filter by tag
cli skills <name>               print one full skill document, verbatim
cli skills <name> --json        one skill, structured
cli skills search <query>       substring match over name, description, tags, triggers
cli skills export <name> [--to <dir>]
                                write the skill to disk for file-based harnesses
cli skills check                validate the corpus against the CLI (dev/CI)
```

Deferred beyond v1 (designed in `04-progressive-learning.md`, not part of
this contract): `skills next` (usage-ranked recommendations), `skills
feedback` (failure recording), and the local usage tracking both need.
They ship only after evals show agents discover and use the v1 surface.

### `cli skills` (human output)

Fixed-width table, sorted by level then name:

```text
NAME            LEVEL         DESCRIPTION
branch-flow     beginner      Create, switch, and clean up branches
commit-flow     beginner      Stage, review, and land one atomic change
pr-flow         intermediate  Open a pull request with a complete description
release         advanced      Tag, changelog, and publish a release
```

### `cli skills --json`

```json
{
  "version": 1,
  "detected_version": "2.47.1",
  "skills": [
    {
      "name": "commit-flow",
      "description": "Stage, review, and land one atomic change",
      "level": "beginner",
      "tags": ["vcs", "commit"],
      "commands": ["status", "diff", "commit"],
      "version_ok": true
    }
  ]
}
```

`detected_version` is the probed target version (skills path only,
`06-shim-pattern.md`); `null` when unknown, never an error. `version_ok`
is `false` when the detected version is below a skill's `min_version`, so
a harness can de-prioritize it and text mode can warn on stderr. Never
silently serve guidance written for a newer target than the one installed.
Both fields are additive within schema version 1.

### `cli skills <name>`

Prints the skill's `SKILL.md` verbatim — frontmatter and body. Verbatim output
matters: the document is already in the shape agents are trained to consume, so
the agent can take it into context unchanged.

### `cli skills <name> --json`

```json
{
  "version": 1,
  "name": "commit-flow",
  "description": "Stage, review, and land one atomic change",
  "level": "beginner",
  "tags": ["vcs", "commit"],
  "commands": ["status", "diff", "commit"],
  "prerequisites": [],
  "next": ["pr-flow"],
  "body": "# Commit flow\n\n## When to use\n...",
  "references": ["references/edge-cases.md"]
}
```

`references` lists the skill's support files (see `02-skill-format.md`); the
agent loads them only if the body points at them.

### Session bootstrap

A harness can inject the layer-1 list into the agent's context once, at
session start, instead of waiting for the agent to discover it:

1. Harness runs `cli skills --json`.
2. Harness injects the result — names, descriptions, levels, tags — into the
   agent's context, the same way skill metadata sits in a system prompt.
3. The agent now knows the curriculum exists before it runs anything, and
   calls `cli skills <name>` only when a workflow is relevant.

Bootstrap output is layer 1 only: small, stable, and cheap enough to keep in
context for the whole session.

The probing contract, in full:

1. Run `cli skills --json`. Exit 0 with the versioned schema → ingest
   `skills[]` as the metadata tier.
2. Exit 3 or non-JSON → not skill-enabled; fall back to `--help`.
3. On task match, run `cli skills <name>`, then a `references/` file only if
   the body points at it.

### `cli skills search <query>`

Substring match over `name`, `description`, `tags`, and `triggers`. Output in
catalog format, matches only. Deterministic order (by name); no fuzzy
matching, no ranking — deterministic beats clever.

## `cli --skill <name>` — loading (optional, non-canonical)

The canonical loading path is the harness calling `cli skills <name>` (or
`--json`) itself and invoking the command separately — see "Recommended
machine path" below. `--skill` is a convenience for simpler harnesses,
scripts, and humans. It is not part of the probing contract, harnesses
must not require it, and a shim may omit it entirely without breaking
the contract (doc 06). What follows specifies it for implementations
that choose to offer it.

```text
cli --skill <name> [--skill <name> ...] <command> [args...]
```

Semantics: before running `<command>`, the CLI prints each requested skill
document to stdout inside a delimited block, then runs the command normally.
The calling harness reads the blocks as context to inject, not as command
output.

```console
$ cli --skill commit-flow commit -m "fix: off-by-one"
<!-- skill: commit-flow -->
---
name: commit-flow
...
---
# Commit flow
...
<!-- /skill: commit-flow -->
[1 file changed, 2 insertions(+)]
...
```

Rules:

- Multiple `--skill` flags compose; blocks print in flag order.
- Both `--skill <name>` and `--skill=<name>` forms are accepted.
- Everything after `--` belongs to the command and passes through untouched.
  The skills layer owns `--skill` before `--`; a command that has its own
  `--skill` flag receives it after `--`
  (`cli --skill pr-flow -- cmd --skill native`).
- The delimiter lines are exactly `<!-- skill: <name> -->` and
  `<!-- /skill: <name> -->`, so a harness can split stdout reliably.
- With `--json`, each skill is emitted as one JSON object on its own line
  (JSONL) before the command runs, using the `cli skills <name> --json`
  schema. The harness reads exactly N objects — one per `--skill` flag, in
  order — and treats everything after them as command output. There is no
  terminator line; command output stays byte-identical.
- An unknown skill name is an error: exit code 2, message on stderr, and the
  block is not printed.

### Recommended machine path

Harnesses call `cli skills <name> --json` themselves, inject the result
into context, and then invoke `cli <command>` normally. That keeps
command stdout byte-identical to the un-skilled invocation and makes the
injection explicit and auditable. This is the path the probing contract
(above) standardizes; `--skill` is strictly a fallback for environments
that cannot do two invocations.

## `cli skills export` — harness-native loading

Materializes a skill as a plain directory (`SKILL.md` + `references/`) so
file-based harnesses can load it natively — the channel for harnesses that
only read skill directories (`.claude/skills/`, `.agents/skills/`, …).

```text
cli skills export <name> [--to <dir>]
```

- Default target: `./.embedded-skills/<name>/`. Prints the path written.
- This is the only `skills` command that writes, and it refuses to overwrite
  a directory it didn't create (exports are stamped with a marker file).
- `--strict` drops this format's extension frontmatter fields (doc 02,
  "Conformance") for harnesses whose validators reject unknown keys; the
  body is exported unchanged.

## Errors and exit codes

| Situation | Exit | Behavior |
|-----------|------|----------|
| `cli skills <unknown>` | 2 | stderr: `unknown skill "foo"` plus `did you mean: commit-flow?` suggestions |
| `cli --skill <unknown> ...` | 2 | same; the command does not run |
| `cli skills check` failures | 1 | lists every violation (used in CI) |
| `cli skills` with an empty corpus | 3 | no skills embedded for this target — distinct so harnesses can probe cheaply |
| success | 0 | — |

The "did you mean" pattern is deliberate: agents are trained on git's version
of it and handle it well.

## Non-goals

- Skills do not wrap or sandbox command execution.
- No network access, ever, from the `skills` command.
- No skill installation or marketplace in v1. The corpus ships with the binary.
- Not an MCP server. MCP gives agents tools to call; skills give agents
  knowledge of how to use the CLI they already have. They compose; they are
  not substitutes.
