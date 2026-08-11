// Package clock define el puerto de tiempo que consumen los módulos de
// dominio. Ningún servicio llama time.Now directamente: recibe un [Clock] por
// constructor para que las pruebas puedan congelar o avanzar el reloj sin
// depender de la hora real del equipo.
package clock

import "time"

// Clock expone el instante actual. Los módulos declaran su propia
// dependencia de esta interfaz junto al código que la consume, según
// docs/03-desarrollo/estandar-backend-go.md.
type Clock interface {
	// Now devuelve el instante actual en UTC.
	Now() time.Time
}

// System es la implementación de [Clock] respaldada por el reloj del
// sistema operativo. Es la única implementación que se inyecta en cmd/api y
// cmd/worker; las pruebas usan un doble propio.
type System struct{}

// Now devuelve time.Now().UTC().
func (System) Now() time.Time {
	return time.Now().UTC()
}
