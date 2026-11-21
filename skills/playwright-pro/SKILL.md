---
name: playwright-pro
description: Production-grade Playwright E2E testing — scaffold a suite, generate tests from flows, fix failing/flaky tests, measure coverage, migrate from other frameworks, run and report, and integrate BrowserStack or TestRail. Use for anything involving Playwright tests or browser end-to-end/UI test automation.
---

# playwright-pro

Playwright E2E toolkit. Routes to the sub-skills in `${CLAUDE_PLUGIN_ROOT}/library/playwright/`; read and follow the matching one. Also called by `ship`'s verify phase for UI/E2E gates.

| User need | Sub-skill |
|---|---|
| Set up Playwright in a repo (config, structure) | `init` |
| Generate tests from a user flow / spec | `generate` |
| Fix failing or flaky tests | `fix` |
| Measure / improve test coverage | `coverage` |
| Migrate from Cypress/Selenium/other | `migrate` |
| Run tests and produce a report | `pw` + `report` |
| Review existing Playwright tests for quality | `review` |
| Cross-browser/device via BrowserStack | `browserstack` |
| Sync results to TestRail | `testrail` |

`pw` is the general-purpose entry when the request spans several of the above. Infer the route; don't hand the user the table.

**Tooling** — `${CLAUDE_PLUGIN_ROOT}/library/policies/service-tooling/SKILL.md` applies: vendor-recommended tool first. For Playwright that is `@playwright/test` at the repo's pinned version (never bump a major to make a test pass) and `@playwright/mcp` for browser driving; for BrowserStack, its official SDK/MCP over hand-rolled grid wiring; TestRail has no first-party tool, so its REST API is the correct rung.
