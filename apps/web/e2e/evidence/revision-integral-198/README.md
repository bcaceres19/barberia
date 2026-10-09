# Revisión integral final del rediseño NAVA / Tailored Grid · #198

Fase 13 de #184. Recorrido real (API y PostgreSQL reales, usuario sintético de `database/testdata/ui_bloqueos_barberos_286.sql`) de las rutas del panel y de acceso, en los cuatro anchos de la guía visual (1280, 768, 360 y 320 px), con movimiento reducido y las transiciones finitas ya terminadas.

## Resultado

| Comprobación                                    | Resultado                                                                                               |
| ----------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| Rutas × anchos medidos                          | 10 rutas × 4 anchos = 40 combinaciones (`report.json`)                                                  |
| Axe WCAG 2.0/2.1/2.2 A y AA, contraste incluido | 0 violaciones en las 40                                                                                 |
| Desbordamiento horizontal                       | Ninguno en las 40                                                                                       |
| Errores de consola y de página                  | 0                                                                                                       |
| Cascarón                                        | Un solo `<main>` y un solo `<h1>` por ruta; títulos con la misma tipografía de display en todo el panel |
| Sesión                                          | Cada ruta del panel se cargó autenticada (la ruta final coincide con la pedida)                         |
| Dock a 1280 px                                  | Ningún rótulo truncado (ver hallazgo)                                                                   |
| `/panel/bloqueos`                               | Redirige a Barberos (`DEC-105`)                                                                         |

Rutas: Agenda (`/panel`), Nuevo turno, Barberos, Servicios, Servicios por barbero, Horarios, Configuración de la barbería, Reserva pública, Acceso y Recuperar acceso. El detalle de un turno, la reserva pública abierta al cliente y el acceso al turno tienen su propia evidencia por fase y no dependen de datos que esta cuenta no tiene.

## Hallazgo corregido en esta revisión

El dock de escritorio repartía el ancho en partes iguales entre siete destinos y a 1280 px «Servicios por barbero» se cortaba con «…» en todas las pantallas del panel. Ahora cada destino mide lo que su rótulo necesita y reparte el sobrante (`flex: 1 1 auto`). La spec lo comprueba (`truncatedNavLabels`); sin la corrección fallaba en las ocho rutas del panel.

## Fases del rediseño y su evidencia

La comparación con el mockup asignado de cada familia se hizo y quedó registrada en el PR de su fase, con capturas lado a lado y diferencia:

| Fase                         | Issue      | PR                                                                                                                |
| ---------------------------- | ---------- | ----------------------------------------------------------------------------------------------------------------- |
| Fundaciones y shell          | #186, #187 | #199, #202                                                                                                        |
| Acceso y recuperación        | #188, #213 | #204, #210, #214, #218                                                                                            |
| Agenda diaria                | #189       | #217                                                                                                              |
| Nuevo turno                  | #190       | #222                                                                                                              |
| Detalle del turno            | #191       | #229                                                                                                              |
| Barberos                     | #192       | #230                                                                                                              |
| Configuración de la barbería | #193       | #231                                                                                                              |
| Servicios                    | #194       | #232                                                                                                              |
| Servicios por barbero        | #195       | #233                                                                                                              |
| Horarios                     | #196       | #234                                                                                                              |
| Bloqueos                     | #197       | `DEC-105` (ventana dentro de Barberos, identidad guiada) y #289; series por fechas, excepciones y edición en #319 |

## Capturas

`{ruta}-1280.png` y `{ruta}-320.png` por ruta, más `report.json` con el detalle por ruta y ancho. Datos sintéticos.

## Reproducción local

Aplicar `database/testdata/ui_bloqueos_barberos_286.sql` con rol migrador solo en la BD local de pruebas, arrancar API y Vite y ejecutar `APP_BASE_URL=<vite> pnpm exec playwright test e2e/revision-integral-nava-198.spec.ts --project=chromium-desktop --workers=1`. La spec corre solo en Chromium de escritorio, reutiliza la sesión de `apps/web/.auth/` (ignorado por Git) y no modifica datos.
