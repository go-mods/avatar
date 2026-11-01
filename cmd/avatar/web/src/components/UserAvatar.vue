<!--
  Asks the server for an avatar and displays it as an image from a data url. An application which
  receives the svg as text displays it the same way, so the page shows what it would show.
-->
<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { svgToDataUrl } from '../lib/options'

const props = withDefaults(
  defineProps<{
    query: string
    alt: string
    // Time in milliseconds to wait for the next change before asking the server
    delay?: number
    // A new value asks the server again for the same query
    refresh?: number
  }>(),
  { delay: 0, refresh: 0 },
)

const emit = defineEmits<{
  loaded: [svg: string]
  failed: [message: string]
}>()

const picture = ref<string | null>(null)
let timer: ReturnType<typeof setTimeout> | undefined
let running: AbortController | undefined

async function load(query: string, signal: AbortSignal) {
  try {
    const response = await fetch(`/api?${query}`, { signal })
    if (!response.ok) {
      const body = (await response.json()) as { error?: string }
      emit('failed', body.error ?? `The server answered ${response.status}.`)
      return
    }
    const svg = await response.text()
    if (signal.aborted) return
    picture.value = svgToDataUrl(svg)
    emit('loaded', svg)
  } catch {
    if (!signal.aborted) emit('failed', 'The server does not answer. Check that it is still running.')
  }
}

// Only the last change counts: a slower answer to an earlier request never replaces it
function schedule() {
  clearTimeout(timer)
  running?.abort()
  const controller = new AbortController()
  running = controller
  timer = setTimeout(() => load(props.query, controller.signal), props.delay)
}

watch(() => [props.query, props.refresh], schedule, { immediate: true })

onBeforeUnmount(() => {
  clearTimeout(timer)
  running?.abort()
})
</script>

<template>
  <img v-if="picture" :src="picture" :alt="alt" />
  <div v-else role="img" :aria-label="alt" />
</template>
