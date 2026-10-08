// Vocabulario de la barbería para la reserva pública (DEC-119). El cascarón
// (`layouts/PublicBookingLayout.vue`) lo lee una vez del perfil público del
// enlace y lo ofrece a todas las pantallas del recorrido: cada una puede abrirse
// directamente desde un enlace, así que ninguna puede depender de haber pasado
// por la anterior. Es un `provide`/`inject` acotado a este árbol, no estado
// global; sin cascarón (pruebas de componente en aislamiento) devuelve los
// valores iniciales, es decir, la interfaz de siempre.
import { computed, inject, type ComputedRef, type InjectionKey } from 'vue'
import { buildVocabularyFromTerms, type Vocabulary } from '@/shared/model'

export const publicVocabularyKey: InjectionKey<ComputedRef<Vocabulary>> = Symbol('publicVocabulary')

const initialVocabulary = computed(() => buildVocabularyFromTerms(null))

export function usePublicVocabulary(): ComputedRef<Vocabulary> {
  return inject(publicVocabularyKey, initialVocabulary)
}
