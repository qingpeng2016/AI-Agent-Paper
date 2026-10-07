<script setup lang="ts">
import { RouterView } from 'vue-router'
</script>

<template>
  <div class="workbench-layout">
    <RouterView />
  </div>
</template>

<style scoped>
.workbench-layout {
  height: 100vh;
  max-height: 100vh;
  overflow: hidden;
  background: var(--atm-bg, #f5f3ff);
}
</style>
