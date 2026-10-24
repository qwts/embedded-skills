---
name: safe-history
description: Commit, branch, and rebase safely in a git worktree: preconditions, expected observations, rollback, and no hooks bypassed or published history rewritten.
level: intermediate
risk: destructive
tags: [git, history, rebase, commit, branch]
triggers: [rebase, commit, branch, amend, force push, conflict, rollback, squash]
commands: [status, diff, log, show, commit, rebase, switch, checkout, branch, fetch, reset]
prerequisites: [soul-worktree-flow]
next: []
---

# Safe history

## When to use

You are inside a linked worktree (see `soul-worktree-flow`) and need to commit, move a branch, rebase, or recover from a conflict without rewriting published history or bypassing identity and hooks.

## Steps

1. Confirm you are not on the primary checkout's default branch: `git rev-parse --abbrev-ref HEAD`. Expect a feature branch name, not `main` or `master`. If it is the default branch, stop and create a feature branch first.
2. Check the tree: `git status --porcelain=v1` — expect no output. If there is output, review it with `git status` and either commit or stash before rebasing.
3. Fetch latest upstream state: `git fetch origin`. Expect success; on failure, stop and report the remote error.
4. Review what will change before staging: `git diff`. Expect only intended edits. If unexpected hunks appear, stop and investigate before staging.
5. Stage only the files for this change: `git add <path>...`. Then review the staged set: `git diff --cached`. Expect the staged diff to match the intended change exactly.
6. Commit with a message stating the change and why: `git commit -m "<message>"`. Expect a new commit hash. If a hook fails, read its output and fix the cause; do not use `--no-verify`.
7. Verify the commit: `git show --stat HEAD` and `git status --porcelain=v1`. Expect the commit present and a clean tree.
8. To move the branch onto updated upstream, run `git rebase origin/<base>`. Expect either `Successfully rebased` or a conflict list.
9. If a conflict occurs, resolve the files, stage them with `git add <path>...`, then `git rebase --continue`. To abandon the rebase entirely, run `git rebase --abort` (this returns the branch to its pre-rebase state).
10. To create a new branch, prefer `git switch -c <branch>`. To delete a merged branch, use `git branch -d <branch>`; use `-D` only for a branch you are deliberately discarding.
11. Never rewrite published history: do not `git commit --amend` or `git rebase` commits that have already been pushed, and never run `git push --force` on a shared branch. If a push is rejected, use `git push --force-with-lease` only after confirming no one else advanced the branch.
12. For recovery from a bad local state, `git reset --hard <known-good-commit>` is destructive and discards uncommitted work; prefer `git reset --keep` or `git stash` when you need to preserve edits. Mark the irreversible step before running it.

## Examples

```console
$ git status --porcelain=v1
$ git diff
$ git add src/parser.go
$ git diff --cached
$ git commit -m "fix: handle empty input in parser"
$ git rebase origin/main
```

## Edge cases

- If `git commit` fails because of a hook, fix the reported cause; bypassing the hook with `--no-verify` is not allowed.
- If `git rebase` reports `CONFLICT`, do not guess: open the conflicted files, resolve, then `git rebase --continue`; use `git rebase --abort` to return to the starting state.
- If `git push` is rejected as non-fast-forward, fetch and rebase (`git fetch origin` then `git rebase origin/<base>`) instead of forcing.
- `git reset --hard` cannot be undone for uncommitted changes; prefer `git stash` when in doubt.

## See also

- `soul-worktree-flow` — obtaining the worktree this history work happens in
