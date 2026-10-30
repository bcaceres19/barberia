<script setup lang="ts">
// Contenido de la sección «Código de acceso» de Configuración (DEC-139): por
// dónde llega el código de verificación cuando alguien necesita confirmar su
// identidad para entrar. Se aplica al instante (una sola preferencia, sin
// borrador): elegir una opción la guarda y lo anuncia en una región viva. Si el
// guardado falla, la selección vuelve a la que sigue vigente en el servidor.
//
// No se pide ni se muestra ningún correo o teléfono: el servidor deriva los
// destinos de cada cuenta (DEC-081). La recuperación de acceso no se configura
// aquí; cada persona elige su canal al recuperar (DEC-092).
import { onMounted, ref } from 'vue'
import { BaseAlert, BaseButton, DiamondLoader } from '@/shared/ui'
import { fetchLoginOtpChannel, saveLoginOtpChannel } from '../api/loginOtpApi'
import type { LoginOtpChannel } from '../model/loginOtpOutcome'
import OptionGroup from './OptionGroup.vue'

type Status = 'loading' | 'ready' | 'network-error' | 'unexpected-error'

const OPTIONS: readonly { value: LoginOtpChannel; label: string; hint: string }[] = [
  {
    value: 'email',
    label: 'Correo',
    hint: 'Enviamos el código al correo de la cuenta. Es la opción que traen los negocios nuevos.',
  },
  {
    value: 'whatsapp',
    label: 'WhatsApp',
    hint: 'Lo enviamos al teléfono verificado de la cuenta; quien no tenga uno verificado no recibe código.',
  },
  {
    value: 'both',
    label: 'Correo y WhatsApp',
    hint: 'El mismo código por los dos canales; basta con que llegue por uno.',
  },
]
const LOADING_PHRASES = ['Cargando tu preferencia'] as const

const status = ref<Status>('loading')
const channel = ref<LoginOtpChannel>('email')
const saving = ref(false)
const feedback = ref<'idle' | 'saved' | 'failed'>('idle')

async function load() {
  status.value = 'loading'
  const outcome = await fetchLoginOtpChannel()
  if (outcome.kind === 'success') {
    channel.value = outcome.channel
    status.value = 'ready'
  } else {
    status.value = outcome.kind
  }
}

async function choose(next: LoginOtpChannel) {
  if (saving.value || next === channel.value) return
  const previous = channel.value
  channel.value = next
  saving.value = true
  feedback.value = 'idle'
  const outcome = await saveLoginOtpChannel(next)
  saving.value = false
  if (outcome.kind === 'success') {
    channel.value = outcome.channel
    feedback.value = 'saved'
  } else {
    channel.value = previous
    feedback.value = 'failed'
  }
}

onMounted(load)
</script>

<template>
  <div class="login-otp" :aria-busy="status === 'loading' || saving">
    <div v-if="status === 'loading'" role="status">
      <DiamondLoader label="Cargando tu preferencia" layout="inline" :phrases="LOADING_PHRASES" />
    </div>

    <BaseAlert
      v-else-if="status !== 'ready'"
      :variant="status === 'network-error' ? 'warning' : 'danger'"
      :title="status === 'network-error' ? 'No pudimos conectar' : 'No pudimos cargar esta opción'"
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
      <OptionGroup
        :model-value="channel"
        :options="OPTIONS"
        label="Canal del código de acceso"
        :disabled="saving"
        class="login-otp__options"
        @update:model-value="choose"
      >
        <template #default="{ option, selected }">
          <span class="login-otp__option" :class="{ 'login-otp__option--selected': selected }">
            <span class="login-otp__name">{{ option.label }}</span>
            <span class="login-otp__hint">{{ option.hint }}</span>
          </span>
        </template>
      </OptionGroup>
      <p class="login-otp__note">
        Aplica a todas las cuentas de tu negocio. Al recuperar el acceso, cada persona elige su
        propio canal.
      </p>
      <p class="login-otp__feedback" role="status" aria-live="polite">
        <template v-if="saving">Guardando…</template>
        <template v-else-if="feedback === 'saved'">Guardado.</template>
        <template v-else-if="feedback === 'failed'">
          No pudimos guardar el cambio; se mantiene la opción anterior. Inténtalo de nuevo.
        </template>
      </p>
    </template>
  </div>
</template>

<style scoped>
.login-otp {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-width: 0;
}

.login-otp__options {
  --option-group-gap: 12px;

  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(min(100%, 220px), 1fr));
}

.login-otp__option {
  display: flex;
  flex-direction: column;
  gap: 6px;
  height: 100%;
  min-height: 44px;
  padding: 14px 16px;
  box-sizing: border-box;
  color: var(--color-on-strong);
  background: color-mix(in srgb, var(--color-on-strong) 6%, transparent);
  border: var(--border-width-normal) solid var(--color-field-strong-border);
  border-radius: 2px;
}

.login-otp__option--selected {
  border-color: var(--color-brand-accent-surface);
  box-shadow: inset 0 calc(var(--border-width-emphasis) * -1) 0 var(--color-brand-accent-surface);
}

.login-otp__name {
  font-size: var(--font-size-body);
  font-weight: 600;
}

.login-otp__hint,
.login-otp__note,
.login-otp__feedback {
  margin: 0;
  font-size: var(--font-size-body-sm);
  line-height: 20px;
  overflow-wrap: anywhere;
}

.login-otp__feedback {
  min-height: 20px;
}
</style>
