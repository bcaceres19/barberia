package schedule

import (
	"sort"
	"time"
)

// ColombianHoliday es un festivo colombiano de un año calendario
// (HU-041, RN-BLQ-02): fecha civil y nombre. Dato de referencia
// calendárica, igual para toda barbería y todo barbero.
type ColombianHoliday struct {
	Date time.Time // Medianoche UTC del día civil; solo la fecha importa.
	Name string
}

// ColombianHolidaysForYear calcula, de forma determinista y sin tabla ni
// dependencia externa, los dieciocho festivos colombianos de year, según
// la Ley 51 de 1983 ("Ley Emiliani"): seis fijos, dos ligados a Pascua sin
// trasladarse (Jueves y Viernes Santo) y diez trasladados al lunes
// siguiente cuando no caen ya en lunes (siete fijos y tres ligados a
// Pascua). El cálculo de la fecha de Pascua usa el algoritmo gregoriano
// anónimo (Meeus/Jones/Butcher), válido para cualquier año del calendario
// gregoriano.
//
// Se documenta aquí, en código, la fuente reproducible que el prompt de
// HU-041 exige registrar: no depende de un proveedor externo ni de una
// lista mantenida a mano que pueda desincronizarse año a año.
func ColombianHolidaysForYear(year int) []ColombianHoliday {
	easter := easterSunday(year)

	holidays := []ColombianHoliday{
		{civilDate(year, time.January, 1), "Año Nuevo"},
		{nextMondayOnOrAfter(civilDate(year, time.January, 6)), "Día de los Reyes Magos"},
		{nextMondayOnOrAfter(civilDate(year, time.March, 19)), "Día de San José"},
		{easter.AddDate(0, 0, -3), "Jueves Santo"},
		{easter.AddDate(0, 0, -2), "Viernes Santo"},
		{civilDate(year, time.May, 1), "Día del Trabajo"},
		{nextMondayOnOrAfter(easter.AddDate(0, 0, 39)), "Ascensión del Señor"},
		{nextMondayOnOrAfter(easter.AddDate(0, 0, 60)), "Corpus Christi"},
		{nextMondayOnOrAfter(easter.AddDate(0, 0, 68)), "Sagrado Corazón de Jesús"},
		{nextMondayOnOrAfter(civilDate(year, time.June, 29)), "San Pedro y San Pablo"},
		{civilDate(year, time.July, 20), "Día de la Independencia"},
		{civilDate(year, time.August, 7), "Batalla de Boyacá"},
		{nextMondayOnOrAfter(civilDate(year, time.August, 15)), "Asunción de la Virgen"},
		{nextMondayOnOrAfter(civilDate(year, time.October, 12)), "Día de la Raza"},
		{nextMondayOnOrAfter(civilDate(year, time.November, 1)), "Día de Todos los Santos"},
		{nextMondayOnOrAfter(civilDate(year, time.November, 11)), "Independencia de Cartagena"},
		{civilDate(year, time.December, 8), "Día de la Inmaculada Concepción"},
		{civilDate(year, time.December, 25), "Navidad"},
	}

	sort.Slice(holidays, func(i, j int) bool { return holidays[i].Date.Before(holidays[j].Date) })
	return holidays
}

// IsColombianHoliday informa si date (comparada solo por año/mes/día
// civil, ignorando cualquier componente de hora) es uno de los festivos
// colombianos de su propio año.
func IsColombianHoliday(date time.Time) (ColombianHoliday, bool) {
	target := civilDate(date.Year(), date.Month(), date.Day())
	for _, h := range ColombianHolidaysForYear(date.Year()) {
		if h.Date.Equal(target) {
			return h, true
		}
	}
	return ColombianHoliday{}, false
}

func civilDate(year int, month time.Month, day int) time.Time {
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

// nextMondayOnOrAfter aplica el traslado de la Ley Emiliani: si d ya cae
// en lunes, se conserva; en cualquier otro caso, avanza hasta el lunes
// siguiente.
func nextMondayOnOrAfter(d time.Time) time.Time {
	for d.Weekday() != time.Monday {
		d = d.AddDate(0, 0, 1)
	}
	return d
}

// easterSunday calcula la fecha del Domingo de Pascua para year mediante
// el algoritmo gregoriano anónimo (Meeus/Jones/Butcher), válido para
// cualquier año del calendario gregoriano.
func easterSunday(year int) time.Time {
	a := year % 19
	b := year / 100
	c := year % 100
	d := b / 4
	e := b % 4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i := c / 4
	k := c % 4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := ((h + l - 7*m + 114) % 31) + 1
	return civilDate(year, time.Month(month), day)
}
