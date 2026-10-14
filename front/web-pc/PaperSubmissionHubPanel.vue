<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { formatManuscriptDraftVersion, type ManuscriptSubmissionStatus } from '@/api/manuscripts'
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
const detailTab = ref<'overview' | 'package' | 'checklist' | 'meta' | 'log'>('overview')
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

const detailChecklist = computed(() => activeDetail.value?.checklist ?? wizardChecklistItems)

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

async function openDetail(id: string) {
  const detail = await loadDetail(id)
  if (!detail) return
  activeDetail.value = detail
  detailTab.value = 'overview'
  screen.value = 'detail'
}

async function openRevision(id: string) {
  const detail = await loadDetail(id)
  if (!detail) return
  activeDetail.value = detail
  revisionForm.decision = detail.decision ?? ''
  revisionForm.comments = detail.comments_summary ?? ''
  screen.value = 'revision'
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
</script>

<template>
  <section class="sub-hub">
    <header v-if="screen === 'list'" class="sub-hub-head">
      <div>
        <h2 class="sub-hub-title">论文投稿</h2>
        <p class="sub-hub-lead">
          当前论文「<strong>{{ manuscriptTitle || '未命名' }}</strong>」· 内链版本
          {{ formatManuscriptDraftVersion(draftVersion) }}
        </p>
      </div>
      <button type="button" class="paper-btn-primary" :disabled="!activeManuscriptId" @click="openWizard">
        新建投递
      </button>
    </header>

    <p v-if="!activeManuscriptId" class="sub-hub-empty">请先在顶栏或「我的论文」选择当前论文。</p>

    <!-- 列表 -->
    <template v-else-if="screen === 'list'">
      <p v-if="listLoading" class="sub-hub-empty-card">正在加载投递列表…</p>
      <p v-else-if="!listLoading && submissions.length === 0" class="sub-hub-empty-card">
        暂无对外投递记录。完成写作并冻结稿本后，可「新建投递」选择 venue、绑定 PDF 并标记已在 portal 提交。
      </p>
      <table v-else-if="!listLoading && submissions.length > 0" class="sub-table">
        <colgroup>
          <col class="sub-col-venue" />
          <col class="sub-col-round" />
          <col class="sub-col-status" />
          <col class="sub-col-snap" />
          <col class="sub-col-date" />
          <col class="sub-col-actions" />
        </colgroup>
        <thead>
          <tr>
            <th>Venue</th>
            <th>轮次</th>
            <th>状态</th>
            <th>稿本</th>
            <th>提交时间</th>
            <th class="ul-col-actions">
              <PaperTableActionsHead tip="查看投递详情或进入改稿中心" />
            </th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="row in submissions" :key="row.id">
            <td>{{ row.venue_label }} · {{ submissionTrackLabel(row.track) }}</td>
            <td>R{{ row.round }}</td>
            <td>
              <span class="sub-pill" :class="`sub-pill--${row.status}`">{{
                STATUS_LABEL[row.status]
              }}</span>
            </td>
            <td>{{ submissionSnapshotLabel(row.draft_version) }}</td>
            <td class="sub-col-date">
              {{ row.submitted_at ? new Date(row.submitted_at).toLocaleDateString('zh-CN') : '—' }}
            </td>
            <td class="ul-col-actions">
              <button type="button" class="ul-btn-link" @click="openDetail(row.id)">详情</button>
              <button type="button" class="ul-btn-link" @click="openRevision(row.id)">改稿</button>
            </td>
          </tr>
        </tbody>
      </table>
    </template>

    <!-- 向导 -->
    <template v-else-if="screen === 'wizard'">
      <nav class="sub-breadcrumb">
        <button type="button" class="ul-btn-link" @click="backToList">← 投递列表</button>
        <span class="sub-breadcrumb-sep">/</span>
        <span>新建投递 · 步骤 {{ wizardStep }}/4</span>
      </nav>
      <div class="sub-wizard-steps">
        <span :class="{ 'sub-wizard-step--on': wizardStep >= 1 }">1 投稿目标</span>
        <span :class="{ 'sub-wizard-step--on': wizardStep >= 2 }">2 稿本绑定</span>
        <span :class="{ 'sub-wizard-step--on': wizardStep >= 3 }">3 投稿材料</span>
        <span :class="{ 'sub-wizard-step--on': wizardStep >= 4 }">4 确认投递</span>
      </div>

      <div v-show="wizardStep === 1" class="sub-card">
        <h3 class="sub-card-title">目标 venue 与轮次</h3>
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
          <input v-model="wizardForm.portalUrl" class="sub-input" type="url" placeholder="OpenReview / ScholarOne URL" />
        </label>
      </div>

      <div v-show="wizardStep === 2" class="sub-card">
        <h3 class="sub-card-title">稿本与投稿包</h3>
        <label class="sub-field">
          <span class="sub-label">绑定写作快照</span>
          <PaperSelect v-model="wizardForm.snapshot" :options="SNAPSHOT_OPTIONS" placeholder="选择版本" />
        </label>
        <label class="sub-check">
          <input v-model="wizardForm.anonymous" type="checkbox" />
          双盲投稿（校验 PDF 不含作者信息）
        </label>
        <label class="sub-field">
          <span class="sub-label">主 PDF</span>
          <input class="sub-input" type="file" accept=".pdf" @change="onWizardPdfChange" />
          <span v-if="wizardForm.mainPdfName" class="sub-meta">{{ wizardForm.mainPdfName }}</span>
        </label>
        <ul class="sub-checklist">
          <li v-for="c in wizardChecklistItems" :key="c.id">
            <span :class="c.ok ? 'sub-check-ok' : 'sub-check-warn'">{{ c.ok ? '通过' : '待处理' }}</span>
            {{ c.label }}
          </li>
        </ul>
      </div>

      <div v-show="wizardStep === 3" class="sub-card">
        <h3 class="sub-card-title">元数据与声明</h3>
        <label class="sub-field">
          <span class="sub-label">标题</span>
          <input v-model="wizardForm.title" class="sub-input" type="text" />
        </label>
        <label class="sub-field">
          <span class="sub-label">摘要</span>
          <textarea v-model="wizardForm.abstract" class="sub-textarea" rows="4" placeholder="可与写作快照不一致时在此覆盖" />
        </label>
      </div>

      <div v-show="wizardStep === 4" class="sub-card">
        <h3 class="sub-card-title">确认并标记已提交</h3>
        <p class="sub-meta">
          本系统不替代 OpenReview：请先在 portal 完成提交，再在此记录状态并同步「我的论文」投递列。
        </p>
        <dl class="sub-summary">
          <dt>Venue</dt>
          <dd>{{ VENUE_OPTIONS.find((v) => v.value === wizardForm.venueCode)?.label || '—' }}</dd>
          <dt>稿本</dt>
          <dd>v{{ wizardForm.snapshot }}</dd>
          <dt>PDF</dt>
          <dd>{{ wizardForm.mainPdfName || '未选择' }}</dd>
        </dl>
        <label class="sub-field">
          <span class="sub-label">Portal Submission ID（可选）</span>
          <input v-model="wizardForm.externalId" class="sub-input" type="text" placeholder="OpenReview forum id" />
        </label>
        <label class="sub-check">
          <input v-model="wizardForm.markedSubmitted" type="checkbox" />
          我已在投稿系统完成本次提交
        </label>
      </div>

      <footer class="sub-wizard-footer">
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
    </template>

    <!-- 详情 -->
    <template v-else-if="screen === 'detail' && activeDetail">
      <nav class="sub-breadcrumb">
        <button type="button" class="ul-btn-link" @click="backToList">← 投递列表</button>
        <span class="sub-breadcrumb-sep">/</span>
        <span>{{ activeDetail.venue_label }}</span>
      </nav>
      <header class="sub-detail-head">
        <div>
          <h2 class="sub-hub-title">
            {{ activeDetail.venue_label }} · {{ submissionTrackLabel(activeDetail.track) }}
          </h2>
          <p class="sub-hub-lead">
            R{{ activeDetail.round }} · {{ STATUS_LABEL[activeDetail.status] }} · 稿本
            {{ submissionSnapshotLabel(activeDetail.draft_version) }}
          </p>
        </div>
        <button type="button" class="paper-btn-secondary" @click="openRevision(activeDetail.id)">
          改稿中心
        </button>
      </header>
      <nav class="sub-tabs" role="tablist">
        <button
          v-for="t in [
            { id: 'overview', label: '概览' },
            { id: 'package', label: '投稿包' },
            { id: 'checklist', label: '校验' },
            { id: 'meta', label: '元数据' },
            { id: 'log', label: '日志' },
          ]"
          :key="t.id"
          type="button"
          class="sub-tab"
          :class="{ 'sub-tab--active': detailTab === t.id }"
          @click="detailTab = t.id as typeof detailTab"
        >
          {{ t.label }}
        </button>
      </nav>
      <div class="sub-card">
        <template v-if="detailTab === 'overview'">
          <dl class="sub-summary">
            <dt>状态</dt>
            <dd>{{ STATUS_LABEL[activeDetail.status] }}</dd>
            <dt>提交时间</dt>
            <dd>
              {{
                activeDetail.submitted_at
                  ? new Date(activeDetail.submitted_at).toLocaleString('zh-CN')
                  : '—'
              }}
            </dd>
            <dt>Portal</dt>
            <dd>{{ activeDetail.portal_url || '—' }}</dd>
            <dt>External ID</dt>
            <dd>{{ activeDetail.external_id || '—' }}</dd>
            <dt v-if="activeDetail.revision_deadline">修回截止</dt>
            <dd v-if="activeDetail.revision_deadline">
              {{ new Date(activeDetail.revision_deadline).toLocaleString('zh-CN') }}
            </dd>
          </dl>
        </template>
        <template v-else-if="detailTab === 'package'">
          <p class="sub-meta">主 PDF、补充材料、Response letter（修回轮）将在此列出并可重新打包下载。</p>
          <ul class="sub-file-list">
            <li>
              {{
                activeDetail.main_pdf_filename ||
                  `main_${submissionSnapshotLabel(activeDetail.draft_version)}.pdf`
              }}
            </li>
            <li>supplementary.zip（占位）</li>
          </ul>
        </template>
        <template v-else-if="detailTab === 'checklist'">
          <ul class="sub-checklist">
            <li v-for="c in detailChecklist" :key="c.id">
              <span :class="c.ok ? 'sub-check-ok' : 'sub-check-warn'">{{ c.ok ? '通过' : '待处理' }}</span>
              {{ c.label }}
            </li>
          </ul>
        </template>
        <template v-else-if="detailTab === 'meta'">
          <p class="sub-meta">提交时锁定的标题与摘要。</p>
          <p><strong>{{ activeDetail.title || manuscriptTitle }}</strong></p>
          <p v-if="activeDetail.abstract" class="sub-meta">{{ activeDetail.abstract }}</p>
        </template>
        <template v-else>
          <ul v-if="activeDetail.events?.length" class="sub-log">
            <li v-for="ev in activeDetail.events" :key="ev.id">
              {{ new Date(ev.created_at).toLocaleString('zh-CN') }} · {{ ev.message }}
            </li>
          </ul>
          <p v-else class="sub-meta">暂无日志。</p>
        </template>
      </div>
    </template>

    <!-- 改稿 -->
    <template v-else-if="screen === 'revision' && activeDetail">
      <nav class="sub-breadcrumb">
        <button type="button" class="ul-btn-link" @click="openDetail(activeDetail.id)">← 投递详情</button>
        <span class="sub-breadcrumb-sep">/</span>
        <span>改稿中心</span>
      </nav>
      <div class="sub-card">
        <h3 class="sub-card-title">{{ activeDetail.venue_label }} · 审稿与修回</h3>
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
      </div>
    </template>
  </section>
</template>

<style scoped>
.sub-hub {
  padding: 20px 22px 28px;
}

.sub-hub-head {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 20px;
}

.sub-hub-title {
  margin: 0 0 6px;
  font-size: 20px;
  font-weight: 700;
  color: #1e1b4b;
}

.sub-hub-lead {
  margin: 0;
  font-size: 14px;
  line-height: 1.55;
  color: #64748b;
}

.sub-hub-empty,
.sub-hub-empty-card {
  margin: 0;
  padding: 28px 20px;
  font-size: 14px;
  line-height: 1.55;
  color: #64748b;
  text-align: center;
  background: #f8fafc;
  border-radius: 12px;
}

.sub-table {
  width: 100%;
  table-layout: fixed;
  font-size: 14px;
  border-collapse: collapse;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
  overflow: hidden;
}

.sub-table col.sub-col-venue {
  width: 28%;
}

.sub-table col.sub-col-round {
  width: 8%;
}

.sub-table col.sub-col-status {
  width: 12%;
}

.sub-table col.sub-col-snap {
  width: 10%;
}

.sub-table col.sub-col-date {
  width: 14%;
}

.sub-table col.sub-col-actions {
  width: 18%;
}

.sub-table th,
.sub-table td {
  padding: 12px;
  text-align: left;
  border-bottom: 1px solid #e2e8f0;
}

.sub-table th {
  font-weight: 700;
  color: #64748b;
}

.sub-pill {
  display: inline-block;
  padding: 3px 10px;
  font-size: 12px;
  font-weight: 600;
  border-radius: 999px;
}

.sub-pill--submitted,
.sub-pill--under_review,
.sub-pill--resubmitted {
  color: #4338ca;
  background: #e0e7ff;
}

.sub-pill--revision,
.sub-pill--rebuttal {
  color: #b45309;
  background: #fef3c7;
}

.sub-pill--accepted {
  color: #047857;
  background: #d1fae5;
}

.sub-pill--rejected {
  color: #64748b;
  background: #f1f5f9;
}

.sub-breadcrumb {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  margin-bottom: 16px;
  font-size: 14px;
  color: #64748b;
}

.sub-breadcrumb-sep {
  opacity: 0.5;
}

.sub-wizard-steps {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-bottom: 16px;
  font-size: 13px;
  font-weight: 600;
  color: #94a3b8;
}

.sub-wizard-step--on {
  color: #4338ca;
}

.sub-card {
  padding: 22px 24px;
  margin-bottom: 16px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 14px;
  box-shadow: 0 4px 20px rgba(30, 27, 75, 0.05);
}

.sub-card-title {
  margin: 0 0 16px;
  font-size: 16px;
  font-weight: 700;
  color: #1e1b4b;
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

.sub-wizard-footer {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  justify-content: flex-end;
}

.sub-detail-head {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: flex-start;
  justify-content: space-between;
  margin-bottom: 12px;
}

.sub-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  margin-bottom: 12px;
}

.sub-tab {
  padding: 8px 14px;
  font-size: 13px;
  font-weight: 600;
  color: #64748b;
  cursor: pointer;
  background: transparent;
  border: none;
  border-bottom: 2px solid transparent;
}

.sub-tab--active {
  color: #4338ca;
  border-bottom-color: #6366f1;
}

.sub-file-list,
.sub-log {
  margin: 0;
  padding-left: 18px;
  font-size: 14px;
  color: #334155;
}

.sub-revision-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 8px;
}
</style>
