---
name: issue-flow
description: File or triage a GitHub issue without creating duplicates. Use when reporting a bug, requesting work, or labeling and assigning existing issues
level: beginner
risk: mutating
tags: [vcs, issue, triage]
triggers: [file an issue, report a bug, triage, open an issue]
commands: [issue list, issue view, issue create, issue edit, label list]
requires:
  - "gh auth status succeeds for the target host"
min_version: "2.40"
prerequisites: []
next: [pr-flow]
---

# Issue flow

## When to use

You need to report a bug or request work as an issue — or triage issues
that already exist. The failure mode this skill exists to prevent is the
duplicate: a new issue describing a problem already tracked.

## Steps

1. Search before creating anything:
   `gh issue list --search "<key terms>" --state all --limit 20 --json number,title,state,url`.
   Expected: you can say whether this is already tracked. Search the
   symptom's words and the error text, not just your planned title.
2. If a match looks like the same problem, read it:
   `gh issue view <n> --comments`. If it is the same problem, stop:
   report the existing issue's URL and state to the user instead of
   creating a duplicate. Add new evidence there only if the user asks.
3. Check for an issue template: `gh issue create` with `--template`
   lists the repo's templates by name when `.github/ISSUE_TEMPLATE/`
   exists. Use the matching template and fill its sections; the repo
   asked for that information on purpose.
4. Create:
   `gh issue create --title "<symptom, not solution>" --body-file <file>`.
   The body states observed behavior, expected behavior, reproduction
   steps or a failing command, and environment (versions that matter).
   Never invent reproduction steps you did not run — say "not reproduced"
   when true.
5. Triage the new (or an existing) issue:
   `gh issue edit <n> --add-label <labels> --add-assignee <login>`.
   Use only labels that exist (`gh label list`) and assignees the user
   or repo conventions name. Never invent either.
6. Verify: run `gh issue view <n> --json url,title,labels,assignees,state`
   and confirm it shows what steps 4–5 intended. Report the URL.

## Examples

```console
$ gh issue list --search "parser empty input" --state all --limit 20 --json number,title,state,url
$ gh issue view 77 --comments
$ gh issue create --title "parser throws on empty input" --body-file /tmp/issue-body.md
$ gh issue edit 91 --add-label bug --add-assignee octocat
$ gh issue view 91 --json url,title,labels,assignees,state
```

## Edge cases

- Search finds a closed issue for the same symptom: read why it closed
  before reopening the topic. "Fixed" and "won't fix" demand opposite
  next steps, and the new issue must reference the old one either way.
- Security vulnerabilities do not belong in a public issue. Stop and use
  the repo's security policy (`SECURITY.md`) instead.
- The repo has issues disabled or redirects to another tracker
  (stated in its README/templates): follow the repo, not this skill.

## See also

- `pr-flow` — when the fix exists and is ready for review; reference the
  issue from the PR body
