# Cloud publishing trial

## Task and fixture

The user requested a reusable skill for soul/worktree startup, local work and
commits, internal review, signed bot publishing through Agent Bot MCP, and a
draft PR without using `gh`. They authorized choosing a small repository and
cloning it, or filing an Agent Bot tool request if transport was unavailable.

Repository: `qwts/embedded-skills`, a small existing Agent Bot skill corpus.
Base: `40ce0057384535b85894fefc1bae70c50fdc4706` (`main`).
Branch: `docs/agent-bot-mcp-workflow`.
Target: installed `agent-bot` 0.15.18.
Runtime bundle source: `d5a28db5f4f6cbdd9e36a58d8c4197a0229f61ba`.
App: `trueline-codex-agent`, resolved from the local runtime, then checked
explicitly with hosted MCP. No human credential fallback was used.

## Observed startup and recovery

1. Read-only HTTPS `git clone` succeeded; no clone tool request was necessary.
2. `agent-bot start --repo <clone> --branch docs/agent-bot-mcp-workflow
   --no-enroll --json` returned `start: fetch failed`. A linked worktree and
   soul pin already existed; startup was not atomic.
3. `agent-bot doctor --json` showed the daemon was not running and the
   worktree had no usable live binding. Read-only `git fetch origin`
   succeeded, so retrying Git fetch was not the repair.
4. `agent-bot daemon start --daemonize` started the installed daemon.
   Repeating the same `start` call succeeded with `binding.status: live`.
5. From the returned worktree, `agent-bot start --no-enroll --json`
   returned `resolution: reused-bound` and `binding.alreadyBound: true`.
   Every worktree check in `doctor` was then `ready`, including
   `worktree.binding` and `worktree.session_binding`.
6. Hosted MCP `whoami` reported its default client mapping as unbound.
   A `github_read` using that default mapping returned `no_identity`.
   A read with the locally resolved `trueline-codex-agent` App succeeded
   and returned the expected `branches/main` SHA. Default mapping status
   alone was therefore insufficient to decide whether this App could act.

The user's `setup-worktree` → enter worktree → `start` sequence separates
preparation from binding. The installed 0.15.18 help recommends `start`
for routine startup and marks `setup-worktree` deprecated; the one-command
path performed both once its daemon dependency was running.

## Reproducible local checks

From `targets/agent-bot`, with an actual Go compiler selected through `GO`:

```sh
make manifest
make test
make check
./dist/agent-bot skills
./dist/agent-bot skills publish-mcp
./dist/agent-bot skills publish-mcp --json
./dist/agent-bot skills search 'avoid gh'
```

Check that the served Markdown matches `skills/publish-mcp/SKILL.md`
byte-for-byte, JSON reports the skill and its reference, and search discovers
it through the user's vocabulary. Check proxy fidelity with `--version`,
`start --help`, and an unknown subcommand; compare stdout, stderr and exit
status with the installed target. Keep shim use process-local; do not change
the user's shell profile or install over an identity wrapper.

The image's `go` executable was not the Go compiler (`go version` returned
`Go: Unknown option: version`). An official Go archive was downloaded to a
scratch directory and its SHA-256 checked against Go's release metadata.

The corpus checker probes command paths with `--help`. The valid runtime
command `identity current` rejects that flag in 0.15.18, so the skill uses
the worktree's Git pin plus `doctor` for identity inspection instead. The
final corpus check still validates every declared Agent Bot command path.

Observed results with Go 1.27.4 and Agent Bot 0.15.18:

| Check | Result |
| --- | --- |
| `make manifest` | Generated three skills and checksums |
| `make test` | Passed target corpus and checksum tests |
| `make check` | Passed; three skills, 551-byte catalog |
| Served Markdown | Matched authored bytes |
| Skill JSON | Parsed; test-workflow reference discoverable |
| Catalog JSON | Detected 0.15.18; all skills version-compatible |
| Search for `avoid gh` | Found `publish-mcp` |
| Proxy `--version` and `start --help` | stdout/stderr/status matched real CLI |
| Unknown real subcommand | stdout/stderr/status matched real CLI |
| Unknown skill | Exited 2 as documented |

## Internal review and remote acceptance

Before publishing, review the full staged diff and required checks. Preserve
the runtime's commit trailers. Send the complete tested file changes through
hosted MCP `sign_commit` with the base SHA above and the same resolved App.
Require the published tree SHA to equal `git rev-parse HEAD^{tree}`, a valid
signature, bot authorship, and the expected soul attribution. Open a draft
with `github_pull_request`, then read it back to confirm author, head and
draft state. The PR is the durable remote receipt; no local GitHub write is
part of this procedure.

## Evaluation limits

This is a manual live workflow trial, not a controlled two-agent evaluation.
No automated agent evaluation harness or unprompted discovery score is claimed.
The initial unscripted attempt had two failed `start` calls before daemon
recovery; the skill records that failure and the observed repair. Help-only
startup guidance did not cover the hosted connector's separate identity or
server-side publish transport. A clean paired help-only/skill-loaded run is
still needed before claiming an improvement score or release readiness.
Keep the PR in draft for that review.
