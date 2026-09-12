package availability_test

import (
	"testing"
	"time"

	"system-barbershop/internal/modules/availability"
)

// bogota es una zona horaria fija de referencia para construir instantes
// legibles en las pruebas (RN-DIS-07): el dominio en sí es agnóstico de
// zona, ya que solo trabaja con time.Time absolutos.
var bogota = mustLoadLocation("America/Bogota")

func mustLoadLocation(name string) *time.Location {
	loc, err := time.LoadLocation(name)
	if err != nil {
		panic(err)
	}
	return loc
}

func at(hour, minute int) time.Time {
	return time.Date(2026, time.September, 14, hour, minute, 0, 0, bogota)
}

func atDay(day, hour, minute int) time.Time {
	return time.Date(2026, time.September, day, hour, minute, 0, 0, bogota)
}

// noWindow es una política sin restricción de anticipación/ventana, para
// que las pruebas de generación de franjas dentro de un tramo no tengan que
// preocuparse también por RN-DIS-04.
func noWindowPolicy(duration, grid time.Duration) availability.Policy {
	return availability.Policy{
		ServiceDuration: duration,
		GridStep:        grid,
		EarliestStart:   at(0, 0),
		LatestStart:     atDay(30, 0, 0),
	}
}

func wantTimes(times ...time.Time) []time.Time { return times }

func assertEqualTimes(t *testing.T, got, want []time.Time) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d\ngot:  %v\nwant: %v", len(got), len(want), got, want)
	}
	for i := range got {
		if !got[i].Equal(want[i]) {
			t.Fatalf("got[%d] = %v, want %v\ngot:  %v\nwant: %v", i, got[i], want[i], got, want)
		}
	}
}

// TestGenerateStarts_ExhaustiveFactors cubre CA-094-01 a CA-094-03 con una
// tabla exhaustiva de los factores que RN-DIS-02 exige considerar (jornada,
// duración, citas/bloqueos que ocupan agenda, solapes, rejilla), más las
// fronteras de RN-DIS-01/RN-DIS-05/RN-DIS-06 y DEC-084.
func TestGenerateStarts_ExhaustiveFactors(t *testing.T) {
	tests := []struct {
		name     string
		segments []availability.Interval
		busy     []availability.Interval
		policy   availability.Policy
		want     []time.Time
	}{
		{
			name:     "jornada sin ocupaciones genera toda la rejilla",
			segments: []availability.Interval{{Start: at(9, 0), End: at(9, 45)}},
			busy:     nil,
			policy:   noWindowPolicy(15*time.Minute, 15*time.Minute),
			want:     wantTimes(at(9, 0), at(9, 15), at(9, 30)),
		},
		{
			name:     "hueco insuficiente entre citas no ofrece nada (CA-094-02)",
			segments: []availability.Interval{{Start: at(9, 0), End: at(18, 0)}},
			busy: []availability.Interval{
				{Start: at(9, 0), End: at(10, 0)},
				{Start: at(10, 45), End: at(18, 0)},
			},
			policy: noWindowPolicy(60*time.Minute, 15*time.Minute),
			want:   nil,
		},
		{
			name:     "hueco exacto de 30 min y servicio de 30 se ofrece (RN-DIS-05, semiabierto)",
			segments: []availability.Interval{{Start: at(9, 0), End: at(18, 0)}},
			busy: []availability.Interval{
				{Start: at(9, 0), End: at(10, 0)},
				{Start: at(10, 30), End: at(18, 0)},
			},
			policy: noWindowPolicy(30*time.Minute, 15*time.Minute),
			want:   wantTimes(at(10, 0)),
		},
		{
			name:     "el servicio que termina justo al cierre de la jornada se ofrece",
			segments: []availability.Interval{{Start: at(17, 0), End: at(18, 0)}},
			policy:   noWindowPolicy(60*time.Minute, 15*time.Minute),
			want:     wantTimes(at(17, 0)),
		},
		{
			name:     "duración excede el cierre de la jornada: nada se ofrece",
			segments: []availability.Interval{{Start: at(17, 30), End: at(18, 0)}},
			policy:   noWindowPolicy(60*time.Minute, 15*time.Minute),
			want:     nil,
		},
		{
			name:     "servicio de 25 con rejilla de 15 deja residuo de 5 (RN-DIS-06)",
			segments: []availability.Interval{{Start: at(9, 0), End: at(10, 0)}},
			policy:   noWindowPolicy(25*time.Minute, 15*time.Minute),
			// 9:00-9:25, 9:15-9:40, 9:30-9:55 caben; 9:35 no cabría si
			// existiera (no es múltiplo de la rejilla, no se genera).
			want: wantTimes(at(9, 0), at(9, 15), at(9, 30)),
		},
		{
			name:     "dos factores solapados (bloqueo dentro de almuerzo) restan una sola vez (RN-DIS-02)",
			segments: []availability.Interval{{Start: at(9, 0), End: at(15, 0)}},
			busy: []availability.Interval{
				{Start: at(13, 0), End: at(14, 0)},   // almuerzo
				{Start: at(13, 30), End: at(13, 45)}, // bloqueo dentro del almuerzo
			},
			policy: noWindowPolicy(60*time.Minute, 60*time.Minute),
			want:   wantTimes(at(9, 0), at(10, 0), at(11, 0), at(12, 0), at(14, 0)),
		},
		{
			name: "DEC-084: la rejilla reinicia desde el instante en que termina la interrupción",
			segments: []availability.Interval{
				{Start: at(9, 0), End: at(18, 0)},
			},
			busy: []availability.Interval{
				{Start: at(9, 0), End: at(13, 47)},
			},
			policy: noWindowPolicy(15*time.Minute, 15*time.Minute),
			// Reinicia en 13:47 (no espera a 14:00, múltiplo de la rejilla
			// original del tramo desde las 9:00).
			want: wantTimes(
				at(13, 47), at(14, 2), at(14, 17), at(14, 32), at(14, 47),
				at(15, 2), at(15, 17), at(15, 32), at(15, 47),
				at(16, 2), at(16, 17), at(16, 32), at(16, 47),
				at(17, 2), at(17, 17), at(17, 32),
			),
		},
		{
			name:     "citas contiguas exactas no dejan hueco (RN-DIS-05) y no se ofrece nada entre ellas",
			segments: []availability.Interval{{Start: at(9, 0), End: at(12, 0)}},
			busy: []availability.Interval{
				{Start: at(9, 30), End: at(10, 0)},
				{Start: at(10, 0), End: at(10, 30)}, // contigua a la anterior
			},
			policy: noWindowPolicy(30*time.Minute, 30*time.Minute),
			want:   wantTimes(at(9, 0), at(10, 30), at(11, 0), at(11, 30)),
		},
		{
			name: "turno nocturno que cruza medianoche es un solo intervalo absoluto",
			segments: []availability.Interval{
				{Start: at(22, 0), End: atDay(15, 2, 0)}, // 22:00 -> 02:00 del día siguiente
			},
			busy:   nil,
			policy: noWindowPolicy(60*time.Minute, 60*time.Minute),
			want:   wantTimes(at(22, 0), at(23, 0), atDay(15, 0, 0), atDay(15, 1, 0)),
		},
		{
			name:     "festivo/día no laborable: sin tramos, sin franjas",
			segments: nil,
			busy:     []availability.Interval{{Start: at(0, 0), End: atDay(15, 0, 0)}},
			policy:   noWindowPolicy(30*time.Minute, 30*time.Minute),
			want:     nil,
		},
		{
			name:     "vacaciones/emergencia cubren toda la jornada: nada disponible",
			segments: []availability.Interval{{Start: at(9, 0), End: at(18, 0)}},
			busy:     []availability.Interval{{Start: at(9, 0), End: at(18, 0)}},
			policy:   noWindowPolicy(30*time.Minute, 15*time.Minute),
			want:     nil,
		},
		{
			name: "dos tramos del mismo día (excepción con varios segmentos) se procesan por separado",
			segments: []availability.Interval{
				{Start: at(9, 0), End: at(12, 0)},
				{Start: at(14, 0), End: at(18, 0)},
			},
			policy: noWindowPolicy(60*time.Minute, 60*time.Minute),
			want:   wantTimes(at(9, 0), at(10, 0), at(11, 0), at(14, 0), at(15, 0), at(16, 0), at(17, 0)),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := availability.GenerateStarts(tc.segments, tc.busy, tc.policy)
			assertEqualTimes(t, got, tc.want)
		})
	}
}

// TestGenerateStarts_AnticipationAndWindow cubre RN-DIS-04/CA-094-04: los
// límites de anticipación y ventana se aplican como instantes ya resueltos
// por el llamador (reloj del servidor), nunca recalculados aquí.
func TestGenerateStarts_AnticipationAndWindow(t *testing.T) {
	segments := []availability.Interval{{Start: at(9, 0), End: at(18, 0)}}

	t.Run("candidato antes de la anticipación mínima se descarta", func(t *testing.T) {
		policy := availability.Policy{
			ServiceDuration: 30 * time.Minute,
			GridStep:        30 * time.Minute,
			EarliestStart:   at(10, 0),
			LatestStart:     atDay(30, 0, 0),
		}
		got := availability.GenerateStarts(segments, nil, policy)
		assertEqualTimes(t, got, wantTimes(at(10, 0), at(10, 30), at(11, 0), at(11, 30), at(12, 0), at(12, 30), at(13, 0), at(13, 30), at(14, 0), at(14, 30), at(15, 0), at(15, 30), at(16, 0), at(16, 30), at(17, 0), at(17, 30)))
	})

	t.Run("candidato después de la ventana máxima se descarta", func(t *testing.T) {
		policy := availability.Policy{
			ServiceDuration: 60 * time.Minute,
			GridStep:        60 * time.Minute,
			EarliestStart:   at(0, 0),
			LatestStart:     at(11, 0),
		}
		got := availability.GenerateStarts(segments, nil, policy)
		assertEqualTimes(t, got, wantTimes(at(9, 0), at(10, 0), at(11, 0)))
	})

	t.Run("LatestStart es límite inclusivo", func(t *testing.T) {
		policy := availability.Policy{
			ServiceDuration: 30 * time.Minute,
			GridStep:        30 * time.Minute,
			EarliestStart:   at(9, 0),
			LatestStart:     at(9, 30),
		}
		got := availability.GenerateStarts(segments, nil, policy)
		assertEqualTimes(t, got, wantTimes(at(9, 0), at(9, 30)))
	})
}

// TestGenerateStarts_DeterministicOrderAcrossMultipleDays cubre el orden
// estable exigido por el "Trabajo requerido" del prompt de HU-094 (contrato
// público con orden estable), pasando segmentos de dos fechas distintas
// desordenados a propósito.
func TestGenerateStarts_DeterministicOrderAcrossMultipleDays(t *testing.T) {
	segments := []availability.Interval{
		{Start: atDay(15, 9, 0), End: atDay(15, 10, 0)},
		{Start: at(9, 0), End: at(10, 0)}, // día 14, llega después en el slice
	}
	policy := noWindowPolicy(30*time.Minute, 30*time.Minute)
	got := availability.GenerateStarts(segments, nil, policy)
	assertEqualTimes(t, got, wantTimes(at(9, 0), at(9, 30), atDay(15, 9, 0), atDay(15, 9, 30)))
}

// TestGenerateStarts_InvalidPolicy_ReturnsNil documenta el comportamiento
// defensivo ante una política degenerada (duración o rejilla no positivas):
// nunca debe generar un candidato ni entrar en bucle infinito.
func TestGenerateStarts_InvalidPolicy_ReturnsNil(t *testing.T) {
	segments := []availability.Interval{{Start: at(9, 0), End: at(18, 0)}}

	if got := availability.GenerateStarts(segments, nil, availability.Policy{
		ServiceDuration: 0,
		GridStep:        15 * time.Minute,
		EarliestStart:   at(0, 0),
		LatestStart:     atDay(30, 0, 0),
	}); got != nil {
		t.Fatalf("duración cero: got %v, want nil", got)
	}

	if got := availability.GenerateStarts(segments, nil, availability.Policy{
		ServiceDuration: 15 * time.Minute,
		GridStep:        0,
		EarliestStart:   at(0, 0),
		LatestStart:     atDay(30, 0, 0),
	}); got != nil {
		t.Fatalf("rejilla cero: got %v, want nil", got)
	}
}
