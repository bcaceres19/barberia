# Combinaciones acotadas

La matriz histórica de 05-matriz-combinaciones.md contiene muestras, no prueba cobertura completa de pares. Si se afirma cobertura pairwise, generar matriz con PICT/equivalente y verificar pares cubiertos.

| Factor | Valores |
| --- | --- |
| Texto | vacío, espacios, típico, máximo, máximo+1, Unicode, HTML literal |
| Número | mínimo-1, mínimo, típico, máximo, máximo+1, decimal, negativo |
| Red | normal, offline al enviar, lenta real, respuesta perdida tras escritura |
| Sesión | válida, logout otra pestaña, vencida por servidor |
| Estado | inicial, carga, vacío, error, conflicto, éxito, inactivo |
| Ancho | 320/360/768/1280 y texto 200% |
| Contexto | tenant propio/ajeno, dos pestañas, doble envío |

Primero fronteras de un campo con resto válido; después parejas de riesgo; triples dirigidos: sesión vencida+formulario+guardar, doble envío+red lenta+F5, dos reservas+bloqueo+franja. No ejecutar producto cartesiano ni permitir que un campo inválido oculte otra rama.

Límites verificados en validadores/contrato al preparar #300: nombres 120, descripción/nota 500; duración 1–1440 entera; precio COP >0 y dos decimales; anticipación 0–1440 min, ventana 1–90 días, rejilla 5/10/15/20/30/60, cancelación 0–10080 min. Anticipación < ventana total. Login correo 254/contraseña 256; contraseña nueva usa su política de recuperación. Releer fuentes si cambia SHA.

Duración 1440 válida; precio 0 inválido DEC-067; última asignación se puede retirar DEC-114. No conservar esperados históricos opuestos.
