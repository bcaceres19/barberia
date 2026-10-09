// Package googlecalendar es el módulo dueño de la integración de un barbero
// con SU Google Calendar (DEC-099): conexión OAuth con PKCE, tokens cifrados,
// estado de la conexión, preferencia de recordatorio y desconexión.
//
// La integración es unidireccional, de NAVA hacia Google: NAVA es la única
// autoridad sobre las citas y Google Calendar es una vista. Este paquete no
// publica eventos (issue #324); solo gestiona la conexión que esa publicación
// usará. El núcleo no importa Chi, PostgreSQL ni el SDK de Google: el
// adaptador de Google vive en google/, la persistencia en postgres/ y HTTP en
// httpapi/, y la raíz de composición (cmd/api) los une.
package googlecalendar
