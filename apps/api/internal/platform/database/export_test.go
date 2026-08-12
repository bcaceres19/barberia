package database

import "github.com/jackc/pgx/v5/pgxpool"

// NewForTest construye un *DB directamente sobre un pool ya creado, sin
// pasar por NewDB ni sus hooks AfterRelease/AfterConnect. Existe solo para
// pruebas que necesitan control directo del pool (p. ej. conectarse con un
// rol de prueba específico); el código de producción SIEMPRE debe usar
// NewDB, la única vía que instala los hooks de residuo de contexto y
// statement_timeout.
func NewForTest(pool *pgxpool.Pool) *DB {
	return &DB{pool: pool}
}
