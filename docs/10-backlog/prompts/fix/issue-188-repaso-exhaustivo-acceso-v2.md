---
prompt_id: "PROMPT-FIX-188-REPASO-EXHAUSTIVO-ACCESO-v2"
version: "2.0"
kind: "fix"
status: "ready"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-010"
related_hu:
  - "HU-005"
  - "HU-007"
  - "HU-009"
issue: 188
issue_url: "https://github.com/bcaceres19/barberia/issues/188"
suggested_issue_title: "chore(web): rediseñar acceso y recuperación Tailored Grid"
branch: "fix/188-acceso-recuperacion-fidelidad"
pr: 210
pr_url: "https://github.com/bcaceres19/barberia/pull/210"
depends_on:
  - "PR #209 y `DEC-080` integrados en `main`."
  - "La primera ejecución de fidelidad está conservada como `PROMPT-FIX-188-ACCESO-RECUPERACION-FIDELIDAD-v1`."
rules:
  - "RN-DAT-02"
decisions:
  - "DEC-055"
  - "DEC-056"
  - "DEC-058"
  - "DEC-062"
  - "DEC-077"
  - "DEC-079"
  - "DEC-080"
acceptance_criteria:
  - "CA-188-REVIEW2-01: se revoca el PASS anterior de `/acceso` y se realiza un inventario nuevo de todos los elementos visibles, no solo de las diferencias indicadas por el propietario."
  - "CA-188-REVIEW2-02: `/acceso` coincide con los paneles de escritorio 1440 px y móvil 360 px de `02-acceso-recuperacion.png` en geometría, colores, tipografía, escalas, bordes, iconos, espaciado, alineación y estados."
  - "CA-188-REVIEW2-03: las etiquetas Correo y Contraseña no muestran asteriscos; ambos controles conservan `required`, `aria-required`, mensajes de validación y semántica accesible."
  - "CA-188-REVIEW2-04: existe un único control de visibilidad de contraseña con nombre accesible inequívoco, y las E2E de acceso no presentan selectores ambiguos."
  - "CA-188-REVIEW2-05: Claude genera, abre e inspecciona baseline v2, final v2, lado a lado y overlay/diff en 1440 y 360 px; itera hasta cerrar toda diferencia primaria."
  - "CA-188-REVIEW2-06: 320, 360, 768, 1280, 1440 y zoom 200 %, teclado, foco, estados, consola, red, pruebas de componente, axe, E2E de acceso y build pasan antes de actualizar el PR."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - ".claude/skills/nava-mockup-fidelity/SKILL.md"
  - ".claude/skills/nava-mockup-fidelity/references/acceptance-gates.md"
  - ".claude/skills/browser-viewport-verification/SKILL.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/02-acceso-recuperacion.png"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/14-componentes-formularios-alertas.png"
  - "docs/10-backlog/prompts/fix/issue-188-acceso-recuperacion-fidelidad-v1.md"
  - "apps/web/src/modules/auth/components/AuthSplitLayout.vue"
  - "apps/web/src/modules/auth/components/LoginForm.vue"
  - "apps/web/src/modules/auth/pages/LoginPage.vue"
  - "apps/web/src/shared/ui/BaseInput.vue"
created_at: "2026-09-03"
updated_at: "2026-09-03"
supersedes: "PROMPT-FIX-188-ACCESO-RECUPERACION-FIDELIDAD-v1"
superseded_by: null
---

# Segunda pasada exhaustiva de fidelidad para `/acceso`

## Instrucción para Claude Code

Reabre el trabajo visual de `/acceso` en el PR #210. El `PASS` de la ejecución anterior queda revocado: todavía existe al menos una diferencia visible que el informe justificó incorrectamente —los asteriscos junto a `Correo` y `Contraseña`— y eso demuestra que la revisión no inventarió todos los elementos de la pantalla.

No corrijas únicamente los asteriscos. Realiza desde cero una segunda pasada completa de la pantalla, usando el mockup como contrato visual y el estado actual como candidato no aprobado. No esperes más observaciones del propietario y no le pidas que encuentre las diferencias por ti.

## Alcance exacto

- Ruta: `/acceso`.
- Referencia de escritorio: panel **ACCESO — ESCRITORIO (1440px)** de `02-acceso-recuperacion.png`.
- Referencia móvil: panel **ACCESO — MÓVIL (360px)** del mismo archivo.
- Estados: normal, foco, validación de correo, validación de contraseña, envío/loading, disabled, error de servidor, sesión expirada y reto tras rate limit cuando sean alcanzables.
- Archivos de acceso y primitivas visuales estrictamente necesarias.

`/recuperar-acceso`, backend, OpenAPI, OTP, sesión y otras pantallas quedan fuera salvo una regresión causada directamente por una primitiva compartida modificada en este repaso.

## Reglas de ejecución

1. Usa Claude Sonnet 5 con esfuerzo `high` y comprueba `/status`.
2. Carga antes de editar y antes de cerrar `nava-mockup-fidelity` y `browser-viewport-verification`.
3. Abre el PNG a resolución original; no trabajes desde memoria, miniatura ni el informe anterior.
4. Abre la app real con Chrome DevTools. Comprueba `window.innerWidth`/`innerHeight` y estilos computados.
5. Compara todo el diff de la rama/PR contra `origin/main`. Un defecto introducido por un commit anterior de esta misma rama sigue siendo responsabilidad del PR; no puede descartarse como “preexistente”.
6. No uses el reporte v1 como prueba de conformidad. Conserva sus artefactos como historial y genera evidencia v2 separada.
7. No declares una desviación por accesibilidad sin citar la exigencia concreta y demostrar que la alternativa fiel no puede conservar la semántica. Accesibilidad no significa copiar un asterisco visual cuando el mockup no lo contiene.

## Inventario obligatorio antes de editar

Completa una fila con `mockup | app actual | diferencia | corrección | evidencia` para cada elemento, aunque parezca correcto:

- canvas exterior, frame, alto, ancho, borde, radio y sombra;
- proporción entre panel tinta y panel marfil;
- wordmark NAVA: caja, tamaño, peso, centrado y tracking;
- regla de latón: ancho total, grosor, posición y rombo central;
- tagline: ancho, saltos de línea, tamaño, interlineado y posición;
- título `Accede a NAVA`: caja, centrado, tipografía y espacios;
- ancho y posición del formulario;
- labels, indicadores de requerido y separación label/control;
- inputs: ancho, alto, padding, fondo, borde, radio y estados;
- sobre, candado y ojo: caja visual, trazo, escala y alineación;
- texto/placeholder y alineación vertical;
- CTA: ancho, alto, color, tipografía, radio y espacios;
- enlace de recuperación: tamaño, color, subrayado y posición;
- alertas aplicables: caja, icono, copy, cierre y separación;
- distribución total del espacio vertical y horizontal;
- foco visible, hover, autofill, loading, disabled y errores;
- ausencia de overflow, recorte, salto o desplazamiento inesperado.

No se permite escribir “equivalente” o “ya cumple” sin una medición o comparación visible asociada.

## Corrección obligatoria del requerido

El mockup de acceso no muestra `*` junto a `Correo` ni `Contraseña`. Elimínalos visualmente en esta pantalla.

Conserva:

- el atributo HTML `required`;
- `aria-required="true"` cuando corresponda;
- asociación correcta entre `label` e `input`;
- mensajes explícitos como “El correo es obligatorio” después de validación;
- una forma accesible no visual de anunciar “obligatorio” si la evaluación accesible demuestra que hace falta.

No elimines indiscriminadamente el indicador de todas las pantallas. Si `BaseInput` lo dibuja globalmente, crea una API de presentación explícita y probada, con el comportamiento existente como valor por defecto, y úsala solo en `/acceso` o en las pantallas cuyo mockup lo exija.

Agrega una prueba que confirme simultáneamente que en `/acceso` no existe un asterisco visible y que ambos inputs siguen siendo requeridos semánticamente.

## Control de contraseña y E2E

Resuelve el selector ambiguo informado en la ejecución v1. Debe existir un solo input de contraseña y un solo botón de visibilidad, con nombres accesibles claros y pruebas estables. No aceptes 23 E2E fallidas como “preexistentes” si el origen está en archivos o commits contenidos por el PR #210.

Corrige el producto o el locator según la causa real, sin degradar el nombre accesible. Vuelve a ejecutar los specs de acceso hasta que pasen o demuestra un bloqueo externo reproducible que no pertenezca al branch diff.

## Ciclo visual obligatorio

En 1440 y 360 px:

1. captura baseline v2;
2. implementa el inventario completo;
3. captura final v2 en el mismo viewport/estado;
4. genera lado a lado y overlay/diff;
5. abre y mira ambos;
6. anota todas las diferencias restantes;
7. autocorrige y repite desde el paso 3.

Después valida 320, 360, 768, 1280, 1440 y zoom 200 %. `PASS` es inválido si queda una diferencia visible sin resolver, si el implementador no abrió las comparaciones o si delega la inspección al propietario.

## Verificación final

```text
pnpm --filter @system-barbershop/web format
pnpm --filter @system-barbershop/web lint
pnpm --filter @system-barbershop/web typecheck
pnpm --filter @system-barbershop/web test:unit
pnpm --filter @system-barbershop/web build
pnpm --filter @system-barbershop/web test:e2e -- <specs de acceso>
git diff --check
```

Revisa además teclado, foco, axe, autofill, consola y Network. No reduzcas umbrales, no ocultes fallos y no sustituyas evidencia visual con números DOM.

## Entrega

Actualiza este prompt a `executed`, el catálogo y el PR #210 con resultados reales. Entrega:

`Elemento inventariado | Diferencia inicial | Corrección | Evidencia v2 | Estado`

y luego:

`Viewport/estado | Baseline v2 | Final v2 | Lado a lado | Overlay/diff | Pruebas | PASS/FAIL`

No uses `PASS` si alguna fila del inventario carece de evidencia o conserva una desviación no autorizada.

## Texto corto para reactivar a Claude

```text
El PASS anterior de `/acceso` queda revocado. Lee y ejecuta completo `docs/10-backlog/prompts/fix/issue-188-repaso-exhaustivo-acceso-v2.md` sobre la rama y PR actuales. La referencia vinculante de esta pantalla es `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/02-acceso-recuperacion.png`: usa exclusivamente el panel `ACCESO — ESCRITORIO (1440px)` para escritorio y el panel `ACCESO — MÓVIL (360px)` para móvil; no uses los paneles de recuperación ni otra lámina para decidir la composición de `/acceso`. Abre el PNG a resolución original e identifica físicamente esos dos paneles antes de editar. No corrijas solo el asterisco: vuelve a inventariar y verifica todos los elementos visibles contra esos paneles, elimina los `*` visuales conservando `required`/`aria-required` y validación, resuelve el locator ambiguo de contraseña, genera y abre evidencia v2 baseline/final/lado-a-lado/overlay y autocorrige hasta PASS sin pedirme que encuentre más diferencias.
```
