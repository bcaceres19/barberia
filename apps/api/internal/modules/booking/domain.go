package booking

import (
	"regexp"
	"strings"
	"time"
)

// barberIDPattern refleja la forma de un UUID, mismo criterio que
// staff.LooksLikeBarberID/schedule.LooksLikeBarberID (CA-002-06: cada
// módulo repite su propia validación de forma sin importar el paquete
// dueño). Un identificador que no cumple esta forma no puede corresponder
// a ninguna fila real: se trata igual que "no existe" (mismo
// apperr.NotFound) en vez de dejar que la consulta SQL falle con un error
// de tipo.
var barberIDPattern = regexp.MustCompile(
	`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`,
)

// LooksLikeBarberID informa si id tiene la forma de un UUID válido.
func LooksLikeBarberID(id string) bool {
	return barberIDPattern.MatchString(id)
}

// Status es uno de los cinco estados cerrados de estados-citas.md §2.
// Texto, nunca números (estados-citas.md §11), igual que la columna
// `status` de appointment (database/migrations/20260827110000_create_appointment_core.sql).
type Status string

const (
	StatusConfirmed           Status = "confirmed"
	StatusCompleted           Status = "completed"
	StatusCancelledByCustomer Status = "cancelled_by_customer"
	StatusCancelledByBarber   Status = "cancelled_by_barber"
	StatusNoShow              Status = "no_show"
)

// OccupiesSchedule refleja EXACTAMENTE el mismo criterio que la columna
// generada appointment.occupies_schedule (estados-citas.md §4/§11): solo
// existe aquí para que el dominio Go pueda razonar sobre el resultado sin
// una consulta adicional, nunca como una segunda fuente de verdad que
// pueda divergir de la base de datos.
func (s Status) OccupiesSchedule() bool {
	switch s {
	case StatusConfirmed, StatusCompleted, StatusNoShow:
		return true
	default:
		return false
	}
}

func (s Status) isValid() bool {
	switch s {
	case StatusConfirmed, StatusCompleted, StatusCancelledByCustomer, StatusCancelledByBarber, StatusNoShow:
		return true
	default:
		return false
	}
}

// Origin distingue una cita nacida del flujo público (B4, todavía sin
// construir) de una creada manualmente por el barbero (HU-061). HU-060 no
// expone ninguno de los dos flujos: solo persiste el origen que el
// llamador ya decidió.
type Origin string

const (
	OriginPublic Origin = "public"
	OriginManual Origin = "manual"
)

func (o Origin) isValid() bool {
	return o == OriginPublic || o == OriginManual
}

// EventType es el vocabulario cerrado de appointment_history.event_type
// (DEC-041). HU-060 solo emite EventTypeAppointmentCreated; los siete
// restantes quedan declarados para que las historias futuras de B3 no
// tengan que reintroducir el vocabulario.
type EventType string

const (
	EventTypeAppointmentCreated             EventType = "appointment_created"
	EventTypeAppointmentRescheduled         EventType = "appointment_rescheduled"
	EventTypeAppointmentServiceChanged      EventType = "appointment_service_changed"
	EventTypeAppointmentCompleted           EventType = "appointment_completed"
	EventTypeAppointmentCancelledByCustomer EventType = "appointment_cancelled_by_customer"
	EventTypeAppointmentCancelledByBarber   EventType = "appointment_cancelled_by_barber"
	EventTypeAppointmentNoShow              EventType = "appointment_no_show"
	EventTypeAppointmentStatusCorrected     EventType = "appointment_status_corrected"
)

// ActorType es quién generó una entrada de historial (RN-HIS-01):
// staff/customer/system, mismo vocabulario que appointment_history.actor_type.
type ActorType string

const (
	ActorTypeStaff    ActorType = "staff"
	ActorTypeCustomer ActorType = "customer"
	ActorTypeSystem   ActorType = "system"
)

// Actor identifica quién ejecuta CreateInternal para el registro de
// historial (RN-HIS-01). La forma exigida coincide con
// appointment_history_actor_shape_ck: staff exige StaffUserID sin
// CustomerID, customer exige CustomerID sin StaffUserID, system exige
// ambos vacíos.
type Actor struct {
	Type        ActorType
	StaffUserID *string
	CustomerID  *string
}

func (a Actor) isValid() bool {
	switch a.Type {
	case ActorTypeStaff:
		return a.StaffUserID != nil && *a.StaffUserID != "" && a.CustomerID == nil
	case ActorTypeCustomer:
		return a.CustomerID != nil && *a.CustomerID != "" && a.StaffUserID == nil
	case ActorTypeSystem:
		return a.StaffUserID == nil && a.CustomerID == nil
	default:
		return false
	}
}

// Límites que coinciden con los CHECK de
// database/migrations/20260827110000_create_appointment_core.sql: el
// dominio Go no inventa un límite más estricto ni más laxo que la base ya
// aplica.
const (
	AttendeeNameMaxLength        = 120
	CustomerFullNameMaxLength    = 120
	CustomerNoteMaxLength        = 500
	ServiceNameSnapshotMaxLength = 120
	MinDurationMinutesSnapshot   = 1
	MaxDurationMinutesSnapshot   = 1440
)

var phonePattern = regexp.MustCompile(`^\+[1-9][0-9]{7,14}$`)

// emailShapePattern refleja customer_email_ck: al menos un carácter, '@',
// al menos un carácter, '.', al menos un carácter (mismo criterio que el
// LIKE '_%@_%._%' de la base). La forma canónica (minúsculas, sin
// espacios) se valida aparte.
var emailShapePattern = regexp.MustCompile(`^.+@.+\..+$`)

// ServiceSnapshot son los datos de servicio ya congelados que la cita va a
// conservar (DEC-004): nunca se sincronizan en segundo plano con el
// catálogo. PriceAmountCents es el precio en centavos, entero exacto,
// nunca coma flotante (mismo criterio que catalog.Service.PriceCents).
type ServiceSnapshot struct {
	Name             string
	DurationMinutes  int
	PriceAmountCents int64
	Currency         string
}

func (s ServiceSnapshot) validate() error {
	name := strings.TrimSpace(s.Name)
	if name == "" {
		return errServiceNameSnapshotRequired()
	}
	if len(name) > ServiceNameSnapshotMaxLength {
		return errServiceNameSnapshotTooLong()
	}
	if s.DurationMinutes < MinDurationMinutesSnapshot || s.DurationMinutes > MaxDurationMinutesSnapshot {
		return errDurationSnapshotOutOfRange()
	}
	if s.PriceAmountCents < 0 {
		return errPriceSnapshotNegative()
	}
	if !currencyPattern.MatchString(s.Currency) {
		return errCurrencySnapshotInvalid()
	}
	return nil
}

var currencyPattern = regexp.MustCompile(`^[A-Z]{3}$`)

// NewCustomerInput son los datos de un cliente todavía no persistido.
// Phone y Email son opcionales: RN-CIT-02 permite omitirlos en una cita
// manual. La reconciliación por teléfono (DEC-045) o correo (DEC-046) NO
// vive aquí: HU-060 solo persiste la forma que el llamador ya decidió;
// HU-061 decide cuándo construir NewCustomerInput frente a reutilizar un
// CustomerID existente (DP-CIT-01).
type NewCustomerInput struct {
	FullName string
	Phone    *string
	Email    *string
}

func (c NewCustomerInput) validate() error {
	name := strings.TrimSpace(c.FullName)
	if name == "" {
		return errCustomerFullNameRequired()
	}
	if len(name) > CustomerFullNameMaxLength {
		return errCustomerFullNameTooLong()
	}
	if c.Phone != nil && !phonePattern.MatchString(*c.Phone) {
		return errCustomerPhoneInvalid()
	}
	if c.Email != nil {
		e := *c.Email
		if e != strings.ToLower(e) || e != strings.TrimSpace(e) || strings.ContainsAny(e, " \t\n\r") {
			return errCustomerEmailInvalid()
		}
		if !emailShapePattern.MatchString(e) || len(e) > 254 {
			return errCustomerEmailInvalid()
		}
	}
	return nil
}

// CustomerInput identifica al cliente de una cita: exactamente uno de los
// dos casos. ExistingID es un cliente que el llamador ya resolvió
// (verificado por FK+RLS dentro de la transacción); New crea una fila
// nueva. HU-060 no decide cuál de los dos usar (DP-CIT-01): eso es
// política de HU-061.
type CustomerInput struct {
	ExistingID *string
	New        *NewCustomerInput
}

func (c CustomerInput) validate() error {
	hasExisting := c.ExistingID != nil && *c.ExistingID != ""
	hasNew := c.New != nil
	if hasExisting == hasNew {
		return errCustomerInputInvalid()
	}
	if hasNew {
		return c.New.validate()
	}
	return nil
}

// Customer es un cliente ya persistido.
type Customer struct {
	ID           string
	BarbershopID string
	FullName     string
	Phone        *string
	Email        *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Appointment es una cita ya persistida. OccupiesSchedule refleja la
// columna generada real, leída de vuelta tras el INSERT: nunca se calcula
// en Go de forma independiente (estados-citas.md §11).
type Appointment struct {
	ID                       string
	BarbershopID             string
	BarberID                 string
	ServiceID                string
	CustomerID               string
	AttendeeName             string
	StartsAt                 time.Time
	EndsAt                   time.Time
	Status                   Status
	OccupiesSchedule         bool
	Origin                   Origin
	ServiceNameSnapshot      string
	DurationMinutesSnapshot  int
	PriceAmountCentsSnapshot int64
	CurrencySnapshot         string
	CustomerNote             *string
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

// CreateInternalInput es la entrada ya autorizada de la primitiva
// transaccional interna (§2.2 del prompt de HU-060): quien la construye
// (HU-061 en adelante) ya validó jornada, bloqueos y asignación
// servicio-barbero; CreateInternal solo aplica las garantías de forma e
// integridad que le corresponden a este núcleo.
type CreateInternalInput struct {
	BarberID     string
	ServiceID    string
	AttendeeName string
	StartsAt     time.Time
	EndsAt       time.Time
	Origin       Origin
	Service      ServiceSnapshot
	CustomerNote *string
	Customer     CustomerInput
	Actor        Actor
}

func (in CreateInternalInput) validate() error {
	if strings.TrimSpace(in.BarberID) == "" {
		return errBarberIDRequired()
	}
	if strings.TrimSpace(in.ServiceID) == "" {
		return errServiceIDRequired()
	}
	name := strings.TrimSpace(in.AttendeeName)
	if name == "" {
		return errAttendeeNameRequired()
	}
	if len(name) > AttendeeNameMaxLength {
		return errAttendeeNameTooLong()
	}
	if !in.Origin.isValid() {
		return errOriginInvalid()
	}
	if !in.EndsAt.After(in.StartsAt) {
		return errIntervalInvalid()
	}
	if int64(in.EndsAt.Sub(in.StartsAt).Seconds()) != int64(in.Service.DurationMinutes)*60 {
		return errIntervalDurationMismatch()
	}
	if err := in.Service.validate(); err != nil {
		return err
	}
	if in.CustomerNote != nil && len(*in.CustomerNote) > CustomerNoteMaxLength {
		return errCustomerNoteTooLong()
	}
	if err := in.Customer.validate(); err != nil {
		return err
	}
	if !in.Actor.isValid() {
		return errActorInvalid()
	}
	return nil
}

// CreateInternalResult es el desenlace de CreateInternal: la cita
// `confirmed` recién creada y el cliente vinculado (ya existente o recién
// insertado, según CustomerInput).
type CreateInternalResult struct {
	Appointment Appointment
	Customer    Customer
}
