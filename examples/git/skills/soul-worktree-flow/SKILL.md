---
name: soul-worktree-flow
description: Check in, create or reuse a linked git worktree under <soulDir>/worktrees/<name>, work there, and clean up; never branch or edit in the owner's primary checkout.
level: beginner
risk: mutating
tags: [git, worktree, soul, branch]
triggers: [worktree, soul, linked worktree, primary checkout, branch, checkout, create worktree]
commands: [status, worktree, branch, config, rev-parse, fetch]
prerequisites: []
next: [safe-history]
---

# Soul worktree flow

## When to use

You need to make a change in a repository whose owner's primary checkout must stay untouched. Use this to obtain an isolated linked worktree under your soul directory, do the work there, and remove it when done.

## Steps

1. Verify a clean starting tree: `git status --porcelain=v1` — expect no output. If there is output, stop and either commit or stash the existing changes first.
2. Check whether you are already in a bot-configured worktree: `git config --get agentBot.agentId` — expect a soul id. If it prints nothing, you are in the owner's primary checkout (or an unconfigured clone): do not create branches or edit files there. Record `<soulDir>` from your agent identity and continue at step 4.
3. Inspect existing worktrees: `git worktree list --porcelain`. If `<soulDir>/worktrees/<name>` is already listed and its directory exists, `cd` into it, run `git status --porcelain=v1`, and skip to step 6.
4. Fetch the base before branching: `git fetch origin <base>` (or `git fetch origin` if the base is already tracked). Expect the fetch to succeed; on failure, stop and report the remote error rather than branching from a stale ref.
5. Create the linked worktree and branch: `git worktree add <soulDir>/worktrees/<name> -b <branch> <base>`. Expect `Preparing worktree` and the new path. If the path already exists but is not registered, remove the stale directory first or choose a new `<name>`; do not use `--force`.
6. Enter the worktree and confirm identity/config: `git config --get user.name` and `git config --get agentBot.agentId`. Expect the bot name and a soul id. If identity is missing, set it with `git config --worktree user.name "<bot>"`, `git config --worktree user.email "<bot-email>"`, and `git config --worktree commit.gpgsign false` before committing.
7. Do the work on `<branch>` only inside this worktree. Never `cd` back to the primary checkout to edit or branch.
8. Before finishing, verify: `git status --porcelain=v1` (clean after commit) and `git log --oneline -1`.
9. Clean up when the change is merged or abandoned: `git worktree remove <soulDir>/worktrees/<name>`, then `git branch -d <branch>` (use `-D` only if the branch was deliberately discarded). Run `git worktree prune` to drop stale metadata.
10. If removal fails because of untracked or modified files, run `git status --porcelain=v1` inside the worktree and resolve; do not pass `--force` to `git worktree remove` unless you have confirmed the files are disposable.

## Examples

```console
$ git status --porcelain=v1
$ git config --get agentBot.agentId
agent_4364298d-...
$ git fetch origin main
$ git worktree add <soulDir>/worktrees/fix-thing -b fix/thing origin/main
$ git worktree remove <soulDir>/worktrees/fix-thing
```

## Edge cases

- If `git worktree add` reports `fatal: '<branch>' is already checked out`, another worktree owns that branch. Choose a different branch name or remove the other worktree first.
- If the base ref is not present locally, fetch it explicitly: `git fetch origin <base>:refs/remotes/origin/<base>`.
- Linked worktrees share the object database but not the working tree; uncommitted changes in one worktree do not appear in another.
- Do not delete the primary checkout's branch while a linked worktree is using it.

## See also

- `safe-history` — commit, rebase, and branch discipline inside the worktree
