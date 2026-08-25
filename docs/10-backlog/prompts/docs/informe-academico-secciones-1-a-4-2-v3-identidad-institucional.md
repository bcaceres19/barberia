---
prompt_id: "PROMPT-DOCS-INFORME-ACADEMICO-v3"
version: "3.0"
kind: "docs"
status: "executed"
target_agents:
  - "codex"
repository: "bcaceres19/barberia"
base_branch: "main"
primary_hu: null
related_hu:
  - "HU-001"
  - "HU-002"
  - "HU-003"
  - "HU-004"
  - "HU-005"
  - "HU-006"
  - "HU-007"
  - "HU-008"
  - "HU-009"
  - "HU-010"
  - "HU-011"
  - "HU-012"
  - "HU-020"
  - "HU-021"
  - "HU-022"
  - "HU-023"
  - "HU-024"
issue: "81"
issue_url: "https://github.com/bcaceres19/barberia/issues/81"
branch: "docs/81-informe-academico"
pr: "82"
pr_url: "https://github.com/bcaceres19/barberia/pull/82"
supersedes: "PROMPT-DOCS-INFORME-ACADEMICO-v2"
depends_on:
  - "Informe académico v2 gerencial ejecutado y verificado"
  - "Identidad institucional y nombres suministrados por el usuario durante la ejecución; no se persisten en Git"
  - "Logo oficial suministrado como archivo local durante la ejecución; no se versiona en Git"
rules:
  - "RN-DIS-01"
  - "RN-CON-03"
  - "RN-CNF-01"
  - "RN-HIS-01"
  - "RN-REC-01"
  - "RN-TEN-01"
  - "RN-DAT-01"
  - "RN-IDE-01"
decisions:
  - "DEC-016"
  - "DEC-017"
  - "DEC-018"
  - "DEC-019"
  - "DEC-020"
  - "DEC-021"
  - "DEC-022"
  - "DEC-025"
  - "DEC-026"
  - "DEC-027"
  - "DEC-028"
  - "DEC-030"
  - "DEC-031"
  - "DEC-032"
acceptance_criteria:
  - "CA-DOC-01: la portada muestra el logo oficial, la institución, la asignatura y los dos autores confirmados"
  - "CA-DOC-02: la sección de integrantes refleja responsabilidad compartida, sin asignar cargos permanentes a los dos integrantes"
  - "CA-DOC-03: el primer barbero participante queda identificado en el DOCX y su función incluye validar el uso real y facilitar contacto con posibles usuarios"
  - "CA-DOC-04: todo el documento utiliza Arial y los textos, títulos y encabezados no usan colores decorativos"
  - "CA-DOC-05: el color se reserva para estados integrados, parciales o en ejecución y pendientes"
  - "CA-DOC-06: el logo tiene texto alternativo y todas las páginas finales pasan revisión visual después de la última edición"
excluded_scope:
  - "No cambiar el alcance P0, el corte documental ni los estados de desarrollo"
  - "No inventar asesor, ciudad, fecha, facultad o programa académico"
  - "No persistir nombres personales ni el logo suministrado en archivos versionados del repositorio"
  - "No modificar ni sobrescribir las versiones v1 y v2 del DOCX"
tests:
  - "Validación estructural de autores, institución, asignatura, participante, 45 RU y tabla de contenido"
  - "Auditoría de fuente Arial y de rellenos de color limitados a celdas de estado"
  - "Auditorías de accesibilidad, relaciones, imágenes, privacidad y estilos"
  - "Render final a PNG e inspección individual de todas las páginas"
---

# PROMPT-DOCS-INFORME-ACADEMICO-v3

## Objetivo

Actualizar el informe gerencial v2 con la identidad académica suministrada por el usuario, la composición real del equipo, el primer barbero participante y una presentación visual sobria. Los datos personales se incorporan únicamente al DOCX entregable y no se copian a Git.

## Cambios autorizados

- Insertar el logo oficial suministrado en la portada y en el encabezado, con tamaño legible, proporción preservada y texto alternativo.
- Sustituir el marcador institucional por la institución indicada y mostrar la asignatura `Ingeniería de Software III`.
- Sustituir los marcadores de autor por los dos nombres suministrados en la solicitud vigente.
- Explicar que ambos integrantes comparten análisis, requisitos, diseño, construcción, pruebas, documentación y mantenimiento, sin roles fijos.
- Identificar al barbero suministrado como primer participante, fuente de retroalimentación y enlace para facilitar contacto con posibles usuarios interesados.
- Conservar el marcador del asesor y cualquier dato académico todavía desconocido.

## Sistema visual

- Usar Arial en todos los estilos, texto directo, tablas, encabezados y pies de página.
- Usar texto negro y fondos blancos para portada, títulos, narración, encabezados y tablas sin estado.
- El azul se admite únicamente como parte del logo oficial.
- Aplicar color solo a celdas de estado: verde para integrado o realizado, ámbar para parcial o en ejecución y rojo suave para pendiente o faltante.
- Eliminar rellenos decorativos, alternancia de filas y encabezados de tabla en azul.
- Conservar bordes, espaciado y jerarquía suficientes para que el documento siga siendo legible y profesional.

## Verificación y entrega

- Generar un DOCX nuevo sin sobrescribir v1 ni v2.
- Actualizar la tabla de contenido con las páginas obtenidas en el render final.
- Inspeccionar todas las páginas al 100 % después de la última edición.
- Verificar 45 RU únicos, logo embebido, texto alternativo, Arial y ausencia de rellenos no asociados con estados.
- Actualizar este prompt y el catálogo con páginas, SHA-256 y evidencia de QA.
- Actualizar el PR #82 sin fusionarlo.

## Resultado de ejecución

- Entregable generado en `temp/informe-academico/Informe_Academico_Sistema_Agenda_Barberias_EAM_Final.docx`; no se versiona por contener los datos personales suministrados durante la ejecución.
- Documento final de 26 páginas, 3 secciones, 9 tablas y 45 requisitos `RU-001`–`RU-045` únicos y consecutivos.
- Se incorporaron el logo oficial embebido con texto alternativo, la identidad académica, las responsabilidades compartidas y el primer participante, sin persistir nombres personales ni el archivo del logo en Git.
- La única fuente explícita detectada en el contenido es Arial. Los únicos rellenos no blancos son `E8F5E9`, `FFF4CC` y `FCE8E6`, aplicados exclusivamente a celdas de estado.
- Se validaron las relaciones internas, imágenes, secciones, encabezados, metadatos y privacidad. La auditoría de accesibilidad reportó 0 hallazgos altos; los 3 medios corresponden a tablas de maquetación del encabezado y los 2 bajos a URL bibliográficas visibles.
- Las 26 páginas fueron renderizadas después de la última edición e inspeccionadas individualmente, sin recortes, superposiciones ni desbordamientos.
- SHA-256 final: `735540788ccb942caf6b2fe2333c0d6d083ec5926954e0f22ce103577d472859`.
