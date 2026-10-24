---
name: pr-flow
description: Open a pull request with a complete title, body, and reviewers. Use when a branch has commits ready for review
level: intermediate
risk: mutating
tags: [vcs, pr, review]
triggers: [pull request, open a pr, code review, ready for review]
commands: [pr status, repo view, pr create, pr edit, pr view]
requires:
  - "gh pr status shows no open pull request for this branch"
  - "the branch is pushed to origin (git status shows no unpushed commits)"
  - "gh auth status succeeds for the target host"
min_version: "2.40"
prerequisites: []
next: [review-flow, pr-merge]
completes_with: pr create
---

# PR flow

## When to use

Your branch has commits ready for review and you want to open a pull request
a reviewer can act on without asking questions.

## Steps

1. Run `gh pr status` to check whether a pull request already exists for this
   branch. Expected: the "Current branch" section reports no pull request.
   If one exists, stop here and update it with `gh pr edit` instead (see
   Edge cases).
2. Review the full diff you are about to send. There is no pull request
   yet, so `gh pr diff` cannot do this — diff against the base branch:
   `base=$(gh repo view --json defaultBranchRef --jq .defaultBranchRef.name)`,
   then `git fetch origin "$base"` and `git diff "origin/$base...HEAD"`.
   Expected: only the changes belonging to this branch. If unrelated
   changes appear, they belong on another branch — stop and say so.
3. Read the repo's pull request template and follow its sections. GitHub
   looks in `.github/pull_request_template.md`, `pull_request_template.md`,
   `docs/pull_request_template.md`, and `.github/PULL_REQUEST_TEMPLATE/`
   (details in `references/body-template.md`). If none exists, use
   Summary, Test plan, Risks.
4. Run `gh pr create --title "<title>" --body "<body>"` with a title that
   states the change and a body that states the why, the test plan, and any
   risks. Expected: the command prints the new pull request's URL. If it
   fails with "no commits between", the branch is not pushed or has no new
   commits — see Edge cases.
5. Run `gh pr edit --add-reviewer <logins> --add-label <labels>` to assign
   reviewers and labels. Never invent reviewer logins; use only names the
   user or the repo's CODEOWNERS provides.
6. Verify: run `gh pr view --json url,reviewRequests,labels` and confirm
   the URL resolves, the intended reviewers are listed, and the labels
   match step 5. Report the URL to the user. This step is the definition
   of done — do not report success without it.

## Examples

```console
$ gh pr status
$ base=$(gh repo view --json defaultBranchRef --jq .defaultBranchRef.name)
$ git fetch origin "$base" && git diff "origin/$base...HEAD"
$ gh pr create --title "fix: handle empty input in parser" --body "<body following the repo template>"
$ gh pr edit --add-reviewer octocat --add-label bug
$ gh pr view --json url,reviewRequests,labels
```

## Edge cases

- If `gh pr status` shows an existing pull request, update it with
  `gh pr edit` instead of creating a duplicate.
- If `gh pr create` fails with "no commits between", the branch has nothing
  new to compare — push it first with `git push -u origin <branch>`.
- If the repo has no pull request template, use three body sections:
  Summary, Test plan, Risks.

## See also

- `review-flow` — once the pull request is open, review someone else's
- `pr-merge` — when a pull request is approved and green, land it
- `references/body-template.md` — template locations and `--body-file`
