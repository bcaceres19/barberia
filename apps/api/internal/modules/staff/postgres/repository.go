// Package postgres es el adaptador de persistencia de staff.Repository
// sobre database.DB (CA-002-06: el núcleo de staff no importa este paquete
// ni pgx; es este paquete el que importa staff), mismo patrón que
// shops/postgres.
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"system-barbershop/internal/modules/staff"
	"system-barbershop/internal/platform/database"
	"system-barbershop/internal/platform/idempotency"
)

// createBarberOperation identifica, dentro de una barbería, la operación
// de idempotencia del alta de barberos (RN-IDE-01): una clave ya usada
// para OTRA operación (por ejemplo una futura "crear cita") nunca se
// confunde con esta.
const createBarberOperation idempotency.Operation = "create_barber"

// createBarberIdempotencyTTL es la vigencia de una reclamación de
// idempotencia sobre el alta de un barbero antes de tratarse como si nunca
// hubiera existido (CA-004-05). 24 horas cubre con margen cualquier
// reintento de red razonable sin dejar el mecanismo vigente indefinidamente.
const createBarberIdempotencyTTL = 24 * time.Hour

// Repository implementa staff.Repository sobre database.DB.
type Repository struct {
	db    *database.DB
	coord idempotency.Coordinator
}

// New construye el repositorio con el coordinador de idempotencia real
// (SQLCoordinator). El parámetro coord existe para pruebas: internal/
// platform/idempotency ya prueba SQLCoordinator de forma exhaustiva contra
// PostgreSQL real, así que este paquete no necesita repetir esas pruebas,
// pero sí necesita poder construir el repositorio con el mismo adaptador
// que producción usa.
func New(db *database.DB, coord idempotency.Coordinator) *Repository {
	return &Repository{db: db, coord: coord}
}

var _ staff.Repository = (*Repository)(nil)

// barberResponseWire es la forma JSON EXACTA que también usa
// httpapi.BarberResponse (mismos nombres de campo, mismo orden, mismos
// tipos): la respuesta que Complete persiste para una repetición exacta
// (CA-004-01) debe ser BYTE A BYTE idéntica a la que el handler ya
// construyó para la primera ejecución, y ambas deben tener EXACTAMENTE la
// misma forma que un GET/PATCH de este mismo barbero. Si cambia
// httpapi.BarberResponse, este struct debe cambiar igual en el mismo
// commit (ver TestCreate_StoredResponseBody_MatchesHTTPAPIWireShape en
// repository_test.go).
type barberResponseWire struct {
	ID             string     `json:"id"`
	FullName       string     `json:"fullName"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
	PhotoUpdatedAt *time.Time `json:"photoUpdatedAt"`
}

func marshalBarberResponse(b staff.Barber) ([]byte, error) {
	return json.Marshal(barberResponseWire{
		ID:             b.ID,
		FullName:       b.FullName,
		CreatedAt:      b.CreatedAt,
		UpdatedAt:      b.UpdatedAt,
		PhotoUpdatedAt: b.PhotoUpdatedAt,
	})
}

// barberColumns/barberPhotoJoin componen la lectura de un barbero con la
// versión de su fotografía (DEC-104). LEFT JOIN a barber_photo por la PK
// compuesta (barbershop_id, barber_id): nunca lee los bytes de `image`, solo
// `updated_at`, así que el listado del equipo no arrastra imágenes.
const (
	barberColumns   = `b.id, b.full_name, b.created_at, b.updated_at, p.updated_at`
	barberPhotoJoin = `LEFT JOIN barber_photo p
	                     ON p.barbershop_id = b.barbershop_id AND p.barber_id = b.id`
)

// List implementa staff.Repository.List: orden estable (created_at, id)
// dentro del tenant vigente, cursor opaco decodificado por el núcleo
// (staff.Cursor), paginación "pedir uno de más" para saber si hay página
// siguiente sin una segunda consulta COUNT.
func (r *Repository) List(ctx context.Context, barbershopID string, cursor *staff.Cursor, limit int) (staff.ListResult, error) {
	var result staff.ListResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var (
			rows pgx.Rows
			err  error
		)
		// fetchLimit pide una fila de más: si vuelve, hay página siguiente.
		fetchLimit := limit + 1

		if cursor == nil {
			rows, err = q.Query(ctx,
				`SELECT `+barberColumns+`
				   FROM barber b `+barberPhotoJoin+`
				  WHERE b.barbershop_id = $1
				  ORDER BY b.created_at, b.id
				  LIMIT $2`,
				barbershopID, fetchLimit,
			)
		} else {
			rows, err = q.Query(ctx,
				`SELECT `+barberColumns+`
				   FROM barber b `+barberPhotoJoin+`
				  WHERE b.barbershop_id = $1
				    AND (b.created_at, b.id) > ($2, $3)
				  ORDER BY b.created_at, b.id
				  LIMIT $4`,
				barbershopID, cursor.CreatedAt, cursor.ID, fetchLimit,
			)
		}
		if err != nil {
			return fmt.Errorf("list barbers: query: %w", err)
		}
		defer rows.Close()

		items := make([]staff.Barber, 0, fetchLimit)
		for rows.Next() {
			var b staff.Barber
			if err := rows.Scan(&b.ID, &b.FullName, &b.CreatedAt, &b.UpdatedAt, &b.PhotoUpdatedAt); err != nil {
				return fmt.Errorf("list barbers: scan: %w", err)
			}
			items = append(items, b)
		}
		if err := rows.Err(); err != nil {
			return fmt.Errorf("list barbers: rows: %w", err)
		}

		hasMore := len(items) > limit
		if hasMore {
			items = items[:limit]
		}

		result.Items = items
		if hasMore {
			last := items[len(items)-1]
			result.NextCursor = staff.EncodeCursor(staff.Cursor{CreatedAt: last.CreatedAt, ID: last.ID})
		}
		return nil
	})
	if err != nil {
		return staff.ListResult{}, fmt.Errorf("staff/postgres: list barbers: %w", err)
	}
	return result, nil
}

// Get implementa staff.Repository.Get. RLS (barber_select_tenant_policy) ya
// restringe la fila visible a barbershop_id = current_setting(...); el
// filtro explícito WHERE barbershop_id = $2 es defensa en profundidad,
// mismo patrón que shops/postgres.Repository.Get.
func (r *Repository) Get(ctx context.Context, barbershopID, barberID string) (staff.Barber, bool, error) {
	var b staff.Barber
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := selectBarber(ctx, q, barbershopID, barberID).
			Scan(&b.ID, &b.FullName, &b.CreatedAt, &b.UpdatedAt, &b.PhotoUpdatedAt)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// found queda false: cubre tanto "no existe" como "es de otra
			// barbería" (CA-021-05), sin distinguir la causa.
		default:
			return fmt.Errorf("select barber: %w", err)
		}
		return nil
	})
	if err != nil {
		return staff.Barber{}, false, fmt.Errorf("staff/postgres: get barber: %w", err)
	}
	if !found {
		return staff.Barber{}, false, nil
	}
	return b, true, nil
}

// Create implementa staff.Repository.Create: Begin, INSERT y Complete
// ocurren dentro de la MISMA InTenantTx (apps/api/README.md, "Patrón
// obligatorio: idempotencia reutilizable"). Si Begin no devuelve
// OutcomeProceed, el callback no ejecuta ningún efecto y la transacción
// igual hace COMMIT (Begin ya dejó su propio registro, cuando aplica); si
// el efecto o Complete fallan, InTenantTx hace ROLLBACK de TODA la
// transacción, incluido el INSERT que Begin reclamó, dejando la clave libre
// para un reintento legítimo (CA-004-06).
func (r *Repository) Create(
	ctx context.Context,
	barbershopID string,
	fullName string,
	key idempotency.Key,
	fingerprint idempotency.Fingerprint,
) (staff.CreateResult, error) {
	var result staff.CreateResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		decision, err := r.coord.Begin(ctx, q, database.BarbershopID(barbershopID), key, createBarberOperation, fingerprint, createBarberIdempotencyTTL)
		if err != nil {
			return fmt.Errorf("idempotency begin: %w", err)
		}
		result.Decision = decision

		if decision.Outcome != idempotency.OutcomeProceed {
			if decision.Outcome == idempotency.OutcomeReplay {
				result.Response = decision.Response
			}
			return nil
		}

		var b staff.Barber
		if err := q.QueryRow(ctx,
			`INSERT INTO barber (barbershop_id, full_name)
			      VALUES ($1, $2)
			   RETURNING id, full_name, created_at, updated_at`,
			barbershopID, fullName,
		).Scan(&b.ID, &b.FullName, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return fmt.Errorf("insert barber: %w", err)
		}

		body, err := marshalBarberResponse(b)
		if err != nil {
			return fmt.Errorf("marshal created barber: %w", err)
		}
		stored := idempotency.StoredResponse{
			Status:      201,
			ContentType: "application/json",
			Body:        string(body),
		}

		ok, err := r.coord.Complete(ctx, q, database.BarbershopID(barbershopID), key, stored)
		if err != nil {
			return fmt.Errorf("idempotency complete: %w", err)
		}
		if !ok {
			// No debería ocurrir: Begin reclamó la clave en esta misma
			// transacción hace un instante; ver el comentario de
			// Coordinator.Complete para el único caso legítimo (otra
			// llamada ya la completó), que no puede pasar dentro de una
			// única transacción secuencial como esta.
			return fmt.Errorf("idempotency complete: la reclamación ya no estaba in_progress")
		}

		result.Barber = b
		result.Response = stored
		return nil
	})
	if err != nil {
		return staff.CreateResult{}, fmt.Errorf("staff/postgres: create barber: %w", err)
	}
	return result, nil
}

// Rename implementa staff.Repository.Rename: UPDATE ... RETURNING dentro de
// una única sentencia, filtrando por id Y barbershop_id (defensa en
// profundidad sobre barber_update_tenant_policy). Nunca crea ni duplica una
// fila (CA-021-04); cero filas afectadas cubre tanto "no existe" como "es
// de otra barbería" (CA-021-05).
func (r *Repository) Rename(ctx context.Context, barbershopID, barberID, fullName string) (staff.RenameResult, error) {
	var result staff.RenameResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var b staff.Barber
		err := q.QueryRow(ctx,
			`WITH renamed AS (
			    UPDATE barber
			       SET full_name = $3
			     WHERE id = $1 AND barbershop_id = $2
			 RETURNING id, barbershop_id, full_name, created_at, updated_at
			 )
			 SELECT b.id, b.full_name, b.created_at, b.updated_at, p.updated_at
			   FROM renamed b `+barberPhotoJoin,
			barberID, barbershopID, fullName,
		).Scan(&b.ID, &b.FullName, &b.CreatedAt, &b.UpdatedAt, &b.PhotoUpdatedAt)
		switch {
		case err == nil:
			result.Found = true
			result.Barber = b
		case errors.Is(err, pgx.ErrNoRows):
			// result.Found queda false: defensivo, ver RenameResult.Found.
		default:
			return fmt.Errorf("update barber: %w", err)
		}
		return nil
	})
	if err != nil {
		return staff.RenameResult{}, fmt.Errorf("staff/postgres: rename barber: %w", err)
	}
	return result, nil
}

// selectBarber lee un barbero con la versión de su fotografía dentro de la
// transacción del llamador. Filtra por id Y barbershop_id (defensa en
// profundidad sobre barber_select_tenant_policy).
func selectBarber(ctx context.Context, q database.Queries, barbershopID, barberID string) pgx.Row {
	return q.QueryRow(ctx,
		`SELECT `+barberColumns+`
		   FROM barber b `+barberPhotoJoin+`
		  WHERE b.id = $1 AND b.barbershop_id = $2`,
		barberID, barbershopID,
	)
}

// PutPhoto implementa staff.Repository.PutPhoto: upsert de la fila de
// barber_photo SOLO si el barbero existe en el tenant vigente (el INSERT ...
// SELECT no produce fila para uno inexistente o de otra barbería, CA-021-05) y
// lectura del barbero ya con la nueva versión, todo en la misma transacción.
// ON CONFLICT reemplaza la imagen sin duplicar la fila; el trigger
// barber_photo_set_updated_at renueva updated_at, que es la versión pública.
func (r *Repository) PutPhoto(ctx context.Context, barbershopID, barberID string, photo staff.Photo) (staff.PhotoResult, error) {
	var result staff.PhotoResult

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		tag, err := q.Exec(ctx,
			`INSERT INTO barber_photo (barbershop_id, barber_id, content_type, image)
			 SELECT b.barbershop_id, b.id, $3, $4
			   FROM barber b
			  WHERE b.id = $1 AND b.barbershop_id = $2
			 ON CONFLICT (barbershop_id, barber_id)
			 DO UPDATE SET content_type = EXCLUDED.content_type, image = EXCLUDED.image`,
			barberID, barbershopID, photo.ContentType, photo.Data,
		)
		if err != nil {
			return fmt.Errorf("upsert barber photo: %w", err)
		}
		if tag.RowsAffected() == 0 {
			return nil // result.Found queda false: barbero inexistente o de otra barbería.
		}

		var b staff.Barber
		if err := selectBarber(ctx, q, barbershopID, barberID).
			Scan(&b.ID, &b.FullName, &b.CreatedAt, &b.UpdatedAt, &b.PhotoUpdatedAt); err != nil {
			return fmt.Errorf("select barber after photo upsert: %w", err)
		}
		result.Found = true
		result.Barber = b
		return nil
	})
	if err != nil {
		return staff.PhotoResult{}, fmt.Errorf("staff/postgres: put barber photo: %w", err)
	}
	return result, nil
}

// GetPhoto implementa staff.Repository.GetPhoto.
func (r *Repository) GetPhoto(ctx context.Context, barbershopID, barberID string) (staff.StoredPhoto, bool, error) {
	var photo staff.StoredPhoto
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`SELECT content_type, image, updated_at
			   FROM barber_photo
			  WHERE barbershop_id = $1 AND barber_id = $2`,
			barbershopID, barberID,
		).Scan(&photo.ContentType, &photo.Data, &photo.UpdatedAt)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// found queda false: sin fotografía o barbero de otra barbería.
		default:
			return fmt.Errorf("select barber photo: %w", err)
		}
		return nil
	})
	if err != nil {
		return staff.StoredPhoto{}, false, fmt.Errorf("staff/postgres: get barber photo: %w", err)
	}
	if !found {
		return staff.StoredPhoto{}, false, nil
	}
	return photo, true, nil
}

// DeletePhoto implementa staff.Repository.DeletePhoto. Primero comprueba que
// el barbero exista en el tenant: así "sin fotografía" (éxito idempotente) y
// "barbero inexistente" (404) se distinguen sin depender de cuántas filas
// borró el DELETE.
func (r *Repository) DeletePhoto(ctx context.Context, barbershopID, barberID string) (bool, error) {
	found := false

	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var exists bool
		if err := q.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM barber WHERE id = $1 AND barbershop_id = $2)`,
			barberID, barbershopID,
		).Scan(&exists); err != nil {
			return fmt.Errorf("check barber exists: %w", err)
		}
		if !exists {
			return nil
		}
		found = true

		if _, err := q.Exec(ctx,
			`DELETE FROM barber_photo WHERE barbershop_id = $1 AND barber_id = $2`,
			barbershopID, barberID,
		); err != nil {
			return fmt.Errorf("delete barber photo: %w", err)
		}
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("staff/postgres: delete barber photo: %w", err)
	}
	return found, nil
}

// ListPage obtiene total, límite y filas en un mismo snapshot, incluso para un
// tenant vacío. El total nunca cuenta otras barberías (DEC-107, RN-TEN-01).
func (r *Repository) ListPage(ctx context.Context, shop string, page, size int) (staff.PageResult, error) {
	result := staff.PageResult{Items: []staff.Barber{}, PageSize: size}
	err := r.db.InTenantTx(ctx, database.BarbershopID(shop), func(ctx context.Context, q database.Queries) error {
		rows, err := q.Query(ctx, `WITH counted AS (
   SELECT count(*) AS total FROM barber WHERE barbershop_id = $1
  ), bounds AS (
   SELECT total, GREATEST(1, (total + $2 - 1) / $2) AS pages FROM counted
  ), selected AS (
   SELECT b.id, b.full_name, b.created_at, b.updated_at, p.updated_at AS photo_updated_at FROM barber b `+barberPhotoJoin+`
   WHERE b.barbershop_id = $1 ORDER BY b.created_at, b.id
   LIMIT $2 OFFSET (SELECT (LEAST($3, pages) - 1) * $2 FROM bounds)
  )
  SELECT total, pages, LEAST($3, pages), s.id, s.full_name,
         s.created_at, s.updated_at, s.photo_updated_at
  FROM bounds LEFT JOIN selected s ON true`, shop, size, page)
		if err != nil {
			return fmt.Errorf("list barber page: %w", err)
		}
		defer rows.Close()
		for rows.Next() {
			var id, name *string
			var created, updated, photo *time.Time
			if err := rows.Scan(&result.Total, &result.TotalPages, &result.Page, &id, &name, &created, &updated, &photo); err != nil {
				return err
			}
			if id != nil {
				result.Items = append(result.Items, staff.Barber{ID: *id, FullName: *name, CreatedAt: *created, UpdatedAt: *updated, PhotoUpdatedAt: photo})
			}
		}
		return rows.Err()
	})
	return result, err
}

// GetLinked implementa staff.Repository.GetLinked.
func (r *Repository) GetLinked(ctx context.Context, barbershopID, staffUserID string) (staff.Barber, bool, error) {
	var b staff.Barber
	found := false
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		err := q.QueryRow(ctx,
			`SELECT `+barberColumns+`
			   FROM barber b `+barberPhotoJoin+`
			  WHERE b.barbershop_id = $1 AND b.staff_user_id = $2`,
			barbershopID, staffUserID,
		).Scan(&b.ID, &b.FullName, &b.CreatedAt, &b.UpdatedAt, &b.PhotoUpdatedAt)
		switch {
		case err == nil:
			found = true
		case errors.Is(err, pgx.ErrNoRows):
			// found queda false: el usuario no tiene barbero.
		default:
			return fmt.Errorf("select linked barber: %w", err)
		}
		return nil
	})
	if err != nil {
		return staff.Barber{}, false, fmt.Errorf("staff/postgres: get linked barber: %w", err)
	}
	return b, found, nil
}

// Link implementa staff.Repository.Link. El bloqueo consultivo por usuario
// serializa dos solicitudes simultáneas del MISMO usuario (así la unicidad
// parcial nunca llega a violarse por una carrera). Entre usuarios distintos
// decide el UPDATE condicional `staff_user_id IS NULL`: es atómico, así que de
// dos usuarios que piden el mismo barbero libre solo uno lo obtiene.
func (r *Repository) Link(ctx context.Context, barbershopID, staffUserID, barberID string) (staff.LinkResult, error) {
	var result staff.LinkResult
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		if _, err := q.Exec(ctx,
			`SELECT pg_advisory_xact_lock(hashtextextended('barber_link:' || $1::text, 0))`,
			staffUserID,
		); err != nil {
			return fmt.Errorf("lock link: %w", err)
		}

		var currentOwner *string
		err := q.QueryRow(ctx,
			`SELECT staff_user_id::text FROM barber WHERE id = $1 AND barbershop_id = $2`,
			barberID, barbershopID,
		).Scan(&currentOwner)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil // Found queda false: inexistente o de otra barbería.
		}
		if err != nil {
			return fmt.Errorf("select barber owner: %w", err)
		}
		result.Found = true

		switch {
		case currentOwner != nil && *currentOwner == staffUserID:
			// Ya es suyo: idempotente, nada que cambiar ni liberar.
		case currentOwner != nil:
			result.Taken = true
			return nil
		default:
			var released *string
			err := q.QueryRow(ctx,
				`UPDATE barber SET staff_user_id = NULL
				  WHERE barbershop_id = $1 AND staff_user_id = $2
				  RETURNING id::text`,
				barbershopID, staffUserID,
			).Scan(&released)
			if err != nil && !errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("release previous barber: %w", err)
			}
			if released != nil {
				result.ReleasedBarberID = *released
			}

			tag, err := q.Exec(ctx,
				`UPDATE barber SET staff_user_id = $3
				  WHERE id = $1 AND barbershop_id = $2 AND staff_user_id IS NULL`,
				barberID, barbershopID, staffUserID,
			)
			if err != nil {
				return fmt.Errorf("take barber: %w", err)
			}
			if tag.RowsAffected() == 0 {
				// Otro usuario lo tomó entre la lectura y el UPDATE: se deshace
				// la liberación para conservar el vínculo anterior del solicitante.
				result.Taken = true
				result.ReleasedBarberID = ""
				return errLinkTaken
			}
		}

		return selectBarber(ctx, q, barbershopID, barberID).
			Scan(&result.Barber.ID, &result.Barber.FullName, &result.Barber.CreatedAt,
				&result.Barber.UpdatedAt, &result.Barber.PhotoUpdatedAt)
	})
	if errors.Is(err, errLinkTaken) {
		return staff.LinkResult{Found: true, Taken: true}, nil
	}
	if err != nil {
		return staff.LinkResult{}, fmt.Errorf("staff/postgres: link barber: %w", err)
	}
	return result, nil
}

// errLinkTaken aborta (y por tanto revierte) la transacción de Link cuando
// otro usuario ganó la carrera por el barbero.
var errLinkTaken = errors.New("staff/postgres: barbero tomado por otro usuario")

// Unlink implementa staff.Repository.Unlink.
func (r *Repository) Unlink(ctx context.Context, barbershopID, staffUserID string) (string, error) {
	released := ""
	err := r.db.InTenantTx(ctx, database.BarbershopID(barbershopID), func(ctx context.Context, q database.Queries) error {
		var id *string
		err := q.QueryRow(ctx,
			`UPDATE barber SET staff_user_id = NULL
			  WHERE barbershop_id = $1 AND staff_user_id = $2
			  RETURNING id::text`,
			barbershopID, staffUserID,
		).Scan(&id)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("unlink barber: %w", err)
		}
		if id != nil {
			released = *id
		}
		return nil
	})
	if err != nil {
		return "", fmt.Errorf("staff/postgres: unlink barber: %w", err)
	}
	return released, nil
}
