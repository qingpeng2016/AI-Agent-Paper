<script setup lang="ts">
import { useRouteNavigationLoading } from '@/composables/useRouteNavigationLoading'

const { active: routeLoading } = useRouteNavigationLoading()
</script>

<template>
  <div
    v-if="routeLoading"
    class="route-loading-bar"
    role="progressbar"
    aria-live="polite"
    aria-busy="true"
  />
  <router-view />
</template>

<style scoped>
.route-loading-bar {
  position: fixed;
  top: 0;
  left: 0;
  z-index: 10000;
  width: 38%;
  max-width: 420px;
  height: 3px;
  pointer-events: none;
  background: linear-gradient(
    90deg,
    transparent,
    var(--atm-primary, #7c3aed) 45%,
    #a78bfa 55%,
    transparent
  );
  box-shadow: 0 0 12px rgba(124, 58, 237, 0.45);
  animation: route-loading-slide 0.85s ease-in-out infinite;
}
@keyframes route-loading-slide {
  0% {
    transform: translateX(-120%);
  }
  100% {
    transform: translateX(calc(100vw + 120%));
  }
}
</style>
