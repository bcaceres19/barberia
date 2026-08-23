// Package notification administra los adaptadores de canal de entrega de
// HU-008 (DEC-066): Meta WhatsApp Cloud API y Resend/SES. No conoce
// staff_recovery_code, correo normalizado ni ningún otro tipo del módulo
// auth: expone dos puertos angostos (WhatsAppSender, EmailSender) y una
// composición (RecoveryCodeSender) que satisface el puerto pequeño que
// auth.RecoveryService consume, según
// docs/03-desarrollo/estandar-backend-go.md.
//
// La programación general de recordatorios, intentos y reintentos de B5
// (RN-REC-*) sigue sin implementar: este paquete solo cubre el envío
// síncrono, de un único intento, que HU-008 necesita.
package notification
