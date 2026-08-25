package catalog

import (
	"context"
	"fmt"
	"unicode/utf8"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// CatalogService implementa los cuatro casos de uso de HU-022: listar,
// consultar, crear y editar servicios del catálogo de la barbería activa.
// Se nombra CatalogService (no solo Service, a diferencia de
// staff.Service/shops.Service) porque el propio recurso de este módulo ya
// se llama Service: un tipo del paquete no puede llamarse igual que otro.
type CatalogService struct {
	repo Repository
}

// NewService construye el servicio de casos de uso.
func NewService(repo Repository) *CatalogService {
	return &CatalogService{repo: repo}
}

// List lee una página de servicios de barbershopID. cursorToken es el valor
// opaco que el cliente envió (vacío para la primera página); limit llega
// crudo del parámetro de consulta y se clamped aquí a
// [MinListLimit, MaxListLimit], con DefaultListLimit cuando el cliente no
// lo especifica (limit <= 0). barbershopID llega siempre de un
// auth.Principal ya autenticado (RN-TEN-01).
func (s *CatalogService) List(ctx context.Context, barbershopID string, cursorToken string, limit int) (ListResult, error) {
	if err := ctx.Err(); err != nil {
		return ListResult{}, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de listar servicios: %w", err))
	}

	switch {
	case limit <= 0:
		limit = DefaultListLimit
	case limit < MinListLimit:
		limit = MinListLimit
	case limit > MaxListLimit:
		limit = MaxListLimit
	}

	var cursor *Cursor
	if cursorToken != "" {
		decoded, err := DecodeCursor(cursorToken)
		if err != nil {
			return ListResult{}, err
		}
		cursor = &decoded
	}

	result, err := s.repo.List(ctx, barbershopID, cursor, limit)
	if err != nil {
		return ListResult{}, apperr.Internal(fmt.Errorf("catalog: listar servicios: %w", err))
	}
	return result, nil
}

// Get lee un servicio por id dentro de la barbería activa. Un identificador
// inexistente o de otra barbería produce el mismo apperr.NotFound
// (CA-022-06).
func (s *CatalogService) Get(ctx context.Context, barbershopID, serviceID string) (Service, error) {
	if err := ctx.Err(); err != nil {
		return Service{}, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de leer servicio: %w", err))
	}
	if !LooksLikeServiceID(serviceID) {
		return Service{}, errServiceNotFound()
	}

	service, found, err := s.repo.Get(ctx, barbershopID, serviceID)
	if err != nil {
		return Service{}, apperr.Internal(fmt.Errorf("catalog: leer servicio: %w", err))
	}
	if !found {
		return Service{}, errServiceNotFound()
	}
	return service, nil
}

// CreateInputRaw es el cuerpo crudo de un alta, sin normalizar ni validar
// todavía (tal como llega del handler HTTP).
type CreateInputRaw struct {
	Name            string
	Description     string
	DurationMinutes int
	Price           string
}

// Create registra un servicio (CA-022-02), protegido por el protocolo de
// idempotencia reutilizable de HU-004 (RN-IDE-01, DEC-043). key y
// fingerprint ya fueron interpretados por la capa HTTP. La validación de
// campos ocurre ANTES de tocar el repositorio: una entrada inválida nunca
// reclama ni consume la clave de idempotencia (mismo criterio que
// staff.Service.Create); el único caso que SÍ requiere tocar el
// repositorio para descubrirse es el nombre duplicado (DEC-067), porque
// depende de otra fila ya persistida, no de la forma del campo.
func (s *CatalogService) Create(
	ctx context.Context,
	barbershopID string,
	raw CreateInputRaw,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CreateResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateResult{}, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de crear servicio: %w", err))
	}

	input, err := validateCreateInput(raw)
	if err != nil {
		return CreateResult{}, err
	}

	result, err := s.repo.Create(ctx, barbershopID, input, key, fingerprint)
	if err != nil {
		return CreateResult{}, apperr.Internal(fmt.Errorf("catalog: crear servicio: %w", err))
	}
	if result.NameTaken {
		return CreateResult{}, errNameConflict()
	}
	return result, nil
}

// UpdateInputRaw es el cuerpo crudo de una edición parcial, tal como llega
// del handler HTTP. Cada puntero nil significa "el cliente no envió este
// campo"; DescriptionSet distingue "no envió description" de "la envió
// (posiblemente vacía, para borrarla)".
type UpdateInputRaw struct {
	Name            *string
	DescriptionSet  bool
	Description     *string
	DurationMinutes *int
	Price           *string
}

// Update cambia los campos de catálogo autorizados de un servicio existente
// de la barbería activa (CA-022-04, CA-022-05). Un identificador inexistente
// o de otra barbería produce el mismo apperr.NotFound que Get (CA-022-06);
// un objeto sin ningún campo se rechaza sin tocar el repositorio
// (errUpdateEmptyBody); un nombre que choque con otro servicio activo de la
// misma barbería produce apperr.Conflict (DEC-067). Ningún campo de esta
// operación toca citas ni simula propagación a recursos futuros (RN-SER-04).
func (s *CatalogService) Update(ctx context.Context, barbershopID, serviceID string, raw UpdateInputRaw) (Service, error) {
	if err := ctx.Err(); err != nil {
		return Service{}, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de editar servicio: %w", err))
	}
	if !LooksLikeServiceID(serviceID) {
		return Service{}, errServiceNotFound()
	}

	fields, err := validateUpdateFields(raw)
	if err != nil {
		return Service{}, err
	}
	if !fields.HasAny() {
		return Service{}, errUpdateEmptyBody()
	}

	result, err := s.repo.Update(ctx, barbershopID, serviceID, fields)
	if err != nil {
		return Service{}, apperr.Internal(fmt.Errorf("catalog: editar servicio: %w", err))
	}
	if result.NameTaken {
		return Service{}, errNameConflict()
	}
	if !result.Found {
		return Service{}, errServiceNotFound()
	}
	return result.Service, nil
}

// PreviewDeactivation calcula el impacto real de desactivar serviceID ahora
// mismo (CA-024-01), sin cambiar ningún dato: solo confirma que el servicio
// existe en barbershopID (mismo apperr.NotFound uniforme que Get) y
// devuelve currentDeactivationImpact(), siempre 0 en B1 (DEC-069).
func (s *CatalogService) PreviewDeactivation(ctx context.Context, barbershopID, serviceID string) (DeactivationImpact, error) {
	if err := ctx.Err(); err != nil {
		return DeactivationImpact{}, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de previsualizar desactivación: %w", err))
	}
	if !LooksLikeServiceID(serviceID) {
		return DeactivationImpact{}, errServiceNotFound()
	}

	_, found, err := s.repo.Get(ctx, barbershopID, serviceID)
	if err != nil {
		return DeactivationImpact{}, apperr.Internal(fmt.Errorf("catalog: leer servicio para previsualizar desactivación: %w", err))
	}
	if !found {
		return DeactivationImpact{}, errServiceNotFound()
	}
	return currentDeactivationImpact(), nil
}

// Deactivate transiciona serviceID de activo a inactivo (CA-024-02,
// CA-024-03), protegido por el protocolo de idempotencia reutilizable de
// HU-004 (RN-IDE-01, DEC-043). key y fingerprint ya fueron interpretados
// por la capa HTTP. Un identificador inexistente o de otra barbería produce
// apperr.NotFound (CA-024-07); un servicio ya inactivo bajo una clave nueva
// produce apperr.Conflict (CA-024-06).
func (s *CatalogService) Deactivate(
	ctx context.Context,
	barbershopID, serviceID string,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (LifecycleResult, error) {
	if err := ctx.Err(); err != nil {
		return LifecycleResult{}, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de desactivar servicio: %w", err))
	}
	if !LooksLikeServiceID(serviceID) {
		return LifecycleResult{}, errServiceNotFound()
	}

	result, err := s.repo.Deactivate(ctx, barbershopID, serviceID, key, fingerprint)
	if err != nil {
		return LifecycleResult{}, apperr.Internal(fmt.Errorf("catalog: desactivar servicio: %w", err))
	}
	if result.Decision.Outcome == idempotency.OutcomeProceed {
		if !result.Found {
			return LifecycleResult{}, errServiceNotFound()
		}
		if result.InvalidTransition {
			return LifecycleResult{}, errServiceAlreadyInactive()
		}
	}
	return result, nil
}

// Reactivate transiciona serviceID de inactivo a activo (CA-024-05), mismo
// protocolo de idempotencia que Deactivate, en sentido inverso.
func (s *CatalogService) Reactivate(
	ctx context.Context,
	barbershopID, serviceID string,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (LifecycleResult, error) {
	if err := ctx.Err(); err != nil {
		return LifecycleResult{}, apperr.Internal(fmt.Errorf("catalog: contexto cancelado antes de reactivar servicio: %w", err))
	}
	if !LooksLikeServiceID(serviceID) {
		return LifecycleResult{}, errServiceNotFound()
	}

	result, err := s.repo.Reactivate(ctx, barbershopID, serviceID, key, fingerprint)
	if err != nil {
		return LifecycleResult{}, apperr.Internal(fmt.Errorf("catalog: reactivar servicio: %w", err))
	}
	if result.Decision.Outcome == idempotency.OutcomeProceed {
		if !result.Found {
			return LifecycleResult{}, errServiceNotFound()
		}
		if result.InvalidTransition {
			return LifecycleResult{}, errServiceAlreadyActive()
		}
	}
	return result, nil
}

// validateName recorta y valida name contra CA-022-04: vacío, solo espacios
// o mayor de NameMaxLength caracteres se rechaza sin tocar el repositorio.
// utf8.RuneCountInString cuenta caracteres Unicode, no bytes.
func validateName(raw string) (string, error) {
	name := NormalizeName(raw)
	switch {
	case name == "":
		return "", errNameRequired()
	case utf8.RuneCountInString(name) > NameMaxLength:
		return "", errNameTooLong()
	}
	return name, nil
}

// validateDescription recorta y valida raw contra DescriptionMaxLength. Un
// resultado vacío tras recortar se normaliza a nil (sin descripción).
func validateDescription(raw string) (*string, error) {
	description := NormalizeDescription(raw)
	if description != nil && utf8.RuneCountInString(*description) > DescriptionMaxLength {
		return nil, errDescriptionTooLong()
	}
	return description, nil
}

// validateDuration valida que minutes esté en [MinDurationMinutes,
// MaxDurationMinutes] (CA-022-03): duraciones como 25, 30, 45 o 90 son
// igual de válidas que cualquier otro entero en ese rango.
func validateDuration(minutes int) (int, error) {
	if minutes < MinDurationMinutes || minutes > MaxDurationMinutes {
		return 0, errDurationOutOfRange()
	}
	return minutes, nil
}

// validateCreateInput normaliza y valida los cuatro campos de un alta, en
// el orden fijo del contrato (name, description, durationMinutes, price):
// el primer error de campo encontrado es el que se devuelve, sin ejecutar
// ninguna escritura.
func validateCreateInput(raw CreateInputRaw) (CreateInput, error) {
	name, err := validateName(raw.Name)
	if err != nil {
		return CreateInput{}, err
	}
	description, err := validateDescription(raw.Description)
	if err != nil {
		return CreateInput{}, err
	}
	duration, err := validateDuration(raw.DurationMinutes)
	if err != nil {
		return CreateInput{}, err
	}
	priceCents, err := ParsePriceCOP(raw.Price)
	if err != nil {
		return CreateInput{}, err
	}
	return CreateInput{
		Name:            name,
		Description:     description,
		DurationMinutes: duration,
		PriceCents:      priceCents,
	}, nil
}

// validateUpdateFields normaliza y valida solo los campos presentes en raw,
// mismo orden fijo que validateCreateInput. Un campo ausente en raw queda
// ausente en el resultado (UpdateFields.HasAny lo refleja).
func validateUpdateFields(raw UpdateInputRaw) (UpdateFields, error) {
	var fields UpdateFields

	if raw.Name != nil {
		name, err := validateName(*raw.Name)
		if err != nil {
			return UpdateFields{}, err
		}
		fields.Name = &name
	}

	if raw.DescriptionSet {
		description, err := validateDescription(derefOr(raw.Description, ""))
		if err != nil {
			return UpdateFields{}, err
		}
		fields.Description = OptionalDescription{Set: true, Value: description}
	}

	if raw.DurationMinutes != nil {
		duration, err := validateDuration(*raw.DurationMinutes)
		if err != nil {
			return UpdateFields{}, err
		}
		fields.DurationMinutes = &duration
	}

	if raw.Price != nil {
		priceCents, err := ParsePriceCOP(*raw.Price)
		if err != nil {
			return UpdateFields{}, err
		}
		fields.PriceCents = &priceCents
	}

	return fields, nil
}

func derefOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}
