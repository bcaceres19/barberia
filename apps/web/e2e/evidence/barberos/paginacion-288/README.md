# Barberos · paginación por viewport · #288

Modo de identidad guiada NAVA; patrón funcional y visual del paginador de Servicios, sin mockup exacto. Se mantienen retratos, edición y Bloqueos.

E2E contra API Go y PostgreSQL reales: `barberos-paginacion.spec.ts`; fixture dedicada `database/testdata/ui_paginacion_barberos_288.sql` (17 y 3 barberos sintéticos). Chromium y Firefox pasaron navegación, reemplazo sin acumulación, error/reintento, ajuste al resize, teclado/foco, apertura/cierre de Bloqueos, modo cursor compatible y aislamiento de tenants. Axe no detectó infracciones WCAG 2 A/AA, 2.1 AA y 2.2 AA en los seis viewports.

Capturas con viewport efectivo comprobado: 320×720, 360×800, 768×900, 1280×900, 1280×650 y 640×450. El último es la aproximación de 200 % por reflow prevista por la guía de pruebas; no sustituye una auditoría de zoom/texto completa. Se comprobaron scrollHeight/clientHeight, anchura, pie visible y aprovechamiento del espacio disponible. Las capturas solo muestran fixtures ficticios.

Componente: regreso de página, total, reintento del destino fallido y descarte de respuesta antigua durante resize. PostgreSQL: total, clamping, vacío y dos tenants; HTTP: mezcla/parámetros inválidos y forma de respuesta.
