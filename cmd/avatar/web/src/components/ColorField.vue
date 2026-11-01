<script setup lang="ts">
import { computed } from 'vue'

defineProps<{
  id: string
  label: string
}>()

const model = defineModel<string>({ required: true })

// The color picker only reads a full hex code
const pickerValue = computed(() => (/^#[0-9a-fA-F]{6}$/.test(model.value) ? model.value : '#808080'))
</script>

<template>
  <div>
    <label class="field-label" :for="id">{{ label }}</label>
    <div class="flex gap-2">
      <input
        :aria-label="`${label}: pick a color`"
        class="h-10 w-11 shrink-0 cursor-pointer rounded-md border border-line bg-paper p-1"
        type="color"
        :value="pickerValue"
        @input="model = ($event.target as HTMLInputElement).value"
      />
      <input :id="id" v-model.trim="model" class="field-input" type="text" placeholder="From the name" spellcheck="false" />
    </div>
  </div>
</template>
