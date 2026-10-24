# Skill Format

A skill is a directory containing one `SKILL.md` file — YAML frontmatter plus
a Markdown body — and optional support files. This shape is deliberate: it is
the minimal structured-document convention agents already parse, so a skill
document can move between this CLI and any agent harness without translation.

**Conformance to the Agent Skills specification, stated plainly.** The
[agentskills.io spec](https://agentskills.io/specification) standardizes
`name` and `description` (required) plus `license`, `compatibility`,
`metadata`, and `allowed-tools`. Two of this format's fields are the
spec's; the rest (`level`, `risk`, `tags`, `triggers`, `commands`,
`requires`, `min_version`, `prerequisites`, `next`, `completes_with`) are
**extensions**. List-typed extensions cannot live under the spec's
string-only `metadata` map, so they stay top-level. Consequences, accepted
deliberately:

- Most harness parsers ignore unknown frontmatter keys, so in practice the
  document moves unchanged. A strict validator may reject them.
- `cli skills export` therefore has a strict mode obligation: for
  harnesses known to reject unknown keys, export drops the extension
  fields (the workflow body is untouched; the corpus index lives in the
  manifest, so nothing the wrapper needs is lost).
- Extension fields are routing and validation metadata only. Nothing an
  agent needs to *follow* the workflow may live exclusively in an
  extension field — it must also be readable in the body. That is what
  keeps an exported, extension-stripped skill fully usable.

## Directory layout

```text
skills/
└── commit-flow/
    ├── SKILL.md                  # required: frontmatter + body
    └── references/               # optional: deep-dive docs, loaded on demand
        └── edge-cases.md
```

Support files live in the skill directory and are referenced from the body by
relative path. They are the third disclosure layer (see
`04-progressive-learning.md`): the body stays small, depth lives here.

## Frontmatter schema

```yaml
---
name: commit-flow
description: Stage, review, and land one atomic change
level: beginner
tags: [vcs, commit]
commands: [status, diff, add, commit]
prerequisites: []
next: [pr-flow]
---
```

| Field | Required | Type | Notes |
|-------|----------|------|-------|
| `name` | yes | string | kebab-case, unique, must match the directory name |
| `description` | yes | string | one line, ≤ 200 chars; shown in `cli skills` list output — this is the routing text |
| `level` | no (default `beginner`) | enum | `beginner` \| `intermediate` \| `advanced` |
| `risk` | no (default `mutating`) | enum | `read-only` \| `mutating` \| `destructive` — agents calibrate caution to this marker |
| `tags` | no | string[] | used by `cli skills --tag` and `search` |
| `triggers` | no | string[] | user-vocabulary keywords for `search` — the user's words, not the CLI's (see `08-skill-authoring.md`) |
| `commands` | no | string[] | command paths this skill covers (space-separated, e.g. `pr create`); lint-validated against the target's real command tree |
| `requires` | no | string[] | checkable preconditions — commands whose output decides whether to proceed |
| `min_version` | no | string | minimum target version this skill was written against; overrides the manifest's `target_version` when tighter |
| `prerequisites` | no | string[] | skills to load before this one |
| `next` | no | string[] | suggested follow-up skills |
| `completes_with` | no | string | **Reserved (deferred mechanism).** Command path that will mark this skill complete when usage tracking ships (see `04-progressive-learning.md`). Authors may set it now; v1 ignores it |

`commands` is the drift guard: a skill that references a command path or flag
the target does not have fails `cli skills check`, so the corpus cannot
silently describe a CLI that no longer exists.

## Body conventions

The body is Markdown with a fixed set of sections, in this order:

1. `# <Title>` — human-readable title.
2. `## When to use` — the situations that call for this workflow. Agents use
   this to decide whether to load the skill at all.
3. `## Steps` — numbered, ordered, each step one command in a fenced block.
4. `## Examples` — two or three complete invocations.
5. `## Edge cases` — the failure modes and how to recover.
6. `## See also` — links to prerequisite and next skills by name.

Rules:

- Steps are numbered and imperative. Agents follow ordered steps reliably;
  prose paragraphs are where they drift.
- Every command appears in a fenced code block so it can be copied verbatim.
- Keep `SKILL.md` under 200 lines and 20 KB. If a topic needs more room,
  move it to `references/` and link to it from the body.
- No absolute paths. No instructions that require network access.

## Template

`skills/_template/SKILL.md`:

```markdown
---
name: my-skill
description: One line: what this workflow accomplishes and when to use it.
level: beginner
risk: mutating
tags: []
triggers: []
commands: []
requires: []
# min_version: "2.40"        # optional: overrides the manifest's target_version
prerequisites: []
next: []
# completes_with: pr create   # reserved: completion marker for deferred tracking (doc 04)
---

# My skill

## When to use

Describe the situation that calls for this workflow.

## Steps

1. Run `cli ...` to ...
2. Run `cli ...` to ...

## Examples

```console
$ cli ...
```

## Edge cases

- If ... then ...

## See also

- `other-skill`
```

## Example

`skills/commit-flow/SKILL.md`:

```markdown
---
name: commit-flow
description: Stage, review, and land one atomic change
level: beginner
tags: [vcs, commit]
commands: [status, diff, add, commit]
prerequisites: []
next: [pr-flow]
---

# Commit flow

## When to use

You have uncommitted changes and want to land exactly one logical change as
one commit.

## Steps

1. Run `cli status` to see what changed.
2. Run `cli diff` to review the unstaged changes.
3. Run `cli add <files>` to stage only the files that belong to this change.
4. Run `cli diff --staged` to review what will be committed.
5. Run `cli commit -m "<message>"` with a message that states the change and
   why.

## Examples

```console
$ cli status
$ cli add src/parser.ts
$ cli diff --staged
$ cli commit -m "fix: handle empty input in parser"
```

## Edge cases

- If `cli diff --staged` shows unrelated changes, unstage them with
  `cli reset <files>` and start over.
- If the commit fails because of a hook, read the hook output and fix the
  cause; do not bypass the hook.

## See also

- `branch-flow` — when the change needs its own branch first
- `pr-flow` — once the commit is on a branch and ready for review
```

For a complete skill written against a real target (gh), see
`examples/gh/pr-flow/SKILL.md`. For how to write skills agents actually
trigger on and follow, see `08-skill-authoring.md`.

## Validation (`cli skills check`)

Run in CI and before release. Fails on any violation:

1. `name` matches the directory name and is unique across the corpus.
2. `description` is present and ≤ 200 characters.
3. `level` is one of the three enum values.
4. `risk` is one of the three enum values.
5. Every entry in `commands` resolves to a real command path in the target's
   command tree.
6. Every flag used in the body's fenced commands exists on that command,
   validated against the target's help output at the pinned target version
   (native: the binary's own tree; shim: a per-target help adapter, see
   `06-shim-pattern.md`).
7. Every entry in `prerequisites` and `next` names an existing skill.
8. Every relative path referenced in the body resolves inside the skill
   directory.
9. `SKILL.md` is ≤ 200 lines and ≤ 20 KB; the rendered catalog for the
   target is ≤ 8 KB (~2k tokens).
10. `min_version`, when present, parses as a version.
11. Frontmatter parses as YAML.
