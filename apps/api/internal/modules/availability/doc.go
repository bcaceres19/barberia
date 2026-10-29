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
// franjas en la rejilla fija del tramo laboral (inicio del tramo + k × N),
// dentro de cada intervalo libre resultante. La rejilla se ancla al inicio del
// tramo y NO se reinicia donde termina una interrupción: DEC-135 sustituye el
// reinicio de DEC-084 para que todas las franjas queden alineadas con el paso
// configurado.
package availability
