---
titulo: "Plan de bloques de construcción del MVP"
version: "1.28"
estado: "Propuesta"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-09-08"
documentos_relacionados:
  - "../01-producto/alcance-mvp.md"
  - "../01-producto/prioridades.md"
  - "../01-producto/reglas-negocio.md"
  - "../02-requisitos/historias-usuario.md"
  - "../00-control/matriz-trazabilidad.md"
  - "../00-control/contradicciones.md"
  - "prompts-implementacion.md"
---

# Plan de bloques de construcción del MVP

> **Estado del contenido: Propuesta.** Este documento **no crea alcance**. Ordena las 45 funciones P0 ya confirmadas en [alcance-mvp.md](../01-producto/alcance-mvp.md) y define en qué secuencia se construyen. Ninguna función entra, sale ni cambia de prioridad aquí; hacerlo exigiría un `DEC-*` conforme a [prioridades.md](../01-producto/prioridades.md), sección 5.

---

## 1. Para qué sirve este documento

El alcance dice **qué** se construye y las prioridades dicen **por qué**. Faltaba decir **en qué orden**, porque 45 funciones P0 no se pueden abordar en paralelo por una sola persona y algunas dependen físicamente de otras: no existe agenda sin restricción de exclusión, y no existe restricción de exclusión sin tabla de citas, y no existe tabla de citas sin aislamiento por barbería.

Un bloque es un conjunto de historias que se entrega completo y verificable antes de empezar el siguiente. Cada bloque tiene un **criterio de salida comprobable**: si no se cumple, el bloque no está terminado y el siguiente no empieza.

---

## 2. Criterio de ordenamiento

Los bloques se ordenaron con tres reglas, en este orden de peso:

1. **Lo que es caro de corregir después va primero.** Aislamiento entre barberías, restricción de cruces, historial inmutable e idempotencia no se pueden "agregar al final" sin reescribir migraciones ya aplicadas (una migración aplicada es inmutable, `DEC-036`).
2. **Lo que otras cosas necesitan para existir va antes que ellas.** Nadie puede configurar servicios sin poder iniciar sesión; nadie puede calcular disponibilidad sin horario laboral.
3. **Lo que el piloto mide va antes que lo que el piloto no mide.** Los criterios de inicio del piloto (`alcance-mvp.md`, sección 8) son la lista de verificación final de este plan.

Consecuencia deliberada: **el flujo público de reserva, que es la cara visible del producto, no es lo primero que se construye.** Se apoya en disponibilidad, agenda, horario y catálogo; adelantarlo obligaría a simular esos datos y a rehacer el trabajo.

---

## 3. Mapa de bloques

| Bloque | Nombre | Funciones P0 cubiertas | Rango de historias | Estado |
| --- | --- | --- | --- | --- |
| **B0** | Cimientos, seguridad y primeras pantallas | 8 | `HU-001` – `HU-012` | Redactado |
| **B1** | Identidad de la barbería y catálogo | 4 | `HU-020` – `HU-0xx` | `HU-020`–`HU-024` integradas en `main` ([PR #84](https://github.com/bcaceres19/barberia/pull/84), issue real `#77`) |
| **B2** | Horario laboral y bloqueos | 2 | `HU-040` – `HU-042` | `CT-008` resuelta (`DEC-070`); `HU-040` integrada en `main` ([PR #93](https://github.com/bcaceres19/barberia/pull/93), issue real [#90](https://github.com/bcaceres19/barberia/issues/90) abierto por `CA-040-08` parcial); `HU-041` integrada en `main` ([PR #96](https://github.com/bcaceres19/barberia/pull/96), issue real [#95](https://github.com/bcaceres19/barberia/issues/95) abierto por `CA-041-08` parcial); `HU-042` integrada en `main` ([PR #99](https://github.com/bcaceres19/barberia/pull/99), issue real [#98](https://github.com/bcaceres19/barberia/issues/98) abierto por seguimiento de UI/E2E, issue [#100](https://github.com/bcaceres19/barberia/issues/100)) |
| **B3** | Agenda del barbero, estados e integridad | 11 | `HU-060` – `HU-0xx` | `HU-060`–`HU-067` integradas ([PR #235](https://github.com/bcaceres19/barberia/pull/235), [#225](https://github.com/bcaceres19/barberia/issues/225); [PR #237](https://github.com/bcaceres19/barberia/pull/237), [#226](https://github.com/bcaceres19/barberia/issues/226)); `HU-068` corrección T8 ([#227](https://github.com/bcaceres19/barberia/issues/227)) redactada y bloqueada por secuencia |
| **B4** | Reserva pública y disponibilidad | 13 | `HU-090` – `HU-1xx` | Pendiente de redacción |
| **B5** | Notificaciones y recordatorios | 3 | `HU-130` – `HU-1xx` | Pendiente de redacción |
| **B6** | Operación, privacidad y preparación del piloto | 4 | `HU-150` – `HU-1xx` | Pendiente de redacción |

Los rangos se reservan por bloque para que los códigos `HU-*` no se renumeren cuando un bloque crezca (convención de [glosario.md](../00-control/glosario.md), sección 8). Un rango sin usar no es una deuda: es espacio reservado.

---

## 4. Detalle de cada bloque

### B0 · Cimientos, seguridad y primeras pantallas

**Objetivo:** que exista un sistema en el que un barbero real pueda entrar, que ninguna barbería pueda ver datos de otra, que los errores se comporten igual en todas partes y que las pantallas siguientes se construyan sobre componentes ya definidos.

**Funciones:** `F-SEG-01`, `F-SEG-03`, `F-AUTH-01`, `F-AUTH-02`, `F-OPS-01`, `F-OPS-02`, `F-OPS-03`, `F-OPS-04`.

**Incluye:** esquema inicial con `barbershop_id` y RLS activa; rol de aplicación sin `BYPASSRLS`; contexto de barbería por transacción; Chi v5 y el orden de middleware; errores RFC 9457; registro técnico sin datos personales; idempotencia reutilizable; inicio y cierre de sesión; defensa escalonada contra abuso; recuperación de acceso; sistema visual aplicado en componentes base; pantallas de acceso, recuperación y cascarón del panel.

**No incluye:** ninguna entidad de negocio (servicios, horarios, citas, bloqueos). Si una historia de B0 necesita una tabla de negocio para demostrarse, la historia está mal delimitada.

**Criterio de salida:**

- [ ] Un usuario de la barbería A recibe `404` al pedir cualquier recurso de la barbería B, verificado endpoint por endpoint de los que existan.
- [ ] Una consulta ejecutada con el rol de aplicación y el contexto de A no devuelve filas de B, verificado también con SQL directo.
- [ ] El barbero inicia sesión, la sesión sobrevive al cierre del navegador y el cierre de sesión la invalida de inmediato.
- [ ] Superado el umbral de intentos por IP, el formulario exige verificación telefónica.
- [ ] El barbero recupera el acceso con un código enviado a su teléfono verificado.
- [ ] Todo error responde con el formato uniforme del contrato y ningún registro técnico contiene nombre, teléfono, correo ni contraseña.
- [ ] Repetir una operación crítica con la misma clave no produce un segundo efecto.
- [ ] Las pantallas de acceso, recuperación y panel se ven correctas en 320, 360, 768 y 1280 px y son operables con teclado.

**Depende de:** nada. Es el punto de partida.

---

### B1 · Identidad de la barbería y catálogo

`HU-020`, `HU-021`, `HU-022`, `HU-023` y `HU-024` están integradas en `main`. `DP-SER-01`–`DP-SER-03` quedaron resueltas por `DEC-067`–`DEC-069` y cada HU tiene issue real (`#75`, `#76`, `#77`); `HU-024` (ciclo de vida del servicio, cierre de B1) está verificada contra PostgreSQL real, con [PR #84](https://github.com/bcaceres19/barberia/pull/84) integrado contra el issue real `#77`.

**Objetivo:** que la barbería tenga nombre, zona horaria, contacto, barberos y servicios con duración y precio.

**Funciones:** `F-CONF-01`, `F-CONF-02`, `F-SERV-01`, `F-SERV-02`.

**Reglas que gobiernan el bloque:** `RN-SER-01` a `RN-SER-04`, `RN-DIS-07`, `DEC-019`.

**Criterio de salida:**

- [ ] La zona horaria de la barbería se almacena como zona IANA y toda hora mostrada la respeta.
- [ ] Una barbería con un barbero y otra con tres se configuran sin casos especiales en el código.
- [ ] Desactivar un servicio con citas futuras advierte cuántas hay y no cancela nada por su cuenta *(la parte de citas se cierra en B3; en B1 se construye la advertencia sobre el catálogo)*.
- [ ] Cambiar duración o precio no altera ninguna cita existente sin decisión explícita.

**Depende de:** B0.

---

### B2 · Horario laboral y bloqueos

**Objetivo:** que el sistema sepa cuándo trabaja cada barbero y cuándo no.

**Historias propuestas:** `HU-040` horario laboral recurrente → `HU-041` excepciones de jornada y festivos → `HU-042` bloqueos de agenda. Cada una tiene un prompt individual en `docs/10-backlog/prompts/hu/`. `CT-008` se resolvió como `DEC-070` (FK en `ON DELETE RESTRICT`); `HU-040` está integrada en `main` ([PR #93](https://github.com/bcaceres19/barberia/pull/93)), con `CA-040-08` parcial (falta evidencia responsive real) y el issue real [#90](https://github.com/bcaceres19/barberia/issues/90) abierto por ese pendiente. `HU-041` está integrada en `main` ([PR #96](https://github.com/bcaceres19/barberia/pull/96)), con `CA-041-08` parcial por el mismo motivo que `CA-040-08` y el issue real [#95](https://github.com/bcaceres19/barberia/issues/95) abierto por ese pendiente. `HU-042` está integrada en `main` ([PR #99](https://github.com/bcaceres19/barberia/pull/99)), con el backend completo probado contra PostgreSQL real y el frontend cubriendo bloqueos puntuales y series `weekly`; el issue real [#98](https://github.com/bcaceres19/barberia/issues/98) sigue abierto por el seguimiento de UI de `date_list`/excepciones/edición de serie y evidencia E2E/responsive, detallado en el issue [#100](https://github.com/bcaceres19/barberia/issues/100) (mismo criterio que `CA-040-08`/`CA-041-08`).

**Funciones:** `F-HOR-01`, `F-HOR-02`.

**Reglas que gobiernan el bloque:** `RN-BLQ-01` a `RN-BLQ-04`, `RN-DIS-05`, `RN-DIS-07`, `DEC-020`.

**Criterio de salida:**

- [ ] Los siete tipos de bloqueo (`break`, `lunch`, `unavailable`, `day_off`, `holiday`, `vacation`, `emergency`) se crean y restan tiempo.
- [ ] Recurrencias, listas explícitas de fechas y excepciones individuales funcionan.
- [ ] El calendario colombiano de festivos se activa por barbero y un festivo puede abrirse manualmente.
- [ ] Un bloqueo eliminado deja de restar disponibilidad y conserva su registro.
- [ ] Un tramo que cruza la medianoche cabe completo en el horario configurado.

**Depende de:** B1 (los bloqueos son por barbero).

---

### B3 · Agenda del barbero, estados e integridad

**Objetivo:** el corazón del producto. Que exista la cita, que no se pueda cruzar, que el barbero opere su día y que todo cambio quede registrado.

**Funciones:** `F-CITA-01` a `F-CITA-06`, `F-EST-01` a `F-EST-04`, `F-DISP-02`.

**Reglas que gobiernan el bloque:** `RN-CIT-01` a `RN-CIT-05`, `RN-CON-01`, `RN-CON-03`, `RN-CAN-03`, `RN-CAN-04`, `RN-HIS-01`, `RN-HIS-02`, `RN-RES-01` a `RN-RES-03`, y la máquina completa de [estados-citas.md](../02-requisitos/estados-citas.md).

**Secuencia redactada:** `HU-060` núcleo persistente y exclusión PostgreSQL → `HU-061` creación manual → `HU-062` agenda diaria de hoy → `HU-063` navegación por fecha → `HU-064` detalle e historial → `HU-065` reprogramación auditada T2 → `HU-066` cancelación por barbero T6 → `HU-067` cierre manual T4/T7 → `HU-068` corrección auditada T8. `HU-060`–`HU-067` están integradas en `main` ([PR #235](https://github.com/bcaceres19/barberia/pull/235), [PR #237](https://github.com/bcaceres19/barberia/pull/237)). El tercer lote usa únicamente semánticas confirmadas: `HU-068` ([#227](https://github.com/bcaceres19/barberia/issues/227)) espera la integración de ambas. T3 y cierre automático permanecen pendientes porque requieren un lote posterior que precise, respectivamente, snapshots/precio/duración y configuración de modo/demora.

**Criterio de salida:**

- [ ] La restricción de exclusión existe y se verificó **intentando insertar dos citas cruzadas directamente en la base de datos**, saltándose la aplicación.
- [ ] Las siete transiciones permitidas funcionan y las diez prohibidas se rechazan con el código correcto.
- [ ] El historial no se puede editar ni borrar; una corrección agrega una entrada nueva.
- [ ] La agenda diaria abre en hoy, navega por fecha y opera con una sola mano en un celular de gama media.

**Depende de:** B2.

---

### B4 · Reserva pública y disponibilidad

**Objetivo:** que el cliente reserve solo, sin cuenta y sin equivocarse de horario.

**Funciones:** `F-PUB-01` a `F-PUB-08`, `F-DISP-01`, `F-DISP-03`, `F-DISP-04`, `F-DISP-05`, `F-CITA-07`.

**Reglas que gobiernan el bloque:** `RN-DIS-01` a `RN-DIS-06`, `RN-CON-02`, `RN-CON-04` a `RN-CON-06`, `RN-CNF-01`, `RN-CNF-02`, `RN-CAN-01`, `RN-CAN-02`, `RN-DAT-01`, `RN-IDE-01`.

**Criterio de salida:**

- [ ] N solicitudes simultáneas sobre la misma franja producen exactamente una cita; las demás reciben un mensaje comprensible con alternativas y **sin perder los datos escritos**.
- [ ] Solo se ofrecen franjas donde cabe el servicio completo, con los once factores de `RN-DIS-02` aplicados.
- [ ] Doble toque en "Confirmar turno" no crea dos citas.
- [ ] El enlace del turno permite consultar y cancelar según la política vigente de la barbería.

**Depende de:** B3.

---

### B5 · Notificaciones y recordatorios

**Objetivo:** que el cliente no olvide su turno y que ningún mensaje diga una hora vieja.

**Funciones:** `F-NOT-01`, `F-NOT-02`, `F-NOT-03`.

**Reglas que gobiernan el bloque:** `RN-REC-01` a `RN-REC-06`, `RN-DAT-02`, `RN-IDE-01`, `DEC-027`, `DEC-032`.

**Criterio de salida:**

- [ ] La programación del recordatorio nace en la misma transacción que la cita.
- [ ] Dos reprogramaciones seguidas dejan exactamente un recordatorio vigente.
- [ ] El contenido del mensaje se construye en el instante del envío, nunca al programarlo.
- [ ] Cada intento de envío deja evidencia con su resultado, sin teléfono ni correo en claro.
- [ ] Caer el trabajador y reiniciarlo no duplica ni pierde envíos.

**Depende de:** B4. Requiere además la verificación previa de proveedores oficiales exigida por `DEC-027`.

---

### B6 · Operación, privacidad y preparación del piloto

**Objetivo:** que el sistema se pueda operar, restaurar y auditar, y que el piloto pueda arrancar legalmente.

**Funciones:** `F-SEG-02`, `F-OPS-05`, `F-OPS-06`, `F-OPS-09`.

**Reglas que gobiernan el bloque:** `RN-DAT-01` a `RN-DAT-03`, `DEC-025`, `DEC-030`, `DEC-031`.

**Criterio de salida:**

- [ ] Copias de seguridad diarias con retención de 30 días, funcionando.
- [ ] **Una restauración completa probada al menos una vez**, con evidencia y fecha.
- [ ] Anonimización de datos vencidos sin destruir cita, intervalo, servicio, estado ni historial.
- [ ] Procedimiento de incidentes escrito y ensayado.
- [ ] Política de datos, aviso y acuerdo breve redactados (`DEC-030`).

**Depende de:** B5. Es el último bloque antes de la lista de verificación del piloto.

---

## 5. Reglas de secuencia

1. **Un bloque no empieza si el anterior no cumplió su criterio de salida.** Las casillas anteriores se marcan con evidencia, no con opinión.
2. **Ninguna historia salta de bloque para "adelantar" trabajo.** Si aparece esa necesidad, es señal de que una dependencia estaba mal identificada; se corrige el plan, no se improvisa.
3. **Toda historia entra por su issue y su rama**, conforme a [flujo-git-github.md](../03-desarrollo/flujo-git-github.md).
4. **Una historia que necesita una decisión inexistente no se implementa.** Se registra la duda en [dudas-pendientes.md](../00-control/dudas-pendientes.md) y se espera el `DEC-*`.
5. **Cada historia terminada actualiza la matriz de trazabilidad** en el mismo cambio.

---

## 6. Cobertura de las 45 funciones P0

Tabla de control: ninguna función P0 puede quedar sin bloque.

| Bloque | Funciones |
| --- | --- |
| B0 | `F-AUTH-01`, `F-AUTH-02`, `F-SEG-01`, `F-SEG-03`, `F-OPS-01`, `F-OPS-02`, `F-OPS-03`, `F-OPS-04` |
| B1 | `F-CONF-01`, `F-CONF-02`, `F-SERV-01`, `F-SERV-02` |
| B2 | `F-HOR-01`, `F-HOR-02` |
| B3 | `F-CITA-01`, `F-CITA-02`, `F-CITA-03`, `F-CITA-04`, `F-CITA-05`, `F-CITA-06`, `F-EST-01`, `F-EST-02`, `F-EST-03`, `F-EST-04`, `F-DISP-02` |
| B4 | `F-PUB-01`, `F-PUB-02`, `F-PUB-03`, `F-PUB-04`, `F-PUB-05`, `F-PUB-06`, `F-PUB-07`, `F-PUB-08`, `F-DISP-01`, `F-DISP-03`, `F-DISP-04`, `F-DISP-05`, `F-CITA-07` |
| B5 | `F-NOT-01`, `F-NOT-02`, `F-NOT-03` |
| B6 | `F-SEG-02`, `F-OPS-05`, `F-OPS-06`, `F-OPS-09` |

**Total: 8 + 4 + 2 + 11 + 13 + 3 + 4 = 45.** Coincide con el alcance confirmado.

> `F-SEG-02` aparece en B6 porque su parte pesada es la anonimización, pero su exigencia de que los registros técnicos no contengan datos personales (`RN-DAT-02`) se implementa desde `HU-003` en B0 y se verifica en cada bloque posterior.

---

## 7. Qué queda fuera de este plan

- Las funciones P1 y P2 no tienen bloque asignado. Se abordan después del piloto o cuando el propietario decida, según [prioridades.md](../01-producto/prioridades.md).
- Este plan no fija fechas. Una estimación temporal sin equipo definido y sin una sola historia terminada sería una cifra inventada.
- Este plan no autoriza ninguna dependencia, proveedor ni gasto nuevo.

---

## 8. Estado y siguiente paso

| Elemento | Estado |
| --- | --- |
| División en bloques | Propuesta; requiere aprobación del propietario |
| Historias de B0 | Redactadas en [historias-usuario.md](../02-requisitos/historias-usuario.md) |
| Prompts de implementación de B0 | Redactados en [prompts-implementacion.md](prompts-implementacion.md) |
| Primeras historias y prompts de B1 | `HU-020`–`HU-024` integradas en `main` ([PR #84](https://github.com/bcaceres19/barberia/pull/84), issue real `#77`) |
| Historias restantes de B1 a B6 | B2: `HU-040`–`HU-042` integradas; B3: `HU-060`–`HU-067` integradas ([PR #235](https://github.com/bcaceres19/barberia/pull/235) #225, [PR #237](https://github.com/bcaceres19/barberia/pull/237) #226), `HU-068` redactada y bloqueada por secuencia (#227); T3, cierre automático y B4–B6 pendientes |
| Dudas que bloqueaban B0 | `DP-SEG-04`, `DP-SEG-05`, `DP-SEG-06`, resueltas el 2026-08-11 como `DEC-050`–`DEC-052` (ver [dudas-pendientes.md](../00-control/dudas-pendientes.md)) |

Redactar historias muy por anticipado está desaconsejado: lo aprendido al construir cambia lo que la historia siguiente debe decir. Por eso el tercer lote de B3 termina en `HU-068`; T3 y cierre automático se redactan después de implementar y revisar estas transiciones. B4 solo se redacta cuando B3 cumpla su criterio de salida.
