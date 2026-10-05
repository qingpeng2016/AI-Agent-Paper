<script setup lang="ts">
import { computed, nextTick, onBeforeMount, reactive, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import {
  formatCny,
  formatSignedCny,
  parseMoney,
  signedMoneyClass,
  type OrderType,
  type UserWalletFlowItem,
} from '@ai-token-mall/shared'
import CatalogPickerModal from '@/components/catalog/CatalogPickerModal.vue'
import PurchaseModal from '@/components/checkout/PurchaseModal.vue'
import ChangePasswordModal from '@/components/member/ChangePasswordModal.vue'
import InvoiceConfigModal from '@/components/member/InvoiceConfigModal.vue'
import MemberMessagesPanel from '@/components/member/MemberMessagesPanel.vue'
import MemberSidebar from '@/components/member/MemberSidebar.vue'
import type { CatalogProduct } from '@/mocks/home'
import {
  reloadCatalogProducts,
  useCatalogProducts,
} from '@/composables/useCatalogProducts'
import {
  beginRouteNavigationLoading,
  endRouteNavigationLoading,
} from '@/composables/useRouteNavigationLoading'
import {
  clearSessionUser,
  getSessionUser,
  isLoggedIn,
  setSessionUser,
  userAccountLabel,
} from '@/composables/useSessionUser'
import { apiKeyApi, couponApi, inviteRebateApi, invoiceApi, orderApi, subscriptionApi, userApi } from '@/api'
import type {
  ApiTeamAddableInviteeItem,
  ApiTeamMemberItem,
  InviteCommissionRecord,
  InviteRebateMember,
  InviteRebateOverview,
  InviteWithdrawalRecord,
  UserAPIKeyItem,
  UserInvoiceItem,
  UserCouponItem,
  UserOrderItem,
  UserSubscriptionItem,
} from '@ai-token-mall/shared'
import {
  formatTokens,
  memberNav,
  orderStatusLabel,
  withdrawalChannelLabel,
  withdrawalStatusLabel,
  type MemberTab,
} from '@/mocks/member'

const router = useRouter()
const route = useRoute()
const user = ref(getSessionUser())
const { categories: catalogCategories, products: catalogProducts } = useCatalogProducts()
const catalogOpen = ref(false)
/** 非空时目录弹窗只展示该 products_category_id（升档） */
const catalogProductsCategoryId = ref<number | null>(null)
const catalogUpgradeBaselineLimitTokens = ref<number | null>(null)
const purchaseOpen = ref(false)
const purchaseProduct = ref<CatalogProduct | null>(null)
const purchaseOrderType = ref<OrderType>('purchase')
const purchaseUserSubscriptionId = ref(0)
const purchaseSubscriptionPeriodEnd = ref('')
const purchaseSubscriptionExpiresAt = ref('')
const memberSubscriptions = ref<UserSubscriptionItem[]>([])
const recentOrders = ref<UserOrderItem[]>([])
const memberOrders = ref<UserOrderItem[]>([])
const ordersPage = ref(1)
const ordersPageSize = 9
const ordersTotal = ref(0)
const walletFlows = ref<UserWalletFlowItem[]>([])
const walletFlowsPage = ref(1)
const walletFlowsPageSize = 9
const walletFlowsTotal = ref(0)
const memberInvoices = ref<UserInvoiceItem[]>([])
const invoicesPage = ref(1)
const invoicesPageSize = 9
const invoicesTotal = ref(0)
const invoicesLoaded = ref(false)
const memberCoupons = ref<UserCouponItem[]>([])
const couponsLoaded = ref(false)
const invoiceConfigOpen = ref(false)
const changePasswordOpen = ref(false)
const walletBalance = ref(0)
const plansLoaded = ref(false)
const ordersLoaded = ref(false)
const recentOrdersLoaded = ref(false)
const walletFlowsLoaded = ref(false)
/** 各侧栏 Tab 仅首次进入时拉取接口，关闭弹窗等不重复请求 */
const tabLoadedOnce = ref<Partial<Record<MemberTab | 'profile' | 'recentOrders', boolean>>>({})

const subscriptionStatusLabel: Record<string, string> = {
  active: '使用中',
  expired: '已过期',
  suspended: '已暂停',
  cancelled: '已取消',
}
const teamInviteOpen = ref(false)
const apiKeyPanelTab = ref<'mine' | 'team' | 'group'>('mine')
const mainApiKeys = ref<UserAPIKeyItem[]>([])
const teamApiKeys = ref<UserAPIKeyItem[]>([])
const teamMembers = ref<ApiTeamMemberItem[]>([])
const addableInvitees = ref<ApiTeamAddableInviteeItem[]>([])
const apiKeysLoaded = ref(false)
const addTeamMemberOpen = ref(false)
const addTeamMemberInvitedId = ref<number | null>(null)
const addTeamMemberSubmitting = ref(false)
const teamEnterpriseInquiryOpen = ref(false)
const hasEnterpriseInquiry = ref(false)
const pendingTeamFlow = ref<'addMember' | null>(null)
const teamEnterpriseInquiryForm = reactive({
  company_name: '',
})
const inviteRebatePanelTab = ref<'details' | 'members' | 'rebates' | 'withdrawals'>('details')
const commissionAvailable = ref(0)
const inviteRebateOverview = ref<InviteRebateOverview | null>(null)
const inviteMembers = ref<InviteRebateMember[]>([])
const commissionRecords = ref<InviteCommissionRecord[]>([])
const commissionRecordsPage = ref(1)
const commissionRecordsPageSize = 9
const commissionRecordsTotal = ref(0)
const commissionRecordsLoaded = ref(false)
const withdrawalRecords = ref<InviteWithdrawalRecord[]>([])
const withdrawalsPage = ref(1)
const withdrawalsPageSize = 9
const withdrawalsTotal = ref(0)
const withdrawalsLoaded = ref(false)
const inviteRebateLoaded = ref(false)

/** 返佣说明（与产品文案一致；不依赖后端 notes，避免旧服务缓存旧文案） */
const inviteRebateDisplayNotes = [
  '返佣比例按「邀请下级人数」自动升级，下级为注册时已绑定到您账号的用户。',
  '受邀用户每笔已支付订单，按实付金额 × 当前返佣比例计算返利，支付成功后即时计入「佣金」。',
  '佣金可划转到余额或者提现。',
]
const payoutQr = reactive({ alipay: '', wechat: '' })
const payoutQrConfigured = reactive({ alipay: false, wechat: false })
const payoutQrSetupOpen = ref(false)
const payoutQrSetupChannel = ref<'alipay' | 'wechat'>('alipay')
const payoutQrSetupFile = ref<File | null>(null)
const payoutQrSetupPreview = ref('')
let payoutQrSetupObjectUrl: string | null = null
const withdrawOpen = ref(false)
const withdrawForm = reactive({
  channel: 'alipay' as 'alipay' | 'wechat',
  amountYuan: '',
})
const transferOpen = ref(false)
const transferForm = reactive({
  amountYuan: '',
})
const assignSubKeyOpen = ref(false)
const assignSubKeySubmitting = ref(false)
const assignSubKeyForm = reactive({
  subscriptionId: null as number | null,
  memberUserId: null as number | null,
  limitTokens: 0,
})

const editSubKeyLimitOpen = ref(false)
const editSubKeyLimitTargetId = ref<number | null>(null)
const editSubKeyLimitValue = ref(0)

const assignSubKeySubscription = computed(() => {
  const subId = Number(assignSubKeyForm.subscriptionId)
  return memberSubscriptions.value.find((s) => Number(s.id) === subId)
})

function eligibleMembersForAssignSubKey(subscriptionId: number) {
  const subId = Number(subscriptionId)
  if (!Number.isFinite(subId) || subId <= 0) return []
  const taken = new Set<number>()
  for (const k of teamApiKeys.value) {
    if (k.status !== 'active') continue
    if (Number(k.user_subscription_id) !== subId) continue
    const uid = Number(k.member_user_id)
    if (uid > 0) taken.add(uid)
  }
  return teamMembersForSubKey.value.filter((m) => !taken.has(Number(m.user_id)))
}

const assignSubKeyEligibleMembers = computed(() =>
  eligibleMembersForAssignSubKey(Number(assignSubKeyForm.subscriptionId)),
)

const editSubKeyLimitTarget = computed(() =>
  editSubKeyLimitTargetId.value == null
    ? null
    : teamApiKeys.value.find((k) => k.id === editSubKeyLimitTargetId.value) ?? null,
)

const editSubKeyLimitMainTotal = computed(() => {
  const k = editSubKeyLimitTarget.value
  if (!k) return 0
  const sub = memberSubscriptions.value.find((s) => s.id === k.user_subscription_id)
  return sub?.limit_tokens ?? k.limit_tokens
})

const activeTeamMembers = computed(() =>
  teamMembers.value.filter((m) => m.status === 'active'),
)

/** 可分配子 Key：须已绑定商城 user_id */
const teamMembersForSubKey = computed(() =>
  activeTeamMembers.value.filter((m) => Number(m.user_id) > 0),
)

const activeTeamSubKeys = computed(() =>
  teamApiKeys.value.filter((k) => k.status === 'active'),
)

const myApiKeyRows = computed(() => mainApiKeys.value)

const validTabs = new Set<MemberTab>([...memberNav.map((n) => n.id), 'messages'])

const activeTab = computed<MemberTab>(() => {
  const q = route.query.tab
  const t = typeof q === 'string' ? q : 'overview'
  return validTabs.has(t as MemberTab) ? (t as MemberTab) : 'overview'
})

watch(
  activeTab,
  (tab, prevTab) => {
    if (tab !== 'api-keys') apiKeyPanelTab.value = 'mine'
    if (tab !== 'sub-accounts') inviteRebatePanelTab.value = 'details'
    if (tab === prevTab) return
    if ((tab === 'plans' || tab === 'overview') && !tabLoadedOnce.value.plans) {
      tabLoadedOnce.value.plans = true
      void fetchPlansTabData()
    }
    if (tab === 'orders') {
      void fetchOrdersTabData()
    } else if (tab === 'overview' && !tabLoadedOnce.value.recentOrders) {
      tabLoadedOnce.value.recentOrders = true
      void fetchRecentOrders()
    }
    if ((tab === 'overview' || tab === 'settings') && !tabLoadedOnce.value.profile) {
      tabLoadedOnce.value.profile = true
      void fetchMemberProfile()
    }
    if (tab === 'account') {
      void fetchWalletFlowsTabData()
    }
    if (tab === 'invoices') {
      void fetchInvoicesTabData()
    }
    if (tab === 'coupons') {
      void fetchCouponsTabData()
    }
    if (tab === 'sub-accounts') {
      void fetchInviteRebateTabData()
    }
    if (tab === 'api-keys') {
      void fetchApiKeysTabData()
    }
  },
  { immediate: true },
)

watch(inviteRebatePanelTab, (panel) => {
  if (activeTab.value !== 'sub-accounts' || !inviteRebateLoaded.value) return
  if (panel === 'members') void fetchInviteMembersTabData()
  if (panel === 'rebates') void fetchCommissionRecordsTabData()
  if (panel === 'withdrawals') void fetchWithdrawalsTabData()
})

const pageTitle = computed(() => {
  if (activeTab.value === 'messages') return '站内消息'
  return memberNav.find((n) => n.id === activeTab.value)?.label ?? '会员中心'
})

const displayName = computed(() =>
  user.value ? userAccountLabel(user.value) : '会员',
)

const overviewSubscriptions = computed(() =>
  memberSubscriptions.value.filter((s) => s.status === 'active'),
)

const exclusivePromoDomain = computed(
  () => inviteRebateOverview.value?.promo_domain_url?.trim() ?? '',
)

onBeforeMount(() => {
  const profile = getSessionUser()
  if (!isLoggedIn() || !profile) {
    clearSessionUser()
    void router.replace({ path: '/login', query: { redirect: route.fullPath } })
    return
  }
  user.value = profile
})

function usagePercent(used: number, limit: number) {
  return limit > 0 ? Math.min(100, Math.round((used / limit) * 100)) : 0
}

function isSubscriptionActive(status: string) {
  return status === 'active'
}

function subscriptionTagClass(status: string) {
  if (status === 'active') return 'tag--active'
  // 与 API Key 非 active 一致：默认灰底标签，不用红色 cancelled
  if (status === 'expired') return ''
  return `tag--${status}`
}

async function ensureCatalogReady() {
  if (catalogProducts.value.length) return
  await reloadCatalogProducts({ soft: true })
}

async function withTopLoading<T>(task: () => Promise<T>): Promise<T> {
  beginRouteNavigationLoading()
  try {
    return await task()
  } finally {
    endRouteNavigationLoading()
  }
}

async function fetchPlansTabData() {
  void ensureCatalogReady().catch(() => {})
  await withTopLoading(async () => {
    try {
      memberSubscriptions.value = await subscriptionApi.list()
    } catch (e) {
      const msg = e instanceof Error ? e.message : ''
      if (/unauthorized|401/i.test(msg)) {
        ElMessage.error('登录已失效，请重新登录')
        clearSessionUser()
        void router.replace({ path: '/login', query: { redirect: route.fullPath } })
      } else {
        ElMessage.error(msg || '套餐列表加载失败')
      }
    } finally {
      plansLoaded.value = true
    }
  })
}

async function fetchRecentOrders() {
  await withTopLoading(async () => {
    try {
      const data = await orderApi.list({ page: 1, page_size: 3 })
      recentOrders.value = data.items
    } catch (e) {
      handleMemberAuthError(e, '最近订单加载失败')
    } finally {
      recentOrdersLoaded.value = true
    }
  })
}

async function fetchOrdersTabData() {
  await withTopLoading(async () => {
    try {
      const data = await orderApi.list({
        page: ordersPage.value,
        page_size: ordersPageSize,
      })
      memberOrders.value = data.items
      ordersTotal.value = data.total
      ordersPage.value = data.page
    } catch (e) {
      handleMemberAuthError(e, '订单列表加载失败')
    } finally {
      ordersLoaded.value = true
    }
  })
}

function onOrdersPageChange(page: number) {
  // v-model 已先更新 current-page，不可再与 ordersPage 比较后跳过请求
  ordersPage.value = page
  void fetchOrdersTabData()
}

const toastAboveModal = { zIndex: 10000 }

function handleMemberAuthError(e: unknown, fallback: string) {
  const msg = e instanceof Error ? e.message : ''
  if (/unauthorized|401/i.test(msg)) {
    ElMessage.error({ message: '登录已失效，请重新登录', ...toastAboveModal })
    clearSessionUser()
    void router.replace({ path: '/login', query: { redirect: route.fullPath } })
    return
  }
  ElMessage.error({ message: msg || fallback, ...toastAboveModal })
}

async function fetchMemberProfile() {
  await withTopLoading(async () => {
    try {
      const profile = await userApi.me()
      user.value = profile
      setSessionUser(profile)
      walletBalance.value = parseMoney(profile.wallet_balance)
      commissionAvailable.value = parseMoney(profile.commission_balance)
    } catch (e) {
      handleMemberAuthError(e, '用户信息加载失败')
    }
  })
}

async function fetchWalletFlowsTabData() {
  await withTopLoading(async () => {
    try {
      const data = await userApi.walletFlows({
        page: walletFlowsPage.value,
        page_size: walletFlowsPageSize,
      })
      walletFlows.value = data.items
      walletFlowsTotal.value = data.total
      walletFlowsPage.value = data.page
    } catch (e) {
      handleMemberAuthError(e, '资金流水加载失败')
    } finally {
      walletFlowsLoaded.value = true
    }
  })
}

function onWalletFlowsPageChange(page: number) {
  walletFlowsPage.value = page
  void fetchWalletFlowsTabData()
}

async function fetchInvoicesTabData() {
  await withTopLoading(async () => {
    try {
      const data = await invoiceApi.list({
        page: invoicesPage.value,
        page_size: invoicesPageSize,
      })
      memberInvoices.value = data.items
      invoicesTotal.value = data.total
      invoicesPage.value = data.page
    } catch (e) {
      handleMemberAuthError(e, '发票列表加载失败')
    } finally {
      invoicesLoaded.value = true
    }
  })
}

function onInvoicesPageChange(page: number) {
  invoicesPage.value = page
  void fetchInvoicesTabData()
}

const couponStatusLabel: Record<string, string> = {
  available: '可使用',
  used: '已使用',
  expired: '已过期',
}

function couponDiscountLabel(c: UserCouponItem) {
  if (c.discount_type === 'percent') {
    return `${parseFloat(c.discount_value)}% 折扣`
  }
  return `立减 ¥${parseFloat(c.discount_value).toFixed(0)}`
}

function couponStatusTagClass(status: string) {
  if (status === 'available') return 'tag--active'
  if (status === 'used') return 'tag--used'
  if (status === 'expired') return 'tag--expired'
  return ''
}

async function fetchCouponsTabData() {
  await withTopLoading(async () => {
    try {
      const data = await couponApi.listMine()
      memberCoupons.value = data.items
    } catch (e) {
      handleMemberAuthError(e, '优惠券加载失败')
    } finally {
      couponsLoaded.value = true
    }
  })
}

async function fetchApiKeysTabData() {
  await withTopLoading(async () => {
    if (!plansLoaded.value) {
      await fetchPlansTabData().catch(() => {})
    }
    if (!inviteRebateOverview.value) {
      void inviteRebateApi
        .overview()
        .then((overview) => {
          inviteRebateOverview.value = overview
        })
        .catch(() => {})
    }
    try {
      mainApiKeys.value = await apiKeyApi.listMainKeys()
    } catch (e) {
      handleMemberAuthError(e, '我的 Key 加载失败')
      mainApiKeys.value = []
    }
    try {
      teamApiKeys.value = await apiKeyApi.listTeamKeys()
    } catch {
      teamApiKeys.value = []
    }
    try {
      teamMembers.value = await apiKeyApi.listTeamMembers()
    } catch {
      teamMembers.value = []
    }
    try {
      const inquiryStatus = await apiKeyApi.getEnterpriseInquiry()
      hasEnterpriseInquiry.value = inquiryStatus.has_inquiry
    } catch {
      hasEnterpriseInquiry.value = false
    }
    apiKeysLoaded.value = true
  })
}

async function fetchAddableInvitees() {
  try {
    addableInvitees.value = await apiKeyApi.listAddableInvitees()
  } catch (e) {
    handleMemberAuthError(e, '可添加成员加载失败')
  }
}

async function fetchInviteRebateTabData() {
  await withTopLoading(async () => {
    try {
      const overview = await inviteRebateApi.overview()
      inviteRebateOverview.value = overview
      commissionAvailable.value = parseMoney(overview.commission_balance)
      const payout = await inviteRebateApi.payoutConfig()
      applyInvitePayoutConfig(payout)
      inviteRebateLoaded.value = true
      const panel = inviteRebatePanelTab.value
      if (panel === 'members') await fetchInviteMembersTabData()
      else if (panel === 'rebates') await fetchCommissionRecordsTabData()
      else if (panel === 'withdrawals') await fetchWithdrawalsTabData()
    } catch (e) {
      handleMemberAuthError(e, '邀请返利加载失败')
    }
  })
}

async function fetchInviteMembersTabData() {
  try {
    const data = await inviteRebateApi.members()
    inviteMembers.value = data.items
  } catch (e) {
    handleMemberAuthError(e, '邀请成员加载失败')
  }
}

async function fetchCommissionRecordsTabData() {
  try {
    const data = await inviteRebateApi.commissionRecords({
      page: commissionRecordsPage.value,
      page_size: commissionRecordsPageSize,
    })
    commissionRecords.value = data.items
    commissionRecordsTotal.value = data.total
    commissionRecordsPage.value = data.page
  } catch (e) {
    handleMemberAuthError(e, '返利记录加载失败')
  } finally {
    commissionRecordsLoaded.value = true
  }
}

function onCommissionRecordsPageChange(page: number) {
  commissionRecordsPage.value = page
  void fetchCommissionRecordsTabData()
}

async function fetchWithdrawalsTabData() {
  try {
    const data = await inviteRebateApi.withdrawals({
      page: withdrawalsPage.value,
      page_size: withdrawalsPageSize,
    })
    withdrawalRecords.value = data.items
    withdrawalsTotal.value = data.total
    withdrawalsPage.value = data.page
  } catch (e) {
    handleMemberAuthError(e, '提现记录加载失败')
  } finally {
    withdrawalsLoaded.value = true
  }
}

function onWithdrawalsPageChange(page: number) {
  withdrawalsPage.value = page
  void fetchWithdrawalsTabData()
}

function invoiceStatusLabel(status: string) {
  if (status === 'issued') return '已开具'
  if (status === 'failed') return '失败'
  return '处理中'
}

function invoiceStatusTagClass(status: string) {
  if (status === 'issued') return 'tag--completed'
  if (status === 'failed') return 'tag--cancelled'
  return 'tag--pending_payment'
}

function walletFlowTypeLabel(type: string) {
  if (type === 'recharge') return '充值'
  if (type === 'refund') return '退款'
  if (type === 'commission') return '佣金'
  if (type === 'withdraw') return '提现'
  return '消费'
}

function openTransferModal() {
  if (commissionAvailable.value <= 0) {
    ElMessage.warning('暂无可划转佣金')
    return
  }
  transferForm.amountYuan = ''
  transferOpen.value = true
}

async function confirmCommissionTransfer() {
  const yuan = Number(transferForm.amountYuan)
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
    walletBalance.value = parseMoney(data.wallet_balance)
    transferOpen.value = false
    ElMessage.success('已划转到账户余额')
  } catch (e) {
    handleMemberAuthError(e, '划转失败')
  }
}

function findCatalogProductById(productId: number) {
  return catalogProducts.value.find((item) => item.id === productId)
}

function logout() {
  void userApi.logout()
  clearSessionUser()
  ElMessage.success('已退出登录')
  router.push('/')
}

function mockRecharge() {
  ElMessage.info('余额充值接口对接中，可先使用支付宝 / 微信直接购套餐')
}

function goInviteRebateTab() {
  inviteRebatePanelTab.value = 'members'
  void router.push({ path: '/member', query: { tab: 'sub-accounts' } })
}

function goCommissionWithdrawFromOverview() {
  inviteRebatePanelTab.value = 'withdrawals'
  const open = () => {
    void nextTick(() => openWithdrawModal())
  }
  if (route.path === '/member' && route.query.tab === 'sub-accounts') {
    open()
    return
  }
  void router.push({ path: '/member', query: { tab: 'sub-accounts' } }).then(open)
}

function openCatalogPicker() {
  purchaseOrderType.value = 'purchase'
  purchaseUserSubscriptionId.value = 0
  purchaseSubscriptionPeriodEnd.value = ''
  purchaseSubscriptionExpiresAt.value = ''
  catalogProductsCategoryId.value = null
  catalogUpgradeBaselineLimitTokens.value = null
  catalogOpen.value = true
}

function resolveUpgradeCategoryId(sub: UserSubscriptionItem): number | null {
  const byProduct = catalogProducts.value.find((item) => item.id === sub.product_id)
  if (byProduct?.products_category_id != null && byProduct.products_category_id > 0) {
    return byProduct.products_category_id
  }
  for (const cat of catalogCategories.value) {
    if (cat.products.some((p) => p.id === sub.product_id)) {
      return cat.id
    }
  }
  const catName = sub.products_category_name?.trim()
  if (catName) {
    const cat = catalogCategories.value.find((c) =>
      c.products.some((p) => p.products_category_name === catName),
    )
    if (cat) return cat.id
  }
  return null
}

async function openUpgradeCatalog(sub: UserSubscriptionItem) {
  await ensureCatalogReady()
  const categoryId = resolveUpgradeCategoryId(sub)
  if (categoryId == null) {
    ElMessage.warning('未找到该套餐分类，请稍后重试或联系客服')
    return
  }
  purchaseOrderType.value = 'upgrade'
  purchaseUserSubscriptionId.value = sub.id
  purchaseSubscriptionPeriodEnd.value = sub.period_end
  purchaseSubscriptionExpiresAt.value = sub.expires_at
  catalogProductsCategoryId.value = categoryId
  catalogUpgradeBaselineLimitTokens.value = sub.limit_tokens
  catalogOpen.value = true
}

async function openQuotaAddon(sub: UserSubscriptionItem) {
  await ensureCatalogReady()
  const p = findCatalogProductById(sub.product_id)
  if (!p) {
    ElMessage.warning('未找到该套餐商品，请稍后重试或联系客服')
    return
  }
  purchaseOrderType.value = 'quota_addon'
  purchaseUserSubscriptionId.value = sub.id
  purchaseProduct.value = p
  purchaseOpen.value = true
}

async function openRenewSubscription(sub: UserSubscriptionItem) {
  await ensureCatalogReady()
  const p = findCatalogProductById(sub.product_id)
  if (!p) {
    ElMessage.warning('未找到该套餐商品，请稍后重试或联系客服')
    return
  }
  purchaseOrderType.value = 'renewal'
  purchaseUserSubscriptionId.value = sub.id
  purchaseProduct.value = p
  purchaseOpen.value = true
}

function onCatalogBuy(p: CatalogProduct) {
  if (purchaseOrderType.value !== 'upgrade') {
    purchaseOrderType.value = 'purchase'
    purchaseUserSubscriptionId.value = 0
  }
  purchaseProduct.value = p
  purchaseOpen.value = true
}

function onPurchasePaid() {
  catalogOpen.value = false
  if (activeTab.value === 'plans' || activeTab.value === 'api-keys') {
    void fetchPlansTabData()
  }
  if (activeTab.value === 'api-keys') {
    void fetchApiKeysTabData()
  }
}

function mockAction(msg: string) {
  ElMessage.info(msg)
}

function fallbackCopyText(text: string): boolean {
  try {
    const ta = document.createElement('textarea')
    ta.value = text
    ta.setAttribute('readonly', 'true')
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

async function copyToClipboard(text: string, successMessage: string) {
  const value = text.trim()
  if (!value) {
    ElMessage.warning({ message: '没有可复制的内容', ...toastAboveModal })
    return
  }
  let copied = false
  if (typeof navigator !== 'undefined' && navigator.clipboard?.writeText) {
    try {
      await navigator.clipboard.writeText(value)
      copied = true
    } catch {
      copied = fallbackCopyText(value)
    }
  } else {
    copied = fallbackCopyText(value)
  }
  if (copied) {
    ElMessage.success({ message: successMessage, ...toastAboveModal })
  } else {
    ElMessage.warning({ message: '复制失败，请手动选择输入框内容复制', ...toastAboveModal })
  }
}

async function openTeamInviteModal() {
  if (!exclusivePromoDomain.value) {
    try {
      const overview = await inviteRebateApi.overview()
      inviteRebateOverview.value = overview
    } catch (e) {
      handleMemberAuthError(e, '推广域名加载失败')
      return
    }
  }
  if (!exclusivePromoDomain.value) {
    ElMessage.warning('请先在「邀请返利」配置专属推广域名')
    return
  }
  teamInviteOpen.value = true
}

function copyTeamInviteLink() {
  const value = exclusivePromoDomain.value.trim()
  teamInviteOpen.value = false
  void nextTick(() => {
    if (!value) {
      ElMessage.warning('暂无可复制的推广域名')
      return
    }
    const copied = fallbackCopyText(value)
    if (copied) {
      ElMessage.success('已复制邀请地址')
    } else {
      ElMessage.warning('复制失败，请手动复制')
    }
  })
}

function copyExclusivePromoDomain() {
  void copyToClipboard(exclusivePromoDomain.value, '已复制推广域名')
}

function applyInvitePayoutConfig(payout: {
  alipay_qr_data_url?: string
  wechat_qr_data_url?: string
  alipay_configured?: boolean
  wechat_configured?: boolean
}) {
  payoutQr.alipay = payout.alipay_qr_data_url ?? ''
  payoutQr.wechat = payout.wechat_qr_data_url ?? ''
  payoutQrConfigured.alipay = Boolean(payout.alipay_configured ?? payoutQr.alipay)
  payoutQrConfigured.wechat = Boolean(payout.wechat_configured ?? payoutQr.wechat)
}

function clearPayoutQrSetupFile() {
  if (payoutQrSetupObjectUrl) {
    URL.revokeObjectURL(payoutQrSetupObjectUrl)
    payoutQrSetupObjectUrl = null
  }
  payoutQrSetupFile.value = null
}

function closePayoutQrModal() {
  payoutQrSetupOpen.value = false
  clearPayoutQrSetupFile()
}

function openPayoutQrModal(channel: 'alipay' | 'wechat') {
  clearPayoutQrSetupFile()
  payoutQrSetupChannel.value = channel
  payoutQrSetupPreview.value = channel === 'alipay' ? payoutQr.alipay : payoutQr.wechat
  payoutQrSetupOpen.value = true
}

function onPayoutQrFileChange(ev: Event) {
  const input = ev.target as HTMLInputElement
  const file = input.files?.[0] ?? null
  clearPayoutQrSetupFile()
  if (!file) {
    payoutQrSetupPreview.value =
      payoutQrSetupChannel.value === 'alipay' ? payoutQr.alipay : payoutQr.wechat
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
  payoutQrSetupFile.value = file
  payoutQrSetupObjectUrl = URL.createObjectURL(file)
  payoutQrSetupPreview.value = payoutQrSetupObjectUrl
}

async function confirmPayoutQrSetup() {
  const file = payoutQrSetupFile.value
  if (!file) {
    ElMessage.warning('请选择收款码图片')
    return
  }
  try {
    const cfg = await inviteRebateApi.uploadPayoutQr(payoutQrSetupChannel.value, file)
    applyInvitePayoutConfig(cfg)
    payoutQrSetupOpen.value = false
    clearPayoutQrSetupFile()
    ElMessage.success('收款码已保存')
  } catch (e) {
    handleMemberAuthError(e, '保存收款码失败')
  }
}

function openWithdrawModal() {
  if (!payoutQrConfigured.alipay && !payoutQrConfigured.wechat) {
    ElMessage.warning('请先设置支付宝或微信收款码')
    return
  }
  withdrawForm.channel = payoutQrConfigured.alipay ? 'alipay' : 'wechat'
  withdrawForm.amountYuan = ''
  withdrawOpen.value = true
}

async function confirmWithdraw() {
  const yuan = Number(withdrawForm.amountYuan)
  if (!Number.isFinite(yuan) || yuan <= 0) {
    ElMessage.warning('请输入有效的提现金额')
    return
  }
  if (yuan > commissionAvailable.value) {
    ElMessage.warning('提现金额不能超过可提现佣金')
    return
  }
  const channel = withdrawForm.channel
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
    handleMemberAuthError(e, '提现申请失败')
  }
}

const payoutQrSetupTitle = computed(() =>
  payoutQrSetupChannel.value === 'alipay' ? '设置支付宝收款码' : '设置微信收款码',
)

function syncAssignSubKeyLimitFromPlan() {
  const subId = Number(assignSubKeyForm.subscriptionId)
  const sub = memberSubscriptions.value.find((s) => Number(s.id) === subId)
  assignSubKeyForm.limitTokens = sub?.limit_tokens ?? 0
}

function syncAssignSubKeyMemberFromPlan() {
  const eligible = assignSubKeyEligibleMembers.value
  const current = Number(assignSubKeyForm.memberUserId)
  if (eligible.some((m) => Number(m.user_id) === current)) return
  assignSubKeyForm.memberUserId = eligible[0]?.user_id ?? null
}

function onAssignSubKeyPlanChange() {
  syncAssignSubKeyLimitFromPlan()
  syncAssignSubKeyMemberFromPlan()
}

function openAssignSubKeyModal() {
  assignSubKeySubmitting.value = false
  const activeSubs = memberSubscriptions.value.filter((s) => s.status === 'active')
  if (!activeSubs.length) {
    ElMessage.warning({ message: '请先开通套餐', ...toastAboveModal })
    return
  }
  if (!teamMembersForSubKey.value.length) {
    ElMessage.warning({
      message: '请先在「我的团队」添加已注册成员（须绑定商城账号）',
      ...toastAboveModal,
    })
    return
  }
  const subWithSlot =
    activeSubs.find((s) => eligibleMembersForAssignSubKey(Number(s.id)).length > 0) ?? null
  if (!subWithSlot) {
    ElMessage.warning({
      message: '各生效套餐下的团队成员均已分配子 Key',
      ...toastAboveModal,
    })
    return
  }
  assignSubKeyForm.subscriptionId = subWithSlot.id
  syncAssignSubKeyLimitFromPlan()
  syncAssignSubKeyMemberFromPlan()
  assignSubKeyOpen.value = true
}

function openEditSubKeyLimitModal(k: UserAPIKeyItem) {
  editSubKeyLimitTargetId.value = k.id
  editSubKeyLimitValue.value = k.limit_tokens
  editSubKeyLimitOpen.value = true
}

async function confirmEditSubKeyLimit() {
  const k = editSubKeyLimitTarget.value
  if (!k) return
  const max = editSubKeyLimitMainTotal.value
  const next = Math.floor(editSubKeyLimitValue.value)
  if (!Number.isFinite(next) || next <= 0) {
    ElMessage.warning('请输入有效的用量上限')
    return
  }
  if (next > max) {
    ElMessage.warning(`不能超过主 Key 套餐总量（${formatTokens(max)}）`)
    return
  }
  if (next < k.used_tokens) {
    ElMessage.warning(`上限不能低于已用量（${formatTokens(k.used_tokens)}）`)
    return
  }
  try {
    await apiKeyApi.updateSubKeyLimit(k.id, next)
    k.limit_tokens = next
    editSubKeyLimitOpen.value = false
    ElMessage.success('子 Key 用量上限已更新')
  } catch (e) {
    handleMemberAuthError(e, '更新失败')
  }
}

async function confirmAssignSubKey() {
  if (assignSubKeySubmitting.value) return
  const subId = Number(assignSubKeyForm.subscriptionId)
  const memberUserId = Number(assignSubKeyForm.memberUserId)
  if (!Number.isFinite(subId) || subId <= 0 || !Number.isFinite(memberUserId) || memberUserId <= 0) {
    ElMessage.warning({ message: '请选择套餐与团队成员', ...toastAboveModal })
    return
  }
  if (!assignSubKeyEligibleMembers.value.some((m) => Number(m.user_id) === memberUserId)) {
    ElMessage.warning({
      message: '该成员在此套餐下已分配子 Key，请重新选择',
      ...toastAboveModal,
    })
    return
  }
  const subscription = memberSubscriptions.value.find((s) => Number(s.id) === subId)
  if (!subscription) {
    ElMessage.warning({ message: '未找到所选套餐，请关闭弹窗后重试', ...toastAboveModal })
    return
  }

  const mainLimit = Number(subscription.limit_tokens)
  let limitTokens = Math.floor(Number(assignSubKeyForm.limitTokens))
  if (!Number.isFinite(limitTokens) || limitTokens <= 0) limitTokens = mainLimit
  if (!Number.isFinite(mainLimit) || mainLimit <= 0) {
    ElMessage.warning({ message: '套餐额度无效，请刷新页面后重试', ...toastAboveModal })
    return
  }
  if (limitTokens > mainLimit) {
    ElMessage.warning({
      message: `用量上限不能超过主 Key 套餐总量（${formatTokens(mainLimit)}）`,
      ...toastAboveModal,
    })
    return
  }

  assignSubKeySubmitting.value = true
  try {
    const created = await apiKeyApi.createSubKey({
      user_subscription_id: subId,
      member_user_id: memberUserId,
      limit_tokens: limitTokens,
    })
    teamApiKeys.value.unshift(created)
    assignSubKeyOpen.value = false
    apiKeyPanelTab.value = 'team'
    void nextTick(() => {
      if (created.api_key) {
        void copyToClipboard(
          created.api_key,
          '子 Key 已复制（完整 Key 仅展示一次，请妥善保存）',
        )
      } else {
        ElMessage.success({ message: '子 Key 已分配', ...toastAboveModal })
      }
    })
    void fetchApiKeysTabData()
  } catch (e) {
    handleMemberAuthError(e, '分配子 Key 失败')
  } finally {
    assignSubKeySubmitting.value = false
  }
}

function prefillTeamEnterpriseInquiryForm() {
  teamEnterpriseInquiryForm.company_name = ''
}

async function ensureEnterpriseInquiryForTeam(): Promise<boolean> {
  if (hasEnterpriseInquiry.value) return true
  try {
    const status = await apiKeyApi.getEnterpriseInquiry()
    hasEnterpriseInquiry.value = status.has_inquiry
    if (status.has_inquiry) return true
    prefillTeamEnterpriseInquiryForm()
    teamEnterpriseInquiryOpen.value = true
    return false
  } catch (e) {
    handleMemberAuthError(e, '企业信息校验失败')
    return false
  }
}

async function proceedOpenAddTeamMemberModal() {
  await fetchAddableInvitees()
  if (!addableInvitees.value.length) {
    ElMessage.warning('暂无已邀请且未加入团队的成员，请先分享专属推广域名')
    return
  }
  addTeamMemberInvitedId.value = addableInvitees.value[0]?.user_id ?? null
  addTeamMemberOpen.value = true
}

async function openAddTeamMemberModal() {
  const ready = await ensureEnterpriseInquiryForTeam()
  if (!ready) {
    pendingTeamFlow.value = 'addMember'
    return
  }
  await proceedOpenAddTeamMemberModal()
}

async function confirmTeamEnterpriseInquiry() {
  const company = teamEnterpriseInquiryForm.company_name.trim()
  if (!company) {
    ElMessage.warning('请填写企业 / 团队名称')
    return
  }
  try {
    await apiKeyApi.submitEnterpriseInquiry({ company_name: company })
    hasEnterpriseInquiry.value = true
    teamEnterpriseInquiryOpen.value = false
    ElMessage.success('企业信息已保存')
    const flow = pendingTeamFlow.value
    pendingTeamFlow.value = null
    if (flow === 'addMember') {
      await proceedOpenAddTeamMemberModal()
    }
  } catch (e) {
    handleMemberAuthError(e, '保存企业信息失败')
  }
}

async function confirmAddTeamMember() {
  if (addTeamMemberSubmitting.value) return
  const invitedId = Number(addTeamMemberInvitedId.value)
  if (!Number.isFinite(invitedId) || invitedId <= 0) {
    ElMessage.warning('请选择成员')
    return
  }
  addTeamMemberSubmitting.value = true
  try {
    await apiKeyApi.addTeamMember(invitedId)
    teamMembers.value = await apiKeyApi.listTeamMembers()
    addableInvitees.value = addableInvitees.value.filter((u) => u.user_id !== invitedId)
    addTeamMemberOpen.value = false
    ElMessage.success('已加入团队')
  } catch (e) {
    handleMemberAuthError(e, '添加成员失败')
  } finally {
    addTeamMemberSubmitting.value = false
  }
}
</script>

<template>
  <div v-if="user" class="member-layout">
    <section class="member-hero">
      <div class="member-hero-bg" aria-hidden="true" />
      <div class="atm-container member-hero-inner">
        <div class="member-hero-copy">
          <h1 class="member-hero-title">
            <span class="member-hero-greeting">你好</span>
            <span class="member-hero-account">{{ displayName }}</span>
          </h1>
        </div>
        <div class="member-hero-actions">
          <button type="button" class="hero-btn hero-btn--light" @click="openCatalogPicker">
            选购套餐
          </button>
          <button type="button" class="hero-btn hero-btn--ghost" @click="goInviteRebateTab">
            邀请返利
          </button>
        </div>
      </div>
    </section>

    <div class="atm-container member-body">
      <div class="member-grid">
        <div class="member-sidebar-wrap">
          <MemberSidebar :active="activeTab" :user="user" @pick-plan="openCatalogPicker" />
        </div>

        <main class="member-main">
          <section class="panel">
            <header
              class="panel-head"
              :class="{
                'panel-head--segmented':
                  activeTab === 'api-keys' || activeTab === 'sub-accounts',
              }"
            >
              <div class="panel-head-row">
                <h2 class="panel-head-title">{{ pageTitle }}</h2>
                <button
                  v-if="activeTab === 'invoices'"
                  type="button"
                  class="atm-btn-primary btn-xs panel-head-action"
                  @click="invoiceConfigOpen = true"
                >
                  管理发票抬头
                </button>
              </div>
              <div
                v-if="activeTab === 'sub-accounts'"
                class="api-key-tabs"
                role="tablist"
                aria-label="邀请返利"
              >
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': inviteRebatePanelTab === 'details' }"
                  role="tab"
                  :aria-selected="inviteRebatePanelTab === 'details'"
                  @click="inviteRebatePanelTab = 'details'"
                >
                  返佣详情
                </button>
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': inviteRebatePanelTab === 'withdrawals' }"
                  role="tab"
                  :aria-selected="inviteRebatePanelTab === 'withdrawals'"
                  @click="inviteRebatePanelTab = 'withdrawals'"
                >
                  我的佣金
                </button>
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': inviteRebatePanelTab === 'members' }"
                  role="tab"
                  :aria-selected="inviteRebatePanelTab === 'members'"
                  @click="inviteRebatePanelTab = 'members'"
                >
                  邀请成员
                </button>
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': inviteRebatePanelTab === 'rebates' }"
                  role="tab"
                  :aria-selected="inviteRebatePanelTab === 'rebates'"
                  @click="inviteRebatePanelTab = 'rebates'"
                >
                  返利记录
                </button>
              </div>
              <div
                v-if="activeTab === 'api-keys'"
                class="api-key-tabs"
                role="tablist"
                aria-label="API Key 类型"
              >
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': apiKeyPanelTab === 'mine' }"
                  role="tab"
                  :aria-selected="apiKeyPanelTab === 'mine'"
                  @click="apiKeyPanelTab = 'mine'"
                >
                  我的 Key
                </button>
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': apiKeyPanelTab === 'team' }"
                  role="tab"
                  :aria-selected="apiKeyPanelTab === 'team'"
                  @click="apiKeyPanelTab = 'team'"
                >
                  团队子 Key
                </button>
                <button
                  type="button"
                  class="api-key-tab"
                  :class="{ 'api-key-tab--active': apiKeyPanelTab === 'group' }"
                  role="tab"
                  :aria-selected="apiKeyPanelTab === 'group'"
                  @click="apiKeyPanelTab = 'group'"
                >
                  我的团队
                </button>
              </div>
            </header>

            <!-- 概览 -->
            <div v-if="activeTab === 'overview'" class="panel-body">
            <div class="stat-row">
              <article class="stat-card stat-card--balance">
                <div class="stat-main">
                  <span class="stat-label">余额</span>
                  <strong class="stat-value">{{ formatCny(walletBalance) }}</strong>
                </div>
                <button type="button" class="stat-link" @click="mockRecharge">充值</button>
              </article>
              <article class="stat-card stat-card--commission">
                <div class="stat-main">
                  <span class="stat-label">佣金</span>
                  <strong class="stat-value">{{
                    formatCny(commissionAvailable)
                  }}</strong>
                </div>
                <div class="stat-card-actions">
                  <button type="button" class="stat-link" @click="openTransferModal">
                    划转
                  </button>
                  <button type="button" class="stat-link" @click="goCommissionWithdrawFromOverview">
                    提现
                  </button>
                </div>
              </article>
            </div>

            <h2 class="panel-subtitle">套餐用量</h2>
            <div v-if="overviewSubscriptions.length" class="plan-mini-list">
              <article
                v-for="sub in overviewSubscriptions"
                :key="sub.id"
                class="plan-mini"
              >
                <div class="plan-mini-head">
                  <strong>{{ sub.product_name }}</strong>
                  <span>至 {{ sub.expires_at }}</span>
                </div>
                <div class="progress-track">
                  <div
                    class="progress-fill"
                    :style="{ width: `${usagePercent(sub.used_tokens, sub.limit_tokens)}%` }"
                  />
                </div>
                <p class="progress-meta">
                  已用 {{ formatTokens(sub.used_tokens) }} / {{ formatTokens(sub.limit_tokens) }}
                </p>
              </article>
            </div>
            <p v-else-if="plansLoaded" class="empty">暂无使用中的套餐</p>

            <h2 class="panel-subtitle">最近订单</h2>
            <div v-if="recentOrders.length" class="table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>订单号</th>
                    <th>商品</th>
                    <th>金额</th>
                    <th>状态</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="o in recentOrders" :key="o.order_no">
                    <td>{{ o.order_no }}</td>
                    <td>{{ o.product_name }}</td>
                    <td>{{ formatCny(o.total_amount) }}</td>
                    <td>
                      <span class="tag" :class="`tag--${o.status}`">{{
                        orderStatusLabel[o.status]
                      }}</span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else-if="recentOrdersLoaded" class="empty">暂无订单</p>
            <p v-if="recentOrders.length" class="overview-orders-more muted">
              <RouterLink :to="{ path: '/member', query: { tab: 'orders' } }">查看全部订单 →</RouterLink>
            </p>
            </div>

            <!-- 站内消息（演示 UI，mock 数据） -->
            <div v-else-if="activeTab === 'messages'" class="panel-body">
              <MemberMessagesPanel />
            </div>

            <!-- 我的套餐 -->
            <div v-else-if="activeTab === 'plans'" class="panel-body">
              <article
                v-for="sub in memberSubscriptions"
                :key="sub.id"
                class="plan-card"
                :class="{ 'plan-card--inactive': !isSubscriptionActive(sub.status) }"
              >
                <div class="plan-card-head">
                  <div>
                    <h2>{{ sub.product_name }}</h2>
                  </div>
                  <div class="plan-card-tags">
                    <span v-if="isSubscriptionActive(sub.status)" class="tag tag--active">
                      余额自动续费中
                    </span>
                    <span
                      v-else
                      class="tag"
                      :class="subscriptionTagClass(sub.status)"
                    >
                      {{ subscriptionStatusLabel[sub.status] ?? sub.status }}
                    </span>
                  </div>
                </div>
                <div class="plan-metrics">
                  <div>
                    <span class="metric-label">本周期额度</span>
                    <strong>{{ formatTokens(sub.limit_tokens) }} tokens</strong>
                  </div>
                  <div>
                    <span class="metric-label">已使用</span>
                    <strong>{{ formatTokens(sub.used_tokens) }}</strong>
                  </div>
                  <div>
                    <span class="metric-label">周期截止</span>
                    <strong>{{ sub.period_end }}</strong>
                  </div>
                  <div>
                    <span class="metric-label">服务到期</span>
                    <strong>{{ sub.expires_at }}</strong>
                  </div>
                </div>
                <div
                  class="progress-track progress-track--lg"
                  :class="{ 'progress-track--inactive': !isSubscriptionActive(sub.status) }"
                >
                  <div
                    v-if="isSubscriptionActive(sub.status)"
                    class="progress-fill"
                    :style="{ width: `${usagePercent(sub.used_tokens, sub.limit_tokens)}%` }"
                  />
                </div>
                <div v-if="isSubscriptionActive(sub.status)" class="plan-actions">
                  <button
                    type="button"
                    class="atm-btn-primary btn-xs"
                    @click="openRenewSubscription(sub)"
                  >
                    续费
                  </button>
                  <button type="button" class="atm-btn-primary btn-xs" @click="openUpgradeCatalog(sub)">
                    升档
                  </button>
                  <button
                    type="button"
                    class="atm-btn-primary btn-xs"
                    @click="openQuotaAddon(sub)"
                  >
                    加本期额度
                  </button>
                </div>
              </article>
              <p v-if="plansLoaded && !memberSubscriptions.length" class="empty">
                暂无套餐记录，去首页选购吧。
              </p>
            </div>

            <!-- API Key -->
            <div v-else-if="activeTab === 'api-keys'" class="panel-body panel-body--segmented">
            <div v-if="apiKeyPanelTab === 'mine'" role="tabpanel">
              <div v-if="myApiKeyRows.length" class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>类型</th>
                      <th>关联套餐</th>
                      <th>API Key</th>
                      <th>使用量 / 总量</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="row in myApiKeyRows" :key="row.id">
                      <td>
                        <span
                          class="tag"
                          :class="row.key_type === 'main' ? 'tag--key-main' : 'tag--key-team'"
                        >
                          {{ row.key_type === 'main' ? '主 Key' : '团队子 Key' }}
                        </span>
                      </td>
                      <td>{{ row.subscription_name }}</td>
                      <td class="mono">{{ row.key_masked }}</td>
                      <td class="subkey-usage">
                        <span class="subkey-usage-text">
                          {{ formatTokens(row.used_tokens) }} / {{ formatTokens(row.limit_tokens) }}
                        </span>
                        <div class="progress-track progress-track--sm">
                          <div
                            class="progress-fill"
                            :style="{
                              width: `${usagePercent(row.used_tokens, row.limit_tokens)}%`,
                            }"
                          />
                        </div>
                      </td>
                      <td>
                        <span
                          class="tag"
                          :class="row.status === 'active' ? 'tag--active' : ''"
                        >{{
                          row.status === 'active' ? '使用中' : '不可用'
                        }}</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else class="empty">
                暂无 Key。开通套餐获得主 Key，或由团队负责人为您分配子 Key 后将在此展示。
              </p>
            </div>

            <div v-else-if="apiKeyPanelTab === 'team'" role="tabpanel">
              <div class="panel-tab-toolbar">
                <button type="button" class="atm-btn-primary btn-xs" @click="openAssignSubKeyModal">
                  分配子 Key
                </button>
              </div>
              <div v-if="activeTeamSubKeys.length" class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>成员</th>
                      <th>关联套餐</th>
                      <th>使用量 / 总量</th>
                      <th />
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="k in activeTeamSubKeys" :key="k.id">
                      <td>
                        <strong>{{ k.member_nickname }}</strong>
                        <span class="cell-sub muted">{{ k.member_email }}</span>
                      </td>
                      <td>{{ k.subscription_name }}</td>
                      <td class="subkey-usage">
                        <span class="subkey-usage-text">
                          {{ formatTokens(k.used_tokens) }} / {{ formatTokens(k.limit_tokens) }}
                        </span>
                        <div class="progress-track progress-track--sm">
                          <div
                            class="progress-fill"
                            :style="{
                              width: `${usagePercent(k.used_tokens, k.limit_tokens)}%`,
                            }"
                          />
                        </div>
                      </td>
                      <td class="subkey-row-actions">
                        <button
                          type="button"
                          class="link-btn"
                          @click="openEditSubKeyLimitModal(k)"
                        >
                          设置
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else class="empty">
                暂无团队子 Key。点击上方「分配子 Key」，为成员指定套餐下的独立 Key。
              </p>
            </div>

            <div v-else role="tabpanel">
              <div class="panel-tab-toolbar">
                <button type="button" class="atm-btn-primary btn-xs" @click="openTeamInviteModal">
                  邀请链接
                </button>
                <button type="button" class="atm-btn-primary btn-xs" @click="openAddTeamMemberModal">
                  添加成员
                </button>
              </div>
              <div v-if="activeTeamMembers.length" class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>昵称</th>
                      <th>邮箱</th>
                      <th>状态</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="m in teamMembers" :key="m.id">
                      <td>{{ m.nickname }}</td>
                      <td>{{ m.email }}</td>
                      <td>
                        <span class="tag tag--active">{{
                          m.status === 'active' ? '正常' : '停用'
                        }}</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else class="empty">
                团队暂无成员，点击上方「添加成员」从已邀请用户中选择。
              </p>
            </div>
            </div>

            <!-- 订单记录 -->
            <div v-else-if="activeTab === 'orders'" class="panel-body">
            <div v-if="memberOrders.length" class="table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>订单号</th>
                    <th>商品</th>
                    <th>数量</th>
                    <th>金额</th>
                    <th>开票</th>
                    <th>状态</th>
                    <th>时间</th>
                    <th />
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="o in memberOrders" :key="o.order_no">
                    <td class="mono">{{ o.order_no }}</td>
                    <td>{{ o.product_name }}</td>
                    <td>{{ o.quantity }}</td>
                    <td>{{ formatCny(o.total_amount) }}</td>
                    <td>{{ o.enterprise_invoice ? '企业' : '—' }}</td>
                    <td>
                      <span class="tag" :class="`tag--${o.status}`">{{
                        orderStatusLabel[o.status]
                      }}</span>
                    </td>
                    <td class="muted">{{ o.created_at }}</td>
                    <td>
                      <button
                        v-if="o.status === 'pending_payment'"
                        type="button"
                        class="link-btn"
                        @click="mockAction('继续支付对接中')"
                      >
                        去支付
                      </button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else-if="ordersLoaded" class="empty">暂无订单</p>
            <div v-if="ordersTotal > ordersPageSize" class="orders-pagination">
              <el-pagination
                v-model:current-page="ordersPage"
                :page-size="ordersPageSize"
                :total="ordersTotal"
                layout="total, prev, pager, next"
                background
                @current-change="onOrdersPageChange"
              />
            </div>
            </div>

            <!-- 资金流水 -->
            <div v-else-if="activeTab === 'coupons'" class="panel-body">
              <div v-if="memberCoupons.length" class="table-wrap">
                <table class="data-table">
                  <thead>
                    <tr>
                      <th>名称</th>
                      <th>优惠</th>
                      <th>门槛</th>
                      <th>有效期</th>
                      <th>状态</th>
                      <th>券码</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="c in memberCoupons" :key="c.id">
                      <td>{{ c.title || c.campaign_code }}</td>
                      <td>{{ couponDiscountLabel(c) }}</td>
                      <td>
                        {{
                          parseFloat(c.min_order_amount) > 0
                            ? `满 ¥${parseFloat(c.min_order_amount).toFixed(0)}`
                            : '无门槛'
                        }}
                      </td>
                      <td class="muted">{{ c.valid_from }} ~ {{ c.valid_until }}</td>
                      <td>
                        <span class="tag" :class="couponStatusTagClass(c.status)">{{
                          couponStatusLabel[c.status] ?? c.status
                        }}</span>
                      </td>
                      <td class="mono">{{ c.coupon_code }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else-if="couponsLoaded" class="empty">暂无优惠券，新用户注册可领取新人礼</p>
            </div>

            <div v-else-if="activeTab === 'account'" class="panel-body">
            <div v-if="walletFlows.length" class="table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>类型</th>
                    <th>金额</th>
                    <th>说明</th>
                    <th>时间</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="tx in walletFlows" :key="tx.id">
                    <td>{{ walletFlowTypeLabel(tx.type) }}</td>
                    <td :class="signedMoneyClass(tx.amount)">
                      {{ formatSignedCny(tx.amount) }}
                    </td>
                    <td>{{ tx.remark }}</td>
                    <td class="muted">{{ tx.created_at }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else-if="walletFlowsLoaded" class="empty">暂无流水</p>
            <div v-if="walletFlowsTotal > walletFlowsPageSize" class="orders-pagination">
              <el-pagination
                v-model:current-page="walletFlowsPage"
                :page-size="walletFlowsPageSize"
                :total="walletFlowsTotal"
                layout="total, prev, pager, next"
                background
                @current-change="onWalletFlowsPageChange"
              />
            </div>
            </div>

            <!-- 发票 -->
            <div v-else-if="activeTab === 'invoices'" class="panel-body">
            <div v-if="memberInvoices.length" class="table-wrap">
              <table class="data-table">
                <thead>
                  <tr>
                    <th>订单号</th>
                    <th>抬头</th>
                    <th>金额</th>
                    <th>状态</th>
                    <th>申请时间</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="inv in memberInvoices" :key="inv.id">
                    <td class="mono">{{ inv.order_no }}</td>
                    <td>{{ inv.title }}</td>
                    <td>{{ formatCny(inv.amount) }}</td>
                    <td>
                      <span class="tag" :class="invoiceStatusTagClass(inv.status)">{{
                        invoiceStatusLabel(inv.status)
                      }}</span>
                    </td>
                    <td class="muted">{{ inv.created_at }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
            <p v-else-if="invoicesLoaded" class="empty">暂无发票记录</p>
            <div v-if="invoicesTotal > invoicesPageSize" class="orders-pagination">
              <el-pagination
                v-model:current-page="invoicesPage"
                :page-size="invoicesPageSize"
                :total="invoicesTotal"
                layout="total, prev, pager, next"
                background
                @current-change="onInvoicesPageChange"
              />
            </div>
            </div>

            <!-- 邀请返利 -->
            <div v-else-if="activeTab === 'sub-accounts'" class="panel-body panel-body--segmented">
            <div v-if="inviteRebatePanelTab === 'details'" role="tabpanel">
              <article class="rebate-details-summary">
                <div class="rebate-details-section">
                  <span class="metric-label">专属推广域名</span>
                  <div class="promo-domain-row">
                    <code class="promo-domain-value">{{ exclusivePromoDomain }}</code>
                    <button
                      type="button"
                      class="atm-btn-primary btn-xs"
                      @click="copyExclusivePromoDomain"
                    >
                      复制
                    </button>
                  </div>
                </div>
                <div class="rebate-details-section rebate-details-metrics-row">
                  <div class="rebate-details-metric">
                    <span class="metric-label">您当前的返佣比例</span>
                    <p class="rebate-rate-value">
                      {{ parseMoney(inviteRebateOverview?.current_rate_percent ?? 0)
                      }}<span class="rebate-rate-unit">%</span>
                    </p>
                  </div>
                  <div class="rebate-details-metric">
                    <span class="metric-label">累积邀请用户</span>
                    <p class="rebate-rate-value">
                      {{ inviteRebateOverview?.valid_invite_count ?? 0
                      }}<span class="rebate-rate-unit"> 人</span>
                    </p>
                  </div>
                  <div class="rebate-details-metric">
                    <span class="metric-label">下级累计消费（已完成订单）</span>
                    <p class="rebate-rate-value">
                      {{ formatCny(inviteRebateOverview?.invitee_paid_total ?? 0) }}
                    </p>
                  </div>
                </div>
              </article>

              <h3 class="panel-subtitle">等级与比例</h3>
              <div class="table-wrap">
                <table class="data-table">
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
                      v-for="tier in inviteRebateOverview?.tiers ?? []"
                      :key="tier.level_label"
                      :class="{
                        'rebate-tier-row--current': tier.is_current,
                      }"
                    >
                      <td>
                        <strong>{{ tier.level_label }}</strong>
                        <span v-if="tier.is_current" class="tag tag--key-main rebate-tier-badge">
                          当前
                        </span>
                      </td>
                      <td>{{ tier.min_valid_invites }} 人</td>
                      <td>{{ formatCny(tier.min_invitee_paid_amount) }}</td>
                      <td>{{ parseMoney(tier.rate_percent) }}%</td>
                    </tr>
                  </tbody>
                </table>
              </div>

              <h3 class="panel-subtitle">说明</h3>
              <ul class="rebate-notes">
                <li v-for="(note, i) in inviteRebateDisplayNotes" :key="i">{{ note }}</li>
              </ul>
            </div>

            <div v-else-if="inviteRebatePanelTab === 'members'" role="tabpanel">
              <div class="table-wrap">
                <table class="data-table">
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
                      <td>
                        <span class="tag tag--active">已绑定</span>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-if="inviteRebateLoaded && !inviteMembers.length" class="empty">
                暂无邀请成员，请分享您的专属推广域名。
              </p>
            </div>

            <div v-else-if="inviteRebatePanelTab === 'rebates'" role="tabpanel">
              <p class="rebate-lead muted">
                受邀用户完成支付后，返利将自动加到「佣金」。
              </p>
              <div v-if="commissionRecords.length" class="table-wrap">
                <table class="data-table">
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
                        <span v-if="r.invitee_email" class="cell-sub muted">{{ r.invitee_email }}</span>
                      </td>
                      <td>{{ r.product_name }}</td>
                      <td>{{ formatCny(r.order_amount) }}</td>
                      <td class="amount-plus">+{{ formatCny(r.rebate_amount) }}</td>
                      <td class="muted">{{ r.created_at }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else-if="commissionRecordsLoaded" class="empty">暂无返利记录。</p>
              <div v-if="commissionRecordsTotal > commissionRecordsPageSize" class="orders-pagination">
                <el-pagination
                  v-model:current-page="commissionRecordsPage"
                  :page-size="commissionRecordsPageSize"
                  :total="commissionRecordsTotal"
                  layout="total, prev, pager, next"
                  background
                  @current-change="onCommissionRecordsPageChange"
                />
              </div>
            </div>

            <div v-else-if="inviteRebatePanelTab === 'withdrawals'" role="tabpanel">
              <section class="withdraw-section-box">
                <h3 class="withdraw-section-title">佣金提现</h3>
                <div class="withdraw-toolbar">
                  <span class="withdraw-balance">
                    <span class="withdraw-payout-name">可提现佣金</span>
                    <strong>{{ formatCny(commissionAvailable) }}</strong>
                  </span>
                  <div class="withdraw-toolbar-actions">
                    <button type="button" class="atm-btn-primary btn-xs" @click="openTransferModal">
                      划转
                    </button>
                    <button type="button" class="atm-btn-primary btn-xs" @click="openWithdrawModal">
                      提现
                    </button>
                  </div>
                </div>

                <h3 class="withdraw-section-title withdraw-section-title--sub">收款方式</h3>
                <div class="withdraw-payout-list">
                  <article class="withdraw-payout-row">
                    <span class="withdraw-payout-name">支付宝</span>
                    <p class="withdraw-payout-status muted">
                      {{ payoutQrConfigured.alipay ? '收款码已配置' : '未设置收款码' }}
                    </p>
                    <button
                      type="button"
                      class="atm-btn-primary btn-xs withdraw-payout-btn"
                      @click="openPayoutQrModal('alipay')"
                    >
                      {{ payoutQrConfigured.alipay ? '更换收款码' : '设置收款码' }}
                    </button>
                  </article>
                  <article class="withdraw-payout-row">
                    <span class="withdraw-payout-name">微信</span>
                    <p class="withdraw-payout-status muted">
                      {{ payoutQrConfigured.wechat ? '收款码已配置' : '未设置收款码' }}
                    </p>
                    <button
                      type="button"
                      class="atm-btn-primary btn-xs withdraw-payout-btn"
                      @click="openPayoutQrModal('wechat')"
                    >
                      {{ payoutQrConfigured.wechat ? '更换收款码' : '设置收款码' }}
                    </button>
                  </article>
                </div>
              </section>

              <h3 class="panel-subtitle">提现记录</h3>
              <div v-if="withdrawalRecords.length" class="table-wrap">
                <table class="data-table">
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
                          class="tag"
                          :class="{
                            'tag--active': w.status === 'completed',
                            'tag--pending_payment': w.status === 'pending',
                            'tag--failed': w.status === 'failed',
                          }"
                        >
                          {{ withdrawalStatusLabel[w.status as keyof typeof withdrawalStatusLabel] ?? w.status }}
                        </span>
                      </td>
                      <td class="muted">{{ w.created_at }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
              <p v-else-if="withdrawalsLoaded" class="empty">暂无提现记录。</p>
              <div v-if="withdrawalsTotal > withdrawalsPageSize" class="orders-pagination">
                <el-pagination
                  v-model:current-page="withdrawalsPage"
                  :page-size="withdrawalsPageSize"
                  :total="withdrawalsTotal"
                  layout="total, prev, pager, next"
                  background
                  @current-change="onWithdrawalsPageChange"
                />
              </div>
            </div>
            </div>

            <!-- 账户设置 -->
            <div v-else-if="activeTab === 'settings'" class="panel-body">
            <dl class="settings-list">
              <div class="settings-row">
                <dt>登录账号</dt>
                <dd>{{ userAccountLabel(user) }}</dd>
              </div>
              <div class="settings-row">
                <dt>昵称</dt>
                <dd>{{ user.nickname || '未设置' }}</dd>
              </div>
              <div class="settings-row">
                <dt>密码</dt>
                <dd>
                  <button type="button" class="link-btn" @click="changePasswordOpen = true">
                    修改密码
                  </button>
                </dd>
              </div>
            </dl>
            <div class="settings-actions">
              <button type="button" class="atm-btn-primary btn-xs" @click="logout">退出登录</button>
            </div>
            </div>
          </section>
        </main>
      </div>
    </div>

    <CatalogPickerModal
      v-model:open="catalogOpen"
      :products-category-id="catalogProductsCategoryId"
      :upgrade-baseline-limit-tokens="catalogUpgradeBaselineLimitTokens"
      @buy="onCatalogBuy"
    />
    <InvoiceConfigModal v-model:open="invoiceConfigOpen" />
    <ChangePasswordModal v-model:open="changePasswordOpen" />
    <PurchaseModal
      v-model:open="purchaseOpen"
      :product="purchaseProduct"
      :user="user"
      :order-type="purchaseOrderType"
      :user-subscription-id="purchaseUserSubscriptionId"
      :subscription-period-end="purchaseSubscriptionPeriodEnd"
      :subscription-expires-at="purchaseSubscriptionExpiresAt"
      @paid="onPurchasePaid"
    />

    <Teleport to="body">
      <div
        v-if="teamInviteOpen"
        class="team-invite-backdrop"
        @click.self="teamInviteOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="team-invite-title">
          <header class="team-invite-head">
            <h3 id="team-invite-title">邀请成员</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="teamInviteOpen = false"
            >
              ×
            </button>
          </header>
          <div class="team-invite-field">
            <span class="metric-label">专属推广域名</span>
            <div class="team-invite-row">
              <input
                class="team-invite-input"
                type="text"
                readonly
                :value="exclusivePromoDomain"
                aria-label="专属推广域名"
              />
              <button
                type="button"
                class="atm-btn-primary btn-xs"
                @click.stop="copyTeamInviteLink"
              >
                复制链接
              </button>
            </div>
          </div>
        </div>
      </div>

      <div
        v-if="teamEnterpriseInquiryOpen"
        class="team-invite-backdrop"
        @click.self="teamEnterpriseInquiryOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="team-enterprise-inquiry-title">
          <header class="team-invite-head">
            <h3 id="team-enterprise-inquiry-title">完善企业信息</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="teamEnterpriseInquiryOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            添加 API 团队成员前，请先填写企业/团队名称，便于团队管理与后续企业服务。
          </p>
          <label class="team-invite-field">
            <span class="metric-label">
              企业 / 团队名称
              <span class="field-required" aria-hidden="true">*</span>
            </span>
            <input
              v-model="teamEnterpriseInquiryForm.company_name"
              class="team-invite-input"
              type="text"
              maxlength="256"
              placeholder="公司或团队全称"
              autocomplete="organization"
            />
          </label>
          <div class="assign-subkey-actions">
            <button
              type="button"
              class="atm-btn-ghost btn-xs"
              @click="teamEnterpriseInquiryOpen = false"
            >
              取消
            </button>
            <button type="button" class="atm-btn-primary btn-xs" @click="confirmTeamEnterpriseInquiry">
              保存并继续
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="addTeamMemberOpen"
        class="team-invite-backdrop"
        @click.self="addTeamMemberOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="add-team-member-title">
          <header class="team-invite-head">
            <h3 id="add-team-member-title">添加成员</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="addTeamMemberOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            选择已通过邀请链接注册的用户、且尚未加入团队。
          </p>
          <label class="team-invite-field">
            <select v-model.number="addTeamMemberInvitedId" class="member-select">
              <option v-for="u in addableInvitees" :key="u.user_id" :value="u.user_id">
                {{ u.nickname }}（{{ u.email }}）
              </option>
            </select>
          </label>
          <div class="assign-subkey-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="addTeamMemberOpen = false">
              取消
            </button>
            <button
              type="button"
              class="atm-btn-primary btn-xs"
              :disabled="addTeamMemberSubmitting"
              @click="confirmAddTeamMember"
            >
              {{ addTeamMemberSubmitting ? '提交中…' : '确认添加' }}
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="payoutQrSetupOpen"
        class="team-invite-backdrop"
        @click.self="closePayoutQrModal"
      >
        <div class="team-invite-panel" role="dialog" :aria-labelledby="'payout-qr-title'">
          <header class="team-invite-head">
            <h3 :id="'payout-qr-title'">{{ payoutQrSetupTitle }}</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="closePayoutQrModal"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            上传 PNG / JPEG / WebP 收款码（不超过 2MB）。已有收款码可直接更换图片覆盖保存。
          </p>
          <div v-if="payoutQrSetupPreview" class="payout-qr-preview-wrap">
            <img :src="payoutQrSetupPreview" alt="收款码预览" class="payout-qr-preview" />
          </div>
          <label class="team-invite-field">
            <span class="metric-label">选择图片</span>
            <input
              type="file"
              accept="image/png,image/jpeg,image/webp"
              class="team-invite-input"
              @change="onPayoutQrFileChange"
            />
          </label>
          <div class="assign-subkey-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="closePayoutQrModal">
              取消
            </button>
            <button type="button" class="atm-btn-primary btn-xs" @click="confirmPayoutQrSetup">
              保存
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="transferOpen"
        class="team-invite-backdrop"
        @click.self="transferOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="transfer-title">
          <header class="team-invite-head">
            <h3 id="transfer-title">划转到余额</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="transferOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            将佣金转入账户余额，可用于购买套餐；当前可划转 {{ formatCny(commissionAvailable) }}。
          </p>
          <label class="team-invite-field">
            <span class="metric-label">划转金额（元）</span>
            <input
              v-model="transferForm.amountYuan"
              type="number"
              min="0.01"
              step="0.01"
              class="team-invite-input"
              placeholder="例如 50.00"
            />
          </label>
          <div class="assign-subkey-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="transferOpen = false">
              取消
            </button>
            <button type="button" class="atm-btn-primary btn-xs" @click="confirmCommissionTransfer">
              确认划转
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="withdrawOpen"
        class="team-invite-backdrop"
        @click.self="withdrawOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="withdraw-title">
          <header class="team-invite-head">
            <h3 id="withdraw-title">提现</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="withdrawOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            可提现 {{ formatCny(commissionAvailable) }}，提现将打款至所选渠道的收款码。
          </p>
          <label class="team-invite-field">
            <span class="metric-label">到账方式</span>
            <select v-model="withdrawForm.channel" class="member-select">
              <option v-if="payoutQrConfigured.alipay" value="alipay">支付宝</option>
              <option v-if="payoutQrConfigured.wechat" value="wechat">微信</option>
            </select>
          </label>
          <label class="team-invite-field">
            <span class="metric-label">提现金额（元）</span>
            <input
              v-model="withdrawForm.amountYuan"
              type="number"
              min="0.01"
              step="0.01"
              class="team-invite-input"
              placeholder="例如 20.00"
            />
          </label>
          <div class="assign-subkey-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="withdrawOpen = false">
              取消
            </button>
            <button type="button" class="atm-btn-primary btn-xs" @click="confirmWithdraw">
              确认提现
            </button>
          </div>
        </div>
      </div>

      <div
        v-if="assignSubKeyOpen"
        class="team-invite-backdrop"
        @click.self="assignSubKeyOpen = false"
      >
        <div
          class="team-invite-panel team-invite-panel--assign-subkey"
          role="dialog"
          aria-labelledby="assign-subkey-title"
          @click.stop
        >
          <header class="team-invite-head">
            <h3 id="assign-subkey-title">分配子 Key</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="assignSubKeyOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            从您的套餐额度中为团队成员生成独立子 Key，调用计入该套餐，用量单独归属成员。
          </p>
          <form class="assign-subkey-form" novalidate @submit.prevent="confirmAssignSubKey">
            <label class="team-invite-field">
              <span class="metric-label">关联套餐</span>
              <select
                v-model.number="assignSubKeyForm.subscriptionId"
                class="member-select"
                @change="onAssignSubKeyPlanChange"
              >
                <option
                  v-for="s in memberSubscriptions.filter((x) => x.status === 'active')"
                  :key="s.id"
                  :value="s.id"
                >
                  {{ s.product_name }}
                </option>
              </select>
            </label>
            <label class="team-invite-field">
              <span class="metric-label">用量上限（tokens）</span>
              <input
                v-model.number="assignSubKeyForm.limitTokens"
                type="number"
                class="team-invite-input"
                min="1"
                :max="assignSubKeySubscription?.limit_tokens"
              />
              <p v-if="assignSubKeySubscription" class="field-hint">
                默认与主 Key 一致：{{ formatTokens(assignSubKeySubscription.limit_tokens) }}（本周期套餐总量）
              </p>
            </label>
            <label class="team-invite-field">
              <span class="metric-label">团队成员</span>
              <select
                v-model.number="assignSubKeyForm.memberUserId"
                class="member-select"
                :disabled="!assignSubKeyEligibleMembers.length"
              >
                <option v-for="m in assignSubKeyEligibleMembers" :key="m.id" :value="m.user_id">
                  {{ m.nickname }}（{{ m.email }}）
                </option>
              </select>
              <p v-if="!assignSubKeyEligibleMembers.length" class="field-hint">
                该套餐下可分配成员已全部拥有子 Key，请切换关联套餐或先在「我的团队」添加成员。
              </p>
            </label>
            <div class="assign-subkey-actions">
              <button
                type="button"
                class="atm-btn-ghost btn-xs"
                :disabled="assignSubKeySubmitting"
                @click="assignSubKeyOpen = false"
              >
                取消
              </button>
              <button
                type="submit"
                class="atm-btn-primary btn-xs"
                :disabled="assignSubKeySubmitting || !assignSubKeyEligibleMembers.length"
              >
                {{ assignSubKeySubmitting ? '提交中…' : '确认分配' }}
              </button>
            </div>
          </form>
        </div>
      </div>

      <div
        v-if="editSubKeyLimitOpen && editSubKeyLimitTarget"
        class="team-invite-backdrop"
        @click.self="editSubKeyLimitOpen = false"
      >
        <div class="team-invite-panel" role="dialog" aria-labelledby="edit-subkey-limit-title">
          <header class="team-invite-head">
            <h3 id="edit-subkey-limit-title">设置子 Key 用量上限</h3>
            <button
              type="button"
              class="team-invite-close"
              aria-label="关闭"
              @click="editSubKeyLimitOpen = false"
            >
              ×
            </button>
          </header>
          <p class="team-invite-lead">
            {{ editSubKeyLimitTarget.member_nickname }} ·
            {{ editSubKeyLimitTarget.subscription_name }}
          </p>
          <label class="team-invite-field">
            <span class="metric-label">用量上限（tokens）</span>
            <input
              v-model.number="editSubKeyLimitValue"
              type="number"
              class="team-invite-input"
              min="1"
              :max="editSubKeyLimitMainTotal"
            />
            <p class="field-hint">
              主 Key 套餐总量 {{ formatTokens(editSubKeyLimitMainTotal) }}；已用
              {{ formatTokens(editSubKeyLimitTarget.used_tokens) }}
            </p>
          </label>
          <div class="assign-subkey-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="editSubKeyLimitOpen = false">
              取消
            </button>
            <button type="button" class="atm-btn-primary btn-xs" @click="confirmEditSubKeyLimit">
              保存
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>

<style scoped>
.member-layout {
  background: var(--atm-bg);
  min-height: calc(100vh - 56px);
}

.member-hero {
  position: relative;
  overflow: hidden;
  padding: 44px 0 88px;
  color: #fff;
}

.member-hero-bg {
  position: absolute;
  inset: 0;
  background: linear-gradient(125deg, #1e1b4b 0%, #5b21b6 45%, #6366f1 100%);
}

.member-hero-bg::after {
  content: '';
  position: absolute;
  inset: 0;
  opacity: 0.15;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.12) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.12) 1px, transparent 1px);
  background-size: 40px 40px;
}

.member-hero-inner {
  position: relative;
  z-index: 1;
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 24px 32px;
}

.member-hero-title {
  margin: 0 0 10px;
  font-weight: 700;
  line-height: 1.4;
}

.member-hero-greeting {
  display: block;
  font-size: clamp(18px, 2.2vw, 22px);
  letter-spacing: -0.02em;
}

.member-hero-account {
  display: block;
  margin-top: 4px;
  max-width: min(100%, 520px);
  font-size: clamp(14px, 1.6vw, 16px);
  font-weight: 600;
  letter-spacing: 0;
  opacity: 0.92;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.member-hero-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.hero-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 12px 26px;
  font-size: 15px;
  font-weight: 600;
  text-decoration: none;
  border-radius: 999px;
  cursor: pointer;
  border: none;
  transition:
    transform 0.15s,
    box-shadow 0.15s;
}

.hero-btn--light {
  color: var(--atm-primary);
  background: #fff;
  box-shadow: 0 8px 24px rgba(0, 0, 0, 0.15);
}

.hero-btn--light:hover {
  transform: translateY(-1px);
}

.hero-btn--ghost {
  color: #fff;
  background: rgba(255, 255, 255, 0.1);
  border: 1px solid rgba(255, 255, 255, 0.35);
}

.member-body {
  position: relative;
  z-index: 2;
  margin-top: -52px;
  padding: 0 0 64px;
}

.member-grid {
  display: flex;
  flex-direction: column;
  gap: 28px;
}

@media (min-width: 960px) {
  .member-grid {
    flex-direction: row;
    align-items: stretch;
  }

  .member-sidebar-wrap {
    flex-shrink: 0;
    width: 272px;
    align-self: stretch;
  }
}

.member-main {
  flex: 1;
  min-width: 0;
}

.panel-head {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 16px;
  margin: 0 0 24px;
}

.panel-head--segmented {
  margin-bottom: 0;
  gap: 12px;
}

.panel-body--segmented {
  margin-top: 12px;
}

.panel-head-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  width: 100%;
}

.panel-head-title {
  margin: 0;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.panel-head-action {
  flex-shrink: 0;
  font-size: 14px;
  font-weight: 600;
}

.panel-head-action--hidden {
  visibility: hidden;
  pointer-events: none;
}

.panel-body {
  min-height: 200px;
}

.panel {
  padding: 32px 28px;
  background: #fff;
  border: 1px solid rgba(237, 233, 254, 0.9);
  border-radius: 20px;
  box-shadow:
    0 4px 24px rgba(124, 58, 237, 0.06),
    0 16px 48px rgba(30, 27, 75, 0.06);
}

.panel-desc {
  margin: 0 0 20px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--atm-text-muted);
}

.panel-desc a {
  color: var(--atm-primary);
  font-weight: 500;
}

.panel-subtitle {
  margin: 28px 0 14px;
  font-size: 17px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.panel-subtitle:first-child {
  margin-top: 0;
}

.panel-subtitle--tight {
  margin-top: 0;
}

.withdraw-section-box {
  margin-bottom: 20px;
  padding: 20px 22px;
  background: linear-gradient(135deg, #f5f3ff 0%, #faf5ff 55%, #fff 100%);
  border: 1px solid rgba(124, 58, 237, 0.12);
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.06);
}

.withdraw-section-title {
  margin: 0 0 14px;
  font-size: 15px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.withdraw-section-title--sub {
  margin-top: 22px;
  padding-top: 20px;
  border-top: 1px solid rgba(124, 58, 237, 0.1);
}

.withdraw-toolbar {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  width: 100%;
}

.withdraw-toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 8px;
}

.withdraw-balance {
  display: inline-flex;
  flex-wrap: wrap;
  align-items: baseline;
  gap: 6px;
}

.withdraw-balance strong {
  margin-left: 0;
  font-size: 18px;
  font-weight: 800;
  color: var(--atm-text);
}

.withdraw-payout-list {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.withdraw-payout-row {
  display: flex;
  flex-direction: row;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px 16px;
}

.withdraw-payout-name {
  flex-shrink: 0;
  min-width: 56px;
  font-size: 15px;
  font-weight: 600;
  color: var(--atm-text-muted);
}

.withdraw-payout-status {
  flex: 1;
  min-width: 0;
  margin: 0;
  font-size: 12px;
  font-weight: 400;
  line-height: 1.5;
  color: var(--atm-text-muted);
}

.withdraw-payout-btn {
  flex-shrink: 0;
  margin-left: auto;
}

.withdraw-qr-preview {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.withdraw-qr-preview img {
  width: 120px;
  height: 120px;
  object-fit: contain;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
}


.stat-row {
  display: grid;
  gap: 16px;
  padding: 4px 0;
}

@media (min-width: 640px) {
  .stat-row {
    grid-template-columns: repeat(2, 1fr);
    gap: 20px;
  }
}

.stat-card {
  display: flex;
  flex-direction: row;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  min-height: 0;
  padding: 16px 18px;
  background: #fff;
  border: 1px solid rgba(124, 58, 237, 0.1);
  border-radius: 16px;
  box-shadow: 0 6px 24px rgba(124, 58, 237, 0.06);
  transition:
    background 0.15s ease,
    box-shadow 0.15s ease;
}

.stat-card--balance,
.stat-card--commission {
  background: linear-gradient(160deg, #faf5ff 0%, #fff 70%);
}

.panel-tab-toolbar {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-start;
  gap: 10px;
  margin-bottom: 16px;
}

.rebate-lead {
  margin: 0 0 16px;
  font-size: 13px;
  line-height: 1.5;
}

.rebate-rate-hero {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  justify-content: space-between;
  gap: 20px 24px;
  margin-bottom: 8px;
  padding: 20px 22px;
  background: linear-gradient(135deg, #f5f3ff 0%, #faf5ff 55%, #fff 100%);
  border: 1px solid rgba(124, 58, 237, 0.12);
  border-radius: 16px;
}

.rebate-details-summary {
  margin-bottom: 8px;
  padding: 20px 22px;
  background: linear-gradient(135deg, #f5f3ff 0%, #faf5ff 55%, #fff 100%);
  border: 1px solid rgba(124, 58, 237, 0.12);
  border-radius: 16px;
  box-shadow: 0 4px 16px rgba(124, 58, 237, 0.06);
}

.rebate-details-section + .rebate-details-section {
  margin-top: 20px;
  padding-top: 20px;
  border-top: 1px solid rgba(124, 58, 237, 0.1);
}

.rebate-details-metrics-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16px 32px;
  align-items: flex-start;
  justify-content: space-between;
}

.rebate-details-metric {
  flex: 1 1 0;
  min-width: min(100%, 160px);
}

.promo-domain-row {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px 16px;
  margin-top: 8px;
}

.promo-domain-value {
  flex: 1;
  min-width: 0;
  margin: 0;
  padding: 0;
  font-size: 18px;
  font-weight: 600;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: var(--atm-primary-dark);
  word-break: break-all;
  background: transparent;
  border: none;
}

.rebate-rate-value {
  margin: 6px 0 0;
  font-size: 32px;
  font-weight: 800;
  line-height: 1.15;
  letter-spacing: -0.02em;
  color: var(--atm-primary-dark);
}

.rebate-rate-unit {
  margin-left: 2px;
  font-size: 18px;
  font-weight: 700;
}

.rebate-rate-meta {
  margin: 8px 0 0;
  font-size: 13px;
  color: var(--atm-text-muted);
}

.rebate-rate-next {
  max-width: 280px;
  padding: 12px 14px;
  background: #fff;
  border: 1px solid rgba(124, 58, 237, 0.1);
  border-radius: 12px;
}

.rebate-rate-next strong {
  display: block;
  margin-top: 4px;
  font-size: 13px;
  line-height: 1.5;
  color: var(--atm-text);
}

.rebate-tier-row--current {
  background: rgba(237, 233, 254, 0.35);
}

.rebate-tier-badge {
  margin-left: 8px;
  vertical-align: middle;
}

.rebate-notes {
  margin: 0;
  padding-left: 1.25rem;
  font-size: 14px;
  line-height: 1.65;
  color: var(--atm-text-muted);
}

.rebate-notes li + li {
  margin-top: 8px;
}

.stat-card:hover {
  background: #fdfcff;
  box-shadow: 0 10px 28px rgba(124, 58, 237, 0.1);
}

.stat-card--balance:hover,
.stat-card--commission:hover {
  background: linear-gradient(160deg, #f3ebff 0%, #fff 70%);
}

.stat-main {
  flex: 1;
  min-width: 0;
}

.stat-label {
  display: block;
  font-size: 13px;
  font-weight: 500;
  color: var(--atm-text-muted);
}

.stat-value {
  display: block;
  margin-top: 6px;
  font-size: 28px;
  font-weight: 800;
  font-variant-numeric: tabular-nums;
  letter-spacing: -0.03em;
  line-height: 1.15;
  color: var(--atm-text);
}

.stat-card--balance .stat-value,
.stat-card--commission .stat-value {
  color: var(--atm-primary-dark);
}

.stat-card-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-shrink: 0;
}

.stat-link {
  flex-shrink: 0;
  padding: 7px 15px;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.2;
  color: var(--atm-primary-dark);
  background: var(--atm-primary-light);
  border: none;
  border-radius: 999px;
  cursor: pointer;
  text-decoration: none;
  transition:
    background 0.15s ease,
    color 0.15s ease;
}

.stat-link:hover {
  color: #fff;
  background: var(--atm-primary);
}

.plan-mini-list {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.plan-mini {
  padding: 18px 20px;
  background: #f8fafc;
  border: 1px solid #f1f5f9;
  border-radius: 14px;
}

.plan-mini-head {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 10px;
  font-size: 14px;
}

.plan-mini-head strong {
  color: var(--atm-text);
}

.plan-mini-head span {
  font-size: 12px;
  color: var(--atm-text-muted);
}

.progress-track {
  height: 8px;
  overflow: hidden;
  background: #e2e8f0;
  border-radius: 999px;
}

.progress-track--lg {
  height: 10px;
  margin-top: 16px;
}

.progress-track--inactive {
  background: #cbd5e1;
}

.progress-track--sm {
  height: 6px;
  margin-top: 6px;
}

.subkey-usage {
  min-width: 140px;
}

.subkey-usage-text {
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-text);
  white-space: nowrap;
}

.subkey-row-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  white-space: nowrap;
}

.field-hint {
  margin: 8px 0 0;
  font-size: 12px;
  line-height: 1.5;
  color: var(--atm-text-muted);
}

.progress-fill {
  height: 100%;
  background: var(--atm-gradient);
  border-radius: 999px;
}

.progress-meta {
  margin: 8px 0 0;
  font-size: 12px;
  color: var(--atm-text-muted);
}

.plan-card {
  padding: 28px 26px;
  margin-bottom: 20px;
  border: 1px solid #ddd6fe;
  border-radius: 18px;
  background: linear-gradient(165deg, #f5f3ff 0%, #fff 42%);
  box-shadow: 0 8px 28px rgba(124, 58, 237, 0.08);
}

.plan-card:last-child {
  margin-bottom: 0;
}

.plan-card--inactive {
  border-color: #e2e8f0;
  background: #f8fafc;
  box-shadow: none;
}

.plan-card--inactive .plan-card-head h2,
.plan-card--inactive .plan-metrics strong {
  color: #64748b;
}

.plan-card-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
  margin-bottom: 16px;
}

.plan-card-head h2 {
  margin: 0 0 4px;
  font-size: 18px;
  font-weight: 700;
  color: var(--atm-text);
}

.plan-card-tags {
  display: flex;
  flex-shrink: 0;
  flex-wrap: wrap;
  gap: 8px;
  align-items: center;
  justify-content: flex-end;
}

.plan-card-head .tag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  padding: 6px 14px;
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
}

.sku {
  font-size: 12px;
  color: var(--atm-text-muted);
}

.plan-metrics {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px 20px;
}

@media (min-width: 640px) {
  .plan-metrics {
    grid-template-columns: repeat(4, 1fr);
  }
}

.metric-label {
  display: block;
  font-size: 12px;
  color: var(--atm-text-muted);
  margin-bottom: 4px;
}

.plan-metrics strong {
  font-size: 14px;
  color: var(--atm-text);
}

.plan-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  align-items: center;
  margin-top: 16px;
}

.orders-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 20px;
}

.overview-orders-more {
  margin-top: 12px;
  font-size: 13px;
}

.overview-orders-more a {
  color: var(--atm-primary, #7c3aed);
  text-decoration: none;
}

.overview-orders-more a:hover {
  text-decoration: underline;
}

.table-wrap {
  overflow-x: auto;
  margin: 0 -4px;
}

.table-wrap--spaced {
  margin-top: 16px;
}

.data-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.data-table th {
  padding: 14px 14px;
  text-align: left;
  font-weight: 600;
  color: var(--atm-text-muted);
  border-bottom: 1px solid #e2e8f0;
  white-space: nowrap;
}

.data-table td {
  padding: 14px;
  border-bottom: 1px solid #f1f5f9;
  color: var(--atm-text);
}

.mono {
  font-family: ui-monospace, monospace;
  font-size: 12px;
}

.muted {
  color: var(--atm-text-muted);
  white-space: nowrap;
}

.tag {
  display: inline-block;
  padding: 3px 8px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 999px;
  background: #f1f5f9;
  color: #64748b;
}

.tag--key-main {
  background: #ede9fe;
  color: #5b21b6;
}

.tag--key-team {
  background: #e0f2fe;
  color: #0369a1;
}

.tag--active,
.tag--completed {
  background: #dcfce7;
  color: #15803d;
}

.tag--pending_payment {
  background: #fef3c7;
  color: #b45309;
}

.tag--used {
  background: #e2e8f0;
  color: #475569;
}

.tag--expired {
  background: #f1f5f9;
  color: #94a3b8;
}

.tag--failed,
.tag--cancelled {
  background: #fee2e2;
  color: #b91c1c;
}

.amount-plus {
  color: #15803d;
  font-weight: 600;
}

.amount-minus {
  color: #b91c1c;
  font-weight: 600;
}

.link-btn {
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-primary);
  background: none;
  border: none;
  cursor: pointer;
  padding: 0;
}

.link-btn:hover {
  text-decoration: underline;
}

.btn-xs {
  padding: 8px 16px;
  font-size: 13px;
}

.settings-list {
  margin: 0 0 28px;
}

.settings-row {
  display: grid;
  grid-template-columns: 100px 1fr;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid #f1f5f9;
  font-size: 14px;
}

.settings-row dt {
  color: var(--atm-text-muted);
}

.settings-row dd {
  margin: 0;
  color: var(--atm-text);
}

.settings-list .link-btn {
  text-decoration: underline;
  text-underline-offset: 3px;
}

.settings-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
}

.empty {
  margin: 0;
  font-size: 14px;
  color: var(--atm-text-muted);
}

.api-key-tabs {
  display: inline-flex;
  gap: 4px;
  width: fit-content;
  padding: 4px;
  background: #f1f5f9;
  border-radius: 12px;
}

.api-key-tab {
  padding: 8px 18px;
  font-size: 13px;
  font-weight: 600;
  font-family: inherit;
  color: var(--atm-text-muted);
  background: transparent;
  border: none;
  border-radius: 9px;
  cursor: pointer;
  transition:
    background 0.15s ease,
    color 0.15s ease,
    box-shadow 0.15s ease;
}

.api-key-tab--active {
  color: var(--atm-primary-dark);
  background: #fff;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.12);
}

.cell-sub {
  display: block;
  margin-top: 2px;
  font-size: 12px;
  font-weight: 400;
}

.member-select {
  display: block;
  width: 100%;
  margin-top: 8px;
  padding: 10px 12px;
  font-size: 14px;
  font-family: inherit;
  color: var(--atm-text);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.assign-subkey-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
}

.team-invite-field + .team-invite-field {
  margin-top: 14px;
}

.assign-subkey-form {
  margin: 0;
}

.team-invite-backdrop {
  position: fixed;
  inset: 0;
  z-index: 3000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px 16px;
  background: rgba(15, 10, 40, 0.52);
  backdrop-filter: blur(4px);
}

.team-invite-panel {
  width: 100%;
  max-width: 520px;
  padding: 22px 24px 24px;
  background: #fff;
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(30, 27, 75, 0.22);
}

.team-invite-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.team-invite-head h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 800;
  color: var(--atm-text);
}

.team-invite-close {
  width: 36px;
  height: 36px;
  font-size: 22px;
  line-height: 1;
  color: var(--atm-text-muted);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
}

.team-invite-lead {
  margin: 0 0 18px;
  font-size: 14px;
  line-height: 1.6;
  color: var(--atm-text-muted);
}

.team-invite-field {
  display: block;
}

.field-required {
  margin-left: 2px;
  color: #e53935;
  font-weight: 600;
}

.team-invite-row {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-top: 8px;
}

.team-invite-input {
  flex: 1;
  min-width: 200px;
  padding: 10px 12px;
  font-size: 13px;
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  color: var(--atm-text);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.payout-qr-preview-wrap {
  display: flex;
  justify-content: center;
  margin-bottom: 16px;
  padding: 12px;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 12px;
}

.payout-qr-preview {
  display: block;
  max-width: 220px;
  max-height: 220px;
  object-fit: contain;
  border-radius: 8px;
}
</style>
