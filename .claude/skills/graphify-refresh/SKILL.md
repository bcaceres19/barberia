---
name: graphify-refresh
description: "Decide whether and how to refresh the local Graphify knowledge graph at the moments the owner defined. Use right before running graphify query/path/explain, when starting a cross-module task that will consult the graph (multi-module HU, refactor, architecture investigation), when closing a block of user stories, or when asked whether the graph is up to date. Do not use after commits, branch switches, docs-only edits, or routine work that will not consult the graph."
---

# graphify-refresh · adaptador de Claude Code

Este archivo solo expone el skill a Claude Code. Su fuente canónica, compartida con Codex, es `.agents/skills/graphify-refresh/SKILL.md`.

1. Lee completo `.agents/skills/graphify-refresh/SKILL.md` antes de actuar y síguelo como si fuera este archivo.
2. Resuelve sus enlaces relativos (`references/...`) desde `.agents/skills/graphify-refresh/` y cárgalos solo cuando el skill lo indique.
3. No edites el contenido aquí: cambia el canónico y ejecuta `tools/ai/validate-agent-system.sh`.
