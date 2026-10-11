<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'

export type PaperSelectOption = {
  value: string
  label: string
}

const props = withDefaults(
  defineProps<{
    modelValue: string
    options: readonly PaperSelectOption[]
    placeholder?: string
    /** 下拉内搜索框，按 label / value 过滤选项 */
    filterable?: boolean
    filterPlaceholder?: string
  }>(),
  {
    placeholder: '请选择',
    filterable: false,
    filterPlaceholder: '输入关键词筛选',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const root = ref<HTMLElement | null>(null)
const filterInputRef = ref<HTMLInputElement | null>(null)
/** 仅下拉内搜索用，与触发区展示文案分离 */
const filterQuery = ref('')
/** 选项点击后 label 可能再触发一次 button，需忽略 */
const ignoreNextToggle = ref(false)

const matchedOption = computed(() =>
  props.options.find((o) => o.value === props.modelValue),
)

const currentLabel = computed(() => {
  if (matchedOption.value) return matchedOption.value.label
  if (!props.modelValue) return props.placeholder
  if (!props.options.length) return '加载中…'
  return props.placeholder
})

const showPlaceholderStyle = computed(() => !matchedOption.value)

const filteredOptions = computed(() => {
  if (!props.filterable) return props.options
  const q = filterQuery.value.trim().toLowerCase()
  if (!q) return props.options
  return props.options.filter(
    (o) =>
      o.label.toLowerCase().includes(q) || o.value.toLowerCase().includes(q),
  )
})

function resetFilterQuery() {
  filterQuery.value = ''
}

function closeMenu() {
  open.value = false
  resetFilterQuery()
}

function focusFilterInput() {
  if (!props.filterable) return
  void nextTick(() => {
    filterInputRef.value?.focus()
  })
}

function openMenu() {
  open.value = true
  resetFilterQuery()
  focusFilterInput()
}

function toggle() {
  if (ignoreNextToggle.value) {
    ignoreNextToggle.value = false
    return
  }
  if (open.value) {
    closeMenu()
  } else {
    openMenu()
  }
}

function onFilterKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    e.preventDefault()
    closeMenu()
  }
}

function pick(value: string) {
  if (value !== props.modelValue) {
    emit('update:modelValue', value)
  }
  open.value = false
  resetFilterQuery()
  ignoreNextToggle.value = true
  window.setTimeout(() => {
    ignoreNextToggle.value = false
  }, 0)
}

function onDocumentPointerDown(e: PointerEvent) {
  if (!root.value?.contains(e.target as Node)) {
    closeMenu()
  }
}

onMounted(() => document.addEventListener('pointerdown', onDocumentPointerDown))
onUnmounted(() => document.removeEventListener('pointerdown', onDocumentPointerDown))
</script>

<template>
  <div ref="root" class="paper-select" :class="{ 'paper-select--open': open }">
    <button type="button" class="paper-select-trigger" @click.stop="toggle">
      <span class="paper-select-value" :class="{ 'paper-select-value--muted': showPlaceholderStyle }">
        {{ currentLabel }}
      </span>
      <span class="paper-select-chevron" aria-hidden="true">▾</span>
    </button>

    <div
      v-if="open && filterable"
      class="paper-select-menu paper-select-menu--filterable"
      @click.stop
      @mousedown.stop
    >
      <div class="paper-select-search">
        <input
          ref="filterInputRef"
          v-model="filterQuery"
          type="search"
          class="paper-select-search-input"
          :placeholder="filterPlaceholder"
          autocomplete="off"
          spellcheck="false"
          @keydown="onFilterKeydown"
        />
      </div>
      <ul class="paper-select-options" role="listbox">
        <li
          v-if="!filteredOptions.length"
          class="paper-select-empty"
          role="presentation"
        >
          无匹配项
        </li>
        <li
          v-for="o in filteredOptions"
          :key="o.value"
          role="option"
          class="paper-select-option"
          :class="{ 'paper-select-option--active': o.value === modelValue }"
          :aria-selected="o.value === modelValue"
          @mousedown.prevent.stop="pick(o.value)"
        >
          {{ o.label }}
        </li>
      </ul>
    </div>

    <ul
      v-else-if="open"
      class="paper-select-menu"
      role="listbox"
      @click.stop
      @mousedown.stop
    >
      <li
        v-for="o in options"
        :key="o.value"
        role="option"
        class="paper-select-option"
        :class="{ 'paper-select-option--active': o.value === modelValue }"
        :aria-selected="o.value === modelValue"
        @mousedown.prevent.stop="pick(o.value)"
      >
        {{ o.label }}
      </li>
    </ul>
  </div>
</template>

<style scoped>
.paper-select {
  position: relative;
  width: 100%;
}

.paper-select-trigger {
  display: flex;
  gap: 8px;
  align-items: center;
  justify-content: space-between;
  width: 100%;
  padding: 10px 12px;
  font-size: 14px;
  color: #1e1b4b;
  text-align: left;
  cursor: pointer;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.paper-select-trigger:hover {
  border-color: #cbd5e1;
}

.paper-select--open .paper-select-trigger {
  border-color: #94a3b8;
  box-shadow: 0 0 0 3px rgba(148, 163, 184, 0.22);
}

.paper-select-value {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.paper-select-value--muted {
  color: #94a3b8;
}

.paper-select-chevron {
  flex-shrink: 0;
  font-size: 12px;
  color: #64748b;
}

.paper-select-menu {
  position: absolute;
  z-index: 80;
  top: calc(100% + 4px);
  right: 0;
  left: 0;
  max-height: min(400px, 58vh);
  margin: 0;
  padding: 6px;
  overflow-y: auto;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
  list-style: none;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.08);
}

.paper-select-menu--filterable {
  display: flex;
  flex-direction: column;
  height: min(400px, 58vh);
  max-height: min(400px, 58vh);
  padding: 0;
  overflow: hidden;
}

.paper-select-search {
  flex-shrink: 0;
  padding: 8px;
  border-bottom: 1px solid #e2e8f0;
}

.paper-select-search-input {
  box-sizing: border-box;
  width: 100%;
  padding: 8px 10px;
  font-size: 14px;
  color: #1e1b4b;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  outline: none;
}

.paper-select-search-input:focus {
  border-color: #94a3b8;
  box-shadow: 0 0 0 3px rgba(148, 163, 184, 0.18);
}

.paper-select-search-input::placeholder {
  color: #94a3b8;
}

.paper-select-options {
  flex: 1;
  min-height: 0;
  margin: 0;
  padding: 6px;
  overflow-y: auto;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
  list-style: none;
}

.paper-select-option {
  padding: 8px 10px;
  font-size: 14px;
  color: #1e1b4b;
  cursor: pointer;
  border-radius: 8px;
}

.paper-select-option:hover {
  background: #f8fafc;
}

.paper-select-option--active {
  font-weight: 600;
  background: #f1f5f9;
}

.paper-select-empty {
  padding: 8px 10px;
  font-size: 13px;
  color: #94a3b8;
  cursor: default;
}
</style>
