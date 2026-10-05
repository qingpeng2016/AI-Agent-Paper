<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { couponApi } from '@/api'
import type { RegisterCouponPromo } from '@ai-token-mall/shared'
import {
  dismissRegisterCouponPromoForSession,
  isRegisterCouponPromoDismissed,
} from '@/composables/useSessionUser'

const props = defineProps<{
  /** 与 guestPromoRemountKey 同步，退出登录后递增以重新拉取 */
  generation: number
}>()

const visible = ref(false)
const promo = ref<RegisterCouponPromo | null>(null)

function dismissForSession() {
  dismissRegisterCouponPromoForSession()
  visible.value = false
}

async function loadPromo() {
  if (isRegisterCouponPromoDismissed()) {
    visible.value = false
    return
  }
  try {
    const data = await couponApi.registerPromo()
    if (!data.active) {
      visible.value = false
      promo.value = null
      return
    }
    promo.value = data
    visible.value = true
  } catch (e) {
    visible.value = false
    if (import.meta.env.DEV) {
      console.warn('[RegisterCouponPromo] register-promo failed', e)
    }
  }
}

watch(
  () => props.generation,
  () => {
    void loadPromo()
  },
  { immediate: true },
)
</script>

<template>
  <Teleport to="body">
    <div v-if="visible && promo" class="promo-root" role="dialog" aria-modal="true" aria-labelledby="promo-title">
      <div class="promo-backdrop" @click="dismissForSession" />
      <div class="promo-card">
        <button type="button" class="promo-close" aria-label="关闭" @click="dismissForSession">×</button>
        <h2 id="promo-title" class="promo-title">{{ promo.title }}</h2>
        <p v-if="promo.subtitle" class="promo-sub">{{ promo.subtitle }}</p>
        <div class="promo-amount">{{ promo.discount_label }}</div>
        <p class="promo-meta">注册自动到账 · {{ promo.valid_days }} 天内有效</p>
        <RouterLink class="promo-cta" to="/register">立即注册领取</RouterLink>
      </div>
    </div>
  </Teleport>
</template>

<style scoped>
.promo-root {
  position: fixed;
  inset: 0;
  z-index: 9000;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 1.5rem;
}
.promo-backdrop {
  position: absolute;
  inset: 0;
  background: rgba(15, 23, 42, 0.55);
  backdrop-filter: blur(4px);
}
.promo-card {
  position: relative;
  width: min(420px, 100%);
  padding: 2rem 1.75rem 1.5rem;
  border-radius: 20px;
  text-align: center;
  color: #fff;
  background: linear-gradient(145deg, #6366f1 0%, #8b5cf6 45%, #ec4899 100%);
  box-shadow: 0 24px 60px rgba(79, 70, 229, 0.45);
}
.promo-close {
  position: absolute;
  top: 0.75rem;
  right: 0.85rem;
  border: none;
  background: transparent;
  color: rgba(255, 255, 255, 0.85);
  font-size: 1.5rem;
  line-height: 1;
  cursor: pointer;
}
.promo-title {
  margin: 0 0 0.5rem;
  font-size: 1.35rem;
  font-weight: 700;
}
.promo-sub {
  margin: 0 0 1rem;
  font-size: 0.9rem;
  opacity: 0.92;
  line-height: 1.5;
}
.promo-amount {
  font-size: 2.75rem;
  font-weight: 800;
  letter-spacing: -0.02em;
  margin: 0.25rem 0 0.5rem;
  text-shadow: 0 2px 12px rgba(0, 0, 0, 0.15);
}
.promo-meta {
  margin: 0 0 1.25rem;
  font-size: 0.8rem;
  opacity: 0.88;
}
.promo-cta {
  display: block;
  width: 100%;
  padding: 0.85rem 1rem;
  border-radius: 12px;
  font-weight: 600;
  text-decoration: none;
  color: #4f46e5;
  background: #fff;
  box-shadow: 0 4px 14px rgba(0, 0, 0, 0.12);
}
.promo-cta:hover {
  filter: brightness(1.02);
}
</style>
