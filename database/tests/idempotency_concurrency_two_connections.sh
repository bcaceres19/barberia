#!/usr/bin/env bash
# database/tests/idempotency_concurrency_two_connections.sh
#
# Prueba de concurrencia real (CA-004-03, DEC-043): dos conexiones psql
# GENUINAS y solapadas intentan idempotency_begin con la MISMA clave.
# idempotency_concurrency.sql ya prueba la forma del protocolo (proceed,
# replay, conflict_*) con llamadas secuenciales en una sola sesión; eso
# basta para la lógica, pero no demuestra que pg_try_advisory_xact_lock
# rechace de verdad a un segundo llamador mientras el primero sigue con su
# transacción abierta (un mismo backend nunca bloquea su propio lock
# consultivo). Esta prueba sí lo exige: la sesión A mantiene su transacción
# abierta tras tomar el lock, mientras la sesión B -un proceso `psql` real y
# aparte- intenta la misma clave.
#
# Uso:
#   DATABASE_TEST_URL=postgres://user:pass@host:5432/db \
#     ./idempotency_concurrency_two_connections.sh
#
# Requiere, ya aplicadas sobre esa base: las cinco migraciones de
# database/migrations y testdata/dos_barberias.sql.
#
# Coordinación sin sleeps de negocio (estrategia-pruebas.md §7): el orden
# A-antes-que-B lo fija un archivo marcador (a_ready) que A escribe justo
# después de llamar idempotency_begin, con su transacción todavía abierta;
# el script principal sondea ese archivo con un máximo de intentos, nunca
# una duración adivinada. A, a su vez, espera un segundo archivo marcador
# (b_done) -que B escribe al terminar- antes de hacer ROLLBACK, así que su
# transacción permanece abierta mientras B actúa. El único `sleep` real es
# el intervalo de sondeo de esos archivos, acotado y observable, no una
# apuesta sobre cuánto tarda el otro proceso.

set -euo pipefail

: "${DATABASE_TEST_URL:?Defina DATABASE_TEST_URL antes de ejecutar esta prueba.}"

BARBERSHOP_ID='11111111-1111-1111-1111-111111111111'
IDEMPOTENCY_KEY='audit-test-two-connections'
FINGERPRINT_A=$(printf 'a%.0s' {1..64})
FINGERPRINT_B=$(printf 'b%.0s' {1..64})

WORKDIR=$(mktemp -d)
cleanup() {
  rm -rf "$WORKDIR"
}
trap cleanup EXIT

# No hace falta borrar nada antes ni después: tanto la sesión A como la B
# terminan con ROLLBACK (abajo), así que ninguna fila queda commiteada en
# una ejecución normal. $DATABASE_TEST_URL se conecta como barberia_app -el
# rol real de la aplicación, sin GRANT sobre barberia_owner- a propósito:
# esta prueba verifica el comportamiento real del rol de producción, no el
# de un rol administrativo.

echo "=== idempotency_begin · concurrencia real con dos conexiones (CA-004-03) ==="

# Sesión A: llama idempotency_begin, avisa por archivo que ya lo hizo, y
# espera (bloqueada dentro de la MISMA transacción, sin ROLLBACK todavía) a
# que B termine su propio intento.
cat > "$WORKDIR/a.sql" <<SQL
\\set ON_ERROR_STOP on
SET ROLE barberia_app;
BEGIN;
SELECT outcome AS a_outcome
FROM idempotency_begin(
  '$BARBERSHOP_ID'::uuid, '$IDEMPOTENCY_KEY', 'audit_test_operation', '$FINGERPRINT_A', 60
) \\gset
\\! touch "$WORKDIR/a_ready"
\\! i=0; while [ ! -f "$WORKDIR/b_done" ] && [ "\$i" -lt 100 ]; do sleep 0.1; i=\$((i+1)); done; true
ROLLBACK;
\\o $WORKDIR/a_outcome.txt
\\qecho :a_outcome
\\o
SQL

psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -q -f "$WORKDIR/a.sql" > "$WORKDIR/a.log" 2>&1 &
A_PID=$!

tries=0
until [ -f "$WORKDIR/a_ready" ]; do
  tries=$((tries + 1))
  if [ "$tries" -ge 100 ]; then
    echo "La sesión A no llamó a idempotency_begin a tiempo (10 s)." >&2
    kill "$A_PID" 2>/dev/null || true
    exit 1
  fi
  sleep 0.1
done

# Sesión B: llama idempotency_begin con la MISMA clave mientras la
# transacción de A sigue abierta (bloqueada esperando b_done). Si
# pg_try_advisory_xact_lock funciona, B debe recibir 'locked' de inmediato,
# sin quedar bloqueado dentro de la función.
psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -q -c "
SET ROLE barberia_app;
SELECT outcome
FROM idempotency_begin(
  '$BARBERSHOP_ID'::uuid, '$IDEMPOTENCY_KEY', 'audit_test_operation', '$FINGERPRINT_B', 60
);
" > "$WORKDIR/b.log" 2>&1

touch "$WORKDIR/b_done"
wait "$A_PID"

A_OUTCOME=$(tr -d '[:space:]' < "$WORKDIR/a_outcome.txt")
B_OUTCOME=$(grep -A2 'outcome' "$WORKDIR/b.log" | tail -1 | tr -d '[:space:]')

if [ "$A_OUTCOME" != "proceed" ]; then
  echo "FALLÓ: la sesión A esperaba 'proceed', obtuvo '$A_OUTCOME'. Log:" >&2
  cat "$WORKDIR/a.log" >&2
  exit 1
fi

if [ "$B_OUTCOME" != "locked" ]; then
  echo "FALLÓ: la sesión B esperaba 'locked' (transacción de A todavía abierta), obtuvo '$B_OUTCOME'. Logs:" >&2
  cat "$WORKDIR/a.log" "$WORKDIR/b.log" >&2
  exit 1
fi

echo "OK · A obtuvo 'proceed' y B, con su transacción solapada, obtuvo 'locked' de inmediato (sin espera bloqueante)."
echo "=== idempotency_begin · concurrencia real: PASÓ ==="
