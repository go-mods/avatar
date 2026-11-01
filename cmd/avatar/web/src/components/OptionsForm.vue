<script setup lang="ts">
import { fontFamilies } from '../lib/examples'
import { shapes, type AvatarOptions } from '../lib/options'
import ColorField from './ColorField.vue'

const options = defineModel<AvatarOptions>({ required: true })

defineEmits<{ reset: [] }>()

const initialsSwitches = [
  { key: 'uppercase', label: 'Upper case' },
  { key: 'lowercase', label: 'Lower case' },
  { key: 'camelcase', label: 'Camel case' },
  { key: 'sensitive', label: 'Keep the case of the name' },
  { key: 'wordLength', label: 'One letter per word' },
] as const

const randomSwitches = [
  { key: 'randomColor', label: 'All colors, matching' },
  { key: 'randomBackgroundColor', label: 'Background' },
  { key: 'randomFontColor', label: 'Text' },
  { key: 'randomBorderColor', label: 'Border' },
] as const

// The regular weight is the default one: it is not sent
const weights = [100, 200, 300, 500, 600, 700, 800, 900]
</script>

<template>
  <form class="space-y-9" @submit.prevent>
    <fieldset>
      <legend class="mb-3 text-lg font-semibold">Shape and size</legend>
      <div class="mb-4 flex flex-wrap gap-x-5 gap-y-2">
        <label v-for="shape in shapes" :key="shape" class="flex cursor-pointer items-center gap-2 capitalize">
          <input v-model="options.shape" type="radio" name="shape" :value="shape" /> {{ shape }}
        </label>
      </div>
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3">
        <div>
          <label class="field-label" for="width">Width</label>
          <input id="width" v-model.number="options.width" class="field-input" type="number" min="1" max="4096" placeholder="48" />
        </div>
        <div>
          <label class="field-label" for="height">Height</label>
          <input id="height" v-model.number="options.height" class="field-input" type="number" min="1" max="4096" placeholder="48" />
        </div>
        <div>
          <label class="field-label" for="padding">Space around</label>
          <input id="padding" v-model.number="options.padding" class="field-input" type="number" min="0" max="4096" placeholder="0" />
        </div>
      </div>
    </fieldset>

    <fieldset>
      <legend class="mb-3 text-lg font-semibold">Colors</legend>
      <div class="grid gap-4 sm:grid-cols-3">
        <ColorField id="background-color" v-model="options.backgroundColor" label="Background" />
        <ColorField id="font-color" v-model="options.fontColor" label="Text" />
        <ColorField id="border-color" v-model="options.borderColor" label="Border" />
      </div>
      <p class="mt-2 text-sm text-muted">A color name or a hex code, like navy or #1e3a8a.</p>
      <p class="mt-4 mb-2 text-sm text-muted">Draw at random, unless a color is set above:</p>
      <div class="flex flex-wrap gap-x-5 gap-y-2">
        <label v-for="option in randomSwitches" :key="option.key" class="flex cursor-pointer items-center gap-2">
          <input v-model="options[option.key]" type="checkbox" /> {{ option.label }}
        </label>
      </div>
    </fieldset>

    <fieldset>
      <legend class="mb-3 text-lg font-semibold">Border</legend>
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3">
        <div>
          <label class="field-label" for="border-width">Width</label>
          <input id="border-width" v-model.number="options.borderWidth" class="field-input" type="number" min="0" max="4096" placeholder="0" />
        </div>
        <div>
          <label class="field-label" for="border-dash">Dashes</label>
          <input id="border-dash" v-model.trim="options.borderDash" class="field-input" type="text" placeholder="None, or 4 2" spellcheck="false" />
        </div>
        <div v-if="options.shape === 'square'">
          <label class="field-label" for="border-radius">Corner radius</label>
          <input id="border-radius" v-model.number="options.borderRadius" class="field-input" type="number" min="0" max="4096" placeholder="0" />
        </div>
      </div>
    </fieldset>

    <fieldset>
      <legend class="mb-3 text-lg font-semibold">Font</legend>
      <div class="grid grid-cols-2 gap-4 sm:grid-cols-3">
        <div class="col-span-2 sm:col-span-1">
          <label class="field-label" for="font-family">Google Fonts family</label>
          <input id="font-family" v-model.trim="options.fontFamily" class="field-input" type="text" list="font-families" placeholder="System sans-serif" spellcheck="false" />
          <datalist id="font-families">
            <option v-for="family in fontFamilies" :key="family" :value="family" />
          </datalist>
        </div>
        <div>
          <label class="field-label" for="font-weight">Weight</label>
          <select id="font-weight" v-model.number="options.fontWeight" class="field-input">
            <option v-for="weight in weights.filter((value) => value < 400)" :key="weight" :value="weight">{{ weight }}</option>
            <option value="">400</option>
            <option v-for="weight in weights.filter((value) => value > 400)" :key="weight" :value="weight">{{ weight }}</option>
          </select>
        </div>
        <div>
          <label class="field-label" for="font-size">Size</label>
          <input id="font-size" v-model.number="options.fontSize" class="field-input" type="number" min="1" max="4096" placeholder="Automatic" />
        </div>
      </div>
      <p class="mt-2 text-sm text-muted">The font travels inside the SVG, reduced to the letters it needs.</p>
    </fieldset>

    <fieldset>
      <legend class="mb-3 text-lg font-semibold">Initials</legend>
      <div class="mb-4 grid grid-cols-2 gap-4 sm:grid-cols-3">
        <div>
          <label class="field-label" for="length">Letters</label>
          <input id="length" v-model.number="options.length" class="field-input" type="number" min="1" max="64" placeholder="2" />
        </div>
        <div>
          <label class="field-label" for="separator">Separator</label>
          <input id="separator" v-model="options.separator" class="field-input" type="text" placeholder="None, or ." spellcheck="false" />
        </div>
      </div>
      <div class="flex flex-wrap gap-x-5 gap-y-2">
        <label v-for="option in initialsSwitches" :key="option.key" class="flex cursor-pointer items-center gap-2">
          <input v-model="options[option.key]" type="checkbox" /> {{ option.label }}
        </label>
      </div>
    </fieldset>

    <button class="quiet-button" type="button" @click="$emit('reset')">Reset all options</button>
  </form>
</template>
