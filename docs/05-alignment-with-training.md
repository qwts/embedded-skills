# Alignment With Agent Training

Short answer: yes. The design is aligned with what agents are trained on,
because it reuses conventions agents already know instead of inventing new
ones. The novel part is *where* the skills live (inside the CLI, tied to its
real command tree) — not the shape of the content.

## Pattern by pattern

| What agents are trained on | How this design mirrors it |
|----------------------------|----------------------------|
| Tool-calling loops: enumerate capabilities, pick one, call it, read the output | `cli skills` is the enumeration step for workflows; the agent then calls the commands the skill describes |
| `--help`, `help`, `-h` conventions | `skills` is the agent-facing sibling of `--help`: same discoverability role, different audience |
| Structured output: `--json`, git's `--porcelain` | Every `skills` surface has a `--json` form with a versioned schema |
| Progressive disclosure: metadata always visible, full body loaded on demand | `cli skills` (layer 1) vs `cli skills <name>` (layer 2) vs `references/` (layer 3) |
| Frontmatter + Markdown skill documents | The skill format is exactly that: YAML frontmatter for routing, Markdown body for instructions |
| Injected instructions: system prompt, loaded skill bodies steer behavior | The harness fetches `cli skills <name>` and puts the body in context, exactly as with a file-loaded skill; the optional `--skill` flag is a fallback for harnesses that can't |
| git's "did you mean …?" suggestions | Unknown skill names produce the same pattern, which agents handle well |
| Composability: pipes, `&&`, small tools that do one thing | Skills teach composed workflows; each skill is one workflow, chains compose them |
| Numbered, imperative procedures | Skill bodies are numbered steps with fenced commands |

## The one-sentence version

A human learns a CLI from `--help`; an agent learns it from `skills` — and
because `skills` speaks in the formats agents are already trained to parse
(frontmatter, Markdown, JSON, numbered steps), the agent needs no new training
to use it.

## Cautions — where alignment can break

1. **Trust boundary.** Skill text arriving inside tool output can look like
   prompt injection to a cautious agent. Mitigations: skills are first-party
   and embedded in the binary (not read from a user-writable path), the
   canonical path is an explicit harness fetch (`cli skills <name>`) rather
   than content interleaved with command output, any `--skill` output is
   clearly delimited, and `cli skills --help` states that embedded skills
   are trusted first-party instructions.
2. **Context budget.** Agents are trained to be economical with context. Cap
   `SKILL.md` at 200 lines, push depth into `references/`, and keep
   `description` to one line. A bloated skill is a skill the agent skips.
3. **Output stability.** Agents parse `--json` output. Treat the schemas as a
   versioned API: no field renames without a major bump, no surprise type
   changes.
4. **Don't invent a new format.** The alignment argument only holds if the
   skill format stays the familiar frontmatter + Markdown shape. This
   design's extra fields (`level`, `commands`, `prerequisites`, `next`, …)
   are declared extensions to the Agent Skills spec, not spec fields —
   additive metadata, removable on export, never the sole carrier of
   anything an agent needs (doc 02, "Conformance"). The temptation to grow
   them into a new document type is the thing to resist.
5. **Help is not skills.** Don't merge the two surfaces. `--help` answers
   "what does this flag do"; `skills` answers "how do I accomplish X". Agents
   need both, and conflating them makes each worse at its job.

## Relationship to MCP

Skills teach; MCP executes. An MCP server gives agents callable tools and
live data; embedded skills give agents knowledge of how to use the CLI they
already have. They compose — a CLI with an MCP server may need fewer skills,
not zero — and they are not substitutes.

## Verdict

Aligned. The design's real risk is not misalignment with agent training — it
is drift between the corpus and the CLI, which the build-time embedding and
`cli skills check` (see `03-repo-structure.md`) exist to prevent.
