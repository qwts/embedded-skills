---
name: embedded-skills
description: Work on this repository's embedded CLI skills and PATH-shadow shims. Use when adding or editing a target skill, adding a CLI target, building or installing a skills shim locally, validating a corpus, or debugging why an agent does not see or use the shim.
---

# Embedded Skills

This repository's one entry-point contributor skill. It routes three
different jobs to three reference procedures. Load exactly the reference
for the requested job; do not load all three by default.

## Route the request

| Request | Load | Outcome |
|---|---|---|
| Add or edit a workflow skill for a CLI that already has an example/target corpus | `references/add-skill.md` | A validated `SKILL.md` added to that target's corpus, with its chain links and manifest updated. |
| Add skills for a CLI that has no shim/corpus yet, or generalize the example into another target | `references/add-target.md` | A runnable PATH-shadow shim for the new CLI, using the gh example as the current reference implementation. |
| Build, install, verify, troubleshoot, or remove an already-built shim on this machine | `references/install-shim.md` | A reversible local installation in front of the real CLI, verified end-to-end. |

If a request spans jobs, do them in this order: add-target, then
add-skill, then install-shim. Finish and verify each procedure before
starting the next.

## Repository facts (do not rediscover)

All repository paths in this skill and its references are relative to the
repository root, not this skill directory, unless a step explicitly says
otherwise. Start from the repository root.

- Design source of truth: `docs/`. When this skill and the docs disagree,
  stop and report the disagreement; do not choose silently.
- Delivery mechanism is always a compiled PATH-shadow binary named
  exactly `<target>`. Never use a target's plugin/extension system
  (gh extensions, `git-<command>` dispatch, kubectl plugins, etc.).
- No script shims for delivered targets. Build-time helper scripts are
  allowed only where the existing `examples/gh` pattern already uses them.
- Current reference implementation: `examples/gh/`. The shared shim module
  is `shim/` (one codebase, many targets, docs/06); the first target built
  on it is `targets/agent-bot/`. `examples/gh/` keeps its own copy of the
  shim code until it is migrated onto `shim/`.
- A skill embedded for a target agent (for example under
  `examples/gh/skills/`) is **content in this repo's product format**.
  This root skill is a **contributor skill for working on the repo**.
  Keep the two uses distinct.

## Always verify against the installed target

Target commands, flags, JSON fields, versions, and help text are facts to
observe from the currently installed binary, never recollections. Before
writing or changing commands in a skill, run at least:

```console
<target> --version
<target> <command> --help
```

When testing through a locally installed shim, distinguish the shim from
the real binary explicitly (`command -v <target>`, the shim's `doctor`
mode where available, or the documented real-binary override).

## Safety rails

- Do not change the target CLI itself, its configuration, credentials, or
  global state unless the user explicitly asks.
- Do not place a shim directory on a user's persistent shell PATH or edit
  shell profiles. Print the exact line for them to add; a process-scoped
  PATH change for the current test shell is fine.
- Do not claim an installation works from build success alone. Exercise
  the public surface: catalog, one skill load, one non-mutating real
  command through the proxy, and version/proxy behavior.
- Publishing to this GitHub repository follows the repository's active
  branch ruleset. If the user asks to publish, use the verified guidance
  loaded for that operation at the time; do not substitute an ad-hoc push.
