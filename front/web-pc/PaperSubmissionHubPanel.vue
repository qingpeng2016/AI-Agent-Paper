<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type { ManuscriptSubmissionStatus } from '@/api/manuscripts'
import {
  createManuscriptSubmission,
  fetchManuscriptSubmissionDetail,
  fetchManuscriptSubmissions,
  markManuscriptSubmissionResubmitted,
  submissionSnapshotLabel,
  submissionTrackLabel,
  updateManuscriptSubmissionRevision,
  uploadManuscriptSubmissionMainPdf,
  type PaperManuscriptSubmissionDetail,
  type PaperManuscriptSubmissionItem,
} from '@/api/manuscriptSubmissions'
import PaperSelect from './PaperSelect.vue'
import PaperTableActionsHead from './PaperTableActionsHead.vue'

type HubScreen = 'list' | 'wizard' | 'detail' | 'revision'

const props = defineProps<{
  activeManuscriptId: string
  manuscriptTitle: string
  draftVersion?: number
}>()

const emit = defineEmits<{
  manuscriptSummaryChanged: []
  startRevisionWriting: [payload: { submissionId: string }]
}>()

const STATUS_LABEL: Record<ManuscriptSubmissionStatus, string> = {
  not_submitted: '未投递',
  submitted: '已提交',
  under_review: '审稿中',
  revision: '待改稿',
  rebuttal: 'Rebuttal',
  resubmitted: '已修回',
  accepted: '录用',
  rejected: '拒稿',
}

function submissionStatusTone(status: ManuscriptSubmissionStatus): string {
  switch (status) {
    case 'accepted':
      return 'success'
    case 'revision':
    case 'rebuttal':
      return 'draft'
    case 'rejected':
      return 'muted'
    case 'submitted':
    case 'under_review':
    case 'resubmitted':
      return 'progress'
    default:
      return 'default'
  }
}

const VENUE_OPTIONS = [
  { value: 'neurips_2026', label: 'NeurIPS 2026' },
  { value: 'iclr_2026', label: 'ICLR 2026' },
  { value: 'ieee_tmi', label: 'IEEE TMI' },
  { value: 'acl_2026', label: 'ACL 2026' },
]

const SNAPSHOT_OPTIONS = computed(() => {
  const max = Math.max(1, props.draftVersion ?? 1)
  return Array.from({ length: max }, (_, i) => {
    const v = i + 1
    return { value: String(v), label: `v${v} · 写作快照` }
  })
})

const screen = ref<HubScreen>('list')
const wizardStep = ref(1)
const detailTab = ref<'overview' | 'package' | 'checklist' | 'log'>('overview')
const listLoading = ref(false)
const actionBusy = ref(false)
const submissions = ref<PaperManuscriptSubmissionItem[]>([])
const activeDetail = ref<PaperManuscriptSubmissionDetail | null>(null)

const wizardForm = reactive({
  venueCode: '',
  track: 'main' as 'main' | 'short',
  portalUrl: '',
  snapshot: '1',
  anonymous: true,
  mainPdfFile: null as File | null,
  mainPdfName: '',
  title: '',
  abstract: '',
  externalId: '',
  markedSubmitted: false,
})

const revisionForm = reactive({
  decision: '',
  comments: '',
})

const wizardChecklistItems = [
  { id: 'pages', label: '页数符合 venue 限制', ok: true },
  { id: 'anon', label: 'PDF 匿名化检查', ok: true },
  { id: 'supp', label: '补充材料已打包', ok: false },
]

const WIZARD_STEPS = [
  { id: 1, label: '投稿目标' },
  { id: 2, label: '稿本绑定' },
  { id: 3, label: '投稿材料' },
  { id: 4, label: '确认投递' },
] as const

function goWizardStep(step: number) {
  if (step >= 1 && step <= 4 && step <= wizardStep.value) {
    wizardStep.value = step
  }
}

const detailChecklist = computed(() => activeDetail.value?.checklist ?? wizardChecklistItems)

type SubmissionRoundLogBlock = {
  round: number
  outgoing: {
    submittedAt?: string
    title?: string
    abstract?: string
    draftVersion?: number
    pdfFilename?: string
    externalId?: string
  }
  incoming: {
    returnedAt?: string
    decision?: string
    comments?: string
    pending?: boolean
  } | null
}

function buildSubmissionRoundLogs(d: PaperManuscriptSubmissionDetail): SubmissionRoundLogBlock[] {
  const events = [...(d.events ?? [])].sort(
    (a, b) => new Date(a.created_at).getTime() - new Date(b.created_at).getTime(),
  )
  const maxR = Math.max(0, d.round)
  const blocks: SubmissionRoundLogBlock[] = []

  for (let r = 0; r <= maxR; r++) {
    const isLatestRound = r === maxR
    let submittedAt: string | undefined
    if (r === 0) {
      submittedAt =
        events.find((e) => e.event_type === 'submitted')?.created_at ?? d.submitted_at ?? undefined
    } else {
      submittedAt =
        events.find((e) => e.event_type === 'resubmitted' && e.message.includes(`R${r}`))
          ?.created_at ??
        events.filter((e) => e.event_type === 'resubmitted')[r - 1]?.created_at ??
        undefined
    }

    const outgoing = {
      submittedAt,
      title: r === 0 || isLatestRound ? d.title || props.manuscriptTitle : undefined,
      abstract: r === 0 || isLatestRound ? d.abstract : undefined,
      draftVersion: isLatestRound ? d.draft_version : undefined,
      pdfFilename: isLatestRound ? d.main_pdf_filename : undefined,
      externalId: r === 0 ? d.external_id : undefined,
    }

    let incoming: SubmissionRoundLogBlock['incoming'] = null

    if (r < maxR) {
      if (r === 0 && (d.decision || d.comments_summary)) {
        incoming = {
          returnedAt: events.find((e) => e.event_type === 'revision_saved')?.created_at,
          decision: d.decision ?? undefined,
          comments: d.comments_summary ?? undefined,
        }
      }
    } else if (maxR === 0) {
      if (d.decision || d.comments_summary) {
        incoming = {
          returnedAt: events.find((e) => e.event_type === 'revision_saved')?.created_at,
          decision: d.decision ?? undefined,
          comments: d.comments_summary ?? undefined,
        }
      } else if (d.status === 'under_review' || d.status === 'submitted') {
        incoming = { pending: true, comments: '已提交，等待编辑部/审稿意见。' }
      }
    } else if (isLatestRound && r > 0) {
      if (d.status === 'resubmitted' || d.status === 'under_review') {
        incoming = {
          pending: true,
          comments: '修回已提交，等待本轮审稿或 meta 结论。',
          returnedAt: submittedAt,
        }
      } else if (d.decision || d.comments_summary) {
        incoming = {
          decision: d.decision ?? undefined,
          comments: d.comments_summary ?? undefined,
        }
      }
    }

    blocks.push({ round: r, outgoing, incoming })
  }

  return blocks
}

const submissionRoundLogs = computed(() => {
  const d = activeDetail.value
  if (!d) return []
  return buildSubmissionRoundLogs(d)
})

async function loadList() {
  const msId = props.activeManuscriptId
  if (!msId || !/^\d+$/.test(msId)) {
    submissions.value = []
    return
  }
  listLoading.value = true
  try {
    const data = await fetchManuscriptSubmissions(msId)
    submissions.value = data.items ?? []
  } catch (e) {
    submissions.value = []
    ElMessage.error(e instanceof Error ? e.message : '加载投递列表失败')
  } finally {
    listLoading.value = false
  }
}

watch(
  () => props.activeManuscriptId,
  () => {
    screen.value = 'list'
    activeDetail.value = null
    void loadList()
  },
  { immediate: true },
)

watch(
  () => props.manuscriptTitle,
  (t) => {
    if (t && !wizardForm.title) wizardForm.title = t
  },
  { immediate: true },
)

function openWizard() {
  wizardStep.value = 1
  wizardForm.venueCode = ''
  wizardForm.track = 'main'
  wizardForm.portalUrl = ''
  wizardForm.snapshot = SNAPSHOT_OPTIONS.value.at(-1)?.value ?? '1'
  wizardForm.anonymous = true
  wizardForm.mainPdfFile = null
  wizardForm.mainPdfName = ''
  wizardForm.title = props.manuscriptTitle || ''
  wizardForm.abstract = ''
  wizardForm.externalId = ''
  wizardForm.markedSubmitted = false
  screen.value = 'wizard'
}

async function loadDetail(id: string) {
  const msId = props.activeManuscriptId
  if (!msId) return null
  try {
    return await fetchManuscriptSubmissionDetail(msId, id)
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '加载投递详情失败')
    return null
  }
}

async function openDetail(id: string, tab: typeof detailTab.value = 'overview') {
  const detail = await loadDetail(id)
  if (!detail) return
  activeDetail.value = detail
  detailTab.value = tab
  screen.value = 'detail'
}

function openRevision(id: string) {
  emit('startRevisionWriting', { submissionId: id })
}

function backToList() {
  screen.value = 'list'
  activeDetail.value = null
  void loadList()
}

async function finishWizard() {
  if (!wizardForm.venueCode) {
    ElMessage.warning('请选择目标 venue')
    return
  }
  if (!wizardForm.markedSubmitted) {
    ElMessage.warning('请确认已在投稿系统完成提交，并勾选确认项')
    return
  }
  const msId = props.activeManuscriptId
  if (!msId) return
  const venueLabel = VENUE_OPTIONS.find((v) => v.value === wizardForm.venueCode)?.label ?? wizardForm.venueCode
  actionBusy.value = true
  try {
    let detail = await createManuscriptSubmission(msId, {
      venue_code: wizardForm.venueCode,
      venue_label: venueLabel,
      track: wizardForm.track,
      draft_version: Number(wizardForm.snapshot) || 1,
      anonymous: wizardForm.anonymous,
      portal_url: wizardForm.portalUrl.trim() || undefined,
      external_id: wizardForm.externalId.trim() || undefined,
      title: wizardForm.title.trim() || undefined,
      abstract: wizardForm.abstract.trim() || undefined,
      marked_submitted: true,
    })
    if (wizardForm.mainPdfFile) {
      detail = await uploadManuscriptSubmissionMainPdf(msId, detail.id, wizardForm.mainPdfFile)
    }
    ElMessage.success('已记录投递')
    emit('manuscriptSummaryChanged')
    await loadList()
    activeDetail.value = detail
    detailTab.value = 'overview'
    screen.value = 'detail'
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '创建投递失败')
  } finally {
    actionBusy.value = false
  }
}

async function saveRevisionPlan() {
  const row = activeDetail.value
  const msId = props.activeManuscriptId
  if (!row || !msId) return
  actionBusy.value = true
  try {
    const detail = await updateManuscriptSubmissionRevision(
      msId,
      row.id,
      revisionForm.decision,
      revisionForm.comments,
    )
    activeDetail.value = detail
    ElMessage.success('改稿计划已保存')
    emit('manuscriptSummaryChanged')
    await loadList()
    screen.value = 'detail'
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '保存改稿计划失败')
  } finally {
    actionBusy.value = false
  }
}

async function onDetailSubmitAction() {
  const d = activeDetail.value
  if (!d) return
  if (d.status === 'revision' || d.status === 'rebuttal') {
    await markResubmitted()
    return
  }
  const url = d.portal_url?.trim()
  if (url) {
    window.open(url, '_blank', 'noopener,noreferrer')
    ElMessage.info('已打开投稿系统；修回完成后请保存状态为待改稿/修回后再点「投递」标记已修回')
  } else {
    ElMessage.warning('暂无投稿系统链接，请在「概览」中补充 Portal URL')
  }
}

async function markResubmitted() {
  const row = activeDetail.value
  const msId = props.activeManuscriptId
  if (!row || !msId) return
  actionBusy.value = true
  try {
    const detail = await markManuscriptSubmissionResubmitted(msId, row.id)
    activeDetail.value = detail
    ElMessage.success('已标记修回提交')
    emit('manuscriptSummaryChanged')
    await loadList()
    screen.value = 'detail'
  } catch (e) {
    ElMessage.error(e instanceof Error ? e.message : '标记修回失败')
  } finally {
    actionBusy.value = false
  }
}

function onWizardPdfChange(e: Event) {
  const f = (e.target as HTMLInputElement).files?.[0]
  wizardForm.mainPdfFile = f ?? null
  wizardForm.mainPdfName = f?.name ?? ''
}

defineExpose({ openWizard })
</script>

<template>
  <!-- 与文献综述模块相同：wf-stack 直下 wf-panel（wf-root 在 WorkflowPanels 组件层） -->
  <div class="wf-stack sub-hub">
    <template v-if="screen === 'list'">
      <section v-if="!activeManuscriptId" class="wf-panel">
        <p class="wf-meta">请先在顶栏或「我的论文」选择当前论文。</p>
      </section>

      <section v-else-if="listLoading" class="wf-panel">
        <p class="wf-meta">正在加载投递列表…</p>
      </section>

      <section v-else-if="submissions.length === 0" class="wf-panel">
        <p class="wf-lead wf-lead--no-head">
          当前论文尚无对外投递记录。完成写作并冻结稿本后，请点顶栏「新建投递」：选择 venue、绑定 PDF，并标记已在 portal 提交。
        </p>
      </section>

      <section v-else class="wf-panel wf-lit-list-panel sub-hub-list-panel">
        <table class="wf-table wf-lit-table wf-lit-table--submission">
          <colgroup>
            <col class="wf-sub-col-venue-pct" />
            <col class="wf-sub-col-round-pct" />
            <col class="wf-sub-col-status-pct" />
            <col class="wf-sub-col-version-pct" />
            <col class="wf-sub-col-date-pct" />
            <col class="wf-sub-col-actions-pct" />
          </colgroup>
          <thead>
            <tr>
              <th class="wf-lit-col-title">Venue</th>
              <th class="wf-lit-col-meta wf-sub-col-round">轮次</th>
              <th class="wf-lit-col-meta wf-sub-col-status">状态</th>
              <th class="wf-lit-col-meta wf-sub-col-version">稿本</th>
              <th class="wf-lit-col-meta wf-sub-col-date">提交时间</th>
              <th class="wf-lit-col-actions ul-col-actions">
                <PaperTableActionsHead tip="改稿前往论文写作；投递打开投递管理" />
              </th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in submissions" :key="row.id">
              <td class="wf-lit-col-title">
                {{ row.venue_label }} · {{ submissionTrackLabel(row.track) }}
              </td>
              <td class="wf-lit-col-meta wf-sub-col-round">R{{ row.round }}</td>
              <td class="wf-lit-col-meta wf-sub-col-status">
                <span
                  class="wf-status-pill"
                  :class="`wf-status-pill--${submissionStatusTone(row.status)}`"
                >
                  {{ STATUS_LABEL[row.status] }}
                </span>
              </td>
              <td class="wf-lit-col-meta wf-sub-col-version">
                <span class="wf-output-version">{{ submissionSnapshotLabel(row.draft_version) }}</span>
              </td>
              <td class="wf-lit-col-meta wf-sub-col-date wf-lit-created-at">
                {{ row.submitted_at ? new Date(row.submitted_at).toLocaleDateString('zh-CN') : '—' }}
              </td>
              <td class="wf-lit-col-actions ul-col-actions">
                <button type="button" class="ul-btn-link" @click="openRevision(row.id)">改稿</button>
                <button type="button" class="ul-btn-link" @click="openDetail(row.id)">投递</button>
              </td>
            </tr>
          </tbody>
        </table>
      </section>
    </template>

    <template v-else-if="screen === 'wizard'">
      <nav class="wf-subnav">
        <button type="button" class="ul-btn-link" @click="backToList">← 投递列表</button>
        <span class="wf-subnav-sep">/</span>
        <span>新建投递</span>
      </nav>

      <section class="wf-panel sub-wizard-panel">
        <header class="sub-detail-panel-head sub-wizard-panel-head">
          <div class="sub-detail-panel-head-text">
            <h2 class="sub-detail-panel-title">新建投递</h2>
            <p class="sub-detail-panel-lead">步骤 {{ wizardStep }} / 4 · 先在 Portal 完成提交，再在此记录并同步「我的论文」</p>
          </div>
        </header>

        <div class="ul-tabs sub-wizard-step-tabs" role="tablist" aria-label="新建投递步骤">
          <button
            v-for="s in WIZARD_STEPS"
            :key="s.id"
            type="button"
            role="tab"
            class="ul-tab sub-wizard-step-tab"
            :class="{
              'ul-tab--active': wizardStep === s.id,
              'sub-wizard-step-tab--done': wizardStep > s.id,
            }"
            :aria-selected="wizardStep === s.id"
            :disabled="s.id > wizardStep"
            @click="goWizardStep(s.id)"
          >
            <span class="sub-wizard-step-num">{{ s.id }}</span>
            {{ s.label }}
          </button>
        </div>

        <div class="ul-tab-pane sub-wizard-pane" role="tabpanel">
          <template v-if="wizardStep === 1">
            <p class="sub-pane-lead">选择目标会议/期刊、Track 与投稿系统链接。</p>
            <div class="sub-wizard-fields">
              <label class="sub-field">
                <span class="sub-label">目标 venue</span>
                <PaperSelect
                  v-model="wizardForm.venueCode"
                  :options="VENUE_OPTIONS"
                  placeholder="请选择"
                  filterable
                />
              </label>
              <label class="sub-field">
                <span class="sub-label">Track</span>
                <select v-model="wizardForm.track" class="sub-input">
                  <option value="main">Main</option>
                  <option value="short">Short</option>
                </select>
              </label>
              <label class="sub-field">
                <span class="sub-label">投稿系统链接</span>
                <input
                  v-model="wizardForm.portalUrl"
                  class="sub-input"
                  type="url"
                  placeholder="OpenReview / ScholarOne URL"
                />
              </label>
            </div>
          </template>

          <template v-else-if="wizardStep === 2">
            <p class="sub-pane-lead">绑定写作快照并上传主 PDF；校验项可在提交前自检。</p>
            <div class="sub-wizard-fields">
              <label class="sub-field">
                <span class="sub-label">绑定写作快照</span>
                <PaperSelect v-model="wizardForm.snapshot" :options="SNAPSHOT_OPTIONS" placeholder="选择版本" />
              </label>
              <label class="sub-check sub-check--block">
                <input v-model="wizardForm.anonymous" type="checkbox" />
                双盲投稿（校验 PDF 不含作者信息）
              </label>
              <label class="sub-field">
                <span class="sub-label">主 PDF</span>
                <input class="sub-input sub-input--file" type="file" accept=".pdf" @change="onWizardPdfChange" />
                <span v-if="wizardForm.mainPdfName" class="sub-field-hint">{{ wizardForm.mainPdfName }}</span>
              </label>
            </div>
            <ul class="sub-check-rows sub-check-rows--compact">
              <li v-for="c in wizardChecklistItems" :key="c.id" class="sub-check-row">
                <span class="sub-check-row-label">{{ c.label }}</span>
                <span
                  class="sub-check-row-status"
                  :class="c.ok ? 'sub-check-row-status--ok' : 'sub-check-row-status--warn'"
                >
                  {{ c.ok ? '通过' : '待处理' }}
                </span>
              </li>
            </ul>
          </template>

          <template v-else-if="wizardStep === 3">
            <p class="sub-pane-lead">提交用标题与摘要；可与写作快照不一致时在此覆盖。</p>
            <div class="sub-wizard-fields">
              <label class="sub-field">
                <span class="sub-label">标题</span>
                <input v-model="wizardForm.title" class="sub-input" type="text" />
              </label>
              <label class="sub-field">
                <span class="sub-label">摘要</span>
                <textarea
                  v-model="wizardForm.abstract"
                  class="sub-textarea"
                  rows="5"
                  placeholder="可与写作快照不一致时在此覆盖"
                />
              </label>
            </div>
          </template>

          <template v-else>
            <p class="sub-pane-lead">确认信息无误后，勾选已在 Portal 完成提交。</p>
            <dl class="sub-detail-dl sub-wizard-confirm-dl">
              <dt>Venue</dt>
              <dd>{{ VENUE_OPTIONS.find((v) => v.value === wizardForm.venueCode)?.label || '—' }}</dd>
              <dt>Track</dt>
              <dd>{{ wizardForm.track === 'short' ? 'Short' : 'Main' }}</dd>
              <dt>稿本</dt>
              <dd>v{{ wizardForm.snapshot }}</dd>
              <dt>主 PDF</dt>
              <dd>{{ wizardForm.mainPdfName || '未选择' }}</dd>
            </dl>
            <div class="sub-wizard-fields sub-wizard-fields--after-summary">
              <label class="sub-field">
                <span class="sub-label">Portal Submission ID（可选）</span>
                <input
                  v-model="wizardForm.externalId"
                  class="sub-input"
                  type="text"
                  placeholder="OpenReview forum id"
                />
              </label>
              <label class="sub-check sub-check--block">
                <input v-model="wizardForm.markedSubmitted" type="checkbox" />
                我已在投稿系统完成本次提交
              </label>
            </div>
          </template>
        </div>

        <footer class="sub-wizard-panel-footer">
          <button v-if="wizardStep > 1" type="button" class="paper-btn-secondary" @click="wizardStep--">
            上一步
          </button>
          <button v-if="wizardStep < 4" type="button" class="paper-btn-primary" @click="wizardStep++">
            下一步
          </button>
          <button v-else type="button" class="paper-btn-primary" :disabled="actionBusy" @click="finishWizard">
            {{ actionBusy ? '提交中…' : '完成并记录投递' }}
          </button>
        </footer>
      </section>
    </template>

    <template v-else-if="screen === 'detail' && activeDetail">
      <nav class="wf-subnav">
        <button type="button" class="ul-btn-link" @click="backToList">← 投递列表</button>
        <span class="wf-subnav-sep">/</span>
        <span>{{ activeDetail.venue_label }}</span>
      </nav>

      <section class="wf-panel sub-detail-panel">
        <header class="sub-detail-panel-head">
          <div class="sub-detail-panel-head-text">
            <h2 class="sub-detail-panel-title">
              {{ activeDetail.venue_label }} · {{ submissionTrackLabel(activeDetail.track) }}
            </h2>
            <p class="sub-detail-panel-lead">
              <span
                class="wf-status-pill wf-status-pill--inline"
                :class="`wf-status-pill--${submissionStatusTone(activeDetail.status)}`"
              >
                {{ STATUS_LABEL[activeDetail.status] }}
              </span>
              <span class="sub-detail-panel-lead-sep">·</span>
              轮次 R{{ activeDetail.round }}
              <span class="sub-detail-panel-lead-sep">·</span>
              稿本 {{ submissionSnapshotLabel(activeDetail.draft_version) }}
            </p>
          </div>
          <button
            type="button"
            class="paper-btn-primary sub-detail-head-action"
            :disabled="actionBusy"
            @click="onDetailSubmitAction"
          >
            {{ actionBusy ? '处理中…' : '投递' }}
          </button>
        </header>

        <div class="ul-tabs" role="tablist" aria-label="投递">
          <button
            v-for="t in [
              { id: 'overview', label: '概览' },
              { id: 'package', label: '投稿包' },
              { id: 'checklist', label: '校验' },
              { id: 'log', label: '投递日志' },
            ]"
            :key="t.id"
            type="button"
            role="tab"
            class="ul-tab"
            :class="{ 'ul-tab--active': detailTab === t.id }"
            :aria-selected="detailTab === t.id"
            @click="detailTab = t.id as typeof detailTab"
          >
            {{ t.label }}
          </button>
        </div>

        <div class="ul-tab-pane" role="tabpanel">
          <template v-if="detailTab === 'overview'">
            <p class="sub-pane-lead">本投递在系统中的状态与 Portal 关联信息。</p>
            <dl class="sub-detail-dl">
              <dt>投递状态</dt>
              <dd>
                <span
                  class="wf-status-pill wf-status-pill--inline"
                  :class="`wf-status-pill--${submissionStatusTone(activeDetail.status)}`"
                >
                  {{ STATUS_LABEL[activeDetail.status] }}
                </span>
              </dd>
              <dt>提交时间</dt>
              <dd>
                {{
                  activeDetail.submitted_at
                    ? new Date(activeDetail.submitted_at).toLocaleString('zh-CN')
                    : '—'
                }}
              </dd>
              <dt>投稿系统</dt>
              <dd>
                <a
                  v-if="activeDetail.portal_url"
                  :href="activeDetail.portal_url"
                  class="sub-detail-link"
                  target="_blank"
                  rel="noopener noreferrer"
                >
                  {{ activeDetail.portal_url }}
                </a>
                <span v-else class="sub-detail-muted">—</span>
              </dd>
              <dt>Submission ID</dt>
              <dd>
                <code v-if="activeDetail.external_id" class="sub-detail-code">{{
                  activeDetail.external_id
                }}</code>
                <span v-else class="sub-detail-muted">—</span>
              </dd>
              <dt v-if="activeDetail.revision_deadline">修回截止</dt>
              <dd v-if="activeDetail.revision_deadline">
                {{ new Date(activeDetail.revision_deadline).toLocaleString('zh-CN') }}
              </dd>
              <dt v-if="activeDetail.decision">审稿结论</dt>
              <dd v-if="activeDetail.decision">{{ activeDetail.decision }}</dd>
            </dl>
          </template>

          <template v-else-if="detailTab === 'package'">
            <p class="sub-pane-lead">绑定稿本对应的投稿包文件；修回轮可在此追加 Response letter。</p>
            <ul class="sub-file-rows">
              <li class="sub-file-row">
                <span class="sub-file-row-icon" aria-hidden="true">PDF</span>
                <div class="sub-file-row-body">
                  <span class="sub-file-row-name">{{
                    activeDetail.main_pdf_filename ||
                      `main_${submissionSnapshotLabel(activeDetail.draft_version)}.pdf`
                  }}</span>
                  <span class="sub-file-row-hint">主文稿 · 绑定 {{ submissionSnapshotLabel(activeDetail.draft_version) }}</span>
                </div>
                <span class="sub-file-row-badge">主 PDF</span>
              </li>
              <li class="sub-file-row sub-file-row--muted">
                <span class="sub-file-row-icon" aria-hidden="true">ZIP</span>
                <div class="sub-file-row-body">
                  <span class="sub-file-row-name">supplementary.zip</span>
                  <span class="sub-file-row-hint">补充材料（占位）</span>
                </div>
                <span class="sub-file-row-badge sub-file-row-badge--muted">待上传</span>
              </li>
            </ul>
          </template>

          <template v-else-if="detailTab === 'checklist'">
            <p class="sub-pane-lead">投递前校验项快照；接入规则引擎后可自动跑检并更新状态。</p>
            <ul class="sub-check-rows">
              <li v-for="c in detailChecklist" :key="c.id" class="sub-check-row">
                <span class="sub-check-row-label">{{ c.label || c.id }}</span>
                <span
                  class="sub-check-row-status"
                  :class="c.ok ? 'sub-check-row-status--ok' : 'sub-check-row-status--warn'"
                >
                  {{ c.ok ? '通过' : '待处理' }}
                </span>
              </li>
            </ul>
          </template>

          <template v-else-if="detailTab === 'log'">
            <p class="sub-pane-lead">按轮次记录我方投稿内容与编辑部/审稿反馈（后续可接入分轮归档与 PDF）。</p>
            <div v-if="submissionRoundLogs.length" class="sub-round-log-stack">
              <article
                v-for="block in submissionRoundLogs"
                :key="block.round"
                class="sub-round-log"
              >
                <h3 class="sub-round-log-title">轮次 R{{ block.round }}</h3>
                <section class="sub-round-log-section sub-round-log-section--out">
                  <h4 class="sub-round-log-section-title">我方投递</h4>
                  <dl class="sub-round-log-dl">
                    <dt>提交时间</dt>
                    <dd>
                      {{
                        block.outgoing.submittedAt
                          ? new Date(block.outgoing.submittedAt).toLocaleString('zh-CN')
                          : '—'
                      }}
                    </dd>
                    <template v-if="block.outgoing.title">
                      <dt>标题</dt>
                      <dd>{{ block.outgoing.title }}</dd>
                    </template>
                    <template v-if="block.outgoing.abstract">
                      <dt>摘要</dt>
                      <dd class="sub-round-log-abstract">{{ block.outgoing.abstract }}</dd>
                    </template>
                    <template v-if="block.outgoing.draftVersion != null">
                      <dt>绑定稿本</dt>
                      <dd>{{ submissionSnapshotLabel(block.outgoing.draftVersion) }}</dd>
                    </template>
                    <template v-if="block.outgoing.pdfFilename">
                      <dt>主 PDF</dt>
                      <dd>{{ block.outgoing.pdfFilename }}</dd>
                    </template>
                    <template v-if="block.outgoing.externalId">
                      <dt>Submission ID</dt>
                      <dd>
                        <code class="sub-detail-code">{{ block.outgoing.externalId }}</code>
                      </dd>
                    </template>
                    <template
                      v-if="
                        !block.outgoing.title &&
                        !block.outgoing.abstract &&
                        block.outgoing.draftVersion == null &&
                        !block.outgoing.pdfFilename
                      "
                    >
                      <dt>说明</dt>
                      <dd class="sub-detail-muted">本轮详细稿本未单独归档，见「投稿包」或后续分轮存储。</dd>
                    </template>
                  </dl>
                </section>
                <section class="sub-round-log-section sub-round-log-section--in">
                  <h4 class="sub-round-log-section-title">对方返回</h4>
                  <template v-if="block.incoming">
                    <dl v-if="!block.incoming.pending" class="sub-round-log-dl">
                      <dt v-if="block.incoming.returnedAt">反馈时间</dt>
                      <dd v-if="block.incoming.returnedAt">
                        {{ new Date(block.incoming.returnedAt).toLocaleString('zh-CN') }}
                      </dd>
                      <dt v-if="block.incoming.decision">结论</dt>
                      <dd v-if="block.incoming.decision">{{ block.incoming.decision }}</dd>
                      <dt v-if="block.incoming.comments">意见</dt>
                      <dd v-if="block.incoming.comments" class="sub-round-log-comments">
                        {{ block.incoming.comments }}
                      </dd>
                    </dl>
                    <p v-else class="sub-round-log-pending">{{ block.incoming.comments }}</p>
                  </template>
                  <p v-else class="sub-detail-muted sub-round-log-empty">本轮暂无审稿或编辑部返回。</p>
                </section>
              </article>
            </div>
            <p v-else class="sub-detail-empty">暂无投递轮次记录。</p>
          </template>
        </div>
      </section>
    </template>

    <template v-else-if="screen === 'revision' && activeDetail">
      <nav class="wf-subnav">
        <button type="button" class="ul-btn-link" @click="openDetail(activeDetail.id)">← 投递</button>
        <span class="wf-subnav-sep">/</span>
        <span>改稿中心</span>
      </nav>
      <section class="wf-panel sub-revision-panel">
        <h3 class="sub-revision-panel-title">{{ activeDetail.venue_label }} · 审稿与修回</h3>
        <label class="sub-field">
          <span class="sub-label">Decision</span>
          <input
            v-model="revisionForm.decision"
            class="sub-input"
            type="text"
            placeholder="Major Revision / Minor / Reject / Accept"
          />
        </label>
        <label class="sub-field">
          <span class="sub-label">意见摘要与改稿要点</span>
          <textarea
            v-model="revisionForm.comments"
            class="sub-textarea"
            rows="6"
            placeholder="逐条 reviewer 意见；接入 Agent 后可自动拆分并派工到综述/实验/写作"
          />
        </label>
        <p class="sub-meta">
          改稿完成后请回到「文献综述 / 实验方案 / 论文写作」更新内容，再绑定新快照发起修回。
        </p>
        <div class="sub-revision-actions">
          <button type="button" class="paper-btn-primary" :disabled="actionBusy" @click="saveRevisionPlan">
            {{ actionBusy ? '保存中…' : '保存改稿计划' }}
          </button>
          <button type="button" class="paper-btn-secondary" :disabled="actionBusy" @click="markResubmitted">
            标记已修回
          </button>
        </div>
      </section>
    </template>
  </div>
</template>

<style scoped>
/* 列表白卡片顶留白与 PaperWorkflowPanels 文献综述列表一致（避免全局层叠未生效） */
.sub-hub-list-panel.wf-panel.wf-lit-list-panel {
  box-sizing: border-box;
  display: flow-root;
  width: 100%;
  padding: 24px 22px 52px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.sub-revision-panel-title {
  margin: 0 0 16px;
  font-size: 16px;
  font-weight: 700;
  color: #1e1b4b;
}

.sub-wizard-step-tabs {
  flex-wrap: nowrap;
  overflow-x: auto;
  -webkit-overflow-scrolling: touch;
}

.sub-wizard-step-tab {
  display: inline-flex;
  gap: 8px;
  align-items: center;
  min-width: 120px;
  white-space: nowrap;
}

.sub-wizard-step-tab:disabled {
  cursor: default;
  opacity: 0.55;
}

.sub-wizard-step-tab--done:not(.ul-tab--active) {
  color: #4338ca;
}

.sub-wizard-step-num {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  font-size: 12px;
  font-weight: 700;
  color: #64748b;
  background: #e2e8f0;
  border-radius: 999px;
}

.ul-tab--active .sub-wizard-step-num {
  color: #fff;
  background: #6366f1;
}

.sub-wizard-step-tab--done .sub-wizard-step-num {
  color: #4338ca;
  background: #e0e7ff;
}

.sub-wizard-pane {
  min-height: 280px;
}

.sub-wizard-fields {
  max-width: 560px;
}

.sub-wizard-fields .sub-field {
  margin-bottom: 18px;
}

.sub-wizard-fields .sub-label {
  margin-bottom: 8px;
}

.sub-wizard-fields--after-summary {
  margin-top: 8px;
}

.sub-field-hint {
  display: block;
  margin-top: 8px;
  font-size: 13px;
  color: #64748b;
}

.sub-check--block {
  padding: 12px 14px;
  margin-bottom: 16px;
  background: #f8fafc;
  border: 1px solid #e8eaf0;
  border-radius: 10px;
}

.sub-input--file {
  padding: 8px 10px;
  font-size: 13px;
}

.sub-check-rows--compact {
  max-width: 560px;
  margin-top: 8px;
}

.sub-wizard-confirm-dl {
  max-width: 560px;
  padding: 16px 18px;
  background: #f8fafc;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
}

.sub-wizard-panel-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: flex-end;
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #e2e8f0;
}

.sub-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 14px;
}

.sub-label {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}

.sub-input,
.sub-textarea {
  padding: 9px 12px;
  font-size: 14px;
  color: #1e293b;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.sub-textarea {
  resize: vertical;
  font-family: inherit;
}

.sub-check {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 14px;
  font-size: 14px;
  color: #334155;
}

.sub-meta {
  margin: 0 0 12px;
  font-size: 13px;
  line-height: 1.55;
  color: #64748b;
}

.sub-checklist {
  padding-left: 0;
  margin: 0;
  list-style: none;
}

.sub-checklist li {
  padding: 8px 0;
  font-size: 14px;
  border-bottom: 1px solid #f1f5f9;
}

.sub-check-ok {
  margin-right: 8px;
  font-size: 12px;
  font-weight: 600;
  color: #047857;
}

.sub-check-warn {
  margin-right: 8px;
  font-size: 12px;
  font-weight: 600;
  color: #b45309;
}

.sub-summary {
  display: grid;
  grid-template-columns: 120px 1fr;
  gap: 8px 16px;
  margin: 0;
  font-size: 14px;
}

.sub-summary dt {
  margin: 0;
  color: #64748b;
}

.sub-summary dd {
  margin: 0;
  color: #1e1b4b;
}

/* 详情 / 向导 / 改稿：外壳用全局 .wf-panel，Tab 用全局 .ul-tabs / .ul-tab */
.sub-detail-panel-head {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 4px;
  padding-bottom: 16px;
}

.sub-detail-head-action {
  flex-shrink: 0;
  align-self: center;
}

.sub-detail-panel-title {
  margin: 0 0 10px;
  font-size: 20px;
  font-weight: 700;
  line-height: 1.35;
  color: #1e1b4b;
}

.sub-detail-panel-lead {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin: 0;
  font-size: 14px;
  line-height: 1.5;
  color: #64748b;
}

.sub-detail-panel-lead-sep {
  opacity: 0.45;
}

.sub-detail-panel .ul-tab-pane,
.sub-wizard-panel .ul-tab-pane {
  min-height: 200px;
  font-size: 14px;
  line-height: 1.6;
  color: #334155;
}

.sub-pane-lead {
  margin: 0 0 20px;
  font-size: 14px;
  line-height: 1.65;
  color: #64748b;
}

.sub-detail-dl {
  display: grid;
  grid-template-columns: 108px 1fr;
  gap: 16px 20px;
  padding: 4px 0 0;
  margin: 0;
}

.sub-detail-dl dt {
  margin: 0;
  font-weight: 600;
  color: #64748b;
}

.sub-detail-dl dd {
  margin: 0;
  color: #1e293b;
  word-break: break-word;
}

.sub-detail-link {
  color: #4338ca;
  text-decoration: none;
  word-break: break-all;
}

.sub-detail-link:hover {
  text-decoration: underline;
}

.sub-detail-code {
  padding: 2px 8px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 13px;
  color: #4338ca;
  word-break: break-all;
  background: #eef2ff;
  border-radius: 6px;
}

.sub-detail-muted {
  color: #94a3b8;
}

.sub-detail-empty {
  margin: 8px 0 0;
  padding: 24px 16px;
  font-size: 14px;
  color: #64748b;
  text-align: center;
  background: #f8fafc;
  border-radius: 12px;
}

.sub-file-rows {
  padding: 0;
  margin: 0;
  list-style: none;
}

.sub-file-row {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  align-items: center;
  padding: 16px 18px;
  margin-bottom: 10px;
  background: #f8fafc;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
}

.sub-file-row--muted {
  opacity: 0.85;
}

.sub-file-row-icon {
  flex-shrink: 0;
  width: 44px;
  height: 44px;
  font-size: 11px;
  font-weight: 700;
  line-height: 44px;
  color: #4338ca;
  text-align: center;
  background: #e0e7ff;
  border-radius: 10px;
}

.sub-file-row-body {
  flex: 1;
  min-width: 160px;
}

.sub-file-row-name {
  display: block;
  font-weight: 600;
  color: #1e1b4b;
}

.sub-file-row-hint {
  display: block;
  margin-top: 4px;
  font-size: 13px;
  color: #64748b;
}

.sub-file-row-badge {
  flex-shrink: 0;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 600;
  color: #4338ca;
  background: #e0e7ff;
  border-radius: 999px;
}

.sub-file-row-badge--muted {
  color: #64748b;
  background: #e2e8f0;
}

.sub-check-rows {
  padding: 0;
  margin: 0;
  list-style: none;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
  overflow: hidden;
}

.sub-check-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  justify-content: space-between;
  padding: 16px 18px;
  font-size: 14px;
  background: #fff;
  border-bottom: 1px solid #f1f5f9;
}

.sub-check-row:last-child {
  border-bottom: none;
}

.sub-check-row-label {
  flex: 1;
  min-width: 200px;
  color: #334155;
}

.sub-check-row-status {
  flex-shrink: 0;
  padding: 4px 12px;
  font-size: 12px;
  font-weight: 600;
  border-radius: 999px;
}

.sub-check-row-status--ok {
  color: #047857;
  background: #d1fae5;
}

.sub-check-row-status--warn {
  color: #b45309;
  background: #fef3c7;
}

.sub-round-log-stack {
  display: flex;
  flex-direction: column;
  gap: 18px;
}

.sub-round-log {
  padding: 18px 20px;
  background: #f8fafc;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
}

.sub-round-log-title {
  margin: 0 0 14px;
  font-size: 15px;
  font-weight: 700;
  color: #1e1b4b;
}

.sub-round-log-section {
  padding-top: 12px;
  margin-top: 12px;
  border-top: 1px solid #e2e8f0;
}

.sub-round-log-section:first-of-type {
  padding-top: 0;
  margin-top: 0;
  border-top: none;
}

.sub-round-log-section-title {
  margin: 0 0 10px;
  font-size: 13px;
  font-weight: 700;
  color: #64748b;
}

.sub-round-log-dl {
  display: grid;
  grid-template-columns: 96px 1fr;
  gap: 10px 16px;
  margin: 0;
  font-size: 14px;
}

.sub-round-log-dl dt {
  margin: 0;
  font-weight: 600;
  color: #64748b;
}

.sub-round-log-dl dd {
  margin: 0;
  line-height: 1.55;
  color: #334155;
  word-break: break-word;
}

.sub-round-log-abstract,
.sub-round-log-comments {
  white-space: pre-wrap;
}

.sub-round-log-pending {
  margin: 0;
  font-size: 14px;
  line-height: 1.55;
  color: #475569;
}

.sub-round-log-empty {
  margin: 0;
  font-size: 14px;
}

.sub-revision-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 8px;
}
</style>
