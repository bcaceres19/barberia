---
name: graphify-refresh
description: "Decide whether and how to refresh the local Graphify knowledge graph at the moments the owner defined. Use right before running graphify query/path/explain, when starting a cross-module task that will consult the graph (multi-module HU, refactor, architecture investigation), when closing a block of user stories, or when asked whether the graph is up to date. Do not use after commits, branch switches, docs-only edits, or routine work that will not consult the graph."
---

# Graphify refresh

`graphify-out/` is local, regenerable, and never versioned (`DEC-087`). Graphify is optional (`DEC-086`). Refresh it only when it is about to be used; a fresh graph nobody queries is wasted work.

## 1. Check freshness deterministically

Run from the repository root:

```bash
tools/ai/graphify-freshness.sh
```

It is read-only. Never guess staleness from memory or file dates.

## 2. Act on the recommendation

| `recommendation=` | Action |
|---|---|
| `SKIP` | Use the graph as is. Say nothing unless asked. |
| `UPDATE` | Run `graphify update .` (local, no LLM, no cost), then continue with the query. If it reports fewer nodes after a refactor that deleted code, rerun with `--force`. |
| `SEMANTIC_UPDATE_SUGGESTED` | Do not run it automatically. Tell the user how many documentation files changed since the last semantic extraction and ask whether to run `/graphify . --update` now (incremental, re-extracts only changed files, but uses the LLM, so it costs and is slower). Meanwhile, if `code_changes` is not 0 (`-1` means unknown), run `graphify update .` so code answers stay correct. |
| `BUILD_FIRST` | No graph exists. Ask the user before building it: a first full `/graphify` has LLM cost. If they decline, continue without the graph. |

After a confirmed `/graphify . --update` or full `/graphify` finishes, record its base so the next check measures from it:

```bash
git rev-parse HEAD > graphify-out/.graphify_semantic_commit
```

## 3. Guard against reinstalled hooks

If the script prints `hooks=installed`, tell the user that Graphify's git hooks are back (they rebuild on every commit and branch switch) and offer `graphify hook uninstall`. Do not remove them without consent.

## When not to refresh

- After each commit or branch switch.
- After changes that only touch documentation, evidence images, or lockfiles, unless the documentation threshold is reached.
- During routine work that will not query the graph. Plain text search and module ownership come first (`CLAUDE.md`).

## Tuning

The documentation threshold defaults to 30 changed files since the last semantic extraction. Override it for one run with `GRAPHIFY_DOCS_THRESHOLD=<n> tools/ai/graphify-freshness.sh`. Change the default only by owner request, updating this skill and the script together.
