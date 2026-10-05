<template>
  <div class="article-skeleton atm-container article-container" aria-busy="true" aria-label="加载文章">
    <div class="sk sk-crumb" />
    <div class="sk sk-meta" />
    <div class="sk sk-headline" />
    <div class="sk sk-para" />
    <div class="sk sk-para" />
    <div class="sk sk-para sk-para--short" />
  </div>
</template>

<style scoped>
.article-container {
  max-width: 720px;
  padding: 28px 0 72px;
}

.sk {
  border-radius: 8px;
  background: linear-gradient(90deg, #eef1f6 0%, #e2e8f0 45%, #eef1f6 90%);
  background-size: 200% 100%;
  animation: article-sk-shimmer 1.2s ease-in-out infinite;
}

.sk-crumb {
  height: 14px;
  width: 200px;
  margin-bottom: 28px;
}

.sk-meta {
  height: 28px;
  width: 160px;
  margin-bottom: 16px;
  border-radius: 999px;
}

.sk-headline {
  height: 32px;
  width: 92%;
  margin-bottom: 28px;
}

.sk-para {
  height: 14px;
  width: 100%;
  margin-bottom: 14px;
}

.sk-para--short {
  width: 76%;
}

@keyframes article-sk-shimmer {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: -100% 0;
  }
}
</style>
