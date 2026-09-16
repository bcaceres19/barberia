// Package customeraccess implementa HU-098: la lectura mínima del turno del
// cliente mediante el token de acceso emitido por publicbooking (HU-097,
// DEC-089). No importa publicbooking ni booking: consume directamente la
// tabla appointment_access_token y las tablas ya existentes de appointment/
// barber/barbershop a través de su propio adaptador postgres, exactamente
// el mismo criterio de independencia entre módulos que CA-002-06 ya exige
// (duplicar la forma pequeña que necesita en vez de importar el paquete
// dueño). Fuera de alcance: cancelar (HU-099), editar, listar más de un
// turno, portal o cuenta de cliente.
package customeraccess
