---
name: review-flow
description: Review a pull request properly before approving or requesting changes. Use when asked to review, check, or give feedback on a PR
level: intermediate
risk: mutating
tags: [vcs, pr, review]
triggers: [review this pr, code review, approve, request changes]
commands: [pr view, pr diff, pr checks, pr review]
requires:
  - "gh pr view <n> succeeds (the pull request exists and is readable)"
  - "you are not the pull request author (gh cannot approve your own PR)"
min_version: "2.40"
prerequisites: [pr-flow]
next: [pr-merge]
---

# Review flow

## When to use

Someone asks you to review a pull request — really review it, not rubber-stamp
it. The deliverable is one submitted review: approve, comment, or request
changes, with evidence.

## Steps

1. Read the pull request as the author presented it:
   `gh pr view <n> --json title,body,author,labels,reviewDecision,mergeStateStatus`.
   Expected: you can state in one sentence what it changes and why.
   If you cannot, that is a review finding, not a reason to guess.
2. Read the actual diff: `gh pr diff <n>`. Verify the diff matches what
   step 1 claimed. A description that disagrees with the diff is a finding.
3. Check CI before reading code closely: `gh pr checks <n>`.
   Expected: all required checks pass, or failures are clearly unrelated
   flakes (say which). Do not approve over a red required check.
4. Read the changed code in the repo (not just the diff hunks) for the
   areas the diff touches: callers, tests, conventions in neighboring
   files. Note concrete findings as file:line with the failure mode.
5. Submit exactly one review:
   - No blocking findings: `gh pr review <n> --approve --body "<what you verified>"`.
   - Questions or non-blocking notes: `gh pr review <n> --comment --body "<notes>"`.
   - Blocking findings: `gh pr review <n> --request-changes --body "<findings with file:line>"`.
   The body must name what you checked, not just the verdict.
6. Verify: run `gh pr view <n> --json reviewDecision,latestReviews` and
   confirm your review is recorded with the intended decision. Report
   the decision and the top findings to the user.

## Examples

```console
$ gh pr view 123 --json title,body,author,labels,reviewDecision,mergeStateStatus
$ gh pr diff 123
$ gh pr checks 123
$ gh pr review 123 --request-changes --body "src/parser.ts:42 — empty input still throws; test at tests/parser.test.ts:18 covers only null"
$ gh pr view 123 --json reviewDecision,latestReviews
```

## Edge cases

- Your own pull request: GitHub forbids approving it. Do the full
  self-review anyway and leave findings with `--comment`.
- Draft PR: review the diff, but submit `--comment` — an approval on a
  draft signals readiness the author has not claimed.
- Huge diff: state explicitly what you did and did not read in the review
  body. A scoped, honest review beats a blanket approval.
- Checks still running: `gh pr checks <n> --watch` waits for them. Do not
  approve before they settle unless the user instructs it.

## See also

- `pr-flow` — how the pull request you are reviewing should have been opened
- `pr-merge` — landing it once it is approved and green
