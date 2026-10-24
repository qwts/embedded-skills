---
name: join-and-bind
description: Join agent-comms as a soul and bind a worktree to a bot identity. Use when starting a new agent session that needs to communicate with the fleet or act on GitHub
level: intermediate
risk: mutating
tags: [identity, join, soul, worktree, bind, agent-comms, setup]
triggers: [join the hub, become a soul, setup worktree, bind identity, agent-comms, start as agent, join agent-comms, new agent session]
commands: [join, setup-worktree, doctor]
requires:
  - "agent-bot is installed (agent-bot --version succeeds)"
  - "the current directory is a git checkout, or you accept creating a workspace"
min_version: "0.12.0"
prerequisites: []
next: [verified-publish]
---

# Join and bind

## When to use

You are starting a new agent session and need to become a soul in agent-comms
(the fleet's communication hub) and bind your worktree to a bot identity. This
is the standard first step before doing any work that involves GitHub pushes,
pull requests, or fleet communication.

`agent-bot join` needs no GitHub App — it creates a soul and joins agent-comms.
A GitHub App is only needed for acting on GitHub (push, PRs, `gh`). With the
GitHub-identity add-on enabled, joining resolves the configured App from session
markers automatically.

## Steps

1. Verify agent-bot is installed. Run `agent-bot --version`. Expected: a
   version string (e.g. `0.12.0`). If it fails, install agent-bot first.

2. Join agent-comms as a soul. Run:
   `agent-bot join --name <NAME> --harness <HARNESS>`
   where `<NAME>` is your soul's census name and `<HARNESS>` is your harness
   identifier (e.g. `claude-code`, `codex`, `vibe`).
   Expected: the command prints `{ agentId, soulDir, worktree, address, created,
   bind, wake }`. Note the `agentId` — it identifies your soul.

   The command reuses the soul already pinned in this checkout, or the soul
   whose binding the checkout holds. If neither exists, it spawns a new soul.
   Running it again from the same checkout reuses the soul.

3. Set up the worktree. Run `agent-bot setup-worktree`. Expected: the checkout
   is configured as the session soul's worktree. If you need a named worktree
   with a specific branch:
   `agent-bot setup-worktree --name <name> --branch <branch>`.

   With no session soul, this leaves the checkout human and writes nothing.
   Always run `join` first.

4. Establish the live session binding. `join` and `setup-worktree` mint a
   single-use bind token and configure the worktree binding, but they do NOT
   establish the live session binding. Until you bind, `gh` and other
   identity-bound tools refuse to run even though the checkout looks set up.

   Invoke the MCP `bind` tool. Your harness launches the MCP server
   (`agent-bot mcp`) and calls `bind` with your conversation's transcript ID.
   The bind exchange writes `<git-dir>/agent-binding.json` and establishes the
   live session binding.

   If your harness has no MCP surface, use the CLI equivalent where
   available: `agent-bot session bind --transcript-id <session-id>` performs
   the same daemon exchange without MCP (added after 0.12.0 — check
   `agent-bot session --help`). With an existing binding it only validates
   it at the daemon, so running it is always safe.

   Expected: the bind tool returns the bound agent ID and worktree. If the
   bind token was already spent or expired, re-run `agent-bot join` to mint a
   fresh one, then retry the bind.

5. Verify the setup. Run `agent-bot doctor --json`. Expected: the JSON report
   shows:
   - A resolved bot identity (not `human`).
   - The correct worktree layout.
   - The resolved and pinned App slug.
   - Git author and committer matching the bot identity.
   - HTTPS origin and credential helper configured.
   - Installed hooks and chained hooks path.
   - Identity daemon supervisor healthy.
   - `worktree.binding` reports the worktree is bound to your App.
   - `worktree.session_binding` reports `ready` — the live-binding check,
     distinct from the worktree pin above.

   The doctor report separates worktree binding from session binding. A
   worktree can be configured (App pinned, attribution written) while the
   session is unbound — identity-bound capabilities refuse to run until the
   MCP bind exchange completes.

## Examples

```console
$ agent-bot --version
$ agent-bot join --name my-soul --harness claude-code
$ agent-bot setup-worktree
# Invoke the MCP bind tool with your transcript ID
$ agent-bot doctor --json
```

With a named worktree:

```console
$ agent-bot join --name my-soul --harness codex
$ agent-bot setup-worktree --name feature-branch --branch feat/my-feature
# Invoke the MCP bind tool with your transcript ID
$ agent-bot doctor --json
```

With JSON output for scripting:

```console
$ agent-bot join --name my-soul --harness vibe --json
# Invoke the MCP bind tool with your transcript ID
$ agent-bot doctor --json --require-schema-version 1
```

## Edge cases

- **Checkout already pinned to another soul:** `join` refuses. You are in a
  checkout bound to a different soul. Either work in that soul's context or
  choose a different checkout.
- **No session soul for setup-worktree:** `setup-worktree` leaves the checkout
  human and writes nothing (an error only with `--name`). Run `join` first.
- **Session binding missing:** `doctor` reports the worktree as configured but
  identity-bound tools refuse to run. Complete the MCP `bind` exchange (step 4)
  before attempting any GitHub operations.
- **Bind token expired or spent:** Re-run `agent-bot join` from the same
  checkout to mint a fresh token, then retry the MCP `bind`.
- **Missing App:** If the GitHub-identity add-on is enabled but no App resolves,
  `join` still succeeds (it needs no App). GitHub operations will fail until an
  App is configured. Run `agent-bot doctor` to diagnose.
- **Cross-device placement:** `setup-worktree` refuses cross-device worktree
  placement. No TMPDIR fallback. Choose a worktree location on the same device.
- **Conflicting pins:** A new worktree never keeps another soul's pin that git
  copied into it from the source checkout. The command handles this
  automatically.
- **Outside a git checkout:** `join` creates the soul's own `worktrees/workspace`
  and `git init`s it on first join. No GitHub attribution (author, credential
  helper, hooks) is written.
- **Rejoining:** Running `join` again from the same checkout reuses the soul.
  Rejoining fills a missing App on an active soul without replacing its identity.

## See also

- `verified-publish` — once you have a bot identity, publish signed commits
- `agent-bot skill agent-bot` — the bundled router skill for this CLI
- `agent-bot skill agent-bot --for setup-worktree` — the bundled operations
  reference with additional detail on bootstrap, minting, and diagnostics
- `agent-bot skill agent-bot --for doctor` — the bundled diagnostics reference
