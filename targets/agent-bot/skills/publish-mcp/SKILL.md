---
name: publish-mcp
description: Start a soul-bound worktree, review a change, and publish a verified bot commit and draft PR through hosted Agent Bot MCP when cloud credentials override local GitHub writes.
level: advanced
risk: mutating
tags: [cloud, mcp, worktree, identity, commit, publish]
triggers: [cloud bot workflow, signed commit, draft PR, avoid gh, soul worktree, bot identity]
commands: [start, doctor]
requires:
  - "agent-bot --version reports 0.15.18 or a compatible newer release"
  - "the identity daemon is running and the device is enrolled for the intended App"
  - "hosted Agent Bot MCP can access the repository as that same App"
min_version: "0.15.18"
prerequisites: []
next: []
---

# Publish through hosted Agent Bot MCP

## When to use

Use for a coding task in a cloud environment where the egress proxy replaces
GitHub credentials. Keep local work in a soul-bound worktree and make GitHub
writes on the Agent Bot server. Do not use `gh`, local `git push`, local REST
writes, or `agent-bot signed-commit` for this path: setting a bot token does
not establish the identity that the proxy sends upstream.

Done means the tested tree is published with a verified bot signature, the
correct soul attribution, and a draft PR authored by the intended App.

## Steps

1. Inspect the installed startup surface, then start the task from a checkout:

   ```sh
   agent-bot --version
   agent-bot start --help
   agent-bot start --repo <repo-path> --branch <task-branch> --no-enroll --json
   ```

   Read project instructions first. On 0.15.18, `start` creates/reuses the
   soul and linked worktree, sets attribution/hooks, and binds the session.
   Routine startup does not require a separate `setup-worktree` call. Enter
   the receipt's `worktree.path`. If startup fails, inspect partial state;
   a directory or identity pin alone does not prove a live binding. Repair
   the reported dependency, then rerun `start`; never bypass identity hooks.
   Missing enrollment needs the owner's enrollment procedure, not human auth.

2. Check the worktree identity and live session:

   ```sh
   git config --get agentBot.agentId
   agent-bot doctor --json
   git status --short
   ```

   Require the intended App, bot attribution, hooks, and both worktree binding
   and `worktree.session_binding` to be ready. Preserve unrelated changes.
   Record the repository, branch and tested base SHA. Never print binding files.

3. Discover hosted MCP tools and describe their current schemas:

   ```text
   search_tools({"query":"sign_commit github_read github_pull_request get_commit_identity"})
   describe_tool({"name":"github_read"})
   call_tool({"name":"whoami","arguments":{}})
   ```

   The authorizing human is not the bot. Match the explicit `harness` to the
   App resolved in step 2; do not guess another App to escape a refusal.
   A client's default `whoami` mapping may be unbound while its authorized
   explicit App works. Verify that exact App with a read before any write:

   ```text
   call_tool({"name":"github_read","arguments":{"harness":"<resolved-app>","repo":"<owner/repo>","path":"branches/<base-branch>"}})
   ```

   Require success and a base SHA matching the intended checkout. Stop on
   `no_identity` for that App or a scope/credential refusal. A local daemon
   binding does not automatically bind the hosted connector.

4. Make the change inside the returned worktree. Run the repository's required
   checks, inspect the full diff and new files, fix findings, and rerun affected
   checks. Review before publishing; this is not a GitHub self-approval.

   ```sh
   git diff --check
   git diff
   git add <intended-files>
   git diff --cached
   git commit -m "<why this change is needed>"
   git status --porcelain
   git show -s --format=%B HEAD
   ```

   Require a clean worktree and the runtime-produced `Agent-Identity` and soul
   co-author trailers. Never invent another soul's attribution or bypass hooks.
   Inspect all new files explicitly: an unstaged diff omits untracked files.

5. **Remote write:** describe `sign_commit`, then send the exact tested change
   through the write dispatcher. This creates a new signed commit and publishes
   the branch; it does not sign an existing local commit in place.

   ```text
   call_write_tool({"name":"sign_commit","arguments":{"harness":"<resolved-app>","repo":"<owner/repo>","branch":"<task-branch>","base":"<tested-base-sha>","message":"<reviewed-message-with-runtime-trailers>","session_id":"<conversation-id>","files":[{"path":"<relative-file>","content":"<complete-tested-content>"}]}})
   ```

   For an existing remote branch, replace `base` with `parent`: the exact head
   SHA the change was built on. A moved parent must fail, never be overwritten.
   Include deletions, binary bytes and executable modes using the described
   schema. Limits are 100 files and 700 KiB per call. Split only independent,
   separately validated changes; otherwise stop and report the transport limit.
   Do not fall back to local publishing through the credential-replacing proxy.

6. Verify the returned SHA before opening the PR:

   ```text
   call_tool({"name":"get_commit_identity","arguments":{"repo":"<owner/repo>","sha":"<published-sha>"}})
   call_tool({"name":"github_read","arguments":{"harness":"<resolved-app>","repo":"<owner/repo>","path":"git/commits/<published-sha>"}})
   ```

   Require the intended bot author, verified signature, soul trailers, and a
   tree SHA equal to the tested local tree (`git rev-parse HEAD^{tree}`). If
   the tree differs, stop and investigate; do not report a successful publish.
   Read the remote branch head after a timeout before retrying a remote write.

7. Describe `github_pull_request`, then create a draft as the same App:

   ```text
   call_write_tool({"name":"github_pull_request","arguments":{"harness":"<resolved-app>","repo":"<owner/repo>","head":"<task-branch>","base":"<base-branch>","title":"<change outcome>","body":"<behavior, validation, limitations>","draft":true,"session_id":"<conversation-id>"}})
   ```

   Read the PR back with `github_read` at `pulls/<number>`. Require `draft:true`,
   the intended bot author, and the published head SHA. Return its URL and any
   incomplete evals. Do not approve this App's own PR.

## Examples

```sh
agent-bot start --repo <repo-path> --branch docs/cloud-workflow --no-enroll --json
```

Inside the returned worktree, recheck an existing session without a new branch:

```sh
agent-bot start --no-enroll --json
```

MCP JSON above is a tool invocation template, not a shell command. Substitute
observed values and use the connector's actual meta-tool names.

## Edge cases

- `start: fetch failed`: inspect daemon health and setup state; do not assume
  it means `git fetch` failed. In the trial, starting the missing daemon and
  rerunning `start` established the binding.
- Missing clone capability: try authorized read-only HTTPS Git cloning first.
  If unavailable, request a repo-scoped clone/archive tool in Agent Bot using
  `github_issue` through `call_write_tool`, when the same App has issue access.
- Missing identity: report the structured setup link. Issue creation also
  needs identity; if it fails, retain the issue draft locally and report it.
- Failed verification: preserve the branch and investigate. Correct with a
  follow-up signed commit or close the draft PR; never force-push over others.
- Full trial and remaining evidence: [test workflow](references/test-workflow.md).

## See also

- `verified-publish` — local publishing only where credentials reach GitHub intact.
- `agent-bot skill agent-bot` — installed startup and identity guidance.
