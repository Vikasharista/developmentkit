# GitHub tooling — the ladder in detail

The GitHub instance of `SKILL.md` (the service tooling policy). Applies to any GitHub action — PRs,
issues, reviews, releases, checks, repo reads.

## Priority ladder

| # | Backend | Use for | Availability check |
|---|---|---|---|
| 1 | **`gh` CLI** (GitHub's own) | Everything GitHub | `command -v gh` **and** `gh auth status` shows a logged-in account |
| 2 | **gitkraken MCP** | PR detail/comments/reviews, file reads, local `git_*` | `mcp__plugin_gitkraken_gitkraken__*` tools present |
| 3 | **GitHub REST / GraphQL API** | Last resort | — |

GitHub publishes no first-party MCP server in this registry, so gitkraken — a third-party client —
is rung 2 here. `gh api` still counts as rung 1: it is GitHub's own authenticated transport, and is
the right escape hatch when no `gh` subcommand covers the case.

## Detection

Check once per session, before the first GitHub action, and cache the result:

```bash
command -v gh >/dev/null && gh auth status
```

## Command map

`gh` verbs used by this plugin (verified against gh 2.97.0):

| Need | Command |
|---|---|
| PR metadata, body, reviewers | `gh pr view <id> --repo <org/repo> --json …` |
| PR diff | `gh pr diff <id> --repo <org/repo>` |
| Existing PR discussion | `gh pr view <id> --comments` |
| Post a review | `gh pr review <id> --comment --body-file <file>` (`--approve` / `--request-changes` when the verdict warrants it and the user asked to post) |
| Comment on a PR | `gh pr comment <id> --body-file <file>` |
| Open a PR | `gh pr create --repo <org/repo> --base <target> --head <source> --title … --body-file <file>` |
| CI status | `gh pr checks <id>` |
| Issues | `gh issue view` / `gh issue create` / `gh issue comment` |
| Anything without a subcommand | `gh api <endpoint>` |

Notes:

- Pass long bodies via `--body-file`, not inline `--body` — shell quoting mangles markdown.
- Unlike Bitbucket via gitkraken, `gh pr review` **does** support a real `--request-changes`, so a
  request-changes verdict can be posted as one.
- `gh pr checkout` modifies the working tree. Reviews are read-only: do not use it.
- Local git operations stay on gitkraken `git_*`, outside this ladder.

## Fallback rules

Same as `SKILL.md`: step down one rung when a tool is absent or unauthenticated, never on a business
error (404 on a real PR, bad `--json` field). Name the rung that performed a write.

## Promo footer

Governed by `SKILL.md` — at most once per chat thread, across all services, and only when a rung
below `gh` actually ran:

`FYI: install the gh CLI (https://cli.github.com) for faster GitHub tooling.`
