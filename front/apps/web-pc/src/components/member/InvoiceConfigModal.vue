<script setup lang="ts">
import { computed, reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import type {
  EnterpriseInvoiceLookupItem,
  InvoiceConfigItem,
  SaveInvoiceConfigBody,
} from '@ai-token-mall/shared'
import { invoiceApi } from '@/api'
import {
  beginRouteNavigationLoading,
  endRouteNavigationLoading,
} from '@/composables/useRouteNavigationLoading'

const MESSAGE_Z_INDEX = 10_000

function toastSuccess(message: string) {
  ElMessage.success({ message, zIndex: MESSAGE_Z_INDEX })
}

function toastError(message: string) {
  ElMessage.error({ message, zIndex: MESSAGE_Z_INDEX })
}

function toastWarning(message: string) {
  ElMessage.warning({ message, zIndex: MESSAGE_Z_INDEX })
}

function toastInfo(message: string) {
  ElMessage.info({ message, zIndex: MESSAGE_Z_INDEX })
}

const open = defineModel<boolean>('open', { default: false })

type PanelTab = 'list' | 'form'
const activeTab = ref<PanelTab>('list')

const configs = ref<InvoiceConfigItem[]>([])
const editingId = ref<number | null>(null)
const saving = ref(false)
const defaultSettingId = ref<number | null>(null)
const lookupLoading = ref(false)
const lookupCandidates = ref<EnterpriseInvoiceLookupItem[]>([])
const lastLookupKeyword = ref('')

const emptyForm = (): SaveInvoiceConfigBody & { profile_type: 'enterprise' | 'personal' } => ({
  profile_type: 'enterprise',
  title: '',
  tax_no: '',
  bank_name: '',
  bank_account: '',
  address: '',
  phone: '',
  is_default: true,
})

const form = reactive(emptyForm())

const isEnterprise = computed(() => form.profile_type === 'enterprise')

function isDefaultRow(row: InvoiceConfigItem): boolean {
  const v = row.is_default as unknown
  return v === true || v === 1 || v === '1'
}

function configToSaveBody(row: InvoiceConfigItem, isDefault: boolean): SaveInvoiceConfigBody {
  const pt = row.profile_type === 'personal' ? 'personal' : 'enterprise'
  return {
    profile_type: pt,
    title: row.title,
    tax_no: row.tax_no || undefined,
    bank_name: row.bank_name || undefined,
    bank_account: row.bank_account || undefined,
    address: row.address || undefined,
    phone: row.phone || undefined,
    is_default: isDefault,
  }
}

function normalizeConfigItem(row: InvoiceConfigItem): InvoiceConfigItem {
  return { ...row, is_default: isDefaultRow(row) }
}

function resetForm() {
  Object.assign(form, emptyForm())
  editingId.value = null
  lookupCandidates.value = []
  lastLookupKeyword.value = ''
}

async function loadConfigs() {
  beginRouteNavigationLoading()
  try {
    const rows = await invoiceApi.listConfigs()
    configs.value = rows.map(normalizeConfigItem)
  } catch (e) {
    toastError(e instanceof Error ? e.message : '加载发票抬头失败')
  } finally {
    endRouteNavigationLoading()
  }
}

function openFormForCreate() {
  resetForm()
  if (!configs.value.length) {
    form.is_default = true
  }
  activeTab.value = 'form'
}

function startEdit(row: InvoiceConfigItem) {
  editingId.value = row.id
  form.profile_type = row.profile_type === 'personal' ? 'personal' : 'enterprise'
  form.title = row.title
  form.tax_no = row.tax_no ?? ''
  form.bank_name = row.bank_name ?? ''
  form.bank_account = row.bank_account ?? ''
  form.address = row.address ?? ''
  form.phone = row.phone ?? ''
  form.is_default = row.is_default
  lookupCandidates.value = []
  lastLookupKeyword.value = ''
  activeTab.value = 'form'
}

function goListTab() {
  if (activeTab.value === 'list') return
  activeTab.value = 'list'
  resetForm()
}

function onFormTabClick() {
  if (activeTab.value === 'form') return
  openFormForCreate()
}

function applyLookupItem(item: EnterpriseInvoiceLookupItem) {
  form.title = item.title
  form.tax_no = item.tax_no ?? form.tax_no
  form.bank_name = item.bank_name ?? ''
  form.bank_account = item.bank_account ?? ''
  form.address = item.address ?? ''
  form.phone = item.phone ?? ''
}

function lookupKeyword(): string {
  const name = form.title?.trim() ?? ''
  const tax = form.tax_no?.trim() ?? ''
  return name || tax
}

async function runEnterpriseLookup(silent = false) {
  const keyword = lookupKeyword()
  if (!keyword) {
    if (!silent) {
      toastWarning('请先填写公司名称或公司税号（至少一项）')
    }
    return
  }
  if (lookupLoading.value || keyword === lastLookupKeyword.value) {
    return
  }
  lookupLoading.value = true
  lookupCandidates.value = []
  beginRouteNavigationLoading()
  try {
    const res = await invoiceApi.lookupEnterprise(keyword)
    lastLookupKeyword.value = keyword
    applyLookupItem(res.match)
    if (res.candidates?.length) {
      lookupCandidates.value = res.candidates
      toastInfo('有多条匹配结果，已选最相近的一条，可点击下方切换')
    }
  } catch (e) {
    if (!silent) {
      toastError(e instanceof Error ? e.message : '查询失败，请核对名称或税号')
    }
  } finally {
    lookupLoading.value = false
    endRouteNavigationLoading()
  }
}

function onEnterpriseLookupBlur() {
  if (!isEnterprise.value) return
  void runEnterpriseLookup(true)
}

function bodyFromForm(): SaveInvoiceConfigBody {
  return {
    profile_type: form.profile_type,
    title: form.title.trim(),
    tax_no: form.tax_no?.trim() || undefined,
    bank_name: form.bank_name?.trim() || undefined,
    bank_account: form.bank_account?.trim() || undefined,
    address: form.address?.trim() || undefined,
    phone: form.phone?.trim() || undefined,
    is_default: form.is_default,
  }
}

async function save() {
  if (form.profile_type === 'personal') {
    if (!form.title.trim()) {
      toastWarning('请填写发票抬头')
      return
    }
  } else {
    if (!form.title.trim() && !form.tax_no?.trim()) {
      toastWarning('请填写公司名称或公司税号（至少一项），建议先查询补全')
      return
    }
  }
  saving.value = true
  beginRouteNavigationLoading()
  try {
    const body = bodyFromForm()
    if (editingId.value) {
      await invoiceApi.updateConfig(editingId.value, body)
      toastSuccess('已保存')
    } else {
      await invoiceApi.createConfig(body)
      toastSuccess('已添加')
    }
    await loadConfigs()
    resetForm()
    activeTab.value = 'list'
  } catch (e) {
    toastError(e instanceof Error ? e.message : '保存失败')
  } finally {
    saving.value = false
    endRouteNavigationLoading()
  }
}

async function setAsDefault(row: InvoiceConfigItem) {
  if (isDefaultRow(row)) return
  defaultSettingId.value = row.id
  beginRouteNavigationLoading()
  try {
    const updated = await invoiceApi.setDefaultConfig(row.id, configToSaveBody(row, true))
    const normalized = normalizeConfigItem(updated)
    configs.value = configs.value.map((c) =>
      c.id === row.id ? normalized : { ...c, is_default: false },
    )
    toastSuccess('已设为默认抬头')
  } catch (e) {
    toastError(e instanceof Error ? e.message : '操作失败')
  } finally {
    defaultSettingId.value = null
    endRouteNavigationLoading()
  }
}

function close() {
  open.value = false
}

watch(open, (v) => {
  if (v) {
    activeTab.value = 'list'
    void loadConfigs()
    resetForm()
  }
})

</script>

<template>
  <Teleport to="body">
    <div v-if="open" class="icm-backdrop" @click.self="close">
      <div class="icm-panel" role="dialog" aria-labelledby="invoice-config-title">
        <header class="icm-head">
          <h3 id="invoice-config-title">发票抬头</h3>
          <button type="button" class="icm-close" aria-label="关闭" @click="close">×</button>
        </header>

        <div class="icm-tabs" role="tablist" aria-label="发票抬头">
          <button
            type="button"
            class="icm-tab"
            :class="{ 'icm-tab--active': activeTab === 'list' }"
            role="tab"
            :aria-selected="activeTab === 'list'"
            @click="goListTab"
          >
            抬头列表
          </button>
          <button
            type="button"
            class="icm-tab"
            :class="{ 'icm-tab--active': activeTab === 'form' }"
            role="tab"
            :aria-selected="activeTab === 'form'"
            @click="onFormTabClick"
          >
            添加抬头
          </button>
        </div>

        <div v-if="activeTab === 'list'" class="icm-tab-panel" role="tabpanel">
          <div class="icm-table-wrap">
            <table class="icm-table">
              <thead>
                <tr>
                  <th>抬头</th>
                  <th>税号</th>
                  <th>默认</th>
                  <th />
                </tr>
              </thead>
              <tbody>
                <tr v-if="!configs.length">
                  <td colspan="4" class="icm-empty">暂无抬头</td>
                </tr>
                <tr v-for="row in configs" :key="row.id">
                  <td>{{ row.title }}</td>
                  <td class="icm-mono">{{ row.tax_no || '—' }}</td>
                  <td>{{ isDefaultRow(row) ? '是' : '—' }}</td>
                  <td class="icm-row-actions">
                    <button type="button" class="icm-link" @click="startEdit(row)">编辑</button>
                    <button
                      v-if="!isDefaultRow(row)"
                      type="button"
                      class="icm-link"
                      :disabled="defaultSettingId === row.id"
                      @click.stop="setAsDefault(row)"
                    >
                      {{ defaultSettingId === row.id ? '设置中…' : '设为默认' }}
                    </button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <div class="icm-actions icm-actions--list">
            <button type="button" class="atm-btn-ghost btn-xs" @click="close">关闭</button>
          </div>
        </div>

        <form v-else class="icm-form" role="tabpanel" @submit.prevent="save">
          <div class="icm-form-stack">
            <label class="icm-field">
              <span class="icm-label">类型</span>
              <select v-model="form.profile_type" class="icm-input">
                <option value="enterprise">企业</option>
                <option value="personal">个人</option>
              </select>
            </label>

            <template v-if="isEnterprise">
              <label class="icm-field">
                <span class="icm-label">公司名称 <span class="icm-req">*</span></span>
                <input
                  v-model="form.title"
                  class="icm-input"
                  type="text"
                  maxlength="256"
                  @blur="onEnterpriseLookupBlur"
                />
              </label>
              <label class="icm-field">
                <span class="icm-label">公司税号 <span class="icm-req">*</span></span>
                <input
                  v-model="form.tax_no"
                  class="icm-input"
                  type="text"
                  maxlength="64"
                  @blur="onEnterpriseLookupBlur"
                />
              </label>
              <label class="icm-field">
                <span class="icm-label">开户银行</span>
                <input v-model="form.bank_name" class="icm-input" type="text" maxlength="128" />
              </label>
              <label class="icm-field">
                <span class="icm-label">银行账号</span>
                <input v-model="form.bank_account" class="icm-input" type="text" maxlength="64" />
              </label>
              <label class="icm-field">
                <span class="icm-label">地址</span>
                <input v-model="form.address" class="icm-input" type="text" maxlength="512" />
              </label>
              <label class="icm-field">
                <span class="icm-label">电话</span>
                <input v-model="form.phone" class="icm-input" type="text" maxlength="32" />
              </label>
              <div v-if="lookupCandidates.length" class="icm-candidates">
                <p class="icm-label">其他匹配结果</p>
                <button
                  v-for="(c, idx) in lookupCandidates"
                  :key="idx"
                  type="button"
                  class="icm-candidate"
                  @click="applyLookupItem(c)"
                >
                  {{ c.title }} · {{ c.tax_no }}
                </button>
              </div>
            </template>

            <template v-else>
              <label class="icm-field">
                <span class="icm-label">发票抬头 <span class="icm-req">*</span></span>
                <input
                  v-model="form.title"
                  class="icm-input"
                  type="text"
                  maxlength="256"
                  placeholder="个人姓名"
                />
              </label>
            </template>
          </div>

          <label class="icm-check">
            <input v-model="form.is_default" type="checkbox" />
            设为默认抬头
          </label>
          <div class="icm-actions">
            <button type="button" class="atm-btn-ghost btn-xs" @click="goListTab">返回列表</button>
            <button type="submit" class="atm-btn-primary btn-xs" :disabled="saving">
              {{ saving ? '保存中…' : editingId ? '保存' : '添加' }}
            </button>
          </div>
        </form>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.icm-backdrop {
  position: fixed;
  inset: 0;
  z-index: 2050;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px 16px;
  background: rgba(15, 10, 40, 0.52);
  backdrop-filter: blur(4px);
}

.icm-panel {
  width: 100%;
  max-width: min(720px, 96vw);
  max-height: min(90vh, 720px);
  overflow-y: auto;
  padding: 22px 24px 24px;
  background: #fff;
  border-radius: 18px;
  box-shadow: 0 24px 64px rgba(30, 27, 75, 0.22);
}

.icm-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.icm-head h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 800;
  color: var(--atm-text);
}

.icm-close {
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

.icm-tabs {
  display: inline-flex;
  gap: 4px;
  width: fit-content;
  max-width: 100%;
  margin-bottom: 16px;
  padding: 4px;
  background: #f1f5f9;
  border-radius: 12px;
}

.icm-tab {
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

.icm-tab--active {
  color: var(--atm-primary-dark, #6d28d9);
  background: #fff;
  box-shadow: 0 2px 8px rgba(124, 58, 237, 0.12);
}

.icm-tab-panel {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.icm-empty {
  padding: 24px 12px;
  text-align: center;
  font-size: 14px;
  color: var(--atm-text-muted);
}

.icm-req {
  color: #dc2626;
}

.icm-table-wrap {
  overflow-x: auto;
}

.icm-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.icm-table th {
  padding: 10px 12px;
  text-align: left;
  font-weight: 600;
  color: var(--atm-text-muted);
  border-bottom: 1px solid #e2e8f0;
}

.icm-table td {
  padding: 10px 12px;
  border-bottom: 1px solid #f1f5f9;
}

.icm-mono {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  font-size: 12px;
  line-height: 1.4;
  word-break: break-all;
  color: var(--atm-text-muted);
}

.icm-row-actions {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 10px;
  white-space: nowrap;
}

.icm-link {
  padding: 0;
  font-size: 13px;
  color: var(--atm-primary, #7c3aed);
  background: none;
  border: none;
  cursor: pointer;
}

.icm-link:disabled {
  opacity: 0.55;
  cursor: not-allowed;
}

.icm-link:hover:not(:disabled) {
  text-decoration: underline;
}

.icm-form {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.icm-form-stack {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.icm-field {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.icm-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-text-muted);
}

.icm-input {
  width: 100%;
  box-sizing: border-box;
  padding: 10px 12px;
  font-size: 13px;
  color: var(--atm-text);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
}

.icm-candidates {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.icm-candidate {
  padding: 8px 12px;
  font-size: 12px;
  text-align: left;
  color: var(--atm-text);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  cursor: pointer;
}

.icm-candidate:hover {
  border-color: var(--atm-primary, #7c3aed);
}

.icm-check {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--atm-text);
}

.icm-actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 4px;
}

.icm-actions--list {
  margin-top: 8px;
}
</style>
