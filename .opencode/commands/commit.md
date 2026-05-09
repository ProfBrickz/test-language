---
description: Stage all changes, preview commit message, then commit locally (no push)
---

Stage all changes (`git add -A`), then analyze the staged diff and recent commit history to draft a commit message.

First, run these commands to understand the current state:
- `git add -A`
- `git status`
- `git diff --staged`
- `git log --oneline -10`

Based on the changes, draft a commit message with:
- A short summary line (imperative mood, no period, ~50 chars)
- A blank line followed by a longer description explaining what changed and why, formatted as a bullet points list (with headers to split it up if it is long)

The summary format should follow existing patterns in the repo (use `git log` output as style reference).

Output the summary as a separate code block:
```
<summary line>
```

Output the description as a separate code block:
```
<description paragraph>
```

Then create the commit locally with:
```
git commit -m "<summary>" -m "<description>"
```

DO NOT push the commit to any remote. This is a local commit only.
