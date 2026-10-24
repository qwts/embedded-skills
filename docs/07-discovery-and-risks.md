# Discovery, Adoption, and Risks

## The bootstrap problem

Everything else in this package is engineering; this is the bet. No model is
trained to run `cli skills`. Left alone, an agent does what it always does:
`cli --help`, then trial and error. Skills can be perfect and never be read.
Discovery must hang on surfaces agents already read, ranked by reliability:

1. **Harness/user instructions — strongest.** The session bootstrap (doc 01):
   the harness injects `cli skills --json` into the agent's context at session
   start. Plus a paste-able `AGENTS.md` snippet
   (`examples/agents-md-snippet.md`) for users who install the shim
   themselves.
2. **`--help` augmentation — opt-in, off by default.** One appended line:
   `Agent skills for this CLI: run 'gh skills'.` Agents read `--help`
   constantly, so this converts existing behavior into discovery. It mutates
   another program's output, so it stays a conscious opt-in, limited to help
   text (doc 06).
3. **Harness-native export.** `cli skills export <name>` drops the skill
   directory where file-based harnesses already look (`.claude/skills/`,
   `.agents/skills/`, …). Zero new agent behavior — the pragmatic default for
   harness users, with the subcommand as the universal fallback.
4. **Error-message pointers — future, careful.** When the real CLI fails in a
   way a skill addresses, the shim could append `hint: gh skills <name>` to
   stderr. High value, high intrusiveness: stderr is parsed by scripts. Only
   behind an explicit flag, never in v1.
5. **Upstream adoption.** The endgame is CLIs implementing the probing
   contract natively. Keep the contract tiny and CLI-agnostic so this stays
   possible; the shim proves the concept on tools you don't own.

Measure discovery honestly: in evals, does the agent find skills *without*
being told the exact commands? If only channel 1 works, say so and lean into
it.

## Risk register

| Risk | Severity | Mitigation |
|------|----------|------------|
| Agents never discover `skills` | Fatal if unaddressed | Bootstrap channels above; eval discovery explicitly (doc 09) |
| Shim breaks the real CLI (argv/TTY/signal/exit-code infidelity) | Fatal to trust | `syscall.Exec`; proxy-fidelity test suite as release gate (doc 09) |
| PATH shim shadows wrong or stale binaries | High | Skip-self resolution via `os.SameFile`; per-target override env var; install never edits rc files; `doctor` diagnostics (doc 06) |
| Skill content is wrong for the installed target version | High | `target_version` + version probe + per-skill `min_version`; never serve silently out-of-range (docs 02, 06) |
| Skills become a prompt-injection / supply-chain vector | High | Treat skill PRs as code PRs; validator linting; signed releases; no bundled executables; no runtime fetching (doc 09) |
| Agents bypass the shim and call the real binary directly | Expected behavior, not a vulnerability | The shim shapes the convenient path; enforcement belongs to the CLI, the OS, and credentials. Don't market it as a security control |
| Target grows its own `skills` subcommand | Medium | Collision policy: real CLI wins; alternate name via env (doc 06) |
| Catalog bloat as skills accrete | Medium | Hard budgets in the validator; trigger quality over coverage (doc 08) |
| Maintenance rot: skills drift from target behavior | Inevitable without process | Corpus maintenance process (doc 06); eval fixtures re-run on target upgrades; a skill with a failing eval is removed, not left stale |

## Anti-patterns to refuse

- **Script-based shims** (`.sh`/`.py` per target): readable, editable,
  interpreter-dependent, leak real paths mid-script.
- **Hiding the shim's existence** from humans: `doctor` and `--version` should
  make shimming fully inspectable. Invisible PATH magic destroys trust the
  first time something odd happens.
- **Skills that restate `--help`:** pure drift, zero workflow value.
- **A skill per subcommand:** catalog bloat; trigger collapse.
- **Runtime-downloaded skills without signing/pinning:** combines supply-chain
  risk with prompt injection.
- **Measuring success by skill count.** Measure by eval win-rate over
  `--help`-only baselines and by discovery rate.

## Phasing

1. **v0 — prove the shim.** One target, 3–5 skills mined from real agent
   failure transcripts. Discovery via harness bootstrap + `AGENTS.md` snippet
   only.
2. **v0.2 — prove the value.** Evals vs `--help`-only; `export`; version
   probing; validator + golden tests in CI.
3. **v0.3 — widen.** Second and third targets; `search`; opt-in `--help`
   pointer line; the authoring guide (doc 08) as the contributor guide.
4. **Later, gated on stage-2 evals:** usage tracking, `skills next`,
   `skills feedback` (doc 04's deferred design) — only once discovery and
   completion are measured to work without them.
5. **Later / maybe never:** user overlay skills, error hinting, signed runtime
   skill packs, upstreaming the contract to a CLI you maintain.
