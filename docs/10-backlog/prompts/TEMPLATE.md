---
prompt_id: "PROMPT-<TIPO>-<REFERENCIA>-v1"
version: "1.0"
kind: "hu|fix|audit|review|test|docs|ci|ops|chore|spike|orchestration"
status: "draft"
target_agents:
  - "any"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu: []
issue: "pending"
issue_url: null
suggested_issue_title: "<tipo>(<scope>): <resultado>"
branch: null
pr: null
pr_url: null
depends_on: []
rules: []
decisions: []
acceptance_criteria: []
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
created_at: "YYYY-MM-DD"
updated_at: "YYYY-MM-DD"
supersedes: null
superseded_by: null
---

# <Título del prompt>

## Instrucción para el agente

Implementa, revisa o investiga únicamente la preocupación descrita en este archivo. Trabaja de forma autónoma hasta producir la entrega definida, sin ampliar el alcance ni inventar decisiones.

## Objetivo

<Resultado observable que debe existir al finalizar.>

## Preflight obligatorio

1. Comprueba que el árbol esté limpio y no sobrescribas cambios ajenos.
2. Actualiza `main` mediante fast-forward y verifica la dependencia indicada en los metadatos.
3. Ejecuta la consulta Graphify apropiada cuando exista `graphify-out/graph.json`.
4. Lee completamente `AGENTS.md`, `CLAUDE.md` y cada archivo de `source_docs`.
5. Comprueba dudas y contradicciones. Si falta una decisión, registra el bloqueo antes de cambiar código.
6. Localiza el issue real. Si el metadato sigue `pending`, créalo o solicita autorización; actualiza este archivo antes de ejecutar.
7. Crea la rama `<tipo>/<issue>-<descripcion>` desde `main` actualizada.

## Alcance incluido

- <Entrega incluida.>

## Fuera de alcance

- <Preocupación que no debe mezclarse.>

## Estado existente que debe conservarse

- <Código, migraciones, contrato, decisiones o pruebas ya disponibles.>

## Trabajo requerido

1. <Paso verificable.>
2. <Paso verificable.>

## Pruebas y evidencia

- <Prueba asociada a regla, criterio o defecto.>

## Documentación y trazabilidad

- Actualiza la matriz, historial, README, contrato o diccionario que realmente resulte afectado.
- Actualiza los metadatos y el índice de este catálogo con issue, rama, PR y estado reales.

## Verificación final

```text
<Comandos aplicables de formato, lint, tipos, pruebas, build y validación>
```

Entrega una tabla:

`Criterio | Estado | Prueba o evidencia`

No declares cumplido aquello que no esté probado.

## Git y PR

- Commit/PR: `<tipo>(<scope>): <descripción>`.
- Usa `Closes #<issue>` solo si se cubre el issue completo; de lo contrario, `Refs #<issue>`.
- No hagas push directo, force push ni merge de `main`.
