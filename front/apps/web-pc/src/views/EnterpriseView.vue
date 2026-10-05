<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { ElMessage } from 'element-plus'
import { enterpriseApi } from '@/api'
import EnterpriseBenefitIcon from '@/components/enterprise/EnterpriseBenefitIcon.vue'
import { SITE_NAME } from '@/constants/brand'
import {
  enterpriseBenefits,
  enterpriseFaqs,
  enterpriseChannelHighlights,
  enterpriseHero,
  enterpriseInvoiceNotes,
  enterpriseStats,
  enterpriseSteps,
} from '@/mocks/enterprise'
import {
  useEnterpriseProducts,
  type EnterprisePlanView,
} from '@/composables/useEnterpriseProducts'

const submitting = ref(false)

const { plans: enterprisePlans } = useEnterpriseProducts()
const selectedPlanId = ref('')

watch(
  enterprisePlans,
  (list) => {
    if (!list.length) return
    if (!list.some((p) => p.id === selectedPlanId.value)) {
      selectedPlanId.value = list[0]?.id ?? ''
    }
  },
  { immediate: true },
)

function selectPlan(planId: string) {
  selectedPlanId.value = planId
}

const form = reactive({
  company: '',
  contact: '',
  phone: '',
  email: '',
})

const quotingPlanId = ref('')

const centerNotice = reactive({
  open: false,
  title: '',
  lines: [] as string[],
})

function openCenterNotice(title: string, lines: string[]) {
  centerNotice.title = title
  centerNotice.lines = lines.filter(Boolean)
  centerNotice.open = true
}

function closeCenterNotice() {
  centerNotice.open = false
}

async function onQuoteClick(plan: EnterprisePlanView) {
  quotingPlanId.value = plan.id
  try {
    const p = await enterpriseApi.getProduct(plan.id)
    openCenterNotice(p.name, [
      p.price_hint,
      p.seats,
      ...(p.features ?? []).map((f) => `· ${f}`),
      p.tagline,
    ])
  } catch (e) {
    const msg = e instanceof Error ? e.message : '获取方案信息失败'
    ElMessage.error(msg)
  } finally {
    quotingPlanId.value = ''
  }
}

async function onSubmit() {
  if (!form.company.trim() || !form.contact.trim() || !form.phone.trim()) {
    ElMessage.warning('请填写公司、联系人与手机号')
    return
  }
  submitting.value = true
  try {
    const res = await enterpriseApi.submitInquiry({
      company_name: form.company.trim(),
      contact_name: form.contact.trim(),
      phone: form.phone.trim(),
      email: form.email.trim() || undefined,
    })
    openCenterNotice('提交成功', [
      res.message ?? '已收到采购需求，顾问将尽快联系您',
    ])
    form.company = ''
    form.contact = ''
    form.phone = ''
    form.email = ''
  } catch (e) {
    const msg = e instanceof Error ? e.message : '提交失败，请稍后重试'
    ElMessage.error(msg)
  } finally {
    submitting.value = false
  }
}
</script>

<template>
  <Teleport to="body">
    <div
      v-if="centerNotice.open"
      class="ent-center-notice"
      role="dialog"
      aria-modal="true"
      @click.self="closeCenterNotice"
    >
      <div class="ent-center-notice-card">
        <h3 v-if="centerNotice.title" class="ent-center-notice-title">
          {{ centerNotice.title }}
        </h3>
        <p v-for="(line, i) in centerNotice.lines" :key="i" class="ent-center-notice-line">
          {{ line }}
        </p>
        <button type="button" class="ent-center-notice-btn" @click="closeCenterNotice">
          知道了
        </button>
      </div>
    </div>
  </Teleport>

  <div class="enterprise-page">
    <section class="ent-hero">
      <div class="atm-container ent-hero-inner">
        <div class="ent-hero-copy">
          <span class="ent-eyebrow">{{ enterpriseHero.eyebrow }}</span>
          <h1 class="ent-title">{{ enterpriseHero.title }}</h1>
          <div class="ent-hero-highlights">
            <ul class="ent-highlight-list">
              <li v-for="item in enterpriseChannelHighlights" :key="item">{{ item }}</li>
            </ul>
          </div>
        </div>
        <form id="contact" class="ent-form ent-hero-form" @submit.prevent="onSubmit">
          <div class="ent-form-row">
            <label>
              <span>公司名称 *</span>
              <input v-model="form.company" type="text" placeholder="与开票抬头一致" />
            </label>
            <label>
              <span>联系人 *</span>
              <input v-model="form.contact" type="text" placeholder="采购 / IT 负责人" />
            </label>
          </div>
          <div class="ent-form-row">
            <label>
              <span>手机 *</span>
              <input v-model="form.phone" type="tel" placeholder="用于顾问回电" />
            </label>
            <label>
              <span>邮箱</span>
              <input v-model="form.email" type="email" placeholder="接收报价单" />
            </label>
          </div>
          <button type="submit" class="atm-btn-primary ent-submit" :disabled="submitting">
            {{ submitting ? '提交中…' : '提交需求' }}
          </button>
        </form>
      </div>
    </section>

    <section class="ent-stats">
      <div class="atm-container ent-stats-grid">
        <div v-for="s in enterpriseStats" :key="s.label" class="ent-stat">
          <strong class="ent-stat-value">{{ s.value }}</strong>
          <span class="ent-stat-label">{{ s.label }}</span>
        </div>
      </div>
    </section>

    <section id="plans" class="ent-section">
      <div class="atm-container">
        <header class="ent-section-head">
          <h2>方案参考</h2>
        </header>
        <div class="ent-plans">
          <article
            v-for="plan in enterprisePlans"
            :key="plan.id"
            class="ent-plan"
            :class="{ 'ent-plan--featured': selectedPlanId === plan.id }"
            role="button"
            tabindex="0"
            @click="selectPlan(plan.id)"
            @keydown.enter.prevent="selectPlan(plan.id)"
            @keydown.space.prevent="selectPlan(plan.id)"
          >
            <div class="ent-plan-top">
              <h3>{{ plan.name }}</h3>
              <span v-if="plan.badge" class="ent-plan-badge">{{ plan.badge }}</span>
            </div>
            <p class="ent-plan-price">{{ plan.priceHint }}</p>
            <p class="ent-plan-seats">{{ plan.seats }}</p>
            <ul>
              <li v-for="f in plan.features" :key="f">{{ f }}</li>
            </ul>
            <p class="ent-plan-cta">{{ plan.cta }}</p>
            <button
              type="button"
              class="ent-plan-btn"
              :disabled="quotingPlanId === plan.id"
              @click.stop="onQuoteClick(plan)"
            >
              {{ quotingPlanId === plan.id ? '加载中…' : plan.buttonLabel }}
            </button>
          </article>
        </div>
      </div>
    </section>

    <section class="ent-section">
      <div class="atm-container">
        <header class="ent-section-head">
          <h2>为什么走企业采购</h2>
        </header>
        <div class="ent-benefits">
          <article v-for="b in enterpriseBenefits" :key="b.title" class="ent-benefit">
            <span class="ent-benefit-icon">
              <EnterpriseBenefitIcon :name="b.icon" />
            </span>
            <h3>{{ b.title }}</h3>
            <p>{{ b.desc }}</p>
          </article>
        </div>
      </div>
    </section>

    <section class="ent-section">
      <div class="atm-container">
        <header class="ent-section-head">
          <h2>合作流程</h2>
        </header>
        <ol class="ent-steps">
          <li v-for="step in enterpriseSteps" :key="step.step" class="ent-step">
            <span class="ent-step-num">{{ step.step }}</span>
            <div>
              <h3>{{ step.title }}</h3>
              <p>{{ step.desc }}</p>
            </div>
          </li>
        </ol>
      </div>
    </section>

    <section id="invoice" class="ent-section">
      <div class="atm-container ent-invoice-grid">
        <div>
          <header class="ent-section-head ent-section-head--left">
            <h2>开票说明</h2>
            <p>与个人下单规则一致，企业批量可集中由一人提交抬头。</p>
          </header>
          <ul class="ent-invoice-list">
            <li v-for="(note, i) in enterpriseInvoiceNotes" :key="i">{{ note }}</li>
          </ul>
          <RouterLink to="/member?tab=invoices" class="ent-inline-link">会员中心 · 发票 →</RouterLink>
        </div>
        <aside class="ent-invoice-aside">
          <h3>对公账户（示例）</h3>
          <dl>
            <div><dt>户名</dt><dd>{{ SITE_NAME }} 科技有限公司</dd></div>
            <div><dt>开户行</dt><dd>某某银行某某支行</dd></div>
            <div><dt>账号</dt><dd>6222 **** **** 1234</dd></div>
            <div><dt>备注</dt><dd>AI-Plan-企业采购-公司简称</dd></div>
          </dl>
          <p class="ent-aside-note">正式账号以顾问邮件 / 报价单为准，切勿向私人账户转账。</p>
        </aside>
      </div>
    </section>

    <section class="ent-section ent-faq-wrap">
      <div class="atm-container ent-faq-inner">
        <h2 class="ent-faq-title">常见问题</h2>
        <dl class="ent-faq">
          <div v-for="item in enterpriseFaqs" :key="item.q" class="ent-faq-item">
            <dt>{{ item.q }}</dt>
            <dd>{{ item.a }}</dd>
          </div>
        </dl>
      </div>
    </section>
  </div>
</template>

<style scoped>
.enterprise-page {
  padding-bottom: 64px;
  background: var(--atm-bg);
}

.ent-hero {
  padding: 48px 0 56px;
  background: linear-gradient(145deg, #1e1b4b 0%, #4c1d95 45%, #6d28d9 100%);
  color: #fff;
}

.ent-hero-inner {
  display: grid;
  gap: 32px;
  align-items: start;
}

@media (min-width: 960px) {
  .ent-hero-inner {
    grid-template-columns: 1.05fr 0.95fr;
    gap: 40px;
  }
}

.ent-eyebrow {
  display: inline-block;
  margin-bottom: 14px;
  padding: 6px 12px;
  font-size: 12px;
  font-weight: 700;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  background: rgba(255, 255, 255, 0.12);
  border-radius: 999px;
}

.ent-title {
  margin: 0 0 16px;
  font-size: clamp(1.75rem, 4vw, 2.5rem);
  font-weight: 800;
  line-height: 1.2;
  letter-spacing: -0.03em;
}

.ent-hero-highlights {
  margin-top: 8px;
  max-width: 28em;
}

.ent-highlight-list {
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: 14px;
  line-height: 1.65;
  color: rgba(255, 255, 255, 0.92);
}

.ent-highlight-list li {
  position: relative;
  padding-left: 18px;
  margin-bottom: 10px;
}

.ent-highlight-list li:last-child {
  margin-bottom: 0;
}

.ent-highlight-list li::before {
  content: '';
  position: absolute;
  left: 0;
  top: 0.55em;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #c4b5fd;
}

.ent-hero-form {
  margin: 0;
  padding: 22px 20px;
  box-shadow: 0 20px 50px rgba(15, 10, 40, 0.35);
}

.ent-hero-form .ent-submit {
  margin-top: 20px;
}

@media (min-width: 960px) {
  .ent-hero-form {
    padding: 24px 22px;
  }

  .ent-hero-form .ent-form-row {
    grid-template-columns: 1fr;
    gap: 12px;
    margin-bottom: 12px;
  }

  .ent-hero-form .ent-form-checks {
    margin-bottom: 12px;
  }

  .ent-hero-form .ent-form-full {
    margin-bottom: 16px;
  }
}

@media (min-width: 1100px) {
  .ent-hero-form .ent-form-row {
    grid-template-columns: 1fr 1fr;
  }
}

.ent-stats {
  margin-top: -28px;
  position: relative;
  z-index: 2;
  padding: 0 0 8px;
}

.ent-stats-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  padding: 20px 24px;
  background: #fff;
  border-radius: 18px;
  border: 1px solid rgba(124, 58, 237, 0.12);
  box-shadow: 0 16px 48px rgba(30, 27, 75, 0.1);
}

@media (min-width: 640px) {
  .ent-stats-grid {
    grid-template-columns: repeat(4, 1fr);
    gap: 8px;
  }
}

.ent-stat {
  text-align: center;
  padding: 8px 4px;
}

.ent-stat-value {
  display: block;
  font-size: 22px;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-primary-dark);
}

.ent-stat-label {
  display: block;
  margin-top: 4px;
  font-size: 12px;
  color: var(--atm-text-muted);
}

.ent-section {
  padding: 56px 0;
  background: transparent;
}

.ent-section-head {
  text-align: center;
  max-width: 36em;
  margin: 0 auto 36px;
}

.ent-section-head--left {
  text-align: left;
  margin-left: 0;
  margin-right: 0;
}

.ent-section-head h2 {
  margin: 0 0 10px;
  font-size: 1.5rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--atm-text);
}

.ent-section-head p {
  margin: 0;
  font-size: 14px;
  line-height: 1.65;
  color: var(--atm-text-muted);
}

.ent-benefits {
  display: grid;
  gap: 18px;
}

@media (min-width: 640px) {
  .ent-benefits {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 960px) {
  .ent-benefits {
    grid-template-columns: repeat(3, 1fr);
  }
}

.ent-benefit {
  padding: 22px 22px 20px;
  background: #fff;
  border-radius: 16px;
  border: 1px solid rgba(124, 58, 237, 0.1);
  box-shadow: 0 6px 24px rgba(124, 58, 237, 0.05);
}

.ent-benefit-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  margin-bottom: 12px;
  color: var(--atm-primary-dark);
  background: var(--atm-primary-light);
  border-radius: 12px;
}

.ent-benefit h3 {
  margin: 0 0 8px;
  font-size: 16px;
  font-weight: 800;
  color: var(--atm-text);
}

.ent-benefit p {
  margin: 0;
  font-size: 13px;
  line-height: 1.65;
  color: var(--atm-text-muted);
}

.ent-steps {
  margin: 0;
  padding: 0;
  list-style: none;
  display: grid;
  gap: 16px;
}

@media (min-width: 768px) {
  .ent-steps {
    grid-template-columns: repeat(2, 1fr);
  }
}

@media (min-width: 1024px) {
  .ent-steps {
    grid-template-columns: repeat(4, 1fr);
  }
}

.ent-step {
  display: flex;
  gap: 14px;
  padding: 20px 18px;
  background: #fff;
  border-radius: 16px;
  border: 1px solid #f1f5f9;
}

.ent-step-num {
  flex-shrink: 0;
  font-size: 13px;
  font-weight: 800;
  color: var(--atm-primary);
}

.ent-step h3 {
  margin: 0 0 6px;
  font-size: 15px;
  font-weight: 800;
  color: var(--atm-text);
}

.ent-step p {
  margin: 0;
  font-size: 13px;
  line-height: 1.6;
  color: var(--atm-text-muted);
}

.ent-plans {
  display: grid;
  gap: 20px;
}

@media (min-width: 768px) {
  .ent-plans {
    grid-template-columns: repeat(3, 1fr);
    align-items: stretch;
  }
}

.ent-plan {
  display: flex;
  flex-direction: column;
  padding: 24px 22px;
  background: #fff;
  border-radius: 18px;
  border: 1px solid rgba(124, 58, 237, 0.12);
  cursor: pointer;
  transition:
    border-color 0.2s ease,
    box-shadow 0.2s ease,
    transform 0.2s ease;
}

.ent-plan--featured {
  border-color: var(--atm-primary);
  box-shadow: 0 16px 40px rgba(124, 58, 237, 0.15);
  transform: translateY(-4px);
}

.ent-plan-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-bottom: 8px;
}

.ent-plan-top h3 {
  margin: 0;
  font-size: 18px;
  font-weight: 800;
}

.ent-plan-badge {
  padding: 4px 10px;
  font-size: 11px;
  font-weight: 700;
  color: #fff;
  background: var(--atm-gradient);
  border-radius: 999px;
}

.ent-plan-price {
  margin: 0 0 4px;
  font-size: 22px;
  font-weight: 800;
  color: var(--atm-primary-dark);
}

.ent-plan-seats {
  margin: 0 0 16px;
  font-size: 13px;
  color: var(--atm-text-muted);
}

.ent-plan ul {
  flex: 1;
  margin: 0 0 12px;
  padding-left: 1.1em;
  font-size: 13px;
  line-height: 1.65;
  color: var(--atm-text-muted);
}

.ent-plan-cta {
  margin: 0 0 16px;
  font-size: 12px;
  font-weight: 600;
  color: var(--atm-primary);
}

.ent-plan-btn {
  width: 100%;
  padding: 12px 16px;
  font-size: 14px;
  font-weight: 700;
  color: var(--atm-primary-dark);
  background: var(--atm-primary-light);
  border: none;
  border-radius: 12px;
  cursor: pointer;
  transition: background 0.15s ease;
}

.ent-plan--featured .ent-plan-btn {
  color: #fff;
  background: var(--atm-gradient);
}

.ent-plan-btn:hover {
  filter: brightness(1.03);
}

.ent-invoice-grid {
  display: grid;
  gap: 32px;
}

@media (min-width: 900px) {
  .ent-invoice-grid {
    grid-template-columns: 1.1fr 0.9fr;
    align-items: start;
  }
}

.ent-invoice-list {
  margin: 0 0 16px;
  padding-left: 1.2em;
  font-size: 14px;
  line-height: 1.7;
  color: var(--atm-text-muted);
}

.ent-inline-link {
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-primary);
  text-decoration: none;
}

.ent-inline-link:hover {
  color: var(--atm-primary-dark);
}

.ent-invoice-aside {
  padding: 22px 24px;
  background: #fff;
  border-radius: 16px;
  border: 1px solid #e2e8f0;
}

.ent-invoice-aside h3 {
  margin: 0 0 16px;
  font-size: 16px;
  font-weight: 800;
}

.ent-invoice-aside dl {
  margin: 0;
}

.ent-invoice-aside div {
  display: grid;
  grid-template-columns: 4.5em 1fr;
  gap: 8px;
  margin-bottom: 10px;
  font-size: 13px;
}

.ent-invoice-aside dt {
  color: var(--atm-text-muted);
}

.ent-invoice-aside dd {
  margin: 0;
  font-weight: 600;
  color: var(--atm-text);
}

.ent-aside-note {
  margin: 16px 0 0;
  font-size: 12px;
  line-height: 1.55;
  color: #b45309;
}

.ent-form {
  padding: 28px 26px;
  background: #fff;
  border-radius: 20px;
  border: 1px solid rgba(124, 58, 237, 0.12);
  box-shadow: 0 12px 40px rgba(124, 58, 237, 0.08);
}

.ent-form-row {
  display: grid;
  gap: 16px;
  margin-bottom: 16px;
}

@media (min-width: 560px) {
  .ent-form-row {
    grid-template-columns: 1fr 1fr;
  }
}

.ent-form label span {
  display: block;
  margin-bottom: 6px;
  font-size: 13px;
  font-weight: 600;
  color: var(--atm-text);
}

.ent-form input,
.ent-form textarea {
  width: 100%;
  padding: 10px 12px;
  font-size: 14px;
  font-family: inherit;
  color: var(--atm-text);
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 10px;
  transition: border-color 0.15s ease;
}

.ent-form input:focus,
.ent-form textarea:focus {
  outline: none;
  border-color: var(--atm-primary);
  background: #fff;
}

.ent-form-checks {
  display: flex;
  flex-wrap: wrap;
  gap: 16px 24px;
  margin-bottom: 16px;
}

.ent-check {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--atm-text-muted);
  cursor: pointer;
}

.ent-form-full {
  display: block;
  margin-bottom: 20px;
}

.ent-submit {
  width: 100%;
}

.ent-faq-wrap {
  padding-top: 24px;
}

.ent-faq-title {
  margin: 0 0 24px;
  text-align: left;
  font-size: 1.35rem;
  font-weight: 800;
}

.ent-faq {
  margin: 0;
  max-width: 720px;
}

.ent-faq-item {
  padding: 18px 0;
  border-bottom: 1px solid #e2e8f0;
}

.ent-faq-item dt {
  margin-bottom: 8px;
  font-size: 15px;
  font-weight: 700;
  color: var(--atm-text);
}

.ent-faq-item dd {
  margin: 0;
  font-size: 14px;
  line-height: 1.65;
  color: var(--atm-text-muted);
}

.ent-center-notice {
  position: fixed;
  inset: 0;
  z-index: 10000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
  background: rgba(15, 23, 42, 0.45);
}

.ent-center-notice-card {
  width: min(420px, 100%);
  padding: 32px 28px 24px;
  text-align: center;
  background: #fff;
  border-radius: 20px;
  box-shadow: 0 24px 48px rgba(15, 23, 42, 0.12);
  border: 1px solid #e2e8f0;
}

.ent-center-notice-title {
  margin: 0 0 16px;
  font-size: 1.25rem;
  font-weight: 800;
  color: var(--atm-text);
}

.ent-center-notice-line {
  margin: 0 0 10px;
  font-size: 15px;
  line-height: 1.6;
  font-weight: 500;
  color: var(--atm-text-muted);
}

.ent-center-notice-line:last-of-type {
  margin-bottom: 24px;
}

.ent-center-notice-btn {
  padding: 10px 28px;
  font-size: 14px;
  font-weight: 700;
  color: #fff;
  background: var(--atm-gradient);
  border: none;
  border-radius: 999px;
  cursor: pointer;
}

.ent-center-notice-btn:hover {
  opacity: 0.92;
}
</style>
