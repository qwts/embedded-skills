---
name: verified-publish
description: Publish bot-authored commits with GitHub-verified signatures. Use when you need to push signed commits as a bot identity to a repository
level: advanced
risk: destructive
tags: [git, publish, sign, bot, identity, commit]
triggers: [signed commit, verified commit, publish as bot, signed-commit, push as agent, verified push]
commands: [signed-commit]
requires:
  - "the live session binding is established (MCP bind exchange completed, or agent-binding.json present in the git dir)"
  - "agent-bot doctor shows a resolved bot identity for this worktree"
  - "the worktree is clean (git status --porcelain is empty)"
  - "the commit range is linear (no merge commits between base and HEAD)"
min_version: "0.12.0"
prerequisites: []
next: []
---

# Verified publish

## When to use

You are a bot identity and need to push commits to a remote repository with
GitHub-verified signatures. The `agent-bot signed-commit` command replays each
local commit through GitHub's Git Data API, verifies the returned signature and
tree, then updates the remote branch under a force-with-lease equivalent.

Do not use this for human-authored commits or when you are not operating under
a bot identity.

## Steps

1. Verify the live session binding is established. A worktree can resolve a
   bot identity while the session is unbound — publishing then fails closed
   mid-flow. Check that `<git-dir>/agent-binding.json` exists:
   `test -f "$(git rev-parse --git-dir)/agent-binding.json"`.
   Expected: exit 0. If the file is missing, complete the MCP `bind` exchange
   (see `join-and-bind` step 4) before proceeding. Do not attempt to publish
   from an unbound session.

2. Verify your bot identity is resolved. Run `agent-bot doctor --json` and
   confirm `identity.class` is not `human` and an App slug is resolved.
   Expected: the JSON report shows a bot identity. If it shows `human` or no
   App, stop — you cannot publish as a bot from this checkout.

3. Confirm the worktree is clean. Run `git status --porcelain`. Expected: empty
   output. If there are uncommitted changes, commit or stash them first.

4. Confirm the commit range is linear. Run `git log --merges <base>..HEAD` where
   `<base>` is your target branch (usually `origin/main` or `origin/master`).
   Expected: no output. If merge commits appear, rebase them away first:
   `git rebase <base>`.

5. Fetch the remote branch to ensure you have the latest state:
   `git fetch origin <branch>`.

6. Preview the publish without minting credentials or contacting GitHub:
   `agent-bot signed-commit --dry-run`. Expected: a preview of the commits
   that would be replayed. If the dry-run fails, read the error and fix the
   cause before proceeding.

7. Publish the current branch:
   `agent-bot signed-commit`. Expected: the command replays each commit,
   verifies signatures, and updates the remote branch. The remote update is
   necessarily forced because replayed commits have new object IDs; the lease
   prevents overwriting a remote head this checkout did not observe.

8. Verify the result on GitHub. Open the commit or PR page and confirm:
   - The author shows the bot App identity (not a human).
   - GitHub displays the "Verified" badge on the commits.
   This step is the definition of done — do not report success without it.

## Examples

```console
$ agent-bot doctor --json
$ git status --porcelain
$ git fetch origin main
$ agent-bot signed-commit --dry-run
$ agent-bot signed-commit
```

With explicit overrides when repository discovery is insufficient:

```console
$ agent-bot signed-commit --base origin/main --branch my-feature --repo owner/repo
```

## Edge cases

- **Session not bound:** `signed-commit` fails closed mid-flow if the live
  session binding is missing. Complete the MCP `bind` exchange first (see
  `join-and-bind` step 4), then retry.
- **Dirty worktree:** `agent-bot signed-commit` refuses to run. Commit or stash
  your changes first.
- **Non-linear range:** If the range contains merge commits, rebase them away
  before publishing. The command will refuse.
- **Base not an ancestor of HEAD:** The worktree was built on an older commit.
  Rebase onto the base first: `git rebase <base>`.
- **Lease failure (remote moved):** The remote branch was updated by someone
  else while you were working. Run the recovery command printed by the CLI,
  revalidate the new range, and start again from step 6.
- **Local reset failure after remote update:** The remote update succeeded but
  the local reset failed. Follow the printed fetch/reset recovery instructions;
  do not republish the branch.
- **Default branch protection:** The default branch is protected. Do not pass
  `--allow-default-branch` unless the user explicitly requests a default-branch
  rewrite and repository policy allows it.
- **Submodule updates:** If a submodule update appears in the range, the
  command stops. Resolve submodule changes separately.
- **App identity cannot be resolved:** The command refuses to mint. Run
  `agent-bot doctor` to diagnose and fix the identity before retrying.

## See also

- `agent-bot skill agent-bot --for signed-commit` — the bundled reference for
  this workflow, with additional detail on safety checks and recovery
- `join-and-bind` — if you need to set up a bot identity first
