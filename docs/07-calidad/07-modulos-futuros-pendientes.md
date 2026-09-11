---
titulo: "Módulos futuros pendientes de checklist"
version: "1.2"
estado: "Herramienta operativa, no normativa"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-11"
documentos_relacionados:
  - "README.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/estados-citas.md"
---

# Módulos futuros pendientes de checklist

Al 2026-08-25, estas carpetas del código son solo el esqueleto de un módulo (`index.ts` sin páginas ni rutas en el frontend; `doc.go` sin dominio en el backend). **No existe interfaz que explorar todavía.** No se fabrica un checklist contra una pantalla que no existe: eso produciría pasos falsos que nadie puede ejecutar y confusión sobre qué está realmente probado.

`apps/web/src/modules/public-booking` salió de esta tabla el 2026-09-11: `HU-090` le agrega una primera pantalla real (`/reservar/:slug`, entrada pública sin sesión). Un checklist completo de "reserva pública" sigue prematuro -concurrencia, idempotencia de creación, catálogo, selección de barbero y disponibilidad llegan con `HU-091` en adelante, todavía bloqueadas por `DP-PUB-02`–`DP-PUB-06`/`CT-011`-, así que las filas de concurrencia/idempotencia de la lista de abajo esperan a que esas capacidades existan de verdad.

| Módulo | Frontend | Backend | Bloque |
| --- | --- | --- | --- |
| Agenda | `apps/web/src/modules/agenda/index.ts` | `apps/api/internal/modules/booking` (HU-060: núcleo persistente y primitiva transaccional interna; sin endpoint HTTP, sin pantalla) | B2/B3 |
| Horarios y bloqueos | `apps/web/src/modules/schedules/index.ts` | `apps/api/internal/modules/schedule` (solo `doc.go`) | B2 |
| Notificaciones | `apps/web/src/modules/notifications/index.ts` | `apps/api/internal/modules/notification` (senders ya existen, sin flujo de UI ni programación de recordatorios) | B2/B3 |

## Cómo extender esta carpeta cuando cada módulo exista

Cuando un módulo de esta lista tenga una pantalla real, crea `0N-checklist-<modulo>.md` siguiendo exactamente la estructura de [`03`](03-checklist-autenticacion-sesion.md)/[`04`](04-checklist-modulos-catalogo.md): límites reales tomados del código de validación (no inventados), y agrega sus filas correspondientes a [`05-matriz-combinaciones.md`](05-matriz-combinaciones.md). Las reglas de negocio ya confirmadas que ese checklist deberá cubrir, con su bloque más crítico marcado, son:

- **Disponibilidad y rejilla** — `RN-DIS-01` a `RN-DIS-07` (huecos exactos, semiabiertos `[inicio, fin)`, paso de 15 minutos, zona horaria).
- **Concurrencia — el bloque de mayor riesgo del producto** — `RN-CON-01` a `RN-CON-06`: dos reservas simultáneas por la misma franja, un bloqueo creado mientras alguien confirma esa franja, la restricción de exclusión de la base de datos como última línea de defensa. Ningún checklist futuro de agenda/reserva pública está completo sin una fila que dispare `N` solicitudes simultáneas reales sobre la misma franja.
- **Estados y transiciones** — la máquina completa de [`estados-citas.md`](../02-requisitos/estados-citas.md): las 8 transiciones permitidas y las 10 prohibidas (tabla §7), cada una como una fila explícita, no como una frase genérica de "probar transiciones inválidas".
- **Cancelación** — `RN-CAN-01` a `RN-CAN-04`, con el plazo configurable y el caso límite del instante exacto evaluado contra el reloj del servidor, no el del navegador.
- **Bloqueos sobre citas existentes** — `RN-BLQ-03`: crear un bloqueo nunca falla por chocar con citas ya agendadas; verificar que la interfaz nunca cancela automáticamente y siempre deja la decisión explícita al barbero.
- **Recordatorios** — `RN-REC-01` a `RN-REC-06`: reprogramar dos veces seguidas debe dejar exactamente un recordatorio vigente, nunca tres; cancelar una cita no debe enviar su recordatorio ya programado.
- **Idempotencia de creación de cita** — `RN-IDE-01` aplicado a la reserva pública: doble toque en "Confirmar turno" con pérdida de señal debe dejar una sola cita.

Hasta entonces, cualquier exploración de estas áreas se limita a leer las reglas y anotar preguntas abiertas en `docs/00-control/dudas-pendientes.md`, nunca a "probar" una pantalla que no existe.
