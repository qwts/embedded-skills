---
name: join-and-bind
description: Join agent-comms as a soul and bind a worktree to a bot identity. Use when starting a new agent session that needs to communicate with the fleet or act on GitHub
level: intermediate
risk: mutating
tags: [identity, join, soul, worktree, mcp, bind, agent-comms, setup]
triggers: [join the hub, become a soul, setup worktree, bind identity, agent-comms, start as agent, join agent-comms, new agent session]
commands: [join, setup-worktree, doctor, mcp]
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

4. Verify the setup. Run `agent-bot doctor --json`. Expected: the JSON report
   shows:
   - A resolved bot identity (not `human`).
   - The correct worktree layout.
   - The resolved and pinned App slug.
   - Git author and committer matching the bot identity.
   - HTTPS origin and credential helper configured.
   - Installed hooks and chained hooks path.
   - Identity daemon supervisor healthy.

5. The MCP bind is automatic. The `join` command mints a single-use bind token,
   so the MCP `bind` tool works in this checkout. No separate MCP configuration
   is needed. The MCP server (`agent-bot mcp`) provides tools: `bind`,
   `whoami`, `population`, `space_path`, `credential`, `take_inbox`.

## Examples

```console
$ agent-bot --version
$ agent-bot join --name my-soul --harness claude-code
$ agent-bot setup-worktree
$ agent-bot doctor --json
```

With a named worktree:

```console
$ agent-bot join --name my-soul --harness codex
$ agent-bot setup-worktree --name feature-branch --branch feat/my-feature
$ agent-bot doctor --json
```

With JSON output for scripting:

```console
$ agent-bot join --name my-soul --harness vibe --json
$ agent-bot doctor --json --require-schema-version 1
```

## Edge cases

- **Checkout already pinned to another soul:** `join` refuses. You are in a
  checkout bound to a different soul. Either work in that soul's context or
  choose a different checkout.
- **No session soul for setup-worktree:** `setup-worktree` leaves the checkout
  human and writes nothing (an error only with `--name`). Run `join` first.
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
