<script setup lang="ts">
// Pantalla "Configuración" (HU-020, DEC-110): nombre, zona horaria y contacto
// de la barbería activa (HU-020) más la marca y el vocabulario de toda la
// barbería y las preferencias de pantalla de ESTE dispositivo. Mismo lenguaje
// que Agenda, Servicios y Barberos (tinta, latón, campos reglados, rombo de
// estado, entrada escalonada, carga con el rombo).
//
// Dos clases de ajuste, y la pantalla lo dice en cada sección:
//  - Este dispositivo (modo, tamaño de texto, animaciones): se aplican al
//    instante y viven en el navegador; no hay nada que guardar.
//  - Toda la barbería (nombre, acento, vocabulario, zona, contacto): se
//    editan en un borrador y solo un guardado confirmado por el servidor los
//    aplica. Un error recuperable NUNCA borra lo que la persona escribió
//    (CA-020-08). Marca y datos básicos son dos recursos distintos, así que se
//    envían solo los que cambiaron y cada uno se confirma por separado.
// El acento elegido se previsualiza en toda la app mientras se edita; si se
// descarta o se sale sin guardar, vuelve al confirmado.
import { computed, nextTick, onBeforeUnmount, onMounted, reactive, ref, watch } from 'vue'
import { PAGE_MIN_HOLD_MS, useMinHoldLoading, useToast, useVocabulary } from '@/shared/composables'
import {
  appearance,
  BRAND_ACCENTS,
  buildVocabulary,
  DEFAULT_BRAND,
  isDefaultAppearance,
  normalizeTerm,
  resetAppearance,
  resolvedTheme,
  setAccentPreview,
  setBrand,
  setMotion,
  setTextScale,
  setTheme,
  TEXT_SCALES,
  TEXT_SCALE_KEYS,
  validateTerm,
  type BrandAccentKey,
  type BrandSettings,
  type Gender,
  type MotionMode,
  type PanelProfile,
  type TextScaleKey,
  type ThemeMode,
} from '@/shared/model'
import { BaseAlert, BaseButton, BaseInput, DiamondLoader } from '@/shared/ui'
import { updateBarbershopName } from '@/modules/auth'
import { fetchBrand, saveBrand } from '../api/brandApi'
import { fetchBarbershopSettings, saveBarbershopSettings } from '../api/settingsApi'
import { toFormValues, type BarbershopSettingsFormValues } from '../model/barbershopSettings'
import {
  brandForSave,
  BUSINESS_PRESETS,
  PROFESSIONAL_PRESETS,
  sameBarbershop,
  sameBrand,
  suggestPlural,
} from '../model/settingsDraft'
import {
  validateContactEmail,
  validateContactPhone,
  validateName,
  validateTimezone,
} from '../validation/settingsValidation'
import OptionGroup from '../components/OptionGroup.vue'
import SaveBar from '../components/SaveBar.vue'
import SettingsIndex from '../components/SettingsIndex.vue'
import SettingsPanel from '../components/SettingsPanel.vue'
import SwitchField from '../components/SwitchField.vue'
import TimezoneField from '../components/TimezoneField.vue'

type LoadStatus = 'loading' | 'ready' | 'load-error'
type GroupError = 'validation' | 'network' | 'unexpected'

const LOADING_PHRASES = [
  'Ajustando la casa',
  'Afinando los detalles',
  'Alineando la marca',
  'Todo a su hora',
] as const

const v = useVocabulary()
const toast = useToast()

const loadStatus = ref<LoadStatus>('loading')
const saving = ref(false)
const attempted = ref(false)
const shopError = ref<GroupError | null>(null)
const brandError = ref<GroupError | null>(null)

const shop = reactive<BarbershopSettingsFormValues>({
  name: '',
  timezone: '',
  contactEmail: '',
  contactPhone: '',
})
const brand = reactive<BrandSettings>({ ...DEFAULT_BRAND })
// Último valor confirmado por el servidor: contra él se mide qué cambió.
const savedShop = ref<BarbershopSettingsFormValues>({ ...shop })
const savedBrand = ref<BrandSettings>({ ...DEFAULT_BRAND })

const { start: startLoadingHold, hold: holdLoadingReveal } = useMinHoldLoading(PAGE_MIN_HOLD_MS)
let requestToken = 0

async function load() {
  const token = ++requestToken
  loadStatus.value = 'loading'
  startLoadingHold()
  const [shopOutcome, brandOutcome] = await Promise.all([fetchBarbershopSettings(), fetchBrand()])
  if (token !== requestToken) return
  holdLoadingReveal(() => {
    if (token !== requestToken) return
    if (shopOutcome.kind === 'success' && brandOutcome.kind === 'success') {
      Object.assign(shop, toFormValues(shopOutcome.settings))
      Object.assign(brand, brandOutcome.brand)
      savedShop.value = { ...shop }
      savedBrand.value = { ...brandOutcome.brand }
      loadStatus.value = 'ready'
      return
    }
    loadStatus.value = 'load-error'
  })
}

onMounted(load)

// --- Borrador --------------------------------------------------------------

const shopDirty = computed(() => !sameBarbershop(shop, savedShop.value))
const brandDirty = computed(() => !sameBrand(brand, savedBrand.value))
const dirty = computed(() => shopDirty.value || brandDirty.value)

const dirtySummary = computed(() => {
  const parts: string[] = []
  if (brandDirty.value) parts.push('marca y vocabulario')
  if (shopDirty.value) parts.push('datos básicos')
  const text = parts.join(' y ')
  return text.charAt(0).toLocaleUpperCase('es') + text.slice(1)
})

// Vocabulario del borrador para la vista previa: si un término aún no es válido
// se conserva el confirmado, así la vista previa nunca muestra basura a medias.
const draftVocabulary = computed(() =>
  buildVocabulary({
    ...brand,
    businessTerm: validateTerm(brand.businessTerm)
      ? savedBrand.value.businessTerm
      : normalizeTerm(brand.businessTerm),
    professionalTerm: validateTerm(brand.professionalTerm)
      ? savedBrand.value.professionalTerm
      : normalizeTerm(brand.professionalTerm),
    professionalTermPlural: validateTerm(brand.professionalTermPlural)
      ? savedBrand.value.professionalTermPlural
      : normalizeTerm(brand.professionalTermPlural),
  }),
)

type FieldErrors = Partial<
  Record<
    | 'name'
    | 'timezone'
    | 'contactEmail'
    | 'contactPhone'
    | 'businessTerm'
    | 'professionalTerm'
    | 'professionalTermPlural',
    string
  >
>
const fieldErrors = ref<FieldErrors>({})

function runValidation(): FieldErrors {
  const errors: FieldErrors = {}
  const set = (key: keyof FieldErrors, message: string | undefined) => {
    if (message) errors[key] = message
  }
  set('name', validateName(shop.name))
  set('timezone', validateTimezone(shop.timezone))
  set('contactEmail', validateContactEmail(shop.contactEmail))
  set('contactPhone', validateContactPhone(shop.contactPhone))
  set('businessTerm', validateTerm(brand.businessTerm))
  set('professionalTerm', validateTerm(brand.professionalTerm))
  set('professionalTermPlural', validateTerm(brand.professionalTermPlural))
  return errors
}

function revalidate() {
  if (attempted.value) fieldErrors.value = runValidation()
}

function onShopInput<K extends keyof BarbershopSettingsFormValues>(
  field: K,
  value: string | number,
) {
  shop[field] = String(value)
  revalidate()
}

function onBusinessTermInput(value: string | number) {
  brand.businessTerm = String(value)
  revalidate()
}

// El plural acompaña al singular mientras siga siendo la sugerencia automática:
// en cuanto la persona lo corrige a mano, deja de pisarlo.
function onProfessionalTermInput(value: string | number) {
  const next = String(value)
  const follows =
    !brand.professionalTermPlural ||
    normalizeTerm(brand.professionalTermPlural) === suggestPlural(brand.professionalTerm)
  brand.professionalTerm = next
  if (follows) brand.professionalTermPlural = suggestPlural(next)
  revalidate()
}

function onProfessionalPluralInput(value: string | number) {
  brand.professionalTermPlural = String(value)
  revalidate()
}

function applyBusinessPreset(index: number) {
  const preset = BUSINESS_PRESETS[index]
  if (!preset) return
  brand.businessTerm = preset.term
  brand.businessTermGender = preset.gender
  revalidate()
}

function applyProfessionalPreset(index: number) {
  const preset = PROFESSIONAL_PRESETS[index]
  if (!preset) return
  brand.professionalTerm = preset.singular
  brand.professionalTermPlural = preset.plural
  brand.professionalTermGender = preset.gender
  revalidate()
}

const isBusinessPreset = (index: number) =>
  normalizeTerm(brand.businessTerm) === BUSINESS_PRESETS[index]?.term &&
  brand.businessTermGender === BUSINESS_PRESETS[index]?.gender
const isProfessionalPreset = (index: number) => {
  const p = PROFESSIONAL_PRESETS[index]
  return (
    !!p &&
    normalizeTerm(brand.professionalTerm) === p.singular &&
    normalizeTerm(brand.professionalTermPlural) === p.plural &&
    brand.professionalTermGender === p.gender
  )
}

// El acento se previsualiza en toda la app mientras se edita.
watch(
  () => brand.accent,
  (accent) => {
    if (loadStatus.value === 'ready') setAccentPreview(accent)
  },
)
onBeforeUnmount(() => setAccentPreview(null))

// --- Guardar y descartar ----------------------------------------------------

async function focusFirstError() {
  await nextTick()
  const first = document.querySelector<HTMLElement>(
    '.settings-page [aria-invalid="true"], .settings-page [data-invalid="true"] input',
  )
  first?.focus()
}

async function onSave() {
  // Guardia de doble envío: la asignación síncrona bloquea un segundo clic.
  if (saving.value) return

  attempted.value = true
  const errors = runValidation()
  fieldErrors.value = errors
  if (Object.keys(errors).length > 0) {
    void focusFirstError()
    return
  }

  saving.value = true
  shopError.value = null
  brandError.value = null
  let savedSomething = false

  if (shopDirty.value) {
    const outcome = await saveBarbershopSettings(shop)
    if (outcome.kind === 'success') {
      Object.assign(shop, toFormValues(outcome.settings))
      savedShop.value = { ...shop }
      // Solo el valor confirmado por el servidor cambia la cabecera (CA-020-02).
      updateBarbershopName(outcome.settings.name)
      savedSomething = true
    } else {
      shopError.value =
        outcome.kind === 'validation-error'
          ? 'validation'
          : outcome.kind === 'network-error'
            ? 'network'
            : 'unexpected'
    }
  }

  if (brandDirty.value) {
    const outcome = await saveBrand(brandForSave(brand))
    if (outcome.kind === 'success') {
      Object.assign(brand, outcome.brand)
      savedBrand.value = { ...outcome.brand }
      setBrand(outcome.brand)
      setAccentPreview(null)
      savedSomething = true
    } else {
      brandError.value =
        outcome.kind === 'validation-error'
          ? 'validation'
          : outcome.kind === 'network-error'
            ? 'network'
            : 'unexpected'
    }
  }

  saving.value = false
  if (savedSomething) {
    attempted.value = false
    fieldErrors.value = {}
    toast.success('Configuración guardada', {
      detail: `Los cambios ya están aplicados en ${v.value.theBusiness}.`,
    })
  }
}

function onDiscard() {
  if (saving.value) return
  Object.assign(shop, savedShop.value)
  Object.assign(brand, savedBrand.value)
  setAccentPreview(null)
  attempted.value = false
  fieldErrors.value = {}
  shopError.value = null
  brandError.value = null
}

// --- Preferencias de este dispositivo ---------------------------------------

const THEME_OPTIONS: readonly { value: ThemeMode; label: string; hint: string }[] = [
  { value: 'ink', label: 'Tinta', hint: 'Oscuro, como hasta ahora' },
  { value: 'ivory', label: 'Marfil', hint: 'Claro y cálido' },
  { value: 'auto', label: 'Automático', hint: 'Sigue a tu sistema' },
]

const TEXT_OPTIONS = TEXT_SCALE_KEYS.map((key) => ({
  value: key,
  label: TEXT_SCALES[key].label,
}))

const GENDER_OPTIONS: readonly { value: Gender; label: string }[] = [
  { value: 'feminine', label: 'Femenino' },
  { value: 'masculine', label: 'Masculino' },
]

// Perfil del panel (DEC-115): se escribe con las palabras de la barbería confirmadas, no
// con las del borrador, para que la opción diga lo mismo que el resto del panel.
const PROFILE_OPTIONS = computed<readonly { value: PanelProfile; label: string; hint: string }[]>(
  () => [
    {
      value: 'shop',
      label: `${v.value.Business} con equipo`,
      hint: `Gestión de ${v.value.professionals}, servicios por ${v.value.professional} y selector de ${v.value.professional} en la agenda.`,
    },
    {
      value: 'solo',
      label: `${v.value.Professional} individual`,
      hint: 'Para quien trabaja solo: agenda, servicios, horarios y tu perfil, sin gestión de equipo.',
    },
  ],
)

const ACCENT_OPTIONS = BRAND_ACCENTS.map((a) => ({ value: a.key, label: a.label }))

/** Valor de cada acento tal como se verá en el modo vigente. */
function swatchColor(key: BrandAccentKey): string {
  const accent = BRAND_ACCENTS.find((a) => a.key === key)!
  return resolvedTheme.value === 'ivory' ? accent.ivory : accent.ink
}

// El sistema ya pide menos movimiento: la elección de "completas" no lo anula.
const systemReducesMotion =
  typeof window !== 'undefined' &&
  typeof window.matchMedia === 'function' &&
  window.matchMedia('(prefers-reduced-motion: reduce)').matches

const reduceMotion = computed({
  get: () => appearance.motion === 'reduced',
  set: (on: boolean) => setMotion((on ? 'reduced' : 'full') as MotionMode),
})

// --- Índice de secciones ----------------------------------------------------

const SECTIONS = [
  { id: 'configuracion-pantalla', number: '01', label: 'Pantalla' },
  { id: 'configuracion-perfil', number: '02', label: 'Perfil del panel' },
  { id: 'configuracion-marca', number: '03', label: 'Marca' },
  { id: 'configuracion-vocabulario', number: '04', label: 'Vocabulario' },
  { id: 'configuracion-hora', number: '05', label: 'Hora' },
  { id: 'configuracion-contacto', number: '06', label: 'Contacto' },
  { id: 'configuracion-reservas', number: '07', label: 'Reservas' },
] as const

const activeSection = ref<string>(SECTIONS[0].id)
const pageRef = ref<HTMLElement | null>(null)

function scrollParent(el: HTMLElement): HTMLElement | null {
  let node = el.parentElement
  while (node) {
    if (['auto', 'scroll'].includes(getComputedStyle(node).overflowY)) return node
    node = node.parentElement
  }
  return null
}

// La sección "activa" es la última cuya cabecera ya pasó la línea de lectura
// (un tercio de la ventana). Se recalcula completa en cada aviso del
// observador, no solo con las entradas que cambiaron: un salto largo de
// desplazamiento no deja el índice atrás.
function computeActiveSection() {
  const page = pageRef.value
  if (!page) return
  const root = scrollParent(page)
  const rootTop = root ? root.getBoundingClientRect().top : 0
  const rootHeight = root ? root.clientHeight : window.innerHeight
  const line = rootTop + rootHeight * 0.33
  let current: string = SECTIONS[0].id
  for (const section of SECTIONS) {
    const el = document.getElementById(section.id)
    if (el && el.getBoundingClientRect().top <= line) current = section.id
  }
  // Al fondo del todo, la última sección gana aunque no alcance la línea.
  if (root && root.scrollTop + root.clientHeight >= root.scrollHeight - 4) {
    current = SECTIONS[SECTIONS.length - 1].id
  }
  activeSection.value = current
}

let scroller: HTMLElement | null = null
let scrollFrame = 0

function onScroll() {
  if (scrollFrame) return
  scrollFrame = requestAnimationFrame(() => {
    scrollFrame = 0
    computeActiveSection()
  })
}

function observeSections() {
  const page = pageRef.value
  if (!page) return
  scroller?.removeEventListener('scroll', onScroll)
  scroller = scrollParent(page)
  scroller?.addEventListener('scroll', onScroll, { passive: true })
  computeActiveSection()
}

function goToSection(id: string) {
  activeSection.value = id
  const el = document.getElementById(id)
  if (!el) return
  const reduced =
    appearance.motion === 'reduced' ||
    (typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches)
  el.scrollIntoView({ behavior: reduced ? 'auto' : 'smooth', block: 'start' })
}

const previewClass = computed(() => ({ 'settings-page__preview--changed': brandDirty.value }))
</script>

<template>
  <section ref="pageRef" class="settings-page" aria-labelledby="settings-page-title">
    <header class="settings-page__header">
      <h1 id="settings-page-title" class="settings-page__title">Configuración</h1>
      <p class="settings-page__subtitle">
        Ajusta cómo se ve NAVA en este dispositivo y cómo se llama {{ v.theBusiness }}.
      </p>
      <span class="settings-page__rule" aria-hidden="true"
        ><span class="settings-page__rule-line" /><span class="settings-page__rule-diamond" /><span
          class="settings-page__rule-line"
      /></span>
    </header>

    <Transition name="settings-content" mode="out-in" @after-enter="observeSections">
      <div
        v-if="loadStatus === 'loading'"
        key="loading"
        class="settings-page__state"
        role="status"
        aria-live="polite"
      >
        <DiamondLoader
          label="Cargando la configuración…"
          layout="inline"
          :phrases="LOADING_PHRASES"
        />
        <div class="settings-page__skeletons" aria-hidden="true">
          <span v-for="n in 3" :key="n" class="settings-page__skeleton" :style="{ '--n': n }" />
        </div>
      </div>

      <BaseAlert
        v-else-if="loadStatus === 'load-error'"
        key="error"
        variant="warning"
        title="No pudimos cargar la configuración"
        role="alert"
      >
        Revisa tu conexión e inténtalo de nuevo.
        <template #action>
          <BaseButton variant="secondary" type="button" @click="load">Reintentar</BaseButton>
        </template>
      </BaseAlert>

      <div v-else key="ready" class="settings-page__layout">
        <SettingsIndex :items="SECTIONS" :active="activeSection" @select="goToSection" />

        <form class="settings-page__form settings-ink" novalidate @submit.prevent="onSave">
          <BaseAlert
            v-if="shopError === 'validation'"
            variant="danger"
            title="No pudimos guardar los datos básicos"
            role="alert"
          >
            Revisa los datos, especialmente la zona horaria: debe ser una zona IANA reconocida (ej.
            America/Bogota).
          </BaseAlert>
          <BaseAlert
            v-if="brandError === 'validation'"
            variant="danger"
            title="No pudimos guardar la marca ni el vocabulario"
            role="alert"
          >
            Revisa las palabras: usa solo letras, espacios, guion o apóstrofo.
          </BaseAlert>
          <BaseAlert
            v-if="shopError === 'network' || brandError === 'network'"
            variant="warning"
            title="No pudimos conectar"
            role="alert"
          >
            Revisa tu conexión e inténtalo de nuevo. No perdiste lo que escribiste.
          </BaseAlert>
          <BaseAlert
            v-if="shopError === 'unexpected' || brandError === 'unexpected'"
            variant="danger"
            title="Ocurrió un error inesperado"
            role="alert"
          >
            Inténtalo de nuevo en unos segundos. No perdiste lo que escribiste.
          </BaseAlert>

          <!-- 01 · Pantalla (este dispositivo) -->
          <SettingsPanel
            id="configuracion-pantalla"
            number="01"
            title="Pantalla"
            description="Cómo se ve NAVA en este aparato. Se aplica al instante y no hace falta guardar."
            scope="device"
            scope-label="Solo este dispositivo"
            style="--panel-index: 0"
          >
            <div class="settings-field">
              <span class="settings-field__label" id="field-theme">Modo</span>
              <OptionGroup
                :model-value="appearance.theme"
                :options="THEME_OPTIONS"
                label="Modo de la pantalla"
                class="settings-themes"
                @update:model-value="setTheme"
              >
                <template #default="{ option, selected }">
                  <span class="theme-tile" :class="{ 'theme-tile--selected': selected }">
                    <span
                      class="theme-tile__mini"
                      :class="`theme-tile__mini--${option.value}`"
                      aria-hidden="true"
                    >
                      <span class="theme-tile__bar" />
                      <span class="theme-tile__body">
                        <span class="theme-tile__line theme-tile__line--title" />
                        <span class="theme-tile__line" />
                        <span class="theme-tile__line theme-tile__line--short" />
                        <span class="theme-tile__accent" />
                      </span>
                    </span>
                    <span class="theme-tile__name">
                      {{ option.label }}
                      <span v-if="selected" class="theme-tile__check" aria-hidden="true">✓</span>
                    </span>
                    <span class="theme-tile__hint">{{ option.hint }}</span>
                  </span>
                </template>
              </OptionGroup>
            </div>

            <div class="settings-field">
              <span class="settings-field__label">Tamaño del texto</span>
              <OptionGroup
                :model-value="appearance.textScale"
                :options="TEXT_OPTIONS"
                label="Tamaño del texto"
                class="settings-sizes"
                @update:model-value="(value: TextScaleKey) => setTextScale(value)"
              >
                <template #default="{ option, selected }">
                  <span class="size-step" :class="{ 'size-step--selected': selected }">
                    <span
                      class="size-step__glyph"
                      :class="`size-step__glyph--${option.value}`"
                      aria-hidden="true"
                      >Aa</span
                    >
                    <span class="size-step__label">{{ option.label }}</span>
                  </span>
                </template>
              </OptionGroup>
              <p class="settings-field__hint">
                Escala toda la interfaz. En pantallas angostas se limita para que nada se corte.
              </p>
            </div>

            <div class="settings-field">
              <SwitchField
                v-model="reduceMotion"
                label="Reducir animaciones"
                :hint="
                  systemReducesMotion
                    ? 'Tu sistema ya pide menos movimiento, así que se respeta siempre.'
                    : 'Quita transiciones y destellos. Útil si el movimiento te distrae o te marea.'
                "
              />
            </div>

            <div class="settings-field settings-field--row">
              <BaseButton
                type="button"
                variant="secondary"
                :disabled="isDefaultAppearance(appearance)"
                @click="resetAppearance"
              >
                Restablecer pantalla
              </BaseButton>
            </div>
          </SettingsPanel>

          <!-- 02 · Perfil del panel (toda la barbería) -->
          <SettingsPanel
            id="configuracion-perfil"
            number="02"
            title="Perfil del panel"
            description="Elige qué muestra NAVA según cómo trabajas. Cambia solo lo que ves: no borra ningún dato."
            scope="shop"
            :scope-label="`Para ${v.theBusiness}`"
            style="--panel-index: 1"
          >
            <div class="settings-field">
              <span class="settings-field__label" id="field-panel-profile">Cómo trabajas</span>
              <OptionGroup
                :model-value="brand.panelProfile"
                :options="PROFILE_OPTIONS"
                label="Perfil del panel"
                :disabled="saving"
                class="settings-profiles"
                @update:model-value="(value: PanelProfile) => (brand.panelProfile = value)"
              >
                <template #default="{ option, selected }">
                  <span class="profile-card" :class="{ 'profile-card--selected': selected }">
                    <span class="profile-card__mark" aria-hidden="true" />
                    <span class="profile-card__copy">
                      <span class="profile-card__title">{{ option.label }}</span>
                      <span class="profile-card__hint">{{ option.hint }}</span>
                    </span>
                  </span>
                </template>
              </OptionGroup>
              <p class="settings-field__hint">
                Puedes volver al panel con equipo cuando quieras, por ejemplo al contratar a
                alguien. Si ya tienes más de {{ v.aProfessional }}, el panel individual los sigue
                mostrando donde hace falta elegir.
              </p>
            </div>
          </SettingsPanel>

          <!-- 03 · Marca (toda la barbería) -->
          <SettingsPanel
            id="configuracion-marca"
            number="03"
            title="Marca"
            description="El nombre y el color con los que tu equipo reconoce el panel."
            scope="shop"
            :scope-label="`Para ${v.theBusiness}`"
            style="--panel-index: 2"
          >
            <BaseInput
              :model-value="shop.name"
              name="name"
              label="Nombre"
              required
              :show-required-marker="false"
              :maxlength="120"
              :disabled="saving"
              :error="fieldErrors.name"
              :class="{ 'settings-input--filled': !!shop.name }"
              @update:model-value="(value) => onShopInput('name', value)"
            />

            <div class="settings-field">
              <span class="settings-field__label">Color de acento</span>
              <OptionGroup
                :model-value="brand.accent"
                :options="ACCENT_OPTIONS"
                label="Color de acento"
                :disabled="saving"
                class="settings-accents"
                @update:model-value="(value: BrandAccentKey) => (brand.accent = value)"
              >
                <template #default="{ option, selected }">
                  <span class="accent-swatch" :class="{ 'accent-swatch--selected': selected }">
                    <span class="accent-swatch__stage" aria-hidden="true">
                      <span
                        class="accent-swatch__gem"
                        :style="{ '--swatch': swatchColor(option.value) }"
                      />
                    </span>
                    <span class="accent-swatch__label">{{ option.label }}</span>
                  </span>
                </template>
              </OptionGroup>
              <p class="settings-field__hint">
                Se ve al instante en todo el panel; se guarda para todo tu equipo. Cada color
                conserva el contraste de lectura en modo Tinta y Marfil.
              </p>
            </div>
          </SettingsPanel>

          <!-- 04 · Vocabulario (toda la barbería) -->
          <SettingsPanel
            id="configuracion-vocabulario"
            number="04"
            title="Vocabulario"
            description="Las palabras con las que NAVA nombra tu negocio y a quien atiende los turnos."
            scope="shop"
            :scope-label="`Para ${v.theBusiness}`"
            style="--panel-index: 3"
          >
            <div class="vocab-group">
              <BaseInput
                :model-value="brand.businessTerm"
                name="businessTerm"
                label="Cómo llamas a tu negocio"
                required
                :show-required-marker="false"
                :maxlength="30"
                :disabled="saving"
                :error="fieldErrors.businessTerm"
                :class="{ 'settings-input--filled': !!brand.businessTerm }"
                @update:model-value="onBusinessTermInput"
              />
              <div class="settings-field">
                <span class="settings-field__label settings-field__label--sub">Género</span>
                <OptionGroup
                  :model-value="brand.businessTermGender"
                  :options="GENDER_OPTIONS"
                  label="Género de la palabra del negocio"
                  :disabled="saving"
                  class="settings-genders"
                  @update:model-value="(value: Gender) => (brand.businessTermGender = value)"
                >
                  <template #default="{ option, selected }">
                    <span class="gender-pill" :class="{ 'gender-pill--selected': selected }">{{
                      option.label
                    }}</span>
                  </template>
                </OptionGroup>
              </div>
              <ul class="preset-list" aria-label="Palabras sugeridas para el negocio">
                <li v-for="(preset, index) in BUSINESS_PRESETS" :key="preset.term">
                  <button
                    type="button"
                    class="preset-chip"
                    :class="{ 'preset-chip--active': isBusinessPreset(index) }"
                    :aria-pressed="isBusinessPreset(index)"
                    :disabled="saving"
                    @click="applyBusinessPreset(index)"
                  >
                    {{ preset.term }}
                  </button>
                </li>
              </ul>
            </div>

            <div class="vocab-group">
              <div class="vocab-group__pair">
                <BaseInput
                  :model-value="brand.professionalTerm"
                  name="professionalTerm"
                  label="Cómo llamas a quien atiende"
                  required
                  :show-required-marker="false"
                  :maxlength="30"
                  :disabled="saving"
                  :error="fieldErrors.professionalTerm"
                  :class="{ 'settings-input--filled': !!brand.professionalTerm }"
                  @update:model-value="onProfessionalTermInput"
                />
                <BaseInput
                  :model-value="brand.professionalTermPlural"
                  name="professionalTermPlural"
                  label="En plural"
                  required
                  :show-required-marker="false"
                  :maxlength="30"
                  :disabled="saving"
                  :error="fieldErrors.professionalTermPlural"
                  :class="{ 'settings-input--filled': !!brand.professionalTermPlural }"
                  @update:model-value="onProfessionalPluralInput"
                />
              </div>
              <div class="settings-field">
                <span class="settings-field__label settings-field__label--sub">Género</span>
                <OptionGroup
                  :model-value="brand.professionalTermGender"
                  :options="GENDER_OPTIONS"
                  label="Género de la palabra del profesional"
                  :disabled="saving"
                  class="settings-genders"
                  @update:model-value="(value: Gender) => (brand.professionalTermGender = value)"
                >
                  <template #default="{ option, selected }">
                    <span class="gender-pill" :class="{ 'gender-pill--selected': selected }">{{
                      option.label
                    }}</span>
                  </template>
                </OptionGroup>
              </div>
              <ul class="preset-list" aria-label="Palabras sugeridas para quien atiende">
                <li v-for="(preset, index) in PROFESSIONAL_PRESETS" :key="preset.singular">
                  <button
                    type="button"
                    class="preset-chip"
                    :class="{ 'preset-chip--active': isProfessionalPreset(index) }"
                    :aria-pressed="isProfessionalPreset(index)"
                    :disabled="saving"
                    @click="applyProfessionalPreset(index)"
                  >
                    {{ preset.singular }}
                  </button>
                </li>
              </ul>
            </div>

            <aside
              class="settings-page__preview"
              :class="previewClass"
              aria-label="Vista previa del vocabulario"
            >
              <p class="settings-page__preview-kicker">Así se leerá en tu panel</p>
              <ul class="settings-page__preview-list">
                <li>
                  <span class="settings-page__preview-tag">Menú</span>
                  <span class="settings-page__preview-dock">{{
                    draftVocabulary.Professionals
                  }}</span>
                </li>
                <li>
                  <span class="settings-page__preview-tag">Botón</span>
                  <span class="settings-page__preview-button"
                    >+ Agregar {{ draftVocabulary.professional }}</span
                  >
                </li>
                <li>
                  <span class="settings-page__preview-tag">Aviso</span>
                  <span
                    >Aún no tienes {{ draftVocabulary.professionalsRegistered }} en
                    {{ draftVocabulary.theBusiness }}.</span
                  >
                </li>
                <li>
                  <span class="settings-page__preview-tag">Agenda</span>
                  <span>Elige {{ draftVocabulary.toTheProfessional }} y el servicio.</span>
                </li>
              </ul>
            </aside>
          </SettingsPanel>

          <!-- 05 · Hora (toda la barbería) -->
          <SettingsPanel
            id="configuracion-hora"
            number="05"
            title="Hora"
            description="Toda hora de la agenda, de los horarios y de las reservas se calcula en esta zona, nunca en la de cada dispositivo."
            scope="shop"
            :scope-label="`Para ${v.theBusiness}`"
            style="--panel-index: 4"
          >
            <TimezoneField
              :model-value="shop.timezone"
              label="Zona horaria"
              :disabled="saving"
              :error="fieldErrors.timezone"
              @update:model-value="(value) => onShopInput('timezone', value)"
            />
          </SettingsPanel>

          <!-- 06 · Contacto (toda la barbería) -->
          <SettingsPanel
            id="configuracion-contacto"
            number="06"
            title="Contacto"
            description="Opcional. Es el correo y el teléfono con los que tus clientes pueden escribirte."
            scope="shop"
            :scope-label="`Para ${v.theBusiness}`"
            style="--panel-index: 5"
          >
            <div class="settings-grid">
              <BaseInput
                :model-value="shop.contactEmail"
                type="email"
                name="contactEmail"
                label="Correo de contacto"
                :maxlength="254"
                :disabled="saving"
                :error="fieldErrors.contactEmail"
                :class="{ 'settings-input--filled': !!shop.contactEmail }"
                @update:model-value="(value) => onShopInput('contactEmail', value)"
              />
              <BaseInput
                :model-value="shop.contactPhone"
                type="tel"
                name="contactPhone"
                label="Teléfono de contacto"
                hint="Con indicativo, por ejemplo +573001234567"
                :disabled="saving"
                :error="fieldErrors.contactPhone"
                :class="{ 'settings-input--filled': !!shop.contactPhone }"
                @update:model-value="(value) => onShopInput('contactPhone', value)"
              />
            </div>
          </SettingsPanel>

          <!-- 07 · Reservas públicas: enlace a su propia pantalla (HU-093) -->
          <SettingsPanel
            id="configuracion-reservas"
            number="07"
            title="Reservas"
            description="Cuándo y con cuánta anticipación pueden reservar tus clientes, y cómo cancelan."
            style="--panel-index: 6"
          >
            <RouterLink :to="{ name: 'configuracion-reserva-publica' }" class="settings-link">
              <span class="settings-link__copy">
                <span class="settings-link__title">Reglas de reserva pública</span>
                <span class="settings-link__detail"
                  >Anticipación mínima, ventana, rejilla de horarios y cancelación.</span
                >
              </span>
              <span class="settings-link__arrow" aria-hidden="true">→</span>
            </RouterLink>
          </SettingsPanel>

          <div class="settings-page__save">
            <SaveBar
              v-if="dirty"
              :summary="dirtySummary"
              :saving="saving"
              @discard="onDiscard"
              @save="onSave"
            />
          </div>
        </form>
      </div>
    </Transition>
  </section>
</template>

<style scoped>
/* Superficie tinta de punta a punta (estandar-diseno-visual.md §3), el mismo
   canvas que Agenda, Servicios y Barberos. */
.settings-page {
  --settings-width: 1040px;

  display: flex;
  flex-direction: column;
  gap: 24px;
  min-height: 100%;
  padding: 34px 32px 48px;
  color: var(--color-on-strong);
  background: var(--color-surface-strong);
  box-sizing: border-box;
}

.settings-page__header,
.settings-page__state,
.settings-page__layout,
.settings-page > :deep(.base-alert) {
  width: min(100%, var(--settings-width));
  margin-inline: auto;
}

.settings-page__title {
  margin: 0;
  font-family: var(--font-display);
  font-size: var(--font-size-h1);
  font-weight: var(--font-weight-h1);
  line-height: var(--font-size-h1-line);
  animation: settings-title-enter 520ms var(--motion-easing-standard) both;
}

@media (min-width: 1024px) {
  .settings-page__title {
    font-size: 40px;
    line-height: 46px;
  }
}

.settings-page__subtitle {
  margin: 4px 0 0;
  font-size: var(--font-size-body-sm);
  line-height: var(--font-size-body-sm-line);
  color: var(--color-on-strong-muted);
  animation: settings-title-enter 520ms var(--motion-easing-standard) 80ms both;
}

/* Regla – rombo – regla, el divisor de la casa, que se dibuja al entrar. */
.settings-page__rule {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-top: 18px;
}

.settings-page__rule-line {
  height: 1px;
  flex: 1;
  background: color-mix(in srgb, var(--color-brand-accent-surface) 45%, transparent);
  transform: scaleX(0);
  transform-origin: left center;
  animation: settings-rule-draw 700ms var(--motion-easing-standard) 160ms both;
}

.settings-page__rule-line:last-child {
  transform-origin: right center;
}

.settings-page__rule-diamond {
  width: 7px;
  height: 7px;
  background: var(--color-brand-accent-surface);
  transform: rotate(45deg);
  animation: settings-diamond-pop 420ms cubic-bezier(0.34, 1.56, 0.64, 1) 260ms both;
}

.settings-page__state {
  display: flex;
  flex-direction: column;
  gap: 16px;
  color: var(--color-on-strong-muted);
}

.settings-page__skeletons {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.settings-page__skeleton {
  display: block;
  height: 132px;
  background: color-mix(in srgb, var(--color-on-strong) 8%, transparent);
  border-radius: 3px;
  animation: settings-skeleton-pulse 1400ms ease-in-out infinite;
  animation-delay: calc(var(--n) * 120ms);
}

.settings-page__layout {
  display: grid;
  grid-template-columns: 188px minmax(0, 1fr);
  gap: 32px;
  align-items: start;
}

.settings-page__form {
  display: flex;
  flex-direction: column;
  gap: 20px;
  min-width: 0;
}

.settings-page__save {
  position: sticky;
  bottom: 16px;
  z-index: var(--layer-sticky);
}

/* --- Campos de una sección ------------------------------------------------ */

.settings-field {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.settings-field--row {
  flex-direction: row;
}

.settings-field__label {
  color: var(--color-brand-accent-surface);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  line-height: 14px;
  text-transform: uppercase;
}

.settings-field__label--sub {
  color: var(--color-on-strong-muted);
}

.settings-field__hint {
  margin: 0;
  max-width: 60ch;
  font-size: 12px;
  line-height: 18px;
  color: var(--color-on-strong-muted);
}

.settings-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 20px;
}

/* --- Modo: tres miniaturas del panel ------------------------------------- */

.settings-themes {
  --option-group-gap: 14px;
}

.settings-themes :deep(.option-group__option) {
  flex: 1 1 168px;
  max-width: 244px;
}

.theme-tile {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 100%;
  padding: 10px;
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 3px;
  transition:
    transform var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.theme-tile:hover {
  transform: translateY(-3px);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
}

.theme-tile--selected {
  background: color-mix(in srgb, var(--color-brand-accent-surface) 10%, transparent);
  border-color: var(--color-brand-accent-surface);
  box-shadow: 0 0 0 1px var(--color-brand-accent-surface);
}

/* Miniatura: colores literales a propósito, porque cada una muestra SU modo
   sin importar cuál esté activo. */
.theme-tile__mini {
  --mini-bg: #101b2b;
  --mini-chrome: #101b2b;
  --mini-line: rgb(244 240 231 / 70%);
  --mini-faint: rgb(244 240 231 / 22%);
  --mini-accent: #b8955a;

  position: relative;
  display: flex;
  flex-direction: column;
  height: 78px;
  overflow: hidden;
  background: var(--mini-bg);
  border: 1px solid rgb(127 127 127 / 35%);
  border-radius: 2px;
}

.theme-tile__mini--ivory {
  --mini-bg: #f4f0e7;
  --mini-chrome: #ebe5d9;
  --mini-line: rgb(16 27 43 / 70%);
  --mini-faint: rgb(16 27 43 / 20%);
  --mini-accent: #765c2f;
}

.theme-tile__mini--auto {
  background: linear-gradient(115deg, #101b2b 0 50%, #f4f0e7 50% 100%);
}

.theme-tile__mini--auto .theme-tile__bar {
  background: linear-gradient(115deg, #101b2b 0 50%, #ebe5d9 50% 100%);
}

.theme-tile__mini--auto .theme-tile__line {
  background: linear-gradient(115deg, rgb(244 240 231 / 70%) 0 50%, rgb(16 27 43 / 70%) 50% 100%);
}

.theme-tile__mini--auto .theme-tile__accent {
  background: linear-gradient(115deg, #b8955a 0 50%, #765c2f 50% 100%);
}

.theme-tile__bar {
  height: 14px;
  background: var(--mini-chrome);
  border-bottom: 1px solid var(--mini-accent);
}

.theme-tile__body {
  position: relative;
  display: flex;
  flex: 1;
  flex-direction: column;
  gap: 5px;
  padding: 9px 10px;
}

.theme-tile__line {
  display: block;
  height: 4px;
  width: 78%;
  background: var(--mini-faint);
  border-radius: 1px;
}

.theme-tile__line--title {
  width: 46%;
  height: 6px;
  background: var(--mini-line);
}

.theme-tile__line--short {
  width: 52%;
}

.theme-tile__accent {
  position: absolute;
  right: 10px;
  bottom: 9px;
  width: 16px;
  height: 6px;
  background: var(--mini-accent);
  border-radius: 1px;
}

.theme-tile__name {
  display: flex;
  align-items: center;
  justify-content: space-between;
  font-size: 13px;
  font-weight: 600;
  line-height: 18px;
  color: var(--color-on-strong);
}

.theme-tile__check {
  color: var(--color-brand-accent-surface);
  animation: settings-check-pop 360ms cubic-bezier(0.34, 1.56, 0.64, 1) both;
}

.theme-tile__hint {
  font-size: 11px;
  line-height: 15px;
  color: var(--color-on-strong-muted);
}

/* --- Tamaño del texto: cuatro escalones ---------------------------------- */

.settings-sizes {
  --option-group-gap: 8px;
}

.size-step {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-end;
  gap: 6px;
  width: 92px;
  min-height: 76px;
  padding: 10px 8px 8px;
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-bottom: var(--border-width-emphasis) solid var(--color-field-strong-border);
  border-radius: 3px;
  transition:
    transform var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard);
}

.size-step:hover {
  transform: translateY(-2px);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
}

.size-step--selected {
  background: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-color: var(--color-brand-accent-surface);
}

.size-step__glyph {
  font-family: var(--font-display);
  line-height: 1;
  color: var(--color-on-strong);
  transition: color var(--motion-duration-base) var(--motion-easing-standard);
}

/* La elegida se marca con el filete y el fondo de acento; el glifo conserva el
   color de texto del modo para no perder contraste sobre el tinte (en Marfil,
   el latón sobre su propio tinte no alcanza AA). */
.size-step--selected {
  border-bottom-color: var(--color-brand-accent-surface);
}

.size-step__glyph--small {
  font-size: 16px;
}

.size-step__glyph--normal {
  font-size: 20px;
}

.size-step__glyph--large {
  font-size: 25px;
}

.size-step__glyph--xlarge {
  font-size: 31px;
}

.size-step__label {
  font-size: 11px;
  font-weight: 600;
  line-height: 14px;
  color: var(--color-on-strong-muted);
}

.size-step--selected .size-step__label {
  color: var(--color-on-strong);
}

/* --- Acento: rombos de color --------------------------------------------- */

.settings-accents {
  --option-group-gap: 10px;
}

.accent-swatch {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  width: 92px;
  padding: 8px 4px 10px;
  border: var(--border-width-normal) solid transparent;
  border-radius: 3px;
  transition:
    transform var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.accent-swatch:hover {
  transform: translateY(-3px);
  background: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
}

.accent-swatch--selected {
  background: color-mix(in srgb, var(--swatch-tint, var(--color-on-strong)) 5%, transparent);
  border-color: color-mix(in srgb, var(--color-on-strong) 28%, transparent);
}

.accent-swatch__stage {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
}

.accent-swatch__gem {
  position: relative;
  width: 24px;
  height: 24px;
  background: var(--swatch);
  transform: rotate(45deg);
  transition:
    transform 320ms cubic-bezier(0.34, 1.56, 0.64, 1),
    box-shadow var(--motion-duration-base) var(--motion-easing-standard);
}

.accent-swatch:hover .accent-swatch__gem {
  transform: rotate(45deg) scale(1.12);
}

.accent-swatch--selected .accent-swatch__gem {
  transform: rotate(225deg) scale(1.2);
  box-shadow:
    0 0 0 3px var(--color-surface-strong),
    0 0 0 5px var(--swatch);
}

.accent-swatch__label {
  font-size: 12px;
  font-weight: 600;
  line-height: 16px;
  color: var(--color-on-strong-muted);
}

.accent-swatch--selected .accent-swatch__label {
  color: var(--color-on-strong);
}

/* --- Vocabulario ---------------------------------------------------------- */

.vocab-group {
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding-bottom: 22px;
  border-bottom: var(--border-width-normal) solid var(--color-field-strong-border);
}

.vocab-group:nth-of-type(2) {
  border-bottom: 0;
  padding-bottom: 0;
}

.vocab-group__pair {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 16px;
}

.settings-genders {
  --option-group-gap: 0;
}

/* Perfil del panel (DEC-115): dos tarjetas lado a lado, con el rombo de la casa como
   marca de la elegida. La selección se dice con texto y con borde, no solo con color. */
.settings-profiles {
  --option-group-gap: 12px;
}

.settings-profiles > :deep(.option-group__option) {
  flex: 1 1 260px;
}

.profile-card {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  height: 100%;
  padding: 16px;
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 3px;
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.profile-card:hover {
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
}

.profile-card--selected {
  background: color-mix(in srgb, var(--color-brand-accent-surface) 10%, transparent);
  border-color: var(--color-brand-accent-surface);
  box-shadow: 0 0 0 1px var(--color-brand-accent-surface);
}

.profile-card__mark {
  flex: 0 0 auto;
  width: 12px;
  height: 12px;
  margin-top: 5px;
  border: var(--border-width-normal) solid var(--color-on-strong-muted);
  transform: rotate(45deg);
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.profile-card--selected .profile-card__mark {
  background: var(--color-brand-accent-surface);
  border-color: var(--color-brand-accent-surface);
}

.profile-card__copy {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.profile-card__title {
  font-size: 15px;
  font-weight: 600;
  line-height: 22px;
  color: var(--color-on-strong);
}

.profile-card--selected .profile-card__title {
  color: var(--color-brand-accent-surface);
}

.profile-card__hint {
  font-size: 13px;
  line-height: 19px;
  color: var(--color-on-strong-muted);
}

.gender-pill {
  display: inline-flex;
  align-items: center;
  min-height: 44px;
  padding: 0 18px;
  color: var(--color-on-strong-muted);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  font-size: 13px;
  font-weight: 600;
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.settings-genders > :first-child .gender-pill {
  border-radius: 3px 0 0 3px;
}

.settings-genders > :last-child .gender-pill {
  margin-left: -1px;
  border-radius: 0 3px 3px 0;
}

.gender-pill:hover {
  color: var(--color-on-strong);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
}

.gender-pill--selected {
  position: relative;
  z-index: 1;
  background: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-color: var(--color-brand-accent-surface);
}

.preset-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin: 0;
  padding: 0;
  list-style: none;
}

.preset-chip {
  min-height: 36px;
  padding: 0 14px;
  color: var(--color-on-strong-muted);
  background: transparent;
  border: var(--border-width-normal) dashed
    color-mix(in srgb, var(--color-on-strong) 28%, transparent);
  border-radius: 999px;
  font-family: var(--font-sans);
  font-size: 12px;
  cursor: pointer;
  transition:
    color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard),
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    transform var(--motion-duration-fast) var(--motion-easing-standard);
}

.preset-chip:hover:not(:disabled) {
  color: var(--color-on-strong);
  border-color: var(--color-brand-accent-surface);
  transform: translateY(-1px);
}

.preset-chip:active:not(:disabled) {
  transform: scale(0.97);
}

.preset-chip:focus-visible {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: 2px;
}

.preset-chip--active {
  color: var(--color-on-strong);
  background: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
  border-style: solid;
  border-color: var(--color-brand-accent-surface);
}

.preset-chip:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.settings-page__preview {
  padding: 16px 18px 18px;
  background: color-mix(in srgb, var(--color-brand-accent-surface) 7%, transparent);
  border-left: 3px solid var(--color-brand-accent-surface);
  border-radius: 2px;
}

.settings-page__preview--changed {
  animation: settings-preview-flash 900ms var(--motion-easing-standard);
}

.settings-page__preview-kicker {
  margin: 0 0 12px;
  color: var(--color-brand-accent-surface);
  font-size: 11px;
  font-weight: 600;
  letter-spacing: 0.08em;
  line-height: 14px;
  text-transform: uppercase;
}

.settings-page__preview-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: 13px;
  line-height: 20px;
}

.settings-page__preview-list li {
  display: flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 4px 14px;
}

.settings-page__preview-tag {
  flex: 0 0 56px;
  color: var(--color-on-strong-muted);
  font-size: 11px;
  letter-spacing: 0.06em;
  text-transform: uppercase;
}

.settings-page__preview-dock {
  font-weight: 600;
}

.settings-page__preview-button {
  display: inline-flex;
  align-items: center;
  height: 30px;
  padding: 0 12px;
  background: var(--color-brand-accent-surface);
  color: var(--color-brand-accent-text);
  border-radius: 3px;
  font-size: 12px;
  font-weight: 600;
}

/* --- Reservas: enlace a su propia pantalla ------------------------------- */

.settings-link {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px;
  color: var(--color-on-strong);
  background: color-mix(in srgb, var(--color-on-strong) 3%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-left: 3px solid var(--color-brand-accent-surface);
  border-radius: 2px;
  text-decoration: none;
  transition:
    background-color var(--motion-duration-base) var(--motion-easing-standard),
    border-color var(--motion-duration-base) var(--motion-easing-standard);
}

.settings-link:hover {
  background: color-mix(in srgb, var(--color-brand-accent-surface) 10%, transparent);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 55%, transparent);
  border-left-color: var(--color-brand-accent-surface);
}

.settings-link:focus-visible {
  outline: var(--border-width-emphasis) solid var(--color-focus);
  outline-offset: 2px;
}

.settings-link__copy {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.settings-link__title {
  font-size: 14px;
  font-weight: 600;
  line-height: 20px;
}

.settings-link__detail {
  font-size: 12px;
  line-height: 18px;
  color: var(--color-on-strong-muted);
}

.settings-link__arrow {
  flex: 0 0 auto;
  font-size: 20px;
  color: var(--color-brand-accent-surface);
  transition: transform 280ms cubic-bezier(0.34, 1.56, 0.64, 1);
}

.settings-link:hover .settings-link__arrow {
  transform: translateX(6px);
}

/* --- Campos reglados: mismo tratamiento que Servicios y Barberos --------- */

.settings-ink :deep(.base-input) {
  --input-bg: color-mix(in srgb, var(--color-on-strong) 4%, transparent);
  --input-border-color: color-mix(in srgb, var(--color-on-strong) 12%, transparent);
  --input-border-base-color: color-mix(in srgb, var(--color-on-strong) 30%, transparent);
  --input-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);

  height: 44px;
  color: var(--color-on-strong);
}

.settings-ink :deep(.base-input__label),
.settings-ink :deep(.base-input__required) {
  color: var(--color-brand-accent-surface);
}

.settings-ink :deep(.base-input__hint) {
  color: var(--color-on-strong-muted);
}

.settings-ink :deep(.base-input:hover:not(:disabled):not(.base-input--invalid)) {
  border-color: color-mix(in srgb, var(--color-on-strong) 26%, transparent);
}

.settings-ink :deep(.settings-input--filled .base-input) {
  background-color: color-mix(in srgb, var(--color-on-strong) 6%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.settings-ink :deep(.base-input:disabled),
.settings-ink :deep(.base-input--disabled) {
  background-color: var(--input-bg);
  border-color: var(--input-border-color);
  border-bottom-color: color-mix(in srgb, var(--color-on-strong) 20%, transparent);
  color: var(--color-on-strong-muted);
  opacity: 0.45;
}

.settings-ink :deep(.base-input--invalid) {
  background-color: color-mix(in srgb, var(--color-danger-on-strong) 7%, transparent);
  border-color: var(--input-border-color);
  border-bottom-color: var(--color-danger-on-strong);
}

.settings-ink :deep(.base-input__error) {
  color: var(--color-danger-on-strong);
}

/* Botones: par latón-sólido / tinta-fantasma calibrado para este fondo. */
.settings-ink :deep(.base-button) {
  --btn-focus-ring: 0 0 0 2px var(--color-surface-strong), 0 0 0 4px var(--color-focus);
}

.settings-ink :deep(.base-button--secondary) {
  background-color: transparent;
  color: var(--color-brand-accent-surface);
  border-color: color-mix(in srgb, var(--color-brand-accent-surface) 50%, transparent);
  border-bottom-color: var(--color-brand-accent-surface);
}

.settings-ink :deep(.base-button--secondary:hover:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 12%, transparent);
}

.settings-ink :deep(.base-button--secondary:active:not(:disabled):not(.base-button--loading)) {
  background-color: color-mix(in srgb, var(--color-brand-accent-surface) 20%, transparent);
}

.settings-ink :deep(.base-button:disabled) {
  opacity: 0.4;
}

/* Fundido entre carga/error/listo (estandar-diseno-visual.md §12). */
.settings-content-enter-active,
.settings-content-leave-active {
  transition: opacity var(--motion-duration-base) var(--motion-easing-standard);
}

.settings-content-enter-from,
.settings-content-leave-to {
  opacity: 0;
}

@keyframes settings-title-enter {
  from {
    opacity: 0;
    transform: translateY(8px);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@keyframes settings-rule-draw {
  to {
    transform: scaleX(1);
  }
}

@keyframes settings-diamond-pop {
  from {
    opacity: 0;
    transform: rotate(-45deg) scale(0);
  }

  to {
    opacity: 1;
    transform: rotate(45deg) scale(1);
  }
}

@keyframes settings-skeleton-pulse {
  0%,
  100% {
    opacity: 0.6;
  }

  50% {
    opacity: 1;
  }
}

@keyframes settings-check-pop {
  from {
    opacity: 0;
    transform: scale(0.3) rotate(-30deg);
  }

  to {
    opacity: 1;
    transform: none;
  }
}

@keyframes settings-preview-flash {
  0% {
    background-color: color-mix(in srgb, var(--color-brand-accent-surface) 22%, transparent);
  }
}

/* --- Responsive ----------------------------------------------------------- */

@media (max-width: 1099px) {
  .settings-page__layout {
    grid-template-columns: minmax(0, 1fr);
    gap: 16px;
  }
}

@media (max-width: 767px) {
  .settings-grid,
  .vocab-group__pair {
    grid-template-columns: minmax(0, 1fr);
  }

  .settings-themes :deep(.option-group__option) {
    flex-basis: 100%;
    max-width: none;
  }
}

@media (max-width: 640px) {
  .settings-page {
    gap: 16px;
    padding: 16px 16px 28px;
  }

  .settings-page__save {
    bottom: 8px;
  }

  .size-step {
    width: 76px;
  }

  .accent-swatch {
    width: 76px;
  }
}

@media (prefers-reduced-motion: reduce) {
  .settings-page__title,
  .settings-page__subtitle,
  .settings-page__rule-line,
  .settings-page__rule-diamond,
  .settings-page__skeleton,
  .theme-tile__check,
  .settings-page__preview--changed {
    animation: none;
  }

  .settings-page__rule-line {
    transform: none;
  }
}
</style>
