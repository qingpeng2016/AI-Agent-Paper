<template>
  <div class="blog-list-skeleton" aria-busy="true" aria-label="加载教程列表">
    <div class="sk sk-title" />
    <div class="sk-row-pills">
      <div v-for="i in 8" :key="i" class="sk sk-pill" />
    </div>
    <div class="sk-grid">
      <div v-for="i in 4" :key="i" class="sk sk-card">
        <div class="sk sk-line sk-line--sm" />
        <div class="sk sk-line sk-line--lg" />
        <div class="sk sk-line" />
        <div class="sk sk-line sk-line--md" />
      </div>
    </div>
  </div>
</template>

<style scoped>
.blog-list-skeleton {
  position: relative;
}

.sk {
  border-radius: 10px;
  background: linear-gradient(90deg, #eef1f6 0%, #e2e8f0 45%, #eef1f6 90%);
  background-size: 200% 100%;
  animation: blog-sk-shimmer 1.2s ease-in-out infinite;
}

.sk-title {
  height: 36px;
  width: min(320px, 70%);
  margin-bottom: 28px;
}

.sk-row-pills {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
  margin-bottom: 28px;
}

.sk-pill {
  height: 34px;
  width: 72px;
  border-radius: 999px;
}

.sk-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

@media (max-width: 720px) {
  .sk-grid {
    grid-template-columns: 1fr;
  }
}

.sk-card {
  padding: 22px;
  min-height: 168px;
  border-radius: 16px;
  background: #fff;
  border: 1px solid #e8ecf4;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.sk-line {
  height: 14px;
  width: 100%;
}

.sk-line--sm {
  width: 40%;
  height: 12px;
}

.sk-line--md {
  width: 88%;
}

.sk-line--lg {
  height: 20px;
  width: 92%;
}

@keyframes blog-sk-shimmer {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: -100% 0;
  }
}
</style>
