<template>
  <article class="sku-card-skeleton" aria-hidden="true">
    <div class="sk sk-badge" />
    <div class="sk sk-title" />
    <div class="sk sk-price" />
    <div class="sk sk-desc" />
    <ul class="sk-features">
      <li v-for="i in 4" :key="i"><span class="sk sk-line" /></li>
    </ul>
    <div class="sk sk-btn" />
    <div class="sk sk-after" />
  </article>
</template>

<style scoped>
.sku-card-skeleton {
  display: flex;
  flex-direction: column;
  min-height: 380px;
  padding: 22px 22px 18px;
  background: #fff;
  border: 1px solid #eef2ff;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(30, 27, 75, 0.06);
  pointer-events: none;
}

.sk {
  border-radius: 8px;
  background: linear-gradient(
    90deg,
    #f1f5f9 0%,
    #e8ecf4 45%,
    #f1f5f9 90%
  );
  background-size: 200% 100%;
  animation: sku-shimmer 1.2s ease-in-out infinite;
}

.sk-badge {
  align-self: flex-end;
  width: 48px;
  height: 22px;
  margin-bottom: 8px;
  border-radius: 999px;
  opacity: 0;
}

.sk-title {
  width: 78%;
  height: 22px;
  margin-bottom: 12px;
}

.sk-price {
  width: 42%;
  height: 32px;
  margin-bottom: 14px;
}

.sk-desc {
  width: 100%;
  height: 40px;
  margin-bottom: 16px;
}

.sk-features {
  margin: 0 0 18px;
  padding: 0;
  list-style: none;
  flex: 1;
}

.sk-features li {
  margin-bottom: 10px;
}

.sk-line {
  display: block;
  width: 88%;
  height: 14px;
}

.sk-features li:nth-child(2) .sk-line {
  width: 72%;
}

.sk-btn {
  width: 100%;
  height: 44px;
  margin-bottom: 12px;
  border-radius: 12px;
}

.sk-after {
  width: 65%;
  height: 12px;
  margin: 0 auto;
  border-radius: 6px;
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
