package staff

import (
	"context"
	"fmt"
	"unicode/utf8"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/idempotency"
)

// Service implementa los casos de uso de HU-021: listar, consultar,
// registrar y renombrar barberos de la barbería activa; y, por DEC-104, subir,
// leer y quitar su fotografía opcional.
type Service struct {
	repo      Repository
	observers []LinkObserver
}

// NewService construye el servicio.
func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// List lee una página de barberos de barbershopID. cursorToken es el valor
// opaco que el cliente envió (vacío para la primera página); limit llega
// crudo del parámetro de consulta y se clamped aquí a
// [MinListLimit, MaxListLimit], con DefaultListLimit cuando el cliente no
// lo especifica (limit <= 0). barbershopID llega siempre de un
// auth.Principal ya autenticado (RN-TEN-01); este método no acepta ni
// valida ningún identificador que el cliente pueda controlar.
func (s *Service) List(ctx context.Context, barbershopID string, cursorToken string, limit int) (ListResult, error) {
	if err := ctx.Err(); err != nil {
		return ListResult{}, apperr.Internal(fmt.Errorf("staff: contexto cancelado antes de listar barberos: %w", err))
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
		return ListResult{}, apperr.Internal(fmt.Errorf("staff: listar barberos: %w", err))
	}
	return result, nil
}

// Get lee un barbero por id dentro de la barbería activa. Un identificador
// inexistente o de otra barbería produce el mismo apperr.NotFound
// (CA-021-05).
func (s *Service) Get(ctx context.Context, barbershopID, barberID string) (Barber, error) {
	if err := ctx.Err(); err != nil {
		return Barber{}, apperr.Internal(fmt.Errorf("staff: contexto cancelado antes de leer barbero: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		// Nunca deja que un identificador con forma inválida llegue a
		// PostgreSQL como comparación contra una columna uuid (eso sería un
		// error de tipo, 500, no el 404 uniforme de CA-021-05).
		return Barber{}, errBarberNotFound()
	}

	barber, found, err := s.repo.Get(ctx, barbershopID, barberID)
	if err != nil {
		return Barber{}, apperr.Internal(fmt.Errorf("staff: leer barbero: %w", err))
	}
	if !found {
		return Barber{}, errBarberNotFound()
	}
	return barber, nil
}

// Create registra un barbero (CA-021-02), protegido por el protocolo de
// idempotencia reutilizable de HU-004 (RN-IDE-01, DEC-043). fullNameRaw es
// el valor crudo del cuerpo, sin normalizar; key y fingerprint ya fueron
// interpretados por la capa HTTP (httpserver.IdempotencyKeyFromRequest/
// IdempotencyFingerprint). La validación de fullName ocurre ANTES de tocar
// el repositorio: un nombre inválido nunca reclama ni consume la clave de
// idempotencia, porque no es el efecto que RN-IDE-01 protege.
func (s *Service) Create(
	ctx context.Context,
	barbershopID string,
	fullNameRaw string,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (CreateResult, error) {
	if err := ctx.Err(); err != nil {
		return CreateResult{}, apperr.Internal(fmt.Errorf("staff: contexto cancelado antes de registrar barbero: %w", err))
	}

	fullName, err := validateFullName(fullNameRaw)
	if err != nil {
		return CreateResult{}, err
	}

	result, err := s.repo.Create(ctx, barbershopID, fullName, key, fingerprint)
	if err != nil {
		return CreateResult{}, apperr.Internal(fmt.Errorf("staff: registrar barbero: %w", err))
	}
	return result, nil
}

// Rename cambia el nombre de un barbero existente de la barbería activa
// (CA-021-04). fullNameRaw es el valor crudo del cuerpo, sin normalizar. Un
// identificador inexistente o de otra barbería produce el mismo
// apperr.NotFound que Get (CA-021-05); nombres duplicados dentro de la
// misma barbería son válidos (esta historia no impone unicidad).
func (s *Service) Rename(ctx context.Context, barbershopID, barberID, fullNameRaw string) (Barber, error) {
	if err := ctx.Err(); err != nil {
		return Barber{}, apperr.Internal(fmt.Errorf("staff: contexto cancelado antes de renombrar barbero: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return Barber{}, errBarberNotFound()
	}

	fullName, err := validateFullName(fullNameRaw)
	if err != nil {
		return Barber{}, err
	}

	result, err := s.repo.Rename(ctx, barbershopID, barberID, fullName)
	if err != nil {
		return Barber{}, apperr.Internal(fmt.Errorf("staff: renombrar barbero: %w", err))
	}
	if !result.Found {
		return Barber{}, errBarberNotFound()
	}
	return result.Barber, nil
}

// SetPhoto sube o reemplaza la fotografía de un barbero (DEC-104). declaredType
// es el Content-Type ya normalizado del cliente y data el cuerpo crudo: la
// validación (formato real, tamaño, dimensiones) ocurre ANTES de tocar el
// repositorio, así que una imagen inválida nunca llega a la base. Un barbero
// inexistente o de otra barbería produce el mismo apperr.NotFound que Get.
func (s *Service) SetPhoto(ctx context.Context, barbershopID, barberID, declaredType string, data []byte) (Barber, error) {
	if err := ctx.Err(); err != nil {
		return Barber{}, apperr.Internal(fmt.Errorf("staff: contexto cancelado antes de guardar la fotografía: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return Barber{}, errBarberNotFound()
	}

	photo, err := ValidatePhoto(declaredType, data)
	if err != nil {
		return Barber{}, err
	}

	result, err := s.repo.PutPhoto(ctx, barbershopID, barberID, photo)
	if err != nil {
		return Barber{}, apperr.Internal(fmt.Errorf("staff: guardar fotografía: %w", err))
	}
	if !result.Found {
		return Barber{}, errBarberNotFound()
	}
	return result.Barber, nil
}

// Photo lee la fotografía de un barbero. Un barbero inexistente, de otra
// barbería o sin fotografía produce el mismo apperr.NotFound (CA-021-05).
func (s *Service) Photo(ctx context.Context, barbershopID, barberID string) (StoredPhoto, error) {
	if err := ctx.Err(); err != nil {
		return StoredPhoto{}, apperr.Internal(fmt.Errorf("staff: contexto cancelado antes de leer la fotografía: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return StoredPhoto{}, errBarberNotFound()
	}

	photo, found, err := s.repo.GetPhoto(ctx, barbershopID, barberID)
	if err != nil {
		return StoredPhoto{}, apperr.Internal(fmt.Errorf("staff: leer fotografía: %w", err))
	}
	if !found {
		return StoredPhoto{}, errBarberNotFound()
	}
	return photo, nil
}

// RemovePhoto quita la fotografía de un barbero. Es idempotente: quitar una
// fotografía inexistente no es un error mientras el barbero exista.
func (s *Service) RemovePhoto(ctx context.Context, barbershopID, barberID string) error {
	if err := ctx.Err(); err != nil {
		return apperr.Internal(fmt.Errorf("staff: contexto cancelado antes de quitar la fotografía: %w", err))
	}
	if !LooksLikeBarberID(barberID) {
		return errBarberNotFound()
	}

	found, err := s.repo.DeletePhoto(ctx, barbershopID, barberID)
	if err != nil {
		return apperr.Internal(fmt.Errorf("staff: quitar fotografía: %w", err))
	}
	if !found {
		return errBarberNotFound()
	}
	return nil
}

// validateFullName recorta y valida fullNameRaw contra CA-021-03: vacío,
// solo espacios o mayor de FullNameMaxLength caracteres se rechaza sin
// tocar el repositorio. utf8.RuneCountInString cuenta caracteres Unicode,
// no bytes, para que un nombre con tildes o ñ no se rechace de forma
// prematura por el tamaño en bytes de su codificación UTF-8.
func validateFullName(raw string) (string, error) {
	fullName := NormalizeFullName(raw)
	switch {
	case fullName == "":
		return "", errFullNameRequired()
	case utf8.RuneCountInString(fullName) > FullNameMaxLength:
		return "", errFullNameTooLong()
	}
	return fullName, nil
}

// ListPage permite saltar a una página de la tabla sin descargar todo el equipo.
func (s *Service) ListPage(ctx context.Context, shop string, page, size int) (PageResult, error) {
	if err := ctx.Err(); err != nil {
		return PageResult{}, apperr.Internal(err)
	}
	if page < 1 {
		page = 1
	}
	if size <= 0 {
		size = DefaultListLimit
	}
	if size > MaxListLimit {
		size = MaxListLimit
	}
	result, err := s.repo.ListPage(ctx, shop, page, size)
	if err != nil {
		return PageResult{}, apperr.Internal(fmt.Errorf("staff: listar página: %w", err))
	}
	return result, nil
}
