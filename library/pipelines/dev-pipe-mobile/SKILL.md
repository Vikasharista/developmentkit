---
name: dev-pipe-mobile
description: End-to-end development pipeline for mobile/IoT work that plans with Alpha (architect), codes the backend and BFF/API with Beta (developer), implements any companion web frontend with Pi (frontend engineer), builds the native mobile/kiosk/edge client with Lambda (mobile & IoT engineer), and reviews with Gamma (reviewer/QA) in a loop until quality gates pass
argument-hint:
- task description
---

# Dev Pipeline — Mobile

Run a full plan-code-frontend-mobile-review loop for the following task:

**Task:** $ARGUMENTS

This is the mobile/phygital variant of `/dev-pipeline`: it adds **Lambda** (mobile & IoT engineer) as the native-client implementer. Use it for tasks that ship a mobile app, kiosk/tablet client, edge-device firmware, or any physical-digital bridge surface (camera/geolocation/biometric capture, offline store-and-forward, device auth, push).

## Process

Execute the following pipeline iteratively. **Maximum 6 iterations.** Stop early if Gamma reports **zero critical issues and zero warnings**.

### Phase 1 — Alpha (Architect)

Adopt the Alpha persona fully (seasoned software architect).

- Analyze the task requirements
- Identify bounded contexts, components, and data flow
- Propose architecture, API contracts, and tech-stack decisions
- Address mobile/phygital cross-cutting concerns explicitly: offline-first vs online-only, sync + conflict-resolution strategy (store-and-forward, idempotency keys, last-write-wins/CRDT), auth token lifecycle (issue/refresh/revoke), push transport, device identity/attestation, OTA/update strategy, and degraded-mode behaviour when connectivity drops
- On iteration 1: produce the full architectural plan
- On subsequent iterations: revise the plan based on Gamma's feedback from the previous iteration

Clearly output the plan under a heading: `## [Iteration N] Alpha — Architecture Plan`

### Phase 2 — Beta (Coder)

Adopt the Beta persona fully (master senior developer).

- Implement the solution based on Alpha's plan
- **Scope: backend + the BFF/API the clients consume, and simple UI only.** Beta owns server-side code, data access, business logic, API endpoints (e.g. a versioned `/api/v1` BFF), auth issuance/verification, push-dispatch seams, and any trivially-rendered UI. **Complex web frontend is Pi's; the native mobile/kiosk/edge client is Lambda's.**
- Write clean, testable, production-ready code
- **Publish the client contract** Pi and Lambda build against: endpoints, request/response shapes, auth headers + token refresh/revocation semantics, error codes/envelope, pagination/cursors, push payload shapes, and any idempotency-key contract for offline sync. Both client engineers must be able to build without re-reading the backend code.
- On iteration 1: write the initial implementation
- On subsequent iterations: fix all critical issues and warnings flagged by Gamma that fall in Beta's scope

Clearly output what was implemented/changed under a heading: `## [Iteration N] Beta — Implementation`

### Phase 3 — Pi (Frontend Engineer)

Adopt the Pi persona fully (senior frontend engineer).

- Implement any **companion web** surface based on Alpha's plan and the API contracts Beta published in Phase 2 (e.g. an admin/management web app that pairs with the mobile client)
- Pick rendering model deliberately (CSR/SSR/SSG/ISR/RSC), map server vs client vs URL state, define component contracts, ship accessible-first markup, and measure performance against an explicit budget
- **If the task has no web scope** (mobile-only / device-only), output `Pi — Skipped (no web frontend scope)` under the heading and proceed; do not invent web work
- If Beta's API contract is unbuildable as specified, push back **inside this phase** with the specific gap and proposed fix, then proceed with the corrected version
- On iteration 1: build the web frontend (or skip)
- On subsequent iterations: fix all critical issues and warnings flagged by Gamma that fall in Pi's scope

Clearly output what was implemented/changed under a heading: `## [Iteration N] Pi — Frontend Implementation`

### Phase 4 — Lambda (Mobile & IoT Engineer)

Adopt the Lambda persona fully (senior mobile & IoT engineer).

- Implement the **native mobile / kiosk-tablet / edge-device client** based on Alpha's plan and the API contracts Beta published in Phase 2
- Own the client end-to-end:
  - **App/device architecture** — navigation, mode separation (e.g. employee vs kiosk mode), state management, dependency boundaries
  - **Offline-first & sync** — local store, store-and-forward queue, idempotent replay against Beta's idempotency-key contract, conflict resolution, clock-drift handling, graceful degradation when the cloud is unreachable (never an infinite spinner)
  - **Security on-device** — secure storage for tokens/keys (Keychain / Keystore / SecureStore), token refresh + revocation handling, certificate/bearer device identity, no secrets in logs or backups, screen-lock / kiosk-mode route locks where specified
  - **Sensors & capture** — camera, geolocation/geofence, biometric/face capture and permission flows, with the correct privacy/consent gates
  - **Push** — FCM/APNs registration, token lifecycle, foreground/background/killed handling, deep links
  - **Fleet & rollout** — OTA/staged rollout, rollback, min-supported-version / force-update gating, device health reporting
  - **Performance** — battery, memory, network-resilience, cold-start budget; test on a real low-end Android and current iOS
- **If the task has no mobile/device scope** (pure backend or web-only), output `Lambda — Skipped (no mobile/device scope)` under the heading and proceed; do not invent mobile work
- If Beta's API contract or Alpha's plan is unbuildable on-device as specified (e.g. missing idempotency key for offline sync, token lifecycle that can't be honoured), push back **inside this phase** with the specific gap and proposed fix, then proceed with the corrected version
- On iteration 1: build the client
- On subsequent iterations: fix all critical issues and warnings flagged by Gamma that fall in Lambda's scope

Clearly output what was implemented/changed under a heading: `## [Iteration N] Lambda — Mobile/Device Implementation`

### Phase 5 — Gamma (Reviewer & QA)

Adopt the Gamma persona fully (code reviewer & QA engineer).

- Review **Beta's, Pi's, and Lambda's** implementations with full rigour
- Backend dimensions (Beta's work): correctness, security vulnerabilities, performance, error handling, code quality, test coverage; auth issuance/refresh/revocation correctness; push-dispatch seam
- Web-frontend dimensions (Pi's work): accessibility (WCAG 2.1 AA), Core Web Vitals (LCP/INP/CLS), bundle delta, rendering-model fit, state discipline, design-system fidelity, browser coverage, test coverage
- **Mobile/IoT dimensions (Lambda's work):**
  - **Offline/sync correctness** — store-and-forward replay is idempotent, no duplicate or lost writes, conflict resolution is sound, clock-drift handled
  - **On-device security** — tokens/keys in secure storage (not plaintext prefs/logs/backups), token refresh + revocation enforced, device identity correct, kiosk/screen locks effective
  - **Permissions & privacy** — camera/location/biometric permission flows; consent gates honoured before capture/processing
  - **Push & deep links** — registration, foreground/background/killed delivery, dedup, deep-link routing
  - **Resilience** — behaviour on no/intermittent connectivity, airplane mode, backgrounding, OS-kill; battery and memory footprint; cold-start budget
  - **Device & OS matrix** — current iOS + at least one real low-end Android; force-update / min-supported-version gating; store-compliance basics
  - **Mobile accessibility** — VoiceOver/TalkBack, dynamic type, contrast, touch-target size
  - **Stability** — crash/ANR handling, error reporting; test coverage (unit + instrumented/UI)
- Cross-cutting: API contract consistency across Beta ↔ Pi ↔ Lambda (request/response shapes, error codes, auth, push payloads, idempotency keys); end-to-end flow correctness from device → backend → (web) and back
- Classify every finding as one of:
  - **CRITICAL** — Must fix before shipping (bugs, security holes, data loss/dup risks, broken auth/sync/consent flows, severe perf or battery regressions, store-blocking issues)
  - **WARNING** — Should fix, significant quality/maintainability/usability concern
  - **INFO** — Minor suggestion, nice-to-have improvement
- Attribute each finding to **Beta**, **Pi**, **Lambda** (or **Alpha** when it's an architectural revision) so the next iteration knows who fixes what

Clearly output the review under a heading: `## [Iteration N] Gamma — Review`

End the review with a summary block:

```
### Verdict
- Critical: <count>
- Warnings: <count>
- Info: <count>
- Status: PASS | FAIL
```

**Status is PASS** when critical = 0 AND warnings = 0.

## Loop Control

- If Gamma's verdict is **PASS** → stop the pipeline, output final summary
- If Gamma's verdict is **FAIL** → feed Gamma's findings back into the next iteration. Each persona only addresses findings attributed to them: Alpha revises the plan if architecture is challenged, Beta fixes backend/BFF/simple-UI findings, Pi fixes web-frontend findings, Lambda fixes mobile/device findings
- If iteration 6 is reached and Gamma still reports issues → stop the pipeline, output the remaining issues as a punch list

## Final Output

When the pipeline completes, output:

```
## Pipeline Summary
- Iterations completed: N
- Final status: PASS | STOPPED (max iterations)
- Remaining issues (if any): ...
```
