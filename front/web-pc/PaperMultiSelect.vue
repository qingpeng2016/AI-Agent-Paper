<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

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
  }>(),
  { placeholder: '请选择', disabled: false },
)

const emit = defineEmits<{
  'update:modelValue': [value: string[]]
}>()

const open = ref(false)
const root = ref<HTMLElement | null>(null)
const ignoreNextToggle = ref(false)

const selectedSet = computed(() => new Set(props.modelValue))

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

function closeMenu() {
  open.value = false
}

function togglePanel() {
  if (props.disabled) return
  if (ignoreNextToggle.value) {
    ignoreNextToggle.value = false
    return
  }
  open.value = !open.value
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
    <div v-if="open" class="paper-multi-panel" role="listbox" @click.stop @mousedown.stop>
      <div v-if="options.length" class="paper-multi-grid">
        <label
          v-for="o in options"
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
  max-height: 320px;
  padding: 8px;
  overflow-y: auto;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.08);
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
