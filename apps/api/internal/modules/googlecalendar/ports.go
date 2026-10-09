package googlecalendar

import (
	"context"
	"time"
)

// Repository es el puerto de persistencia del módulo. El núcleo no importa
// pgx ni database: postgres/ traduce entre este contrato y la base. Toda
// operación recibe barbershopID y filtra por él además de RLS.
type Repository interface {
	// GetConnection devuelve la conexión del barbero. found=false si nunca
	// conectó. Incluye las credenciales cifradas: solo el servicio las lee.
	GetConnection(ctx context.Context, barbershopID, barberID string) (conn Connection, found bool, err error)

	// SaveConnected crea o reutiliza la fila del barbero y la deja `connected`
	// con las credenciales nuevas, limpiando el error y la fecha de desconexión.
	// En la MISMA transacción encola la publicación inicial: las citas y bloqueos
	// futuros de los próximos 6 meses y la reconciliación de los vínculos que ya
	// existían (DEC-101.8).
	SaveConnected(ctx context.Context, barbershopID, barberID string, saved ConnectedData) (Connection, error)

	// MarkStatus cambia el estado de la conexión. Para ReauthRequired y
	// Disconnected borra las credenciales (y, al desconectar, el correo de la
	// cuenta); para Error las conserva. found=false si no hay conexión.
	MarkStatus(ctx context.Context, barbershopID, barberID string, status Status, errorCode string, at time.Time) (found bool, err error)

	// SetReminder guarda reminder_minutes (nil = predeterminados) y, en la misma
	// transacción, encola la actualización de los eventos futuros ya publicados
	// (DEC-101.5). found=false si no hay conexión o está desconectada.
	SetReminder(ctx context.Context, barbershopID, barberID string, minutes *int) (Connection, bool, error)

	// CreateState guarda un estado OAuth de un solo uso y purga los vencidos
	// del mismo barbero.
	CreateState(ctx context.Context, state OAuthState, now time.Time) error

	// ConsumeState marca como usado, en una sola operación atómica, el estado
	// vigente con ese hash. found=false si no existe, venció o ya se usó.
	ConsumeState(ctx context.Context, barbershopID, stateHash string, now time.Time) (OAuthState, bool, error)

	// BarberOfUser devuelve el barbero vinculado al usuario (DEC-100).
	BarberOfUser(ctx context.Context, barbershopID, staffUserID string) (barberID string, found bool, err error)

	// JobCounts cuenta los trabajos de la conexión del barbero: los que esperan o
	// se ejecutan y los que agotaron sus intentos.
	JobCounts(ctx context.Context, barbershopID, barberID string) (JobCounts, error)

	// RequeueConnection hace vencer ya los trabajos pendientes de la conexión y
	// reencola los fallidos («Sincronizar ahora»). Solo reordena la cola de ESA
	// conexión y es segura ante varios clics. found=false sin conexión que publique.
	RequeueConnection(ctx context.Context, barbershopID, barberID string, now time.Time) (JobCounts, bool, error)
}

// JobCounts resume la cola de una conexión.
type JobCounts struct {
	Pending int
	Failed  int
}

// ConnectedData son los datos con que una conexión queda `connected`.
type ConnectedData struct {
	AccountEmail string
	Credentials  []byte
	KeyID        string
	At           time.Time
}

// OAuthProvider es el puerto hacia Google. El SDK vive solo en su adaptador
// (google/); el núcleo y las pruebas usan este contrato.
type OAuthProvider interface {
	// AuthorizationURL construye la URL de consentimiento con PKCE (S256),
	// acceso offline y consentimiento forzado.
	AuthorizationURL(state, codeVerifier string) string

	// Exchange canjea el código de autorización por el refresh token y la
	// cuenta conectada.
	Exchange(ctx context.Context, code, codeVerifier string) (Tokens, error)

	// RefreshAccessToken obtiene un access token nuevo. Devuelve ErrTokenRevoked
	// cuando Google ya no acepta el refresh token (revocado, caducado o permiso
	// retirado) y cualquier otro error para fallos transitorios.
	RefreshAccessToken(ctx context.Context, refreshToken string) (accessToken string, expiresAt time.Time, err error)

	// Revoke revoca el refresh token en Google. Un token ya inválido no es un error.
	Revoke(ctx context.Context, refreshToken string) error
}

// Randomness aísla la fuente de aleatoriedad y el reloj para pruebas
// deterministas.
type Randomness interface {
	// Token devuelve n bytes aleatorios codificados en base64 URL sin relleno.
	Token(n int) (string, error)
}
