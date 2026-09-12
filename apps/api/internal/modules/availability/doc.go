// Package availability implementa el dominio puro de HU-094: proyectar los
// inicios públicos válidos de un servicio, dado el barbero, dentro de una
// ventana de fechas.
//
// Este paquete NO importa schedule, booking, catalog, shops, Chi ni pgx
// (docs/03-desarrollo/estandar-backend-go.md §5.23): recibe únicamente
// tramos laborales y ocupaciones ya resueltos como instantes absolutos
// (time.Time), y una política ya resuelta (duración del servicio, rejilla,
// anticipación y ventana como instantes límite). No sabe qué es un festivo,
// una excepción o una cita: esas reglas viven en schedule (B2) y booking
// (B3, RN-CON-01/RN-CON-03), y HU-094 tiene prohibido duplicarlas.
//
// GenerateStarts es la única función exportada relevante para el cálculo:
// fusiona los tramos que se solapan (RN-DIS-02), resta las ocupaciones
// (bloqueos + citas que ocupan agenda, ya fusionadas entre sí) y genera
// franjas cada N minutos DENTRO de cada intervalo libre resultante,
// anclando la rejilla al inicio de ESE intervalo libre -nunca al inicio del
// tramo laboral original- lo que implementa DEC-084 (resuelve DP-PUB-03)
// sin necesitar un caso especial: cualquier intervalo libre que empieza
// justo después de una interrupción ya reinicia la rejilla por construcción.
package availability
