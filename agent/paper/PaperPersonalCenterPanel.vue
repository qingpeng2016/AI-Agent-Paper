<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import PaperOperationLogPanel from './PaperOperationLogPanel.vue'
import PaperSelect from './PaperSelect.vue'
import {
  DISCIPLINE_OPTIONS,
  ENV_PREFERENCE_STORAGE_KEY,
  LITERATURE_SOURCE_OPTIONS,
  getOperationLogs,
  type EnvironmentPreferenceForm,
} from './types'

type PersonalCenterTabId = 'profile' | 'environment' | 'operation-log'

type StoredUserProfile = {
  id?: number
  email?: string | null
  phone?: string | null
  nickname?: string | null
  wallet_balance?: number | string
  created_at?: string
  last_login_at?: string | null
}

const SESSION_USER_STORAGE_KEY = 'atm_user'

const props = defineProps<{
  manuscriptId: string
  manuscriptTitle: string
  envPreference: EnvironmentPreferenceForm
}>()

const emit = defineEmits<{
  environmentSaved: []
}>()

const router = useRouter()
const activeTab = ref<PersonalCenterTabId>('profile')
const envSaving = ref(false)
const operationLogRef = ref<InstanceType<typeof PaperOperationLogPanel> | null>(null)
const profileRefreshTick = ref(0)

const tabs: { id: PersonalCenterTabId; label: string }[] = [
  { id: 'profile', label: '我的信息' },
  { id: 'environment', label: '默认配置' },
  { id: 'operation-log', label: '操作日志' },
]

const sessionUser = computed(() => {
  try {
    const raw = localStorage.getItem(SESSION_USER_STORAGE_KEY)
    if (!raw) return null
    const profile = JSON.parse(raw) as StoredUserProfile
    return profile?.id ? profile : null
  } catch {
    return null
  }
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

/** 演示：套餐额度（接入后来自订阅 API） */
const demoPlanLimit = 50_000
const demoPlanUsed = computed(() => tokenUsage.value.totalUsed)
const demoPlanPercent = computed(() =>
  Math.min(100, Math.round((demoPlanUsed.value / demoPlanLimit) * 100)),
)

const recentTokenMoves = computed(() => {
  void profileRefreshTick.value
  const fromLogs = getOperationLogs()
    .filter((e) => e.status === 'success' && e.tokensTotal > 0)
    .slice(0, 4)
    .map((e) => ({
      id: e.id,
      at: e.occurredAt,
      label: `${e.moduleLabel} · ${e.action}`,
      delta: -e.tokensTotal,
      kind: 'consume' as const,
    }))
  const demoRecharge = {
    id: 'demo-recharge',
    at: new Date(Date.now() - 86400000 * 3).toISOString(),
    label: '余额充值（演示）',
    delta: 10_000,
    kind: 'recharge' as const,
  }
  return [demoRecharge, ...fromLogs].sort(
    (a, b) => new Date(b.at).getTime() - new Date(a.at).getTime(),
  )
})

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

function formatTokens(n: number) {
  return n.toLocaleString('zh-CN')
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

function onRecharge() {
  void router.push({ path: '/member', query: { tab: 'overview' } })
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
  profileRefreshTick.value += 1
}

watch(activeTab, (tab) => {
  if (tab === 'profile') profileRefreshTick.value += 1
})

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
        <ul v-if="recentTokenMoves.length" class="pc-moves-list">
          <li v-for="item in recentTokenMoves" :key="item.id" class="pc-move-row">
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
