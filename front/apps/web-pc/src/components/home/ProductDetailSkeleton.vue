<template>
  <div class="product-page-skeleton atm-container" aria-busy="true" aria-label="加载中">
    <div class="hero-card">
      <div class="hero-main">
        <div class="sk sk-eyebrow" />
        <div class="sk sk-h1" />
        <div class="sk sk-lead" />
        <ul class="sk-list">
          <li v-for="i in 4" :key="i"><span class="sk sk-line" /></li>
        </ul>
        <div class="sk sk-btn" />
      </div>
      <aside class="buy-card">
        <div class="sk sk-title" />
        <div class="sk sk-price" />
        <div class="sk sk-desc" />
        <div v-for="i in 3" :key="`f-${i}`" class="sk sk-line buy-line" />
        <div class="sk sk-cta" />
      </aside>
    </div>
  </div>
</template>

<style scoped>
.product-page-skeleton {
  padding: 28px 0 56px;
}

.hero-card {
  display: grid;
  gap: 28px;
  padding: 32px 28px;
  background: #fff;
  border: 1px solid #ede9fe;
  border-radius: 24px;
  box-shadow: var(--atm-shadow-lg);
}

@media (min-width: 960px) {
  .hero-card {
    grid-template-columns: 1fr minmax(280px, 340px);
    gap: 36px;
    padding: 40px 36px;
  }
}

.buy-card {
  padding: 24px 22px;
  border: 2px solid #eef2ff;
  border-radius: 20px;
}

.sk {
  border-radius: 8px;
  background: linear-gradient(90deg, #f1f5f9 0%, #e8ecf4 45%, #f1f5f9 90%);
  background-size: 200% 100%;
  animation: sku-shimmer 1.2s ease-in-out infinite;
}

.sk-eyebrow {
  width: 140px;
  height: 28px;
  margin-bottom: 16px;
  border-radius: 999px;
}

.sk-h1 {
  width: 85%;
  height: 36px;
  margin-bottom: 16px;
}

.sk-lead {
  width: 100%;
  height: 48px;
  margin-bottom: 20px;
}

.sk-list {
  margin: 0 0 24px;
  padding: 0;
  list-style: none;
}

.sk-list li {
  margin-bottom: 12px;
}

.sk-line {
  display: block;
  width: 80%;
  height: 16px;
}

.buy-line {
  width: 90%;
  margin-bottom: 10px;
}

.sk-btn {
  width: 140px;
  height: 44px;
  border-radius: 12px;
}

.sk-title {
  width: 70%;
  height: 20px;
  margin-bottom: 12px;
}

.sk-price {
  width: 50%;
  height: 40px;
  margin-bottom: 16px;
}

.sk-desc {
  width: 100%;
  height: 36px;
  margin-bottom: 16px;
}

.sk-cta {
  width: 100%;
  height: 48px;
  margin-top: 8px;
  border-radius: 12px;
}

@keyframes sku-shimmer {
  0% {
    background-position: 100% 0;
  }
  100% {
    background-position: -100% 0;
  }
}
</style>
