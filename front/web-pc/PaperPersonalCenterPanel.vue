<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  formatCny,
  formatSignedCny,
  signedMoneyClass,
  type UserProfile,
  type UserWalletFlowItem,
} from '@ai-agent-paper/shared'
import { userApi } from '@/api'
import { getSessionUser, isLoggedIn, setSessionUser } from '@/composables/useSessionUser'
import PaperOperationLogPanel from './PaperOperationLogPanel.vue'
import PaperSelect from './PaperSelect.vue'
import {
  DISCIPLINE_OPTIONS,
  ENV_PREFERENCE_STORAGE_KEY,
  LITERATURE_SOURCE_OPTIONS,
  type EnvironmentPreferenceForm,
} from './types'

type PersonalCenterTabId = 'profile' | 'environment' | 'wallet-records' | 'operation-log'

const props = defineProps<{
  manuscriptId: string
  manuscriptTitle: string
  envPreference: EnvironmentPreferenceForm
}>()

const emit = defineEmits<{
  environmentSaved: []
}>()

const activeTab = ref<PersonalCenterTabId>('profile')
const envSaving = ref(false)
const operationLogRef = ref<InstanceType<typeof PaperOperationLogPanel> | null>(null)
const profile = ref<UserProfile | null>(getSessionUser())
const profileLoading = ref(false)

const walletFlows = ref<UserWalletFlowItem[]>([])
const walletFlowsTotal = ref(0)
const walletFlowsPage = ref(1)
const walletFlowsPageSize = 9
const walletFlowsLoaded = ref(false)
const walletFlowsLoading = ref(false)

const walletFlowTypeLabel: Record<string, string> = {
  recharge: '充值',
  pay: '消费',
  refund: '退款',
  commission: '佣金',
  withdraw: '提现',
}

const statusLabel: Record<string, string> = {
  active: '正常',
  disabled: '已禁用',
  banned: '已封禁',
}

const tabs: { id: PersonalCenterTabId; label: string }[] = [
  { id: 'profile', label: '我的信息' },
  { id: 'environment', label: '默认配置' },
  { id: 'wallet-records', label: '资金记录' },
  { id: 'operation-log', label: '操作日志' },
]

const displayName = computed(() => {
  const u = profile.value
  return u?.nickname?.trim() || '科研用户'
})

const accountLabel = computed(() => {
  const u = profile.value
  if (!u) return '未登录'
  return u.email?.trim() || u.phone?.trim() || `用户 #${u.id}`
})

const avatarLetter = computed(() => displayName.value.slice(0, 1).toUpperCase())

function displayField(value: string | null | undefined, fallback = '—') {
  const v = value?.trim()
  return v || fallback
}

function apiErrorMessage(e: unknown, fallback: string) {
  const msg = e instanceof Error ? e.message : fallback
  if (msg === 'unauthorized') {
    return '未登录或登录已失效，请重新登录后再试'
  }
  return msg || fallback
}

async function loadProfile() {
  if (!isLoggedIn()) {
    profile.value = getSessionUser()
    return
  }
  profileLoading.value = true
  try {
    const u = await userApi.me()
    profile.value = u
    setSessionUser(u)
  } catch (e) {
    ElMessage.error(apiErrorMessage(e, '加载账户信息失败'))
    profile.value = getSessionUser()
  } finally {
    profileLoading.value = false
  }
}

async function loadWalletFlows() {
  if (!isLoggedIn()) {
    walletFlows.value = []
    walletFlowsTotal.value = 0
    walletFlowsLoaded.value = true
    return
  }
  walletFlowsLoading.value = true
  try {
    const page = await userApi.walletFlows({
      page: walletFlowsPage.value,
      page_size: walletFlowsPageSize,
    })
    walletFlows.value = page.items ?? []
    walletFlowsTotal.value = page.total ?? 0
    walletFlowsLoaded.value = true
  } catch (e) {
    ElMessage.error(apiErrorMessage(e, '加载资金记录失败'))
    walletFlows.value = []
    walletFlowsTotal.value = 0
    walletFlowsLoaded.value = true
  } finally {
    walletFlowsLoading.value = false
  }
}

function walletFlowTypeText(type: string) {
  return walletFlowTypeLabel[type] ?? type
}

function walletAmountClass(amount: number | string) {
  const key = signedMoneyClass(amount)
  if (key === 'amount-plus') return 'pc-amount-plus'
  if (key === 'amount-minus') return 'pc-amount-minus'
  return 'pc-amount-zero'
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

const disciplineSelectOptions = DISCIPLINE_OPTIONS.map((d) => ({
  value: d.code,
  label: d.label,
}))

function formatDate(iso: string) {
  try {
    return new Date(iso).toLocaleString('zh-CN', { hour12: false })
  } catch {
    return iso
  }
}

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
    ElMessage.success('默认配置已保存（本地演示）')
  } finally {
    envSaving.value = false
  }
}

function reloadLogs() {
  operationLogRef.value?.reload()
}

watch(activeTab, (tab) => {
  if (tab === 'profile') void loadProfile()
  if (tab === 'wallet-records') {
    walletFlowsPage.value = 1
    void loadWalletFlows()
  }
})

watch(walletFlowsPage, () => {
  if (activeTab.value === 'wallet-records') void loadWalletFlows()
})

void loadProfile()

defineExpose({ reloadLogs })
</script>

<template>
  <section class="pc-panel">
    <div class="pc-tabs" role="tablist" aria-label="个人中心">
      <button
        v-for="tab in tabs"
        :key="tab.id"
        type="button"
        role="tab"
        class="pc-tab"
        :class="{ 'pc-tab--active': activeTab === tab.id }"
        :aria-selected="activeTab === tab.id"
        @click="activeTab = tab.id"
      >
        {{ tab.label }}
      </button>
    </div>

    <div v-show="activeTab === 'profile'" class="pc-pane" role="tabpanel">
      <div class="pc-profile-head">
        <div class="pc-avatar" aria-hidden="true">{{ avatarLetter }}</div>
        <div class="pc-profile-id">
          <h2 class="pc-profile-name">{{ displayName }}</h2>
          <p class="pc-profile-account">{{ accountLabel }}</p>
        </div>
        <button
          type="button"
          class="pc-btn-secondary pc-btn--compact pc-profile-refresh"
          :disabled="profileLoading"
          @click="loadProfile"
        >
          {{ profileLoading ? '刷新中…' : '刷新' }}
        </button>
      </div>

      <p v-if="!isLoggedIn()" class="pc-lead">登录后可查看完整账户信息与余额。</p>

      <dl v-else class="pc-metrics">
        <div class="pc-metric">
          <dt>钱包余额</dt>
          <dd class="pc-metric-value">{{ formatCny(profile?.wallet_balance ?? 0) }}</dd>
        </div>
        <div class="pc-metric">
          <dt>佣金余额</dt>
          <dd class="pc-metric-value">{{ formatCny(profile?.commission_balance ?? 0) }}</dd>
        </div>
      </dl>

      <dl class="pc-info-grid">
        <div class="pc-info-item">
          <dt>用户 ID</dt>
          <dd>{{ profile?.id ?? '—' }}</dd>
        </div>
        <div class="pc-info-item">
          <dt>昵称</dt>
          <dd>{{ displayField(profile?.nickname ?? undefined) }}</dd>
        </div>
        <div class="pc-info-item">
          <dt>邮箱</dt>
          <dd>{{ displayField(profile?.email ?? undefined) }}</dd>
        </div>
        <div class="pc-info-item">
          <dt>手机号</dt>
          <dd>{{ displayField(profile?.phone ?? undefined) }}</dd>
        </div>
        <div class="pc-info-item">
          <dt>账号状态</dt>
          <dd>{{ statusLabel[profile?.status ?? ''] ?? profile?.status ?? '—' }}</dd>
        </div>
        <div class="pc-info-item">
          <dt>注册时间</dt>
          <dd>{{ profile?.created_at ? formatDate(profile.created_at) : '—' }}</dd>
        </div>
        <div class="pc-info-item">
          <dt>最后登录</dt>
          <dd>{{ profile?.last_login_at ? formatDate(profile.last_login_at) : '—' }}</dd>
        </div>
      </dl>
    </div>

    <div v-show="activeTab === 'environment'" class="pc-pane pc-pane--env" role="tabpanel">
      <p class="pc-lead">新建工作流时的默认科研偏好与文献策略；保存后对后续「选题发现」等模块预填生效。</p>

      <div class="pc-field-grid">
        <label class="pc-field">
          <span class="pc-label">默认学科</span>
          <PaperSelect v-model="envPreference.disciplineCode" :options="disciplineSelectOptions" />
        </label>

        <label class="pc-field">
          <span class="pc-label">默认目标会议/期刊</span>
          <input v-model="envPreference.defaultVenueText" type="text" class="pc-input" />
        </label>

        <label class="pc-field">
          <span class="pc-label">默认执行强度</span>
          <PaperSelect v-model="envPreference.intensity" :options="intensityOptions" />
        </label>

        <label class="pc-field">
          <span class="pc-label">默认审计等级</span>
          <PaperSelect v-model="envPreference.auditLevel" :options="auditOptions" />
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
          {{ envSaving ? '保存中…' : '保存默认配置' }}
        </button>
      </div>
    </div>

    <div v-show="activeTab === 'wallet-records'" class="pc-pane" role="tabpanel">
      <p class="pc-lead">账户资金变动明细（充值、消费、退款、佣金划入、提现等）。</p>

      <p v-if="!isLoggedIn()" class="pc-empty">请先登录后查看资金记录。</p>
      <p v-else-if="walletFlowsLoading && !walletFlows.length" class="pc-empty">加载中…</p>
      <div v-else-if="walletFlows.length" class="pc-table-wrap">
        <table class="pc-table">
          <thead>
            <tr>
              <th>类型</th>
              <th>金额</th>
              <th>备注</th>
              <th>时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in walletFlows" :key="row.id">
              <td>{{ walletFlowTypeText(row.type) }}</td>
              <td :class="walletAmountClass(row.amount)">{{ formatSignedCny(row.amount) }}</td>
              <td>{{ row.remark?.trim() || '—' }}</td>
              <td class="pc-time">{{ row.created_at }}</td>
            </tr>
          </tbody>
        </table>
      </div>
      <p v-else-if="walletFlowsLoaded" class="pc-empty">暂无资金记录</p>

      <div
        v-if="isLoggedIn() && walletFlowsTotal > walletFlowsPageSize"
        class="pc-pagination"
      >
        <el-pagination
          v-model:current-page="walletFlowsPage"
          :page-size="walletFlowsPageSize"
          :total="walletFlowsTotal"
          layout="total, prev, pager, next"
          background
        />
      </div>
    </div>

    <div v-show="activeTab === 'operation-log'" class="pc-pane" role="tabpanel">
      <PaperOperationLogPanel
        ref="operationLogRef"
        embedded
        list-only
        :manuscript-id="manuscriptId"
        :manuscript-title="manuscriptTitle"
      />
    </div>
  </section>
</template>

<style scoped>
.pc-panel {
  padding: 20px 26px 28px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.06);
}

.pc-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
  padding: 0 2px;
  margin-bottom: 24px;
  background: transparent;
  border-bottom: 1px solid #e2e8f0;
}

.pc-tab {
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

.pc-tab:hover:not(.pc-tab--active) {
  color: #334155;
  background: #f8fafc;
}

.pc-tab--active {
  font-weight: 600;
  color: #1e293b;
  background: #fff;
  border-bottom-color: #6366f1;
  box-shadow: inset 0 -1px 0 #fff;
}

.pc-profile-head {
  display: flex;
  flex-wrap: wrap;
  gap: 14px 16px;
  align-items: center;
  margin-bottom: 18px;
}

.pc-profile-refresh {
  margin-left: auto;
}

.pc-avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 48px;
  height: 48px;
  font-size: 18px;
  font-weight: 700;
  color: #475569;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.pc-profile-name {
  margin: 0 0 4px;
  font-size: 18px;
  font-weight: 700;
  color: #0f172a;
}

.pc-profile-account {
  margin: 0;
  font-size: 13px;
  color: #64748b;
}

.pc-info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 14px 24px;
  margin: 0;
  padding: 20px 0 0;
  border-top: 1px solid #f1f5f9;
}

.pc-info-item dt {
  margin: 0 0 4px;
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
}

.pc-info-item dd {
  margin: 0;
  font-size: 14px;
  font-weight: 500;
  color: #1e293b;
  word-break: break-all;
}

.pc-btn-secondary {
  padding: 8px 16px;
  font-size: 13px;
  font-weight: 600;
  color: #334155;
  background: #fff;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  cursor: pointer;
}

.pc-btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.pc-btn--compact {
  padding: 7px 14px;
  font-size: 13px;
}

.pc-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
  align-items: stretch;
  width: fit-content;
  max-width: 100%;
  margin: 0 0 22px;
  padding: 0;
}

.pc-metric {
  position: relative;
  display: flex;
  flex: 0 1 auto;
  flex-direction: column;
  min-width: 108px;
  padding: 0 20px;
}

.pc-metric:first-child {
  padding-left: 0;
}

.pc-metric:nth-child(2) {
  padding-right: 36px;
}

.pc-metric:not(:last-child)::after {
  position: absolute;
  top: 50%;
  right: 0;
  width: 1px;
  height: 28px;
  content: '';
  background: #e2e8f0;
  transform: translateY(-50%);
}

.pc-metric dt {
  margin: 0 0 4px;
  font-size: 12px;
  font-weight: 500;
  color: #94a3b8;
}

.pc-metric dd {
  margin: 0;
  margin-top: auto;
  font-size: 20px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: #0f172a;
  line-height: 1.2;
}

.pc-metric-inline {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  align-items: flex-end;
}

.pc-metric:nth-child(3) .pc-metric-inline {
  gap: 8px 28px;
}

.pc-metric-value {
  font-size: 20px;
  font-weight: 700;
  line-height: 1.2;
}

.pc-table-wrap {
  overflow-x: auto;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
}

.pc-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.pc-table th,
.pc-table td {
  padding: 11px 14px;
  text-align: left;
  border-bottom: 1px solid #eef2f6;
  vertical-align: top;
}

.pc-table th {
  font-size: 12px;
  font-weight: 600;
  color: #64748b;
  white-space: nowrap;
  background: #f8fafc;
}

.pc-table tbody tr:hover {
  background: #fafbff;
}

.pc-table tbody tr:last-child td {
  border-bottom: none;
}

.pc-time {
  font-size: 12px;
  color: #64748b;
  white-space: nowrap;
}

.pc-amount-plus {
  font-weight: 600;
  color: #059669;
  font-variant-numeric: tabular-nums;
}

.pc-amount-minus {
  font-weight: 600;
  color: #dc2626;
  font-variant-numeric: tabular-nums;
}

.pc-amount-zero {
  font-variant-numeric: tabular-nums;
  color: #64748b;
}

.pc-empty {
  padding: 28px 16px;
  font-size: 14px;
  color: #64748b;
  text-align: center;
  background: #f8fafc;
  border-radius: 12px;
}

.pc-pagination {
  margin-top: 16px;
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
