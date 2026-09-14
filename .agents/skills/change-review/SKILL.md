---
name: change-review
description: "Independently review a proposed or implemented repository change for correctness, architecture, tests, security, contracts, and frontend quality. Use for code review, pre-PR review, architecture review, regression analysis, or self-critique after implementation. Read-only by default; do not implement fixes unless explicitly asked."
---

# Change Review

Review the change against its issue and authoritative project sources. A diff without intent is not enough; intent without a diff is not proof.

## Scope the review

1. Read the issue or persisted prompt, acceptance criteria, relevant `RN-*`/`DEC-*`/`HU-*`, and repository instructions.
2. Inspect the diff, neighboring ownership boundaries, and tests. Ignore unrelated dirty worktree changes.
3. Select only the applicable lenses below.

## Lenses

- **Behavior:** acceptance criteria, edge cases, errors, idempotency, concurrency, time zones, and backwards compatibility.
- **Architecture:** dependency direction, capability ownership, domain isolation from Chi/PostgreSQL, typed Vue API boundary, cohesion, and unnecessary dependencies.
- **Database/API:** normalized schema, tenant isolation, RLS, migration safety, Atlas integrity, OpenAPI-first consistency, generated client impact, and recovery.
- **Frontend:** semantic HTML, state ownership, component boundaries, responsiveness, accessibility, performance, and consistency with the selected NAVA mode.
- **Security/privacy:** authorization, input/output exposure, secrets, personal data, logs, fixtures, dependency scripts, and destructive paths.
- **Tests/evidence:** changed rule tested at the lowest adequate layer; PostgreSQL/tenant, component, E2E, visual, and accessibility coverage when applicable.
- **Delivery:** issue/branch/commit/PR traceability, focused change, docs/contracts synchronized, and no unrelated generated artifacts.

## Findings standard

Report only actionable findings supported by a file, line, behavior, command result, or missing required evidence.

For each finding include:

- severity (`P0` critical, `P1` high, `P2` medium, `P3` low);
- location and evidence;
- violated requirement or plausible failure scenario;
- smallest corrective action;
- test that would prevent recurrence.

Distinguish verified defects from questions and residual risks. Avoid generic approval language, speculative rewrites, style-only preferences, or demanding tests unrelated to the changed behavior.

If there are no findings, say so and list remaining evidence gaps. Do not modify files, resolve conversations, approve, merge, or mark a prompt complete unless the user requested that action.
