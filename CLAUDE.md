## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).

## Merge de pull requests

- Autorización permanente del propietario (2026-08-13): en cuanto todos los checks de CI de un PR contra `main` estén en verde, haz squash-merge sin pedir confirmación adicional por ese paso. No hace falta preguntar cada vez.
- Si algún check falla o queda pendiente, no mergees: investiga o espera.
- Esta autorización cubre el merge en sí. Sigue pidiendo confirmación aparte para otras acciones destructivas o que afecten estado compartido (force-push, borrar ramas, reescribir `main`, etc.), salvo que el usuario las autorice explícitamente también.

## Prompts persistentes

- Antes de entregar un prompt pensado para otro chat, Claude, Codex o una ejecución futura, aplica `AGENTS.md` y guárdalo en `docs/10-backlog/prompts/` con la plantilla e índice del catálogo.
- Un prompt mutable no se ejecuta con `issue: pending`: permanece `draft` hasta enlazar un issue real. No inventes issue, rama, PR, evidencia ni estado de una HU.
- La respuesta al usuario enlaza el archivo persistido. Si no puedes escribirlo, declara de forma explícita que solo entregas un borrador no guardado.
