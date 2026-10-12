<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, ref } from 'vue'

export type PaperMultiSelectOption = {
  value: string
  label: string
}

const props = withDefaults(
  defineProps<{
    modelValue: readonly string[]
    options: readonly PaperMultiSelectOption[]
    placeholder?: string
    disabled?: boolean
    /** 下拉内搜索框，按 label / value 过滤选项 */
    filterable?: boolean
    filterPlaceholder?: string
  }>(),
  {
    placeholder: '请选择',
    disabled: false,
    filterable: false,
    filterPlaceholder: '输入关键词筛选',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const open = ref(false)
const root = ref<HTMLElement | null>(null)
const filterInputRef = ref<HTMLInputElement | null>(null)
const filterQuery = ref('')
const ignoreNextToggle = ref(false)

const selectedSet = computed(() => new Set(props.modelValue))

const filteredOptions = computed(() => {
  if (!props.filterable) return props.options
  const q = filterQuery.value.trim().toLowerCase()
  if (!q) return props.options
  return props.options.filter(
    (o) =>
      o.label.toLowerCase().includes(q) || o.value.toLowerCase().includes(q),
  )
})

const currentLabel = computed(() => {
  if (!props.modelValue.length) {
    if (!props.options.length) return '加载中…'
    return props.placeholder
  }
  return props.modelValue
    .map((v) => props.options.find((o) => o.value === v)?.label ?? v)
    .filter(Boolean)
    .join('、')
})

const showPlaceholderStyle = computed(() => props.modelValue.length === 0)

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

function openPanel() {
  open.value = true
  resetFilterQuery()
  focusFilterInput()
}

function togglePanel() {
  if (props.disabled) return
  if (ignoreNextToggle.value) {
    ignoreNextToggle.value = false
    return
  }
  if (open.value) {
    closeMenu()
  } else {
    openPanel()
  }
}

function onFilterKeydown(e: KeyboardEvent) {
  if (e.key === 'Escape') {
    e.preventDefault()
    closeMenu()
  }
}

function toggleValue(value: string, checked: boolean) {
  const set = new Set(props.modelValue)
  if (checked) set.add(value)
  else set.delete(value)
  emit('update:modelValue', [...set])
}

function isChecked(value: string) {
  return selectedSet.value.has(value)
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
  <div ref="root" class="paper-select paper-multi" :class="{ 'paper-select--open': open }">
    <button
      type="button"
      class="paper-select-trigger"
      :disabled="disabled"
      @click.stop="togglePanel"
    >
      <span
        class="paper-select-value"
        :class="{ 'paper-select-value--muted': showPlaceholderStyle }"
      >
        {{ currentLabel }}
      </span>
      <span class="paper-select-chevron" aria-hidden="true">▾</span>
    </button>
    <div
      v-if="open"
      class="paper-multi-panel"
      :class="{ 'paper-multi-panel--filterable': filterable }"
      role="listbox"
      @click.stop
      @mousedown.stop
    >
      <div v-if="filterable" class="paper-select-search">
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
      <div v-if="options.length" class="paper-multi-body">
        <div v-if="filteredOptions.length" class="paper-multi-grid">
          <label
            v-for="o in filteredOptions"
            :key="o.value"
            class="paper-multi-row"
            :class="{ 'paper-multi-row--checked': isChecked(o.value) }"
          >
            <input
              type="checkbox"
              :checked="isChecked(o.value)"
              @change="toggleValue(o.value, ($event.target as HTMLInputElement).checked)"
            />
            <span class="paper-multi-label">{{ o.label }}</span>
          </label>
        </div>
        <p v-else class="paper-multi-empty">无匹配项</p>
      </div>
      <p v-else class="paper-multi-empty">暂无可选 venue</p>
    </div>
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

.paper-select-trigger:hover:not(:disabled) {
  border-color: #cbd5e1;
}

.paper-select--open .paper-select-trigger {
  border-color: #94a3b8;
  box-shadow: 0 0 0 3px rgba(148, 163, 184, 0.22);
}

.paper-select-value {
  flex: 1;
  min-width: 0;
  line-height: 1.4;
  white-space: normal;
  word-break: break-word;
}

.paper-select-value--muted {
  color: #94a3b8;
}

.paper-select-chevron {
  flex-shrink: 0;
  font-size: 12px;
  color: #64748b;
}

.paper-select-trigger:disabled {
  cursor: not-allowed;
  opacity: 0.65;
}

.paper-multi-panel {
  position: absolute;
  z-index: 80;
  top: calc(100% + 4px);
  right: 0;
  left: 0;
  max-height: min(400px, 58vh);
  padding: 8px;
  overflow-y: auto;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.08);
}

.paper-multi-panel--filterable {
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

.paper-multi-body {
  flex: 1;
  min-height: 0;
  padding: 8px;
  overflow-y: auto;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
}

.paper-multi-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 4px 8px;
}

.paper-multi-row {
  display: flex;
  gap: 6px;
  align-items: flex-start;
  min-width: 0;
  padding: 6px 8px;
  font-size: 13px;
  color: #1e1b4b;
  cursor: pointer;
  border-radius: 8px;
}

.paper-multi-row:hover {
  background: #f8fafc;
}

.paper-multi-row--checked {
  background: #f1f5f9;
}

.paper-multi-row input {
  flex-shrink: 0;
  margin-top: 3px;
}

.paper-multi-label {
  flex: 1;
  min-width: 0;
  line-height: 1.35;
  word-break: break-word;
}

.paper-multi-empty {
  margin: 0;
  padding: 8px 10px;
  font-size: 13px;
  color: #94a3b8;
}
</style>
