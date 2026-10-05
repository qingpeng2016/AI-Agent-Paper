<script setup lang="ts">
export type CompareIconId =
  | 'pay'
  | 'shield'
  | 'terminal'
  | 'chip'
  | 'devices'
  | 'invoice'
  | 'support'

defineProps<{
  name: CompareIconId
}>()
</script>

<template>
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
    <!-- 支付方式 -->
    <template v-if="name === 'pay'">
      <rect x="2" y="5" width="20" height="14" rx="2.5" />
      <path d="M2 10h20" />
      <path d="M6 15h3" />
    </template>
    <!-- 额度 / 保障 -->
    <template v-else-if="name === 'shield'">
      <path d="M12 3l8 4.5v5.5c0 4.5-3.2 7.8-8 9-4.8-1.2-8-4.5-8-9V7.5L12 3z" />
      <path d="M9 12l2 2 4-4" />
    </template>
    <!-- Codex CLI -->
    <template v-else-if="name === 'terminal'">
      <path d="M6 9l4 3-4 3" />
      <path d="M12 15h6" />
      <rect x="3" y="5" width="18" height="14" rx="2" />
    </template>
    <!-- 顶级模型 -->
    <template v-else-if="name === 'chip'">
      <rect x="7" y="7" width="10" height="10" rx="1.5" />
      <path d="M10 7V4M14 7V4M10 17v3M14 17v3M7 10H4M7 14H4M17 10h3M17 14h3" />
    </template>
    <!-- 手机 + PC -->
    <template v-else-if="name === 'devices'">
      <rect x="5" y="3" width="8" height="14" rx="1.5" />
      <path d="M9 15h.01" />
      <rect x="13" y="6" width="8" height="11" rx="1.5" />
      <path d="M15 17h6" />
    </template>
    <!-- 企业报销 -->
    <template v-else-if="name === 'invoice'">
      <path d="M14 3H8a2 2 0 00-2 2v16l3-2 3 2 3-2 3 2V5a2 2 0 00-2-2z" />
      <path d="M10 8h6M10 12h6M10 16h4" />
    </template>
    <!-- 售后客服 -->
    <template v-else-if="name === 'support'">
      <path d="M4 11v3a2 2 0 002 2h1" />
      <path d="M20 11v3a2 2 0 01-2 2h-1" />
      <path d="M4 11a8 8 0 0116 0v-1a3 3 0 00-3-3h-1l-1.5-2h-5L8 7H7a3 3 0 00-3 3v1z" />
    </template>
  </svg>
</template>
