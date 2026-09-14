# Effective viewport verification

This reference proves the effective viewport; it does not prove visual fidelity.

Never trust a browser resize tool response by itself. After requesting a
viewport/window resize, measure the rendered viewport with JavaScript:

```js
window.innerWidth
window.innerHeight
```

The viewport counts as changed only if the measured dimensions approximately
match the requested ones. If a resize reports success but the dimensions did
not change:

1. Do not claim the responsive layout was tested.
2. Try browser/device emulation if available.
3. Use an alternative viewport mechanism if available.
4. Otherwise report that the requested viewport could not be verified.

Never mark mobile/tablet verification as passed without checking the actual
dimensions.

## Known working alternative in this repository

Claude-in-Chrome `resize_window` has been observed to report success while
leaving `window.innerWidth` at the OS-maximized width. When that happens, use
the repository's Playwright device-emulation harness instead. It sets the
viewport through CDP rather than an OS window resize, so it is unaffected:

- `apps/web/e2e/*-evidencia-responsiva.spec.ts` already fixes the required
  breakpoints `{320, 360, 768, 1280}` plus `1280-zoom200` (a `640x450`
  viewport approximating 200% text zoom).
- Reuse that `viewports` array/pattern for any new route's responsive evidence
  instead of inventing another set of breakpoints.
- Interactive verification (clicks, keyboard, focus, console, network,
  computed styles) can still run at the real effective viewport; only the
  breakpoint-accurate screenshots need to come from Playwright when the
  resize is confirmed broken.
