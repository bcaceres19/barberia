---
titulo: "Checklist de autenticación y sesión"
version: "1.0"
estado: "Herramienta operativa, no normativa"
responsable: "Propietario del proyecto"
ultima_actualizacion: "2026-08-25"
documentos_relacionados:
  - "README.md"
  - "02-checklist-transversal.md"
  - "../02-requisitos/historias-usuario.md"
---

# Checklist de autenticación y sesión

Cubre `LoginPage`/`LoginForm`, `PhoneChallengeForm`, `RecoveryPage` (3 pasos), la guardia del panel privado y el cierre de sesión. Ejecuta primero el bloque completo de [`02-checklist-transversal.md`](02-checklist-transversal.md) sobre cada pantalla de este módulo.

## 1. Login (`LoginPage`)

Límites reales de `loginValidation.ts`: correo máx. 254 caracteres, contraseña máx. 256.

- [ ] Correo vacío / solo espacios → error de "escribe tu correo", no un 422 del servidor.
- [ ] Correo sin `@`, sin punto en el dominio, con espacios internos → rechazado en cliente.
- [ ] Correo exactamente en 254 caracteres con forma válida → aceptado; en 255 → rechazado en cliente.
- [ ] Contraseña vacía → error de cliente. Contraseña de 256 caracteres exactos → aceptada; de 257 → rechazada en cliente.
- [ ] Correo válido pero inexistente vs. correo existente con contraseña incorrecta vs. correo existente pero **usuario inactivo**: la respuesta visible (mensaje, tiempo de respuesta aproximado, forma del error) debe ser **indistinguible** entre los tres casos — es una garantía explícita del backend (`TestLogin_StructuralNonEnumeration...`, `TestLogin_InactiveUser_IsIndistinguishableFromUnknownEmail`); confírmalo también desde la interfaz, no solo leyendo el test.
- [ ] Varios intentos fallidos seguidos y rápidos sobre el mismo correo: verificar que aparece el mecanismo de defensa escalonada (`HU-007`) sin que el mensaje revele si el correo existe.
- [ ] Doble clic en "Iniciar sesión" con credenciales correctas: una sola sesión creada, no dos.
- [ ] Autocompletado del navegador/gestor de contraseñas rellenando el campo: el botón se habilita correctamente sin requerir un evento de teclado manual.
- [ ] Login exitoso y luego **navegación directa a `/login` de nuevo** con la sesión ya activa: ¿redirige al panel o vuelve a mostrar el formulario?

## 2. Reto telefónico (`PhoneChallengeForm`, si aplica al flujo)

- [ ] Código correcto, incorrecto, vencido y ya usado antes: cada uno con su resultado esperado; el vencido/ya usado no debe aceptarse silenciosamente.
- [ ] Reenviar el código en clics rápidos y repetidos: no debe disparar múltiples envíos reales del proveedor ni resetear una cuenta regresiva ya en curso de forma inconsistente.
- [ ] Pegar un código con espacios o caracteres no numéricos: se limpia o se rechaza con claridad, nunca se envía tal cual a una comparación que silenciosamente falla.

## 3. Recuperación de acceso (`RecoveryPage`, 3 pasos)

Límites reales: código de 6 dígitos numéricos exactos (`validateRecoveryCode`); contraseña nueva entre 10 y 128 caracteres (`NEW_PASSWORD_MIN_LENGTH`/`MAX_LENGTH`); la contraseña nueva **no puede ser igual al correo** (comparación insensible a mayúsculas).

- [ ] Paso 1 — correo inexistente: el mensaje debe ser el mismo que para uno existente (no enumeración, `CT-007`/`DEC-064`/`DEC-065`).
- [ ] Paso 2 — código de 5 dígitos, de 7 dígitos, con letras, vacío: rechazado en cliente sin llegar al servidor.
- [ ] Paso 2 — código correcto pero **vencido**, código correcto pero **ya usado en un intento anterior**, código incorrecto: mismo mensaje uniforme para los tres (`CT-007`), verificar que ninguno filtra cuál era el motivo real.
- [ ] Paso 3 — contraseña nueva de 9 caracteres (rechazada), de exactamente 10 (aceptada), de 128 (aceptada), de 129 (rechazada).
- [ ] Paso 3 — contraseña nueva idéntica al correo de la cuenta, en minúsculas y en mayúsculas distintas al correo original: ambas rechazadas por la misma regla.
- [ ] Recargar la página (F5) a mitad del paso 2 o 3: ¿se pierde el destino enmascarado ya revelado, o el flujo se puede reanudar razonablemente, o al menos falla con un mensaje claro de "vuelve a empezar"?
- [ ] Completar la recuperación y luego intentar **reutilizar el mismo código** en una segunda pestaña que quedó abierta en el paso 2: debe rechazarse.
- [ ] Reenvío del código (`CA-011-04`, cuenta regresiva de 60 s) en clics repetidos antes de que expire la cuenta regresiva: no debe reiniciar el conteo de forma que permita reenvíos ilimitados burlando el propósito de la espera.

## 4. Sesión y guardia del panel

- [ ] Navegación directa por URL a `/panel` (o cualquier ruta privada) sin sesión: redirección a login, y tras iniciar sesión correctamente, aterriza en el destino originalmente pretendido, no siempre en la raíz del panel.
- [ ] Sesión activa, cerrar la pestaña sin logout explícito, volver a abrir la app: la sesión persistente (`HU-006`) sigue activa según el tiempo configurado, sin pedir login de nuevo antes de tiempo.
- [ ] Logout explícito y luego botón "Atrás" del navegador: no debe mostrar contenido privado cacheado ni permitir una acción real sin sesión.
- [ ] Con la sesión a punto de expirar (o forzada a expirar manualmente en base de datos/cookie), realizar una acción que muta datos (crear un servicio, por ejemplo): el 401 resultante no debe simular un éxito ni dejar el formulario en un estado ambiguo — instala el mismo manejo único de 401 que ya cubre `installSessionHandling`.
- [ ] Cabecera con el nombre de la barbería activa: tras cambiar el nombre desde "Configuración" en otra pestaña, ¿la cabecera de la primera pestaña queda desactualizada indefinidamente o se corrige en la siguiente navegación/petición?
