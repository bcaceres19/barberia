package availability

import (
	"sort"
	"time"
)

// Interval es un intervalo semiabierto [Start, End) de instantes absolutos
// (RN-DIS-05). Se usa tanto para un tramo laboral como para una ocupación
// (bloqueo o cita que ocupa agenda): el dominio no distingue su origen.
type Interval struct {
	Start time.Time
	End   time.Time
}

// valid informa si el intervalo no está vacío ni invertido.
func (iv Interval) valid() bool {
	return iv.End.After(iv.Start)
}

// Policy son los parámetros ya resueltos que GenerateStarts necesita:
// ninguno se recalcula aquí (RN-DIS-04, RN-DIS-06, RN-SER-01/02).
type Policy struct {
	// ServiceDuration es la duración candidata del servicio (snapshot,
	// DEC-002/DEC-004): el intervalo [inicio, inicio+ServiceDuration) debe
	// caber entero (RN-DIS-01).
	ServiceDuration time.Duration
	// GridStep es el paso de la rejilla (RN-DIS-06, DEC-006/DEC-018/
	// DEC-083). Debe ser estrictamente positivo.
	GridStep time.Duration
	// EarliestStart es el instante más temprano aceptable (ahora +
	// anticipación mínima, RN-DIS-04). Un candidato anterior se descarta.
	EarliestStart time.Time
	// LatestStart es el instante más tardío aceptable (ahora + ventana
	// máxima, RN-DIS-04). Un candidato posterior se descarta.
	LatestStart time.Time
}

// mergeIntervals ordena y fusiona intervalos solapados o contiguos
// (RN-DIS-02: "dos factores que se solapan deben restar una sola vez").
// Intervalos inválidos (End <= Start) se descartan silenciosamente: nunca
// deberían llegar aquí, pero un dato así no puede representar una
// ocupación real.
func mergeIntervals(intervals []Interval) []Interval {
	clean := make([]Interval, 0, len(intervals))
	for _, iv := range intervals {
		if iv.valid() {
			clean = append(clean, iv)
		}
	}
	if len(clean) == 0 {
		return nil
	}
	sort.Slice(clean, func(i, j int) bool { return clean[i].Start.Before(clean[j].Start) })

	merged := make([]Interval, 0, len(clean))
	current := clean[0]
	for _, iv := range clean[1:] {
		if !iv.Start.After(current.End) {
			// Solape o contigüidad: RN-DIS-05 trata la contigüidad exacta
			// como "sin cruce", pero para restar disponibilidad la
			// contigüidad SÍ debe fusionarse (dos bloqueos pegados no dejan
			// un hueco de cero minutos entre ellos).
			if iv.End.After(current.End) {
				current.End = iv.End
			}
			continue
		}
		merged = append(merged, current)
		current = iv
	}
	merged = append(merged, current)
	return merged
}

// subtractBusy resta busy (ya fusionado y ordenado por Start) de seg y
// devuelve los intervalos libres resultantes, en orden. Cada intervalo
// libre devuelto ancla su propio inicio: el llamador que genere la rejilla
// sobre cada uno de ellos por separado implementa DEC-084 sin ningún caso
// especial.
func subtractBusy(seg Interval, busy []Interval) []Interval {
	if !seg.valid() {
		return nil
	}
	var free []Interval
	cursor := seg.Start
	for _, b := range busy {
		if !b.Start.Before(seg.End) {
			break // b empieza en o después del cierre de seg: ya no afecta.
		}
		if !b.End.After(seg.Start) {
			continue // b termina antes o al abrir seg: irrelevante.
		}
		if b.Start.After(cursor) {
			free = append(free, Interval{Start: cursor, End: b.Start})
		}
		if b.End.After(cursor) {
			cursor = b.End
		}
		if !cursor.Before(seg.End) {
			return free
		}
	}
	if cursor.Before(seg.End) {
		free = append(free, Interval{Start: cursor, End: seg.End})
	}
	return free
}

// GenerateStarts proyecta los inicios públicos válidos (CA-094-01,
// CA-094-02, CA-094-03): fusiona segments y busy por separado, resta busy
// de cada tramo fusionado, y genera candidatos cada policy.GridStep dentro
// de cada intervalo libre resultante, anclado al inicio de ESE intervalo
// (DEC-084). Un candidato solo se conserva si el servicio completo cabe
// ([inicio, inicio+duración) enteramente dentro del intervalo libre,
// RN-DIS-01/RN-DIS-05) y si cae dentro de
// [policy.EarliestStart, policy.LatestStart] (RN-DIS-04). El resultado
// viene ordenado cronológicamente y sin duplicados, porque segments no se
// solapa entre sí una vez fusionado y cada intervalo libre se recorre una
// sola vez.
//
// GenerateStarts no reserva nada, no consulta nada y no depende de la hora
// real (RN-DIS-03, CA-094-04): todos los instantes -incluidos los límites
// de policy- ya llegan resueltos por el llamador.
func GenerateStarts(segments []Interval, busy []Interval, policy Policy) []time.Time {
	if policy.ServiceDuration <= 0 || policy.GridStep <= 0 {
		return nil
	}
	mergedSegments := mergeIntervals(segments)
	mergedBusy := mergeIntervals(busy)

	var starts []time.Time
	for _, seg := range mergedSegments {
		for _, free := range subtractBusy(seg, mergedBusy) {
			for t := free.Start; !t.Add(policy.ServiceDuration).After(free.End); t = t.Add(policy.GridStep) {
				if t.Before(policy.EarliestStart) {
					continue
				}
				if t.After(policy.LatestStart) {
					break // t solo crece dentro de este intervalo libre.
				}
				starts = append(starts, t)
			}
		}
	}
	return starts
}
