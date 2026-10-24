# Embedded Skills — Design Package

Status: design only. No implementation in this pass.

Design notes: read-only / clean-stdout discipline, version-probing
semantics, and an explicit no-telemetry default run through the whole
package. `--skill` is an optional convenience, not the canonical path;
usage tracking, `skills next`, and `skills feedback` are designed but
deferred until evals prove discovery (docs 01, 04, 07).

## The idea

A CLI ships with a corpus of embedded skills: structured, machine-readable
workflow guides that live inside the binary and are discoverable at runtime.

The canonical surface is `cli skills`: discovery and loading. List,
inspect, search. The harness loads a skill by fetching it
(`cli skills <name> --json`) and invoking the command separately, so real
command output is never altered.

An optional second surface is `cli --skill <name>` — a convenience that
prints the skill into the invocation's stdout for simple harnesses and
scripts. It is non-canonical: not part of the probing contract, and a
shim may omit it (docs 01, 06).

Delivery is the **shim**, for any executable: a binary named exactly
`<target>`, placed earlier on PATH, answers `skills` itself and execs the
real target for everything else (doc 06). No target cooperation, and in
particular no per-CLI plugin/extension mechanisms (gh extensions,
`git-<command>` dispatch, …) — those don't exist on most CLIs, so they
cannot be the universal path. A CLI building the surface in natively is
the long-term upstreaming endgame (docs 03, 07), not how this repo ships.

The split mirrors how the two audiences consume a CLI:

| Audience | Surface | Shape of content |
|----------|---------|------------------|
| Human | `cli --help` | prose reference, paginated |
| Agent | `cli skills` | structured workflows, JSON-capable |

The intent: an agent working with the CLI runs `cli skills` the way a
human runs `git --help` — to see what is possible before acting. No agent
is trained to do this spontaneously today; making it happen is the
bootstrap problem (doc 07), not an assumption.

## Why agents, specifically

Agents are trained on a discover, load, act loop: enumerate the available
tools, pull the documentation for the one that fits, then execute. `--help`
serves humans in that loop; nothing serves agents. `skills` fills the gap with
content shaped like a workflow (when to use it, ordered steps, examples)
rather than like a reference page (flags, arguments, exit codes).

The progressive-learning property comes from the corpus being organized as
levels and chains: beginner skills link to intermediate ones, and the agent
follows the chain as it works. In v1 this progression is static curation —
nothing is tracked. Ranking recommendations by observed use is designed
but deferred (doc 04). A harness can also inject the layer-1 list into the
agent's context at session start (`cli skills --json`), so the curriculum
is visible before anything runs.

## Questions this package answers

1. How should this repo be structured? → `03-repo-structure.md`
2. Is the idea aligned with what agents are trained on? → `05-alignment-with-training.md`
   (short answer: yes — it reuses conventions agents already know instead of
   inventing new ones)

## Contents

| Doc | Covers |
|-----|--------|
| `01-cli-surface.md` | The `skills` contract (canonical), the optional `--skill` flag, output contracts, exit codes |
| `02-skill-format.md` | Skill document format: frontmatter schema (spec fields + declared extensions), body conventions, template, validation |
| `03-repo-structure.md` | Repo layout, manifest, build-time embedding, versioning, lint and CI |
| `04-progressive-learning.md` | Disclosure layers, levels, chains; deferred usage tracking/`next`/feedback design |
| `05-alignment-with-training.md` | Alignment analysis against agent training patterns, plus cautions |
| `06-shim-pattern.md` | The shim pattern: delivering skills for any executable you don't own (gh, git, kubectl) |
| `07-discovery-and-risks.md` | The bootstrap problem: discovery channels, probing contract, risk register, anti-patterns, phasing |
| `08-skill-authoring.md` | How to write skills agents trigger on and follow: trigger fields, tier discipline, review checklist |
| `09-implementation-and-evals.md` | Reference implementation: proxy fidelity, version probing, validation thresholds, evals, signed releases |
| `../../examples/gh/` | A runnable gh reference example: shim binary source, 4-skill corpus, manifest, PATH-install walkthrough |
| `examples/agents-md-snippet.md` | Paste-able AGENTS.md snippet that bootstraps skill discovery |

## Example

```console
$ cli skills
NAME            LEVEL         DESCRIPTION
branch-flow     beginner      Create, switch, and clean up branches
commit-flow     beginner      Stage, review, and land one atomic change
pr-flow         intermediate  Open a pull request with a complete description

$ cli skills commit-flow
---
name: commit-flow
description: Stage, review, and land one atomic change
---

# Commit flow
...

$ cli commit -m "fix: ..."        # normal invocation; output untouched
...normal command output...
```

The harness fetched `commit-flow` first (`cli skills commit-flow --json`)
and put it in the agent's context; the `commit` invocation itself is
indistinguishable from an un-skilled one. The optional convenience form
`cli --skill commit-flow commit …` prints the skill into that
invocation's stdout instead — same content, dirtier stream (doc 01).
