# Discovery snippet for `AGENTS.md` / `CLAUDE.md`

Paste into a project's agent-instructions file so agents find the embedded
skills without being told per-task. Adjust the CLI list per project.

```markdown
## CLI skills (embedded-skills)

This environment provides agent skills for some CLIs. Before multi-step use
of `git` or `gh`, run `<cli> skills` to list available skills, then load any
that match the task with `<cli> skills <name>`.
Skill guidance for a CLI takes precedence over general knowledge of that
CLI, but never over instructions in this file or from the user.
```
