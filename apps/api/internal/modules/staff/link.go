package staff

import (
	"context"
	"fmt"

	"system-barbershop/internal/platform/apperr"
)

// LinkResult es el desenlace de declarar «este barbero soy yo» (DEC-100).
//   - Found false: el barbero no existe o es de otra barbería (CA-021-05).
//   - Taken true: el barbero ya está vinculado a OTRO usuario; no se cambió nada.
//   - ReleasedBarberID: barbero que el usuario tenía antes y que esta
//     operación liberó (vacío si no tenía vínculo o era el mismo barbero).
type LinkResult struct {
	Barber           Barber
	Found            bool
	Taken            bool
	ReleasedBarberID string
}

// LinkObserver es el puerto de aviso de «vínculo cambiado o quitado»: otros
// módulos (la integración con Google Calendar, DEC-099) lo implementan para
// reaccionar cuando un barbero deja de pertenecer a un usuario, y lo
// registran en la raíz de composición (cmd/api); staff nunca los importa.
//
// Se invoca DESPUÉS de confirmar la transacción y sin garantía de entrega: el
// vínculo ya cambió aunque el observador falle, así que un consumidor que
// necesite consistencia estricta debe además reconciliar al leer.
type LinkObserver interface {
	BarberLinkReleased(ctx context.Context, barbershopID, barberID string)
}

// ObserveLinkReleased registra un observador del aviso de vínculo liberado.
// Se llama durante la composición, antes de atender solicitudes.
func (s *Service) ObserveLinkReleased(observer LinkObserver) {
	s.observers = append(s.observers, observer)
}

func (s *Service) notifyReleased(ctx context.Context, barbershopID, barberID string) {
	if barberID == "" {
		return
	}
	for _, observer := range s.observers {
		observer.BarberLinkReleased(ctx, barbershopID, barberID)
	}
}

// errNoLinkedBarber es el 404 de «no tienes un barbero vinculado»: el vínculo
// es opcional (DEC-100), no un fallo.
func errNoLinkedBarber() error {
	return apperr.NotFound("el usuario no tiene un barbero vinculado")
}

// errBarberTaken es el 409 de «ese barbero ya es de otro usuario». No revela
// quién es: el mensaje es el mismo sea cual sea el otro usuario.
func errBarberTaken() error {
	return apperr.Conflict("ese barbero ya está vinculado a otro usuario")
}

// LinkedBarber devuelve el barbero del usuario autenticado. staffUserID llega
// siempre de un auth.Principal ya autenticado: nunca de la solicitud (DEC-100).
func (s *Service) LinkedBarber(ctx context.Context, barbershopID, staffUserID string) (Barber, error) {
	if err := ctx.Err(); err != nil {
		return Barber{}, apperr.Internal(fmt.Errorf("staff: contexto cancelado antes de leer el barbero propio: %w", err))
	}
	barber, found, err := s.repo.GetLinked(ctx, barbershopID, staffUserID)
	if err != nil {
		return Barber{}, apperr.Internal(fmt.Errorf("staff: leer barbero propio: %w", err))
	}
	if !found {
		return Barber{}, errNoLinkedBarber()
	}
	return barber, nil
}

// LinkMyBarber vincula al usuario autenticado con barberID, liberando el
// barbero que tuviera antes. Es idempotente: repetir la misma selección deja
// el mismo estado. Un barbero ya vinculado a otro usuario responde conflicto
// y conserva el vínculo anterior del solicitante.
func (s *Service) LinkMyBarber(ctx context.Context, barbershopID, staffUserID, barberID string) (Barber, error) {
	if err := ctx.Err(); err != nil {
		return Barber{}, apperr.Internal(fmt.Errorf("staff: contexto cancelado antes de vincular el barbero propio: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return Barber{}, errBarberNotFound()
	}

	result, err := s.repo.Link(ctx, barbershopID, staffUserID, barberID)
	if err != nil {
		return Barber{}, apperr.Internal(fmt.Errorf("staff: vincular barbero propio: %w", err))
	}
	if !result.Found {
		return Barber{}, errBarberNotFound()
	}
	if result.Taken {
		return Barber{}, errBarberTaken()
	}
	s.notifyReleased(ctx, barbershopID, result.ReleasedBarberID)
	return result.Barber, nil
}

// UnlinkMyBarber quita el vínculo del usuario autenticado. Sin vínculo es un
// éxito: la operación es idempotente. No borra ni desactiva al barbero.
func (s *Service) UnlinkMyBarber(ctx context.Context, barbershopID, staffUserID string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("staff: contexto cancelado antes de quitar el vínculo: %w", err))
	}
	released, err := s.repo.Unlink(ctx, barbershopID, staffUserID)
	if err != nil {
		return apperr.Internal(fmt.Errorf("staff: quitar vínculo del barbero propio: %w", err))
	}
	s.notifyReleased(ctx, barbershopID, released)
	return nil
}

// UserBarberLookup adapta *Service a un puerto mínimo que otro módulo puede
// consumir para resolver «el barbero del usuario autenticado» sin importar el
// núcleo de staff (mismo patrón que BarberLookup).
type UserBarberLookup struct {
	service *Service
}

// NewUserBarberLookup construye el adaptador sobre el *Service compartido.
func NewUserBarberLookup(service *Service) UserBarberLookup {
	return UserBarberLookup{service: service}
}

// BarberIDOfUser devuelve el id del barbero vinculado a staffUserID. found es
// false si el usuario no tiene barbero (no es un error: el vínculo es opcional).
func (l UserBarberLookup) BarberIDOfUser(ctx context.Context, barbershopID, staffUserID string) (barberID string, found bool, err error) {
	barber, err := l.service.LinkedBarber(ctx, barbershopID, staffUserID)
	if err != nil {
		if appErr, ok := apperr.As(err); ok && appErr.Kind == apperr.KindNotFound {
			return "", false, nil
		}
		return "", false, err
	}
	return barber.ID, true, nil
}
