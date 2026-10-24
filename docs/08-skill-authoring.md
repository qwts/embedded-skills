# Authoring Skills

A CLI skill is not documentation squeezed smaller. It is a **procedure an
agent executes**: preconditions to verify, commands in order, decision points,
recovery paths. Write for a competent operator who has never seen this CLI's
conventions and will follow you literally.

## The field that decides everything: `description`

Triggering happens on `description` (and `triggers`). The agent sees the
catalog, matches the current task against descriptions, loads at most a few.
Rules:

- State **what** and **when**: "Open a pull request with a complete title,
  body, and reviewers. Use when a branch has commits ready for review."
- Name the user's vocabulary, not the CLI's: "pull request", "code review",
  "squash" — that is what `triggers` is for.
- One skill, one job. "Helps with branching" triggers on everything and helps
  with nothing. ≤ 200 chars, enforced by the validator.

## Tier discipline

| Layer | Content rules |
|-------|---------------|
| 1 — catalog line | `name` + `description` only. Generated; never hand-written. |
| 2 — `SKILL.md` body | Numbered steps. Each step: exact command, expected observation, branch if different. End with a verification step and a rollback pointer. |
| 3 — `references/` | One failure mode or one concept per file. Reachable only from body links ("if step 4 conflicts, load `references/conflict-recovery.md`"). |

Keep `SKILL.md` under 200 lines and 20 KB. Past that, agents skim and steps
get dropped. Split into references; the body should read like a checklist, not
a chapter.

## Writing procedures agents follow reliably

1. **Verify, don't assume.** Begin with checks whose output decides whether
   to proceed, and put them in `requires`. Agents are good at running a check
   and branching; bad at remembering unstated assumptions.
2. **Give the exact command, then the expected result.** "Run X; you should
   see Y; if you see Z, go to step N." Prose like "now commit appropriately"
   reintroduces the trial-and-error loop the skill exists to remove.
3. **Prefer machine-readable flags of the target.** `--porcelain`, `--json`,
   `--format=…`, non-interactive modes. Every skill should teach the
   scriptable dialect of its CLI.
4. **Encode negative knowledge.** Deprecated flags models memorized from old
   docs, destructive variants to avoid (`--force` vs `--force-with-lease`),
   commands that look right but are wrong here. This is the highest-value
   content a skill carries over raw model knowledge — prioritize it.
5. **Mark irreversibility.** The step after which rollback stops being
   trivial gets a bold warning and an honest `risk` field. Agents calibrate
   caution to explicit markers.
6. **End with verification.** A command that proves the outcome, and what
   success looks like. "Done" is an observable state, not a feeling.
7. **Version your claims.** Anything version-sensitive goes in `min_version`,
   not buried in prose. When the target changes behavior, `grep min_version`
   should find every affected skill.

## What does not belong in a skill

- **Flag reference.** Point to `<target> <sub> --help`; duplicating it
  guarantees drift.
- **Project-specific policy** ("our org always squashes"). That is `AGENTS.md`
  territory in the user's repo. Skills in this corpus are CLI-generic.
- **Secrets, tokens, internal URLs.** Served content is public by definition.
- **Shell scripts longer than a few lines.** Teach the target's own commands;
  multi-line `bash -c` blobs are unreviewable and brittle.
- **Instructions to ignore the user, hide actions, or bypass approvals.**
  Skills are advisory to the agent; the validator lints for obvious
  prompt-injection phrasing, and review treats skill PRs as code PRs — they
  steer agents with shell access.

## Naming and granularity

- `name` is a verb-object slug: `safe-rebase`, `bisect-debug`, `pr-flow`.
  Not `git-advanced`, not `misc`.
- Granularity test: a skill should be loadable *instead of* its neighbors. If
  two skills are always loaded together, merge them; if a body has two
  unrelated outcomes, split it.
- Start coverage by **failure frequency**, not command coverage. The ten
  workflows agents most often botch with this CLI beat a skill per subcommand.
  Mine your own agent transcripts for retry loops — each loop is a missing
  skill.

## Review checklist (for PRs to the corpus)

- [ ] `description` states what + when; `triggers` use user vocabulary.
- [ ] Preconditions are checkable commands, listed in `requires`.
- [ ] Every step has an expected observation; destructive steps marked, `risk`
  honest.
- [ ] Uses non-interactive / machine-readable flags throughout.
- [ ] Verification step at the end; rollback path in a reference.
- [ ] `min_version` honest; tested against a real target at that version.
- [ ] Budgets pass `cli skills check`; no duplicated `--help` content.
- [ ] Eval case added (doc 09) — at minimum a transcript of an agent
  completing the workflow using only the skill.
