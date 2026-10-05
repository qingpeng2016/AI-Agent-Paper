<template>
  <div class="mega-grid mega-grid--skeleton" aria-busy="true" aria-label="菜单加载中">
    <div v-for="c in columns" :key="c" class="mega-col">
      <div class="sk sk-col-title" />
      <ul class="mega-list">
        <li v-for="r in rows" :key="r">
          <div class="sk sk-row" />
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
withDefaults(
  defineProps<{ columns?: number; rows?: number }>(),
  { columns: 5, rows: 4 },
)
</script>

<style scoped>
.sk {
  border-radius: 8px;
  background: linear-gradient(90deg, #f1f5f9 0%, #e8ecf4 45%, #f1f5f9 90%);
  background-size: 200% 100%;
  animation: nav-sk-shimmer 1.2s ease-in-out infinite;
}

.sk-col-title {
  width: 72px;
  height: 16px;
  margin: 0 8px 8px;
}

.sk-row {
  height: 32px;
  margin-bottom: 4px;
}

.mega-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(158px, 1fr));
  gap: 18px 28px;
  align-items: start;
  min-height: 220px;
}

@media (min-width: 1100px) {
  .mega-grid {
    grid-template-columns: repeat(5, minmax(172px, 1fr));
    gap: 16px 20px;
  }
}

.mega-list {
  margin: 0;
  padding: 0;
  list-style: none;
}

.mega-col {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  min-width: 0;
}

@keyframes nav-sk-shimmer {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: -100% 0;
  }
}
</style>
