---
name: nava-mockup-fidelity
description: "Use whenever implementing, redesigning, or correcting a NAVA screen, flow, or component from an explicit mockup, reference PNG, screenshot, or instruction to match a visual reference. Required before editing and before declaring the UI complete. Enforces measured geometry, colors, typography, alignment, browser comparison, responsive states, and visual evidence. Do not use for UI work without an assigned exact reference or for read-only design critique."
---

# NAVA mockup fidelity

Use this skill together with `browser-viewport-verification` whenever an exact
mockup or screenshot is assigned to the implementation.

## Select the conformance mode

- **Mockup fidelity:** an issue, prompt, or user message assigns a specific
  image to the screen/component or asks to match, reproduce, correct against,
  or implement it. The image is a visual contract for the viewport and state
  it represents.
- **Guided identity:** no exact image is assigned. Composition remains free
  within NAVA / Tailored Grid and the normative frontend documents.

Do not downgrade mockup fidelity to “inspiration” because the code already uses
NAVA colors or because another composition is easier to implement.

## Required workflow in mockup-fidelity mode

1. Read `AGENTS.md`, `CLAUDE.md`, the issue, the assigned prompt,
   `docs/03-desarrollo/estandar-diseno-visual.md`, and the affected product
   rules before editing.
2. Open the source image at original resolution. Identify the exact panel,
   viewport, and state that map to the route. Do not judge from a thumbnail.
3. Inventory the current route and capture a baseline from the real app. Keep
   existing behavior, copy, security, API contracts, and state transitions.
4. Complete the contract and evidence tables in
   [references/acceptance-gates.md](references/acceptance-gates.md). Record
   geometry, type, colors, icons, spacing, states, and justified differences.
5. Implement the whole represented region. A token swap, isolated wordmark
   change, or screenshot refresh is not a redesign.
6. Render the app at the exact reference viewport. Verify `window.innerWidth`
   and `window.innerHeight`; use the repository Playwright harness when browser
   resize is unreliable.
7. Compare reference and implementation side by side and with an overlay or
   image diff. Inspect computed styles as supporting evidence, not as a
   substitute for visual comparison.
8. Iterate until every unexplained difference in the contract table is closed.
9. Then validate 320, 360, 768, and 1280 px, 200% zoom, keyboard, focus,
   reduced motion, applicable async states, console, and network.
10. Do not declare the screen complete or ready to merge until the evidence
    gate passes and all remaining deviations are documented and authorized.

## What must match

At the represented viewport and state, preserve the reference's visible:

- canvas and brand colors;
- proportions and occupied area of major regions;
- centering, alignment, spacing rhythm, and content density;
- type hierarchy, approximate face/weight, size, and line height;
- control, icon, border, and illustration scale;
- visual priority, contrast, and state treatment.

Aim for the tolerances defined in the acceptance-gates reference. Exact CSS
coordinates are not the goal; equivalent rendered geometry is.

## Allowed deviations

Deviate only when required by real content, an authoritative product rule,
accessibility, responsive reflow, privacy, security, or browser rendering. Log
the reason and the resulting treatment. A mockup never creates data, an API,
business behavior, or a navigation destination.

Implementation details remain free: CSS, scoped CSS, CSS Modules, utility-first,
Tailwind, or a visual library may be used under the repository standards.
Tailwind and other dependencies are not installed automatically.

## Failure conditions

The gate fails when any of the following is true:

- the brand block, form, main action, or primary content is visibly too small,
  too large, off-center, or differently proportioned;
- declared NAVA colors differ in computed CSS without justification;
- typography or icon scale changes the hierarchy;
- the implementation was checked only through DOM numbers, unit tests, or an
  automated screenshot without viewing the comparison;
- only one viewport or the happy path was reviewed;
- the evidence hides overflow, personal data, errors, or a known discrepancy.
