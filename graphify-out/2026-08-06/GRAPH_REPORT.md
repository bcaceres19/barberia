# Graph Report - .  (2026-08-06)

## Corpus Check
- cluster-only mode — file stats not available

## Summary
- 200 nodes · 173 edges · 45 communities (42 shown, 3 thin omitted)
- Extraction: 93% EXTRACTED · 7% INFERRED · 0% AMBIGUOUS · INFERRED: 12 edges (avg confidence: 0.8)
- Token cost: 31,880 input · 1,720 output

## Community Hubs (Navigation)
- Frontend Tooling Dependencies
- Node TypeScript Config
- HTTP Middleware Handlers
- OpenAPI Tooling Config
- API Server Bootstrap
- Web Package Manifest
- NPM Build Scripts
- App TypeScript Config
- Vite TypeScript Config
- Vitest TypeScript Config
- Worker Config Loading
- Clock Abstraction
- Prettier Formatting Config
- Vue Router Setup
- Root TypeScript Config
- Barbershop System Root

## God Nodes (most connected - your core abstractions)
1. `compilerOptions` - 15 edges
2. `scripts` - 10 edges
3. `compilerOptions` - 9 edges
4. `run()` - 8 edges
5. `Recover()` - 6 edges
6. `NewLogger()` - 6 edges
7. `Load()` - 5 edges
8. `RequestID()` - 5 edges
9. `New()` - 5 edges
10. `include` - 5 edges

## Surprising Connections (you probably didn't know these)
- `run()` --calls--> `Load()`  [INFERRED]
  apps/api/cmd/api/main.go → apps/api/internal/platform/config/config.go
- `run()` --calls--> `HealthHandler()`  [INFERRED]
  apps/api/cmd/api/main.go → apps/api/internal/platform/httpserver/health.go
- `run()` --calls--> `Recover()`  [INFERRED]
  apps/api/cmd/api/main.go → apps/api/internal/platform/httpserver/recover.go
- `run()` --calls--> `NewLogger()`  [INFERRED]
  apps/api/cmd/worker/main.go → apps/api/internal/platform/observability/logger.go
- `run()` --calls--> `Chain()`  [INFERRED]
  apps/api/cmd/api/main.go → apps/api/internal/platform/httpserver/server.go

## Import Cycles
- None detected.

## Communities (45 total, 3 thin omitted)

### Community 0 - "Frontend Tooling Dependencies"
Cohesion: 0.06
Nodes (31): devDependencies, eslint, eslint-plugin-vue, jsdom, @playwright/test, prettier, @types/node, typescript (+23 more)

### Community 1 - "Node TypeScript Config"
Cohesion: 0.09
Nodes (22): compilerOptions, allowImportingTsExtensions, erasableSyntaxOnly, lib, module, moduleDetection, noEmit, noFallthroughCasesInSwitch (+14 more)

### Community 2 - "HTTP Middleware Handlers"
Cohesion: 0.18
Nodes (11): HealthHandler(), Handler, Logger, Recover(), Handler, newRequestID(), RequestID(), RequestIDFromContext() (+3 more)

### Community 3 - "OpenAPI Tooling Config"
Cohesion: 0.14
Nodes (13): description, devDependencies, @redocly/cli, name, private, scripts, openapi:bundle, openapi:check-config (+5 more)

### Community 4 - "API Server Bootstrap"
Cohesion: 0.22
Nodes (10): main(), run(), Chain(), Handler, New(), Logger, NewLogger(), parseLevel() (+2 more)

### Community 5 - "Web Package Manifest"
Cohesion: 0.20
Nodes (9): dependencies, vue, vue-router, name, private, type, version, vue (+1 more)

### Community 6 - "NPM Build Scripts"
Cohesion: 0.20
Nodes (10): scripts, build, dev, format, format:write, lint, preview, test:e2e (+2 more)

### Community 7 - "App TypeScript Config"
Cohesion: 0.20
Nodes (9): exclude, extends, include, src/**/*.spec.ts, e2e/**/*, src/**/*.ts, src/**/*.tsx, src/**/*.vue (+1 more)

### Community 8 - "Vite TypeScript Config"
Cohesion: 0.20
Nodes (10): compilerOptions, allowArbitraryExtensions, erasableSyntaxOnly, noFallthroughCasesInSwitch, noUnusedLocals, noUnusedParameters, paths, tsBuildInfoFile (+2 more)

### Community 9 - "Vitest TypeScript Config"
Cohesion: 0.20
Nodes (9): compilerOptions, tsBuildInfoFile, types, extends, include, node, src/**/*.spec.ts, jsdom (+1 more)

### Community 10 - "Worker Config Loading"
Cohesion: 0.43
Nodes (5): main(), run(), getEnv(), Load(), Config

### Community 11 - "Clock Abstraction"
Cohesion: 0.40
Nodes (3): Clock, System, Time

### Community 12 - "Prettier Formatting Config"
Cohesion: 0.40
Nodes (4): printWidth, semi, singleQuote, trailingComma

## Knowledge Gaps
- **90 isolated node(s):** `system-barbershop`, `Clock`, `requestIDKey`, `semi`, `singleQuote` (+85 more)
  These have ≤1 connection - possible missing edges or undocumented components.
- **3 thin communities (<3 nodes) omitted from report** — run `graphify query` to explore isolated nodes.

## Suggested Questions
_Questions this graph is uniquely positioned to answer:_

- **Why does `devDependencies` connect `Frontend Tooling Dependencies` to `Web Package Manifest`?**
  _High betweenness centrality (0.052) - this node is a cross-community bridge._
- **Why does `scripts` connect `NPM Build Scripts` to `Web Package Manifest`?**
  _High betweenness centrality (0.021) - this node is a cross-community bridge._
- **Why does `run()` connect `API Server Bootstrap` to `Worker Config Loading`, `HTTP Middleware Handlers`?**
  _High betweenness centrality (0.017) - this node is a cross-community bridge._
- **Are the 6 inferred relationships involving `run()` (e.g. with `Load()` and `HealthHandler()`) actually correct?**
  _`run()` has 6 INFERRED edges - model-reasoned connections that need verification._
- **What connects `system-barbershop`, `Clock`, `requestIDKey` to the rest of the system?**
  _90 weakly-connected nodes found - possible documentation gaps or missing edges._
- **Should `Frontend Tooling Dependencies` be split into smaller, more focused modules?**
  _Cohesion score 0.06451612903225806 - nodes in this community are weakly interconnected._
- **Should `Node TypeScript Config` be split into smaller, more focused modules?**
  _Cohesion score 0.08695652173913043 - nodes in this community are weakly interconnected._