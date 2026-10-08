import { readonly, ref } from 'vue'
import {
  fetchTopicDiscoveryFormOptions,
  type PaperSelectOption,
} from '@/api/topicDiscovery'
import { DISCIPLINE_OPTIONS } from '@paper/types'

const disciplineSelectOptions = ref<PaperSelectOption[]>([])
const intensityOptions = ref<PaperSelectOption[]>([])
const auditOptions = ref<PaperSelectOption[]>([])
const defaultIntensityCode = ref('balanced')
const defaultAuditLevelCode = ref('polished')
const ready = ref(false)
const loading = ref(false)

let reloadSeq = 0

function applyFallbackOptions() {
  disciplineSelectOptions.value = DISCIPLINE_OPTIONS.map((d) => ({
    value: d.code,
    label: d.label,
  }))
  intensityOptions.value = [
    { value: 'fast', label: '更快' },
    { value: 'balanced', label: 'Balanced（平衡）' },
    { value: 'deep', label: '更深' },
  ]
  auditOptions.value = [
    { value: 'standard', label: '标准' },
    { value: 'polished', label: '精修' },
    { value: 'strict', label: '严格' },
  ]
  defaultIntensityCode.value = 'balanced'
  defaultAuditLevelCode.value = 'polished'
}

export function applyCatalogIntensityAuditDefaults(target: {
  intensity: string
  auditLevel: string
}) {
  const i = defaultIntensityCode.value
  const a = defaultAuditLevelCode.value
  if (i && intensityOptions.value.some((o) => o.value === i)) {
    target.intensity = i
  }
  if (a && auditOptions.value.some((o) => o.value === a)) {
    target.auditLevel = a
  }
}

/** 每次进入相关模块时重新拉取；失败时用本地 fallback */
export async function reloadTopicDiscoveryFormOptions(): Promise<void> {
  const seq = ++reloadSeq
  loading.value = true
  try {
    const opts = await fetchTopicDiscoveryFormOptions()
    if (seq !== reloadSeq) return
    disciplineSelectOptions.value = opts.disciplineSelectOptions
    intensityOptions.value = opts.intensityOptions
    auditOptions.value = opts.auditOptions
    if (opts.defaultIntensityCode) {
      defaultIntensityCode.value = opts.defaultIntensityCode
    }
    if (opts.defaultAuditLevelCode) {
      defaultAuditLevelCode.value = opts.defaultAuditLevelCode
    }
    if (
      !disciplineSelectOptions.value.length ||
      !intensityOptions.value.length ||
      !auditOptions.value.length
    ) {
      applyFallbackOptions()
    }
  } catch {
    if (seq !== reloadSeq) return
    applyFallbackOptions()
  } finally {
    if (seq === reloadSeq) {
      ready.value = true
      loading.value = false
    }
  }
}

/** @deprecated 使用 reloadTopicDiscoveryFormOptions */
export function ensureTopicDiscoveryFormOptions(): Promise<void> {
  return reloadTopicDiscoveryFormOptions()
}

export function useTopicDiscoveryFormOptions() {
  return {
    ready: readonly(ready),
    loading: readonly(loading),
    disciplineSelectOptions: readonly(disciplineSelectOptions),
    intensityOptions: readonly(intensityOptions),
    auditOptions: readonly(auditOptions),
    defaultIntensityCode: readonly(defaultIntensityCode),
    defaultAuditLevelCode: readonly(defaultAuditLevelCode),
    applyCatalogIntensityAuditDefaults,
  }
}
