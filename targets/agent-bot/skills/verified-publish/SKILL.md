---
name: verified-publish
description: Publish a branch as the bot with GitHub-verified signatures: signed-commit preconditions, dry-run, publish, recovery, and on-GitHub confirmation. Use when a branch is ready to publish as the bot
level: advanced
risk: destructive
tags: [agent-bot, publish, signed-commit, github, verified, bot]
triggers: [publish, push, signed commit, verified commit, publish branch, ship it, signed-commit, verified publish, bot commit]
commands: [signed-commit, doctor, mint-token]
requires:
  - "agent-bot doctor --json reports the github-identity feature gate enabled and an App resolved for this checkout"
  - "a bot identity resolves here: --app, GH_AGENT_APP, the checkout's agentBot.app pin, or the harness account"
  - "the worktree is clean (git status --porcelain is empty)"
min_version: "0.12.0"
prerequisites: [join-and-bind]
next: []
completes_with: signed-commit
---

# Verified publish

## When to use

Your branch has commits ready to publish as the bot, and the remote branch
must carry GitHub-verified signatures. `agent-bot signed-commit` replays
each commit through GitHub's Git Data API, verifies the returned signature
and tree, then updates the remote branch under a force-with-lease
equivalent. This skill is the safe ordering around that command:
preconditions, dry-run, publish, recovery, and the on-GitHub confirmation.

## Steps

1. Confirm the identity gate and the App. Run
   `agent-bot doctor --json --require-schema-version 1`. Expected: the
   `github-identity` feature gate is enabled (the report carries it under
   the machine checks' `evidence.gates`), this checkout resolves an App,
   and that App's credential is valid with a live mint succeeding.
   `signed-commit` refuses with `github-identity add-on is off` when the
   gate is off — enable `features.github-identity` in the user config and
   re-check; do not work around the gate. If no App resolves, stop here and
   finish `join-and-bind` first.
2. State the bot identity and prove it mints as the bot. From the bound,
   pinned worktree, `agent-bot mint-token` mints through the daemon.
   Verify the login end to end:

   ```console
   GH_TOKEN="$(agent-bot mint-token)" || exit 1
   export GH_TOKEN
   login="$(gh api graphql -f 'query={viewer{login}}' --jq .data.viewer.login)" || { unset GH_TOKEN; exit 1; }
   unset GH_TOKEN
   [ "$login" = "<slug>[bot]" ] || exit 1
   ```

   Expected: the login is `<slug>[bot]`. Never echo, persist, or include the
   token in output; treat mint output as secret-bearing. Do not combine the
   command substitution with `export` — the shell then returns the status of
   `export`, masking a failed mint. Never run a bare
   `agent-bot mint-token --app <slug>` from an unpinned context with no
   stated identity: it fires the owner-approval prompt on every call. If no
   identity resolves, `signed-commit` refuses with `no bot identity resolves
   here — pass --app, set GH_AGENT_APP, pin the checkout, or run from the
   harness account`.
3. Confirm the worktree is clean and the range is linear and current:
   `git status --porcelain` must be empty, `git fetch origin` first if
   another writer may have updated the branch, and
   `git log --merges --oneline origin/<default>..HEAD` must be empty.
   `signed-commit` refuses a dirty tree, a merge commit in the range, and a
   base that is not an ancestor of HEAD — fix these before publishing, not
   after.
4. Dry-run: `agent-bot signed-commit --dry-run`. Expected: it prints
   `repo`, `branch`, `base`, `remote (not checked — dry run)`, `commits`,
   and one line per commit with its changed paths, ending with `dry run —
   nothing was created or pushed`. A dry run mints nothing and contacts no
   network. If it fails, read the error, fix the cause (see Edge cases),
   and dry-run again.
5. Publish: `agent-bot signed-commit`. **This is the irreversible step.**
   It replays every commit in the range through GitHub, verifies each
   returned signature and tree, then force-updates the remote branch under
   a lease. Expected: one `-> <sha> verified (<reason>)` line per commit,
   then `<branch> -> <sha>` and `local branch reset to the signed history`.
   The remote update is necessarily forced — replayed commits have new
   object IDs — but the lease prevents overwriting a remote head this
   checkout did not observe. Never replace this with a blind
   `git push --force`.
6. Verify on GitHub: open the commit or pull request page and confirm it
   shows the App author (`<slug>[bot]`) and GitHub's `Verified` status.
   This is the definition of done — do not report success without it.

## Examples

```console
$ agent-bot doctor --json --require-schema-version 1
$ GH_TOKEN="$(agent-bot mint-token)" || exit 1
$ agent-bot signed-commit --dry-run
$ agent-bot signed-commit
```

## Edge cases

- `working tree is dirty — commit or stash first, so the signed commit
  matches what you tested`: commit or stash, then re-run.
- `github-identity add-on is off`: enable the feature gate in the user
  config; `doctor` shows each gate. Off is not an error to work around.
- `no bot identity resolves here — pass --app, set GH_AGENT_APP, pin the
  checkout, or run from the harness account`: state the identity explicitly;
  see `join-and-bind`. Never retry as the human account.
- `origin is not a repository on <host> — pass --repo owner/name`: pass
  `--repo <owner/name>`. Use an HTTPS origin; SSH commonly bypasses the App
  identity.
- `detached HEAD — pass --branch <name>`: pass `--branch`.
- `<base> is not an ancestor of HEAD … Rebase first (git rebase <base>)`:
  the worktree was built on an older commit, and publishing would revert
  what the base gained since. Rebase onto it, then re-run.
- `could not find a merge-base with <default> — pass --base <ref>`: pass
  `--base <ref>`.
- `nothing to sign — HEAD is already at the base`: there is nothing to
  publish.
- `<sha> is a merge commit — rebase into linear history first`: rebase the
  range into linear history.
- `<path> is a submodule — submodule bumps are not supported`: publish
  without the submodule bump.
- `<branch> on the remote is at <sha>, but this checkout last saw <sha> —
  someone else pushed. Nothing was written.`: run the printed recovery
  (`git fetch origin <branch> && git rebase origin/<branch>`), revalidate
  the new range, and start again.
- `<branch> moved on the remote during replay — nothing was pushed`: fetch,
  rebase, and re-run.
- `<sha> came back unsigned (<reason>)` or `tree mismatch on <sha>`: stop.
  Do not retry as the human account; report the failure.
- `PUSH SUCCEEDED — <branch> is at <sha>; local reset failed: <msg>`: the
  remote is already published. Follow the printed fetch/reset recovery
  (`git fetch origin <branch> && git reset --hard <sha>`); do not republish
  the branch.
- `refusing to rewrite the default branch (<name>)`: the default branch is
  protected. Pass `--allow-default-branch` only when the user explicitly
  requests a default-branch rewrite and repository policy allows it.
- A commit message that credits another soul is refused (`<sha>: … Nothing
  was written.`). The message is replayed verbatim, so the soul's co-author
  trailer survives — never write another soul's trailer.
- Identity resolution order is fixed: explicit App, `GH_AGENT_APP`, the
  checkout's `agentBot.app` pin, the managed soul assignment, the configured
  account, then the harness mapping. Present but unreadable or conflicting
  pins are errors; directory names never select identity.
- Stop on mint, credential, identity-pin, lease, verification, or tree-match
  failures. Never retry as the human account.

## See also

- `join-and-bind` — the prerequisite: a bound, set-up worktree with a
  resolved App
- `agent-bot skill agent-bot --for signed-commit` — the bundled reference
  with the full safety-check list
- `agent-bot skill agent-bot` — the bundled router skill for this CLI
