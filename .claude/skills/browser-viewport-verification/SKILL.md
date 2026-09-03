---
name: browser-viewport-verification
description: "Use whenever verifying responsive/mobile/tablet layout with claude-in-chrome browser automation (resize_window, device emulation, or any responsive QA pass) — before trusting a resize result or marking a breakpoint as checked. Trigger words: viewport, responsive, breakpoint, resize_window, mobile/tablet verification, 320/360/768/1280px checks."
---

# Viewport verification

Never trust the browser resize tool response by itself.

After requesting a viewport/window resize, verify the actual rendered
viewport dimensions using JavaScript:

```js
window.innerWidth
window.innerHeight
```

The viewport is considered successfully changed ONLY if the measured
dimensions approximately match the requested dimensions.

If `resize_window` reports success but the viewport dimensions did not
change:

1. Do not claim the responsive layout was tested.
2. Try browser/device emulation if available.
3. Use an alternative viewport mechanism if available.
4. Otherwise report that the requested viewport could not be verified.

Never mark mobile/tablet responsive verification as passed without
checking the actual viewport dimensions.

## Known working alternative in this repo

`resize_window` has been observed to report success while leaving
`window.innerWidth` at the OS-maximized width (verified on two separate
tabs). When that happens, use the repository's own Playwright
device-emulation harness instead — it sets the viewport via CDP, not an
OS window resize, so it is unaffected by this failure mode:

- `apps/web/e2e/*-evidencia-responsiva.spec.ts` already fixes the exact
  breakpoints this project requires: `{320, 360, 768, 1280}` plus
  `1280-zoom200` (a `640x450` viewport approximating 200% text zoom).
- Reuse that same `viewports` array/pattern for any new route's
  responsive evidence instead of inventing another set of breakpoints.
- Chrome DevTools interactive verification (clicks, keyboard, focus,
  console, network, computed styles) still runs at whatever the real
  effective viewport is; only the breakpoint-accurate screenshots need
  to come from Playwright when `resize_window` is confirmed broken.
