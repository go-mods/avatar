<script setup lang="ts">
import { computed, ref, useId } from 'vue'
import { goCode, type AvatarOptions } from '../lib/options'

const props = defineProps<{
  options: AvatarOptions
  query: string
}>()

const groupName = useId()
const view = ref<'http' | 'go'>('http')
const copyLabel = ref('Copy')
const codeElement = ref<HTMLElement | null>(null)
let copiedTimer: ReturnType<typeof setTimeout> | undefined

const requestUrl = computed(() => `${window.location.origin}/api?${props.query}`)
const code = computed(() => (view.value === 'http' ? requestUrl.value : goCode(props.options)))

// The clipboard api only exists on https and localhost: a server reached by http on a local network
// goes through the selection instead.
function copyFromSelection(): boolean {
  const selection = window.getSelection()
  if (!selection || !codeElement.value) return false
  selection.selectAllChildren(codeElement.value)
  return document.execCommand('copy')
}

async function copy() {
  let done = false
  try {
    await navigator.clipboard.writeText(code.value)
    done = true
  } catch {
    done = copyFromSelection()
  }
  copyLabel.value = done ? 'Copied' : 'Selected: press Ctrl+C'
  clearTimeout(copiedTimer)
  copiedTimer = setTimeout(() => (copyLabel.value = 'Copy'), done ? 1500 : 4000)
}
</script>

<template>
  <section aria-label="How to get this avatar">
    <div class="mb-2 flex items-center justify-between gap-2">
      <div class="flex flex-wrap gap-x-4" role="group" aria-label="Show the code as">
        <label class="flex cursor-pointer items-center gap-1.5">
          <input v-model="view" type="radio" :name="groupName" value="http" /> HTTP request
        </label>
        <label class="flex cursor-pointer items-center gap-1.5">
          <input v-model="view" type="radio" :name="groupName" value="go" /> Go code
        </label>
      </div>
      <button class="quiet-button" type="button" @click="copy">{{ copyLabel }}</button>
    </div>
    <pre class="avatar-panel-code max-h-56 overflow-auto rounded-md p-3 font-mono text-sm break-all whitespace-pre-wrap" tabindex="0"><code ref="codeElement">{{ code }}</code></pre>
  </section>
</template>
