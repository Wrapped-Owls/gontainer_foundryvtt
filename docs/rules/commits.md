# Commits

We use **Conventional Commits** with an emoji prefix that mirrors the existing `git log` of this
repo.

## Format

```
:emoji: type: short subject in lowercase
```

- `:emoji:` is a single gitmoji shortcode (rendered by GitHub). Pick one that summarises the
  intent — `:sparkles:` for new features, `:wrench:` for fixes, etc.
- `type` is a Conventional Commit type: `feat`, `fix`, `chore`, `refactor`, `docs`, `test`,
  `perf`, `build`, `ci`. Scopes are optional and use parentheses: `feat(msgsender):`.
- The subject is **lowercase**, imperative mood, no trailing period, ≤ 72 characters.

## Examples (from this repo's history)

```
:cloud: feat: Complete overhaul of HTTP server and context handling
:gear: feat: [wip] Implement infra configuration loading
:open_file_folder: refactor: Move tools to root folder
:package: chore: Update tools dependencies to latest
:page_with_curl: docs: Create more docs about implementation plan
```

The leading `:emoji:` and the `type:` are both required for new commits.

## Body and footer

- Wrap the body at 72 columns. Explain **why**, not what (the diff already shows what).
- Reference issues/tickets in the footer: `Refs: #42`, `Closes: #99`.
- For breaking changes, include a `BREAKING CHANGE:` footer with migration notes.

## What goes in one commit

- One logical change. A refactor and a feature do not share a commit.
- A commit must compile and pass `task main:test` on its own.
- Doc‑only commits use `docs:` and do not need to pass tests.

## Reviews and squashing

- PRs may contain multiple commits. Keep them clean (`git rebase -i`) before requesting review.
- Merge strategy is **rebase**, not squash, so the curated history reaches `main`.

## Co‑authoring (for AI‑assisted commits)

When committing on behalf of a Copilot/Claude‑assisted change, append the trailer:

```
Co-authored-by: Copilot <223556219+Copilot@users.noreply.github.com>
```
