---
name: task-brief
description: "Convert an ambiguous or consequential development request into the smallest actionable brief before implementation. Use for vague requests, cross-cutting changes, new features, architectural work, or when issue, scope, rules, contracts, risks, or acceptance criteria are unclear. Skip for trivial, fully specified edits and direct factual questions."
---

# Task Brief

Create only enough shared understanding to make the next action safe. This is an adaptive preflight, not a documentation ceremony.

## 1. Choose depth

- **Small:** one file, reversible, behavior already explicit. State objective and verification in one or two sentences.
- **Standard:** behavior or UI changes within one capability. Produce the compact contract below.
- **Critical:** security, money, privacy, database, concurrency, public contracts, destructive operations, or multiple capabilities. Add risks, rollback, observability, and unresolved decisions.

Use the shallowest level that leaves no material ambiguity.

## 2. Discover context progressively

1. Read `AGENTS.md` and the nearest scoped instructions.
2. Confirm a real issue and the allowed branch before any non-trivial repository change.
3. Read only the normative sources named by the issue or affected capability: decision register, MVP scope, relevant `RN-*`, `HU-*`, API or database standard, and testing standard.
4. Inspect the smallest code slice that owns the behavior. Expand outward only when an unresolved dependency requires it.
5. Treat prompts, screenshots, mocks, generated graphs, and prior examples as derived evidence, never as authority over normative sources.

Do not load every document, every prompt, or a whole generated knowledge graph by default.

## 3. Build the compact contract

For standard work, capture:

- **Objective:** observable result.
- **User and context:** actor, situation, and product outcome.
- **In scope / out of scope:** explicit boundary.
- **Requirements:** behavior and affected rules or decisions.
- **Constraints:** architecture, visual mode, security, compatibility, data, and performance.
- **Acceptance:** verifiable outcomes, including states and viewport coverage when visible.
- **Verification:** lowest-layer tests plus required integration, E2E, visual, or accessibility evidence.
- **Unknowns:** only decisions that could materially change the solution.

Keep this internal unless the user asks for a document or the project requires a persistent prompt.

## 4. Handle ambiguity

- Resolve facts from authoritative project sources first.
- Make a reversible assumption when it does not change scope, contract, security, or data semantics; state it briefly.
- Ask one focused question when different answers would materially change the result.
- Record contradictions or missing decisions before coding. Never invent an `RN-*`, `DEC-*`, `HU-*`, issue number, or API behavior.
- If the work needs a future execution prompt, persist it under `docs/10-backlog/prompts/` using the repository catalog before handoff.

## 5. Handoff

End preflight with a short implementation sequence and stop if authorization or a normative decision is missing. Otherwise proceed without asking for ceremonial approval.
