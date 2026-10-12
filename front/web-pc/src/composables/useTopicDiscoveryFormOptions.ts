import { readonly, ref } from 'vue'
import {
  fetchTopicDiscoveryFormOptions,
  DEFAULT_EXEC_INTENSITY_CODE,
  EXEC_INTENSITY_FALLBACK_OPTIONS,
  type LiteratureSourceOption,
  type PaperSelectOption,
  type VenueOption,
} from '@/api/topicDiscovery'
import { DISCIPLINE_OPTIONS, LITERATURE_SOURCE_OPTIONS } from '@paper/types'

const disciplineSelectOptions = ref<PaperSelectOption[]>([])
const intensityOptions = ref<PaperSelectOption[]>([])
const auditOptions = ref<PaperSelectOption[]>([])
const defaultIntensityCode = ref(DEFAULT_EXEC_INTENSITY_CODE)
const defaultAuditLevelCode = ref('standard')
const literatureSourceOptions = ref<LiteratureSourceOption[]>([])
const venueOptions = ref<VenueOption[]>([])
const venueSelectOptions = ref<PaperSelectOption[]>([])
const defaultLiteratureSourceCodes = ref<string[]>([])
const ready = ref(false)
const loading = ref(false)

let reloadSeq = 0

function applyFallbackOptions() {
  disciplineSelectOptions.value = DISCIPLINE_OPTIONS.map((d) => ({
    value: d.code,
    label: d.label,
  }))
  intensityOptions.value = [...EXEC_INTENSITY_FALLBACK_OPTIONS]
  auditOptions.value = [
    { value: 'polished', label: '基础' },
    { value: 'standard', label: '标准' },
    { value: 'strict', label: '严格' },
  ]
  defaultIntensityCode.value = DEFAULT_EXEC_INTENSITY_CODE
  defaultAuditLevelCode.value = 'standard'
  // form-options 不可用时的本地 fallback（默认勾选对齐 all.sql 种子，非线上库表）
  const fallbackDefaultSelected = new Set(['arxiv', 'openalex', 'semantic_scholar'])
  literatureSourceOptions.value = LITERATURE_SOURCE_OPTIONS.map((s, i) => ({
    code: s.code,
    label: s.label,
    sort: i + 1,
    defaultSelected: fallbackDefaultSelected.has(s.code),
  }))
  defaultLiteratureSourceCodes.value = literatureSourceOptions.value
    .filter((s) => s.defaultSelected)
    .map((s) => s.code)
}

export function literatureSourceLabel(code: string): string {
  const hit = literatureSourceOptions.value.find((o) => o.code === code)
  if (hit) return hit.label
  return LITERATURE_SOURCE_OPTIONS.find((o) => o.code === code)?.label ?? code
}

export function applyDefaultLiteratureSources(target: { sourceCodes: string[] }) {
  const codes = defaultLiteratureSourceCodes.value
  if (codes.length > 0) {
    target.sourceCodes = [...codes]
  }
}

export function applyDefaultLiteratureSourceCodes(target: { literatureSourceCodes: string[] }) {
  const codes = defaultLiteratureSourceCodes.value
  if (codes.length > 0) {
    target.literatureSourceCodes = [...codes]
  }
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
    intensityOptions.value = opts.intensityOptions.length
      ? opts.intensityOptions
      : [...EXEC_INTENSITY_FALLBACK_OPTIONS]
    auditOptions.value = opts.auditOptions
    defaultIntensityCode.value = opts.intensityOptions.length
      ? opts.defaultIntensityCode || DEFAULT_EXEC_INTENSITY_CODE
      : DEFAULT_EXEC_INTENSITY_CODE
    if (opts.defaultAuditLevelCode) {
      defaultAuditLevelCode.value = opts.defaultAuditLevelCode
    }
    if (opts.venueOptions.length) {
      venueOptions.value = opts.venueOptions
      venueSelectOptions.value = opts.venueSelectOptions
    }
    if (opts.literatureSources.length) {
      literatureSourceOptions.value = opts.literatureSources
      defaultLiteratureSourceCodes.value = opts.defaultLiteratureSourceCodes
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
    literatureSourceOptions: readonly(literatureSourceOptions),
    venueOptions: readonly(venueOptions),
    venueSelectOptions: readonly(venueSelectOptions),
    defaultLiteratureSourceCodes: readonly(defaultLiteratureSourceCodes),
    applyCatalogIntensityAuditDefaults,
    literatureSourceLabel,
    applyDefaultLiteratureSources,
    applyDefaultLiteratureSourceCodes,
  }
}
