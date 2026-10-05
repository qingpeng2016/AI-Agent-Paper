<script setup lang="ts">
import {
  UPSTREAM_LABEL,
  formatCny,
  type ProductsCategoryName,
} from '@ai-token-mall/shared'
import type { CatalogProduct } from '@/mocks/home'

defineProps<{
  product: CatalogProduct
}>()

const emit = defineEmits<{
  buy: [product: CatalogProduct]
}>()
</script>

<template>
  <article class="row">
    <div class="row-main">
      <div class="row-tags">
        <span v-if="product.hot_tag_name" class="tag hot">{{ product.hot_tag_name }}</span>
        <span class="tag line">{{
          UPSTREAM_LABEL[product.products_category_name as ProductsCategoryName] ??
            product.products_category_name
        }}</span>
      </div>
      <h3 class="title">{{ product.card_title }}</h3>
      <p class="desc">{{ product.card_subtitle }}</p>
      <ul class="feat">
        <li v-for="(f, i) in product.card_features.slice(0, 3)" :key="i">{{ f }}</li>
      </ul>
    </div>
    <div class="row-side">
      <p class="price">{{ formatCny(product.price) }}</p>
      <p class="sku">{{ product.sku_code }}</p>
      <button type="button" class="buy" @click="emit('buy', product)">购买</button>
    </div>
  </article>
</template>

<style scoped>
.row {
  display: flex;
  flex-direction: column;
  gap: 16px;
  padding: 20px 24px;
  background: #fff;
  border: 1px solid #e2e8f0;
  border-radius: 14px;
  transition: border-color 0.15s, box-shadow 0.15s;
}
@media (min-width: 640px) {
  .row {
    flex-direction: row;
    align-items: center;
  }
}
.row:hover {
  border-color: #c4b5fd;
  box-shadow: 0 8px 24px rgba(124, 58, 237, 0.08);
}
.row-main {
  flex: 1;
  min-width: 0;
}
.row-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 8px;
}
.tag {
  padding: 2px 8px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 6px;
  background: #f1f5f9;
  color: var(--atm-text-muted);
}
.tag.hot {
  background: #fef3c7;
  color: #b45309;
}
.tag.line {
  background: var(--atm-primary-light);
  color: var(--atm-primary);
}
.row .title {
  margin: 0 0 6px;
  font-size: 18px;
  font-weight: 700;
}
.desc {
  margin: 0 0 10px;
  font-size: 13px;
  color: var(--atm-text-muted);
  line-height: 1.5;
}
.feat {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 16px;
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: 12px;
  color: #64748b;
}
.feat li::before {
  content: '·';
  margin-right: 6px;
  color: #cbd5e1;
}
.row-side {
  flex-shrink: 0;
  text-align: right;
  min-width: 120px;
}
.price {
  margin: 0;
  font-size: 28px;
  font-weight: 800;
  color: var(--atm-primary);
}
.sku {
  margin: 4px 0 12px;
  font-size: 11px;
  color: #94a3b8;
  font-family: ui-monospace, monospace;
}
.buy {
  padding: 10px 28px;
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  background: var(--atm-text);
  border: none;
  border-radius: 10px;
  cursor: pointer;
}
.buy:hover {
  background: var(--atm-primary);
}
</style>
