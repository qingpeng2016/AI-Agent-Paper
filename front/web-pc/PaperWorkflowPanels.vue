<script setup lang="ts">
import { onUnmounted, reactive, ref, toRefs, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  fetchLiteratureReviews,
  formatLiteratureReviewStatus,
  generateExperimentPlanFromLiteratureReview,
  isLiteratureReviewGeneratingExperimentPlan,
  softDeleteLiteratureReview,
  type PaperLiteratureReviewItem,
} from '@/api/literatureReviews'
import {
  DEMO_AUTO_REVIEW,
  DEMO_EXPERIMENT_PLAN,
  DEMO_FIGURES,
  DEMO_MANUSCRIPT,
  DEMO_MANUSCRIPT_ANALYSIS,
} from './demoModuleOutputs'
import PaperFigureUploadPanel from './PaperFigureUploadPanel.vue'
import PaperSelect from './PaperSelect.vue'
import {
  DEFAULT_AUTO_REVIEW,
  DEFAULT_EXPERIMENT_PLANNING,
  DEFAULT_FIGURE_GENERATION,
  DEFAULT_MANUSCRIPT_ANALYSIS,
  DEFAULT_PAPER_WRITING,
  FIGURE_CHART_OPTIONS,
  PAPER_WRITING_SECTION_OPTIONS,
  type AutoReviewForm,
  type ExperimentPlanningForm,
  type FigureGenerationForm,
  type ManuscriptAnalysisForm,
  type PaperModuleId,
  type PaperWritingForm,
} from './types'

export type FigureManagementTabId = 'upload' | 'generate'

const figureTab = defineModel<FigureManagementTabId>('figureTab', { default: 'upload' })

const props = withDefaults(
  defineProps<{
    moduleId: PaperModuleId
    manuscriptId?: string
    manuscriptTitle?: string
  }>(),
  {
    manuscriptId: '',
    manuscriptTitle: '未命名',
  },
)
const { moduleId, manuscriptId, manuscriptTitle } = toRefs(props)

const planForm = reactive<ExperimentPlanningForm>({ ...DEFAULT_EXPERIMENT_PLANNING })
const reviewForm = reactive<AutoReviewForm>({ ...DEFAULT_AUTO_REVIEW })
const writeForm = reactive<PaperWritingForm>({
  ...DEFAULT_PAPER_WRITING,
  sections: [...DEFAULT_PAPER_WRITING.sections],
})
const figureForm = reactive<FigureGenerationForm>({
  ...DEFAULT_FIGURE_GENERATION,
  chartTypes: [...DEFAULT_FIGURE_GENERATION.chartTypes],
})
const analysisForm = reactive<ManuscriptAnalysisForm>({ ...DEFAULT_MANUSCRIPT_ANALYSIS })

const resultVisible = ref<Partial<Record<PaperModuleId, boolean>>>({})

const litReviewItems = ref<PaperLiteratureReviewItem[]>([])
const litReviewsLoading = ref(false)
const litReviewViewItem = ref<PaperLiteratureReviewItem | null>(null)
const litReviewDeletePending = ref<PaperLiteratureReviewItem | null>(null)
const litReviewDeleteSubmitting = ref(false)
const litReviewGeneratePending = ref<PaperLiteratureReviewItem | null>(null)
const litReviewGenerateSubmitting = ref(false)
let litReviewPollTimer: ReturnType<typeof setInterval> | null = null

const LIT_STRUCTURE_LABEL: Record<string, string> = {
  thematic: '按主题',
  chronological: '按时间线',
  method: '按方法族',
}

let litReviewReloadSeq = 0

async function reloadLiteratureReviews(opts?: { silent?: boolean }) {
  const silent = opts?.silent === true
  const ms = manuscriptId.value
  if (!ms || !/^\d+$/.test(ms)) {
    if (!silent) litReviewItems.value = []
    return
  }
  const seq = ++litReviewReloadSeq
  if (!silent) litReviewsLoading.value = true
  try {
    const data = await fetchLiteratureReviews(ms)
    if (seq !== litReviewReloadSeq) return
    litReviewItems.value = data.items ?? []
  } catch (e) {
    if (seq !== litReviewReloadSeq) return
    if (!silent) {
      litReviewItems.value = []
      const msg = e instanceof Error ? e.message : '加载文献综述失败'
      ElMessage.error(msg)
    }
  } finally {
    if (seq === litReviewReloadSeq) {
      if (!silent) litReviewsLoading.value = false
      syncLitReviewPollTimer()
    }
  }
}

/** 进入文献综述 / 切换当前论文 / 刷新：同一条路径拉列表（不依赖父组件 ref 时序） */
watch(
  () => [moduleId.value, manuscriptId.value] as const,
  ([mod, ms]) => {
    if (mod !== 'literature-review') return
    if (!ms || !/^\d+$/.test(ms)) {
      litReviewItems.value = []
      return
    }
    void reloadLiteratureReviews()
  },
  { immediate: true },
)

function litReviewHasGeneratingExperimentPlan(): boolean {
  return litReviewItems.value.some((it) => isLiteratureReviewGeneratingExperimentPlan(it.status))
}

function syncLitReviewPollTimer() {
  if (moduleId.value !== 'literature-review') {
    if (litReviewPollTimer) {
      clearInterval(litReviewPollTimer)
      litReviewPollTimer = null
    }
    return
  }
  if (litReviewHasGeneratingExperimentPlan()) {
    if (!litReviewPollTimer) {
      litReviewPollTimer = setInterval(() => {
        void reloadLiteratureReviews({ silent: true })
      }, 4000)
    }
  } else if (litReviewPollTimer) {
    clearInterval(litReviewPollTimer)
    litReviewPollTimer = null
  }
}

watch(litReviewItems, () => syncLitReviewPollTimer(), { deep: true })
watch(moduleId, () => syncLitReviewPollTimer())

onUnmounted(() => {
  if (litReviewPollTimer) {
    clearInterval(litReviewPollTimer)
    litReviewPollTimer = null
  }
})

function isGenerateExperimentPlanDisabled(item: PaperLiteratureReviewItem): boolean {
  return isLiteratureReviewGeneratingExperimentPlan(item.status)
}

function openGenerateExperimentPlanConfirm(item: PaperLiteratureReviewItem) {
  if (isGenerateExperimentPlanDisabled(item)) return
  litReviewGeneratePending.value = item
}

function closeGenerateExperimentPlanConfirm() {
  if (litReviewGenerateSubmitting.value) return
  litReviewGeneratePending.value = null
}

async function submitGenerateExperimentPlanConfirm() {
  const item = litReviewGeneratePending.value
  const ms = manuscriptId.value
  if (!item || !ms) return
  litReviewGenerateSubmitting.value = true
  try {
    await generateExperimentPlanFromLiteratureReview(ms, item.id)
    ElMessage.success('操作成功')
    litReviewGeneratePending.value = null
    await reloadLiteratureReviews({ silent: true })
  } catch (e) {
    const msg = e instanceof Error ? e.message : '提交失败'
    ElMessage.error(msg)
  } finally {
    litReviewGenerateSubmitting.value = false
  }
}

const intensityOptions = [
  { value: 'fast', label: '更快' },
  { value: 'balanced', label: 'Balanced（平衡）' },
  { value: 'deep', label: '更深' },
]

const auditOptions = [
  { value: 'standard', label: 'Standard' },
  { value: 'polished', label: 'Polished（精修）' },
  { value: 'strict', label: 'Strict' },
]

const writeToneOptions = [
  { value: 'concise', label: '简洁' },
  { value: 'standard', label: '标准 conference' },
]

const figureDatasetOptions = [
  { value: 'results_main.csv', label: 'results_main.csv（演示）' },
  { value: 'ablation_B.csv', label: 'ablation_B.csv（演示）' },
  { value: 'throughput.csv', label: 'throughput.csv（演示）' },
]

const captionLangOptions = [
  { value: 'en', label: 'English' },
  { value: 'zh', label: '中文' },
]

function toggleWritingSection(value: string, checked: boolean) {
  const set = new Set(writeForm.sections)
  if (checked) set.add(value)
  else set.delete(value)
  writeForm.sections = [...set]
}

function isWritingSectionChecked(value: string) {
  return writeForm.sections.includes(value)
}

function toggleChartType(value: string, checked: boolean) {
  const set = new Set(figureForm.chartTypes)
  if (checked) set.add(value)
  else set.delete(value)
  figureForm.chartTypes = [...set]
}

function isChartChecked(value: string) {
  return figureForm.chartTypes.includes(value)
}

function litReviewRowTitle(item: PaperLiteratureReviewItem): string {
  const t = item.title?.trim()
  if (t) return t
  return `文献综述 v${item.version}`
}

function formatLitReviewCreatedAt(iso: string): string {
  try {
    return new Date(iso).toLocaleString('zh-CN')
  } catch {
    return iso
  }
}

function openLitReviewView(item: PaperLiteratureReviewItem) {
  litReviewViewItem.value = item
}

function closeLitReviewView() {
  litReviewViewItem.value = null
}

function openDeleteLitReviewConfirm(item: PaperLiteratureReviewItem) {
  litReviewDeletePending.value = item
}

function closeDeleteLitReviewConfirm() {
  if (litReviewDeleteSubmitting.value) return
  litReviewDeletePending.value = null
}

async function submitDeleteLitReviewConfirm() {
  const item = litReviewDeletePending.value
  const ms = manuscriptId.value
  if (!item || !ms) return
  litReviewDeleteSubmitting.value = true
  try {
    await softDeleteLiteratureReview(ms, item.id)
    ElMessage.success('已删除')
    litReviewDeletePending.value = null
    if (litReviewViewItem.value?.id === item.id) closeLitReviewView()
    await reloadLiteratureReviews()
  } catch (e) {
    const msg = e instanceof Error ? e.message : '删除失败'
    ElMessage.error(msg)
  } finally {
    litReviewDeleteSubmitting.value = false
  }
}

async function runModule(id: PaperModuleId): Promise<boolean> {
  if (id === 'literature-review') {
    await reloadLiteratureReviews()
    if (!litReviewItems.value.length) {
      ElMessage.warning('暂无文献综述，请先在「选题发现」完成审计步骤')
      return false
    }
    return true
  }
  if (id === 'experiment-planning' && !planForm.ideaSummary.trim()) {
    ElMessage.warning('请填写核心 idea / 假设')
    return false
  }
  if (id === 'paper-writing' && writeForm.sections.length === 0) {
    ElMessage.warning('请至少选择一个章节')
    return false
  }
  if (id === 'figure-generation' && figureForm.chartTypes.length === 0) {
    ElMessage.warning('请至少选择一种图表类型')
    return false
  }

  await new Promise((r) => setTimeout(r, 750))
  resultVisible.value = { ...resultVisible.value, [id]: true }
  return true
}

defineExpose({ runModule, reloadLiteratureReviews })
</script>

<template>
  <div class="wf-root">
  <!-- 文献综述：paper_output_literature_review -->
  <div v-if="moduleId === 'literature-review'" class="wf-stack">
    <section v-if="litReviewsLoading" class="wf-panel">
      <p class="wf-meta">正在加载文献综述…</p>
    </section>

    <section v-else-if="!litReviewItems.length" class="wf-panel">
      <h2 class="wf-title">文献综述</h2>
      <p class="wf-lead">
        当前论文尚无文献综述。请先在「选题发现」完成审计步骤（第三步会自动生成并入库）。
      </p>
    </section>

    <section v-else class="wf-panel wf-lit-list-panel">
      <table class="wf-table wf-lit-table">
        <thead>
          <tr>
            <th>标题</th>
            <th>版本</th>
            <th>创建时间</th>
            <th>状态</th>
            <th class="wf-lit-col-actions">
              <div class="wf-lit-actions">
                <span class="wf-lit-head-slot">
                  <button
                    type="button"
                    tabindex="-1"
                    aria-hidden="true"
                    class="paper-btn-primary paper-btn--compact wf-lit-width-ruler"
                  >
                    查看
                  </button>
                  <span class="wf-lit-col-head-label">操作</span>
                </span>
                <button
                  type="button"
                  tabindex="-1"
                  aria-hidden="true"
                  class="paper-btn-primary paper-btn--compact wf-lit-width-ruler"
                >
                  生成实验方案
                </button>
                <button
                  type="button"
                  tabindex="-1"
                  aria-hidden="true"
                  class="paper-btn-danger paper-btn--compact wf-lit-width-ruler"
                >
                  删除
                </button>
              </div>
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="item in litReviewItems" :key="item.id">
            <td class="wf-lit-col-title">{{ litReviewRowTitle(item) }}</td>
            <td>v{{ item.version }}</td>
            <td>{{ formatLitReviewCreatedAt(item.created_at) }}</td>
            <td>{{ formatLiteratureReviewStatus(item.status) }}</td>
            <td class="wf-lit-col-actions">
              <div class="wf-lit-actions">
                <button
                  type="button"
                  class="paper-btn-primary paper-btn--compact"
                  @click="openLitReviewView(item)"
                >
                  查看
                </button>
                <button
                  type="button"
                  class="paper-btn-primary paper-btn--compact"
                  :disabled="isGenerateExperimentPlanDisabled(item)"
                  @click="openGenerateExperimentPlanConfirm(item)"
                >
                  生成实验方案
                </button>
                <button
                  type="button"
                  class="paper-btn-danger paper-btn--compact"
                  @click="openDeleteLitReviewConfirm(item)"
                >
                  删除
                </button>
              </div>
            </td>
          </tr>
        </tbody>
      </table>
    </section>

    <Teleport to="body">
      <div
        v-if="litReviewGeneratePending"
        class="paper-modal-overlay wf-lit-generate-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="wf-lit-generate-title"
        @keydown.escape="closeGenerateExperimentPlanConfirm"
      >
        <div
          class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-generate-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="wf-lit-generate-title" class="paper-modal-title">生成实验方案</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="litReviewGenerateSubmitting"
              @click="closeGenerateExperimentPlanConfirm"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body">
            <p class="wf-lit-delete-lead">
              确定为「{{ litReviewRowTitle(litReviewGeneratePending) }}」生成实验方案？
            </p>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="litReviewGenerateSubmitting"
              @click="submitGenerateExperimentPlanConfirm"
            >
              {{ litReviewGenerateSubmitting ? '提交中…' : '确定' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="litReviewGenerateSubmitting"
              @click="closeGenerateExperimentPlanConfirm"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="litReviewDeletePending"
        class="paper-modal-overlay wf-lit-delete-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="wf-lit-delete-title"
        @keydown.escape="closeDeleteLitReviewConfirm"
      >
        <div
          class="paper-message-box paper-message-box--danger paper-modal-panel wf-lit-delete-panel"
          @click.stop
        >
          <header class="paper-modal-header">
            <h2 id="wf-lit-delete-title" class="paper-modal-title">删除文献综述</h2>
            <button
              type="button"
              class="paper-modal-close"
              aria-label="关闭"
              :disabled="litReviewDeleteSubmitting"
              @click="closeDeleteLitReviewConfirm"
            >
              ×
            </button>
          </header>
          <div class="paper-modal-body">
            <p class="wf-lit-delete-lead">
              确定删除「{{ litReviewRowTitle(litReviewDeletePending) }}」？
            </p>
          </div>
          <footer class="paper-message-box__btns">
            <button
              type="button"
              class="paper-btn-danger wf-lit-delete-dialog-btn"
              :disabled="litReviewDeleteSubmitting"
              @click="submitDeleteLitReviewConfirm"
            >
              {{ litReviewDeleteSubmitting ? '删除中…' : '删除' }}
            </button>
            <button
              type="button"
              class="paper-btn-primary wf-lit-delete-dialog-btn"
              :disabled="litReviewDeleteSubmitting"
              @click="closeDeleteLitReviewConfirm"
            >
              取消
            </button>
          </footer>
        </div>
      </div>
    </Teleport>

    <Teleport to="body">
      <div
        v-if="litReviewViewItem"
        class="paper-modal-overlay wf-lit-view-overlay"
        role="dialog"
        aria-modal="true"
        aria-labelledby="wf-lit-view-title"
      >
        <div class="paper-message-box paper-message-box--default paper-modal-panel wf-lit-view-panel" @click.stop>
          <header class="paper-modal-header wf-lit-view-header">
            <h2 id="wf-lit-view-title" class="paper-modal-title wf-lit-view-title">
              {{ litReviewRowTitle(litReviewViewItem) }}
            </h2>
            <button type="button" class="paper-modal-close" aria-label="关闭" @click="closeLitReviewView">
              ×
            </button>
          </header>
          <div class="paper-modal-body wf-lit-view-body">
            <p class="wf-meta wf-lit-meta">
              <span>版本 v{{ litReviewViewItem.version }}</span>
              <span>·</span>
              <span>{{ formatLitReviewCreatedAt(litReviewViewItem.created_at) }}</span>
              <span v-if="litReviewViewItem.structure">·</span>
              <span v-if="litReviewViewItem.structure">{{
                LIT_STRUCTURE_LABEL[litReviewViewItem.structure] ?? litReviewViewItem.structure
              }}</span>
              <span>·</span>
              <span>{{ formatLiteratureReviewStatus(litReviewViewItem.status) }}</span>
            </p>
            <p v-if="litReviewViewItem.summary?.trim()" class="wf-lit-summary">
              {{ litReviewViewItem.summary }}
            </p>
            <pre v-if="litReviewViewItem.content_medium?.trim()" class="wf-pre wf-pre--lit">{{
              litReviewViewItem.content_medium
            }}</pre>
            <p v-else class="wf-artifact-empty">暂无正文（content_medium 为空）</p>
          </div>
          <footer class="paper-message-box__btns wf-lit-view-btns">
            <button type="button" class="paper-btn-primary" @click="closeLitReviewView">关闭</button>
          </footer>
        </div>
      </div>
    </Teleport>
  </div>

  <!-- 实验规划 -->
  <div v-else-if="moduleId === 'experiment-planning'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <label class="wf-field wf-field--block">
        <span class="wf-label">核心 idea / 假设 <em class="req">*</em></span>
        <textarea v-model="planForm.ideaSummary" class="wf-textarea" rows="2" />
      </label>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">目标 venue</span>
          <input v-model="planForm.venue" type="text" class="wf-input" />
        </label>
        <label class="wf-field">
          <span class="wf-label">执行强度</span>
          <PaperSelect v-model="planForm.intensity" :options="intensityOptions" />
          <span class="wf-hint">控制检索数量、迭代轮数和输出深度。</span>
        </label>
        <label class="wf-field wf-field--span2">
          <span class="wf-label">基线（逗号或换行）</span>
          <textarea v-model="planForm.baselinesText" class="wf-textarea" rows="2" />
        </label>
        <label class="wf-field wf-field--span2">
          <span class="wf-label">资源与时间</span>
          <input v-model="planForm.resources" type="text" class="wf-input" />
        </label>
      </div>
    </section>
    <section v-if="resultVisible['experiment-planning']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">实验计划（演示）</h2>
      <p class="wf-kv"><span>假设</span>{{ DEMO_EXPERIMENT_PLAN.hypothesis }}</p>
      <p class="wf-kv"><span>基线</span>{{ DEMO_EXPERIMENT_PLAN.baselines.join(' · ') }}</p>
      <p class="wf-kv"><span>指标</span>{{ DEMO_EXPERIMENT_PLAN.metrics.join(' · ') }}</p>
      <p class="wf-kv"><span>消融</span>{{ DEMO_EXPERIMENT_PLAN.ablations.join(' · ') }}</p>
      <ol class="wf-list wf-list--ordered">
        <li v-for="(s, i) in DEMO_EXPERIMENT_PLAN.steps" :key="i">{{ s }}</li>
      </ol>
    </section>
  </div>

  <!-- 结果审查（写前） -->
  <div v-else-if="moduleId === 'auto-review'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <p class="wf-lead">写作前审查：实验方案与上传数据（非完整稿）。</p>
      <div class="wf-field wf-field--block">
        <span class="wf-label">审查对象</span>
        <div class="wf-radios">
          <label><input v-model="reviewForm.reviewFocus" type="radio" value="plan" /> 仅实验方案</label>
          <label><input v-model="reviewForm.reviewFocus" type="radio" value="data" /> 仅上传数据</label>
          <label><input v-model="reviewForm.reviewFocus" type="radio" value="both" /> 方案 + 数据</label>
        </div>
      </div>
      <label class="wf-field wf-field--block">
        <span class="wf-label">数据文件</span>
        <input v-model="reviewForm.dataFileLabel" type="text" class="wf-input" readonly />
      </label>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">审计等级</span>
          <PaperSelect v-model="reviewForm.auditLevel" :options="auditOptions" />
          <span class="wf-hint">控制 citation audit、claim audit、kill argument 等门禁强度。</span>
        </label>
        <label class="wf-check wf-check--solo">
          <input v-model="reviewForm.strictKill" type="checkbox" />
          <span>启用 kill argument（严苛反驳）</span>
        </label>
      </div>
    </section>
    <section v-if="resultVisible['auto-review']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">审查意见（演示）</h2>
      <p class="wf-meta">{{ DEMO_AUTO_REVIEW.target }}</p>
      <div class="wf-scores">
        <span>Rigor {{ DEMO_AUTO_REVIEW.scores.rigor }}</span>
        <span>Complete {{ DEMO_AUTO_REVIEW.scores.completeness }}</span>
        <span>Claims {{ DEMO_AUTO_REVIEW.scores.claimSupport }}</span>
      </div>
      <ul class="wf-findings">
        <li v-for="(f, i) in DEMO_AUTO_REVIEW.findings" :key="i" :class="`wf-finding--${f.level}`">
          <strong>{{ f.level === 'major' ? 'Major' : 'Minor' }}</strong> {{ f.text }}
        </li>
      </ul>
      <p class="wf-callout wf-callout--warn"><strong>Kill：</strong>{{ DEMO_AUTO_REVIEW.kill }}</p>
    </section>
  </div>

  <!-- 论文写作 -->
  <div v-else-if="moduleId === 'paper-writing'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">目标 venue</span>
          <input v-model="writeForm.venue" type="text" class="wf-input" />
        </label>
        <label class="wf-field">
          <span class="wf-label">文风</span>
          <PaperSelect v-model="writeForm.tone" :options="writeToneOptions" />
        </label>
      </div>
      <div class="wf-field wf-field--block">
        <span class="wf-label">生成章节</span>
        <div class="wf-check-group">
          <label v-for="opt in PAPER_WRITING_SECTION_OPTIONS" :key="opt.value" class="wf-check wf-check--inline">
            <input
              type="checkbox"
              :checked="isWritingSectionChecked(opt.value)"
              @change="toggleWritingSection(opt.value, ($event.target as HTMLInputElement).checked)"
            />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </div>
      <label class="wf-check">
        <input v-model="writeForm.useLitReview" type="checkbox" />
        <span>引用已有文献综述 artifact</span>
      </label>
      <label class="wf-check">
        <input v-model="writeForm.useReviewArtifact" type="checkbox" />
        <span>参考写前审查意见修订 Experiments 表述</span>
      </label>
    </section>
    <section v-if="resultVisible['paper-writing']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">稿件（演示）</h2>
      <p class="wf-kv"><span>标题</span>{{ DEMO_MANUSCRIPT.title }}</p>
      <table class="wf-table">
        <thead>
          <tr>
            <th>章节</th>
            <th>字数</th>
            <th>状态</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="s in DEMO_MANUSCRIPT.sections" :key="s.name">
            <td>{{ s.name }}</td>
            <td>{{ s.words }}</td>
            <td>
              <span class="wf-tag" :class="s.status === 'done' ? 'wf-tag--ok' : 'wf-tag--draft'">{{
                s.status
              }}</span>
            </td>
          </tr>
        </tbody>
      </table>
      <p class="wf-meta">引用 {{ DEMO_MANUSCRIPT.cites }} · 已验证 {{ DEMO_MANUSCRIPT.verifiedCites }}</p>
      <pre class="wf-pre">{{ DEMO_MANUSCRIPT.excerpt }}</pre>
    </section>
  </div>

  <!-- 图表管理 -->
  <section v-else-if="moduleId === 'figure-generation'" class="fig-mgmt-panel">
    <div class="fig-mgmt-tabs" role="tablist" aria-label="图表管理">
      <button
        type="button"
        role="tab"
        class="fig-mgmt-tab"
        :class="{ 'fig-mgmt-tab--active': figureTab === 'upload' }"
        :aria-selected="figureTab === 'upload'"
        @click="figureTab = 'upload'"
      >
        上传图表
      </button>
      <button
        type="button"
        role="tab"
        class="fig-mgmt-tab"
        :class="{ 'fig-mgmt-tab--active': figureTab === 'generate' }"
        :aria-selected="figureTab === 'generate'"
        @click="figureTab = 'generate'"
      >
        一键生成
      </button>
    </div>

    <div v-show="figureTab === 'upload'" class="fig-mgmt-pane" role="tabpanel">
      <PaperFigureUploadPanel
        embedded
        :manuscript-id="manuscriptId"
        :manuscript-title="manuscriptTitle"
      />
    </div>

    <div v-show="figureTab === 'generate'" class="fig-mgmt-pane fig-mgmt-pane--stack" role="tabpanel">
    <section class="wf-panel wf-panel--in-fig-mgmt">
      <h2 class="wf-title">参数</h2>
      <p class="wf-lead">基于实验数据生成可复现统计图（非通用美工工具）。</p>
      <label class="wf-field wf-field--block">
        <span class="wf-label">数据文件</span>
        <PaperSelect v-model="figureForm.datasetLabel" :options="figureDatasetOptions" />
      </label>
      <div class="wf-field wf-field--block">
        <span class="wf-label">图表类型</span>
        <div class="wf-check-group">
          <label v-for="opt in FIGURE_CHART_OPTIONS" :key="opt.value" class="wf-check wf-check--inline">
            <input
              type="checkbox"
              :checked="isChartChecked(opt.value)"
              @change="toggleChartType(opt.value, ($event.target as HTMLInputElement).checked)"
            />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </div>
      <label class="wf-check">
        <input v-model="figureForm.includeErrorBars" type="checkbox" />
        <span>误差条 / 多 seed 汇总</span>
      </label>
      <label class="wf-field">
        <span class="wf-label">Caption 语言</span>
        <PaperSelect v-model="figureForm.captionLang" :options="captionLangOptions" />
      </label>
    </section>
    <section
      v-if="resultVisible['figure-generation']"
      class="wf-panel wf-panel--in-fig-mgmt wf-panel--in-fig-mgmt-result"
    >
      <h2 class="wf-title">Figure 产出（演示）</h2>
      <div class="wf-figure-grid">
        <div v-for="fig in DEMO_FIGURES.files" :key="fig.id" class="wf-figure-card">
          <div class="wf-figure-preview" :class="`wf-figure-preview--${fig.type}`">
            <div v-for="n in 5" :key="n" class="wf-bar" :style="{ height: `${30 + n * 12}%` }" />
          </div>
          <p class="wf-figure-cap">{{ fig.title }}</p>
          <code class="wf-figure-file">{{ fig.note }}</code>
        </div>
      </div>
      <table class="wf-table wf-table--compact">
        <thead>
          <tr>
            <th>Method</th>
            <th>Score</th>
            <th>±</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in DEMO_FIGURES.dataPreview" :key="row.method">
            <td>{{ row.method }}</td>
            <td>{{ row.score }}</td>
            <td>{{ row.std }}</td>
          </tr>
        </tbody>
      </table>
    </section>
    </div>
  </section>

  <!-- 论文审查 -->
  <div v-else-if="moduleId === 'manuscript-analysis'" class="wf-stack">
    <section class="wf-panel">
      <h2 class="wf-title">参数</h2>
      <p class="wf-lead">写作完成后对全文做投稿前全面分析。</p>
      <div class="wf-grid">
        <label class="wf-field">
          <span class="wf-label">分析深度</span>
          <PaperSelect v-model="analysisForm.depth" :options="auditOptions" />
        </label>
      </div>
      <label class="wf-check">
        <input v-model="analysisForm.includeFigures" type="checkbox" />
        <span>纳入图表与表（{{ DEMO_FIGURES.files.length }} 个 demo figure）</span>
      </label>
      <label class="wf-check">
        <input v-model="analysisForm.includeBib" type="checkbox" />
        <span>纳入参考文献门禁结果</span>
      </label>
    </section>
    <section v-if="resultVisible['manuscript-analysis']" class="wf-panel wf-panel--result">
      <h2 class="wf-title">论文审查报告（演示）</h2>
      <p class="wf-score-big">总评 {{ DEMO_MANUSCRIPT_ANALYSIS.overall }} / 10</p>
      <p class="wf-meta">{{ DEMO_MANUSCRIPT_ANALYSIS.recommendation }}</p>
      <div class="wf-dim-grid">
        <div v-for="d in DEMO_MANUSCRIPT_ANALYSIS.dimensions" :key="d.name" class="wf-dim">
          <span>{{ d.name }}</span>
          <strong>{{ d.score }}</strong>
        </div>
      </div>
      <h3 class="wf-subtitle">Must fix</h3>
      <ul class="wf-list">
        <li v-for="(m, i) in DEMO_MANUSCRIPT_ANALYSIS.mustFix" :key="i">{{ m }}</li>
      </ul>
      <p class="wf-callout wf-callout--warn"><strong>Kill argument：</strong>{{ DEMO_MANUSCRIPT_ANALYSIS.kill }}</p>
    </section>
  </div>
  </div>
</template>

<style scoped>
.wf-root {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.wf-stack {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.fig-mgmt-panel {
  padding: 20px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.fig-mgmt-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 0 2px;
  margin-bottom: 24px;
  background: transparent;
  border-bottom: 1px solid #e2e8f0;
}

.fig-mgmt-pane--stack {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.fig-mgmt-pane .wf-panel--in-fig-mgmt {
  padding: 0;
  margin: 0;
  background: transparent;
  border: none;
  border-radius: 0;
  box-shadow: none;
}

.fig-mgmt-pane .wf-panel--in-fig-mgmt-result {
  padding-top: 24px;
  margin-top: 24px;
  border-top: 1px solid #e2e8f0;
}

.fig-mgmt-tab {
  position: relative;
  flex: 0 1 auto;
  min-width: 112px;
  padding: 12px 20px;
  margin-bottom: -1px;
  font-size: 14px;
  font-weight: 500;
  color: #64748b;
  cursor: pointer;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
  border-radius: 8px 8px 0 0;
  transition:
    color 0.15s,
    background 0.15s,
    border-color 0.15s;
}

.fig-mgmt-tab:hover:not(.fig-mgmt-tab--active) {
  color: #334155;
  background: #f8fafc;
}

.fig-mgmt-tab--active {
  font-weight: 600;
  color: #1e293b;
  background: #fff;
  border-bottom-color: #6366f1;
  box-shadow: inset 0 -1px 0 #fff;
}

.wf-panel {
  padding: 24px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.wf-panel--result {
  margin-top: 18px;
  border-color: #ddd6fe;
  background: linear-gradient(180deg, #faf5ff 0%, #fff 120px);
}

.wf-title {
  margin: 0 0 18px;
  font-size: 16px;
  font-weight: 700;
  color: #1e1b4b;
}

.wf-title--tight {
  margin-bottom: 8px;
}

.wf-lit-list-panel {
  padding: 20px 22px 24px;
}

.wf-lit-table {
  margin: 0;
}

.wf-lit-col-title {
  max-width: min(420px, 40vw);
  font-weight: 600;
  color: #1e1b4b;
}

.wf-lit-col-actions {
  min-width: 280px;
}

.wf-lit-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
}

.wf-lit-head-slot {
  position: relative;
  display: inline-flex;
}

.wf-lit-width-ruler {
  visibility: hidden;
  pointer-events: none;
}

.wf-lit-col-head-label {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  padding: 6px 12px;
  font-size: 13px;
  font-weight: 700;
  color: #64748b;
  pointer-events: none;
}

.wf-lit-delete-overlay,
.wf-lit-generate-overlay,
.wf-lit-view-overlay {
  z-index: 3200;
}

.wf-lit-generate-panel {
  width: min(420px, calc(100vw - 32px));
}

.wf-lit-delete-panel {
  width: min(420px, calc(100vw - 32px));
}

.wf-lit-delete-panel .paper-message-box__btns {
  flex-direction: row-reverse;
  justify-content: flex-start;
}

.wf-lit-delete-panel .paper-message-box__btns .wf-lit-delete-dialog-btn {
  min-width: 88px;
  box-sizing: border-box;
}

.wf-lit-delete-lead {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--atm-text, #1e1b4b);
  line-height: 1.5;
}

.wf-lit-view-panel {
  display: flex;
  flex-direction: column;
  width: min(920px, calc(100vw - 32px));
  max-height: min(90vh, 880px);
}

.wf-lit-view-header {
  display: flex;
  gap: 12px;
  align-items: flex-start;
  justify-content: space-between;
}

.wf-lit-view-title {
  margin: 0;
  font-size: 16px;
  line-height: 1.45;
}

.wf-lit-view-body {
  flex: 1;
  overflow: auto;
}

.wf-lit-view-btns {
  flex-direction: row-reverse;
  justify-content: flex-start;
}

.wf-lit-meta {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.wf-lit-summary {
  margin: 0 0 16px;
  font-size: 14px;
  line-height: 1.65;
  color: #334155;
}

.wf-pre--lit {
  max-height: min(60vh, 720px);
  overflow: auto;
}

.wf-subtitle {
  margin: 16px 0 8px;
  font-size: 14px;
  font-weight: 700;
  color: #1e1b4b;
}

.wf-pipeline {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-bottom: 20px;
  font-size: 13px;
  color: #475569;
}

.wf-pipeline > span:not(.wf-pipeline-sep) {
  padding: 6px 12px;
  background: #f1f5f9;
  border-radius: 999px;
}

.wf-pipeline-sep {
  color: #94a3b8;
}

.wf-subhead {
  margin: 22px 0 10px;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #64748b;
}

.wf-hint {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  line-height: 1.45;
  color: #64748b;
}

.wf-check-group--tight {
  margin-top: 12px;
}

.wf-lead {
  margin: -8px 0 16px;
  font-size: 13px;
  color: #64748b;
}

.wf-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 18px 20px;
}

@media (max-width: 900px) {
  .wf-grid {
    grid-template-columns: 1fr;
  }
}

.wf-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.wf-field--block {
  margin-bottom: 4px;
}

.wf-field--span2 {
  grid-column: 1 / -1;
}

.wf-label {
  font-size: 13px;
  font-weight: 600;
  color: #1e1b4b;
}

.wf-label .req {
  color: #e53935;
  font-style: normal;
}

.wf-input,
.wf-textarea {
  width: 100%;
  padding: 10px 12px;
  font-size: 14px;
  color: #1e1b4b;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.wf-textarea {
  resize: vertical;
  font-family: inherit;
  line-height: 1.5;
}

.wf-check {
  display: flex;
  gap: 10px;
  align-items: flex-start;
  margin-top: 14px;
  font-size: 14px;
  cursor: pointer;
}

.wf-check--inline {
  margin-top: 0;
}

.wf-check--solo {
  justify-content: flex-end;
  align-self: end;
}

.wf-check-group {
  display: flex;
  flex-wrap: wrap;
  gap: 12px 18px;
  margin-top: 8px;
}

.wf-radios {
  display: flex;
  flex-wrap: wrap;
  gap: 14px 20px;
  margin-top: 8px;
  font-size: 14px;
}

.wf-meta {
  margin: 0 0 12px;
  font-size: 13px;
  color: #64748b;
}

.wf-kv {
  margin: 0 0 10px;
  font-size: 14px;
  line-height: 1.5;
}

.wf-kv span {
  display: inline-block;
  min-width: 4em;
  margin-right: 8px;
  font-weight: 700;
  color: #64748b;
}

.wf-list {
  margin: 0;
  padding-left: 20px;
  font-size: 13px;
  line-height: 1.55;
}

.wf-list--ordered {
  margin-top: 12px;
}

.wf-pre {
  margin: 12px 0 0;
  padding: 12px;
  font-size: 12px;
  line-height: 1.5;
  color: #334155;
  white-space: pre-wrap;
  background: #f1f5f9;
  border-radius: 10px;
}

.wf-callout {
  margin: 16px 0 0;
  padding: 12px 14px;
  font-size: 13px;
  line-height: 1.5;
  background: #f0fdf4;
  border-radius: 10px;
}

.wf-callout--warn {
  background: #fffbeb;
}

.wf-artifact-grid {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 8px;
}

.wf-artifact-card {
  padding: 14px 16px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.wf-artifact-card--wide {
  grid-column: 1 / -1;
}

.wf-artifact-title {
  margin: 0 0 8px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.04em;
  text-transform: uppercase;
  color: #64748b;
}

.wf-artifact-body {
  margin: 0 0 6px;
  font-size: 14px;
  line-height: 1.5;
  color: #1e1b4b;
}

.wf-artifact-meta {
  margin: 0;
  font-size: 12px;
  color: #64748b;
}

.wf-artifact-stat {
  margin: 0 0 4px;
  font-size: 22px;
  font-weight: 700;
  color: #4f46e5;
}

.wf-artifact-stat span {
  font-size: 13px;
  font-weight: 500;
  color: #64748b;
}

.wf-artifact-empty {
  margin: 0;
  font-size: 13px;
  color: #94a3b8;
}

.wf-list--tight {
  margin-top: 0;
  font-size: 13px;
}

.wf-grid--run {
  margin-top: 18px;
}

.wf-block {
  margin-top: 8px;
}

.wf-scores {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 12px;
  font-size: 13px;
  font-weight: 600;
}

.wf-scores span {
  padding: 6px 10px;
  background: #ede9fe;
  border-radius: 8px;
}

.wf-findings {
  margin: 0;
  padding: 0;
  list-style: none;
}

.wf-finding--major,
.wf-finding--minor {
  margin-bottom: 8px;
  padding: 10px 12px;
  font-size: 13px;
  border-radius: 10px;
}

.wf-finding--major {
  background: #fee2e2;
}

.wf-finding--minor {
  background: #f1f5f9;
}

.wf-table {
  width: 100%;
  margin: 12px 0;
  font-size: 13px;
  border-collapse: collapse;
}

.wf-table th,
.wf-table td {
  padding: 8px 10px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.wf-table th {
  font-weight: 700;
  color: #64748b;
}

.wf-tag {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
}

.wf-tag--ok {
  color: #15803d;
}

.wf-tag--draft {
  color: #b45309;
}

.wf-figure-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 14px;
}

@media (max-width: 900px) {
  .wf-figure-grid {
    grid-template-columns: 1fr;
  }
}

.wf-figure-card {
  padding: 12px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.wf-figure-preview {
  display: flex;
  gap: 6px;
  align-items: flex-end;
  justify-content: center;
  height: 100px;
  padding: 8px;
  background: #f8fafc;
  border-radius: 8px;
}

.wf-bar {
  flex: 1;
  max-width: 24px;
  background: linear-gradient(180deg, #7c3aed, #6366f1);
  border-radius: 4px 4px 0 0;
}

.wf-figure-preview--line .wf-bar {
  background: #6366f1;
  opacity: 0.85;
}

.wf-figure-cap {
  margin: 10px 0 4px;
  font-size: 13px;
  font-weight: 600;
}

.wf-figure-file {
  font-size: 11px;
  color: #64748b;
}

.wf-score-big {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  color: #5b21b6;
}

.wf-dim-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 10px;
  margin: 16px 0;
}

@media (max-width: 700px) {
  .wf-dim-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}

.wf-dim {
  padding: 12px;
  text-align: center;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.wf-dim span {
  display: block;
  font-size: 11px;
  color: #64748b;
}

.wf-dim strong {
  font-size: 18px;
  color: #1e1b4b;
}
</style>
