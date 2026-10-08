<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  formatSignedCny,
  signedMoneyClass,
  type UserWalletFlowItem,
} from '@ai-agent-paper/shared'
import { userApi } from '@/api'
import { clearSessionUser, getSessionUser, isLoggedIn } from '@/composables/useSessionUser'
import PaperEnvironmentPanel from './PaperEnvironmentPanel.vue'
import PaperOperationLogPanel from './PaperOperationLogPanel.vue'
import { getOperationLogs, type EnvironmentPreferenceForm } from './types'

type StoredUserProfile = {
  id?: number
  email?: string | null
  phone?: string | null
  nickname?: string | null
  wallet_balance?: number | string
  created_at?: string
  last_login_at?: string | null
}

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
const operationLogRef = ref<InstanceType<typeof PaperOperationLogPanel> | null>(null)
const profileRefreshTick = ref(0)

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

const tabs: { id: PersonalCenterTabId; label: string }[] = [
  { id: 'profile', label: '我的信息' },
  { id: 'environment', label: '环境配置' },
  { id: 'wallet-records', label: '资金记录' },
  { id: 'operation-log', label: '操作日志' },
]

const sessionUser = computed((): StoredUserProfile | null => {
  void profileRefreshTick.value
  const u = getSessionUser()
  return u as StoredUserProfile | null
})

function parseWalletBalance(raw: number | string | undefined | null): number {
  if (raw == null || raw === '') return 12_800
  const n = typeof raw === 'number' ? raw : parseFloat(String(raw))
  return Number.isFinite(n) ? n : 12_800
}

const displayName = computed(() => {
  const u = sessionUser.value
  return u?.nickname?.trim() || '科研用户'
})

const accountLabel = computed(() => {
  const u = sessionUser.value
  return u?.email?.trim() || u?.phone?.trim() || '演示账号 · 未登录'
})

const avatarLetter = computed(() => displayName.value.slice(0, 1).toUpperCase())

const walletBalance = computed(() => parseWalletBalance(sessionUser.value?.wallet_balance))

const tokenUsage = computed(() => {
  void profileRefreshTick.value
  const logs = getOperationLogs().filter((e) => e.status === 'success')
  const now = new Date()
  const monthStart = new Date(now.getFullYear(), now.getMonth(), 1).getTime()
  let monthUsed = 0
  let totalUsed = 0
  for (const row of logs) {
    totalUsed += row.tokensTotal
    if (new Date(row.occurredAt).getTime() >= monthStart) {
      monthUsed += row.tokensTotal
    }
  }
  return { monthUsed, totalUsed }
})

const demoPlanLimit = 50_000
const demoPlanUsed = computed(() => tokenUsage.value.totalUsed)
const demoPlanPercent = computed(() =>
  Math.min(100, Math.round((demoPlanUsed.value / demoPlanLimit) * 100)),
)

type TokenMoveRow = {
  id: string
  at: string
  label: string
  delta: number
  kind: 'recharge' | 'consume'
}

const tokenMoveRecords = computed((): TokenMoveRow[] => {
  void profileRefreshTick.value
  const fromLogs = getOperationLogs()
    .filter((e) => e.status === 'success' && e.tokensTotal > 0)
    .map((e) => ({
      id: e.id,
      at: e.occurredAt,
      label: `${e.moduleLabel} · ${e.action}`,
      delta: -e.tokensTotal,
      kind: 'consume' as const,
    }))
  const demoRecharge: TokenMoveRow = {
    id: 'demo-recharge',
    at: new Date(Date.now() - 86400000 * 3).toISOString(),
    label: '余额充值（演示）',
    delta: 10_000,
    kind: 'recharge',
  }
  return [demoRecharge, ...fromLogs].sort(
    (a, b) => new Date(b.at).getTime() - new Date(a.at).getTime(),
  )
})

function formatTokens(n: number) {
  return n.toLocaleString('zh-CN')
}

const router = useRouter()
const logoutLoading = ref(false)

function onRecharge() {
  ElMessage.info('余额充值请前往会员中心（演示）')
}

async function onLogout() {
  if (logoutLoading.value) return
  logoutLoading.value = true
  try {
    try {
      await userApi.logout()
    } catch {
      /* 仍清除本地登录态 */
    }
    clearSessionUser()
    profileRefreshTick.value += 1
    ElMessage.success('已退出登录')
    await router.push('/login')
  } finally {
    logoutLoading.value = false
  }
}

function apiErrorMessage(e: unknown, fallback: string) {
  const msg = e instanceof Error ? e.message : fallback
  if (msg === 'unauthorized') {
    return '未登录或登录已失效，请重新登录后再试'
  }
  return msg || fallback
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

function formatDate(iso: string) {
  try {
    return new Date(iso).toLocaleString('zh-CN', { hour12: false })
  } catch {
    return iso
  }
}

function reloadLogs() {
  operationLogRef.value?.reload()
  profileRefreshTick.value += 1
}

watch(activeTab, (tab) => {
  if (tab === 'profile') profileRefreshTick.value += 1
  if (tab === 'wallet-records') {
    walletFlowsPage.value = 1
    void loadWalletFlows()
  }
})

watch(walletFlowsPage, () => {
  if (activeTab.value === 'wallet-records') void loadWalletFlows()
})

async function reloadFromMenu() {
  profileRefreshTick.value += 1
  if (activeTab.value === 'wallet-records') {
    walletFlowsPage.value = 1
    await loadWalletFlows()
  }
  operationLogRef.value?.reload()
}

defineExpose({ reloadLogs, reloadFromMenu })
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
          <p class="pc-profile-account-row">
            <span class="pc-profile-account">{{ accountLabel }}</span>
            <button
              v-if="isLoggedIn()"
              type="button"
              class="pc-btn-logout"
              :disabled="logoutLoading"
              @click="onLogout"
            >
              {{ logoutLoading ? '退出登录中…' : '退出登录' }}
            </button>
          </p>
        </div>
      </div>

      <dl class="pc-metrics">
        <div class="pc-metric">
          <dt>Token 余额</dt>
          <dd class="pc-metric-inline">
            <span class="pc-metric-value">{{ formatTokens(walletBalance) }}</span>
          </dd>
        </div>
        <div class="pc-metric">
          <dt>本月已用</dt>
          <dd class="pc-metric-inline">
            <span class="pc-metric-value">{{ formatTokens(tokenUsage.monthUsed) }}</span>
          </dd>
        </div>
        <div class="pc-metric pc-metric--with-action">
          <dt>累计消耗</dt>
          <dd class="pc-metric-inline">
            <span class="pc-metric-value">{{ formatTokens(tokenUsage.totalUsed) }}</span>
            <button type="button" class="pc-btn-recharge" @click="onRecharge">充值</button>
          </dd>
        </div>
      </dl>

      <section class="pc-plan-section">
        <h3 class="pc-section-title pc-plan-title">套餐用量</h3>
        <div class="pc-plan-bar">
          <div class="pc-plan-bar-fill" :style="{ width: `${demoPlanPercent}%` }" />
        </div>
        <p class="pc-plan-meta">
          已用 <strong>{{ formatTokens(demoPlanUsed) }}</strong> /
          {{ formatTokens(demoPlanLimit) }} tokens（演示额度）
        </p>
      </section>

      <section class="pc-moves-section">
        <h3 class="pc-section-title">Token 动态</h3>
        <ul v-if="tokenMoveRecords.length" class="pc-moves-list">
          <li v-for="item in tokenMoveRecords" :key="item.id" class="pc-move-row">
            <div class="pc-move-main">
              <span class="pc-move-type" :class="`pc-move-type--${item.kind}`">
                {{ item.kind === 'recharge' ? '充值' : '消耗' }}
              </span>
              <span class="pc-move-label">{{ item.label }}</span>
            </div>
            <div class="pc-move-side">
              <span
                class="pc-move-delta"
                :class="item.delta > 0 ? 'pc-move-delta--plus' : 'pc-move-delta--minus'"
              >
                {{ item.delta > 0 ? '+' : '' }}{{ formatTokens(Math.abs(item.delta)) }}
              </span>
              <time class="pc-move-time">{{ formatDate(item.at) }}</time>
            </div>
          </li>
        </ul>
        <p v-else class="pc-moves-empty">暂无 Token 变动记录</p>
      </section>
    </div>

    <div v-show="activeTab === 'environment'" class="pc-pane" role="tabpanel">
      <PaperEnvironmentPanel
        embedded
        :env-preference="props.envPreference"
        @environment-saved="emit('environmentSaved')"
      />
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

.pc-profile-account-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  align-items: center;
  margin: 0;
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

.pc-btn-recharge {
  padding: 8px 18px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  background: var(--atm-gradient, linear-gradient(135deg, #7c3aed, #6366f1));
  border: none;
  border-radius: 10px;
  cursor: pointer;
  box-shadow: 0 6px 20px rgba(91, 33, 182, 0.28);
  transition:
    transform 0.12s,
    opacity 0.12s;
}

.pc-btn-recharge:hover {
  transform: translateY(-1px);
}

.pc-btn-recharge:disabled {
  opacity: 0.65;
  cursor: not-allowed;
  transform: none;
}

.pc-btn-logout {
  padding: 0;
  font: inherit;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.4;
  color: #0f172a;
  text-decoration: underline;
  text-underline-offset: 3px;
  background: none;
  border: none;
  cursor: pointer;
}

.pc-btn-logout:hover:not(:disabled) {
  color: #334155;
}

.pc-btn-logout:disabled {
  opacity: 0.65;
  cursor: not-allowed;
}

.pc-plan-section {
  margin-bottom: 24px;
  padding: 0 0 4px;
  border-bottom: 1px solid #f1f5f9;
}

.pc-section-title {
  margin: 0;
  font-size: 15px;
  font-weight: 700;
  color: #1e293b;
}

.pc-plan-title {
  margin-bottom: 12px;
}

.pc-plan-bar {
  height: 8px;
  overflow: hidden;
  background: #e2e8f0;
  border-radius: 999px;
}

.pc-plan-bar-fill {
  height: 100%;
  background: #64748b;
  border-radius: 999px;
  transition: width 0.35s ease;
}

.pc-plan-meta {
  margin: 10px 0 0;
  font-size: 13px;
  color: #64748b;
}

.pc-plan-meta strong {
  color: #334155;
}

.pc-moves-section {
  padding-top: 4px;
}

.pc-moves-list {
  margin: 12px 0 0;
  padding: 0;
  list-style: none;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
  overflow: hidden;
}

.pc-move-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
}

.pc-move-row:last-child {
  border-bottom: none;
}

.pc-move-main {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  min-width: 0;
}

.pc-move-type {
  flex-shrink: 0;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 6px;
}

.pc-move-type--recharge {
  color: #166534;
  background: #dcfce7;
}

.pc-move-type--consume {
  color: #5b21b6;
  background: #ede9fe;
}

.pc-move-label {
  font-size: 13px;
  color: #334155;
  line-height: 1.4;
}

.pc-move-side {
  flex-shrink: 0;
  text-align: right;
}

.pc-move-delta {
  display: block;
  font-size: 14px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
}

.pc-move-delta--plus {
  color: #059669;
}

.pc-move-delta--minus {
  color: #4f46e5;
}

.pc-move-time {
  display: block;
  margin-top: 2px;
  font-size: 11px;
  color: #94a3b8;
}

.pc-moves-empty {
  margin: 12px 0 0;
  font-size: 13px;
  color: #94a3b8;
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
