<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage } from 'element-plus'
import PaperSelect from './PaperSelect.vue'
import { useTopicDiscoveryFormOptions } from '@/composables/useTopicDiscoveryFormOptions'
import {
  ENV_PREFERENCE_STORAGE_KEY,
  LITERATURE_SOURCE_OPTIONS,
  type EnvironmentPreferenceForm,
} from './types'

const props = withDefaults(
  defineProps<{
    envPreference: EnvironmentPreferenceForm
    /** 嵌入个人中心 Tab：无外层卡片 */
    embedded?: boolean
  }>(),
  { embedded: false },
)

const emit = defineEmits<{
  environmentSaved: []
}>()

const envSaving = ref(false)

const {
  disciplineSelectOptions,
  intensityOptions,
  auditOptions,
} = useTopicDiscoveryFormOptions()

function isLiteratureSourceChecked(code: string) {
  return props.envPreference.literatureSourceCodes.includes(code)
}

function toggleLiteratureSource(code: string, checked: boolean) {
  const set = new Set(props.envPreference.literatureSourceCodes)
  if (checked) set.add(code)
  else set.delete(code)
  props.envPreference.literatureSourceCodes = [...set]
}

async function saveEnvironment() {
  if (props.envPreference.literatureSourceCodes.length === 0) {
    ElMessage.warning('请至少选择一个文献来源')
    return
  }
  envSaving.value = true
  try {
    await new Promise((r) => setTimeout(r, 400))
    localStorage.setItem(
      ENV_PREFERENCE_STORAGE_KEY,
      JSON.stringify({ preference: { ...props.envPreference } }),
    )
    emit('environmentSaved')
    ElMessage.success('环境配置已保存（本地演示）')
  } finally {
    envSaving.value = false
  }
}
</script>

<template>
  <component :is="embedded ? 'div' : 'section'" :class="embedded ? 'pc-embed' : 'pc-panel'">
    <div class="pc-pane pc-pane--env" role="region" aria-label="环境配置">
      <p class="pc-lead">
        新建工作流时的默认科研偏好与文献策略；保存后对后续「选题发现」等模块预填生效。
      </p>

      <div class="pc-field-grid">
        <label class="pc-field">
          <span class="pc-label">默认学科</span>
          <PaperSelect v-model="envPreference.disciplineCode" :options="disciplineSelectOptions" />
        </label>

        <label class="pc-field">
          <span class="pc-label">默认目标会议/期刊</span>
          <input v-model="envPreference.defaultVenueText" type="text" class="pc-input" />
          <span class="pc-hint">用于约束贡献类型、实验标准和写作风格。</span>
        </label>

        <label class="pc-field">
          <span class="pc-label">默认执行强度</span>
          <PaperSelect v-model="envPreference.intensity" :options="intensityOptions" />
          <span class="pc-hint">控制检索数量、迭代轮数和输出深度。</span>
        </label>

        <label class="pc-field">
          <span class="pc-label">默认审计等级</span>
          <PaperSelect v-model="envPreference.auditLevel" :options="auditOptions" />
          <span class="pc-hint">控制 citation audit、claim audit、kill argument 等门禁强度。</span>
        </label>
      </div>

      <div class="pc-field pc-field--block">
        <span class="pc-label">默认文献来源</span>
        <div class="pc-check-group">
          <label v-for="src in LITERATURE_SOURCE_OPTIONS" :key="src.code" class="pc-check pc-check--inline">
            <input
              type="checkbox"
              :checked="isLiteratureSourceChecked(src.code)"
              @change="toggleLiteratureSource(src.code, ($event.target as HTMLInputElement).checked)"
            />
            <span>{{ src.label }}</span>
          </label>
        </div>
        <span class="pc-hint">所有结果必须进入真实文献验证流程。</span>
      </div>

      <label class="pc-check">
        <input v-model="envPreference.humanCheckpoint" type="checkbox" />
        <span><strong>默认开启人工检查点</strong></span>
      </label>

      <label class="pc-check">
        <input v-model="envPreference.referenceGateEnabled" type="checkbox" />
        <span>
          <strong>参考文献门禁</strong>
          <span class="pc-hint">未验证文献不得进入引用与正文</span>
        </span>
      </label>

      <div class="pc-env-actions">
        <button type="button" class="pc-btn-primary" :disabled="envSaving" @click="saveEnvironment">
          {{ envSaving ? '保存中…' : '保存环境配置' }}
        </button>
      </div>
    </div>
  </component>
</template>

<style scoped>
.pc-embed {
  padding: 0;
}

.pc-panel {
  padding: 20px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.pc-lead {
  margin: 0 0 20px;
  font-size: 14px;
  line-height: 1.55;
  color: #475569;
}

.pc-field-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(240px, 1fr));
  gap: 16px 20px;
  margin-bottom: 20px;
}

.pc-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pc-field--block {
  margin-bottom: 16px;
}

.pc-label {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}

.pc-input {
  padding: 10px 12px;
  font-size: 14px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
}

.pc-check-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 20px;
}

.pc-check {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  margin-bottom: 12px;
  font-size: 14px;
  color: #334155;
  cursor: pointer;
}

.pc-check--inline {
  margin-bottom: 0;
}

.pc-hint {
  display: block;
  margin-top: 2px;
  font-size: 12px;
  font-weight: 400;
  color: #94a3b8;
}

.pc-env-actions {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
}

.pc-btn-primary {
  padding: 10px 22px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  background: linear-gradient(135deg, #6366f1, #4f46e5);
  border: none;
  border-radius: 10px;
  cursor: pointer;
}

.pc-btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
