# Verificación del kit de pruebas UI — #300

Fecha: 2026-10-07. Rama: test/300-exploracion-ui-luna. Base de producto comprobada: ebac3a7f4dec3e0158446a010ab6ccd4df7ebd8b, main/origin/main al iniciar. Este informe valida preparación, no la campaña completa #301.

## Artefactos

- Skill canónico ui-app-testing, adaptador Claude y autorización explícita DEC-116.
- Catálogo de 17 pantallas, fichas, transversales, recorridos, combinaciones, límites e informe.
- 21 prompts persistentes: 17 pantallas, journeys, orquestación, preparación y evaluación; issues reales #300/#301.
- Aprovisionador local con guard de ambiente/destino/roles, Argon2 del módulo existente y credenciales privadas. Sin dependencias nuevas, permisos nuevos ni cambios de producto/esquema.

## Resultado de preparación

PostgreSQL real, Atlas validado con 26 migraciones aplicadas y ninguna pendiente (versión 20261004120000_add_barbershop_panel_profile). API local :8080, frontend :5173, Chromium de Playwright. Marcadores locales para OTP y recuperación; credenciales de proveedores desactivadas en el proceso QA. No se inició worker.

49 comprobaciones reales PASS en verify-ui-ready.mjs:

| Comprobación | Cantidad | Resultado |
| --- | --- | --- |
| Health API/PostgreSQL | 1 | 200 en ambos destinos |
| Login de cada cuenta activa por interfaz | 20 | Panel/Agenda visible; Barberos 200 |
| Agenda diaria autenticada por tenant | 20 | 200 |
| Credencial válida de cuenta inactiva | 1 | Login 401 |
| Login adicional de staff/aislamiento | 3 | Panel/Agenda visible |
| Crear profesional por UI y recuperar tras recarga/paginación | 1 | Dato persistido y encontrado |
| Aislamiento A/B | 2 | Tenant correcto, recurso cruzado 404 en ambas direcciones |
| Entrada pública | 1 | Navegación y resolución real 200 |

Se crearon 20 tenants/cuentas activas y 1 cuenta inactiva. Baseline por tenant: 2 profesionales, 3 servicios y horarios/asignaciones. El smoke agrega registros QA en staff; no restaura estados de escenarios. Aprovisionamiento inicial y repetición contra BD real comprobados; la repetición reconoce marcador privado y no vuelve a sembrar.

Manifiesto en apps/web/.auth/nava-qa/current.json, fuera de Git, permisos 0600/directorio 0700. Evidencia cruda en apps/web/test-results/ui-qa/<campaña>/preparation/: results.json y capturas staff a 360/1280 px. Ambas inspeccionadas; solo es evidencia de la pantalla de preparación, no conformidad responsive/accesible de toda la app. No se publican credenciales, cookies, OTP ni URLs con tokens.

## Controles y revisión

- node tools/qa/provision-ui.test.mjs: 3 pruebas PASS (destino, identidades/campañas, SQL sin borrados/restauración/password en claro).
- node --check de herramientas y bash -n del wrapper: PASS. Argon2 Go ejercitado por aprovisionamiento real.
- quick_validate.py: skill válido.
- tools/ai/validate-agent-system.sh --strict: 8 skills, 0 fallos, 0 advertencias.
- Enlaces locales y referencias HU/RN/DEC/CA de los nuevos documentos: comprobados. git diff --check: limpio.
- Evaluación independiente mediante spawn_agent(model=gpt-6-luna, reasoning_effort=medium, fork_turns=none). Informe privado skill-evaluation/report.md; señaló falta de instrucciones para seleccionar campaña nueva y reanudar tras interrupción. Ambas incorporadas en protocolo/skill. Su smoke inicial no terminó; el resultado final de 49 PASS corresponde al coordinador tras corregir los selectores/paginación del helper.
- GitHub comprobado con herramienta autenticada: #300/#301 abiertos al preparar entrega.

## Uso y pendientes

Pedir **«Prueba toda la app»**. La skill prepara/verifica el entorno y reparte tareas a Luna medium, máximo dos simultáneos; access/recovery y cierre cruzado seriales. Para usar Luna también como coordinador, seleccionarlo en el cliente: un skill no cambia el modelo del hilo actual.

[Entrada del kit](../../../07-calidad/pruebas-ui/README.md) · [Skill](../../../../.agents/skills/ui-app-testing/SKILL.md) · [Campaña lista #301](https://github.com/bcaceres19/barberia/issues/301).

La campaña completa todavía no se ejecutó: expiración/renovación/revocación de sesiones, todos los estados/combinaciones, viewports, accesibilidad y recorridos cruzados recibirán evidencia y resultado durante #301. La app usa cookie de sesión con token opaco, no JWT; el kit prueba el mecanismo real. Integración de esta rama mediante PR queda fuera de esta preparación local.
