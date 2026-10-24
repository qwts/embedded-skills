# Working in this repository

Before adding a target skill, adding a CLI target, or building/installing
a skills shim, load the repository's root contributor skill:

`.agents/skills/embedded-skills/SKILL.md`

It routes the work to exactly one procedure under
`.agents/skills/embedded-skills/references/`. The design documents in
`docs/` remain the source of truth if the skill and docs ever disagree.

Delivery is always the PATH-shadow shim. Do not add target-specific
plugin/extension integrations.
