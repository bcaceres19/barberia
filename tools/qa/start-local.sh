#!/usr/bin/env bash
# Arranque exclusivo de QA de #300. No modifica .env.local ni lanza proveedores.
set -euo pipefail
repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
cd "$repo_root"
if [[ ! -f apps/api/.env.local ]]; then
  echo 'Falta apps/api/.env.local; aplicar local-app-startup.' >&2
  exit 1
fi
set -a
source apps/api/.env.local
set +a
if [[ "${APP_ENVIRONMENT:-}" != local && "${APP_ENVIRONMENT:-}" != test ]]; then
  echo 'QA requiere APP_ENVIRONMENT=local/test.' >&2
  exit 1
fi
export NAVA_QA_DATABASE_URL="${DATABASE_URL:?Falta conexión del migrador}"
node tools/qa/provision-ui.mjs "$@"
qa_private="$repo_root/apps/web/.auth/nava-qa"
mkdir -p "$qa_private"
chmod 700 "$qa_private"
# La captura envuelve al remitente: quitar credenciales impide envíos reales.
export OTP_PROVIDER=meta
unset APP_META_WHATSAPP_PHONE_NUMBER_ID APP_META_WHATSAPP_ACCESS_TOKEN APP_META_WHATSAPP_TEMPLATE_NAME
unset APP_RESEND_API_KEY APP_RESEND_FROM_ADDRESS TWILIO_AUTH_TOKEN TWILIO_API_KEY_SECRET
export APP_PHONE_CHALLENGE_CAPTURE_FILE="$qa_private/challenge.json"
export APP_RECOVERY_CAPTURE_FILE="$qa_private/recovery.json"
export APP_TRUSTED_PROXIES=127.0.0.1/32,::1/128
cd apps/api
exec go run ./cmd/api
