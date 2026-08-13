# Graph Report - web (2026-08-07)

## Corpus Check

- 40 files · ~9,324 words
- Verdict: corpus is large enough that graph structure adds value.

## Summary

- 221 nodes · 207 edges · 34 communities (31 shown, 3 thin omitted)
- Extraction: 99% EXTRACTED · 1% INFERRED · 0% AMBIGUOUS · INFERRED: 2 edges (avg confidence: 0.5)
- Token cost: 0 input · 0 output

## Community Hubs (Navigation)

- devDependencies
- compilerOptions
- BaseDialog.vue
- scripts
- compilerOptions
- BaseAlert.vue
- BaseInput.vue
- BaseBadge.vue
- tsconfig.vitest.json
- BaseButton.vue
- .prettierrc.json
- apps/web
- router/index.ts
- tsconfig.json
- e2e/README.md

## God Nodes (most connected - your core abstractions)

1. `compilerOptions` - 15 edges
2. `scripts` - 10 edges
3. `compilerOptions` - 9 edges
4. `close()` - 5 edges
5. `emit` - 5 edges
6. `include` - 5 edges
7. `include` - 4 edges
8. `apps/web` - 4 edges
9. `emit` - 3 edges
10. `handleDismiss()` - 3 edges

## Surprising Connections (you probably didn't know these)

- None detected - all connections are within the same source files.

## Import Cycles

- None detected.

## Communities (34 total, 3 thin omitted)

### Community 0 - "devDependencies"

Cohesion: 0.06
Nodes (31): eslint, eslint-plugin-vue, jsdom, devDependencies, eslint, eslint-plugin-vue, jsdom, @playwright/test (+23 more)

### Community 1 - "compilerOptions"

Cohesion: 0.09
Nodes (22): ES2023, eslint.config.js, playwright.config.ts, vite.config.ts, vitest.config.ts, compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly (+14 more)

### Community 2 - "BaseDialog.vue"

Cohesion: 0.12
Nodes (18): classes, close(), closeButtonRef, dialogRef, emit, focusableElementsRef, handleBackdropClick(), handleCloseClick() (+10 more)

### Community 3 - "scripts"

Cohesion: 0.10
Nodes (19): dependencies, vue, vue-router, name, private, scripts, build, dev (+11 more)

### Community 4 - "compilerOptions"

Cohesion: 0.10
Nodes (19): e2e/**/\*, src/**/_.ts, src/\**/_.tsx, src/**/*.vue, vite/client, @vue/tsconfig/tsconfig.dom.json, compilerOptions, allowArbitraryExtensions (+11 more)

### Community 5 - "BaseAlert.vue"

Cohesion: 0.12
Nodes (16): alertRef, borderVar, classes, emit, focusableElementsRef, handleActionClick(), handleDismiss(), handleKeyDown() (+8 more)

### Community 6 - "BaseInput.vue"

Cohesion: 0.15
Nodes (15): classes, describedBy, emit, errorId, handleBlur(), handleChange(), handleFocus(), handleInput() (+7 more)

### Community 7 - "BaseBadge.vue"

Cohesion: 0.18
Nodes (10): borderVar, classes, dotColorVar, dotStyle, emit, handleDismiss(), Props, style (+2 more)

### Community 8 - "tsconfig.vitest.json"

Cohesion: 0.20
Nodes (9): jsdom, ./tsconfig.app.json, compilerOptions, tsBuildInfoFile, types, extends, include, node (+1 more)

### Community 9 - "BaseButton.vue"

Cohesion: 0.25
Nodes (5): buttonRef, classes, emit, handleClick(), Props

### Community 10 - ".prettierrc.json"

Cohesion: 0.40
Nodes (4): printWidth, semi, singleQuote, trailingComma

### Community 11 - "apps/web"

Cohesion: 0.40
Nodes (4): apps/web, Comandos, Estado, Requisitos

## Knowledge Gaps

- **124 isolated node(s):** `semi`, `singleQuote`, `printWidth`, `trailingComma`, `name` (+119 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **3 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions

_Questions this graph is uniquely positioned to answer:_

- **Why does `devDependencies` connect `devDependencies` to `scripts`?**
  _High betweenness centrality (0.042) - this node is a cross-community bridge._
- **What connects `semi`, `singleQuote`, `printWidth` to the rest of the system?**
  _124 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `devDependencies` be split into smaller, more focused modules?**
  _Cohesion score 0.06451612903225806 - nodes in this community are weakly interconnected._
- **Should `compilerOptions` be split into smaller, more focused modules?**
  _Cohesion score 0.08695652173913043 - nodes in this community are weakly interconnected._
- **Should `BaseDialog.vue` be split into smaller, more focused modules?**
  _Cohesion score 0.12380952380952381 - nodes in this community are weakly interconnected._
- **Should `scripts` be split into smaller, more focused modules?**
  _Cohesion score 0.1 - nodes in this community are weakly interconnected._
- **Should `compilerOptions` be split into smaller, more focused modules?**
  _Cohesion score 0.1 - nodes in this community are weakly interconnected._
