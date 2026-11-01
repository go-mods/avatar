<script setup lang="ts">
import { computed, ref } from 'vue'
import { svgToDataUrl, usesRandomColors, type AvatarOptions } from '../lib/options'
import CodeView from './CodeView.vue'
import UserAvatar from './UserAvatar.vue'

const props = defineProps<{
  options: AvatarOptions
  query: string
  svg: string
  error: string
}>()

defineEmits<{
  loaded: [svg: string]
  failed: [message: string]
}>()

const refresh = ref(0)
const download = computed(() => (props.svg ? svgToDataUrl(props.svg) : undefined))
</script>

<template>
  <aside class="avatar-panel flex flex-col gap-4 p-5 md:sticky md:top-0 md:h-screen md:justify-center md:overflow-y-auto md:p-10 max-md:sticky max-md:top-0 max-md:z-10">
    <div class="flex items-center gap-4 md:flex-col md:gap-6">
      <UserAvatar
        class="h-24 w-24 shrink-0 object-contain md:h-auto md:max-h-[44vh] md:w-full md:max-w-80"
        :query="query"
        :alt="`Avatar of ${options.name || 'nobody'}`"
        :delay="180"
        :refresh="refresh"
        @loaded="$emit('loaded', $event)"
        @failed="$emit('failed', $event)"
      />
      <div class="flex flex-wrap gap-2 md:justify-center">
        <button v-if="usesRandomColors(options)" class="quiet-button" type="button" @click="refresh++">Draw new colors</button>
        <a v-if="download" class="quiet-button inline-flex items-center" :href="download" download="avatar.svg">Download SVG</a>
      </div>
    </div>

    <p v-if="error" class="avatar-panel-code rounded-md p-3 text-sm" role="alert">
      {{ error }}. The avatar shown is the last valid one.
    </p>

    <CodeView class="w-full max-md:hidden" :options="options" :query="query" />
  </aside>
</template>
