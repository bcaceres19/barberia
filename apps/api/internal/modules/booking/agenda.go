package booking

import (
	"context"
	"fmt"
	"strings"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/clock"
)

// dailyAgendaDateLayout es el único formato que ListDailyAgenda admite para
// una fecha civil explícita: "AAAA-MM-DD", sin hora ni zona (RFC 3339
// solo-fecha), mismo criterio de rechazo estricto que civilLocalLayout en
// manual.go.
const dailyAgendaDateLayout = "2006-01-02"

// DailyAgendaLimit acota la respuesta de ListDailyAgenda: un solo día de
// un solo barbero nunca se acerca a este volumen en la práctica (HU-062,
// "límite verificable"), pero el límite existe para que la consulta nunca
// dependa de que el llamador se comporte bien.
const DailyAgendaLimit = 300

// DailyAgendaEntry es la proyección mínima de una cita que HU-062 expone
// (CA-062-05): sin teléfono, correo, nota ni customerId. Los snapshots de
// servicio son la fuente visible (DEC-004): esta lectura nunca hace join
// con el catálogo vigente.
type DailyAgendaEntry struct {
	ID                       string
	AttendeeName             string
	StartsAt                 time.Time
	EndsAt                   time.Time
	Status                   Status
	Origin                   Origin
	ServiceNameSnapshot      string
	DurationMinutesSnapshot  int
	PriceAmountCentsSnapshot int64
	CurrencySnapshot         string
}

// AgendaService implementa el caso de uso de lectura de HU-062: agenda
// diaria de un único barbero (DEC-074), calculada en la zona IANA de la
// barbería, con un turno nocturno visible en cada agenda diaria cuyo rango
// interseca su intervalo (DEC-075). Colabora con staff/shops exclusivamente
// mediante BarberPort/TimezonePort, mismo criterio que ManualBookingService
// frente a catalog/schedule/shops.
type AgendaService struct {
	repo      Repository
	barbers   BarberPort
	timezones TimezonePort
	clock     clock.Clock
}

// NewAgendaService construye el caso de uso a partir de sus cuatro
// colaboradores.
func NewAgendaService(repo Repository, barbers BarberPort, timezones TimezonePort, clk clock.Clock) *AgendaService {
	return &AgendaService{repo: repo, barbers: barbers, timezones: timezones, clock: clk}
}

// ListDailyAgenda devuelve, ordenadas por StartsAt y luego por ID, las citas
// del barbero barberID cuyo intervalo interseca el rango civil de
// dateLocal (DEC-075). dateLocal vacío significa "hoy" en la zona de la
// barbería, resuelto a partir de s.clock.Now() (nunca time.Now
// directamente, docs/03-desarrollo/estandar-backend-go.md). Un barberID con
// forma inválida, inexistente o de otra barbería responde el mismo
// apperr.NotFound (DEC-074, RN-TEN-01), verificado ANTES de tocar el
// repositorio (mismo criterio que schedule.Service.List).
func (s *AgendaService) ListDailyAgenda(ctx context.Context, barbershopID, barberID, dateLocal string) ([]DailyAgendaEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, apperr.Internal(fmt.Errorf("booking: contexto cancelado antes de listar la agenda diaria: %w", err))
	}

	barberID = strings.TrimSpace(barberID)
	if !LooksLikeBarberID(barberID) {
		return nil, errBarberNotFound()
	}

	exists, err := s.barbers.Exists(ctx, barbershopID, barberID)
	if err != nil {
		return nil, apperr.Internal(fmt.Errorf("booking: verificar barbero para la agenda diaria: %w", err))
	}
	if !exists {
		return nil, errBarberNotFound()
	}

	timezone, err := s.timezones.Timezone(ctx, barbershopID)
	if err != nil {
		return nil, err
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return nil, errTimezoneUnresolvable()
	}

	day, err := resolveCivilDate(dateLocal, loc, s.clock.Now())
	if err != nil {
		return nil, err
	}

	// time.Date normaliza el desbordamiento de Day() (día 32 de agosto se
	// convierte en 1 de septiembre) usando la ubicación civil de loc, nunca
	// sumando una duración fija de 24 horas: un día local de 23 o 25 horas
	// por un cambio de horario de verano queda correctamente representado
	// (RN-DIS-07).
	rangeStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, loc)
	rangeEnd := time.Date(day.Year(), day.Month(), day.Day()+1, 0, 0, 0, 0, loc)

	entries, err := s.repo.ListDailyAgenda(ctx, barbershopID, barberID, rangeStart, rangeEnd)
	if err != nil {
		return nil, apperr.Internal(fmt.Errorf("booking: listar agenda diaria: %w", err))
	}
	return entries, nil
}

// resolveCivilDate interpreta raw como fecha civil en loc. raw vacío
// devuelve "hoy": now (ya resuelto por el reloj inyectado) convertido a la
// zona de la barbería. Un raw no vacío que no cumple
// dailyAgendaDateLayout, o que trae hora/zona, se rechaza.
func resolveCivilDate(raw string, loc *time.Location, now time.Time) (time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return now.In(loc), nil
	}
	parsed, err := time.ParseInLocation(dailyAgendaDateLayout, trimmed, loc)
	if err != nil {
		return time.Time{}, errDateInvalid()
	}
	return parsed, nil
}
