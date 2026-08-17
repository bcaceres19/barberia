// Contexto mínimo de sesión que HU-012 rehidrata desde
// GET /private/auth/session (DEC-060). Espejo tipado del cuerpo real del
// contrato: nunca StaffUserID, correo ni nombre del barbero (RN-DAT-02).
export interface SessionContext {
  barbershopId: string
  barbershopName: string
  expiresAt: string
}
