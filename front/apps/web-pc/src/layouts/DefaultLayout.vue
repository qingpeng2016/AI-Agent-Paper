<script setup lang="ts">
import { computed } from 'vue'
import { RouterView, useRoute } from 'vue-router'
import FloatingCustomerService from '@/components/layout/FloatingCustomerService.vue'
import SiteFooter from '@/components/layout/SiteFooter.vue'
import SiteHeader from '@/components/layout/SiteHeader.vue'
import RegisterCouponPromo from '@/components/home/RegisterCouponPromo.vue'
import {
  authSessionRevision,
  guestPromoRemountKey,
  isLoggedIn,
} from '@/composables/useSessionUser'

const route = useRoute()

const showGuestCouponPromo = computed(() => {
  authSessionRevision.value
  return route.name === 'home' && !isLoggedIn()
})
</script>

<template>
  <div class="layout">
    <SiteHeader />
    <main class="main">
      <RouterView />
    </main>
    <SiteFooter />
    <FloatingCustomerService />
    <RegisterCouponPromo
      v-if="showGuestCouponPromo"
      :key="guestPromoRemountKey"
      :generation="guestPromoRemountKey"
    />
  </div>
</template>

<style scoped>
.layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
.main {
  flex: 1;
  background: var(--atm-bg);
}
</style>
