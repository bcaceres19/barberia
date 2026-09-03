---
prompt_id: "PROMPT-FIX-188-ACCESO-RECUPERACION-FIDELIDAD-v1"
version: "1.0"
kind: "fix"
status: "in_progress"
target_agents:
  - "claude"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: "HU-010"
related_hu:
  - "HU-005"
  - "HU-007"
  - "HU-008"
  - "HU-009"
  - "HU-011"
issue: 188
issue_url: "https://github.com/bcaceres19/barberia/issues/188"
suggested_issue_title: "chore(web): rediseñar acceso y recuperación Tailored Grid"
branch: "fix/188-acceso-recuperacion-fidelidad"
pr: null
pr_url: null
depends_on:
  - "El atlas del issue #183 está integrado en `main`."
  - "El protocolo de fidelidad visual del issue #208 y PR #209 está integrado en `main`."
  - "El rediseño inicial de acceso del issue #206 y PR #207 está integrado en `main`."
rules:
  - "RN-DAT-02"
decisions:
  - "DEC-055"
  - "DEC-056"
  - "DEC-058"
  - "DEC-062"
  - "DEC-065"
  - "DEC-077"
  - "DEC-078"
  - "DEC-079"
  - "DEC-080"
acceptance_criteria:
  - "CA-188-FID-01: `/acceso` reproduce el panel de escritorio de 1440 px y el panel móvil de 360 px de `02-acceso-recuperacion.png` dentro de las tolerancias de la skill de fidelidad."
  - "CA-188-FID-02: el bloque NAVA, la regla completa de latón, el título, el formulario, los controles, los iconos y el CTA conservan escala, proporción, centrado y jerarquía del mockup."
  - "CA-188-FID-03: los colores computados usan exactamente la paleta normativa NAVA; no quedan aproximaciones visibles del tinta, marfil, blanco, piedra, grafito, salvia o latón."
  - "CA-188-FID-04: `/recuperar-acceso` reproduce solicitud, verificación, nueva contraseña y estados representados sin cambiar el contrato, la seguridad ni la lógica OTP."
  - "CA-188-FID-05: existen baseline, render final, lado a lado y overlay/diff vistos por el implementador para escritorio y móvil, junto con prueba del viewport efectivo."
  - "CA-188-FID-06: acceso y recuperación pasan teclado, foco, zoom 200 %, movimiento reducido, estados aplicables, consola, red, pruebas de componente, axe, E2E afectadas y build."
  - "CA-188-FID-07: Claude itera y autocorrige hasta `PASS`; no delega la comparación visual al propietario ni declara terminado por haber cambiado CSS o screenshots."
source_docs:
  - "AGENTS.md"
  - "CLAUDE.md"
  - "CONTRIBUTING.md"
  - ".claude/skills/nava-mockup-fidelity/SKILL.md"
  - ".claude/skills/nava-mockup-fidelity/references/acceptance-gates.md"
  - ".claude/skills/browser-viewport-verification/SKILL.md"
  - "docs/00-control/registro-decisiones.md"
  - "docs/00-control/contradicciones.md"
  - "docs/00-control/dudas-pendientes.md"
  - "docs/01-producto/alcance-mvp.md"
  - "docs/01-producto/reglas-negocio.md"
  - "docs/02-requisitos/historias-usuario.md"
  - "docs/03-desarrollo/estandar-diseno-visual.md"
  - "docs/03-desarrollo/especificacion-frontend-nava.md"
  - "docs/03-desarrollo/estandar-frontend-vue.md"
  - "docs/03-desarrollo/estrategia-pruebas.md"
  - "docs/03-desarrollo/flujo-git-github.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/README.md"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/00-referencia-tailored-grid.png"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/02-acceso-recuperacion.png"
  - "docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/14-componentes-formularios-alertas.png"
  - "docs/10-backlog/prompts/orchestration/redisenio-integral-nava-tailored-grid-v3.md"
  - "apps/web/package.json"
  - "apps/web/src/app/router/index.ts"
  - "apps/web/src/modules/auth"
  - "apps/web/src/shared/ui"
created_at: "2026-09-03"
updated_at: "2026-09-03"
supersedes: null
superseded_by: null
---

# Terminar acceso y recuperación con fidelidad al mockup NAVA

## Instrucción para Claude Code

Continúa de forma autónoma el rediseño visual de `/acceso` y `/recuperar-acceso` hasta que ambos flujos pasen la puerta de fidelidad. No esperes que el propietario compare la pantalla, enumere diferencias o te diga qué corregir. Tú debes abrir el mockup a resolución original, medirlo, abrir la app real, generar las comparaciones, mirarlas, identificar las diferencias y autocorregir el código hasta `PASS`.

Este trabajo es netamente visual. Conserva comportamiento, copy normativo, contratos, rutas, sesión, rate limits, privacidad, seguridad y lógica OTP. No conviertas los borradores de canal/destino OTP que existen en el árbol en decisiones de producto y no los implementes dentro de este issue.

## Modelo y skills obligatorias

Usa Claude Sonnet 5 con esfuerzo `high`. Comprueba el modelo efectivo con `/status`. Antes de editar y nuevamente antes de declarar terminado:

1. carga `.claude/skills/nava-mockup-fidelity/SKILL.md`;
2. completa `.claude/skills/nava-mockup-fidelity/references/acceptance-gates.md`;
3. carga `.claude/skills/browser-viewport-verification/SKILL.md`;
4. trabaja en modo **mockup fidelity**, nunca como inspiración libre.

## Preflight obligatorio

1. Lee completos `AGENTS.md`, `CLAUDE.md`, este prompt y todos los `source_docs` textuales.
2. Verifica que estás en `fix/188-acceso-recuperacion-fidelidad`, que el árbol no contiene cambios ajenos nuevos y que la rama incluye `origin/main` con los PR #207 y #209.
3. No descartes los cambios ya presentes: trátalos como un candidato que aún debe demostrarse visualmente, no como trabajo aprobado.
4. Ejecuta Graphify si el repositorio lo exige y localiza componentes, estilos, pruebas y E2E afectados.
5. Ejecuta baseline de formato, lint, tipos, unitarias de auth y build. Registra por separado cualquier fallo preexistente.
6. Arranca el stack con los scripts oficiales. Usa Chrome DevTools sobre la app real, no una maqueta HTML ni una captura aislada.
7. Usa únicamente la cuenta QA que el propietario haya indicado para esta ejecución. Escribe las credenciales solo en la UI; nunca las guardes en archivos, prompts, comandos, logs, capturas o reportes.

## Referencia vinculante

La referencia exacta es `docs/10-backlog/evidence/ui-mockups-nava-tailored-grid-2026-09-02/02-acceso-recuperacion.png`:

- panel **ACCESO — ESCRITORIO (1440px)** para `/acceso`;
- panel **ACCESO — MÓVIL (360px)** para `/acceso`;
- los tres paneles **RECUPERAR ACCESO — MÓVIL (360px)** para solicitud, verificación y nueva contraseña;
- la franja **ESTADOS CLAVE** para error de correo, error de contraseña, foco y éxito.

Usa `00-referencia-tailored-grid.png` para el lenguaje maestro y `14-componentes-formularios-alertas.png` para controles y alertas. Si se contradicen, el panel exacto de acceso manda para geometría y la documentación normativa manda para seguridad, accesibilidad y comportamiento.

## Contrato cromático exacto

Verifica en estilos computados, no solo en variables declaradas:

- tinta `#101B2B`;
- marfil `#F4F0E7`;
- blanco `#FFFFFF`;
- grafito `#2A2D32` y secundario `#5E625F`;
- piedra `#E8E2D8` y borde `#C9C0B2`;
- salvia `#748477`;
- latón `#B8955A` y latón oscuro `#765C2F`;
- semánticos de foco, información, éxito, advertencia y error definidos por el estándar.

No aceptes un color “parecido”. Comprueba también que las fuentes editoriales/funcionales realmente carguen, que los pesos usados existan y que no haya un fallback silencioso.

## Qué debes corregir expresamente

Mide en la imagen y en el navegador antes de fijar cifras CSS. Como mínimo resuelve:

1. En escritorio, el conjunto de acceso debe quedar centrado en el viewport y ocupar una proporción equivalente al panel de referencia; no puede verse como una tarjeta pequeña flotando dentro de demasiado espacio.
2. La mitad tinta y la mitad marfil deben conservar la relación del mockup y una altura visual dominante equivalente.
3. `NAVA` debe tener la escala editorial grande del panel, estar centrado en su región y conservar aire suficiente.
4. La regla de latón debe abarcar visualmente el ancho completo del wordmark, con el rombo centrado; no puede cubrir solo una fracción de la palabra.
5. `Accede a NAVA` debe quedar centrado respecto al formulario, con tamaño, peso, interlineado y separación equivalentes al mockup.
6. El formulario no debe quedar subdimensionado: ancho, alto, labels, texto, iconos, campos, botón y espacios deben guardar la proporción del panel.
7. Correo, contraseña, ojo, candado y sobre deben conservar cajas visuales y alineación consistentes; debe existir un único control accesible para mostrar/ocultar contraseña.
8. En móvil, la cabecera tinta, el wordmark, el formulario y el CTA deben reproducir el panel de 360 px sin aplicar a ciegas la composición de escritorio.
9. Recuperación debe cubrir los tres pasos y sus estados, incluido código de seis dígitos, foco, errores, reenvío, éxito y nueva contraseña, sin inventar funcionalidad.
10. Autofill, loading, disabled, error, challenge por rate limit, sesión expirada y navegación de retorno no deben romper la geometría ni la jerarquía.

## Ciclo autónomo obligatorio

Repite este ciclo hasta aprobar, sin pedir confirmación por decisiones rutinarias de layout:

1. abre el PNG original y registra el contrato de Gate A con píxeles o proporciones;
2. captura baseline de la app real en 1440 px para escritorio y 360 px para móvil, demostrando `window.innerWidth`/`innerHeight`;
3. implementa una corrección coherente de la región completa;
4. captura el render final en exactamente el mismo viewport y estado;
5. genera lado a lado y overlay/diff entre referencia y app;
6. **abre y mira** ambos artefactos; una imagen generada pero no inspeccionada no cuenta;
7. registra cada diferencia primaria de geometría, color, tipografía, alineación, escala, iconos o estado;
8. autocorrige y vuelve al paso 4;
9. solo cuando no haya diferencias primarias inexplicadas, ejecuta la matriz responsive y funcional completa.

No declares `PASS` porque tests, tokens o screenshots se actualizaron. No le preguntes al propietario “¿así está bien?”: la referencia y las puertas medibles son el criterio. Solo detente ante un bloqueo real de credenciales/autorización, contradicción normativa que cambie producto o infraestructura externa reproduciblemente caída.

## Responsive, accesibilidad y runtime

Después de aprobar los viewports representados, verifica 320, 360, 768, 1280 y 1440 px, además de zoom 200 %:

- sin overflow horizontal, recortes ni controles solapados;
- objetivo táctil, labels, nombres accesibles y mensajes asociados;
- recorrido completo por teclado, foco visible y orden lógico;
- restauración/movimiento de foco en pasos y errores;
- `prefers-reduced-motion`;
- normal, foco, error, loading, disabled y demás estados aplicables;
- consola sin errores nuevos y Network sin solicitudes inesperadas del rediseño.

Chrome DevTools es obligatorio para interacción, estilos computados, accesibilidad, consola y red. Si el navegador no produce un viewport exacto, usa Playwright para la captura determinista, pero no sustituyas la inspección manual de Chrome.

## Pruebas y evidencia

Conserva evidencias sin datos personales en una estructura que distinga `baseline`, `final`, `side-by-side` y `overlay` por ruta, estado y viewport. No sobrescribas el baseline con el resultado final.

Ejecuta como mínimo:

```text
pnpm --filter @system-barbershop/web format
pnpm --filter @system-barbershop/web lint
pnpm --filter @system-barbershop/web typecheck
pnpm --filter @system-barbershop/web test:unit
pnpm --filter @system-barbershop/web build
pnpm --filter @system-barbershop/web test:e2e -- <specs de acceso y recuperación>
git diff --check
```

No reduzcas umbrales, no ocultes flakiness y no actualices evidencia sin verla. Corrige las pruebas afectadas por cambios legítimos de semántica o controles, conservando la intención funcional.

## Fuera de alcance

- backend, OpenAPI, migraciones, contratos, payloads o reglas de autenticación;
- resolver los borradores OTP de canal único o destino verificado;
- rediseñar agenda, shell u otras pantallas;
- instalar Tailwind o una biblioteca sin justificación completa;
- mezclar un fix funcional no bloqueante: regístralo en otro issue.

## Entrega

Actualiza este prompt y el catálogo con estado, commit y PR reales. Entrega una tabla:

`Ruta/estado | Mockup/panel | Viewport medido | Baseline | Final | Lado a lado | Overlay/diff | Pruebas | Desviaciones | PASS/FAIL`

Commit/PR: `fix(web): completa fidelidad visual de acceso y recuperación`.

Usa `Closes #188` solo si todos los criterios del issue quedan cubiertos; en caso contrario usa `Refs #188`. No integres con diferencias primarias pendientes ni checks rojos.

## Texto corto para iniciar Claude

```text
Lee y ejecuta de principio a fin `docs/10-backlog/prompts/fix/issue-188-acceso-recuperacion-fidelidad-v1.md`. Usa Claude Sonnet 5 con esfuerzo high y carga obligatoriamente las skills `nava-mockup-fidelity` y `browser-viewport-verification`. Continúa sobre la rama actual, abre el mockup original y la app real con Chrome DevTools, mide y compara 1440/360, genera y mira baseline/final/lado-a-lado/overlay, autocorrige hasta PASS y no me delegues la revisión visual. Conserva estrictamente la funcionalidad y no te detengas por decisiones rutinarias de CSS.
```
