package customeraccess

import "time"

// MaxTokenLength acota el valor recibido en la ruta antes de hashear o
// consultar nada: el token real que HU-097 emite (confirm.go,
// accessTokenBytes) siempre codifica en base64.RawURLEncoding.EncodeToString
// exactamente 43 caracteres (32 bytes, DEC-089); un valor más largo nunca
// puede coincidir con ninguna fila y se descarta sin tocar la base de datos
// (defensa barata contra una ruta arbitrariamente larga).
const MaxTokenLength = 256

// AppointmentView es la proyección mínima que HU-098 expone (CA-098-03):
// nunca un identificador interno, contacto del cliente, nota, historial ni
// ningún otro turno. barberName/serviceName/attendeeName ya son los
// snapshots o nombres vigentes que el resto del sistema usa para mostrar al
// cliente (mismo criterio que ConfirmedPublicAppointmentResponse, HU-097).
type AppointmentView struct {
	BarbershopName                 string
	Timezone                       string
	AttendeeName                   string
	ServiceName                    string
	DurationMinutes                int
	BarberName                     string
	StartsAt                       time.Time
	EndsAt                         time.Time
	Status                         string
	CancellationDeadlineMinutes    int
	LateCancellationClientAllowed  bool
	LateCancellationReasonRequired bool
}
