---
name: pr-merge
description: Land an approved, green pull request with the right merge method. Use when a PR is approved and ready to merge
level: advanced
risk: destructive
tags: [vcs, pr, merge]
triggers: [merge this, land it, merge the pr, ship it]
commands: [pr view, pr checks, pr merge]
requires:
  - "gh pr view <n> --json reviewDecision shows APPROVED (or the user explicitly instructs otherwise)"
  - "gh pr checks <n> shows all required checks passing"
min_version: "2.40"
prerequisites: [pr-flow, review-flow]
next: []
completes_with: pr merge
---

# PR merge

## When to use

A pull request is approved and ready to land. Merging is the one step in
the PR lifecycle you cannot quietly undo — a revert is a new, visible
change — so every precondition is checked, not assumed.

## Steps

1. Establish merge readiness from the server's view, not memory:
   `gh pr view <n> --json state,reviewDecision,mergeable,mergeStateStatus,title`.
   Expected: `state` is `OPEN`, `reviewDecision` is `APPROVED`,
   `mergeStateStatus` is `CLEAN`. Anything else: stop and report which
   field blocks.
2. Confirm checks: `gh pr checks <n> --required`. Expected: all required
   checks pass. Never merge over red required checks unless the user
   explicitly instructs it — and then record that instruction in your
   report. Never use `--admin` on your own initiative; it exists to
   bypass the very protections this skill checks.
3. Choose the merge method from the repo's observable convention, not
   preference: check recent history (`git log --oneline -10 --merges`
   and whether merged PRs preserved their commits). If the convention is
   unclear or the user has not named a method, ask — squash vs merge
   vs rebase is the owner's call, not yours.
4. Merge: `gh pr merge <n> --<method>` (one of `--squash`, `--merge`,
   `--rebase`). Add `--delete-branch` only when the user or repo
   convention says branches are deleted after merge.
   **This is the irreversible step.**
5. Verify: run `gh pr view <n> --json state,mergedAt,mergeCommit` and
   confirm `state` is `MERGED`. Report the merge commit and method to
   the user.

## Examples

```console
$ gh pr view 123 --json state,reviewDecision,mergeable,mergeStateStatus,title
$ gh pr checks 123 --required
$ gh pr merge 123 --squash --delete-branch
$ gh pr view 123 --json state,mergedAt,mergeCommit
```

## Edge cases

- `mergeStateStatus` is `BEHIND`: the branch must be updated first
  (`gh pr update-branch <n>`), then re-check from step 1 — reviews can
  be dismissed by the update.
- `mergeStateStatus` is `DIRTY`: there are conflicts. Resolving them is
  author work; do not resolve conflicts inside a merge task.
- Checks pending a long time: `gh pr merge <n> --<method> --auto` queues
  the merge for when requirements are met. Use it only when the user
  wants fire-and-forget landing, and verify with
  `gh pr view <n> --json autoMergeRequest` afterwards.
- The PR was updated after approval: head commit may have changed under
  the approval. `--match-head-commit <sha>` pins the merge to the exact
  commit you verified in step 1.

## See also

- `review-flow` — the approval this skill requires has to mean something
- `pr-flow` — opening the next one
