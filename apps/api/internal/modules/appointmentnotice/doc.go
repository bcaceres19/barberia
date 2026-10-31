// Package appointmentnotice avisa por correo al cliente cuando el barbero
// cancela o reprograma su turno (HU-130, DEC-142, primer tramo de F-NOT-02).
//
// El aviso nace en la misma transacción que el cambio (booking lo encola mediante
// su NoticeHook) y lo envía el worker: reclama con lease, lee el estado VIGENTE
// del turno, redacta el mensaje y deja evidencia de cada intento sin datos de
// contacto (RN-REC-03, RN-REC-04). Este paquete no importa booking ni
// notification: el correo sale por el puerto [Mailer], que cmd/worker satisface
// con el adaptador de Resend.
//
// Fuera de este tramo quedan los recordatorios programados (F-NOT-03), WhatsApp
// y la configuración por evento y canal (RN-REC-05).
package appointmentnotice
