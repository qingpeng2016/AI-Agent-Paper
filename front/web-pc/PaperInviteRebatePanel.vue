<script setup lang="ts">
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import {
  createInviteRebateApi,
  formatCny,
  parseMoney,
  type InviteCommissionRecord,
  type InvitePayoutConfig,
  type InviteRebateMember,
  type InviteRebateOverview,
  type InviteWithdrawalRecord,
} from '@ai-token-mall/shared'

type PanelTabId = 'details' | 'withdrawals' | 'members' | 'rebates'

const AUTH_TOKEN_COOKIE = 'atm_token'

function getAuthToken(): string | null {
  if (typeof document === 'undefined') return null
  const match = document.cookie.match(new RegExp(`(?:^|;\\s*)${AUTH_TOKEN_COOKIE}=([^;]*)`))
  return match ? decodeURIComponent(match[1]) : null
}

const baseURL = import.meta.env.VITE_API_BASE_URL ?? ''
const inviteRebateApi = createInviteRebateApi({
  baseURL,
  getToken: () => getAuthToken(),
})

const tabs: { id: PanelTabId; label: string }[] = [
  { id: 'details', label: '返佣详情' },
  { id: 'withdrawals', label: '我的佣金' },
  { id: 'members', label: '邀请成员' },
  { id: 'rebates', label: '返利记录' },
]

const activeTab = ref<PanelTabId>('details')
const loaded = ref(false)
const overview = ref<InviteRebateOverview | null>(null)
const commissionAvailable = ref(0)
const inviteMembers = ref<InviteRebateMember[]>([])
const commissionRecords = ref<InviteCommissionRecord[]>([])
const commissionRecordsPage = ref(1)
const commissionRecordsPageSize = 10
const commissionRecordsTotal = ref(0)
const commissionRecordsLoaded = ref(false)
const withdrawalRecords = ref<InviteWithdrawalRecord[]>([])
const withdrawalsPage = ref(1)
const withdrawalsPageSize = 9
const withdrawalsTotal = ref(0)
const withdrawalsLoaded = ref(false)

const inviteRebateDisplayNotes = [
  '返佣比例按「邀请下级人数」自动升级，下级为注册时已绑定到您账号的用户。',
  '受邀用户每笔已支付订单，按实付金额 × 当前返佣比例计算返利，支付成功后即时计入「佣金」。',
  '佣金可划转到余额或者提现。',
]

const payoutQr = reactive({ alipay: '', wechat: '' })
const payoutQrConfigured = reactive({ alipay: false, wechat: false })

const exclusivePromoDomain = computed(() => overview.value?.promo_domain_url?.trim() ?? '')

const withdrawalChannelLabel: Record<'alipay' | 'wechat', string> = {
  alipay: '支付宝',
  wechat: '微信',
}

const withdrawalStatusLabel: Record<string, string> = {
  completed: '已完成',
  pending: '处理中',
  failed: '失败',
}

function authError(e: unknown, fallback: string) {
  const msg = e instanceof Error ? e.message : fallback
  ElMessage.error(msg || fallback)
}

function fallbackCopyText(text: string): boolean {
  try {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.left = '-9999px'
    document.body.appendChild(ta)
    ta.select()
    const ok = document.execCommand('copy')
    document.body.removeChild(ta)
    return ok
  } catch {
    return false
  }
}

async function copyPromoDomain() {
  const value = exclusivePromoDomain.value.trim()
  if (!value) {
    ElMessage.warning('暂无推广域名')
    return
  }
  let copied = false
  if (navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value)
      copied = true
    } catch {
      copied = fallbackCopyText(value)
    }
  } else {
    copied = fallbackCopyText(value)
  }
  if (copied) ElMessage.success('已复制推广域名')
  else ElMessage.warning('复制失败，请手动复制')
}

function applyPayoutConfig(payout: InvitePayoutConfig) {
  payoutQr.alipay = payout.alipay_qr_data_url ?? ''
  payoutQr.wechat = payout.wechat_qr_data_url ?? ''
  payoutQrConfigured.alipay = Boolean(payout.alipay_configured ?? payoutQr.alipay)
  payoutQrConfigured.wechat = Boolean(payout.wechat_configured ?? payoutQr.wechat)
}

async function fetchOverviewBundle() {
  const ov = await inviteRebateApi.overview()
  overview.value = ov
  commissionAvailable.value = parseMoney(ov.commission_balance)
  const payout = await inviteRebateApi.payoutConfig()
  applyPayoutConfig(payout)
  loaded.value = true
}

async function fetchMembers() {
  const data = await inviteRebateApi.members()
  inviteMembers.value = data.items
}

async function fetchCommissionRecords() {
  const data = await inviteRebateApi.commissionRecords({
    page: commissionRecordsPage.value,
    page_size: commissionRecordsPageSize,
  })
  commissionRecords.value = data.items
  commissionRecordsTotal.value = data.total
  commissionRecordsPage.value = data.page
  commissionRecordsLoaded.value = true
}

async function fetchWithdrawals() {
  const data = await inviteRebateApi.withdrawals({
    page: withdrawalsPage.value,
    page_size: withdrawalsPageSize,
  })
  withdrawalRecords.value = data.items
  withdrawalsTotal.value = data.total
  withdrawalsPage.value = data.page
  withdrawalsLoaded.value = true
}

async function loadForTab(tab: PanelTabId) {
  try {
    if (!loaded.value) await fetchOverviewBundle()
    if (tab === 'members') await fetchMembers()
    else if (tab === 'rebates') await fetchCommissionRecords()
    else if (tab === 'withdrawals') await fetchWithdrawals()
  } catch (e) {
    authError(e, '邀请返利加载失败')
  }
}

onMounted(() => {
  void loadForTab(activeTab.value)
})

watch(activeTab, (tab: PanelTabId) => {
  void loadForTab(tab)
})

watch(commissionRecordsPage, () => {
  if (activeTab.value === 'rebates') void fetchCommissionRecords().catch((e) => authError(e, '加载失败'))
})

watch(withdrawalsPage, () => {
  if (activeTab.value === 'withdrawals') void fetchWithdrawals().catch((e) => authError(e, '加载失败'))
})

const transferOpen = ref(false)
const transferAmount = ref('')
const withdrawOpen = ref(false)
const withdrawChannel = ref<'alipay' | 'wechat'>('alipay')
const withdrawAmount = ref('')

function openTransferModal() {
  if (commissionAvailable.value <= 0) {
    ElMessage.warning('暂无可划转佣金')
    return
  }
  transferAmount.value = ''
  transferOpen.value = true
}

async function confirmTransfer() {
  const yuan = Number(transferAmount.value)
  if (!Number.isFinite(yuan) || yuan <= 0) {
    ElMessage.warning('请输入有效的划转金额')
    return
  }
  if (yuan > commissionAvailable.value) {
    ElMessage.warning('划转金额不能超过可提现佣金')
    return
  }
  try {
    const data = await inviteRebateApi.transferCommissionToBalance(yuan.toFixed(2))
    commissionAvailable.value = parseMoney(data.commission_balance)
    transferOpen.value = false
    ElMessage.success('已划转到账户余额')
  } catch (e) {
    authError(e, '划转失败')
  }
}

function openWithdrawModal() {
  if (!payoutQrConfigured.alipay && !payoutQrConfigured.wechat) {
    ElMessage.warning('请先设置支付宝或微信收款码')
    return
  }
  withdrawChannel.value = payoutQrConfigured.alipay ? 'alipay' : 'wechat'
  withdrawAmount.value = ''
  withdrawOpen.value = true
}

async function confirmWithdraw() {
  const yuan = Number(withdrawAmount.value)
  if (!Number.isFinite(yuan) || yuan <= 0) {
    ElMessage.warning('请输入有效的提现金额')
    return
  }
  if (yuan > commissionAvailable.value) {
    ElMessage.warning('提现金额不能超过可提现佣金')
    return
  }
  const channel = withdrawChannel.value
  if (channel === 'alipay' && !payoutQrConfigured.alipay) {
    ElMessage.warning('请先设置支付宝收款码')
    return
  }
  if (channel === 'wechat' && !payoutQrConfigured.wechat) {
    ElMessage.warning('请先设置微信收款码')
    return
  }
  try {
    const data = await inviteRebateApi.createWithdrawal({
      amount: yuan.toFixed(2),
      channel,
    })
    commissionAvailable.value = parseMoney(data.commission_balance)
    withdrawalRecords.value.unshift(data.withdrawal)
    withdrawalsTotal.value += 1
    withdrawalsLoaded.value = true
    withdrawOpen.value = false
    ElMessage.success('提现申请已提交')
  } catch (e) {
    authError(e, '提现申请失败')
  }
}

const payoutModalOpen = ref(false)
const payoutModalChannel = ref<'alipay' | 'wechat'>('alipay')
const payoutFile = ref<File | null>(null)
const payoutPreview = ref('')
let payoutObjectUrl: string | null = null

function clearPayoutFile() {
  if (payoutObjectUrl) {
    URL.revokeObjectURL(payoutObjectUrl)
    payoutObjectUrl = null
  }
  payoutFile.value = null
}

function openPayoutModal(channel: 'alipay' | 'wechat') {
  clearPayoutFile()
  payoutModalChannel.value = channel
  payoutPreview.value = channel === 'alipay' ? payoutQr.alipay : payoutQr.wechat
  payoutModalOpen.value = true
}

function closePayoutModal() {
  payoutModalOpen.value = false
  clearPayoutFile()
}

function onPayoutFileChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  clearPayoutFile()
  if (!file) {
    payoutPreview.value =
      payoutModalChannel.value === 'alipay' ? payoutQr.alipay : payoutQr.wechat
    return
  }
  if (!/^image\/(png|jpeg|webp)$/i.test(file.type)) {
    ElMessage.warning('请上传 PNG、JPEG 或 WebP 图片')
    input.value = ''
    return
  }
  if (file.size > 2 * 1024 * 1024) {
    ElMessage.warning('图片大小不能超过 2MB')
    input.value = ''
    return
  }
  payoutFile.value = file
  payoutObjectUrl = URL.createObjectURL(file)
  payoutPreview.value = payoutObjectUrl
}

async function confirmPayoutQr() {
  if (!payoutFile.value) {
    ElMessage.warning('请选择收款码图片')
    return
  }
  try {
    const cfg = await inviteRebateApi.uploadPayoutQr(payoutModalChannel.value, payoutFile.value)
    applyPayoutConfig(cfg)
    closePayoutModal()
    ElMessage.success('收款码已保存')
  } catch (e) {
    authError(e, '保存收款码失败')
  }
}

defineExpose({
  reload: () => loadForTab(activeTab.value),
})
</script>

<template>
  <section class="pc-panel">
    <div class="pc-tabs" role="tablist" aria-label="邀请返利">
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

    <div class="pc-body">
      <div v-show="activeTab === 'details'" class="pc-pane" role="tabpanel">
        <p class="pc-lead">分享专属推广域名邀请注册；下级消费按等级比例返佣，与会员中心数据一致。</p>

        <div class="pc-field pc-field--block">
          <span class="pc-label">专属推广域名</span>
          <div class="pc-promo-row">
            <code class="pc-promo-code">{{ exclusivePromoDomain || '—' }}</code>
            <button type="button" class="pc-btn-primary pc-btn--compact" @click="copyPromoDomain">
              复制
            </button>
          </div>
        </div>

        <dl class="pc-metrics pc-metrics--wide">
          <div class="pc-metric">
            <dt>您当前的返佣比例</dt>
            <dd>{{ parseMoney(overview?.current_rate_percent ?? 0) }}%</dd>
          </div>
          <div class="pc-metric">
            <dt>累积邀请用户</dt>
            <dd>{{ overview?.valid_invite_count ?? 0 }} 人</dd>
          </div>
          <div class="pc-metric pc-metric--with-action">
            <dt>下级累计消费</dt>
            <dd>{{ formatCny(overview?.invitee_paid_total ?? 0) }}</dd>
          </div>
        </dl>

        <section class="pc-section">
          <h3 class="pc-section-title">等级与比例</h3>
          <div class="pc-table-wrap">
            <table class="pc-table">
              <thead>
                <tr>
                  <th>等级</th>
                  <th>邀请下级（≥）</th>
                  <th>下级累计消费（≥）</th>
                  <th>返佣比例</th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="tier in overview?.tiers ?? []"
                  :key="tier.level_label"
                  :class="{ 'pc-table-row--current': tier.is_current }"
                >
                  <td>
                    <strong>{{ tier.level_label }}</strong>
                    <span v-if="tier.is_current" class="pc-pill pc-pill--current">当前</span>
                  </td>
                  <td>{{ tier.min_valid_invites }} 人</td>
                  <td>{{ formatCny(tier.min_invitee_paid_amount) }}</td>
                  <td>{{ parseMoney(tier.rate_percent) }}%</td>
                </tr>
              </tbody>
            </table>
          </div>
        </section>

        <section class="pc-section pc-section--notes">
          <h3 class="pc-section-title">说明</h3>
          <ul class="pc-notes">
            <li v-for="(note, i) in inviteRebateDisplayNotes" :key="i">{{ note }}</li>
          </ul>
        </section>
      </div>

      <div v-show="activeTab === 'members'" class="pc-pane" role="tabpanel">
        <p class="pc-lead">通过您的推广域名注册并已绑定关系的用户。</p>
        <div v-if="inviteMembers.length" class="pc-table-wrap">
          <table class="pc-table">
            <thead>
              <tr>
                <th>昵称</th>
                <th>邮箱</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="u in inviteMembers" :key="u.id">
                <td>{{ u.nickname || u.email || '—' }}</td>
                <td>{{ u.email || '—' }}</td>
                <td><span class="pc-pill pc-pill--ok">已绑定</span></td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else-if="loaded" class="pc-empty">暂无邀请成员，请分享您的专属推广域名。</p>
      </div>

      <div v-show="activeTab === 'rebates'" class="pc-pane" role="tabpanel">
        <p class="pc-lead">受邀用户完成支付后，返利将自动计入「我的佣金」。</p>
        <div v-if="commissionRecords.length" class="pc-table-wrap">
          <table class="pc-table">
            <thead>
              <tr>
                <th>邀请用户</th>
                <th>商品</th>
                <th>订单金额</th>
                <th>返利金额</th>
                <th>时间</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="r in commissionRecords" :key="r.id">
                <td>
                  <strong>{{ r.invitee_nickname || r.invitee_email || '—' }}</strong>
                  <span v-if="r.invitee_email" class="pc-cell-sub">{{ r.invitee_email }}</span>
                </td>
                <td>{{ r.product_name }}</td>
                <td>{{ formatCny(r.order_amount) }}</td>
                <td class="pc-amount-plus">+{{ formatCny(r.rebate_amount) }}</td>
                <td class="pc-time">{{ r.created_at }}</td>
              </tr>
            </tbody>
          </table>
        </div>
        <p v-else-if="commissionRecordsLoaded" class="pc-empty">暂无返利记录。</p>
        <div v-if="commissionRecordsTotal > commissionRecordsPageSize" class="pc-pagination">
          <el-pagination
            v-model:current-page="commissionRecordsPage"
            :page-size="commissionRecordsPageSize"
            :total="commissionRecordsTotal"
            layout="total, prev, pager, next"
            background
          />
        </div>
      </div>

      <div v-show="activeTab === 'withdrawals'" class="pc-pane" role="tabpanel">
        <p class="pc-lead">佣金可划转到余额或提现至已配置的收款码。</p>

        <dl class="pc-metrics">
          <div class="pc-metric pc-metric--with-action">
            <dt>可提现佣金</dt>
            <dd class="pc-metric-inline">
              <span class="pc-metric-value">{{ formatCny(commissionAvailable) }}</span>
              <button type="button" class="pc-btn-secondary pc-btn--compact" @click="openTransferModal">
                划转
              </button>
              <button type="button" class="pc-btn-recharge pc-btn--compact" @click="openWithdrawModal">
                提现
              </button>
            </dd>
          </div>
        </dl>

        <section class="pc-section">
          <h3 class="pc-section-title">收款方式</h3>
          <ul class="pc-payout-list">
            <li class="pc-payout-item">
              <span class="pc-payout-name">支付宝</span>
              <span class="pc-hint">{{
                payoutQrConfigured.alipay ? '收款码已配置' : '未设置收款码'
              }}</span>
              <button type="button" class="pc-btn-secondary pc-btn--compact" @click="openPayoutModal('alipay')">
                {{ payoutQrConfigured.alipay ? '更换收款码' : '设置收款码' }}
              </button>
            </li>
            <li class="pc-payout-item">
              <span class="pc-payout-name">微信</span>
              <span class="pc-hint">{{
                payoutQrConfigured.wechat ? '收款码已配置' : '未设置收款码'
              }}</span>
              <button type="button" class="pc-btn-secondary pc-btn--compact" @click="openPayoutModal('wechat')">
                {{ payoutQrConfigured.wechat ? '更换收款码' : '设置收款码' }}
              </button>
            </li>
          </ul>
        </section>

        <section class="pc-section">
          <h3 class="pc-section-title">提现记录</h3>
          <div v-if="withdrawalRecords.length" class="pc-table-wrap">
            <table class="pc-table">
              <thead>
                <tr>
                  <th>金额</th>
                  <th>到账方式</th>
                  <th>状态</th>
                  <th>时间</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="w in withdrawalRecords" :key="w.id">
                  <td>{{ formatCny(w.amount) }}</td>
                  <td>{{ withdrawalChannelLabel[w.channel as 'alipay' | 'wechat'] ?? w.channel }}</td>
                  <td>
                    <span
                      class="pc-pill"
                      :class="{
                        'pc-pill--ok': w.status === 'completed',
                        'pc-pill--pending': w.status === 'pending',
                      }"
                    >
                      {{ withdrawalStatusLabel[w.status] ?? w.status }}
                    </span>
                  </td>
                  <td class="pc-time">{{ w.created_at }}</td>
                </tr>
              </tbody>
            </table>
          </div>
          <p v-else-if="withdrawalsLoaded" class="pc-empty">暂无提现记录。</p>
          <div v-if="withdrawalsTotal > withdrawalsPageSize" class="pc-pagination">
            <el-pagination
              v-model:current-page="withdrawalsPage"
              :page-size="withdrawalsPageSize"
              :total="withdrawalsTotal"
              layout="total, prev, pager, next"
              background
            />
          </div>
        </section>
      </div>
    </div>

    <div v-if="transferOpen" class="pc-modal-backdrop" @click.self="transferOpen = false">
      <div class="pc-modal" role="dialog" aria-labelledby="pc-transfer-title">
        <header class="pc-modal-head">
          <h3 id="pc-transfer-title">划转到余额</h3>
          <button type="button" class="pc-modal-close" aria-label="关闭" @click="transferOpen = false">×</button>
        </header>
        <p class="pc-lead pc-lead--tight">将佣金划转到账户余额，可用于购套餐等消费。</p>
        <label class="pc-field">
          <span class="pc-label">金额（元）</span>
          <input v-model="transferAmount" type="number" min="0" step="0.01" class="pc-input" />
        </label>
        <footer class="pc-modal-foot">
          <button type="button" class="pc-btn-secondary" @click="transferOpen = false">取消</button>
          <button type="button" class="pc-btn-primary" @click="confirmTransfer">确认划转</button>
        </footer>
      </div>
    </div>

    <div v-if="withdrawOpen" class="pc-modal-backdrop" @click.self="withdrawOpen = false">
      <div class="pc-modal" role="dialog" aria-labelledby="pc-withdraw-title">
        <header class="pc-modal-head">
          <h3 id="pc-withdraw-title">申请提现</h3>
          <button type="button" class="pc-modal-close" aria-label="关闭" @click="withdrawOpen = false">×</button>
        </header>
        <label class="pc-field">
          <span class="pc-label">到账方式</span>
          <select v-model="withdrawChannel" class="pc-input">
            <option value="alipay" :disabled="!payoutQrConfigured.alipay">支付宝</option>
            <option value="wechat" :disabled="!payoutQrConfigured.wechat">微信</option>
          </select>
        </label>
        <label class="pc-field">
          <span class="pc-label">金额（元）</span>
          <input v-model="withdrawAmount" type="number" min="0" step="0.01" class="pc-input" />
        </label>
        <footer class="pc-modal-foot">
          <button type="button" class="pc-btn-secondary" @click="withdrawOpen = false">取消</button>
          <button type="button" class="pc-btn-primary" @click="confirmWithdraw">提交申请</button>
        </footer>
      </div>
    </div>

    <div v-if="payoutModalOpen" class="pc-modal-backdrop" @click.self="closePayoutModal">
      <div class="pc-modal" role="dialog">
        <header class="pc-modal-head">
          <h3>{{ payoutModalChannel === 'alipay' ? '设置支付宝收款码' : '设置微信收款码' }}</h3>
          <button type="button" class="pc-modal-close" aria-label="关闭" @click="closePayoutModal">×</button>
        </header>
        <label class="pc-field">
          <span class="pc-label">收款码图片</span>
          <input type="file" accept="image/png,image/jpeg,image/webp" class="pc-file" @change="onPayoutFileChange" />
        </label>
        <div v-if="payoutPreview" class="pc-qr-preview">
          <img :src="payoutPreview" alt="收款码预览" />
        </div>
        <footer class="pc-modal-foot">
          <button type="button" class="pc-btn-secondary" @click="closePayoutModal">取消</button>
          <button type="button" class="pc-btn-primary" @click="confirmPayoutQr">保存</button>
        </footer>
      </div>
    </div>
  </section>
</template>

<style scoped>
/* 与个人中心 PaperPersonalCenterPanel 同一套工作台卡片 / Tab / 内容区 */
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
  margin-bottom: 0;
  background: transparent;
  border-bottom: 1px solid #e2e8f0;
}

.pc-tab {
  position: relative;
  flex: 0 1 auto;
  min-width: 96px;
  padding: 12px 18px;
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

.pc-body {
  padding-top: 22px;
}

.pc-pane {
  animation: pc-fade-in 0.12s ease;
}

@keyframes pc-fade-in {
  from {
    opacity: 0.6;
  }
  to {
    opacity: 1;
  }
}

.pc-lead {
  margin: 0 0 20px;
  font-size: 14px;
  line-height: 1.55;
  color: #475569;
}

.pc-lead--tight {
  margin-bottom: 16px;
}

.pc-field {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.pc-field--block {
  margin-bottom: 22px;
}

.pc-label {
  font-size: 13px;
  font-weight: 600;
  color: #475569;
}

.pc-promo-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px 12px;
  align-items: center;
}

.pc-promo-code {
  flex: 1;
  min-width: min(100%, 240px);
  margin: 0;
  padding: 10px 12px;
  font-family: ui-monospace, SFMono-Regular, Menlo, monospace;
  font-size: 13px;
  line-height: 1.45;
  color: #334155;
  word-break: break-all;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}

.pc-metrics {
  display: flex;
  flex-wrap: wrap;
  gap: 0;
  align-items: stretch;
  width: fit-content;
  max-width: 100%;
  margin: 0 0 26px;
  padding: 0;
}

.pc-metrics--wide .pc-metric {
  min-width: 120px;
  max-width: 220px;
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

.pc-metric-value {
  font-size: 20px;
  font-weight: 700;
  line-height: 1.2;
}

.pc-section {
  margin-bottom: 24px;
  padding-bottom: 4px;
  border-bottom: 1px solid #f1f5f9;
}

.pc-section--notes {
  margin-bottom: 0;
  border-bottom: none;
}

.pc-section-title {
  margin: 0 0 12px;
  font-size: 15px;
  font-weight: 700;
  color: #1e293b;
}

.pc-table-wrap {
  overflow-x: auto;
  border: 1px solid #e2e8f0;
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

.pc-table-row--current {
  background: #f8fafc;
}

.pc-pill {
  display: inline-block;
  margin-left: 6px;
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 6px;
}

.pc-pill--current {
  color: #4338ca;
  background: #e0e7ff;
}

.pc-pill--ok {
  color: #166534;
  background: #dcfce7;
}

.pc-pill--pending {
  color: #92400e;
  background: #fef3c7;
}

.pc-notes {
  margin: 0;
  padding-left: 1.2em;
  font-size: 13px;
  line-height: 1.65;
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

.pc-cell-sub {
  display: block;
  margin-top: 2px;
  font-size: 12px;
  font-weight: 400;
  color: #94a3b8;
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

.pc-pagination {
  margin-top: 16px;
}

.pc-hint {
  font-size: 13px;
  font-weight: 400;
  color: #94a3b8;
}

.pc-payout-list {
  margin: 0;
  padding: 0;
  list-style: none;
  border: 1px solid #e8eaf0;
  border-radius: 12px;
  overflow: hidden;
}

.pc-payout-item {
  display: grid;
  grid-template-columns: 72px 1fr auto;
  gap: 12px;
  align-items: center;
  padding: 14px 16px;
  border-bottom: 1px solid #f1f5f9;
}

.pc-payout-item:last-child {
  border-bottom: none;
}

.pc-payout-name {
  font-size: 14px;
  font-weight: 600;
  color: #334155;
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

.pc-btn-secondary {
  padding: 10px 18px;
  font-size: 14px;
  font-weight: 600;
  color: #334155;
  background: #fff;
  border: 1px solid #cbd5e1;
  border-radius: 10px;
  cursor: pointer;
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
  box-shadow: 0 6px 20px rgba(91, 33, 182, 0.22);
}

.pc-btn--compact {
  padding: 8px 16px;
  font-size: 13px;
}

.pc-input {
  padding: 10px 12px;
  font-size: 14px;
  border: 1px solid #cbd5e1;
  border-radius: 8px;
}

.pc-file {
  font-size: 13px;
}

.pc-modal-backdrop {
  position: fixed;
  inset: 0;
  z-index: 2000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgb(15 23 42 / 45%);
}

.pc-modal {
  width: min(420px, 100%);
  padding: 22px 24px;
  background: #fff;
  border: 1px solid #e8eaf0;
  border-radius: 16px;
  box-shadow: 0 4px 24px rgba(30, 27, 75, 0.12);
}

.pc-modal-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 8px;
}

.pc-modal-head h3 {
  margin: 0;
  font-size: 17px;
  font-weight: 700;
  color: #0f172a;
}

.pc-modal-close {
  font-size: 24px;
  line-height: 1;
  color: #94a3b8;
  cursor: pointer;
  background: none;
  border: none;
}

.pc-modal-foot {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 8px;
  padding-top: 16px;
  border-top: 1px solid #e2e8f0;
}

.pc-qr-preview {
  margin-bottom: 12px;
  text-align: center;
}

.pc-qr-preview img {
  max-width: 200px;
  max-height: 200px;
  border-radius: 8px;
}
</style>
