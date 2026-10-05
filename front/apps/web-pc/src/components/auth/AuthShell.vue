<script setup lang="ts">
import { RouterLink } from 'vue-router'
import SiteLogo from '@/components/brand/SiteLogo.vue'
import { SITE_NAME } from '@/constants/brand'

defineProps<{
  title: string
  subtitle?: string
}>()
</script>

<template>
  <div class="auth-page">
    <aside class="auth-brand" aria-hidden="false">
      <div class="auth-brand-bg" aria-hidden="true">
        <span class="orb orb-a" />
        <span class="orb orb-b" />
        <span class="grid-lines" />
      </div>
      <div class="auth-brand-inner">
        <SiteLogo variant="auth" class="auth-brand-logo" />

        <p class="auth-brand-eyebrow">ChatGPT · Claude · Cursor</p>
        <h2 class="auth-brand-headline">国内低价<br />自助开通 AI 套餐</h2>
        <p class="auth-brand-lead">
          注册会员后可在线下单，额度与订单在会员中心
        </p>

        <ul class="auth-brand-list">
          <li>
            <span class="check" aria-hidden="true">✓</span>
            支付宝 / 微信自助支付
          </li>
          <li>
            <span class="check" aria-hidden="true">✓</span>
            价格低于官方，开通流程透明
          </li>
          <li>
            <span class="check" aria-hidden="true">✓</span>
            1000+ 用户已开通
          </li>
        </ul>

        <RouterLink to="/" class="auth-brand-back">← 返回首页选购套餐</RouterLink>
      </div>
    </aside>

    <main class="auth-main">
      <header class="auth-main-top">
        <RouterLink to="/" class="auth-main-home">首页</RouterLink>
      </header>

      <div class="auth-main-body">
        <h1 class="auth-title">{{ title }}</h1>
        <p v-if="subtitle" class="auth-subtitle">{{ subtitle }}</p>
        <slot />
      </div>

      <footer class="auth-main-foot">
        <span>© {{ SITE_NAME }}</span>
        <span class="dot" aria-hidden="true">·</span>
        <a href="/#faq">帮助与 FAQ</a>
      </footer>
    </main>
  </div>
</template>

<style scoped>
.auth-page {
  display: grid;
  min-height: 100vh;
  grid-template-columns: minmax(360px, 1.05fr) minmax(400px, 0.95fr);
}

@media (max-width: 960px) {
  .auth-page {
    grid-template-columns: 1fr;
  }
}

/* —— Brand panel —— */
.auth-brand {
  position: relative;
  display: flex;
  align-items: stretch;
  overflow: hidden;
  color: #fff;
  background: linear-gradient(145deg, #4c1d95 0%, #6d28d9 42%, #4338ca 100%);
}

@media (max-width: 960px) {
  .auth-brand {
    min-height: auto;
    padding-bottom: 8px;
  }
}

.auth-brand-bg {
  position: absolute;
  inset: 0;
  pointer-events: none;
}

.grid-lines {
  position: absolute;
  inset: 0;
  opacity: 0.12;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.35) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.35) 1px, transparent 1px);
  background-size: 48px 48px;
  mask-image: linear-gradient(180deg, #000 30%, transparent 95%);
}

.orb {
  position: absolute;
  border-radius: 50%;
  filter: blur(60px);
}

.orb-a {
  top: -80px;
  right: -40px;
  width: 280px;
  height: 280px;
  background: rgba(167, 139, 250, 0.55);
}

.orb-b {
  bottom: -60px;
  left: -80px;
  width: 320px;
  height: 320px;
  background: rgba(99, 102, 241, 0.45);
}

.auth-brand-inner {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 520px;
  margin: 0 auto;
  padding: 48px 48px 40px;
}

@media (max-width: 960px) {
  .auth-brand-inner {
    padding: 32px 24px 28px;
  }

  .auth-brand-headline {
    font-size: 26px !important;
  }

  .auth-brand-list,
  .auth-brand-back {
    display: none;
  }
}

.auth-brand-logo {
  margin-bottom: 40px;
}

@media (max-width: 960px) {
  .auth-brand-logo {
    margin-bottom: 20px;
  }
}

.auth-brand-eyebrow {
  margin: 0 0 12px;
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.06em;
  text-transform: uppercase;
  color: rgba(255, 255, 255, 0.72);
}

.auth-brand-headline {
  margin: 0 0 16px;
  font-size: 36px;
  font-weight: 800;
  line-height: 1.15;
  letter-spacing: -0.03em;
}

.auth-brand-lead {
  margin: 0 0 32px;
  max-width: 380px;
  font-size: 15px;
  line-height: 1.65;
  color: rgba(255, 255, 255, 0.82);
}

.auth-brand-list {
  margin: 0 0 36px;
  padding: 0;
  list-style: none;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.auth-brand-list li {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  font-size: 15px;
  line-height: 1.5;
  color: rgba(255, 255, 255, 0.92);
}

.check {
  flex-shrink: 0;
  width: 22px;
  height: 22px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  color: #4c1d95;
  background: rgba(255, 255, 255, 0.95);
  border-radius: 50%;
}

.auth-brand-back {
  margin-top: auto;
  font-size: 14px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.85);
  text-decoration: none;
  transition: color 0.15s;
}

.auth-brand-back:hover {
  color: #fff;
}

/* —— Form panel —— */
.auth-main {
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  background: #fafafa;
}

.auth-main-top {
  display: flex;
  justify-content: flex-end;
  padding: 20px 32px 0;
}

@media (min-width: 961px) {
  .auth-main-top {
    display: none;
  }
}

.auth-main-home {
  font-size: 14px;
  font-weight: 500;
  color: var(--atm-text-muted);
  text-decoration: none;
}

.auth-main-home:hover {
  color: var(--atm-primary);
}

.auth-main-body {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  width: 100%;
  max-width: 440px;
  margin: 0 auto;
  padding: 32px 32px 48px;
}

@media (min-width: 961px) {
  .auth-main-body {
    padding: 48px 56px 64px;
    max-width: 480px;
  }
}

.auth-title {
  margin: 0 0 10px;
  font-size: 32px;
  font-weight: 800;
  letter-spacing: -0.03em;
  color: var(--atm-text);
}

.auth-subtitle {
  margin: 0 0 32px;
  font-size: 15px;
  line-height: 1.6;
  color: var(--atm-text-muted);
}

.auth-main-foot {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 20px 24px 28px;
  font-size: 13px;
  color: #94a3b8;
}

.auth-main-foot a {
  color: var(--atm-text-muted);
  text-decoration: none;
}

.auth-main-foot a:hover {
  color: var(--atm-primary);
}

.dot {
  opacity: 0.5;
}
</style>
