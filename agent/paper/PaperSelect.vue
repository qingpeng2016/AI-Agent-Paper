<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'

export type PaperSelectOption = {
  value: string
  label: string
}

const props = defineProps<{
  modelValue: string
  options: readonly PaperSelectOption[]
}>()

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const open = ref(false)
const root = ref<HTMLElement | null>(null)
/** 选项点击后 label 可能再触发一次 button，需忽略 */
const ignoreNextToggle = ref(false)

const currentLabel = computed(
  () => props.options.find((o) => o.value === props.modelValue)?.label ?? props.modelValue,
)

function closeMenu() {
  open.value = false
}

function toggle() {
  if (ignoreNextToggle.value) {
    ignoreNextToggle.value = false
    return
  }
  open.value = !open.value
}

function pick(value: string) {
  if (value !== props.modelValue) {
    emit('update:modelValue', value)
  }
  closeMenu()
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
      <span class="paper-select-value">{{ currentLabel }}</span>
      <span class="paper-select-chevron" aria-hidden="true">▾</span>
    </button>
    <ul v-if="open" class="paper-select-menu" role="listbox" @click.stop @mousedown.stop>
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
  margin: 0;
  padding: 6px;
  list-style: none;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.08);
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
</style>
