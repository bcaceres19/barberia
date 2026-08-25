---
titulo: "Matriz de combinaciones por pares"
version: "1.0"
estado: "Herramienta operativa, no normativa"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-25"
documentos_relacionados:
  - "README.md"
  - "01-metodologia-y-uso.md"
  - "04-checklist-modulos-catalogo.md"
---

# Matriz de combinaciones por pares

Tablas ya reducidas por combinación por pares (ver técnica en [`01-metodologia-y-uso.md`](01-metodologia-y-uso.md) §2), listas para ejecutar fila por fila. Cada fila es una sesión concreta: una combinación de valores de distintas dimensiones que aparece junta al menos una vez, cubriendo todos los pares posibles con el mínimo de filas.

## 1. Crear servicio (`CatalogPage`)

Dimensiones: **Nombre** (vacío / típico / 120 exacto / 121 / duplicado-activo), **Duración** (0 / 1 / 1440 / 1441 / "30.5"), **Precio** ("0" / "45000.50" / "45000.999" / vacío / "45,000"), **Ancho** (320 / 768 / 1280), **Red** (normal / offline al enviar / lenta).

| # | Nombre | Duración | Precio | Ancho | Red | Resultado esperado |
| --- | --- | --- | --- | --- | --- | --- |
| 1 | típico | 1 | 45000.50 | 320 | normal | Creado; visible en la lista a 320 px sin desbordes |
| 2 | 121 chars | 1440 | 0 | 768 | normal | Ambos errores de campo mostrados juntos, sin envío |
| 3 | vacío | 0 | 45000.999 | 1280 | normal | Tres errores de campo, ninguno llega al servidor |
| 4 | duplicado-activo | 1441 | vacío | 320 | offline al enviar | Rechazo por `name-conflict` **o** por red, nunca ambos a la vez de forma confusa; datos del formulario conservados |
| 5 | 120 exacto | "30.5" | 45,000 | 768 | lenta | Errores de duración y precio junto con estado de carga visible mientras se valida en cliente (no debería ni llegar a la red) |
| 6 | típico | 1 | 45000.50 | 1280 | offline al enviar | Error de red claro, "Reintentar" disponible, formulario intacto |
| 7 | 121 chars | 1440 | 45000.999 | 320 | lenta | Errores de cliente visibles antes de que la red lenta importe |
| 8 | duplicado-activo | 0 | 45000.50 | 768 | normal | `name-conflict` del servidor (duración ya habría fallado en cliente primero — confirma que la validación de cliente bloquea el envío) |

## 2. Configuración de la barbería — guardar sección (`SettingsPage`)

Dimensiones: **Correo contacto** (vacío / válido / inválido / 254 exacto), **Teléfono contacto** (vacío / E.164 válido / con espacios / sin `+`), **Zona horaria** (válida IANA / vacía / no reconocida por Postgres), **Pestañas** (una / dos concurrentes editando campos distintos).

| # | Correo | Teléfono | Zona horaria | Pestañas | Resultado esperado |
| --- | --- | --- | --- | --- | --- |
| 1 | vacío | vacío | válida | una | Guardado; `contactEmail`/`contactPhone` persisten como `null`, no `""` |
| 2 | válido | con espacios | vacía | una | Error de teléfono (cliente) y de zona (cliente); correo no se envía solo porque los otros fallen |
| 3 | inválido | E.164 válido | no reconocida | una | Error de correo en cliente; si se corrige y se reintenta, el error de zona debe venir del servidor como error general, sin escritura parcial |
| 4 | 254 exacto | sin `+` | válida | una | Error solo de teléfono; correo de 254 se guarda íntegro |
| 5 | válido | vacío | válida | dos | Ambas pestañas guardan su sección casi a la vez; verificar cuál gana y si la otra pestaña queda con datos obsoletos hasta refrescar |

## 3. Ciclo de vida de un servicio + asignación cruzada

Dimensiones: **Estado inicial** (activo con 1 barbero asignado / activo con varios / recién reactivado), **Acción** (desactivar / reactivar / desasignar última / desasignar no-última), **Concurrencia** (una pestaña / dos pestañas simultáneas), **Ancho**.

| # | Estado inicial | Acción | Concurrencia | Ancho | Resultado esperado |
| --- | --- | --- | --- | --- | --- |
| 1 | activo, 1 barbero | desasignar esa única asignación | una | 768 | Rechazado (`DEC-068`, última asignación de servicio activo) con mensaje claro |
| 2 | activo, 1 barbero | desactivar servicio, luego desasignar | una | 768 | Tras desactivar, ¿la regla de "última asignación" se sigue aplicando a un servicio ya inactivo? Documentar comportamiento real |
| 3 | activo, varios barberos | desasignar uno cualquiera (no el último) | dos pestañas, la misma asignación | 320 | Una tiene éxito, la otra recibe conflicto controlado — nunca doble desasignación silenciosa |
| 4 | activo, varios barberos | desactivar servicio (doble clic) | una | 1280 | Una sola desactivación real; segundo clic con `idempotency-conflict` bien comunicado |
| 5 | recién reactivado | crear un servicio nuevo con el mismo nombre que tenía antes de reactivarse | una | 768 | Verificar `name-conflict` esperado entre dos servicios activos con el mismo nombre |

## 4. Login + throttle + reto (si `HU-007` está activo en el entorno probado)

Dimensiones: **Correo** (existente / inexistente), **Contraseña** (correcta / incorrecta), **Intentos previos** (0 / justo antes del umbral / en el umbral), **Ancho**.

| # | Correo | Contraseña | Intentos previos | Ancho | Resultado esperado |
| --- | --- | --- | --- | --- | --- |
| 1 | existente | correcta | 0 | 320 | Login exitoso, sesión creada |
| 2 | inexistente | cualquiera | 0 | 768 | Mismo mensaje que credenciales incorrectas |
| 3 | existente | incorrecta | justo antes del umbral | 320 | Mismo mensaje uniforme; siguiente intento activa el reto |
| 4 | existente | incorrecta | en el umbral | 1280 | Reto telefónico exigido, mensaje no revela por qué |
| 5 | inexistente | cualquiera | en el umbral | 768 | Igual de indistinguible que la fila 4 pese a que el correo no existe |

## 5. Extender esta matriz

Al cubrir un módulo nuevo, sigue el mismo patrón: dimensiones a la izquierda del encabezado, filas que garantizan cada par de valores al menos una vez, columna final siempre con el resultado esperado citando la regla `RN-*`/`CA-*` o el estándar visual que lo respalda cuando exista uno. Una fila sin resultado esperado explícito no es una prueba, es una excursión sin objetivo.
