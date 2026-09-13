---
name: generacion-mockups-nava
description: Genera mockups y atlas visuales raster de NAVA cuando el usuario pide diseñar pantallas, rutas, estados o eventos. Entrega imágenes de diseño; no crea implementación ni generadores de código salvo petición explícita.
---

# Generación de mockups NAVA

## Resultado esperado

Entregar archivos de imagen finales y documentación breve de diseño. Una
petición de “mockups”, “mocks”, “pantallas”, “diseño de una ruta” o “atlas por
eventos” no autoriza crear HTML, CSS, JavaScript, TypeScript, Vue, SVG de
maquetación, generadores de capturas ni modificar la aplicación.

Solo crear código o un generador reproducible cuando el usuario lo pida de
forma explícita después de distinguirlo de los mockups visuales.

## Flujo

1. Inspeccionar de forma read-only la guía visual, los atlas asignados y la
   pantalla real para extraer contenido, acciones, estados y exclusiones.
2. Si el usuario menciona un atlas existente, reproducir su organización:
   un PNG por evento, nombres equivalentes y pares simétricos de escritorio y
   móvil. No sustituir el atlas por una lámina compuesta ni por una muestra.
3. Crear una matriz exhaustiva antes de generar. Incluir estados iniciales,
   carga, vacío, error, validación, envío, errores recuperables y resultados
   exitosos que existan realmente. Estados con idéntica composición pueden
   compartir tratamiento, pero si el usuario pide “todos los eventos”, cada
   copy o resultado visible recibe su propio PNG.
4. Usar ImageGen integrado para cada imagen. Una imagen de referencia local se
   inspecciona primero y se etiqueta como referencia o como objetivo de edición.
5. Mantener invariantes entre eventos: cascarón, navegación, retícula, escala,
   posición de controles, tipografía, paleta y datos ficticios. Cambiar solo el
   estado solicitado en cada variante.
6. Guardar los PNG dentro del directorio del atlas solicitado en el proyecto,
   no únicamente en el directorio interno de imágenes generadas.
7. Abrir y revisar cada salida. Rehacer cualquier archivo con texto ilegible,
   elementos fuera de alcance, diferencias de estructura, recortes o errores
   ortográficos.
8. Documentar la matriz evento/viewport/archivo, referencias visuales,
   decisiones y exclusiones. La documentación no sustituye ningún PNG.

## Reglas de diseño del proyecto

- Respetar NAVA / Tailored Grid, los tokens vigentes y el modo de fidelidad
  definido por `docs/03-desarrollo/estandar-diseno-visual.md`.
- Un mockup nunca crea funciones. Si una acción o dato no existe en la HU, el
  contrato o la pantalla real, no se dibuja.
- Evitar dashboards genéricos, tarjetas flotantes repetitivas, glassmorphism,
  neón y clichés decorativos de barbería.
- Mantener contraste, foco visible, objetivos táctiles, reflow real y texto sin
  overflow en todos los viewports solicitados.
- No declarar implementación, QA funcional ni accesibilidad comprobada por el
  hecho de haber generado imágenes.
