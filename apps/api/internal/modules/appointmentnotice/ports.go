package appointmentnotice

import (
	"context"
	"time"
)

// Tipos de aviso (valores de `appointment_notice.kind`).
const (
	KindCancelledByBarber = "cancelled_by_barber"
	KindRescheduled       = "rescheduled"
)

// Resultados con los que se cierra un aviso (argumento de `notice_finish`).
const (
	OutcomeSent      = "sent"
	OutcomeSkipped   = "skipped"
	OutcomeRetry     = "retry"
	OutcomePermanent = "permanent"
)

// Códigos de causa que quedan en la evidencia. Son códigos cortos, nunca texto
// libre ni datos del cliente (RN-REC-04, RN-DAT-02).
const (
	CodeSuperseded          = "superseded"
	CodeNotConfirmed        = "not_confirmed"
	CodeNotCancelled        = "not_cancelled"
	CodePast                = "past"
	CodeNoEmail             = "no_email"
	CodeProviderUnavailable = "provider_unavailable"
	CodeProviderRejected    = "provider_rejected"
	CodeInvalidTimezone     = "invalid_timezone"
)

// Notice es un aviso reclamado de la cola.
type Notice struct {
	ID         string
	ClaimToken string
	Attempts   int
}

// Context es lo mínimo que el worker necesita para redactar UN aviso con el
// estado actual del turno (RN-REC-03). CustomerEmail y AttendeeName son datos
// personales: solo viajan hasta el puerto de envío y nunca a logs ni evidencia.
type Context struct {
	Kind              string
	AppointmentStatus string
	StartsAt          time.Time
	// StateAt es el updated_at que tenía el turno al leerlo: el estado que este
	// envío cubre.
	StateAt        time.Time
	CustomerEmail  string
	AttendeeName   string
	ServiceName    string
	BarbershopName string
	Timezone       string
	// AlreadyCovered indica que otro aviso ya envió un estado igual o más reciente.
	AlreadyCovered bool
}

// FinishInput cierra un aviso con CAS sobre su claim_token.
type FinishInput struct {
	NoticeID      string
	ClaimToken    string
	Outcome       string
	Code          string
	RetrySeconds  int
	MaxAttempts   int
	CoversStateAt time.Time
	Now           time.Time
}

// Store es el puerto de persistencia del worker. Lo implementa
// postgres.WorkerRepository con las funciones SECURITY DEFINER de la migración
// (DEC-040): el worker no tiene acceso directo a ninguna tabla.
type Store interface {
	ClaimNotices(ctx context.Context, limit, leaseSeconds, maxAttempts int, now time.Time) ([]Notice, error)
	// NoticeContext devuelve found=false cuando la reclamación ya no es vigente.
	NoticeContext(ctx context.Context, noticeID, claimToken string) (Context, bool, error)
	// FinishNotice devuelve false cuando la reclamación se perdió (CAS).
	FinishNotice(ctx context.Context, in FinishInput) (bool, error)
}

// Mailer envía el correo del aviso. El adaptador de Resend en el paquete
// notification lo satisface de forma estructural. startsAtLocal llega ya
// formateado en la zona de la barbería. Un error nunca debe incluir el
// destinatario (RN-DAT-02); si implementa `Permanent() bool` y devuelve true, el
// rechazo no se reintenta.
type Mailer interface {
	SendChangeNotice(
		ctx context.Context,
		kind, email, barbershopName, serviceName, attendeeName, startsAtLocal string,
	) error
}
