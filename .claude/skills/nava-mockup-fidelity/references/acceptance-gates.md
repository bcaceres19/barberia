# Acceptance gates for an assigned mockup

Use these tables in the issue, PR, or evidence report. Do not leave a row blank;
use `not represented` or `not applicable` where appropriate.

## Gate A · Visual contract before code

| Region | Reference measurement | Current implementation | Required change |
| --- | --- | --- | --- |
| Canvas / shell | viewport, background, occupied area | measured value | target |
| Brand block | bounding box, alignment, wordmark size | measured value | target |
| Heading block | position, width, font, size, line height | measured value | target |
| Primary content / form | width, height, offset, padding | measured value | target |
| Controls | height, type size, icon box, gaps | measured value | target |
| Primary action | width, height, position, type/icon scale | measured value | target |
| Feedback / states | surface, border, icon, copy hierarchy | measured value | target |

Measurements can come from source pixels, browser CSS pixels, computed styles,
or ratios. Record which unit is being used. For a bitmap without specifications,
measure ratios relative to the source panel instead of guessing from memory.

## Gate B · Objective tolerance at the reference viewport

- Declared palette roles match their normative CSS values exactly.
- Major region bounds and key alignments differ by no more than 4 CSS px or 2%
  of the relevant dimension, whichever is greater.
- Text size and line height differ by no more than 1 CSS px when the source
  value is known; when inferred from a bitmap, the rendered hierarchy and line
  wrapping must match.
- Functional icon visual boxes differ by no more than 2 CSS px when they map
  one-to-one to the reference.
- No unexplained change remains in centering, whitespace distribution, content
  density, border treatment, or primary-action prominence.
- Platform font rasterization, anti-aliasing, and real copy length are not
  failures by themselves; their layout effects still require review.

These tolerances are review aids, not permission to ignore an obvious visual
difference that falls just inside a numeric threshold.

## Gate C · Evidence after implementation

| Evidence | Required result |
| --- | --- |
| Reference panel | original-resolution image and identified viewport/state |
| Baseline | real app before the change at the same viewport |
| Final render | real app after the change at the same viewport |
| Comparison | side-by-side plus overlay or image diff viewed by the implementer |
| Viewport proof | measured `window.innerWidth` and `window.innerHeight` |
| Responsive | 320, 360, 768, 1280 px and 200% zoom |
| Interaction | keyboard, focus, dialog behavior, reduced motion |
| States | applicable loading, empty, error, conflict, success, disabled |
| Runtime | no new console errors or unexpected failed network requests |
| Deviations | difference, authoritative reason, and approved treatment |

## Gate D · Final report

Report:

1. route and exact mockup panel;
2. observable changes to geometry, type, colors, icons, and states;
3. evidence paths for baseline, final, and comparison;
4. measured viewports and tests executed;
5. every remaining deviation and its authority;
6. explicit `PASS` or `FAIL` for mockup fidelity.

`PASS` is invalid while the implementer can still point to an unexplained visible
difference in a primary region.
