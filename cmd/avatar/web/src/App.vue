<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import CodeView from './components/CodeView.vue'
import OptionsForm from './components/OptionsForm.vue'
import ResultPanel from './components/ResultPanel.vue'
import UserAvatar from './components/UserAvatar.vue'
import { examples, type Example } from './lib/examples'
import { buildQuery, defaultOptions, panelColor, parseQuery, readableInkOn, readBackgroundColor } from './lib/options'

const options = ref(parseQuery(window.location.search))
const svg = ref('')
const error = ref('')
const background = ref('#c4b3a2')

const query = computed(() => buildQuery(options.value))

// The address of the page describes the avatar, so that it can be shared
watch(query, (value) => {
  window.history.replaceState(null, '', value ? `?${value}` : window.location.pathname)
})

function onLoaded(content: string) {
  svg.value = content
  error.value = ''
  background.value = readBackgroundColor(content) ?? background.value
}

const panelColors = computed(() => {
  const panel = panelColor(background.value)
  return { '--avatar-background': panel, '--avatar-ink': readableInkOn(panel) }
})

function show(example: Example) {
  options.value = { ...defaultOptions(), ...example.options }
  window.scrollTo({ top: 0 })
}
</script>

<template>
  <div class="grid md:grid-cols-[minmax(0,1fr)_minmax(0,1fr)] xl:grid-cols-[minmax(0,3fr)_minmax(0,2fr)]">
    <ResultPanel
      class="md:order-2"
      :style="panelColors"
      :options="options"
      :query="query"
      :svg="svg"
      :error="error"
      @loaded="onLoaded"
      @failed="error = $event"
    />

    <main class="min-w-0 px-5 py-8 md:order-1 md:px-10 md:py-12 xl:px-16">
      <div class="mx-auto max-w-2xl">
        <header class="mb-10">
          <h1 class="font-display text-2xl font-bold">avatar</h1>
          <p class="mt-1 text-muted">Turn a name into an SVG avatar, from Go code or over HTTP.</p>
        </header>

        <div class="mb-12">
          <label class="field-label" for="name">Name</label>
          <input
            id="name"
            v-model="options.name"
            class="w-full border-b-2 border-line bg-transparent pb-1 font-display text-4xl font-bold focus:border-current focus:outline-none sm:text-5xl"
            type="text"
            maxlength="256"
            autocomplete="off"
            spellcheck="false"
            placeholder="Type a name"
          />
        </div>

        <OptionsForm v-model="options" @reset="options = defaultOptions()" />

        <CodeView class="mt-12 md:hidden" :options="options" :query="query" />

        <section class="mt-16" aria-labelledby="examples-title">
          <h2 id="examples-title" class="text-lg font-semibold">Start from an example</h2>
          <ul class="mt-4 grid grid-cols-2 gap-x-4 gap-y-6 sm:grid-cols-4">
            <li v-for="example in examples" :key="example.caption">
              <button class="group w-full cursor-pointer rounded-md text-left" type="button" @click="show(example)">
                <UserAvatar
                  class="mb-2 aspect-square w-full max-w-28 transition-transform group-hover:scale-105 motion-reduce:transition-none motion-reduce:group-hover:scale-100"
                  :query="buildQuery({ ...defaultOptions(), ...example.options })"
                  :alt="`Avatar of ${example.options.name}`"
                />
                <span class="block text-sm">{{ example.caption }}</span>
              </button>
            </li>
          </ul>
        </section>

        <section class="mt-16" aria-labelledby="use-title">
          <h2 id="use-title" class="text-lg font-semibold">Use it in your project</h2>
          <p class="mt-3 text-muted">Add the library to a Go module:</p>
          <pre class="mt-2 overflow-x-auto rounded-md border border-line p-3 font-mono text-sm" tabindex="0"><code>go get github.com/go-mods/avatar</code></pre>
          <p class="mt-4 text-muted">Or run this server from a clone of the repository:</p>
          <pre class="mt-2 overflow-x-auto rounded-md border border-line p-3 font-mono text-sm" tabindex="0"><code>docker compose up -d</code></pre>
          <p class="mt-4">
            <a class="underline underline-offset-2" href="https://github.com/go-mods/avatar">Read the code on GitHub</a>
          </p>
        </section>
      </div>
    </main>
  </div>
</template>
