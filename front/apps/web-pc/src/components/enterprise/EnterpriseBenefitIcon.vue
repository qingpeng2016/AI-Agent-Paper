<script setup lang="ts">
import type { EnterpriseBenefitIconName } from '@/mocks/enterprise'

defineProps<{ name: EnterpriseBenefitIconName }>()
</script>

<template>
  <svg
    class="ent-icon-svg"
    viewBox="0 0 24 24"
    fill="none"
    xmlns="http://www.w3.org/2000/svg"
    aria-hidden="true"
  >
    <!-- 批量开通：多用户 -->
    <template v-if="name === 'batch'">
      <circle cx="9" cy="8" r="3" stroke="currentColor" stroke-width="1.75" />
      <path
        d="M3.5 18.5v-.5c0-2.5 2.2-4.5 5.5-4.5s5.5 2 5.5 4.5v.5"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
      />
      <circle cx="17.5" cy="9" r="2.25" stroke="currentColor" stroke-width="1.75" />
      <path
        d="M14 18.5v-.3c0-1.6 1.5-2.9 3.5-2.9"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
      />
    </template>

    <!-- 对公转账：银行 / 机构 -->
    <template v-else-if="name === 'bank'">
      <path
        d="M4 10h16M6 10V19M10 10V19M14 10V19M18 10V19M3 19h18"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
      <path
        d="M12 4L3 10h18L12 4z"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linejoin="round"
      />
    </template>

    <!-- 增值税发票：票据 -->
    <template v-else-if="name === 'invoice'">
      <path
        d="M8 4h9l3 3v13H8V4z"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linejoin="round"
      />
      <path d="M17 4v3h3" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round" />
      <path
        d="M10 11h7M10 14h5M10 17h6"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
      />
    </template>

    <!-- 充到本人账号：用户 + 对勾 -->
    <template v-else-if="name === 'account'">
      <circle cx="12" cy="9" r="3.25" stroke="currentColor" stroke-width="1.75" />
      <path
        d="M5.5 19c0-3.2 2.9-5.5 6.5-5.5s6.5 2.3 6.5 5.5"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
      />
      <path
        d="M17.5 8.5l1.25 1.25L21 7"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
        stroke-linejoin="round"
      />
    </template>

    <!-- 合同与资质：文档 + 笔 -->
    <template v-else-if="name === 'contract'">
      <path
        d="M7 4h8l4 4v12H7V4z"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linejoin="round"
      />
      <path d="M15 4v4h4" stroke="currentColor" stroke-width="1.75" stroke-linejoin="round" />
      <path
        d="M9.5 13h6M9.5 16h4"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
      />
      <path
        d="M16.5 17.5l2.5 2.5M18 16l1 1"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
      />
    </template>

    <!-- 专属客服：耳机 -->
    <template v-else-if="name === 'support'">
      <path
        d="M4 14v-2a8 8 0 0116 0v2"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
      />
      <rect
        x="3"
        y="14"
        width="3"
        height="5"
        rx="1.25"
        stroke="currentColor"
        stroke-width="1.75"
      />
      <rect
        x="18"
        y="14"
        width="3"
        height="5"
        rx="1.25"
        stroke="currentColor"
        stroke-width="1.75"
      />
      <path
        d="M12 19v1.5"
        stroke="currentColor"
        stroke-width="1.75"
        stroke-linecap="round"
      />
    </template>
  </svg>
</template>

<style scoped>
.ent-icon-svg {
  display: block;
  width: 22px;
  height: 22px;
}
</style>
