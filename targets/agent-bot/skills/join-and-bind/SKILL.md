---
name: join-and-bind
description: Check in as a soul and bind this checkout: join, MCP bind, setup-worktree, doctor, and the fail-closed errors for a missing App, missing binding, or wrong checkout. Use when starting a fresh session
level: intermediate
risk: mutating
tags: [agent-bot, join, bind, soul, worktree, mcp, setup]
triggers: [join, check in, bind, setup worktree, soul, agent-comms, hub, new session, fresh checkout, onboard, wake]
commands: [join, setup-worktree, mcp, doctor]
requires:
  - "agent-bot join --name NAME --harness HARNESS succeeds and reports an agentId"
  - "the MCP bind tool returns a binding for this worktree"
  - "agent-bot setup-worktree completes in the linked worktree"
min_version: "0.12.0"
prerequisites: []
next: [verified-publish]
completes_with: setup-worktree
---

# Join and bind

## When to use

You are an agent that nobody launched, working in a checkout that is not
yet checked in: you need to become a soul, join agent-comms, bind this
conversation to the soul, and configure the worktree so commits and pushes
attribute to the bot. Use this when starting work in a fresh checkout or
session, or when `agent-bot doctor` reports a missing binding, a missing
App, or a checkout that belongs to another soul.

## Steps

1. Check in: `agent-bot join --name NAME --harness HARNESS --json`. Run it
   inside the checkout you will work in; from an unpinned checkout, pass
   `--soul AGENT_ID` to reuse your canonical long-lived soul instead of
   spawning a task-scoped one. Expected: JSON with `agentId`, `soulDir`,
   `worktree`, `address`, `created`, `bind`, `wake`. Running join again
   from the same checkout reuses the soul. Join needs no GitHub App: with
   the `github-identity` add-on enabled, configured harness markers resolve
   the App automatically; with no mapping or the add-on off, a hub-only
   join still works. Join writes the soul directory, pins
   `agentBot.agentId` in the checkout's worktree config, records the
   checkout in the census, and mints a single-use bind token — it writes no
   GitHub attribution. `--wake resume:read-only|resume:workspace|acp` makes
   new messages wake the soul; cold wake is owner only, the owner gate runs
   before anything is created, and a refusal changes nothing.
2. Bind the conversation. Set `AGENT_BOT_ID` from the join result, then
   call the MCP `bind` tool — served by `agent-bot mcp`, which your harness
   mounts — once at conversation start, before any other agent-bot tool,
   with `transcript_id` (required), `transcript_provider`, and
   `parent_agent_id` when this conversation was spawned by a bound agent.
   Expected: the tool returns the binding; the secret is held in the
   worktree's private git dir (`<git-dir>/agent-binding.json`, mode 0600)
   and is never returned or logged. Confirm with the `whoami` tool. Treat
   the binding file as a credential: never print, copy, or commit it.
   `AGENT_BOT_BINDING` names a different binding file and wins over the
   worktree's own.
3. Configure the worktree: `agent-bot setup-worktree` in the same linked
   worktree. Expected: it verifies credentials and writes Git attribution
   for the soul's App (`agentBot.app`, `user.name`, `user.email`,
   `core.hooksPath`, `credential.helper`) and mints a fresh bind token.
   `setup-worktree --name NAME [--branch BRANCH]` first creates a linked
   git worktree under the soul's `worktrees/<name>` (branch defaults to
   NAME). It refuses primary checkouts outside the soul, arbitrary
   directories, conflicting pins, and cross-device placement; with no
   session soul it leaves the checkout human and writes nothing.
4. Confirm: `agent-bot doctor --json --require-schema-version 1`. Expected:
   machine and worktree readiness agree — detected harness and account,
   resolved App, Git author and committer, HTTPS origin and credential
   helper, hooks, daemon, and census. `doctor` is read-only and has no
   repair mode; use `agent-bot bootstrap` for the explicit repair boundary.

## Examples

```console
$ agent-bot join --name my-soul --harness codex --json
$ export AGENT_BOT_ID=<agentId from the join result>
# the harness calls the MCP bind tool: {"transcript_id": "<session id>", "transcript_provider": "codex"}
$ agent-bot setup-worktree
$ agent-bot doctor --json --require-schema-version 1
```

## Edge cases

Fail closed — read the error, fix the layer it names, and re-run; never
route around it:

- Missing App: `session soul has no GitHub App; re-run agent-bot join in
  this checkout to resolve its configured harness mapping`. Re-run the same
  join command in the linked checkout — rejoin fills only missing App
  metadata and preserves the soul and binding; it never overrides an
  existing different assignment, a retired identity, or a revoked binding.
  Then re-run setup-worktree.
- Missing binding: `session binding is missing; check in again with
  agent-bot join` (setup-worktree), or `session binding is missing;
  refusing to join` (join, when `AGENT_BOT_BINDING` names a binding file
  that does not exist). MCP bind errors: `bind requires transcript_id`;
  `no bind token is minted for this worktree — run `agent-bot
  setup-worktree` first`; `bind failed: the worktree binding could not be
  read … run `agent-bot doctor`, then `agent-bot setup-worktree` if the
  binding is corrupt`; `bind failed: the daemon is unreachable to re-bind
  this worktree … run `agent-bot daemon status` and `agent-bot doctor`,
  then retry`. Other MCP tools answer `not bound — call the bind tool
  first`.
- Wrong checkout: `this checkout is already pinned to <id>; join from
  another folder`; `checkout belongs to another soul; use a new worktree`;
  `checkout binding belongs to another soul`; `session soul and binding
  disagree`; `multiple souls match this session; supply AGENT_BOT_ID from
  agent-bot join`; `session soul is retired`; `named worktree is not a
  checkout of this repository`; `named worktree uses a different branch`.
- Harness mismatch: `<agentId> runs <harness>; rejoin with its own harness`.
- App conflict: `GitHub App does not match the session soul or checkout
  pin`; `GitHub App does not match the session soul; use a new worktree`.
- Credential backends: for `store: "keyd"` and soul-marked
  `store: "pass-cli"`, setup-worktree stops at local preparation
  (`keyd-held` / `owner-only`) unless the App credential is imported into a
  readable `file` or `keychain` store (`agent-bot ensure-private-key --app
  <slug>`). These backends have no unbound `GH_AGENT_APP` mint path — the
  daemon binding (join + MCP bind, or `join --wake acp`) must exist first.
- gh shim token contract: when any soul marker is present (`GH_AGENT_APP`,
  `agentBot.app`, `AGENT_BOT_ID`, `agentBot.agentId`, or
  `<git-dir>/agent-binding.json`), the fail-closed gh shim that agent-bot
  installs sets `SOUL_BOUND=1`, drops any inherited `GH_TOKEN`, and requires
  a live daemon binding to mint. Never unset `GH_TOKEN` to expose a stored
  human login from an agent session.
- Revoking: `agent-bot binding revoke` ends a binding and its spawned
  descendants.
- Never bind a human's own checkout on harness detection alone, and never
  treat a `.<tool>/worktrees/**` path as an identity signal. A stated bot
  whose worktree setup failed is refused `git commit` and `git push` — fix
  the setup; do not unstate the identity to get past the refusal.

## See also

- `verified-publish` — once the checkout is bound and set up, publish bot
  commits with verified signatures
- `agent-bot skill agent-bot` — the bundled router skill; its operations
  reference covers setup and recovery in depth
- `agent-bot skill agent-bot --for setup-worktree` and `--for doctor` — the
  bundled per-subcommand references
- `agent-bot skill agent-space` — the bundled skill for the soul's work area
