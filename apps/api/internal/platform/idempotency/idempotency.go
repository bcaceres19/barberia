package idempotency

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"system-barbershop/internal/platform/apperr"
	"system-barbershop/internal/platform/database"
)

// Key es una clave de idempotencia ya validada, tal como la envía el
// cliente en la cabecera Idempotency-Key. No es un string desnudo: separa
// un valor ya aceptado de una entrada todavía sin comprobar (ver ParseKey).
type Key string

// keyPattern acota lo que se acepta de un cliente al mismo criterio que
// httpserver.RequestID usa para X-Request-Id (caracteres seguros para log y
// cabecera), extendido a 255 para coincidir exactamente con
// idempotency_record_idempotency_key_ck. Cualquier otra cosa se rechaza
// ANTES de tocar PostgreSQL o cualquier registro.
var keyPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,255}$`)

// ParseKey valida el valor crudo de la cabecera Idempotency-Key. No admite
// vacío: a diferencia de X-Request-Id, esta cabecera es obligatoria para
// toda operación crítica (docs/06-api/estandar-openapi.md sección 10) y no
// existe un valor generado de reemplazo razonable. El mensaje de error
// nunca incluye raw: un formato inválido no es motivo para reflejar de
// vuelta un valor que el cliente pudo escribir con datos sensibles.
func ParseKey(raw string) (Key, error) {
	if raw == "" {
		return "", apperr.Invalid("falta la cabecera Idempotency-Key, obligatoria para esta operación")
	}
	if !keyPattern.MatchString(raw) {
		return "", apperr.Invalid("Idempotency-Key tiene un formato inválido")
	}
	return Key(raw), nil
}

// Redacted devuelve una referencia opaca y estable de k, segura para
// registrar en logs o trazas: la clave literal del cliente nunca se
// registra (podría ser o contener datos sensibles elegidos por el
// cliente). Dos llamadas con la misma Key devuelven la misma referencia,
// lo que basta para correlacionar sin exponer el valor original.
func (k Key) Redacted() string {
	sum := sha256.Sum256([]byte(k))
	return hex.EncodeToString(sum[:6])
}

// Operation identifica, dentro de una barbería, QUÉ operación reclamó una
// clave. No es una entrada de cliente: cada módulo llamador pasa una
// constante propia (por ejemplo "create_appointment"). RN-IDE-01 exige que
// una clave ya usada para una operación no se reutilice para otra
// (idempotency_begin responde 'conflict_operation' en ese caso).
type Operation string

// fingerprintPattern refleja exactamente
// idempotency_record_request_fingerprint_ck (DDL-VAL-01): hexadecimal en
// minúsculas, 32 a 128 caracteres. ComputeFingerprint siempre produce un
// valor que la cumple; esta comprobación es una defensa contra un
// Fingerprint construido a mano fuera de ComputeFingerprint, no una
// duplicación de la restricción de PostgreSQL (que sigue siendo la
// garantía real).
var fingerprintPattern = regexp.MustCompile(`^[0-9a-f]{32,128}$`)

// Fingerprint es la huella canónica del contenido de una solicitud.
type Fingerprint string

func (f Fingerprint) valid() bool {
	return fingerprintPattern.MatchString(string(f))
}

// ComputeFingerprint calcula la huella canónica SHA-256, codificada en
// hexadecimal minúsculo, de method, path y body. La forma canónica es
// exactamente:
//
//	MAYÚSCULAS(method) + "\n" + path + "\n" + body
//
// method se normaliza a mayúsculas (GET y get no deben producir huellas
// distintas). path es la ruta sin query string tal como Chi la resolvió
// (r.URL.Path): dos rutas equivalentes por normalización de query no deben
// colisionar ni divergir por el orden de sus parámetros, así que la query
// queda fuera a propósito — si una operación crítica necesita distinguir
// por query string, debe incluirla explícitamente en el body canónico que
// pase aquí, no asumir que este paquete lo hace. body es el CUERPO CRUDO
// recibido, byte a byte: nunca una re-serialización de un JSON ya
// decodificado, porque reordenar claves o normalizar espacios cambiaría la
// huella de un contenido semánticamente idéntico y volvería la comparación
// inservible. El separador "\n" evita que method+path concatenados sin
// límite produzcan la misma huella para pares distintos (ni el método HTTP
// ni una ruta de Chi contienen salto de línea).
func ComputeFingerprint(method, path string, body []byte) Fingerprint {
	h := sha256.New()
	h.Write([]byte(strings.ToUpper(method)))
	h.Write([]byte{'\n'})
	h.Write([]byte(path))
	h.Write([]byte{'\n'})
	h.Write(body)
	return Fingerprint(hex.EncodeToString(h.Sum(nil)))
}

// Outcome es el pequeño catálogo cerrado de resultados que Begin puede
// devolver, ya traducido del texto que entrega idempotency_begin (DEC-043)
// a un tipo Go que el compilador puede exhaustivamente comprobar en un
// switch.
type Outcome int

const (
	// OutcomeProceed: la clave se reclamó ahora mismo. El llamador DEBE
	// ejecutar el efecto exactamente una vez y confirmarlo con Complete.
	OutcomeProceed Outcome = iota
	// OutcomeReplay: ya se ejecutó con la misma clave y el mismo
	// contenido. Decision.Response trae la respuesta original exacta; el
	// llamador NO debe repetir el efecto (CA-004-01).
	OutcomeReplay
	// OutcomeConflictOperation: la clave ya se usó para OTRA operación
	// (RN-IDE-01). No se ejecuta ningún efecto.
	OutcomeConflictOperation
	// OutcomeConflictFingerprint: la clave ya se usó con contenido
	// distinto (CA-004-02). No se ejecuta ningún efecto.
	OutcomeConflictFingerprint
	// OutcomeConflictInProgress: el lock se obtuvo, pero la fila vigente
	// sigue 'in_progress' de la misma operación (poco frecuente bajo el
	// lock, posible si el TTL de una reclamación previa no venció).
	OutcomeConflictInProgress
	// OutcomeLocked: otra transacción tiene el bloqueo consultivo de esta
	// clave en este momento (DEC-043). Respuesta inmediata, sin espera
	// acotada; el llamador puede reintentar más tarde.
	OutcomeLocked
)

// String describe el Outcome de forma estable para logs y mensajes de
// error; no es el texto que viaja al cliente (eso lo decide Decision.AsError
// vía apperr, con mensajes propios en español seguro para el cliente).
func (o Outcome) String() string {
	switch o {
	case OutcomeProceed:
		return "proceed"
	case OutcomeReplay:
		return "replay"
	case OutcomeConflictOperation:
		return "conflict_operation"
	case OutcomeConflictFingerprint:
		return "conflict_fingerprint"
	case OutcomeConflictInProgress:
		return "conflict_in_progress"
	case OutcomeLocked:
		return "locked"
	default:
		return fmt.Sprintf("outcome_desconocido(%d)", int(o))
	}
}

// StoredResponse es la respuesta HTTP original que idempotency_complete
// persistió y que una repetición exacta debe reproducir byte a byte
// (idempotency_record.response_body: "no debe reordenarse ni normalizarse
// al almacenarla").
type StoredResponse struct {
	Status      int
	ContentType string
	Body        string
}

// Decision es el resultado completo de una llamada a Begin.
type Decision struct {
	Outcome Outcome
	// Response solo es significativo cuando Outcome == OutcomeReplay.
	Response StoredResponse
}

// AsError traduce una Decision que NO debe ejecutar el efecto (cualquier
// Outcome salvo Proceed y Replay) al vocabulario apperr que
// httpserver.Translate ya sabe convertir en application/problem+json.
// Devuelve nil para Proceed y Replay: en ambos casos el llamador tiene
// trabajo útil que hacer (ejecutar el efecto, o reproducir Response), no un
// error que traducir.
func (d Decision) AsError() error {
	switch d.Outcome {
	case OutcomeProceed, OutcomeReplay:
		return nil
	case OutcomeConflictOperation:
		return apperr.IdempotencyConflict(
			"la clave de idempotencia ya se usó para otra operación; usa una clave nueva")
	case OutcomeConflictFingerprint:
		return apperr.IdempotencyConflict(
			"la clave de idempotencia ya se usó con un contenido distinto; usa una clave nueva")
	case OutcomeConflictInProgress, OutcomeLocked:
		return apperr.IdempotencyLocked(
			"la operación con esta clave de idempotencia todavía está en curso; reintenta en unos segundos")
	default:
		return apperr.Internal(fmt.Errorf("idempotency: outcome desconocido %q", d.Outcome))
	}
}

// Coordinator es el puerto de aplicación del protocolo de idempotencia
// (DEC-043), implementado por SQLCoordinator contra las funciones
// SECURITY DEFINER ya aprobadas y migradas. Los tres métodos DEBEN
// ejecutarse dentro de la MISMA transacción de negocio, es decir, con el
// mismo database.Queries que una única llamada a database.DB.InTenantTx
// entrega a su callback:
//
//   - El bloqueo que toma Begin (pg_try_advisory_xact_lock) tiene alcance
//     TRANSACCIONAL: se libera solo al terminar esa transacción (COMMIT o
//     ROLLBACK), nunca antes. Si Begin corriera en una transacción propia y
//     ya cerrada, el lock se liberaría de inmediato y dejaría de proteger al
//     efecto que se ejecuta después, anulando la garantía de DEC-043.
//   - Si el callback de InTenantTx devuelve error (el efecto falló, hizo
//     panic, o el contexto se canceló), InTenantTx hace ROLLBACK de TODA la
//     transacción, lo que automáticamente deshace el INSERT que Begin hizo
//     para reclamar la clave. Un reintento legítimo con la misma clave
//     encuentra la clave libre de nuevo, sin necesidad de llamar Abort
//     (CA-004-06). Abort existe para el caso distinto en que el llamador
//     necesita conservar el resto de esa transacción (por ejemplo, para
//     dejar evidencia de auditoría del fallo) y solo limpiar la
//     reclamación de idempotencia antes de hacer COMMIT; decidir si ese
//     caso aplica es responsabilidad de cada módulo llamador, no de este
//     paquete.
type Coordinator interface {
	// Begin reclama la clave para (shop, key) o informa por qué no puede
	// reclamarse ahora. Nunca ejecuta el efecto: solo el llamador, y solo
	// cuando Decision.Outcome == OutcomeProceed, debe hacerlo. ttl es la
	// vigencia de la reclamación si el resultado es OutcomeProceed;
	// vencida, una clave se trata como si no existiera (CA-004-05).
	Begin(
		ctx context.Context,
		q database.Queries,
		shop database.BarbershopID,
		key Key,
		operation Operation,
		fingerprint Fingerprint,
		ttl time.Duration,
	) (Decision, error)

	// Complete persiste la respuesta de un efecto que tuvo éxito. Solo
	// transiciona una fila 'in_progress' a 'completed'; ok=false indica
	// que la fila ya no estaba en ese estado (otra llamada la completó, o
	// nunca se reclamó con Begin en esta transacción) y el llamador debe
	// decidir con esa información, no asumir éxito silencioso.
	Complete(
		ctx context.Context,
		q database.Queries,
		shop database.BarbershopID,
		key Key,
		response StoredResponse,
	) (ok bool, err error)

	// Abort libera una reclamación 'in_progress' cuyo efecto no debe
	// reproducirse. ok=false indica que no había nada que liberar (la fila
	// ya no estaba 'in_progress'). Abort JAMÁS borra una fila 'completed':
	// esa garantía la impone la propia función SQL, no este adaptador
	// (cierra DDL-IDEM-01).
	Abort(
		ctx context.Context,
		q database.Queries,
		shop database.BarbershopID,
		key Key,
	) (ok bool, err error)
}
