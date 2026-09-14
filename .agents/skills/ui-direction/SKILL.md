---
name: ui-direction
description: "Establish a deliberate visual and interaction direction before implementing a significant new or redesigned interface. Use for greenfield screens, broad redesigns, weak or generic visual requests, or when the user asks for creative directions. Do not use for tiny styling fixes, faithful implementation of an assigned exact mockup, or backend-only work."
---

# UI Direction

Act as a product-minded creative director before acting as an implementer. The goal is a defensible visual target, not decorative novelty.

## Grounding

Read only the relevant sections of:

- `docs/01-producto/alcance-mvp.md` for the product boundary;
- `docs/03-desarrollo/estandar-diseno-visual.md` for NAVA identity, tokens, freedom, and fidelity rules;
- `docs/03-desarrollo/especificacion-frontend-nava.md` for the affected route, components, states, voice, and responsive behavior;
- the exact source image or Figma frame when one is assigned.

A mockup may constrain appearance but never creates functionality.

## Choose the mode first

- **Identity-guided:** no exact target is assigned. Explore composition freely within NAVA / Tailored Grid.
- **Fidelity:** an exact mockup, screenshot, or frame is assigned. Do not invent alternate directions for the represented viewport and state; extract a measurable visual contract and route implementation and verification to `visual-qa`.

## Explore meaningful directions

In identity-guided mode, propose two directions by default and three only for high-impact or highly ambiguous surfaces. Each direction must differ in composition or interaction model, not merely color.

For each direction state:

- product idea and user benefit;
- personality and editorial metaphor;
- information hierarchy and dominant composition;
- typography roles, spatial rhythm, density, and depth;
- interaction and motion principles, including reduced motion;
- mobile-to-desktop transformation;
- strengths, risks, and what makes it recognizably NAVA.

Reject directions that rely without product reason on generic SaaS dashboards, arbitrary purple gradients, card grids for every section, oversized radii, glassmorphism, placeholder iconography, or interchangeable sidebars.

## Select and specify

Recommend one direction against user outcome, scope, accessibility, implementation cost, and consistency with adjacent screens. Do not ask the user to choose when the issue or an assigned reference already determines the answer.

Create a compact visual contract:

- primary hierarchy and layout landmarks;
- token families to reuse, without copying the whole design system;
- component variants and required states;
- responsive transformations at 320, 360, 768, and 1280 px;
- keyboard, focus, contrast, zoom, and motion expectations;
- reference artifacts and acceptance evidence.

If visual exploration would materially improve the decision, use the `generacion-mockups-nava` skill to create an atlas. Do not generate raster assets when prose and a layout sketch are sufficient.
