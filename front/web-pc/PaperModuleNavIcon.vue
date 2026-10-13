<script setup lang="ts">
import type { PaperModuleId } from './types'

defineProps<{
  moduleId: PaperModuleId
}>()
</script>

<template>
  <svg
    class="paper-module-nav-icon"
    viewBox="0 0 24 24"
    fill="none"
    stroke="currentColor"
    stroke-width="1.75"
    stroke-linecap="round"
    stroke-linejoin="round"
    aria-hidden="true"
  >
    <!-- 选题发现：方框 / 模块 -->
    <template v-if="moduleId === 'topic-discovery'">
      <rect x="4" y="4" width="16" height="16" rx="2" />
      <path d="M8 9h8M8 12h6M8 15h4" />
    </template>
    <!-- 文献综述：柱状 -->
    <template v-else-if="moduleId === 'literature-review'">
      <path d="M6 18V10M10 18V6M14 18v-5M18 18V8" />
    </template>
    <!-- 实验方案：烧瓶 -->
    <template v-else-if="moduleId === 'experiment-planning'">
      <path d="M10 3v6l-5 9a2 2 0 0 0 1.7 3h10.6a2 2 0 0 0 1.7-3l-5-9V3" />
      <path d="M8 3h8" />
    </template>
    <!-- 论文写作：笔 + 线 -->
    <template v-else-if="moduleId === 'paper-writing'">
      <path d="M12 20h9" />
      <path d="M14.5 6.5 17.5 3.5a2.1 2.1 0 0 1 3 3L8 19l-4 1 1-4 9.5-9.5z" />
    </template>
    <!-- 文献管理：书页 -->
    <template v-else-if="moduleId === 'user-literature'">
      <path d="M6 4h8a2 2 0 0 1 2 2v14a1 1 0 0 0-1-1H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2z" />
      <path d="M8 4v16M14 8h4M14 12h4M14 16h3" opacity="0.55" />
    </template>
    <!-- 图表管理：图框 -->
    <template v-else-if="moduleId === 'figure-generation'">
      <rect x="4" y="5" width="16" height="14" rx="2" />
      <circle cx="9" cy="11" r="1.5" fill="currentColor" stroke="none" />
      <path d="M4 16l4-4 3 3 5-6 4 5" />
    </template>
    <!-- 我的论文：叠层 -->
    <template v-else-if="moduleId === 'my-manuscripts'">
      <path d="M8 6h8l4 4v10a2 2 0 0 1-2 2H8a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2z" />
      <path d="M16 6v4h4M12 11h4M12 15h4" opacity="0.5" />
      <path d="M6 9H5a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h1" />
    </template>
    <!-- 邀请返利：礼物 -->
    <template v-else-if="moduleId === 'invite-rebate'">
      <rect x="4" y="8" width="16" height="12" rx="2" />
      <path d="M12 8V20M4 12h16" />
      <path d="M12 8c-2-3-5-3-5 0s3 3 5 0 5-3 5 0-3-3-5 0" />
    </template>
    <!-- 个人中心：用户 -->
    <template v-else-if="moduleId === 'personal-center'">
      <circle cx="12" cy="8" r="3.5" />
      <path d="M5 20c0-3.5 3.1-6 7-6s7 2.5 7 6" />
    </template>
    <!-- 环境配置：齿轮 -->
    <template v-else-if="moduleId === 'environment'">
      <circle cx="12" cy="12" r="3" />
      <path
        d="M12 2v2M12 20v2M4.9 4.9l1.4 1.4M17.7 17.7l1.4 1.4M2 12h2M20 12h2M4.9 19.1l1.4-1.4M17.7 6.3l1.4-1.4"
      />
    </template>
    <template v-else>
      <circle cx="12" cy="12" r="8" />
    </template>
  </svg>
</template>

<style scoped>
.paper-module-nav-icon {
  display: block;
  width: 18px;
  height: 18px;
}
</style>
