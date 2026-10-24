# PR body: templates and `--body-file`

Load this from `pr-flow` step 3 when the template situation is unclear.

## Where GitHub looks for a template

In priority order, GitHub uses the first it finds:

1. `.github/PULL_REQUEST_TEMPLATE/<name>.md` — a directory of named
   templates; the author picks one. If the repo uses these, `gh pr
   create --template <name>` selects it instead of hand-copying sections.
2. `.github/pull_request_template.md`
3. `pull_request_template.md` (repo root)
4. `docs/pull_request_template.md`

Capitalization variants (`PULL_REQUEST_TEMPLATE.md`) also match.

## Writing the body without shell-quoting pain

For anything multi-line, prefer `--body-file` over `--body`: write the
body to a temporary file, then

```console
$ gh pr create --title "<title>" --body-file /tmp/pr-body.md
```

This avoids quoting/escaping bugs in `--body "..."` when the text contains
backticks, `$`, or quotes. Do not leave the body file in the repository.

## No template? Use this shape

```markdown
## Summary
<what changed and why, 2–5 lines>

## Test plan
<commands run and observed results — not intentions>

## Risks
<what could break, what was deliberately not done, follow-ups>
```

Never claim tests ran that you did not run. If the test plan could not be
executed, say so and why under Risks.
