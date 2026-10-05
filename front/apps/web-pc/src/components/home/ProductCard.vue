<script setup lang="ts">
import { formatCny } from '@ai-token-mall/shared'
import type { CatalogProduct } from '@/mocks/home'

defineProps<{
  product: CatalogProduct
  selected?: boolean
}>()

const emit = defineEmits<{
  buy: [product: CatalogProduct]
  open: [product: CatalogProduct]
}>()
</script>

<template>
  <article
    class="sku-card"
    :class="{ 'sku-card--selected': selected }"
    role="button"
    tabindex="0"
    @click="emit('open', product)"
    @keydown.enter.prevent="emit('open', product)"
    @keydown.space.prevent="emit('open', product)"
  >
    <span v-if="product.hot_tag_name" class="sku-badge">{{ product.hot_tag_name }}</span>
    <h3 class="sku-title">{{ product.card_title }}</h3>
    <p class="sku-price">{{ formatCny(product.price) }}</p>
    <p class="sku-desc">{{ product.card_subtitle }}</p>
    <ul class="sku-features">
      <li v-for="(f, i) in product.card_features" :key="i">
        <span class="check" aria-hidden="true">✓</span>
        <span class="feature-text">{{ f }}</span>
      </li>
    </ul>
    <button type="button" class="sku-buy" @click.stop="emit('buy', product)">立即购买</button>
  </article>
</template>

<style scoped>
.sku-card {
  position: relative;
  display: flex;
  flex-direction: column;
  height: 100%;
  padding: 22px 22px 18px;
  background: #fff;
  border: 1px solid #eef2ff;
  border-radius: 16px;
  box-shadow: 0 4px 20px rgba(30, 27, 75, 0.06);
  transition:
    transform 0.15s,
    box-shadow 0.15s,
    border-color 0.15s;
  cursor: pointer;
}
.sku-card:focus-visible {
  outline: 2px solid var(--atm-primary);
  outline-offset: 2px;
}
.sku-card--selected {
  border: 2px solid var(--atm-primary);
  box-shadow:
    0 0 0 3px rgba(124, 58, 237, 0.18),
    0 12px 28px rgba(124, 58, 237, 0.14);
}
.sku-card:hover {
  transform: translateY(-2px);
  box-shadow: 0 12px 28px rgba(124, 58, 237, 0.12);
}
.sku-badge {
  position: absolute;
  top: 14px;
  right: 14px;
  padding: 4px 10px;
  font-size: 11px;
  font-weight: 700;
  color: #fff;
  background: var(--atm-gradient);
  border-radius: 999px;
}
.sku-title {
  margin: 0 0 8px;
  padding-right: 52px;
  font-size: 16px;
  font-weight: 700;
  color: var(--atm-text);
  line-height: 1.35;
}
.sku-price {
  margin: 0 0 10px;
  font-size: 30px;
  font-weight: 800;
  color: var(--atm-primary);
  letter-spacing: -0.02em;
}
.sku-desc {
  margin: 0 0 14px;
  font-size: 13px;
  line-height: 1.5;
  color: #64748b;
  min-height: 40px;
}
.sku-features {
  margin: 0 0 18px;
  padding: 0;
  list-style: none;
  flex: 1;
}
.sku-features li {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  margin-bottom: 10px;
  line-height: 1.4;
}
.sku-features .feature-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--atm-text);
}
.check {
  flex-shrink: 0;
  margin-top: 1px;
  font-size: 14px;
  color: #22c55e;
  font-weight: 800;
}
.sku-buy {
  width: 100%;
  padding: 13px;
  font-size: 15px;
  font-weight: 600;
  color: #fff;
  background: var(--atm-gradient);
  border: none;
  border-radius: 12px;
  cursor: pointer;
  box-shadow: 0 4px 14px rgba(124, 58, 237, 0.25);
}
.sku-buy:hover {
  filter: brightness(1.04);
}
</style>
