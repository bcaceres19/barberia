---
name: visual-qa
description: "Verify a rendered NAVA interface with screenshots and interaction evidence, in identity-guided mode or against an assigned exact mockup, screenshot, or Figma frame. Use after visible UI changes, for critique grounded in the running app, for responsive/viewport checks (320/360/768/1280 px, resize_window), and before editing and before declaring done whenever a screen must match a reference. Do not use for code-only review or speculative critique without a renderable target."
---

# Visual QA

Review what is rendered, not what the code appears to intend. This skill replaces the former Claude-only `nava-mockup-fidelity` and `browser-viewport-verification` skills; their gates live here and in `references/`.

## 1. Select the conformance mode

Read `docs/03-desarrollo/estandar-diseno-visual.md` and decide before judging:

- **Identity-guided:** no exact image is assigned. Composition stays free within NAVA / Tailored Grid. Judge product clarity, NAVA identity, composition, hierarchy, consistency, responsive behavior, and interaction quality.
- **Fidelity:** an issue, prompt, or user message assigns a specific image to the screen/component, or asks to match, reproduce, implement, or correct against it. The image is a visual contract for the viewport and state it represents.

Do not downgrade fidelity to "inspiration" because the code already uses NAVA colors or because another composition is easier to implement.

## 2. Evidence gate

- Require a renderable implementation.
- Prove the effective viewport: after any resize, measure `window.innerWidth` and `window.innerHeight`. A successful resize response is not proof. Follow [references/viewport-verification.md](references/viewport-verification.md) whenever browser automation or responsive checks are involved.
- In fidelity mode, also require the original reference at full resolution and match viewport, route, state, theme, content, crop, scale, and device density.

If a required artifact cannot be opened or the states do not match, report `blocked` with the missing evidence. Never claim fidelity from code, tokens, DOM metrics, computed styles, unit tests, or memory alone; those are supporting evidence only.

## 3. Fidelity workflow (before editing and before declaring done)

1. Read `AGENTS.md`, the issue, the assigned prompt, the visual standard, and the affected product rules.
2. Open the source image at original resolution and identify the exact panel, viewport, and state that map to the route. Do not judge from a thumbnail.
3. Inventory the current route and capture a baseline from the real app. Keep existing behavior, copy, security, API contracts, and state transitions.
4. Complete Gate A of [references/fidelity-gates.md](references/fidelity-gates.md): geometry, type, colors, icons, spacing, states, and justified differences.
5. Implement the whole represented region. A token swap, isolated wordmark change, or screenshot refresh is not a redesign.
6. Render at the exact reference viewport (verified as in section 2).
7. Compare reference and implementation side by side and with an overlay or image diff, and view the comparison yourself.
8. Iterate until every unexplained difference in the contract table is closed, within the Gate B tolerances.
9. Then validate the responsive matrix, zoom, keyboard, focus, reduced motion, applicable async states, console, and network.
10. Report with Gates C and D. Do not declare the screen complete or ready to merge while an unexplained visible difference remains in a primary region.

At the represented viewport and state, the implementation must preserve the reference's canvas and brand colors, region proportions and occupied area, centering, alignment, spacing rhythm, density, type hierarchy (approximate face, weight, size, line height), control/icon/border/illustration scale, visual priority, contrast, and state treatment. Equivalent rendered geometry is the goal, not identical CSS coordinates.

**Allowed deviations** only for real content, an authoritative product rule, accessibility, responsive reflow, privacy, security, or browser rendering; log the reason and treatment. A mockup never creates data, an API, business behavior, or a navigation destination. Implementation technique stays free under repository standards; dependencies are never installed automatically.

**The fidelity gate fails** when the brand block, form, main action, or primary content is visibly mis-sized, off-center, or differently proportioned; declared NAVA colors differ without justification; typography or icon scale changes the hierarchy; the check relied only on DOM numbers, tests, or an unviewed automated screenshot; only one viewport or the happy path was reviewed; or evidence hides overflow, personal data, errors, or a known discrepancy.

## 4. Review loop

1. Capture the relevant state at the verified viewport and inspect the saved screenshot.
2. In fidelity mode, build the side-by-side plus overlay or diff.
3. Report findings by severity; fix P0–P2 only when the user asked for implementation.
4. Re-render the same state and compare again.

Default to two correction passes; allow a third when a remaining P1/P2 is clearly converging. Stop and explain the constraint rather than looping indefinitely.

## 5. Inspection matrix

Always check:

- task clarity, hierarchy, alignment, balance, density, spacing, and overflow;
- typography family, weight, scale, line height, wrapping, and truncation;
- semantic colors, contrast, borders, radii, shadows, icons, imagery, and asset quality;
- loading, empty, error, success, conflict, disabled, hover, focus, active, and validation states when affected;
- keyboard order, visible focus, accessible names, reduced motion, and zoom at 200%;
- 320, 360, 768, and 1280 px when composition changes, with mobile and desktop evidence at minimum for any visible change;
- console errors, unexpected failed network requests, and the affected component/Playwright tests.

Do not infer full accessibility compliance from screenshots; combine visual inspection with automated (axe) and keyboard checks.

## 6. Findings

Lead with defects, not praise. For each finding give severity, location, visible evidence, user impact, and a concrete correction.

- **P0:** core task impossible, destructive mistake, or severe accessibility failure.
- **P1:** major visual/interaction mismatch or likely user failure.
- **P2:** responsive drift, inconsistent state, or material polish problem.
- **P3:** non-blocking refinement.

Finish with `passed` (or `PASS` for fidelity) only when no actionable P0–P2 remains. Store evidence where the issue or prompt defines it, normally under `apps/web/e2e/evidence/`; do not invent a new root-level report convention.
