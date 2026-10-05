<script setup lang="ts">
import { RouterView } from 'vue-router'
import FloatingCustomerService from '@/components/layout/FloatingCustomerService.vue'
</script>

<template>
  <div class="auth-layout">
    <RouterView />
    <FloatingCustomerService />
  </div>
</template>

<style scoped>
.auth-layout {
  min-height: 100vh;
  background: #fff;
}
</style>
