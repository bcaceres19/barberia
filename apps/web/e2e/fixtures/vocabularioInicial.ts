// Vocabulario que recibe una barbería que nunca lo configuró (DEC-110, DEC-119): las
// respuestas simuladas del perfil público y del turno del cliente lo incluyen porque el
// contrato lo exige (`vocabulary`). Con estos valores la interfaz dice lo de siempre.
export const INITIAL_VOCABULARY = {
  businessTerm: 'barbería',
  businessTermGender: 'feminine',
  professionalTerm: 'barbero',
  professionalTermPlural: 'barberos',
  professionalTermGender: 'masculine',
} as const
