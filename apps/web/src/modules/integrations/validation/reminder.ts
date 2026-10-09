// Validación del recordatorio (DEC-101): entero de 0 a 40320 minutos (cuatro
// semanas, el máximo de Google). El servidor repite la regla (422); aquí solo se
// evita enviar un valor que ya se sabe inválido.
export const MIN_REMINDER_MINUTES = 0
export const MAX_REMINDER_MINUTES = 40320

export type ReminderValidation = { ok: true; minutes: number } | { ok: false; message: string }

export function validateReminderMinutes(raw: string): ReminderValidation {
  const text = raw.trim()
  if (text === '') {
    return { ok: false, message: 'Escribe cuántos minutos antes quieres el aviso.' }
  }
  if (!/^\d+$/.test(text)) {
    return { ok: false, message: 'Usa solo números enteros, por ejemplo 30.' }
  }
  const minutes = Number(text)
  if (minutes < MIN_REMINDER_MINUTES || minutes > MAX_REMINDER_MINUTES) {
    return {
      ok: false,
      message: `El aviso debe ser de ${MIN_REMINDER_MINUTES} a ${MAX_REMINDER_MINUTES} minutos.`,
    }
  }
  return { ok: true, minutes }
}
