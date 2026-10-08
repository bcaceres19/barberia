<script setup lang="ts">
// Contenido de la sección «Enlace público» de Configuración (issue #304,
// DEC-117): el enlace de reservas de la barbería, siempre a mano para verlo y
// copiarlo si se pierde. Solo lectura: no se edita ni se regenera. La primera
// lectura de una barbería sin enlace lo genera en el servidor.
//
// Copiar usa el portapapeles del navegador; si no está disponible (contexto
// no seguro, permiso denegado) deja el enlace seleccionado en el campo y lo
// dice, para que copiar a mano con Ctrl+C sea un paso, no un callejón sin
// salida. El resultado se anuncia en una región viva (WCAG 4.1.3).
import { computed, onBeforeUnmount, onMounted, ref } from 'vue'
import { BaseAlert, BaseButton, DiamondLoader } from '@/shared/ui'
import { fetchPublicLink } from '../api/publicLinkApi'
import { buildPublicLinkUrl } from '../model/publicLink'

type Status = 'loading' | 'ready' | 'network-error' | 'unexpected-error'
type CopyState = 'idle' | 'copied' | 'manual'

const COPIED_FEEDBACK_MS = 3000
const LOADING_PHRASES = ['Preparando tu enlace'] as const

const status = ref<Status>('loading')
const slug = ref('')
const copyState = ref<CopyState>('idle')
const field = ref<HTMLElement | null>(null)
let resetTimer: ReturnType<typeof setTimeout> | undefined

const url = computed(() =>
  slug.value ? buildPublicLinkUrl(window.location.origin, slug.value) : '',
)

async function load() {
  status.value = 'loading'
  const outcome = await fetchPublicLink()
  if (outcome.kind === 'success') {
    slug.value = outcome.slug
    status.value = 'ready'
  } else {
    status.value = outcome.kind
  }
}

// Selecciona el texto completo del enlace: dejarlo listo para Ctrl+C es lo que
// salva la acción cuando el portapapeles del navegador no responde.
function selectField() {
  const el = field.value
  if (!el) return
  el.focus()
  const selection = window.getSelection()
  selection?.removeAllRanges()
  const range = document.createRange()
  range.selectNodeContents(el)
  selection?.addRange(range)
}

async function copy() {
  clearTimeout(resetTimer)
  try {
    await navigator.clipboard.writeText(url.value)
    copyState.value = 'copied'
    resetTimer = setTimeout(() => (copyState.value = 'idle'), COPIED_FEEDBACK_MS)
  } catch {
    copyState.value = 'manual'
    selectField()
  }
}

onMounted(load)
onBeforeUnmount(() => clearTimeout(resetTimer))
</script>

<template>
  <div class="public-link" :aria-busy="status === 'loading'">
    <div v-if="status === 'loading'" role="status">
      <DiamondLoader label="Preparando tu enlace" layout="inline" :phrases="LOADING_PHRASES" />
    </div>

    <BaseAlert
      v-else-if="status !== 'ready'"
      :variant="status === 'network-error' ? 'warning' : 'danger'"
      :title="status === 'network-error' ? 'No pudimos conectar' : 'No pudimos cargar tu enlace'"
      role="alert"
    >
      {{
        status === 'network-error'
          ? 'Revisa tu conexión e inténtalo de nuevo.'
          : 'Inténtalo de nuevo en unos segundos.'
      }}
      <template #action>
        <BaseButton type="button" @click="load">Reintentar</BaseButton>
      </template>
    </BaseAlert>

    <template v-else>
      <p id="public-link-label" class="public-link__label">Tu enlace de reservas</p>
      <div class="public-link__row">
        <!-- Texto que parte línea en vez de recortarse: el código final es lo que
             distingue este enlace de cualquier otro y debe leerse completo. -->
        <div
          ref="field"
          class="public-link__field"
          role="textbox"
          aria-readonly="true"
          aria-multiline="true"
          aria-labelledby="public-link-label"
          tabindex="0"
        >
          {{ url }}
        </div>
        <div class="public-link__actions">
          <BaseButton type="button" @click="copy">Copiar enlace</BaseButton>
          <a class="public-link__open" :href="url" target="_blank" rel="noopener noreferrer">
            Abrir
            <span class="visually-hidden">(se abre en una pestaña nueva)</span>
          </a>
        </div>
      </div>
      <p class="public-link__hint">
        Compártelo con tus clientes. Es único y solo abre tu reserva; si lo pierdes, vuelve a esta
        pantalla.
      </p>
      <p class="public-link__feedback" role="status" aria-live="polite">
        <template v-if="copyState === 'copied'">Enlace copiado.</template>
        <template v-else-if="copyState === 'manual'">
          No pudimos copiarlo solo: quedó seleccionado, cópialo con Ctrl+C.
        </template>
      </p>
    </template>
  </div>
</template>

<style scoped>
.public-link {
  display: flex;
  flex-direction: column;
  gap: 10px;
  min-width: 0;
}

.public-link__label {
  font-size: var(--font-size-caption);
  font-weight: 600;
  letter-spacing: 0.08em;
  text-transform: uppercase;
  color: var(--color-brand-accent-surface);
}

.public-link__row {
  display: flex;
  flex-wrap: wrap;
  align-items: stretch;
  gap: 12px;
}

/* Enlace completo en monoespaciado para que el código aleatorio se lea sin
   ambigüedad. Ocupa el ancho disponible y las acciones bajan debajo en
   pantallas angostas. */
.public-link__field {
  display: flex;
  flex: 1 1 320px;
  align-items: center;
  min-width: 0;
  min-height: 44px;
  padding: 8px 14px;
  box-sizing: border-box;
  font-family: ui-monospace, 'SFMono-Regular', Menlo, Consolas, monospace;
  font-size: var(--font-size-body-sm);
  line-height: 20px;
  color: var(--color-on-strong);
  overflow-wrap: anywhere;
  user-select: all;
  background: color-mix(in srgb, var(--color-on-strong) 6%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-bottom: var(--border-width-emphasis) solid var(--color-brand-accent-surface);
  border-radius: 2px;
}

.public-link__field:focus-visible {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: 2px;
}

.public-link__actions {
  display: flex;
  flex: 0 0 auto;
  align-items: center;
  gap: 12px;
}

.public-link__open {
  display: inline-flex;
  align-items: center;
  min-height: 44px;
  padding: 0 8px;
  font-size: var(--font-size-body-sm);
  font-weight: 600;
  color: var(--color-brand-accent-surface);
  text-decoration: underline;
  text-underline-offset: 3px;
}

.public-link__open:focus-visible {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: 2px;
}

.public-link__hint {
  margin: 0;
  font-size: var(--font-size-caption);
  line-height: 18px;
  color: var(--color-on-strong-muted);
}

/* La región viva siempre existe (un lector de pantalla solo anuncia cambios en
   una región ya presente) y reserva su alto para que el aviso no empuje nada. */
.public-link__feedback {
  min-height: 18px;
  margin: 0;
  font-size: var(--font-size-caption);
  line-height: 18px;
  color: var(--color-brand-accent-surface);
}

.visually-hidden {
  position: absolute;
  width: 1px;
  height: 1px;
  margin: -1px;
  padding: 0;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}
</style>
