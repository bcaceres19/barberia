#!/usr/bin/env bash
# database/tests/notification_lease_concurrency_two_connections.sh
#
# Prueba de concurrencia real (DDL-CON-02): dos conexiones psql GENUINAS y
# solapadas reclaman a la vez la misma fila 'pending'. notification_lease_
# concurrency.sql ya prueba la forma del protocolo (claim/CAS/recuperación)
# con llamadas secuenciales en una sola sesión, controlando `p_now`; eso basta
# para la lógica, pero no demuestra que SKIP LOCKED salte de verdad una fila
# bloqueada por la transacción ABIERTA Y NO CONFIRMADA de otro proceso. Esta
# prueba sí lo exige: la sesión A mantiene su transacción de reclamo abierta
# mientras la sesión B, un proceso `psql` real y aparte, intenta reclamar la
# misma fila.
#
# Uso:
#   DATABASE_TEST_URL=postgres://user:pass@host:5432/db \
#     ./notification_lease_concurrency_two_connections.sh
#
# Requiere, ya aplicados sobre esa base: las cuatro migraciones de
# database/migrations, database/modelo-fisico-referencia.sql,
# testdata/dos_barberias.sql y testdata/notification_lease_fixture.sql.
#
# Coordinación sin sleeps de negocio (estrategia-pruebas.md §7 · "no usar
# sleep para coordinar concurrencia; usar barreras, canales o condiciones
# observables"): el orden A-antes-que-B lo fija un archivo marcador que A
# escribe justo después de tomar un advisory lock de sesión (una barrera de
# PostgreSQL real, bloqueante); el propio B se queda bloqueado en
# `pg_advisory_lock` -esperando a que A la suelte tras reclamar-, no en un
# `sleep` que adivine cuánto tarda A. El único `sleep` del script es el
# intervalo de sondeo, acotado con un máximo de intentos, mientras se observa
# la aparición del archivo marcador.

set -euo pipefail

: "${DATABASE_TEST_URL:?Defina DATABASE_TEST_URL antes de ejecutar esta prueba.}"

BARBERSHOP_ID='11111111-1111-1111-1111-111111111111'
APPOINTMENT_ID='a99a99a9-a99a-a99a-a99a-a99a99a99a99'
SCHEDULE_ID='00000000-0000-0000-0000-0000000000aa'

WORKDIR=$(mktemp -d)
cleanup() {
  psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -q \
    -c "DELETE FROM notification_schedule WHERE id = '$SCHEDULE_ID';" >/dev/null 2>&1 || true
  rm -rf "$WORKDIR"
}
trap cleanup EXIT

echo "=== notification_claim_due · concurrencia real con dos conexiones ==="

psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -q <<SQL
SET ROLE barberia_migrator;
DELETE FROM notification_schedule WHERE id = '$SCHEDULE_ID';
INSERT INTO notification_schedule (id, barbershop_id, appointment_id, event_type, channel, scheduled_for)
VALUES ('$SCHEDULE_ID', '$BARBERSHOP_ID', '$APPOINTMENT_ID', 'reminder', 'email', '2026-01-05 09:30:00-05');
SQL

# Sesión A: toma el advisory lock 424201 ANTES de reclamar (marca que ya
# reclamó bajo FOR UPDATE), avisa por archivo, luego espera el advisory lock
# 424202 -que B suelta al terminar su intento- antes de confirmar.
cat > "$WORKDIR/a.sql" <<SQL
\\set ON_ERROR_STOP on
BEGIN;
SET ROLE barberia_worker;
SELECT pg_advisory_lock(424201);
\\! touch "$WORKDIR/a_ready"
SELECT schedule_id, claim_token AS a_token
FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz)
WHERE schedule_id = '$SCHEDULE_ID' \\gset
SELECT pg_advisory_unlock(424201);
SELECT pg_advisory_lock(424202);
SELECT pg_advisory_unlock(424202);
COMMIT;
\\o $WORKDIR/a_token.txt
\\qecho :a_token
\\o
SQL

# Sesión B: toma el advisory lock 424202 de inmediato (lo suelta solo al
# terminar), y se bloquea en el 424201 hasta que A confirme -con el archivo
# marcador- que ya reclamó. La transacción de A sigue abierta en ese
# instante: si SKIP LOCKED funciona, este intento de B no debe traer la fila.
cat > "$WORKDIR/b.sql" <<SQL
\\set ON_ERROR_STOP on
SET ROLE barberia_worker;
SELECT pg_advisory_lock(424202);
SELECT pg_advisory_lock(424201);
SELECT pg_advisory_unlock(424201);
SELECT count(*) AS b_got_target
FROM notification_claim_due(10, 60, '2026-01-05 09:30:00-05'::timestamptz)
WHERE schedule_id = '$SCHEDULE_ID' \\gset
SELECT pg_advisory_unlock(424202);
\\o $WORKDIR/b_count.txt
\\qecho :b_got_target
\\o
SQL

psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -q -f "$WORKDIR/a.sql" > "$WORKDIR/a.log" 2>&1 &
A_PID=$!

# Condición observable, acotada: espera a que A haya tomado el lock 424201
# (justo antes de reclamar), nunca una duración adivinada.
tries=0
until [ -f "$WORKDIR/a_ready" ]; do
  tries=$((tries + 1))
  if [ "$tries" -ge 100 ]; then
    echo "La sesión A no tomó el lock a tiempo (10 s)." >&2
    kill "$A_PID" 2>/dev/null || true
    exit 1
  fi
  sleep 0.1
done

psql "$DATABASE_TEST_URL" -v ON_ERROR_STOP=1 -q -f "$WORKDIR/b.sql" > "$WORKDIR/b.log" 2>&1 &
B_PID=$!

wait "$A_PID"
wait "$B_PID"

A_TOKEN=$(tr -d '[:space:]' < "$WORKDIR/a_token.txt")
B_COUNT=$(tr -d '[:space:]' < "$WORKDIR/b_count.txt")

if [ -z "$A_TOKEN" ]; then
  echo "FALLÓ: la sesión A no obtuvo claim_token. Log:" >&2
  cat "$WORKDIR/a.log" >&2
  exit 1
fi

if [ "$B_COUNT" != "0" ]; then
  echo "FALLÓ: la sesión B reclamó la misma fila que A (SKIP LOCKED no la saltó). Logs:" >&2
  cat "$WORKDIR/a.log" "$WORKDIR/b.log" >&2
  exit 1
fi

echo "OK · A reclamó la fila (claim_token $A_TOKEN) y B, con su transacción solapada, obtuvo 0 filas."
echo "=== notification_claim_due · concurrencia real: PASÓ ==="
