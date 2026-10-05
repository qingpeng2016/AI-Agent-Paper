<template>
  <div class="nav-drop-skeleton" aria-busy="true" aria-label="加载中">
    <div v-for="i in rows" :key="i" class="sk sk-row" />
  </div>
</template>

<script setup lang="ts">
withDefaults(defineProps<{ rows?: number }>(), { rows: 5 })
</script>

<style scoped>
.nav-drop-skeleton {
  padding: 4px 0;
}

.sk {
  border-radius: 8px;
  background: linear-gradient(90deg, #f1f5f9 0%, #e8ecf4 45%, #f1f5f9 90%);
  background-size: 200% 100%;
  animation: nav-drop-shimmer 1.2s ease-in-out infinite;
}

.sk-row {
  height: 36px;
  margin-bottom: 4px;
}

@keyframes nav-drop-shimmer {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: -100% 0;
  }
}
</style>
