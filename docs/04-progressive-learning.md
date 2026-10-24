# Progressive Learning

The goal: an agent gets better at the CLI as it uses it — not because it was
retrained, but because the CLI teaches it the next thing at the right time.

**Scope split.** Disclosure layers, levels, and chains are v1: they are
static curation in frontmatter and cost the runtime nothing. Usage
tracking, `skills next`, and `skills feedback` are **deferred** (bottom of
this doc): they add local state and writes, and they ship only after
evals (doc 09) show agents actually discover and complete v1 skills.
Nothing in v1 records anything.

## Three disclosure layers

Agents are trained to consume documentation in layers: see the map, open the
one page, follow the one link. The corpus mirrors that:

| Layer | Surface | Cost | Contents |
|-------|---------|------|----------|
| 1 | `cli skills` | one screen | name, level, one-line description |
| 2 | `cli skills <name>` | one document | full workflow (frontmatter + body) |
| 3 | `references/` files | on demand | deep dives the body links to |

The agent pays for layer 2 only when layer 1 convinces it the skill is
relevant, and for layer 3 only when the body says the depth is needed. Context
budget is respected at every step.

Layer 1 has a second consumer: the harness. At session start it runs
`cli skills --json` and injects the list into the agent's context, so the
curriculum is visible before the agent runs anything (see
`01-cli-surface.md`, "Session bootstrap").

## Levels

`level` gates complexity:

- `beginner` — single-purpose workflows, one command or a short chain.
- `intermediate` — multi-step workflows that combine commands.
- `advanced` — workflows with judgment calls: releases, migrations, recovery.

`cli skills --level` lets an agent (or a harness configured for a skill level)
scope discovery.

## Chains: prerequisites and next

Frontmatter links skills into a graph:

- `prerequisites` — load these first. `pr-flow` lists `commit-flow` and
  `branch-flow`.
- `next` — after this one, try these.

In v1 these links are advisory text an agent (or harness) reads and
follows — exactly like "See also" in a man page. They gate nothing and
record nothing. Their only runtime use is ordering: the catalog sorts by
level, and `skills <name> --json` surfaces `prerequisites`/`next` so a
harness can pre-load or suggest them. *Unlocking* based on observed use
is the deferred mechanism below.

## Deferred: usage tracking, `next`, and feedback

Everything from here to "The growth loop" is a design for a later phase,
kept because it shapes the `completes_with` field in doc 02. The gate for
building any of it: v1 evals showing agents discover skills unprompted
and beat `--help`-only baselines. Until then it is unbuilt complexity on
top of an unproven behavior.

### Usage tracking (deferred)

Opt-in, local, and plain:

- Config flag: `skills.track_usage = true` (default off).
- Storage: one JSON file in the user's config directory, e.g.
  `~/.config/cli/skills-usage.json`:

```json
{
  "loaded": {
    "commit-flow": {"count": 7, "completed": 5, "last_used": "2026-10-23T17:00:00Z"},
    "branch-flow": {"count": 3, "completed": 1, "last_used": "2026-10-23T17:05:00Z"}
  },
  "failures": [
    {"skill": "pr-flow", "step": 3, "note": "gh pr create rejected the body", "at": "2026-10-23T17:10:00Z"}
  ]
}
```

- No network, ever. The file exists to rank `cli skills next` and to hold
  failure reports, nothing else.
- `cli skills next --reset` clears it. Documented in `cli skills --help`.
- Writes are atomic (write-temp-then-rename) so concurrent agents cannot
  corrupt the file.
- The file is per machine, not per agent: multiple agents on one host share a
  history. Treat recommendations as host-level, not agent-level.

Recording happens when a skill is loaded via `--skill` or shown via
`cli skills <name>` — both are explicit acts of learning.

### Signal strength: loaded is not mastered (deferred)

Tracking records that a skill was *loaded*, which is a weak proxy for
*mastered*. `cli skills next` ranks by loads, so a skill an agent opened once
and abandoned still counts as progress.

The completion signal: a skill may declare `completes_with: <command path>`
in frontmatter (see `02-skill-format.md`). When the CLI executes that command
while the skill is loaded, the load is recorded as `completed`, and `next`
ranks completed prerequisites above merely loaded ones. Skills without
`completes_with` fall back to load counts.

### Feedback: failures flow back (deferred)

The loop is not one-directional. When a skill fails an agent — a step errors,
a command is rejected — the agent records it:

```text
cli skills feedback <name> --step <n> [--note <text>]
cli skills feedback --list
```

Failures are stored in the usage file (see above) and listed locally.
Submission is a human action: the `skills` command never touches the
network. Maintainers use failure reports to fix or retire skills, which is
how the corpus learns from the agents using it.

## The growth loop

The v1 loop is stateless — the progression lives in the static chains and
in the harness's own context management, not in CLI-side state:

```text
bootstrap  cli skills --json     harness injects the map into context (layer 1)
    │
discover   cli skills            see the map when needed
    │
load       cli skills <name>     pull the workflow (layer 2)
    ▼
apply      run the steps         the agent does the work
    │
follow     `next` links          the body names what to learn next
    │
    └──────────────────────────► back to discover
```

An agent that has run `commit-flow` and `branch-flow` finds `pr-flow`
through their `next` links. That is v1 progressive learning: a curated
curriculum over the discover, load, act loop, with no recording. The
track/report/suggest stages above it are the deferred phase, gated on
evals.

## Why this shape

Each piece maps to a habit agents are trained on:

- Enumerate before acting → `cli skills` as the first call.
- Read the doc for the one tool you need → layer 2 on demand.
- Follow the "see also" link → layer 3 and the `next` chain.
- *(Deferred)* Learn from what you just did → usage tracking feeding
  suggestions, once discovery itself is proven.

The CLI is not retraining the model; it is feeding the existing
discover, load, act loop with better material at each step.
